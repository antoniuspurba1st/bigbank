export type ApiStatus = "success" | "error" | "rejected";

export type ApiEnvelope<T> = {
  status: ApiStatus;
  message: string;
  correlation_id: string;
  data?: T | null;
};

export type ApiErrorEnvelope = {
  status: "error";
  code: string;
  message: string;
  correlation_id: string;
};

export type TransferRequest = {
  reference: string;
  from_account: string;
  to_account: string;
  amount: number;
};

export type TransferResult = {
  reference: string;
  amount: number;
  fraud_decision: string;
  ledger_status?: string;
  transaction_id?: string;
  duplicate: boolean;
  created_at?: string;
};

export type TransactionHistoryItem = {
  transaction_id: string;
  reference: string;
  from_account: string;
  to_account: string;
  amount: number;
  status: string;
  created_at: string;
};

export type TransactionHistoryPage = {
  items: TransactionHistoryItem[];
  page: number;
  limit: number;
  total_items: number;
  total_pages: number;
  has_next: boolean;
  has_previous: boolean;
};

export function isApiErrorEnvelope(value: unknown): value is ApiErrorEnvelope {
  if (!value || typeof value !== "object") {
    return false;
  }

  const candidate = value as Record<string, unknown>;
  return (
    candidate.status === "error" &&
    typeof candidate.code === "string" &&
    typeof candidate.message === "string" &&
    typeof candidate.correlation_id === "string"
  );
}
