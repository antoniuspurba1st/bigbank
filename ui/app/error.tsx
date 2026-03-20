"use client";

import { useEffect } from "react";
import Link from "next/link";

export default function ErrorBoundary({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Log the error to an error reporting service
    console.error("UI Error Caught:", error);
  }, [error]);

  return (
    <div className="flex h-[70vh] flex-col items-center justify-center p-4">
      <div className="glass-panel p-8 rounded-lg max-w-md w-full text-center border border-red-50 relative overflow-hidden">
        <div className="absolute top-0 left-0 w-full h-1 bg-red-400"></div>
        <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-red-50 mb-5 text-red-500">
          <svg
            className="h-7 w-7"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
              d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        </div>
        <h2 className="text-xl font-bold text-slate-800 mb-2">
          Page Error
        </h2>
        <p className="text-sm text-slate-500 mb-6">
          We encountered an unexpected error while rendering this page.
        </p>
        <div className="flex flex-col gap-3">
          <button
            onClick={() => reset()}
            className="w-full rounded-full bg-[color:var(--accent)] px-5 py-2.5 text-sm font-medium text-white transition hover:bg-[color:var(--accent-strong)]"
          >
            Try reloading
          </button>
          <Link
            href="/"
            className="w-full rounded-full bg-slate-100 px-5 py-2.5 text-sm font-medium text-slate-700 transition hover:bg-slate-200"
          >
            Return to Dashboard
          </Link>
        </div>
      </div>
    </div>
  );
}
