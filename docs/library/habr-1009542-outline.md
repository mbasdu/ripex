# Habr 1009542 — "How TSPU Detects VLESS in 2026 and Why XHTTP Is the Next Step"

- **Source:** https://habr.com/ru/articles/1009542/
- **Author:** `cyberscoper`
- **Date:** 2026-03-12
- **Fetched:** 2026-04-23

Summary, outline, and the author's recommendations. Comments and my own verification live in sibling docs.

---

## Author's thesis

Russian DPI (TSPU) runs four detection layers:

1. **Signature scan** — first 16–32 bytes vs. known patterns (catches Shadowsocks, OpenVPN)
2. **TLS fingerprint** — JA3/JA4 over ClientHello
3. **Active probing** — TSPU itself reaches out to the suspected endpoint and checks behaviour
4. **Behavioural analysis** — packet sizes, timing, direction ratios, session shape

VLESS + Reality defeats layers 1–3 (perfect TLS, real cert chain via proxy-to-real-server, right ASN if SNI chosen well). The author claims it increasingly loses to layer 4 in 2026 because real HTTPS traffic is asymmetric (downloads dominate) while proxied tunnels generate atypical flow shapes. Author's proposed fix: add **XHTTP** transport under VLESS so the tunnel looks like a normal HTTP/2 request-response sequence, not a persistent bidirectional stream.

## Outline (section → 1–3 lines)

| Section | Summary |
|---|---|
| TSPU architecture | Four layers as above. Reality covers L1–L3. L4 is the "new front". |
| JA3/JA4 | JA3 hashes cipher suites + extensions + curves. JA4 (FoxIO 2023) is more robust because it fingerprints more features. Go's default `crypto/tls` has a detectable JA3; uTLS with `fingerprint: "chrome"` mimics real Chrome. |
| Reality's limits | TLS handshake + cert validation is perfect. Post-handshake traffic shape gives it away. |
| XHTTP protocol | HTTP/2 or HTTP/3 transport for VLESS that splits upload and download into separate HTTP transactions. Modes: `packet-up`, `stream-up`, `stream-one`, `auto`. |
| Config examples | Full xray JSON for server + client (VLESS + XHTTP + Reality) are included in the article. |
| SNI selection | High-traffic, same-region donor. Good: github.com, twitch.tv, microsoft.com. Bad: apple.com (owns its own ASN). Bad: small / niche sites. |
| Verification | Use scrapfly.io JA3 checker, direct curl to your endpoint to provoke probes, check inbound ≈ outbound ratio during real use. |

## Concrete recommendations (author's list)

| # | Recommendation |
|---|---|
| R1 | Always set `"fingerprint": "chrome"` in client uTLS config; never ship with bare Go TLS defaults. |
| R2 | Use XHTTP mode `packet-up` by default — it fronts cleanly through CDN / nginx. Use `stream-one` only with XTLS-Vision. |
| R3 | SNI donor must be in the **same geographic region** as the server and share its ASN / hosting class. Hetzner-hosted server + `icloud.com` SNI = instant anomaly. |
| R4 | Listen on **TCP/443 only**. No obscure ports. |
| R5 | Generate Reality keys with `xray x25519`. Keep the private key on the server, ship the public key + shortId to clients. |
| R6 | If tunnelling XHTTP through nginx gRPC: `grpc_read_timeout 315s; grpc_send_timeout 315s; grpc_pass ...`. |
| R7 | Xray client and server versions must match exactly. Mismatch → silent failure or total breakage. |
| R8 | Set `"xPaddingBytes": "100-1000"` so packet size distribution is normalised. |
| R9 | When Reality and XHTTP are both enabled, let Xray pick the mode (it settles on `stream-one` automatically); don't force `packet-up`. |
| R10 | Things to avoid: mismatched Xray versions, SNI that doesn't match the server's real ASN, apple.com as SNI donor, small / niche SNI donors. |

## Tools & verdicts

| Tool | Verdict |
|---|---|
| Xray-core | Primary, active; essential for VLESS / XHTTP / Reality. |
| uTLS | Go library that mimics real-Chrome ClientHello byte-for-byte. Integrated into Xray. Non-optional for JA3 evasion. |
| Reality | Solid against L1–L3. Insufficient alone against L4 under load. |
| XHTTP (Xray built-in) | "Emerging primary" for 2026. HTTP-like request-response pairs; supports packet-level randomisation. |
| XTLS-Vision | Advanced option, compatible only with `stream-one` XHTTP mode. |
| JA3 / JA4 | Detection-side metrics; JA4 (FoxIO 2023) more robust than JA3. Reference checker: scrapfly.io. |

## TSPU / DPI behaviour claims (as asserted)

| Claim | Detail |
|---|---|
| L1 signature | First 16–32 bytes matched against a known-patterns DB. Shadowsocks, OpenVPN fall here. VLESS + TLS passes. |
| L2 JA3 / JA4 | Non-standard TLS stacks entered into detection DBs automatically. Raw Xray (no uTLS chrome) is flagged. |
| L3 active probe | TSPU pokes the endpoint with HTTP/HTTPS/TLS probes. Real sites respond right, proxies don't. |
| L4 behavioural | Size, timing, direction ratio, API-endpoint pattern analysed. VLESS tunnel doesn't mimic real Apple/iCloud traffic. |
| Timeline | 2022-2023 Reality effective; 2025 academic research on XTLS-Reality detection (cites SPbPU thesis); 2026 behavioural blocking increasing. |
| Failure signals | GREASE pattern mismatch; IP / ASN not matching SNI; post-handshake flow looks wrong. |

## Operational hazards

| Risk | Mitigation |
|---|---|
| JA3 leak | `fingerprint: chrome` + matched versions. Verify via scrapfly.io. |
| SNI / ASN mismatch | Pick CDN-fronted donor in the hoster's region. |
| Behavioural profiling under load | "High-load servers fall faster" because the flow shape diverges more from the declared donor. XHTTP + padding mitigates. |
| Version mismatch | No graceful error. Pin xray version on both ends. |
| CDN incompatibility | XHTTP `stream-up` / `stream-one` may fail through some reverse proxies. Prefer `packet-up` until verified. |
| Fingerprint reneging | Chrome 2024+ randomises TLS extension order. Reduces uniqueness but doesn't hide non-browser stacks. uTLS still required. |

## Comparison table (author's verdict)

| | Reality only | XHTTP only | XHTTP + Reality |
|---|---|---|---|
| TLS fingerprint | perfect | good (uTLS) | perfect |
| Active probe resilience | yes | yes | yes |
| Behavioural resilience | **weak under load** | strong | strong |
| CDN / reverse proxy support | no | yes (packet-up) | partial |
| XTLS-Vision compat | yes | stream-one only | stream-one only |
| Setup complexity | moderate | moderate | above moderate |
| 2026 outlook | declining | growing | **best available** |

## What the article does NOT say (absences worth noting)

- No specific TSPU / Tinkoff / Beeline port-based blocking claims.
- No time-of-day variation.
- No retry-limit or rate-limit specifics.
- No multi-server coordination or failover strategy.

---

*Cross-reference: `habr-1009542-comments.md` for community reaction, `habr-1009542-verification.md` for independent confirmation of claims.*
