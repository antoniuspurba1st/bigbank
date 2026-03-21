"use client";

import { useEffect, useState } from "react";
import { getAuthHeaders } from "@/lib/session";

type DashboardSummary = {
  title: string;
  value: string;
};

type UserProfile = {
  email: string;
  account_number?: string;
  balance?: number;
};

const initialSummaryCards: DashboardSummary[] = [
  { title: "Account", value: "n/a" },
  { title: "Balance", value: "$0.00" },
  { title: "Transactions", value: "0" },
];

type TransactionItem = {
  id: string;
  reference: string;
  amount: string;
  status: string;
};

export default function Home() {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [summary, setSummary] = useState<DashboardSummary[]>(initialSummaryCards);
  const [transactions, setTransactions] = useState<TransactionItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadProfile() {
      setIsLoading(true);
      setError(null);

      try {
        const response = await fetch("/api/auth/me", {
          cache: "no-store",
          headers: {
            ...getAuthHeaders(),
          },
        });
        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.message || "Unable to load user profile");
        }

        const data = await response.json();
        const userProfile: UserProfile = {
          email: data.email,
          account_number: data.account_number,
          balance: data.balance,
        };

        setUser(userProfile);

        // Load recent transactions via API for this user.
        const txnResponse = await fetch("/api/transactions?page=0&limit=4", {
          cache: "no-store",
          headers: getAuthHeaders(),
        });

        let fetchedTransactions: TransactionItem[] = [];

        if (txnResponse.ok) {
          const txnData = await txnResponse.json();
          const items = txnData?.data?.items || [];

          fetchedTransactions = items.map((item: any) => ({
            id: item.transaction_id,
            reference: item.reference,
            amount: `$${item.amount.toFixed(2)}`,
            status: item.status,
          }));

          setTransactions(fetchedTransactions);
        }

        setSummary([
          { title: "Account", value: userProfile.account_number || "n/a" },
          { title: "Balance", value: userProfile.balance !== undefined ? `$${userProfile.balance.toFixed(2)}` : "$0.00" },
          { title: "Transactions", value: fetchedTransactions.length.toString() },
        ]);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load profile.");
      } finally {
        setIsLoading(false);
      }
    }

    loadProfile();
  }, []);

  return (
    <div>
      <h1 className="text-3xl font-semibold tracking-tight">Dashboard</h1>

      {isLoading ? (
        <p className="mt-6">Loading account info...</p>
      ) : error ? (
        <p className="mt-6 text-red-600">{error}</p>
      ) : (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-6">
            {summary.map((card) => (
              <div key={card.title} className="glass-panel p-6 rounded-lg">
                <h3 className="eyebrow">{card.title}</h3>
                <p className="text-2xl font-semibold mt-2">{card.value}</p>
              </div>
            ))}
          </div>

          <div className="mt-8">
            <h2 className="text-xl font-semibold tracking-tight">Recent Transactions</h2>
            <div className="mt-4 glass-panel rounded-lg p-6">
              {transactions.length === 0 ? (
                <p className="text-gray-600">No transactions yet. Make a top-up or transfer to get started.</p>
              ) : (
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-(--line)">
                      <th className="text-left py-2 eyebrow">Reference</th>
                      <th className="text-left py-2 eyebrow">Amount</th>
                      <th className="text-left py-2 eyebrow">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {transactions.map((txn) => (
                      <tr key={txn.id} className="border-b border-(--line)">
                        <td className="py-3 font-mono text-sm">{txn.reference}</td>
                        <td className="py-3">{txn.amount}</td>
                        <td className="py-3">
                          <span
                            className={`status-pill px-2 py-1 rounded-full text-xs ${
                              txn.status === "Completed"
                                ? "status-success"
                                : "status-rejected"
                            }`}
                          >
                            {txn.status}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
