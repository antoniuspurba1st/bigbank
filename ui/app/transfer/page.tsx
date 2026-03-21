"use client";

import { useState, useEffect } from "react";
import { transactionServiceUrl } from "@/lib/server-api";
import { getSession, getAuthHeaders } from "@/lib/session";

type TransferResult = {
  status: "success" | "rejected" | "error";
  message: string;
  data?: any;
};

export default function TransferPage() {
  const [fromAccount, setFromAccount] = useState("ACC-001");
  const [toAccount, setToAccount] = useState("ACC-002");
  const [userAccount, setUserAccount] = useState("");
  const [amount, setAmount] = useState("");
  const [reference, setReference] = useState("tx-ui-");

  const [isLoading, setIsLoading] = useState(false);
  const [result, setResult] = useState<TransferResult | null>(null);
  
  // Confirmation state
  const [showConfirmation, setShowConfirmation] = useState(false);

  useEffect(() => {
    let active = true;

    async function fetchProfile() {
      try {
        const response = await fetch("/api/auth/me", {
          cache: "no-store",
          headers: getAuthHeaders(),
        });
        if (!response.ok) return;

        const data = await response.json();
        if (active && data.account_number) {
          setUserAccount(data.account_number);
          setFromAccount(data.account_number);
        }
      } catch (error) {
        // ignore, keep defaults
      }
    }

    fetchProfile();

    return () => {
      active = false;
    };
  }, []);

  const handleInitialSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!amount || !reference) {
      setResult({ status: "error", message: "Amount and Reference are required." });
      return;
    }

    const parsedAmount = parseFloat(amount);
    if (isNaN(parsedAmount) || parsedAmount <= 0) {
      setResult({ status: "error", message: "Please enter a valid amount greater than zero." });
      return;
    }
    
    // Show confirmation before proceeding
    setShowConfirmation(true);
    setResult(null);
  };

  const handleConfirmedSubmit = async () => {
    setShowConfirmation(false);
    setIsLoading(true);
    setResult(null);

    const parsedAmount = parseFloat(amount);
    
    const requestFromAccount = userAccount || fromAccount;

    const transferRequest = {
      from_account: requestFromAccount,
      to_account: toAccount,
      amount: parsedAmount,
      reference: `${reference}${Date.now()}`,
    };

    try {
      const session = getSession();
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
      };
      
      if (session) {
        headers["X-User-Email"] = session.email;
      }

      const response = await fetch("/api/transfer", {
        method: "POST",
        headers,
        body: JSON.stringify(transferRequest),
      });

      const data = await response.json();

      if (!response.ok) {
        setResult({
          status: "rejected",
          message: data.message || "The transfer could not be completed.",
          data,
        });
      } else {
        setResult({
          status: "success",
          message: `Successfully transferred ${parsedAmount} to ${toAccount}.`,
          data,
        });
        
        // Optionally reset amount and reference on success
        setAmount("");
      }
    } catch (error) {
      setResult({
        status: "error",
        message: "An unexpected network error occurred while processing the transfer. Please try again later.",
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="relative">
      <h1 className="text-3xl font-semibold tracking-tight">Transfer</h1>
      <div className="mt-6 grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div className="glass-panel p-6 rounded-lg">
          <form onSubmit={handleInitialSubmit}>
            <div className="space-y-4">
              <div>
                <label htmlFor="fromAccount" className="eyebrow block mb-2">
                  From Account
                </label>
                <input
                  id="fromAccount"
                  type="text"
                  value={fromAccount}
                  onChange={(e) => setFromAccount(e.target.value)}
                  className="w-full p-2 border border-(--line) rounded-md bg-white/50 focus:outline-none focus:ring-2 focus:ring-(--accent)"
                />
              </div>
              <div>
                <label htmlFor="toAccount" className="eyebrow block mb-2">
                  To Account
                </label>
                <input
                  id="toAccount"
                  type="text"
                  value={toAccount}
                  onChange={(e) => setToAccount(e.target.value)}
                  className="w-full p-2 border border-(--line) rounded-md bg-white/50 focus:outline-none focus:ring-2 focus:ring-(--accent)"
                />
              </div>
              <div>
                <label htmlFor="amount" className="eyebrow block mb-2">
                  Amount
                </label>
                <input
                  id="amount"
                  type="number"
                  step="0.01"
                  min="0.01"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  className="w-full p-2 border border-(--line) rounded-md bg-white/50 focus:outline-none focus:ring-2 focus:ring-(--accent)"
                  placeholder="0.00"
                  required
                />
              </div>
              <div>
                <label htmlFor="reference" className="eyebrow block mb-2">
                  Reference
                </label>
                <input
                  id="reference"
                  type="text"
                  value={reference}
                  onChange={(e) => setReference(e.target.value)}
                  className="w-full p-2 border border-(--line) rounded-md bg-white/50 focus:outline-none focus:ring-2 focus:ring-(--accent)"
                  placeholder="e.g., ref-001"
                  required
                />
              </div>
            </div>
            
            <div className="mt-8">
              <button
                type="submit"
                disabled={isLoading}
                className="w-full rounded-full bg-(--accent) px-5 py-3 text-sm font-medium text-white transition hover:bg-(--accent-strong) disabled:bg-slate-300 disabled:cursor-not-allowed flex items-center justify-center gap-2"
              >
                {isLoading ? (
                  <>
                    <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Processing Transfer...
                  </>
                ) : (
                  "Transfer Funds"
                )}
              </button>
            </div>
          </form>
        </div>
        
        {result && (
          <div className="glass-panel p-6 rounded-lg h-fit">
            <h3 className="eyebrow border-b border-(--line) pb-2">Transaction Result</h3>
            <div
              className={`mt-4 p-5 rounded-md border ${
                result.status === "success"
                  ? "bg-green-50/50 border-green-200 text-green-800"
                  : result.status === "error" 
                    ? "bg-amber-50/50 border-amber-200 text-amber-800"
                    : "bg-red-50/50 border-red-200 text-red-800"
              }`}
            >
              <div className="flex items-center gap-2 mb-2">
                {result.status === "success" && (
                  <svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7"></path></svg>
                )}
                {(result.status === "rejected" || result.status === "error") && (
                  <svg className="w-5 h-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                )}
                <p className="font-semibold text-lg">{result.status === "success" ? "Success" : "Failed"}</p>
              </div>
              
              <p className="text-sm mt-1">{result.message}</p>
              
              {(result.status === "rejected" || result.status === "error") && (
                <div className="mt-4 flex gap-2">
                  <button
                    onClick={handleConfirmedSubmit}
                    disabled={isLoading}
                    className="px-3 py-1.5 text-xs font-medium text-white bg-(--accent) hover:bg-(--accent-strong) rounded-md transition-colors disabled:bg-slate-300 disabled:cursor-not-allowed flex items-center gap-1"
                  >
                    {isLoading ? (
                      <>
                        <svg className="animate-spin h-3 w-3" fill="none" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        Retrying...
                      </>
                    ) : (
                      <>
                        <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
                        </svg>
                        Retry
                      </>
                    )}
                  </button>
                </div>
              )}
              
              {result.data && result.status !== "success" && (
                <div className="mt-4 bg-white/60 p-3 rounded text-xs font-mono overflow-auto border border-black/5">
                  <span className="font-semibold block mb-1 text-slate-500">Error Details:</span>
                  {JSON.stringify(result.data, null, 2)}
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Confirmation Modal */}
      {showConfirmation && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
          <div className="bg-white rounded-xl shadow-xl max-w-md w-full p-6 animate-in fade-in zoom-in duration-200">
            <h2 className="text-xl font-semibold mb-4 text-slate-800">Confirm Transfer</h2>
            <div className="bg-slate-50 rounded-lg p-4 mb-6 border border-slate-100">
              <div className="grid grid-cols-2 gap-y-3 text-sm">
                <div className="text-slate-500">From:</div>
                <div className="font-medium text-slate-800 text-right">{fromAccount}</div>
                
                <div className="text-slate-500">To:</div>
                <div className="font-medium text-slate-800 text-right">{toAccount}</div>
                
                <div className="text-slate-500">Amount:</div>
                <div className="font-semibold text-lg text-(--accent) text-right">
                  ${parseFloat(amount).toFixed(2)}
                </div>
              </div>
            </div>
            
            <p className="text-sm text-slate-600 mb-6">
              Please verify the details above. This action cannot be undone once confirmed.
            </p>
            
            <div className="flex gap-3 justify-end">
              <button
                type="button"
                onClick={() => setShowConfirmation(false)}
                className="px-4 py-2 text-sm font-medium text-slate-600 hover:text-slate-800 bg-slate-100 hover:bg-slate-200 rounded-lg transition-colors"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleConfirmedSubmit}
                className="px-4 py-2 text-sm font-medium text-white bg-(--accent) hover:bg-(--accent-strong) rounded-lg shadow-sm transition-colors"
              >
                Confirm Transfer
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
