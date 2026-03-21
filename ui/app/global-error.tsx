"use client";

import { useEffect } from "react";
import Link from "next/link";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Log the error to an error reporting service
    console.error("Global UI Error Caught:", error);
  }, [error]);

  return (
    <html lang="en">
      <body>
        <div className="flex min-h-screen flex-col items-center justify-center p-4 bg-slate-50">
          <div className="glass-panel p-8 rounded-lg max-w-md w-full text-center shadow-xl border border-red-100">
            <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-red-100 mb-6">
              <svg
                className="h-8 w-8 text-red-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                />
              </svg>
            </div>
            <h2 className="text-2xl font-bold text-slate-900 mb-2">
              Something went wrong!
            </h2>
            <p className="text-slate-500 mb-8">
              A critical error occurred in the application interface.
            </p>
            <div className="flex flex-col gap-3">
              <button
                onClick={() => reset()}
                className="w-full rounded-full bg-(--accent) px-5 py-3 text-sm font-medium text-white transition hover:bg-(--accent-strong)"
              >
                Try again
              </button>
              <Link
                href="/"
                className="w-full rounded-full bg-slate-200 px-5 py-3 text-sm font-medium text-slate-800 transition hover:bg-slate-300"
              >
                Go back home
              </Link>
            </div>
          </div>
        </div>
      </body>
    </html>
  );
}
