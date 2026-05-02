# TSPU And VPN Detection Notes

This repository is not only a prefix builder. It also documents practical
signals relevant to Russian filtering and VPN-detection behavior.

The current knowledge model is:

- curated public lists under `lists/`
- explanatory docs under `docs/`
- longer investigative notes under `docs/library/`

## What The Non-RIPE Lists Capture

The non-RIPE list families exist because provider-prefix coverage alone
is not sufficient for real routing decisions.

- `lists/domains/` captures services that geo-block non-RU traffic and
  therefore should not transit a foreign proxy path
- `lists/probe/` captures public-IP and VPN-detection endpoints used by
  Russian apps and services
- `lists/whitelist/` captures a separate hostname-derived feed used as a
  practical proxy for mobile-whitelist reachability

## Research Archive

The raw source material is preserved in:

- `docs/library/habr-1009542-*`
- `docs/library/habr-1027276-*`
- `docs/library/rks-vpn-detection-2026.md`

Those files are retained for traceability, but consumer-facing guidance
should prefer the curated lists and docs first.
