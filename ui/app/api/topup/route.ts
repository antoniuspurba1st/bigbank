import { NextRequest, NextResponse } from "next/server";

import { correlationIdFromHeaders, transactionServiceUrl } from "@/lib/server-api";
import { ApiErrorEnvelope } from "@/lib/types";

export async function POST(request: NextRequest) {
  const correlationId = correlationIdFromHeaders(request.headers);
  const userEmail = request.headers.get("x-user-email") || "";
  const userId = request.headers.get("x-user-id") || "";

  if (!userEmail && !userId) {
    return NextResponse.json(
      {
        status: "error",
        code: "UNAUTHORIZED",
        message: "User header missing",
        correlation_id: correlationId,
      } satisfies ApiErrorEnvelope,
      {
        status: 401,
        headers: { "x-correlation-id": correlationId },
      }
    );
  }

  try {
    const payload = await request.json();
    const upstream = await fetch(`${transactionServiceUrl()}/topup`, {
      method: "POST",
      headers: {
        "content-type": "application/json",
        "x-correlation-id": correlationId,
        "x-user-email": userEmail,
        "x-user-id": userId,
      },
      body: JSON.stringify(payload),
      cache: "no-store",
    });

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
