# Data access and credential register

This project never commits access tokens, passwords, or subscription keys. Most
official sources in the trade and supply-chain registry are deliberately chosen
because their public interfaces require no credential.

| Publisher service | Access status | Credential / configuration |
| --- | --- | --- |
| World Bank Indicators API | Active adapter; public | No credential. Set `GLOBAL_TRADE_MODE=live` to opt into network ingestion; deterministic snapshot mode is the default. |
| Statistics Canada Web Data Service | Registered; public | No credential. |
| ISED Trade Data Online | Registered portal/download | No credential for public reports. |
| OECD Data Explorer SDMX API | Registered; public | No credential for public SDMX data. |
| IMF PortWatch Search API | Registered; public catalogue | No credential for public catalogue access. |
| UNCTADstat Data Centre | Registered portal/download | No credential for public downloads. |
| UN Comtrade public API | Registered; public tier | Public requests can operate without a key subject to publisher limits. `UN_COMTRADE_API_KEY` is reserved for a future promoted adapter; obtain any subscription credential directly from UN Comtrade. |
| WTO Timeseries API | Registered; key required | A free key must be issued to an accountable user at <https://apiportal.wto.org/>. Store it only as `WTO_API_KEY` in the deployment secret manager. |
| Google AdSense Auto ads | Optional monetization | Site approval and a public publisher ID are required. Set `NEXT_PUBLIC_ADSENSE_PUBLISHER_ID=ca-pub-################`; never store the Google account password in this project. |

## Secret handling

- Local values belong in an ignored `.env` file copied from `.env.example`.
- Production values belong in the hosting provider's encrypted environment
  variables. Scope keys to the minimum environments that need them.
- Logs, health payloads, source profiles, exported datasets, and browser bundles
  must report only `configured` / `not configured`, never a credential value.
- Rotate a key immediately if it appears in Git history, build output, logs, an
  issue, or a client-side bundle.

## AdSense activation

1. Obtain site approval and the publisher ID from AdSense.
2. Add the publisher ID as `NEXT_PUBLIC_ADSENSE_PUBLISHER_ID` in Vercel.
3. Enable Auto ads and the left/right side-rail format in AdSense.
4. Redeploy and verify `/ads.txt` returns the account declaration.
5. Ads remain consent-gated and use Google's `disablePersonalization` privacy
   treatment. A publisher remains responsible for its privacy notice, regional
   consent requirements, tax reporting, and AdSense policy compliance.

No software change can guarantee advertising income. Earnings depend on site
approval, eligible traffic, geography, advertiser demand, viewability, and
policy-compliant use.
