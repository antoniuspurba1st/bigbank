import { NextRequest, NextResponse } from "next/server";
import { correlationIdFromHeaders, transactionServiceUrl } from "@/lib/server-api";

export async function GET(request: NextRequest) {
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
      },
      {
        status: 401,
        headers: { "x-correlation-id": correlationId },
      }
    );
  }

  try {
    const upstream = await fetch(`${transactionServiceUrl()}/auth/me`, {
      method: "GET",
      headers: {
        "x-user-email": userEmail,
        "x-user-id": userId,
        "x-correlation-id": correlationId,
      },
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
        message: error instanceof Error ? error.message : "Unable to reach backend auth/me",
        correlation_id: correlationId,
      },
      {
        status: 503,
        headers: { "x-correlation-id": correlationId },
      }
    );
  }
}
