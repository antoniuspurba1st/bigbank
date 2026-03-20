import { NextResponse } from "next/server";

import { transactionServiceUrl } from "@/lib/server-api";

export async function GET() {
  try {
    const upstream = await fetch(`${transactionServiceUrl()}/health`, {
      cache: "no-store",
    });
    const body = await upstream.text();

    return new NextResponse(body, {
      status: upstream.status,
      headers: {
        "content-type": "application/json",
        "x-correlation-id": upstream.headers.get("x-correlation-id") || "",
      },
    });
  } catch {
    return NextResponse.json(
      {
        status: "error",
        code: "UI_PROXY_UNAVAILABLE",
        message: "Transaction service health check failed",
        correlation_id: "ui-health-fallback",
      },
      { status: 503 },
    );
  }
}
