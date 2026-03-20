import { NextRequest, NextResponse } from "next/server";

import { correlationIdFromHeaders, transactionServiceUrl } from "@/lib/server-api";
import { ApiErrorEnvelope } from "@/lib/types";

export async function GET(request: NextRequest) {
  const correlationId = correlationIdFromHeaders(request.headers);

  try {
    const page = request.nextUrl.searchParams.get("page") || "0";
    const limit = request.nextUrl.searchParams.get("limit") || "10";

    const upstream = await fetch(
      `${transactionServiceUrl()}/transactions?page=${encodeURIComponent(page)}&limit=${encodeURIComponent(limit)}`,
      {
        method: "GET",
        headers: {
          "x-correlation-id": correlationId,
        },
        cache: "no-store",
      },
    );

    const body = await upstream.text();
    return new NextResponse(body, {
      status: upstream.status,
      headers: {
        "content-type": "application/json",
        "x-correlation-id":
          upstream.headers.get("x-correlation-id") || correlationId,
      },
    });
  } catch (error) {
    const fallback: ApiErrorEnvelope = {
      status: "error",
      code: "UI_PROXY_UNAVAILABLE",
      message:
        error instanceof Error
          ? error.message
          : "UI proxy could not reach the transaction service",
      correlation_id: correlationId,
    };

    return NextResponse.json(fallback, {
      status: 503,
      headers: { "x-correlation-id": correlationId },
    });
  }
}
