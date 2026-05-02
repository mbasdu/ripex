# Habr 1009542 — Signal comments

Filtered for signal: kept [DISPUTES], [ADDS_EVIDENCE], [NUANCES], [COUNTER_RECOMMENDATION], [USER_REPORT]. Dropped "+1", jokes, insults, pure meta.

Translations are condensed paraphrases, not literal.

---

## TSPU detection — what's actually happening

### C3 — `runetfreedom` | [DISPUTES] [COUNTER_RECOMMENDATION]
Heavy critique of the article's premise. Main points:
- Academic papers that claim to detect Reality use outdated Xray, naive configs with no padding / multiplexing / obfuscation.
- Reported 60–70% detection probabilities would trigger unacceptable false-positive rates if actually deployed, so they aren't.
- TSPU lacks the compute budget for ML-based per-flow classification; even China / Iran (much bigger budgets) haven't productised what a graduate-student lab paper claims.
- Real, present-day threats are **16 KB blocks** and **ISP-level carpet-bombing**, not sophisticated ML.

### C2 — `cyberscoper` (author response to C1) | [NUANCES] [ADDS_EVIDENCE]
Defends: lab detection used known parameters; a properly-configured Reality defeats that method in real life. Name-drops R&S PACE 2 (commercial DPI engine) using kNN / decision trees / CNN / LSTM and weekly signature updates as evidence that the class of attack exists commercially.

### C5 — `zakker` | [ADDS_EVIDENCE]
Shadowsocks detection is based on entropy + early-packet-size distribution. Tested a custom SS variant: first data packet never reaches the server — i.e., TSPU intercepts specifically the first payload-carrying packet, not the handshake.

### C13 — `Z55` | [USER_REPORT] [NUANCES]
WG blocked by one ISP, Outline by another, Xray works everywhere. Rules centralised but with per-ISP customisation.

---

## White SNI lists (important; confirmed by multiple users)

### C14 — `eskars` | [CONFIRMS]
Real. Specific domains bypass 16 KB blocks even on aggressive providers.

### C15 — `Niter43` | [NUANCES]
List is 99.9% bound to CIDR / ASN. Hetzner has hundreds of whitelisted domains (author mentions `pdf24.org` as one). Small hosters have fewer / none.

### C4 — `Crixetro` | [USER_REPORT] [NUANCES]
VLESS + XHTTP helps only if the SNI is on the white list; under 16 KB blocks it's useless. No speed difference vs VLESS + Reality.

---

## Speed / protocol comparisons

### C10 — `InspectorCat` | [USER_REPORT]
Posted benchmarks: no noticeable speed difference after xHTTP migration. (Screenshot in thread.)

### C9 — `Kenya-West` | [NUANCES]
Three questions:
1. Why masquerade as a major site vs self-host a plausible donor?
2. Does xHTTP help if you're on a blacklisted AS? (implied: probably not)
3. Seen reports of xHTTP giving ~50% throughput of VLESS + Reality + TCP.

---

## Operational experience with other protocols

### C11 — `Negat1v9` | [USER_REPORT]
Outline: 320 days uptime, 1 TB+/month, 10+ users, stable, mostly Moscow clients.

### C12 — `sloww` | [COUNTER_REPORT to C11]
Outline detected and blocked ~2 years on mobile, ~1 year on most wired ISPs. Plain SS detected instantly.

### C16 — `Barnaby` | [COUNTER_RECOMMENDATION]
"If SSH is blocked, nothing else works anyway" — fall back to SSH-SOCKS.

### C17 — `MiracleUsr` | [ADDS_EVIDENCE]
Counter: some hosters throttle / cut SSH-SOCKS above traffic thresholds.

### C18 — `Aelliari` | [NUANCES]
SSH-terminal and SSH-carrying-SOCKS have distinguishable traffic shapes (volume / direction patterns). DPI can tell the difference.

### C19 — `NickyX3` | [USER_REPORT]
Runs a custom VPN-over-SSH implementation; it's stable. Disputes claim (C18) that SSH + SOCKS and file-transfer patterns differ meaningfully.

---

## Article-quality criticism (meta, but worth noting)

### C7 — `Destructive` | [DISPUTES] [meta]
Flags AI-generated content markers: suspicious diagrams with empty details, artificial heading structure ("Why WebSocket died"), unnatural repetition, excessive em-dashes, excessive quoted phrases, choppy sentences.

### C8 — `atomlib` | [DISPUTES] [meta]
Extended linguistic analysis of LLM patterns: "this is not X, it's Y" constructions, structural reuse, overuse of em-dashes, clichéd quotation marks, rhythmic choppiness. Humans vary sentence structure more.

### C6 — `0ka` | [DISPUTES]
Points at an earlier detailed critical comment that was apparently downvoted-hidden. Says article is "absolute nonsense disconnected from reality".

### C1 — `Alexinthecold` | [DISPUTES]
Opener: "Wasn't the XTLS-Reality detection paper shown to be garbage? Where's the evidence for ML at RKN?"

---

## Topic matrix

| Topic | Confirm | Dispute / Counter | User report |
|---|---|---|---|
| TSPU L4 behavioural detection | C2 (with caveat) | C3 | — |
| 16 KB block is the real threat | — | — | C3, C4 |
| White SNI lists | C14, C4 | — | C15 |
| SNI bound to ASN / CIDR | — | — | C15 |
| XHTTP speed ≈ Reality | C10 | C9 (50% claim) | C4 |
| Outline viability | C11 | C12 | C13 |
| SSH-SOCKS as fallback | C16 | C17 (throttling), C18 (detectable) | C19 |
| Article is AI-generated | C7, C8 | — | C6 |

---

## What this means for us (raw notes, no conclusions yet)

- The article's "switch to XHTTP" thesis is loudly contested in the thread. The most substantive critic (`runetfreedom`, C3) argues TSPU isn't doing the ML the article claims; real threats are simpler and different.
- **White SNI lists + per-ASN binding** is a concrete observation multiple users confirm and it's actionable.
- **16 KB blocks** is a technical term introduced in the comments that isn't in the main article — needs investigation.
- XHTTP may not buy us anything measurable on actual TSPU today; it may buy us future-proofing.
- User reports wildly disagree on Outline / WG blocking status → regional / ISP variation is the dominant factor, not protocol choice.

---

*Note: the article itself is plausibly AI-generated per C7 / C8. Treat its specific numeric claims ("60–70% detection") with suspicion until independently verified in the sibling `verification.md`.*
