# RIPE Source

The RIPE-derived lists are built from the RIPE split database snapshots:

- `ripe.db.organisation.gz`
- `ripe.db.inetnum.gz`
- `ripe.db.aut-num.gz`
- `ripe.db.route.gz`

Default upstream:

- https://ftp.ripe.net/ripe/dbase/split/

The generator under `tools/ripex/` produces three IPv4 dataset families:

- `ru_org_inetnum_v4`
- `ru_as_route_v4`
- `ru_org_inetnum_plus_ru_as_route_v4`

Tracked public outputs derived from those datasets live in `lists/ripe/`.

Local scratch outputs are written to `build/ripex/ru/` and include CSV,
JSONL, minimized prefixes, `ripex.dat`, and a manifest.

Rebuild command:

```bash
cd tools/ripex
go run ./cmd/ripex run
```
