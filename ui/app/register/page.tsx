"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import Link from "next/link";
import { setSession } from "@/lib/session";
import { transactionServiceUrl } from "@/lib/server-api";

export default function RegisterPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [phone, setPhone] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError("");

    if (password.length < 6) {
      setError("Password must be at least 6 characters long.");
      setIsLoading(false);
      return;
    }

    try {
      const response = await fetch(`${transactionServiceUrl()}/auth/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password, phone }),
      });

      const data = await response.json();

      if (!response.ok) {
        setError(data.message || "Registration failed. Please try again.");
      } else {
        // Automatically set session on successful registration to avoid forcing a separate login
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
    <div className="flex items-center justify-center min-h-screen p-4">
      <div className="glass-panel p-8 rounded-lg max-w-sm w-full shadow-lg">
        <h1 className="text-2xl font-bold tracking-tight text-center text-slate-900">
          Create Account
        </h1>
        <p className="text-center text-sm text-slate-500 mt-2 mb-6">
          Join DD Bank to manage your finances
        </p>

        {error && (
          <div className="mb-4 p-3 bg-red-50 text-red-700 border border-red-200 rounded-md text-sm font-medium">
            {error}
          </div>
        )}

        <form onSubmit={handleRegister} className="space-y-4">
          <div>
            <label htmlFor="email" className="eyebrow block mb-1">
              Email
            </label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full p-2.5 border border-[color:var(--line)] rounded-md bg-white/50 focus:outline-none focus:ring-2 focus:ring-[color:var(--accent)]"
              placeholder="you@example.com"
              required
            />
          </div>
          <div>
            <label htmlFor="password" className="eyebrow block mb-1">
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full p-2.5 border border-[color:var(--line)] rounded-md bg-white/50 focus:outline-none focus:ring-2 focus:ring-[color:var(--accent)]"
              placeholder="min. 6 characters"
              required
            />
          </div>
          <div>
            <label htmlFor="phone" className="eyebrow block mb-1">
              Phone (Optional)
            </label>
            <input
              id="phone"
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              className="w-full p-2.5 border border-[color:var(--line)] rounded-md bg-white/50 focus:outline-none focus:ring-2 focus:ring-[color:var(--accent)]"
              placeholder="+1 555-0192"
            />
          </div>
          
          <div className="pt-4">
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
                "Register"
              )}
            </button>
          </div>
        </form>
        
        <div className="mt-6 text-center text-sm text-slate-500 border-t border-[color:var(--line)] pt-4">
          <p>Already have an account? <Link href="/login" className="text-[color:var(--accent)] hover:underline font-medium">Sign in</Link></p>
        </div>
      </div>
    </div>
  );
}
