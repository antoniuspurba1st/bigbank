"use client";

import { useState, useEffect } from "react";
import { getSession, setSession } from "@/lib/session";
import { transactionServiceUrl } from "@/lib/server-api";

export default function SettingsPage() {
  const [session, setSessionState] = useState<{ email: string } | null>(null);
  const [phone, setPhone] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [result, setResult] = useState<{ status: "success" | "error"; message: string } | null>(null);

  useEffect(() => {
    // Only access localStorage on client side
    setSessionState(getSession());
  }, []);

  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!session?.email) return;

    setIsLoading(true);
    setResult(null);

    try {
      const response = await fetch(`${transactionServiceUrl()}/auth/phone`, {
        method: "PUT",
        headers: { 
          "Content-Type": "application/json",
          "X-User-Email": session.email // Basic prototype identity header match backend expectation
        },
        body: JSON.stringify({ phone }),
      });

      const data = await response.json();

      if (!response.ok) {
        setResult({
          status: "error",
          message: data.message || "Failed to update profile",
        });
      } else {
        setResult({
          status: "success",
          message: "Profile updated successfully!",
        });
      }
    } catch (err) {
      setResult({
        status: "error",
        message: "Network error. Please try again later.",
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-3xl font-semibold tracking-tight text-slate-900 mb-6">Account Settings</h1>
      
      <div className="glass-panel p-6 rounded-xl shadow-sm border border-slate-200">
        <h3 className="text-xl font-medium text-slate-800 mb-4 border-b border-slate-100 pb-3">Personal Information</h3>
        
        {result && (
          <div className={`mb-6 p-4 rounded-lg border flex items-start gap-3 animate-in fade-in slide-in-from-bottom-2 ${
            result.status === "success" 
              ? "bg-green-50 border-green-200 text-green-800" 
              : "bg-red-50 border-red-200 text-red-800"
          }`}>
            {result.status === "success" ? (
              <svg className="w-5 h-5 mt-0.5 text-green-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7" />
              </svg>
            ) : (
              <svg className="w-5 h-5 mt-0.5 text-red-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            )}
            <p className="font-medium">{result.message}</p>
          </div>
        )}

        <form onSubmit={handleUpdateProfile} className="space-y-6">
          <div>
            <label className="eyebrow block mb-1 text-slate-500">
              Email Address (Cannot be changed)
            </label>
            <input
              type="email"
              value={session?.email || "Loading..."}
              disabled
              className="w-full p-2.5 border border-slate-200 rounded-md bg-slate-50 text-slate-500 cursor-not-allowed"
            />
          </div>

          <div>
            <label htmlFor="phone" className="eyebrow block mb-1">
              Phone Number
            </label>
            <input
              id="phone"
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder="e.g. +1 555-0199"
              className="w-full p-2.5 border border-[color:var(--line)] rounded-md bg-white/50 focus:outline-none focus:ring-2 focus:ring-[color:var(--accent)]"
            />
            <p className="text-xs text-slate-500 mt-1">
              Used for important account notifications.
            </p>
          </div>

          <div className="pt-2">
            <button
              type="submit"
              disabled={isLoading}
              className="px-6 py-2.5 rounded-lg bg-[color:var(--accent)] text-sm font-medium text-white transition hover:bg-[color:var(--accent-strong)] disabled:bg-slate-300 flex items-center gap-2"
            >
              {isLoading ? (
                <>
                  <svg className="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Saving...
                </>
              ) : (
                "Save Changes"
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
