import { randomUUID } from "node:crypto";

const defaultApiUrl = "http://localhost:8081";

export function getApiUrl() {
  // Try NEXT_PUBLIC_API_URL first, fallback to TRANSACTION_SERVICE_URL for backwards compatibility, then default
  const url = process.env.NEXT_PUBLIC_API_URL || process.env.TRANSACTION_SERVICE_URL || defaultApiUrl;
  return url.trim().replace(/\/$/, "");
}

// Keep the old function name for backwards compatibility to prevent breaking other UI proxy code
export function transactionServiceUrl() {
  return getApiUrl();
}

export function correlationIdFromHeaders(headers: Headers) {
  return (
    headers.get("x-correlation-id")?.trim() ||
    headers.get("X-Correlation-Id")?.trim() ||
    randomUUID()
  );
}
