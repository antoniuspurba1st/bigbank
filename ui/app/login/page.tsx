"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import Link from "next/link";
import { setSession } from "@/lib/session";
import { transactionServiceUrl } from "@/lib/server-api";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const [error, setError] = useState("");

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError("");

    try {
      const response = await fetch(`${transactionServiceUrl()}/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      const data = await response.json();

      if (!response.ok) {
        setError(data.message || "Invalid email or password.");
      } else {
        setSession(data.user.email);
        router.push("/");
      }
    } catch (err) {
      setError("Network error. Please ensure the backend is running.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="glass-panel p-8 rounded-lg max-w-sm w-full">
        <h1 className="text-2xl font-semibold tracking-tight text-center">
          DD Bank Console
        </h1>
        <p className="text-center text-sm text-slate-500 mt-2 mb-6">
          Sign in to your account
        </p>
        
        {error && (
          <div className="mb-4 p-3 bg-red-50 text-red-700 border border-red-200 rounded-md text-sm font-medium">
            {error}
          </div>
        )}

        <form onSubmit={handleLogin} className="space-y-4">
          <div>
            <label htmlFor="email" className="eyebrow block mb-2">
              Email
            </label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full p-2 border border-[color:var(--line)] rounded-md bg-white/50 focus:outline-none focus:ring-2 focus:ring-[color:var(--accent)]"
              placeholder="user@ddbank.com"
              required
            />
          </div>
          <div>
            <label htmlFor="password" className="eyebrow block mb-2">
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full p-2 border border-[color:var(--line)] rounded-md bg-white/50 focus:outline-none focus:ring-2 focus:ring-[color:var(--accent)]"
              placeholder="********"
              required
            />
          </div>
          <div className="pt-2">
            <button
              type="submit"
              disabled={isLoading || !email || !password}
              className="w-full rounded-full bg-[color:var(--accent)] px-5 py-3 text-sm font-medium text-white transition hover:bg-[color:var(--accent-strong)] disabled:bg-slate-300 flex justify-center items-center"
            >
              {isLoading ? (
                <svg className="animate-spin h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
              ) : (
                "Sign In"
              )}
            </button>
          </div>
        </form>
        
        <div className="mt-6 text-center text-sm text-slate-500 border-t border-[color:var(--line)] pt-4">
          <p>Don't have an account? <Link href="/register" className="text-[color:var(--accent)] hover:underline font-medium">Register here</Link></p>
        </div>
      </div>
    </div>
  );
}
