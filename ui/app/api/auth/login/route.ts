import { NextRequest, NextResponse } from "next/server";
import { correlationIdFromHeaders, transactionServiceUrl } from "@/lib/server-api";

export async function POST(request: NextRequest) {
  const correlationId = correlationIdFromHeaders(request.headers);

  try {
    const payload = await request.json();

    const upstream = await fetch(`${transactionServiceUrl()}/auth/login`, {
      method: "POST",
      headers: {
        "content-type": "application/json",
        "x-correlation-id": correlationId,
      },
      body: JSON.stringify(payload),
      cache: "no-store",
    });

    const body = await upstream.text();

    return new NextResponse(body, {
      status: upstream.status,
      headers: {
        "content-type": "application/json",
        "x-correlation-id": upstream.headers.get("x-correlation-id") || correlationId,
      },
    });
  } catch (error) {
    return NextResponse.json(
      {
        status: "error",
        code: "UI_PROXY_UNAVAILABLE",
        message: error instanceof Error ? error.message : "Unable to reach backend auth/login",
        correlation_id: correlationId,
      },
      {
        status: 503,
        headers: { "x-correlation-id": correlationId },
      }
    );
  }
}
