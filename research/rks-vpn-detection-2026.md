# RKS Global — "How and Why Russian Apps Search for VPN on Users' Phones"

- **Source:** https://rks.global/files/research/russian_apps_search_for_vpn_en.pdf
- **Date:** April 2026
- **Fetched:** 2026-05-01

Investigation into VPN detection mechanisms in popular Russian Android applications.

---

## Executive Summary

The report analyzed 30 popular Russian Android applications and found that 22 of them (73%) contain VPN detection logic, with 19 of those apps actively transmitting the user's VPN status to their backend servers.

## Key Findings

- **Widespread Detection:** Nearly 3/4 of analyzed apps scan for VPNs.
- **Aggressive Reporting:** 63% of apps report VPN status to home servers.
- **Deep Scanning:** Some apps use SDKs (like AppTracer) for deep device scanning.
- **Ecosystem Patterns:** VK and Yandex ecosystems are particularly thorough in their detection and reporting.

## Extracted Hosts and IP Detection Endpoints

These domains and IP addresses are associated with VPN detection, IP verification, and status reporting as identified in the RKS Global research.

### **Hostnames Used for VPN/IP Detection**
- `checkip.amazonaws.com`
- `ifconfig.co`
- `ifconfig.me`
- `icanhazip.com`
- `ident.me`
- `ipecho.net`
- `ipinfo.io`
- `myip.dnsomatic.com`
- `api.ipify.org`
- `api.myip.com`
- `ip-api.com`
- `whoer.net`
- `2ip.ru`
- `vpn.fail`
- `dns.google`
- `cloudflare-dns.com`
- `dns.yandex.ru`

### **Hostnames Used for Reporting (Trackers/Analytics)**
- `mc.yandex.ru`
- `clck.yandex.ru`
- `an.yandex.ru`
- `suggest.yandex.ru`
- `yandex.ru`
- `yandex.net`
- `yandex.com`
- `tracker-api.vk-analytics.ru`
- `sdk-api.apptracer.ru`
- `api.oneme.ru` / `i.oneme.ru`
- `portal-sentry-v2.vk.team`
- `app-measurement.com`
- `google-analytics.com`
- `firebase-settings.crashlytics.com`
- `api.segment.io`
- `graph.facebook.com`
- `vk.com`
- `mail.ru`
- `ok.ru`

### **IP Addresses (DNS Servers & Service Endpoints)**
- `1.1.1.1` (Cloudflare DNS)
- `1.0.0.1` (Cloudflare DNS)
- `8.8.8.8` (Google DNS)
- `8.8.4.4` (Google DNS)
- `9.9.9.9` (Quad9 DNS)
- `149.112.112.112` (Quad9 DNS)
- `208.67.222.222` (OpenDNS)
- `208.67.220.220` (OpenDNS)
- `77.88.8.8` (Yandex DNS)
- `77.88.8.1` (Yandex DNS)
- `213.180.204.62` (Yandex/Tracker IP)
- `87.250.250.119` (Yandex IP)
- `93.158.134.119` (Yandex IP)
- `213.180.193.119` (Yandex IP)
- `77.88.21.119` (Yandex IP)
- `77.88.21.125` (Yandex IP)
- `5.255.255.5` (Yandex IP)
- `5.255.255.50` (Yandex IP)
- `5.255.255.55` (Yandex IP)
- `77.88.55.55` (Yandex IP)
- `77.88.55.60` (Yandex IP)
- `77.88.55.66` (Yandex IP)
- `77.88.55.70` (Yandex IP)
- `77.88.55.77` (Yandex IP)
- `77.88.55.80` (Yandex IP)
- `77.88.55.88` (Yandex IP)
- `95.161.224.68` (Detection-related IP)
- `185.178.208.197` through `185.178.208.210` (Detection-related IP range)

### **IPv6 Addresses (DNS Servers)**
- `2606:4700:4700::1111`
- `2606:4700:4700::1001`
- `2001:4860:4860::8888`
- `2001:4860:4860::8844`
- `2620:fe::fe`
- `2620:fe::9`
- `2a02:6b8::feed:0ff`
- `2a02:6b8:0:1::feed:0ff`

## Detection Methods Identified

1. **Network Interface Scanning:** Checking for `tun0`, `ppp0`, `tap0`, `pptp0` adapters via `NetworkInterface.getNetworkInterfaces()`.
2. **TCP Table Inspection:** Reading `/proc/net/tcp` to find ports used by reverse engineering tools (e.g., `27042` for Frida).
3. **External IP Comparison:** Querying public IP services and comparing with local network state.
4. **App Inventory Scanning:** Scanning for presence of foreign banks, social networks, and competitors (noted in Avito).
5. **Tor Detection:** Specifically noted in Yandex Browser.

## Countermeasures

- **Isolation:** Using Work Profiles (Shelter/Island) to prevent apps from seeing the full device state.
- **Domain Blocking:** Blocking identified analytics and reporting endpoints (e.g., `vk-analytics.ru`, `apptracer.ru`) via DNS or local firewalls.
- **Split Tunneling:** Ensuring domestic traffic does not traverse the VPN, making it harder for simple IP comparison checks to trigger.
