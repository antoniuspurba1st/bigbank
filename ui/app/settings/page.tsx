"use client";

import { useState, useEffect } from "react";
import { getSession } from "@/lib/session";

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
      const response = await fetch(`/api/auth/phone`, {
        method: "PUT",
        headers: { 
          "Content-Type": "application/json",
          "X-User-Email": session.email,
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
            {result.status === "error" && (
              <button
                onClick={handleUpdateProfile}
                disabled={isLoading}
                className="mt-2 px-3 py-1.5 text-xs font-medium text-white bg-red-600 hover:bg-red-700 rounded-md transition-colors disabled:bg-slate-300 disabled:cursor-not-allowed flex items-center gap-1"
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
            )}
          </div>
        )}

        <form onSubmit={handleUpdateProfile} className="space-y-6">
          <div className="space-y-1">
  <label
    htmlFor="email"
    className="block text-sm font-medium text-slate-600"
  >
    Email Address
    <span className="ml-1 text-xs text-slate-400">
      (cannot be changed)
    </span>
  </label>

  <input
    id="email"
    name="email"
    type="email"
    value={session?.email ?? ""}
    readOnly
    aria-readonly="true"
    className="
      w-full
      rounded-md
      border border-slate-200
      bg-slate-50
      px-3 py-2.5
      text-slate-500
      focus:outline-none
      focus:ring-0
      cursor-default
    "
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
              className="w-full p-2.5 border border-(--line) rounded-md bg-white/50 focus:outline-none focus:ring-2 focus:ring-(--accent)"
            />
            <p className="text-xs text-slate-500 mt-1">
              Used for important account notifications.
            </p>
          </div>

          <div className="pt-2">
            <button
              type="submit"
              disabled={isLoading}
              className="px-6 py-2.5 rounded-lg bg-(--accent) text-sm font-medium text-white transition hover:bg-(--accent-strong) disabled:bg-slate-300 flex items-center gap-2"
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
