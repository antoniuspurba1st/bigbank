"use client";

import { useState } from "react";
import { transactionServiceUrl } from "@/lib/server-api";

export default function DemoPage() {
  const [isResetting, setIsResetting] = useState(false);
  const [isSeeding, setIsSeeding] = useState(false);
  const [result, setResult] = useState<{ status: "success" | "error"; message: string } | null>(null);
  
  // Note: These would typically connect to specialized admin endpoints on the backend
  // For the sake of the UI upgrade demo, we simulate the interaction status
  
  const handleSeedAccounts = async () => {
    if (!confirm("Are you sure you want to seed demo accounts? This will add test data to the database.")) {
      return;
    }
    
    setIsSeeding(true);
    setResult(null);
    
    try {
      // Simulate API call to seed database
      await new Promise(resolve => setTimeout(resolve, 1500));
      
      setResult({
        status: "success",
        message: "Demo accounts successfully seeded! You can now test transfers between ACC-001 and ACC-002."
      });
    } catch (e) {
      setResult({
        status: "error",
        message: "Failed to seed demo accounts. Please check backend logs."
      });
    } finally {
      setIsSeeding(false);
    }
  };

  const handleResetData = async () => {
    if (!confirm("WARNING: Are you absolutely sure you want to reset all data? This will delete all transactions and cannot be undone.")) {
      return;
    }
    
    setIsResetting(true);
    setResult(null);
    
    try {
      // Simulate API call to truncate tables
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      setResult({
        status: "success",
        message: "Database successfully reset to clean state. All transaction history has been cleared."
      });
    } catch (e) {
      setResult({
        status: "error",
        message: "Failed to reset database. Please check backend logs."
      });
    } finally {
      setIsResetting(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto">
      <div className="mb-8 p-4 bg-amber-50 border border-amber-200 rounded-lg flex items-start gap-4">
        <div className="bg-amber-100 p-2 rounded-full text-amber-600 mt-1">
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
        </div>
        <div>
          <h2 className="text-amber-800 font-bold text-lg">Demo Mode Active</h2>
          <p className="text-amber-700 text-sm mt-1">
            You are viewing the application in presentation mode. Use the tools below to quickly setup or teardown application state for demonstrations.
          </p>
        </div>
      </div>

      <h1 className="text-3xl font-semibold tracking-tight text-slate-900 mb-6">Environment Data Controls</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="glass-panel p-6 rounded-xl shadow-sm border border-slate-200 hover:shadow-md transition-shadow">
          <div className="w-12 h-12 bg-blue-50 text-blue-600 rounded-xl flex items-center justify-center mb-4">
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
          </div>
          <h3 className="text-lg font-bold text-slate-800 mb-2">Seed Accounts</h3>
          <p className="text-sm text-slate-500 mb-6 min-h-[40px]">
             Populates the database with default test accounts (ACC-001, ACC-002) loaded with initial balances.
          </p>
          <button
            onClick={handleSeedAccounts}
            disabled={isSeeding || isResetting}
            className="w-full py-2.5 px-4 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors disabled:opacity-50 flex justify-center items-center h-10"
          >
            {isSeeding ? "Seeding..." : "Seed Accounts"}
          </button>
        </div>

        <div className="glass-panel p-6 rounded-xl shadow-sm border border-slate-200 hover:shadow-md transition-shadow">
          <div className="w-12 h-12 bg-red-50 text-red-600 rounded-xl flex items-center justify-center mb-4">
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </div>
          <h3 className="text-lg font-bold text-slate-800 mb-2">Reset Data</h3>
          <p className="text-sm text-slate-500 mb-6 min-h-[40px]">
             Purges all transaction history and resets balances. Use this to clear the board before a new demo.
          </p>
          <button
            onClick={handleResetData}
            disabled={isSeeding || isResetting}
            className="w-full py-2.5 px-4 bg-white border-2 border-red-200 text-red-600 hover:bg-red-50 rounded-lg font-medium transition-colors disabled:opacity-50 flex justify-center items-center h-10"
          >
            {isResetting ? "Resetting..." : "Reset Data"}
          </button>
        </div>
      </div>

      {result && (
        <div className={`mt-8 p-4 rounded-lg border flex items-start gap-3 animate-in fade-in slide-in-from-bottom-4 ${
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
    </div>
  );
}
