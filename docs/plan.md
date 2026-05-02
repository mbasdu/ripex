# `ripex` Go CLI for RU Prefix Datasets from RIPE Snapshots

## Summary

Build a single Go command-line tool named `ripex` that downloads the required RIPE split snapshot files, parses them locally, and generates multiple IPv4 raw datasets for different "Russian provider" heuristics. The tool uses full daily snapshot reprocessing only; no diffing, NRTM, or incremental state management is included.

The first version produces JSONL and CSV datasets from these RIPE object types:

- `organisation`
- `inetnum`
- `aut-num`
- `route`

It supports three dataset strategies so they can be compared experimentally.

## CLI and Interfaces

- Binary name: `ripex`
- Primary commands:
  - `ripex fetch`
    - Download required RIPE split files into a cache directory
    - Files fetched:
      - `ripe.db.organisation.gz`
      - `ripe.db.inetnum.gz`
      - `ripe.db.aut-num.gz`
      - `ripe.db.route.gz`
  - `ripex build`
    - Parse cached files and write all datasets
- `ripex run`
  - Execute `fetch` then `build`
- `ripex prefixes`
  - Emit minimized IPv4 prefixes for a selected dataset
  - Supports optional `--source-type inetnum|route`
- Shared flags:
  - `--cache-dir`
  - `--out-dir`
  - `--base-url`
  - `--timeout`
  - `--user-agent`
- Default locations:
  - cache: `data/ripe/cache`
  - outputs: `data/ripe/ru`

## Outputs

- `data/ripe/ru/ru_org_inetnum_v4.jsonl`
- `data/ripe/ru/ru_org_inetnum_v4.csv`
- `data/ripe/ru/ru_org_inetnum_v4.prefixes.txt`
- `data/ripe/ru/ru_org_inetnum_plus_ru_as_route_v4.jsonl`
- `data/ripe/ru/ru_org_inetnum_plus_ru_as_route_v4.csv`
- `data/ripe/ru/ru_org_inetnum_plus_ru_as_route_v4.prefixes.txt`
- `data/ripe/ru/ru_as_route_v4.jsonl`
- `data/ripe/ru/ru_as_route_v4.csv`
- `data/ripe/ru/ru_as_route_v4.prefixes.txt`
- `data/ripe/ru/ripex.dat`
- `data/ripe/ru/manifest.json`

## Implementation Notes

- Download only the four required split snapshot files from `https://ftp.ripe.net/ripe/dbase/split/`
- Stream-decompress `.gz` files and parse RPSL objects separated by blank lines
- Extract only needed attributes:
  - `organisation`: `organisation`, `org-name`, `country`
  - `inetnum`: `inetnum`, `org`, `country`, `netname`, `status`
  - `aut-num`: `aut-num`, `as-name`, `org`
  - `route`: `route`, `origin`
- Build three datasets:
  - `ru_org_inetnum_v4`
  - `ru_org_inetnum_plus_ru_as_route_v4`
  - `ru_as_route_v4`
- For each dataset, also generate a minimized IPv4 prefix list:
  - union overlapping prefixes
  - merge adjacent ranges where a valid supernet exists
  - emit the smallest exact CIDR cover
- Generate an Xray-compatible GeoIP asset file:
  - protobuf `GeoIPList` format compatible with `ext:file:tag`
  - one tag per dataset plus `ru-providers` alias for the combined dataset
- Convert `inetnum` address ranges to minimal covering IPv4 CIDRs
- Preserve route prefixes exactly as registered
- Keep detailed CSV/JSONL rows unchanged; minimized outputs intentionally lose row-level provenance

## Test Plan

- Unit tests for RPSL parsing
- Unit tests for `inetnum` range to CIDR conversion
- Unit tests for prefix minimization and invalid prefix skipping
- Unit tests for dataset inclusion logic
- Fixture-based end-to-end test for `build` and `prefixes`

## Assumptions

- Full snapshot reprocessing once per day is the intended operating model
- No incremental diffs, NRTM handling, or change tracking
- IPv4 only in v1
- `organisation.country == RU` is the root signal for Russian organisations
- The tool generates raw datasets only and does not update firewall config files
