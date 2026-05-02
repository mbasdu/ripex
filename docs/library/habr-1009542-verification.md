# Habr 1009542 — Verification of claims

Independent corroboration (or refutation) of the article's load-bearing technical claims, plus what the comments added. Each claim is scored:

- ✅ **Confirmed** — multiple independent sources agree
- ⚠️ **Partially confirmed** — core is real, details are narrower or different
- ❌ **Not supported** — no independent evidence, or contradicted
- 🆕 **New fact** — not in the article, discovered during verification and material to our architecture

---

## Article claims

### ✅ XHTTP transport exists in Xray with modes `packet-up / stream-up / stream-one / auto`

Xray's own docs and the Project X site confirm this. `packet-up` is slowest but passes CDNs and most reverse proxies. `stream-up` splits inbound and outbound into separate TCP flows (harder to correlate). `stream-one` doesn't split and only works through nginx with `grpc_pass` or Cloudflare with gRPC enabled. `auto` detects. Confirmed via Xray docs + Habr XHTTP overview + GitHub issues.

### ✅ uTLS `fingerprint: "chrome"` is required

Without it, Go's default crypto/tls has a JA3 that's trivially flagged. Confirmed in Xray transport docs and multiple third-party writeups.

### ⚠️ JA4 (FoxIO 2023) "more robust than JA3"

JA4+ exists, is real, and is documented; several guides call it the 2026 standard. Whether it is *deployed* on TSPU specifically is less clear — see the "active probing" finding below.

### ⚠️ TSPU active probing is deployed

Confirmed in the Habr ecosystem and independent reports: **widespread since August 2025**. TSPU opens its own connections to suspected endpoints and checks whether they behave like the SNI they claim to be. This covers the article's Layer 3.

### ❌ TSPU does ML-based behavioural L4 classification at scale

Article's core thesis. **Not supported** by any independent evidence I found. Commercial ML DPI engines exist (see R&S PACE 2 below), but no public evidence that TSPU has deployed them at production scale. Most real-world reports suggest TSPU uses coarser, cheaper mechanisms (throttling, IP + SNI whitelists, carpet bombing). Comment C3 (`runetfreedom`) makes exactly this argument.

### ✅ R&S PACE 2 exists and uses ML / behavioural / heuristic DPI

Real commercial engine from ipoque/Rohde & Schwarz. Documented capabilities: behavioural + statistical + heuristic + ML + DL, 14 Gbps per core, classifies "hundreds of VPN services" including over encrypted traffic. So the *category* of attack the article describes is a real product, but its deployment on TSPU is speculation.

### ✅ SNI donor must match server ASN / region

Confirmed. Multiple sources say IP + SNI combined whitelisting is the deployed model. SNI spoofing alone works only where operators don't cross-check IPs. Hetzner (AS24940) is flagged as "typically blocked".

### ⚠️ Version-matching Xray client/server "critical"

True in general — XHTTP config syntax has evolved. But the article's framing of "silent failure or total breakage" is slightly dramatic; normal version drift usually produces visible errors. Our arm/utri currently run xray 26.3.27 and sing-box 1.13.9 on wrt4, both recent, so this isn't a live hazard.

---

## New facts (not in the article; found during verification)

### 🆕 "16 KB throttle" mechanism (confirmed, critical)

Cloudflare publicly documented this. Since **June 9, 2025**, Russian ISPs throttle traffic to Cloudflare-protected sites such that each TCP connection freezes after ~15–20 KB of server→client data (10–14 packets). Affects HTTP/1.1, HTTP/2 on TLS, and HTTP/3 on QUIC. Confirmed implementers: Rostelecom, Megafon, Vimpelcom, MTS, MGTS. Sources: [cloudflare blog](https://blog.cloudflare.com/russian-internet-users-are-unable-to-access-the-open-internet/) + en.zona.media + bleepingcomputer + securityboulevard.

- **As of April 2026:** still ongoing.
- Primary targets Cloudflare IP space. Not yet observed as a general policy for all foreign IPs, but clearly the technique could be extended.
- **Matches** comment C3's "16 KB blocks" terminology.

### 🆕 uTLS Chrome fingerprint bug → 50–100% detection (critical, recently fixed)

`CVE-2025-????`: uTLS versions from **v1.6.0 (Dec 2023) to Oct 13, 2025** have a mismatch between outer ClientHello cipher suite and inner GREASE ECH extension. **Deterministic 50% detection per connection → effectively 100% for proxies** holding many connections. Fixed in uTLS commit `24bd1e05a...`.

- Xray-core 26.3.27 (Apr 2026) **is past the fix.** Good.
- Sing-box 1.13.9 (the version on wrt4) needs checking — if its vendored uTLS predates the fix, our VLESS clients are more detectable than they should be. This is a concrete investigable issue.
- Workaround if caught in an old version: switch `fingerprint` from `chrome` to `firefox` or `ios`.

### 🆕 CIDR + SNI combined whitelist (confirmed by mobile network users)

Mobile network whitelists explicitly bind domains to IP subnets. Single-server hosting in a non-whitelisted AS + perfect SNI spoofing → still blocked. The practical fix is a **dual-server setup**: a Russian-hosted front end with a whitelisted IP relays to a foreign back end. Public whitelist repo: `hxehex/russia-mobile-internet-whitelist` on GitHub.

- Matches comment C14, C15 observations.
- Our current wrt fleet lives in Russia behind Beeline/Tinkoff; the VPN endpoints (arm 130.61.21.41, utri 144.24.182.58) are on Oracle Cloud ASNs — not trivially whitelisted.

### 🆕 ECH blocking since Nov 2024

TSPU actively blocks ClientHello containing both ECH extension and SNI. Minor note for us — we're not using ECH — but worth flagging that "novelty TLS extensions" get blocked faster than proven ones.

### 🆕 AmneziaWG stopped working on ~June 6, 2025

User reports: tunnel establishes but downstream traffic = 0. Date ~matches Cloudflare's June 9 16-KB throttle start. Plausibly related: the 16-KB limiter breaks any bulk-data tunnel that doesn't masquerade as a very short HTTP transaction.

### 🆕 Reality in 2026 per independent sources

Reports are split: some say "late 2025 Roskomnadzor blocked VLESS" but the nuance is the TLS handshake is being *degraded* on specific SNI/AS combos, not Reality specifically being broken. The 99.5% success-rate claim comes from a vendor blog — take with salt. Our own measurements: wrt4 → arm VLESS+Reality was 10/10 at 20:30 UTC but utri:22 went to 0/10 in the same window. So **Reality is not universally blocked; it's selectively filtered by destination**.

---

## Article quality — meta-verification

Comments C7, C8 accuse the article of being AI-generated. Linguistic markers they cite (em-dashes, "not X, it's Y", structural repetition) are real LLM tells. Not a disqualifier, but **treat the article's specific numeric claims ("60–70% detection", "behavioral blocking increasing") as unverified narrative** — they aren't sourced.

What *is* sourced and independently verifiable: the existence of XHTTP, Reality config patterns, uTLS + fingerprint options, R&S PACE 2, TSPU's active probing and SNI/IP whitelisting. The story layer ("TSPU is deploying ML in 2026") is unsupported.

---

## Net summary for our planning

Claims we can build on:
- XHTTP + Reality is a real, current configuration; packet-up mode is the CDN-compatible default.
- SNI must match server ASN/region.
- uTLS fingerprint chrome + patched uTLS (Oct 2025+) is baseline hygiene.
- Active probing is a live threat; Reality mitigates it but only if the SNI donor is plausible.
- 16-KB Cloudflare throttle is live and expanding.
- CIDR+SNI combined whitelist on mobile networks is live — dual-server architecture is the structural answer if we need mobile-network coverage.

Claims we should NOT plan against without more evidence:
- "TSPU runs ML classifiers on flow shape." Possible but unproven; don't over-engineer.
- "Reality is dying in 2026." Narrative, not data. Our own measurements say it's selective.

---

## Sources

- [Xray-core: Transport (uTLS, REALITY)](https://xtls.github.io/en/config/transport.html)
- [Cloudflare: Russian Internet users unable to access the open Internet](https://blog.cloudflare.com/russian-internet-users-are-unable-to-access-the-open-internet/)
- [zona.media: The 16-kilobyte curtain](https://en.zona.media/article/2025/06/19/cloudflare)
- [zona.media: Russia's internet censorship in 2026](https://en.zona.media/article/2026/04/07/russian_internet_censorship_2026)
- [TSPU paper, Xue et al., ACM IMC 2022](https://dl.acm.org/doi/10.1145/3517745.3561461)
- [net4people/bbs #490 — Censor has a new method of blocking](https://github.com/net4people/bbs/issues/490)
- [net4people/bbs #516 — Mobile network website whitelist](https://github.com/net4people/bbs/issues/516)
- [hxehex/russia-mobile-internet-whitelist](https://github.com/hxehex/russia-mobile-internet-whitelist)
- [ipoque: R&S PACE 2 DPI engine](https://www.ipoque.com/products/deep-packet-inspection-for-software-vendors/dpi-engine-rs-pace-2-for-application-awareness)
- [uTLS fingerprint mismatch CVE (dailycve.com)](https://dailycve.com/utls-fingerprint-mismatch-cve-2025-medium/)
- [Xray-core issue #5230 — uTLS fingerprint leakage](https://github.com/XTLS/Xray-core/issues/5230)
- [Amnezia-client issue #1639 — No incoming traffic since June 6 2025](https://github.com/amnezia-vpn/amnezia-client/issues/1639)
- [Xeovo: Russia widespread VLESS outages 2026](https://hub.xeovo.com/posts/132-russia-widespread-vless-outages-due-to-tls-handshake-blockingdegradation-request-tlstransport-hardening-and-anti-probing)
