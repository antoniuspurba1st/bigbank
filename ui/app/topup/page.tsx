"use client";

import { useState, useEffect } from "react";
import { getSession, getAuthHeaders } from "@/lib/session";
import Link from "next/link";

export default function TopupPage() {
  const [session, setSession] = useState<{ email: string } | null>(null);
  const [amount, setAmount] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [result, setResult] = useState<{
    status: "success" | "error";
    message: string;
    transactionID?: string;
  } | null>(null);

  useEffect(() => {
    setSession(getSession());
  }, []);

  const handleTopup = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!session?.email || !amount) return;

    const normalizedAmount = amount.replace(",", ".");
    const numAmount = parseFloat(normalizedAmount);
    if (isNaN(numAmount) || numAmount <= 0 || numAmount > 10000000) {
      setResult({
        status: "error",
        message: "Amount must be between $0.01 and $10,000,000",
      });
      return;
    }

    setIsLoading(true);
    setResult(null);

    try {
      const response = await fetch(`/api/topup`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeaders(),
        },
        body: JSON.stringify({ amount: numAmount }),
      });

      const data = await response.json();

      if (!response.ok) {
        setResult({
          status: "error",
          message: data.message || "Top-up failed",
        });
      } else {
        setResult({
          status: "success",
          message: `Top-up of $${numAmount.toFixed(2)} processed successfully!`,
          transactionID: data.data?.transaction_id,
        });
        setAmount("");
      }
    } catch (err) {
      setResult({
        status: "error",
        message: "Network error. Please try again.",
      });
    } finally {
      setIsLoading(false);
    }
  };

  if (!session) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-600">Loading...</p>
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-3xl font-semibold tracking-tight mb-8">Add Funds</h1>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Top-up Form */}
        <div className="glass-panel p-8 rounded-lg">
          <h2 className="text-xl font-semibold mb-6">Top-up Your Account</h2>

          <form onSubmit={handleTopup} className="space-y-6">
            <div>
              <label htmlFor="amount" className="block text-sm font-medium mb-2">
                Amount (USD)
              </label>
              <div className="relative">
                <span className="absolute left-3 top-3 text-gray-600">$</span>
                <input
                  id="amount"
                  type="number"
                  step="0.01"
                  min="0.01"
                  max="10000000"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  placeholder="0.00"
                  className="w-full pl-7 pr-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  required
                />
              </div>
              <p className="text-xs text-gray-500 mt-2">
                Maximum per transaction: $10,000,000
              </p>
            </div>

            {result && (
              <div
                className={`p-4 rounded-lg ${
                  result.status === "success"
                    ? "bg-green-50 text-green-800 border border-green-200"
                    : "bg-red-50 text-red-800 border border-red-200"
                }`}
              >
                <p className="font-medium">{result.message}</p>
                {result.transactionID && (
                  <p className="text-xs mt-2 font-mono">ID: {result.transactionID}</p>
                )}
              </div>
            )}

            <button
              type="submit"
              disabled={isLoading || !amount}
              className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-medium py-2 rounded-lg transition"
            >
              {isLoading ? "Processing..." : "Top-up Now"}
            </button>
          </form>

          <p className="text-xs text-gray-500 mt-6">
            Transaction fees: None. Your money will be credited instantly.
          </p>
        </div>

        {/* Information Panel */}
        <div className="glass-panel p-8 rounded-lg">
          <h2 className="text-xl font-semibold mb-6">How It Works</h2>

          <div className="space-y-4">
            <div className="flex gap-3">
              <div className="flex-shrink-0 w-6 h-6 rounded-full bg-blue-100 text-blue-600 flex items-center justify-center text-sm font-semibold">
                1
              </div>
              <div>
                <h3 className="font-medium">Enter Amount</h3>
                <p className="text-sm text-gray-600">
                  Enter the amount you want to add to your account
                </p>
              </div>
            </div>

            <div className="flex gap-3">
              <div className="flex-shrink-0 w-6 h-6 rounded-full bg-blue-100 text-blue-600 flex items-center justify-center text-sm font-semibold">
                2
              </div>
              <div>
                <h3 className="font-medium">Review Details</h3>
                <p className="text-sm text-gray-600">
                  Your account will be updated with the new balance
                </p>
              </div>
            </div>

            <div className="flex gap-3">
              <div className="flex-shrink-0 w-6 h-6 rounded-full bg-blue-100 text-blue-600 flex items-center justify-center text-sm font-semibold">
                3
              </div>
              <div>
                <h3 className="font-medium">Instant Credit</h3>
                <p className="text-sm text-gray-600">
                  Funds are available immediately for transfers
                </p>
              </div>
            </div>
          </div>

          <div className="mt-6 p-4 bg-blue-50 rounded-lg border border-blue-200">
            <p className="text-sm font-medium text-blue-900">Demo Mode</p>
            <p className="text-xs text-blue-800 mt-1">
              This is a demonstration. Top-ups are credited to your account for testing transfers and transactions.
            </p>
          </div>
        </div>
      </div>

      {/* Navigation Links */}
      <div className="mt-8 flex gap-4">
        <Link href="/transfer" className="text-blue-600 hover:text-blue-700">
          ← Back to Transfer
        </Link>
        <Link href="/transactions" className="text-blue-600 hover:text-blue-700">
          View Transactions →
        </Link>
      </div>
    </div>
  );
}
