"use client";

import { useEffect, useState } from "react";

type Transaction = {
  id: string;
  reference: string;
  amount: number;
  status: string;
  created_at: string;
};

export default function TransactionsPage() {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchTransactions = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await fetch("/api/transactions");
      if (!response.ok) {
        throw new Error("Failed to fetch transactions");
      }
      const data = await response.json();
      const items = data?.data?.items || [];
      setTransactions(items.map((item: any) => ({
        id: item.transaction_id,
        reference: item.reference,
        amount: item.amount,
        status: item.status,
        created_at: item.created_at,
      })));
    } catch (err: any) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchTransactions();
  }, []);

  return (
    <div>
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-semibold tracking-tight">Transactions</h1>
        <button
          onClick={fetchTransactions}
          disabled={isLoading}
          className="rounded-full bg-white border border-(--line) px-5 py-2 text-sm font-medium text-slate-900 transition hover:bg-gray-50 disabled:opacity-50"
        >
          {isLoading ? "Refreshing..." : "Refresh"}
        </button>
      </div>
      <div className="mt-6 glass-panel rounded-lg p-6">
        {isLoading && <p>Loading transactions...</p>}
        {error && <p className="text-red-500">{error}</p>}
        {!isLoading && !error && transactions.length === 0 && (
          <p>No transactions found.</p>
        )}
        {!isLoading && !error && transactions.length > 0 && (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-(--line)">
                  <th className="text-left py-2 eyebrow">Reference</th>
                  <th className="text-left py-2 eyebrow">Amount</th>
                  <th className="text-left py-2 eyebrow">Timestamp</th>
                </tr>
              </thead>
              <tbody>
                {transactions.map((txn) => (
                  <tr key={txn.id} className="border-b border-(--line)">
                    <td className="py-3 font-mono text-sm">{txn.reference}</td>
                    <td className="py-3">
                      {new Intl.NumberFormat("en-US", {
                        style: "currency",
                        currency: "USD",
                      }).format(txn.amount)}
                    </td>
                    <td className="py-3 font-mono text-sm">
                      {new Date(txn.created_at).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
