import { NextResponse } from "next/server";

const PUBLIC_HEADERS = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Methods": "GET, OPTIONS",
  "Access-Control-Allow-Headers": "Accept, Content-Type",
  "Cache-Control": "public, max-age=0, s-maxage=300, stale-while-revalidate=86400",
  "X-Content-Type-Options": "nosniff",
} as const;

export function publicJSON(value: unknown, init: ResponseInit = {}): NextResponse {
  return NextResponse.json(value, {
    ...init,
    headers: { ...PUBLIC_HEADERS, ...init.headers },
  });
}

export function publicText(value: string, contentType: string, init: ResponseInit = {}): Response {
  return new Response(value, {
    ...init,
    headers: { ...PUBLIC_HEADERS, "Content-Type": contentType, ...init.headers },
  });
}

export function publicOptions(): Response {
  return new Response(null, { status: 204, headers: PUBLIC_HEADERS });
}
