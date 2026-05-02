package datasets

import (
	"reflect"
	"testing"

	"ripex/internal/output"
	"ripex/internal/ripe"
)

func TestInetnumRangeToCIDRs(t *testing.T) {
	got, err := inetnumRangeToCIDRs("192.0.2.0 - 192.0.2.255")
	if err != nil {
		t.Fatalf("inetnumRangeToCIDRs() error = %v", err)
	}
	if len(got) != 1 || got[0] != "192.0.2.0/24" {
		t.Fatalf("got = %#v", got)
	}
}

func TestBuildDatasets(t *testing.T) {
	snapshot := ripe.SnapshotData{
		Organisations: []ripe.Organisation{
			{ID: "ORG-RU1-RIPE", Name: "RU Org", Country: "RU"},
			{ID: "ORG-NL1-RIPE", Name: "NL Org", Country: "NL"},
		},
		Inetnums: []ripe.Inetnum{
			{Key: "192.0.2.0 - 192.0.2.255", OrgID: "ORG-RU1-RIPE"},
			{Key: "198.51.100.0 - 198.51.100.255", OrgID: "ORG-NL1-RIPE"},
		},
		AutNums: []ripe.AutNum{
			{ASN: "AS65001", AsName: "RU-AS", OrgID: "ORG-RU1-RIPE"},
			{ASN: "AS65002", AsName: "NL-AS", OrgID: "ORG-NL1-RIPE"},
		},
		Routes: []ripe.Route{
			{Prefix: "203.0.113.0/24", Origin: "AS65001"},
			{Prefix: "203.0.114.0/24", Origin: "AS65002"},
			{Prefix: "bad-prefix", Origin: "AS65001"},
		},
	}

	result, err := Build(snapshot, "2026-03-22")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(result.Rows[DatasetOrgInetnum]) != 1 {
		t.Fatalf("org inetnum rows = %d", len(result.Rows[DatasetOrgInetnum]))
	}
	if len(result.Rows[DatasetASRoute]) != 2 {
		t.Fatalf("as route rows = %d", len(result.Rows[DatasetASRoute]))
	}
	if len(result.Rows[DatasetOrgInetnumPlusRoute]) != 3 {
		t.Fatalf("combined rows = %d", len(result.Rows[DatasetOrgInetnumPlusRoute]))
	}
	if result.Stats[DatasetASRoute].InvalidPrefixesSkipped != 1 {
		t.Fatalf("invalid skipped = %d", result.Stats[DatasetASRoute].InvalidPrefixesSkipped)
	}
	if result.Stats[DatasetOrgInetnumPlusRoute].MinimizedPrefixes != 2 {
		t.Fatalf("combined minimized prefixes = %d", result.Stats[DatasetOrgInetnumPlusRoute].MinimizedPrefixes)
	}
}

func TestMinimizeRecords(t *testing.T) {
	records := []output.Record{
		{Prefix: "10.0.0.0/24", SourceObjectType: "route"},
		{Prefix: "10.0.1.0/24", SourceObjectType: "route"},
		{Prefix: "10.0.0.0/23", SourceObjectType: "route"},
		{Prefix: "10.0.2.0/24", SourceObjectType: "inetnum"},
		{Prefix: "10.0.3.0/24", SourceObjectType: "inetnum"},
		{Prefix: "broken", SourceObjectType: "route"},
	}

	got, stats := MinimizeRecords(records, "")
	want := []string{"10.0.0.0/22"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MinimizeRecords() = %#v, want %#v", got, want)
	}
	if stats.UniquePrefixes != 6 {
		t.Fatalf("UniquePrefixes = %d", stats.UniquePrefixes)
	}
	if stats.InvalidPrefixesSkipped != 1 {
		t.Fatalf("InvalidPrefixesSkipped = %d", stats.InvalidPrefixesSkipped)
	}
}

func TestMinimizePrefixes(t *testing.T) {
	got, stats := MinimizePrefixes([]string{
		"10.0.0.0/24",
		"10.0.1.0/24",
		"10.0.0.0/23",
		"bad",
	})
	want := []string{"10.0.0.0/23"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MinimizePrefixes() = %#v, want %#v", got, want)
	}
	if stats.UniquePrefixes != 4 {
		t.Fatalf("UniquePrefixes = %d", stats.UniquePrefixes)
	}
	if stats.InvalidPrefixesSkipped != 1 {
		t.Fatalf("InvalidPrefixesSkipped = %d", stats.InvalidPrefixesSkipped)
	}
}

func TestMinimizeRecordsSourceType(t *testing.T) {
	records := []output.Record{
		{Prefix: "10.0.0.0/24", SourceObjectType: "route"},
		{Prefix: "10.0.1.0/24", SourceObjectType: "route"},
		{Prefix: "10.0.2.0/24", SourceObjectType: "inetnum"},
	}

	got, _ := MinimizeRecords(records, "route")
	want := []string{"10.0.0.0/23"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("route-only prefixes = %#v, want %#v", got, want)
	}
}

func TestRangeToCIDRsProducesCanonicalNetworks(t *testing.T) {
	got := rangeToCIDRs(0x023B4C00, 0x023B53FF) // 2.59.76.0 - 2.59.83.255
	want := []string{"2.59.76.0/22", "2.59.80.0/22"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("rangeToCIDRs() = %#v, want %#v", got, want)
	}
}
