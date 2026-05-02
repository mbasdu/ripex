package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ripex/internal/datasets"
	"ripex/internal/fetch"
	"ripex/internal/output"
	"ripex/internal/parse"
	"ripex/internal/resolve"
	"ripex/internal/ripe"
	"ripex/internal/xraygeo"
)

const defaultBaseURL = "https://ftp.ripe.net/ripe/dbase/split"
const xrayGeoIPFile = "ripex.dat"

type Config struct {
	BaseURL   string
	CacheDir  string
	OutDir    string
	Timeout   time.Duration
	UserAgent string
}

type PrefixesConfig struct {
	Config
	Dataset    string
	Out        string
	Format     string
	SourceType string
}

// ResolveConfig controls the `resolve` subcommand, which turns a plain
// list of domains into the same kind of minimized prefix files that `build`
// produces for RIPE-derived datasets.
type ResolveConfig struct {
	Input       string
	Dataset     string
	OutDir      string
	Resolver    string
	Concurrency int
	Timeout     time.Duration
	StrictFail  bool
}

type MergePrefixesConfig struct {
	Inputs []string
	Out    string
}

func Run(args []string) error {
	if len(args) == 0 {
		return usageError()
	}

	switch args[0] {
	case "fetch":
		cfg, err := parseFlags("fetch", args[1:])
		if err != nil {
			return err
		}
		return fetch.Run(toFetchConfig(cfg))
	case "build":
		cfg, err := parseFlags("build", args[1:])
		if err != nil {
			return err
		}
		return build(cfg)
	case "run":
		cfg, err := parseFlags("run", args[1:])
		if err != nil {
			return err
		}
		if err := fetch.Run(toFetchConfig(cfg)); err != nil {
			return err
		}
		return build(cfg)
	case "prefixes":
		cfg, err := parsePrefixesFlags(args[1:])
		if err != nil {
			return err
		}
		return prefixes(cfg)
	case "resolve":
		cfg, err := parseResolveFlags(args[1:])
		if err != nil {
			return err
		}
		return resolveDataset(cfg)
	case "merge-prefixes":
		cfg, err := parseMergePrefixesFlags(args[1:])
		if err != nil {
			return err
		}
		return mergePrefixes(cfg)
	default:
		return usageError()
	}
}

func parseFlags(name string, args []string) (Config, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	cfg := Config{}
	fs.StringVar(&cfg.BaseURL, "base-url", defaultBaseURL, "base URL for RIPE split snapshots")
	fs.StringVar(&cfg.CacheDir, "cache-dir", ".cache/ripex/ripe", "cache directory for downloaded RIPE snapshots")
	fs.StringVar(&cfg.OutDir, "out-dir", "build/ripex/ru", "output directory for generated datasets")
	fs.DurationVar(&cfg.Timeout, "timeout", 5*time.Minute, "HTTP timeout")
	fs.StringVar(&cfg.UserAgent, "user-agent", "ripex/0.1", "HTTP user agent")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func parsePrefixesFlags(args []string) (PrefixesConfig, error) {
	fs := flag.NewFlagSet("prefixes", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	cfg := PrefixesConfig{}
	fs.StringVar(&cfg.BaseURL, "base-url", defaultBaseURL, "base URL for RIPE split snapshots")
	fs.StringVar(&cfg.CacheDir, "cache-dir", ".cache/ripex/ripe", "cache directory for downloaded RIPE snapshots")
	fs.StringVar(&cfg.OutDir, "out-dir", "build/ripex/ru", "output directory for generated datasets")
	fs.DurationVar(&cfg.Timeout, "timeout", 5*time.Minute, "HTTP timeout")
	fs.StringVar(&cfg.UserAgent, "user-agent", "ripex/0.1", "HTTP user agent")
	fs.StringVar(&cfg.Dataset, "dataset", "", "dataset name")
	fs.StringVar(&cfg.Out, "out", "", "output file path; default is stdout")
	fs.StringVar(&cfg.Format, "format", "text", "output format")
	fs.StringVar(&cfg.SourceType, "source-type", "", "optional source type filter: inetnum or route")

	if err := fs.Parse(args); err != nil {
		return PrefixesConfig{}, err
	}
	return cfg, nil
}

func parseResolveFlags(args []string) (ResolveConfig, error) {
	fs := flag.NewFlagSet("resolve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	cfg := ResolveConfig{}
	fs.StringVar(&cfg.Input, "input", "sources/domains/russia-outside.lst", "newline-delimited list of domains to resolve")
	fs.StringVar(&cfg.Dataset, "dataset", "ru_direct_domains_v4", "output dataset name")
	fs.StringVar(&cfg.OutDir, "out-dir", "build/ripex/ru", "output directory for generated dataset files")
	fs.StringVar(&cfg.Resolver, "resolver", "", "optional DNS resolver host:port (default: system resolver)")
	fs.IntVar(&cfg.Concurrency, "concurrency", 16, "max parallel DNS queries")
	fs.DurationVar(&cfg.Timeout, "timeout", 5*time.Second, "per-domain DNS timeout")
	fs.BoolVar(&cfg.StrictFail, "strict-fail", false, "exit non-zero if any domain fails to resolve")

	if err := fs.Parse(args); err != nil {
		return ResolveConfig{}, err
	}
	return cfg, nil
}

type stringListFlag []string

func (s *stringListFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *stringListFlag) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func parseMergePrefixesFlags(args []string) (MergePrefixesConfig, error) {
	fs := flag.NewFlagSet("merge-prefixes", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	cfg := MergePrefixesConfig{}
	var inputs stringListFlag
	fs.Var(&inputs, "input", "input prefix file; repeatable")
	fs.StringVar(&cfg.Out, "out", "", "output file path; default is stdout")

	if err := fs.Parse(args); err != nil {
		return MergePrefixesConfig{}, err
	}
	cfg.Inputs = append(cfg.Inputs, inputs...)
	return cfg, nil
}

func usageError() error {
	return errors.New("usage: ripex <fetch|build|run|prefixes|resolve|merge-prefixes> [flags]")
}

func toFetchConfig(cfg Config) fetch.Config {
	return fetch.Config{
		BaseURL:   cfg.BaseURL,
		CacheDir:  cfg.CacheDir,
		UserAgent: cfg.UserAgent,
		Timeout:   cfg.Timeout,
	}
}

func build(cfg Config) error {
	start := time.Now()
	if err := os.MkdirAll(cfg.OutDir, 0o755); err != nil {
		return err
	}

	snapshot, counts, sourceURLs, cachedFiles, snapshotDate, err := loadSnapshot(cfg)
	if err != nil {
		return err
	}

	result, err := datasets.Build(snapshot, snapshotDate)
	if err != nil {
		return err
	}

	for _, name := range datasets.DatasetNames() {
		records := result.Rows[name]
		if err := output.WriteJSONL(filepath.Join(cfg.OutDir, name+".jsonl"), records); err != nil {
			return err
		}
		if err := output.WriteCSV(filepath.Join(cfg.OutDir, name+".csv"), records); err != nil {
			return err
		}
		if err := output.WriteLines(filepath.Join(cfg.OutDir, name+".prefixes.txt"), result.Minimized[name]); err != nil {
			return err
		}
	}

	xrayEntries := []xraygeo.Entry{
		{Tag: datasets.DatasetOrgInetnum, Prefixes: result.Minimized[datasets.DatasetOrgInetnum]},
		{Tag: datasets.DatasetASRoute, Prefixes: result.Minimized[datasets.DatasetASRoute]},
		{Tag: datasets.DatasetOrgInetnumPlusRoute, Prefixes: result.Minimized[datasets.DatasetOrgInetnumPlusRoute]},
		{Tag: "ru-providers", Prefixes: result.Minimized[datasets.DatasetOrgInetnumPlusRoute]},
	}
	if err := xraygeo.WriteFile(filepath.Join(cfg.OutDir, xrayGeoIPFile), xrayEntries); err != nil {
		return err
	}

	manifest := datasets.NewManifest(start, snapshotDate, sourceURLs, cachedFiles, counts, result.Stats, datasets.XrayArtifact{
		File: xrayGeoIPFile,
		Tags: []string{
			datasets.DatasetOrgInetnum,
			datasets.DatasetASRoute,
			datasets.DatasetOrgInetnumPlusRoute,
			"ru-providers",
		},
	})
	manifestPath := filepath.Join(cfg.OutDir, "manifest.json")
	return writeJSON(manifestPath, manifest)
}

func prefixes(cfg PrefixesConfig) error {
	if !datasets.IsValidDataset(cfg.Dataset) {
		return fmt.Errorf("unknown dataset %q", cfg.Dataset)
	}
	if cfg.Format != "text" {
		return fmt.Errorf("unsupported format %q", cfg.Format)
	}
	if cfg.SourceType != "" && cfg.SourceType != "inetnum" && cfg.SourceType != "route" {
		return fmt.Errorf("unsupported source type %q", cfg.SourceType)
	}

	snapshot, _, _, _, snapshotDate, err := loadSnapshot(cfg.Config)
	if err != nil {
		return err
	}
	result, err := datasets.Build(snapshot, snapshotDate)
	if err != nil {
		return err
	}

	prefixes, _ := datasets.MinimizeRecords(result.Rows[cfg.Dataset], cfg.SourceType)
	if cfg.Out != "" {
		return output.WriteLines(cfg.Out, prefixes)
	}
	return writeLines(os.Stdout, prefixes)
}

func resolveDataset(cfg ResolveConfig) error {
	start := time.Now()
	if cfg.Input == "" {
		return errors.New("--input is required")
	}
	if cfg.Dataset == "" {
		return errors.New("--dataset must not be empty")
	}
	if err := os.MkdirAll(cfg.OutDir, 0o755); err != nil {
		return err
	}

	domains, err := resolve.LoadDomains(cfg.Input)
	if err != nil {
		return fmt.Errorf("load domains: %w", err)
	}
	if len(domains) == 0 {
		return fmt.Errorf("no domains in %s", cfg.Input)
	}

	resolver := buildResolver(cfg.Resolver, cfg.Timeout)
	records := resolve.Resolve(context.Background(), resolver, domains, resolve.Config{
		Concurrency: cfg.Concurrency,
		Timeout:     cfg.Timeout,
	})
	summary := resolve.Summarise(records)

	snapshotDate := time.Now().UTC().Format("2006-01-02")
	outputs := buildOutputRecords(cfg.Dataset, snapshotDate, records)
	minimized, _ := datasets.MinimizeRecords(outputs, "")

	if err := output.WriteJSONL(filepath.Join(cfg.OutDir, cfg.Dataset+".jsonl"), outputs); err != nil {
		return err
	}
	if err := output.WriteCSV(filepath.Join(cfg.OutDir, cfg.Dataset+".csv"), outputs); err != nil {
		return err
	}
	if err := output.WriteLines(filepath.Join(cfg.OutDir, cfg.Dataset+".prefixes.txt"), minimized); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr,
		"resolve: %s: %d domains, %d resolved, %d failed, %d unique IPs, %d minimized prefixes (%s)\n",
		cfg.Dataset,
		summary.Domains,
		summary.ResolvedDomains,
		summary.FailedDomains,
		summary.UniqueAddrs,
		len(minimized),
		time.Since(start).Round(time.Millisecond),
	)

	if cfg.StrictFail {
		return resolve.FailuresError(records)
	}
	return nil
}

func mergePrefixes(cfg MergePrefixesConfig) error {
	if len(cfg.Inputs) == 0 {
		return errors.New("--input is required")
	}

	var prefixes []string
	for _, path := range cfg.Inputs {
		lines, err := readPrefixFile(path)
		if err != nil {
			return err
		}
		prefixes = append(prefixes, lines...)
	}

	minimized, stats := datasets.MinimizePrefixes(prefixes)
	if cfg.Out != "" {
		if err := output.WriteLines(cfg.Out, minimized); err != nil {
			return err
		}
	} else if err := writeLines(os.Stdout, minimized); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr,
		"merge-prefixes: %d inputs, %d raw prefixes, %d unique prefixes, %d minimized prefixes, %d invalid skipped\n",
		len(cfg.Inputs),
		len(prefixes),
		stats.UniquePrefixes,
		stats.MinimizedPrefixes,
		stats.InvalidPrefixesSkipped,
	)
	return nil
}

func buildResolver(addr string, timeout time.Duration) resolve.Resolver {
	if addr == "" {
		return net.DefaultResolver
	}
	dialer := net.Dialer{Timeout: timeout}
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			// Upstream is fixed by the operator; ignore the stock address.
			return dialer.DialContext(ctx, network, addr)
		},
	}
}

func buildOutputRecords(dataset, snapshotDate string, records []resolve.Record) []output.Record {
	var out []output.Record
	for _, r := range records {
		for _, addr := range r.Addrs {
			out = append(out, output.Record{
				Prefix:           addr.String() + "/32",
				Dataset:          dataset,
				SourceObjectType: "domain",
				SourceKey:        r.Domain,
				SnapshotDate:     snapshotDate,
			})
		}
	}
	return out
}

func readPrefixFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	var prefixes []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		prefixes = append(prefixes, line)
	}
	return prefixes, nil
}

func loadSnapshot(cfg Config) (ripe.SnapshotData, ripe.ParseCounts, []string, []string, string, error) {
	var snapshot ripe.SnapshotData
	var counts ripe.ParseCounts
	snapshotDate := time.Now().UTC().Format("2006-01-02")

	type fileSpec struct {
		name  string
		types map[string]struct{}
		load  func(parse.Object)
	}

	specs := []fileSpec{
		{
			name:  "ripe.db.organisation.gz",
			types: map[string]struct{}{"organisation": {}},
			load: func(obj parse.Object) {
				snapshot.Organisations = append(snapshot.Organisations, parse.ToOrganisation(obj))
				counts.Organisations++
			},
		},
		{
			name:  "ripe.db.inetnum.gz",
			types: map[string]struct{}{"inetnum": {}},
			load: func(obj parse.Object) {
				snapshot.Inetnums = append(snapshot.Inetnums, parse.ToInetnum(obj))
				counts.Inetnums++
			},
		},
		{
			name:  "ripe.db.aut-num.gz",
			types: map[string]struct{}{"aut-num": {}},
			load: func(obj parse.Object) {
				snapshot.AutNums = append(snapshot.AutNums, parse.ToAutNum(obj))
				counts.AutNums++
			},
		},
		{
			name:  "ripe.db.route.gz",
			types: map[string]struct{}{"route": {}},
			load: func(obj parse.Object) {
				snapshot.Routes = append(snapshot.Routes, parse.ToRoute(obj))
				counts.Routes++
			},
		},
	}

	var cachedFiles []string
	var sourceURLs []string
	for _, spec := range specs {
		path := filepath.Join(cfg.CacheDir, spec.name)
		if _, err := os.Stat(path); err != nil {
			return ripe.SnapshotData{}, ripe.ParseCounts{}, nil, nil, "", fmt.Errorf("missing cached snapshot %s; run `ripex fetch` first", path)
		}
		if err := parse.ParseGzipFile(path, spec.types, func(obj parse.Object) error {
			spec.load(obj)
			return nil
		}); err != nil {
			return ripe.SnapshotData{}, ripe.ParseCounts{}, nil, nil, "", err
		}
		cachedFiles = append(cachedFiles, path)
		sourceURLs = append(sourceURLs, cfg.BaseURL+"/"+spec.name)
	}

	return snapshot, counts, sourceURLs, cachedFiles, snapshotDate, nil
}

func writeJSON(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func writeLines(w io.Writer, lines []string) error {
	if len(lines) == 0 {
		return nil
	}
	_, err := io.WriteString(w, strings.Join(lines, "\n")+"\n")
	return err
}
