"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import Link from "next/link";
import { setSession, clearSession } from "@/lib/session";
import { transactionServiceUrl } from "@/lib/server-api";

export default function RegisterPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [phone, setPhone] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [passwordError, setPasswordError] = useState("");
  const [confirmPasswordError, setConfirmPasswordError] = useState("");

  const validatePassword = (pwd: string): boolean => {
    if (pwd.length < 8) {
      setPasswordError("Password must be at least 8 characters long");
      return false;
    }
    if (!/[A-Z]/.test(pwd)) {
      setPasswordError("Password must contain at least one uppercase letter");
      return false;
    }
    if (!/[0-9]/.test(pwd)) {
      setPasswordError("Password must contain at least one number");
      return false;
    }
    if (!/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(pwd)) {
      setPasswordError("Password must contain at least one special character (!@#$%^&*...)");
      return false;
    }
    setPasswordError("");
    return true;
  };

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newPassword = e.target.value;
    setPassword(newPassword);
    if (newPassword.length > 0) {
      validatePassword(newPassword);
    } else {
      setPasswordError("");
    }

    if (confirmPassword.length > 0) {
      if (newPassword !== confirmPassword) {
        setConfirmPasswordError("Passwords do not match");
      } else {
        setConfirmPasswordError("");
      }
    }
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError("");
    setConfirmPasswordError("");

    if (!validatePassword(password)) {
      setIsLoading(false);
      return;
    }

    if (password !== confirmPassword) {
      setConfirmPasswordError("Passwords do not match");
      setIsLoading(false);
      return;
    }

    try {
      const response = await fetch(`/api/auth/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password, phone }),
      });

      const data = await response.json();

      if (!response.ok) {
        setError(data.message || "Registration failed. Please try again.");
      } else {
        clearSession(); // Clear any existing session
        setSession(data.user.id, data.user.email);
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
              className="w-full p-2.5 border border-(--line) rounded-md bg-white/50 focus:outline-none focus:ring-2 focus:ring-(--accent)"
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
              onChange={handlePasswordChange}
              className={`w-full p-2.5 border rounded-md bg-white/50 focus:outline-none focus:ring-2 focus:ring-(--accent) ${
                passwordError ? "border-red-300" : "border-(--line)"
              }`}
              placeholder="min. 8 characters"
              required
            />
            {passwordError && (
              <p className="text-xs text-red-600 mt-1">{passwordError}</p>
            )}
            <div className="mt-2 text-xs text-slate-500 space-y-1">
              <p className="font-medium">Password requirements:</p>
              <ul className="space-y-0.5 ml-4">
                <li className={password.length >= 8 ? "text-green-600" : ""}>
                  ✓ At least 8 characters
                </li>
                <li className={/[A-Z]/.test(password) ? "text-green-600" : ""}>
                  ✓ One uppercase letter
                </li>
                <li className={/[0-9]/.test(password) ? "text-green-600" : ""}>
                  ✓ One number
                </li>
                <li className={/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password) ? "text-green-600" : ""}>
                  ✓ One special character
                </li>
              </ul>
            </div>
          </div>
          <div>
            <label htmlFor="confirmPassword" className="eyebrow block mb-1">
              Confirm Password
            </label>
            <input
              id="confirmPassword"
              type="password"
              value={confirmPassword}
              onChange={(e) => {
                setConfirmPassword(e.target.value);
                if (e.target.value.length > 0 && e.target.value !== password) {
                  setConfirmPasswordError("Passwords do not match");
                } else {
                  setConfirmPasswordError("");
                }
              }}
              className={`w-full p-2.5 border rounded-md bg-white/50 focus:outline-none focus:ring-2 focus:ring-(--accent) ${
                confirmPasswordError ? "border-red-300" : "border-(--line)"
              }`}
              placeholder="Confirm your password"
              required
            />
            {confirmPasswordError && (
              <p className="text-xs text-red-600 mt-1">{confirmPasswordError}</p>
            )}
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
              className="w-full p-2.5 border border-(--line) rounded-md bg-white/50 focus:outline-none focus:ring-2 focus:ring-(--accent)"
              placeholder="+1 555-0192"
            />
          </div>
          
          <div className="pt-4">
            <button
              type="submit"
              disabled={isLoading || !email || !password || !confirmPassword || !!confirmPasswordError || !!passwordError}
              className="w-full rounded-full bg-(--accent) px-5 py-3 text-sm font-medium text-white transition hover:bg-(--accent-strong) disabled:bg-slate-300 flex justify-center items-center"
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
        
        <div className="mt-6 text-center text-sm text-slate-500 border-t border-(--line) pt-4">
          <p>Already have an account? <Link href="/login" className="text-(--accent) hover:underline font-medium">Sign in</Link></p>
        </div>
      </div>
    </div>
  );
}
