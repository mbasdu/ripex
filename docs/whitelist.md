# Mobile Internet Whitelist Coverage Checks

This repository includes an opt-in live test for domains curated from the five published MinTsifry whitelist expansion notices.

## Source and Limits

The current fixtures are curated from these primary-source publications:

- https://digital.gov.ru/news/vazhnye-internet-servisy-ostayutsya-dostupnymi-vo-vremya-ogranichenij
- https://digital.gov.ru/news/spisok-dostupnyh-vo-vremya-ogranichenij-raboty-mobilnogo-interneta-sajtov-dopolnen-2
- https://digital.gov.ru/news/belyj-spisok-popolnilsya-regionalnymi-onlajn-platformami
- https://digital.gov.ru/news/belyj-spisok-dostupnyh-pri-ogranicheniyah-interneta-sajtov-vnov-rasshiren
- https://digital.gov.ru/news/v-belyj-spisok-vklyucheny-novye-servisy

This is still not an official prefix feed. The fixtures map named services to representative public domains, then verify that their current IPv4 A records are covered by `ripex` output.

That means the suite checks a practical proxy, not a perfect canonical whitelist:

- it validates domains, not the entire service hostname surface
- DNS can change over time
- a service may use additional hostnames or CDNs not represented by the chosen domain
- some MinTsifry entries are broad bundles rather than a single stable public domain, so the fixtures use a representative site or omit the item when no safe proxy is available

## Running the Live Coverage Test

Build the current datasets first:

```bash
go run ./cmd/ripex build --cache-dir data/ripe/cache --out-dir data/ripe/ru
```

Then run the opt-in live test:

```bash
RIPEX_LIVE_WHITELIST=1 go test ./internal/whitelist -run TestRussiaMobileWhitelistCoverageLive -count=1
```

The live test loads `data/ripe/ru/ru_org_inetnum_plus_ru_as_route_v4.prefixes.txt`, resolves the curated domains in:

- `testdata/russia_mobile_whitelist_mintsifry_2025_09_05.json`
- `testdata/russia_mobile_whitelist_mintsifry_2025_11_14.json`
- `testdata/russia_mobile_whitelist_mintsifry_2025_12_05.json`
- `testdata/russia_mobile_whitelist_mintsifry_2025_12_18.json`
- `testdata/russia_mobile_whitelist_mintsifry_2026_02_04.json`

It fails if any resolved IPv4 address falls outside the combined minimized prefix list.
