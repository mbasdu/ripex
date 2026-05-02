# List Composition

The repository distinguishes between local scratch build output and
tracked public lists.

## Scratch Output

Generated scratch output goes to:

- `.cache/ripex/ripe`
- `build/ripex/ru`

These directories are implementation detail and are not the public
distribution surface.

## Public Lists

Tracked public lists live under `lists/`:

- `lists/ripe/`
- `lists/domains/`
- `lists/probe/`
- `lists/whitelist/`
- `lists/merged/`

The canonical merged list is:

- `lists/merged/ru_all_v4.prefixes.txt`

That list is assembled by `tools/ripex/scripts/build-ru-router-list.sh`
from:

- `ru_org_inetnum_plus_ru_as_route_v4.prefixes.txt`
- `ru_direct_domains_v4.prefixes.txt`
- `ru_probe_hosts_v4.prefixes.txt`
- `ru_wl_hosts_v4.prefixes.txt`

The same script also republishes the individual source-family prefix
lists into their tracked `lists/` directories and synchronizes
`lists/merged/manifest.json`.
