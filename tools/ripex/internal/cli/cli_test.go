package cli

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"google.golang.org/protobuf/proto"
)

func TestPrefixesCommandAndBuildOutputs(t *testing.T) {
	cacheDir := t.TempDir()
	outDir := t.TempDir()

	writeGzip(t, filepath.Join(cacheDir, "ripe.db.organisation.gz"), `
organisation: ORG-RU1-RIPE
org-name: RU Org
country: RU

organisation: ORG-NL1-RIPE
org-name: NL Org
country: NL
`)
	writeGzip(t, filepath.Join(cacheDir, "ripe.db.inetnum.gz"), `
inetnum: 10.0.0.0 - 10.0.1.255
org: ORG-RU1-RIPE
country: RU
netname: TEST
status: ALLOCATED PA

inetnum: 10.0.2.0 - 10.0.2.255
org: ORG-NL1-RIPE
country: NL
netname: TEST2
status: ALLOCATED PA
`)
	writeGzip(t, filepath.Join(cacheDir, "ripe.db.aut-num.gz"), `
aut-num: AS65001
as-name: RU-AS
org: ORG-RU1-RIPE
`)
	writeGzip(t, filepath.Join(cacheDir, "ripe.db.route.gz"), `
route: 10.0.2.0/24
origin: AS65001

route: bad-prefix
origin: AS65001
`)

	if err := Run([]string{"build", "--cache-dir", cacheDir, "--out-dir", outDir}); err != nil {
		t.Fatalf("build error = %v", err)
	}

	prefixFile := filepath.Join(outDir, "ru_org_inetnum_plus_ru_as_route_v4.prefixes.txt")
	data, err := os.ReadFile(prefixFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.TrimSpace(string(data)) != "10.0.0.0/23\n10.0.2.0/24" {
		t.Fatalf("prefixes file = %q", string(data))
	}

	geoData, err := os.ReadFile(filepath.Join(outDir, xrayGeoIPFile))
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", xrayGeoIPFile, err)
	}
	var geoList routercommon.GeoIPList
	if err := proto.Unmarshal(geoData, &geoList); err != nil {
		t.Fatalf("proto.Unmarshal() error = %v", err)
	}
	if len(geoList.Entry) != 4 {
		t.Fatalf("geoList entries = %d", len(geoList.Entry))
	}

	outPath := filepath.Join(outDir, "route-only.txt")
	if err := Run([]string{
		"prefixes",
		"--cache-dir", cacheDir,
		"--dataset", "ru_org_inetnum_plus_ru_as_route_v4",
		"--source-type", "route",
		"--out", outPath,
	}); err != nil {
		t.Fatalf("prefixes error = %v", err)
	}

	routeOnly, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(route-only) error = %v", err)
	}
	if strings.TrimSpace(string(routeOnly)) != "10.0.2.0/24" {
		t.Fatalf("route-only prefixes = %q", string(routeOnly))
	}
}

func TestPrefixesUnknownDataset(t *testing.T) {
	err := Run([]string{"prefixes", "--dataset", "nope"})
	if err == nil || !strings.Contains(err.Error(), "unknown dataset") {
		t.Fatalf("unexpected error = %v", err)
	}
}

func TestMergePrefixesCommand(t *testing.T) {
	dir := t.TempDir()
	in1 := filepath.Join(dir, "a.txt")
	in2 := filepath.Join(dir, "b.txt")
	out := filepath.Join(dir, "merged.txt")

	if err := os.WriteFile(in1, []byte("10.0.0.0/24\n10.0.1.0/24\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", in1, err)
	}
	if err := os.WriteFile(in2, []byte("10.0.0.0/23\n10.0.2.0/24\nbroken\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", in2, err)
	}

	if err := Run([]string{
		"merge-prefixes",
		"--input", in1,
		"--input", in2,
		"--out", out,
	}); err != nil {
		t.Fatalf("merge-prefixes error = %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", out, err)
	}
	if strings.TrimSpace(string(data)) != "10.0.0.0/23\n10.0.2.0/24" {
		t.Fatalf("merged prefixes = %q", string(data))
	}
}

func writeGzip(t *testing.T, path, content string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create(%s) error = %v", path, err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	if _, err := gz.Write([]byte(strings.TrimSpace(content) + "\n")); err != nil {
		t.Fatalf("Write(%s) error = %v", path, err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("Close(%s) error = %v", path, err)
	}
}
