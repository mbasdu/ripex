# Whitelist Source

The repository tracks two local hostname-source files:

- `sources/domains/russia-outside.lst`
- `sources/probe/ru-probe-hosts.lst`

It also consumes one remote hostname feed at build time:

- `https://raw.githubusercontent.com/hxehex/russia-mobile-internet-whitelist/main/whitelist.txt`

These hostname inputs are resolved into public prefix lists:

- `lists/domains/ru_direct_domains_v4.prefixes.txt`
- `lists/probe/ru_probe_hosts_v4.prefixes.txt`
- `lists/whitelist/ru_wl_hosts_v4.prefixes.txt`

Build commands:

```bash
bash tools/ripex/scripts/build-ru-direct-list.sh
bash tools/ripex/scripts/build-wl-hosts.sh
```

## Live Coverage Check

The opt-in live whitelist test remains in the `ripex` tool module.

Build the RIPE dataset first, then run:

```bash
cd tools/ripex
RIPEX_LIVE_WHITELIST=1 go test ./internal/whitelist -run TestRussiaMobileWhitelistCoverageLive -count=1
```

The live test checks the curated whitelist fixtures in
`tools/ripex/testdata/` against `lists/ripe/ru_org_inetnum_plus_ru_as_route_v4.prefixes.txt`.
