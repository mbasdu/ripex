package datasets

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"net"
	"sort"
	"strings"
	"time"

	"ripex/internal/output"
	"ripex/internal/ripe"
)

const (
	DatasetOrgInetnum          = "ru_org_inetnum_v4"
	DatasetOrgInetnumPlusRoute = "ru_org_inetnum_plus_ru_as_route_v4"
	DatasetASRoute             = "ru_as_route_v4"
)

type DatasetStats struct {
	DetailedRows           int `json:"detailed_rows"`
	UniquePrefixes         int `json:"unique_prefixes"`
	MinimizedPrefixes      int `json:"minimized_prefixes"`
	InvalidPrefixesSkipped int `json:"invalid_prefixes_skipped"`
}

type Manifest struct {
	GeneratedAt  string                  `json:"generated_at"`
	SnapshotDate string                  `json:"snapshot_date"`
	SourceURLs   []string                `json:"source_urls"`
	CachedFiles  []string                `json:"cached_files"`
	Parsed       ripe.ParseCounts        `json:"parsed"`
	Datasets     map[string]DatasetStats `json:"datasets"`
	Xray         XrayArtifact            `json:"xray"`
	DurationMS   int64                   `json:"duration_ms"`
}

type XrayArtifact struct {
	File string   `json:"file"`
	Tags []string `json:"tags"`
}

type BuildResult struct {
	Rows      map[string][]output.Record
	Minimized map[string][]string
	Stats     map[string]DatasetStats
}

type interval struct {
	start uint32
	end   uint32
}

func DatasetNames() []string {
	return []string{
		DatasetOrgInetnum,
		DatasetOrgInetnumPlusRoute,
		DatasetASRoute,
	}
}

func IsValidDataset(name string) bool {
	for _, dataset := range DatasetNames() {
		if dataset == name {
			return true
		}
	}
	return false
}

func Build(snapshot ripe.SnapshotData, snapshotDate string) (BuildResult, error) {
	ruOrgs := make(map[string]ripe.Organisation)
	for _, org := range snapshot.Organisations {
		if strings.EqualFold(org.Country, "RU") && org.ID != "" {
			ruOrgs[org.ID] = org
		}
	}

	ruASNs := make(map[string]ripe.AutNum)
	for _, aut := range snapshot.AutNums {
		if aut.OrgID == "" || aut.ASN == "" {
			continue
		}
		if _, ok := ruOrgs[aut.OrgID]; ok {
			ruASNs[aut.ASN] = aut
		}
	}

	rows := map[string][]output.Record{
		DatasetOrgInetnum:          {},
		DatasetOrgInetnumPlusRoute: {},
		DatasetASRoute:             {},
	}
	seen := make(map[string]struct{})

	add := func(dataset string, record output.Record) {
		key := strings.Join([]string{
			dataset,
			record.Prefix,
			record.SourceObjectType,
			record.SourceKey,
		}, "|")
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		rows[dataset] = append(rows[dataset], record)
	}

	for _, inetnum := range snapshot.Inetnums {
		if inetnum.OrgID == "" {
			continue
		}
		org, ok := ruOrgs[inetnum.OrgID]
		if !ok {
			continue
		}
		prefixes, err := inetnumRangeToCIDRs(inetnum.Key)
		if err != nil {
			return BuildResult{}, fmt.Errorf("inetnum %s: %w", inetnum.Key, err)
		}
		for _, prefix := range prefixes {
			record := output.Record{
				Prefix:           prefix,
				Dataset:          DatasetOrgInetnum,
				SourceObjectType: "inetnum",
				SourceKey:        inetnum.Key,
				OrgID:            org.ID,
				OrgName:          org.Name,
				OrgCountry:       org.Country,
				SnapshotDate:     snapshotDate,
			}
			add(DatasetOrgInetnum, record)
			record.Dataset = DatasetOrgInetnumPlusRoute
			add(DatasetOrgInetnumPlusRoute, record)
		}
	}

	for _, route := range snapshot.Routes {
		if route.Origin == "" || route.Prefix == "" {
			continue
		}
		aut, ok := ruASNs[route.Origin]
		if !ok {
			continue
		}
		org := ruOrgs[aut.OrgID]
		record := output.Record{
			Prefix:           route.Prefix,
			Dataset:          DatasetASRoute,
			SourceObjectType: "route",
			SourceKey:        route.Prefix + "," + route.Origin,
			OrgID:            org.ID,
			OrgName:          org.Name,
			OrgCountry:       org.Country,
			ASN:              aut.ASN,
			AsName:           aut.AsName,
			SnapshotDate:     snapshotDate,
		}
		add(DatasetASRoute, record)
		record.Dataset = DatasetOrgInetnumPlusRoute
		add(DatasetOrgInetnumPlusRoute, record)
	}

	for _, dataset := range DatasetNames() {
		sort.Slice(rows[dataset], func(i, j int) bool {
			if rows[dataset][i].Prefix == rows[dataset][j].Prefix {
				return rows[dataset][i].SourceKey < rows[dataset][j].SourceKey
			}
			return rows[dataset][i].Prefix < rows[dataset][j].Prefix
		})
	}

	stats := make(map[string]DatasetStats, len(rows))
	minimized := make(map[string][]string, len(rows))
	for _, dataset := range DatasetNames() {
		prefixes, datasetStats := MinimizeRecords(rows[dataset], "")
		datasetStats.DetailedRows = len(rows[dataset])
		stats[dataset] = datasetStats
		minimized[dataset] = prefixes
	}

	return BuildResult{
		Rows:      rows,
		Minimized: minimized,
		Stats:     stats,
	}, nil
}

func MinimizeRecords(records []output.Record, sourceType string) ([]string, DatasetStats) {
	filtered := filterRecords(records, sourceType)
	uniqueRaw := make(map[string]struct{})
	uniqueValid := make(map[string]struct{})
	var intervals []interval
	stats := DatasetStats{DetailedRows: len(filtered)}

	for _, record := range filtered {
		uniqueRaw[record.Prefix] = struct{}{}
		start, end, ok := parseIPv4CIDR(record.Prefix)
		if !ok {
			stats.InvalidPrefixesSkipped++
			continue
		}
		if _, seen := uniqueValid[record.Prefix]; seen {
			continue
		}
		uniqueValid[record.Prefix] = struct{}{}
		intervals = append(intervals, interval{start: start, end: end})
	}

	stats.UniquePrefixes = len(uniqueRaw)
	if len(intervals) == 0 {
		return nil, stats
	}

	sort.Slice(intervals, func(i, j int) bool {
		if intervals[i].start == intervals[j].start {
			return intervals[i].end < intervals[j].end
		}
		return intervals[i].start < intervals[j].start
	})

	merged := make([]interval, 0, len(intervals))
	current := intervals[0]
	for _, next := range intervals[1:] {
		if next.start <= current.end || (current.end != ^uint32(0) && next.start == current.end+1) {
			if next.end > current.end {
				current.end = next.end
			}
			continue
		}
		merged = append(merged, current)
		current = next
	}
	merged = append(merged, current)

	var prefixes []string
	for _, item := range merged {
		prefixes = append(prefixes, rangeToCIDRs(item.start, item.end)...)
	}
	stats.MinimizedPrefixes = len(prefixes)
	return prefixes, stats
}

func MinimizePrefixes(prefixes []string) ([]string, DatasetStats) {
	records := make([]output.Record, 0, len(prefixes))
	for _, prefix := range prefixes {
		records = append(records, output.Record{Prefix: prefix})
	}
	return MinimizeRecords(records, "")
}

func NewManifest(start time.Time, snapshotDate string, sourceURLs, cachedFiles []string, counts ripe.ParseCounts, stats map[string]DatasetStats, xray XrayArtifact) Manifest {
	return Manifest{
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		SnapshotDate: snapshotDate,
		SourceURLs:   sourceURLs,
		CachedFiles:  cachedFiles,
		Parsed:       counts,
		Datasets:     stats,
		Xray:         xray,
		DurationMS:   time.Since(start).Milliseconds(),
	}
}

func filterRecords(records []output.Record, sourceType string) []output.Record {
	if sourceType == "" {
		return records
	}
	filtered := make([]output.Record, 0, len(records))
	for _, record := range records {
		if record.SourceObjectType == sourceType {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

func parseIPv4CIDR(value string) (uint32, uint32, bool) {
	ip, network, err := net.ParseCIDR(strings.TrimSpace(value))
	if err != nil {
		return 0, 0, false
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return 0, 0, false
	}
	ones, bits := network.Mask.Size()
	if bits != 32 {
		return 0, 0, false
	}

	start := binary.BigEndian.Uint32(ip4.Mask(network.Mask))
	hostBits := 32 - uint32(ones)
	var end uint32
	if hostBits == 32 {
		end = ^uint32(0)
	} else {
		end = start + (uint32(1) << hostBits) - 1
	}
	return start, end, true
}

func rangeToCIDRs(start, end uint32) []string {
	var cidrs []string
	for start <= end {
		remaining := end - start + 1
		hostBitsByAlignment := uint32(bits.TrailingZeros32(start))
		hostBitsByRemaining := uint32(bits.Len32(remaining) - 1)
		hostBits := hostBitsByAlignment
		if hostBits > hostBitsByRemaining {
			hostBits = hostBitsByRemaining
		}
		prefixLen := uint32(32) - hostBits

		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, start)
		cidrs = append(cidrs, fmt.Sprintf("%s/%d", ip.String(), prefixLen))

		if hostBits == 32 {
			break
		}
		blockSize := uint32(1) << hostBits
		start += blockSize
	}
	return cidrs
}

func inetnumRangeToCIDRs(value string) ([]string, error) {
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid inetnum range")
	}
	startIP := net.ParseIP(strings.TrimSpace(parts[0])).To4()
	endIP := net.ParseIP(strings.TrimSpace(parts[1])).To4()
	if startIP == nil || endIP == nil {
		return nil, fmt.Errorf("not an IPv4 range")
	}

	start := binary.BigEndian.Uint32(startIP)
	end := binary.BigEndian.Uint32(endIP)
	if start > end {
		return nil, fmt.Errorf("range start is after end")
	}
	return rangeToCIDRs(start, end), nil
}
