# Habr 1027276 — Signal comments

Filtered for signal. Kept comments that add evidence, operational detail, or meaningful disagreement; dropped short cheers and low-signal reactions.

---

## Operational guidance from practitioners

### `forc3meat` | [ADDS_EVIDENCE] [COUNTER_RECOMMENDATION]
This is the most useful comment in the thread.

- Reports successful VPN operation through white lists in a corporate setting with users in Russia and servers in Europe.
- Recommends `cloak`, `trojan`, `hysteria`, and `vk-turn-proxy`.
- Says `shadowsocks/outline`, `obfs4proxy`, and `vless` are being detected or blocked more often.
- Recommends multihop through a Russian exit first, then a foreign exit, then the VPN.
- Advises split tunneling and sticking to common ports like `80/443`.

### `myswordishatred` | [NUANCES]
Asks the key architectural question: why cable / home internet would be harder to whitelist than mobile internet. The author answers only with a political / timing claim, not a technical proof.

### `izogfif` | [USER_REPORT]
Reports a real outage / repair window where only white-listed sites worked for a few hours after a backbone failure, which suggests the whitelist mode is not purely theoretical.

### `Espr1tDeCorps` | [USER_REPORT]
Says that after recent Cloudflare tightening, most sites are effectively inaccessible without a VPN. This is a broader censorship observation, not a white-list-specific measurement.

### `K0styan` | [NUANCES]
Gives the most sober reply to the cable-internet question: technically possible, politically uncertain.

---

## Topic matrix

| Topic | Evidence | Notes |
|---|---|---|
| Mobile white lists are real | `forc3meat`, `izogfif`, issue #516 | Strongest combined signal. |
| Multihop is useful | `forc3meat` | Practitioner guidance, not a formal proof. |
| `cloak` / `trojan` / `hysteria` as working options | `forc3meat` | Anecdotal but concrete. |
| `shadowsocks` / `outline` / `obfs4proxy` being detectable | `forc3meat` | Anecdotal; should be treated as field feedback. |
| Cable white lists are inevitable | None | The thread does not prove that. |

