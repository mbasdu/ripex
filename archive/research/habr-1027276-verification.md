# Habr 1027276 — Verification of claims

Independent corroboration or refutation of the article’s load-bearing claims, plus the useful additions from the comment thread. Each claim is scored:

- ✅ **Confirmed** — multiple independent sources agree
- ⚠️ **Partially confirmed** — the core is real, but the article overstates or generalizes
- ❌ **Not supported** — no independent evidence found
- 🆕 **New fact** — not in the article, but material to our planning

---

## Article claims

### ✅ White lists on mobile networks are real

Confirmed by the article itself and by independent community reporting. The article describes a drop-all mode where only approved IPs and SNIs are allowed. That matches the `net4people/bbs` issue and the community whitelist repository.

### ✅ Local wrt13 LTE test shows allowlist behavior

Tested on 2026-04-25 from `wrt13` over the LTE/Wi-Fi uplink `phy0-sta0` (`192.168.43.252/24`, gateway `192.168.43.1`). Important detail: binding only to source IP was not sufficient, because Linux still selected `wan`; valid tests used `curl --interface phy0-sta0`.

Working over LTE:

- `https://habr.com/` -> HTTP 200, ~462 KB downloaded, `178.248.237.68`
- `https://xn--h1aemc7b.xn--e1aatfr8bwa.online/` (`шипко.смешные.online`) -> HTTP 200, `158.160.210.161`
- `https://max.ru/` -> HTTP 200, ~70 KB downloaded
- `https://yandex.ru/` -> HTTP 200 after redirect path

Blocked/timed out over LTE:

- `https://habr.ru/` timed out on `178.248.233.33`
- `https://google.com/` timed out
- `https://2ip.ru/` timed out
- `https://ifconfig.me/` timed out
- ICMP to `1.1.1.1`, `8.8.8.8`, Cloudflare WARP egress IP, `utri:443`, and several non-allowed targets failed.

Interpretation: this matches whitelist behavior, but it is not a clean "all Habr works" result. `habr.com` works; `habr.ru` did not work in this test. The `шипко.смешные.online` claim is locally confirmed on wrt13's LTE path.

### ✅ The whitelist is a two-layer system: IP/CIDR plus SNI

This is the article’s central technical claim and it lines up with the issue report:

- L3 allowlist by destination IP/CIDR
- L7 allowlist by TLS SNI

The `net4people` issue explicitly reports that `yandex.ru`-hosted IPs behave differently depending on the SNI, and that non-whitelisted destinations can be blackholed or reset.

Local wrt13 tests also showed IP-level gating:

- `habr.com` SNI to its normal `178.248.237.68` passed.
- `habr.com` SNI forced to `34.160.111.145` timed out.
- `шипко.смешные.online` SNI to its normal Yandex Cloud IP `158.160.210.161` passed.
- `шипко.смешные.online` SNI forced to `34.160.111.145` timed out.

SNI-only filtering was not proven by the local test; some mismatched SNI tests still reached the origin if the destination IP was allowed, likely because the origin itself accepted/responded or because this provider path is primarily IP-gated for those cases.

### ✅ Whitelist behavior varies by operator, region, and even subnet

Strongly supported by the article and by the whitelist repo’s README. The practical consequence is that "one working SNI" is not enough to infer universal access.

### ⚠️ The whitelist corpus is community-scan based

Plausible and consistent with the linked repository, but the article’s own methodology is still a self-reported scan pipeline. The data is useful; it is not an independent audit.

### ⚠️ Cable / home internet will not be affected soon

This is not established by the article. The thread only shows a technical possibility plus a political guess. Treat it as speculation, not a guarantee.

---

## Comment-thread additions

### ✅ White-list-aware operational strategy

The `forc3meat` comment adds a concrete operator strategy that matches real-world censorship behavior:

- multihop from Russia to Europe
- split tunneling for domestic services
- keep the front door on common ports like `80/443`
- prefer transports that look like ordinary HTTPS traffic

This is practical field advice, but it remains anecdotal.

### ⚠️ Protocol recommendations are anecdotal, not measured

`cloak`, `trojan`, and `hysteria` are widely used tools, but the comment is still a practitioner report, not a controlled benchmark.

---

## External sources

- Article source: [Habr article](https://habr.com/ru/articles/1027276/)
- Important comment: [Habr comment thread](https://habr.com/ru/articles/1027276/comments/#comment_29883166)
- Article repo: [openlibrecommunity/twl](https://github.com/openlibrecommunity/twl)
- Community whitelist repo: [hxehex/russia-mobile-internet-whitelist](https://github.com/hxehex/russia-mobile-internet-whitelist)
- Field report issue: [net4people/bbs #516](https://github.com/net4people/bbs/issues/516)
- Source thread cited by the issue: [ntc.party/t/16325/182](https://ntc.party/t/16325/182)

---

## Net summary

The article’s strongest claim is confirmed: Russian mobile networks can and do run allowlist-style filtering. Local wrt13 LTE tests confirm the practical pattern: selected whitelisted Russian/Yandex-hosted domains work, while unrelated destinations time out. The local evidence strongly confirms IP allowlisting and the `шипко.смешные.online` example; it does not prove a universal SNI+IP rule for every allowed IP, and it specifically found `habr.ru` failing while `habr.com` worked. The main thing not proven here is the article’s implied generalization to other access types such as home cable internet.

## Method 3: Yandex API Gateway reverse proxy

Tested on 2026-04-25 with the deployed API Gateway
`d5dr0ggdagt1gj1kn7s0.nkhmighe.apigw.yandexcloud.net`.

Result: confirmed for fixed reverse-proxy routes.

- Gateway root was reachable from `wrt13` over `phy0-sta0`.
- `/ifconfig/ip` returned Yandex egress IP `84.201.183.12`, while direct LTE
  to `https://ifconfig.me/ip` timed out.
- `/cloudflare/cdn-cgi/trace` returned Yandex egress IP `84.201.183.60`,
  `loc=RU`, `warp=off`, while direct LTE to the same Cloudflare URL timed out.
- `/2ip` returned `200` with Yandex egress IP `84.201.183.81`, while direct LTE
  to `https://2ip.ru` timed out.

Operational caveat: Yandex API Gateway `http` integration is a reverse proxy to
preconfigured upstream URLs. It is not a generic SOCKS/CONNECT tunnel and cannot
be used directly as a stock sing-box/xray outbound.

## Retest on 2026-04-27

Retested both the API Gateway and Cloud Function Fetch proxy from `wrt13`:

1.  **API Gateway:** Failed over the LTE whitelist path (`phy0-sta0` connection timed out) but succeeded over the primary WAN. This confirms that specific API Gateway domains can be blocked or lose their whitelisted status on the mobile network.
2.  **Cloud Function Fetch Proxy:** Succeeded over the LTE whitelist path. Successfully fetched non-RU sites (`ifconfig.me`, `google.com`, `mail.google.com`) using an authenticated GET request, confirming Yandex Cloud Function endpoints are currently accessible through the LTE drop-all filter.
    *   **Limitation:** Cloud Functions are strictly stateless request/response processors. They cannot rewrite complex SPA links (like Gmail) dynamically, and they do not support persistent TCP/WebSocket connections. Therefore, they are useful as a simple programmatic fetch proxy but cannot be used as a backend for standard VPN tools (like `sing-box` or `xray`) or normal browser navigation.
    *   **Next Step Hypothesis:** To achieve a full bidirectional proxy (SOCKS5/TUN) through the whitelisted Yandex IP space, we should investigate **Yandex Serverless Containers**. These allow standard Docker images (like `xray` or `sing-box`) to be deployed and support HTTP/1.1 Upgrades to WebSockets, which can facilitate a true, long-lived proxy connection over standard HTTPS ports.

## openlibrecommunity/twl Repository Analysis

Analyzed on 2026-04-27. The `openlibrecommunity/twl` (Total Whitelist) project is a key community resource documenting the boundaries of the Russian "sovereign internet" drop-all/whitelist mode.

### Methodology
The project relies on empirical probing during throttling/shutdown events, analyzing DPI (TSPU) behavior, and aggregating known approved subnets and domains across major mobile operators (MTS, Beeline, Megafon, Tele2).

### Key Findings & Mechanics
- **TSPU Enforcement:** The whitelist is enforced by TSPU devices at ISP gateways. Whitelisted traffic bypasses restrictive DPI checks.
- **Allowed Categories:**
  - State Services (`*.gosuslugi.ru`, `*.gov.ru`)
  - Russian Tech Giants (Yandex, VK, Mail.ru)
  - Financial Infrastructure (Sber, VTB, Mir)
  - Selected CDNs (Akamai, G-Core, local CDNs) hosting "essential" content.
  - App Store Infrastructure (Apple/Google update servers to prevent device bricking).

### Operational Insights for Bypassing
- **IP Tunneling:** Hosting proxy endpoints within whitelisted IP ranges (especially CDN ranges or Yandex Cloud, as proven by the `wrt13` experiments) is highly effective.
- **SNI Fronting:** Using whitelisted domains in the SNI header can sometimes bypass DPI, though this is actively countered by ECH blocking and IP-SNI correlation.
- **Protocol Mimicry:** Shaping traffic to mimic allowed services (e.g., VK Video, Yandex Disk) helps evade throttling.

This analysis heavily reinforces our local findings on `wrt13` regarding Yandex Cloud's whitelisted status and supports the strategy of using Yandex infrastructure (like Serverless Containers) for tunneling.
