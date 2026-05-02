# Habr 1027276 — "Это — все что вам надо знать о белых списках: ресерч, сканы, обход"

- **Source:** https://habr.com/ru/articles/1027276/
- **Author:** `zarazaexe`
- **Date:** 2026-04-25
- **Fetched:** 2026-04-25

Summary, outline, and the author’s main claims. Comments and verification live in sibling docs.

---

## Core thesis

The article argues that Russian mobile-network "white lists" are a real and active censorship mode:

1. **L3 allowlist** blocks traffic by destination IP/CIDR before DPI sees it.
2. **L7 allowlist** then checks TLS SNI and resets connections that are not explicitly approved.
3. The practical whitelist is not uniform: it varies by operator, region, and even subnet.

The author says the findings come from scans on MegaFon mobile internet and publishes the code/data in a companion repository.

## Outline (section → short summary)

| Section | Summary |
|---|---|
| Intro | White lists are presented as a drop-all mode with only approved IP + SNI combinations allowed. |
| L3 filtering | Non-whitelisted destinations do not pass at all; even ICMP and normal TCP can be blackholed. |
| L7 filtering | On approved IPs, SNI decides whether TLS is allowed, reset, or inconsistently ignored. |
| Scan method | The article says the whitelist corpus was built by scanning live networks and storing the data in a public repo. |
| Inconsistency across networks | The same SNI can pass on one provider / subnet and fail on another. |
| Practical takeaway | For users, "find a whitelisted IP" or use an architecture that exits through a whitelisted Russian front. |

## Important comment

The linked comment from `forc3meat` adds a practitioner view:

- Multihop is recommended: `client -> vps rf -> vps eu -> vpn`.
- Split tunnel is important so domestic traffic stays domestic.
- The comment recommends `cloak`, `trojan`, and `hysteria`; it warns against `shadowsocks/outline`, `obfs4proxy`, `vless`, and says `amnezia` appears detectable.
- It also says port choice matters and suggests sticking to `80/443`.

## What the article does not settle

- Whether cable / home internet will follow the same pattern is left open.
- How universal the whitelist corpus is across ISPs is not proven by the article alone.
- The article presents the scanning results, but not a full independent methodology audit.

## External sources worth keeping

- Article repo: [openlibrecommunity/twl](https://github.com/openlibrecommunity/twl)
- Community whitelist repo: [hxehex/russia-mobile-internet-whitelist](https://github.com/hxehex/russia-mobile-internet-whitelist)
- Related field-report issue: [net4people/bbs #516](https://github.com/net4people/bbs/issues/516)
- Source thread referenced by the issue: [ntc.party/t/16325/182](https://ntc.party/t/16325/182)

