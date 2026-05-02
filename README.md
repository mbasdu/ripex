# `ripex`

`ripex` is a Go CLI that downloads RIPE split database snapshots and builds IPv4 datasets for networks associated with Russian providers.

It currently produces:

- detailed CSV and JSONL datasets
- minimized CIDR prefix lists
- an Xray-compatible `GeoIPList` asset
- a manifest with counts and metadata

The intended use cases are:

- split tunneling and bypass routing
- OpenWrt/Xray routing resources
- firewall or policy-routing prefix feeds
- analysis of different RU-provider selection heuristics

## Datasets

`ripex` builds three dataset variants:

- `ru_org_inetnum_v4`
  - `inetnum` objects whose `org` points to an `organisation` with `country: RU`
- `ru_as_route_v4`
  - `route` objects whose `origin` ASN belongs to an `aut-num` linked to an RU `organisation`
- `ru_org_inetnum_plus_ru_as_route_v4`
  - union of the two datasets above

All datasets are IPv4-only in the current version.

## Inputs

`ripex` downloads these RIPE split snapshot files:

- `ripe.db.organisation.gz`
- `ripe.db.inetnum.gz`
- `ripe.db.aut-num.gz`
- `ripe.db.route.gz`

Default source:

- [https://ftp.ripe.net/ripe/dbase/split/](https://ftp.ripe.net/ripe/dbase/split/)

## Commands

Build the binary:

```bash
go build ./cmd/ripex
```

Fetch the RIPE snapshots:

```bash
go run ./cmd/ripex fetch
```

Build all datasets from the cached snapshots:

```bash
go run ./cmd/ripex build
```

Fetch and build in one step:

```bash
go run ./cmd/ripex run
```

Print minimized prefixes for a dataset:

```bash
go run ./cmd/ripex prefixes --dataset ru_as_route_v4
```

Print only route-derived minimized prefixes for the combined dataset:

```bash
go run ./cmd/ripex prefixes --dataset ru_org_inetnum_plus_ru_as_route_v4 --source-type route
```

Resolve a domain list to minimized IPv4 CIDRs (for example
itdoginfo/allow-domains's `Russia/outside-raw.lst` — Russian services
that geo-block non-RU IPs and therefore must not transit a foreign
proxy):

```bash
# Stage the list somewhere local; default path is data/outside/russia-outside.lst
go run ./cmd/ripex resolve

# Or point it at any newline-delimited domain list
go run ./cmd/ripex resolve --input /path/to/list.lst --dataset my_bypass_v4
```

The resolve step writes the same trio of artifacts as `build`
(`<dataset>.jsonl`, `.csv`, `.prefixes.txt`) into the `--out-dir`, so
the resulting prefix file is a drop-in for the same push/load pipeline
that handles RIPE-derived datasets. DNS lookups run locally against
the caller's configured resolver (override with `--resolver
1.1.1.1:53`), making the artifact reproducible to push to many routers
without ever resolving on-device.

For the curated host feeds, use:

```bash
scripts/build-ru-direct-list.sh
scripts/build-wl-hosts.sh
scripts/build-ru-router-list.sh
```

The resolver helpers accept `MODE=ssh` plus `SSH_TARGET=user@host` and
`SSH_WORKDIR=/path/to/ripex` when resolution should happen on a remote
host instead of the local machine.

`scripts/build-ru-router-list.sh` builds the canonical router artifact:

- `data/ripe/ru/ru_all_v4.prefixes.txt`

It merges and minimizes four inputs into one short deduped prefix list:

- the RIPE-derived RU provider dataset
  `ru_org_inetnum_plus_ru_as_route_v4.prefixes.txt`
- the direct-domain dataset resolved from `Russia/outside-raw.lst`
  `ru_direct_domains_v4.prefixes.txt`
- the curated VPN/IP-check probe host dataset
  `ru_probe_hosts_v4.prefixes.txt`
- the separate whitelist-host dataset fetched from
  `hxehex/russia-mobile-internet-whitelist`
  `ru_wl_hosts_v4.prefixes.txt`

So `ru_all_v4.prefixes.txt` and the release zip are not RIPE-only
artifacts. They intentionally include both RIPE-derived prefixes and
non-RIPE host-derived prefixes.

`scripts/build-ru-router-list.sh` also syncs the tracked release files:

- `assets/ru_all_v4.prefixes.txt`
- `assets/manifest.json`

Those are the release source-of-truth files committed to the repo.

The repository also includes a GitHub Actions workflow at
`.github/workflows/release-ru-net.yml` that does not rebuild datasets.
It only packages the tracked files from `assets/` into a zip bundle with
`SHA256SUMS` and publishes it as a GitHub Release asset on `v*` tags or
manual runs with `release_tag`.

You can build the same merged artifact directly with:

```bash
go run ./cmd/ripex merge-prefixes \
  --input data/ripe/ru/ru_org_inetnum_plus_ru_as_route_v4.prefixes.txt \
  --input data/ripe/ru/ru_direct_domains_v4.prefixes.txt \
  --input data/ripe/ru/ru_probe_hosts_v4.prefixes.txt \
  --input data/ripe/ru/ru_wl_hosts_v4.prefixes.txt \
  --out data/ripe/ru/ru_all_v4.prefixes.txt
```

Write minimized prefixes to a file:

```bash
go run ./cmd/ripex prefixes --dataset ru_org_inetnum_v4 --out /tmp/ru.txt
```

### Global flags

Available on `fetch`, `build`, `run`, `prefixes`, and `resolve` as applicable:

- `--base-url`
- `--cache-dir`
- `--out-dir`
- `--timeout`
- `--user-agent`

Defaults:

- cache dir: `data/ripe/cache`
- output dir: `data/ripe/ru`

## Generated Artifacts

For each dataset, `build` writes:

- `<dataset>.jsonl`
  - one detailed source row per record
- `<dataset>.csv`
  - same data in tabular form
- `<dataset>.prefixes.txt`
  - minimized IPv4 CIDRs, one per line

The full output set is:

- `data/ripe/ru/ru_org_inetnum_v4.jsonl`
- `data/ripe/ru/ru_org_inetnum_v4.csv`
- `data/ripe/ru/ru_org_inetnum_v4.prefixes.txt`
- `data/ripe/ru/ru_as_route_v4.jsonl`
- `data/ripe/ru/ru_as_route_v4.csv`
- `data/ripe/ru/ru_as_route_v4.prefixes.txt`
- `data/ripe/ru/ru_direct_domains_v4.jsonl` *(only when `resolve` or `scripts/build-ru-direct-list.sh` has been run)*
- `data/ripe/ru/ru_direct_domains_v4.csv`
- `data/ripe/ru/ru_direct_domains_v4.prefixes.txt`
- `data/ripe/ru/ru_probe_hosts_v4.jsonl` *(only when `scripts/build-ru-direct-list.sh` has been run)*
- `data/ripe/ru/ru_probe_hosts_v4.csv`
- `data/ripe/ru/ru_probe_hosts_v4.prefixes.txt`
- `data/ripe/ru/ru_wl_hosts_v4.jsonl` *(only when `scripts/build-wl-hosts.sh` has been run)*
- `data/ripe/ru/ru_wl_hosts_v4.csv`
- `data/ripe/ru/ru_wl_hosts_v4.prefixes.txt`
- `data/ripe/ru/ru_all_v4.prefixes.txt` *(canonical router feed)*
- `data/ripe/ru/ru_org_inetnum_plus_ru_as_route_v4.jsonl`
- `data/ripe/ru/ru_org_inetnum_plus_ru_as_route_v4.csv`
- `data/ripe/ru/ru_org_inetnum_plus_ru_as_route_v4.prefixes.txt`
- `data/ripe/ru/ripex.dat`
- `data/ripe/ru/manifest.json`

## Detailed vs Minimized Outputs

Detailed outputs keep source provenance:

- original prefix
- dataset
- source object type
- source key
- organisation and ASN metadata

Minimized outputs intentionally drop row-level provenance and treat each dataset as a pure IPv4 address set:

- exact duplicates are removed
- overlapping ranges are unioned
- adjacent ranges are collapsed into the smallest exact CIDR cover
- malformed prefixes are skipped during minimization and counted in the manifest

## Xray Artifact

`ripex.dat` is a protobuf `GeoIPList` asset compatible with Xray / V2Ray `ext:file:tag` IP routing rules.

Included tags:

- `ru_org_inetnum_v4`
- `ru_as_route_v4`
- `ru_org_inetnum_plus_ru_as_route_v4`
- `ru-providers`

`ru-providers` is an alias for the combined minimized dataset.

Example Xray usage:

```json
{
  "type": "field",
  "ip": [
    "ext:ripex.dat:ru-providers"
  ],
  "outboundTag": "direct"
}
```

Place `ripex.dat` in Xray’s asset directory, or configure the asset path via Xray environment variables/runtime settings.

## Manifest

`manifest.json` contains:

- generation time
- source snapshot date
- source URLs
- cached filenames
- parsed object counts
- per-dataset counts:
  - `detailed_rows`
  - `unique_prefixes`
  - `minimized_prefixes`
  - `invalid_prefixes_skipped`
- Xray artifact file name and exported tags
- build duration

## Whitelist Coverage Tests

The repository includes an opt-in live test that checks whether domains curated from published MinTsifry mobile-internet whitelist notices resolve to IPv4 addresses covered by the combined minimized dataset.

Build the datasets first:

```bash
go run ./cmd/ripex build --cache-dir data/ripe/cache --out-dir data/ripe/ru
```

Run the live whitelist test:

```bash
RIPEX_LIVE_WHITELIST=1 go test ./internal/whitelist -run TestRussiaMobileWhitelistCoverageLive -count=1
```

See [whitelist.md](/Users/bvt/work/ripex/docs/whitelist.md) for source notes and caveats.

## Development

Run the test suite:

```bash
go test ./...
```

Key implementation areas:

- [cli.go](/Users/bvt/work/ripex/internal/cli/cli.go)
- [build.go](/Users/bvt/work/ripex/internal/datasets/build.go)
- [geoip.go](/Users/bvt/work/ripex/internal/xraygeo/geoip.go)
- [rpsl.go](/Users/bvt/work/ripex/internal/parse/rpsl.go)
- `scripts/build-ru-direct-list.sh`
- `scripts/build-wl-hosts.sh`
- `scripts/build-ru-router-list.sh`
- `scripts/resolve-hosts.sh`

## Current Assumptions

- full snapshot reprocessing only
- IPv4 only
- `organisation.country == RU` is the root signal for RU-linked organisations
- route-derived datasets reflect routing registry data and may differ from pure allocation views
