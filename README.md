# `ip-lists`

`ip-lists` publishes curated Russian network lists and supporting
knowledge about filtering, TSPU behavior, and VPN-detection surfaces.

This repository is content-first:

- `lists/` contains the tracked public lists
- `sources/` contains tracked source inputs used to build some lists
- `docs/` explains sources and methodology
- `docs/library/` keeps longer research notes and source material
- `tools/ripex/` contains the Go generator used to rebuild the datasets

## Published Lists

List families are organized by source:

- `lists/ripe/`
  RIPE-derived provider prefix sets built from RIPE split snapshots
- `lists/domains/`
  prefixes resolved from curated direct-domain inputs
- `lists/probe/`
  prefixes resolved from curated VPN/IP-check probe hosts
- `lists/whitelist/`
  prefixes resolved from the separate mobile-whitelist hostname feed
- `lists/merged/`
  canonical merged outputs intended for direct consumption

The main consumer-facing artifact is:

- `lists/merged/ru_all_v4.prefixes.txt`

It is not a RIPE-only list. It intentionally merges:

- RIPE-derived RU provider prefixes
- direct-domain prefixes from `sources/domains/russia-outside.lst`
- probe-host prefixes from `sources/probe/ru-probe-hosts.lst`
- whitelist-host prefixes fetched from
  `hxehex/russia-mobile-internet-whitelist`

The release bundle publishes:

- `ru_all_v4.prefixes.txt`
- `manifest.json`
- `SHA256SUMS`

## Repo Layout

```text
lists/
  ripe/
  domains/
  probe/
  whitelist/
  merged/
sources/
  domains/
  probe/
docs/
  library/
  sources/
  methodology/
  filtering/
tools/
  ripex/
```

## Rebuild Flow

The generator lives under `tools/ripex`.

Build the tool:

```bash
cd tools/ripex
go build ./cmd/ripex
```

Build RIPE-derived datasets into local scratch output:

```bash
cd tools/ripex
go run ./cmd/ripex run
```

Default local paths:

- cache: `.cache/ripex/ripe`
- generated outputs: `build/ripex/ru`

Build the hostname-derived lists and republish tracked public prefixes:

```bash
bash tools/ripex/scripts/build-ru-direct-list.sh
bash tools/ripex/scripts/build-wl-hosts.sh
bash tools/ripex/scripts/build-ru-router-list.sh
```

After that, the tracked public lists under `lists/` are refreshed, and
`lists/merged/manifest.json` is synchronized from the latest RIPE build.

## Documentation

Start here:

- `docs/sources/ripe.md`
- `docs/sources/whitelist.md`
- `docs/methodology/list-composition.md`
- `docs/filtering/tspu-and-vpn-detection.md`

The longer research notes remain in `docs/library/` for traceability,
but they are not the primary documentation surface anymore.

## Releases

GitHub Releases package files from `lists/merged/`. The workflow does
not rebuild datasets in CI. Releases are packaging-only, based on the
tracked list state already committed to the repository.
