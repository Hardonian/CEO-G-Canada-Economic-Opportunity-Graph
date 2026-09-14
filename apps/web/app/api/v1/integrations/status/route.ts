import { publicJSON, publicOptions } from "@/lib/public-api";

export const dynamic = "force-dynamic";

export function GET() {
  return publicJSON({
    secrets_disclosed: false,
    integrations: [
      { id: "world_bank_indicators", status: "ACTIVE", credential: "NOT_REQUIRED", configured: true },
      { id: "statistics_canada_wds", status: "REGISTERED_NOT_INGESTED", credential: "NOT_REQUIRED", configured: true },
      { id: "oecd_sdmx", status: "REGISTERED_NOT_INGESTED", credential: "NOT_REQUIRED", configured: true },
      { id: "un_comtrade", status: "REGISTERED_NOT_INGESTED", credential: "OPTIONAL_SUBSCRIPTION_KEY", configured: Boolean(process.env.UN_COMTRADE_API_KEY) },
      { id: "wto_timeseries", status: "REGISTERED_NOT_INGESTED", credential: "FREE_API_KEY_REQUIRED", configured: Boolean(process.env.WTO_API_KEY) },
      { id: "adsense_auto_ads", status: "OPTIONAL_MONETIZATION", credential: "APPROVED_PUBLISHER_ID_REQUIRED", configured: /^ca-pub-\d{16}$/.test(process.env.NEXT_PUBLIC_ADSENSE_PUBLISHER_ID ?? "") },
    ],
  }, { headers: { "Cache-Control": "private, no-store" } });
}

export const OPTIONS = publicOptions;
