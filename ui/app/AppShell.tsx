"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { isSessionValid, clearSession, getSession, getAuthHeaders } from "@/lib/session";
import { transactionServiceUrl } from "@/lib/server-api";

const navLinks = [
  { href: "/", label: "Dashboard" },
  { href: "/transfer", label: "Transfer" },
  { href: "/topup", label: "Top-up" },
  { href: "/transactions", label: "Transactions" },
  { href: "/demo", label: "Demo" },
  { href: "/settings", label: "Settings" },
];

function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [isAuth, setIsAuth] = useState(false);

  useEffect(() => {
    if (typeof window !== "undefined") {
      if (!isSessionValid()) {
        router.push("/login");
      } else {
        setIsAuth(true);
      }
      
      const interval = setInterval(() => {
        if (!isSessionValid()) {
          setIsAuth(false);
          router.push("/login");
        }
      }, 5000);
      
      return () => clearInterval(interval);
    }
  }, [router]);

  if (!isAuth) {
    return null;
  }

  return <>{children}</>;
}

export default function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [userEmail, setUserEmail] = useState<string | null>(null);

  useEffect(() => {
    const session = getSession();
    if (session) {
      setUserEmail(session.email);
      
      fetch(`/api/auth/me`, {
        headers: {
          ...getAuthHeaders(),
        },
      })
        .then((res) => {
          if (res.ok) {
            return res.json();
          }
          return null;
        })
        .then((data) => {
          if (data && data.email) {
            setUserEmail(data.email);
          }
        })
        .catch(() => {});
    }
  }, []);

  const handleLogout = () => {
    clearSession();
    router.push("/login");
  };

  if (pathname === "/login" || pathname === "/register") {
    return <>{children}</>;
  }

  return (
    <AuthGuard>
      {/* Demo Mode Banner */}
      <div className="bg-amber-100/80 backdrop-blur-sm text-center py-1.5 px-4 sticky top-0 z-50 border-b border-amber-200">
        <p className="text-xs font-semibold text-amber-800 tracking-wide uppercase flex items-center justify-center gap-2">
          <span className="w-2 h-2 rounded-full bg-amber-500 animate-pulse"></span>
          Demo Environment
        </p>
      </div>
      
      <header className="main-header top-8.25">
        <div className="main-header__logo">
          <Link href="/">DD Bank</Link>
        </div>
        <nav className="main-nav desktop-nav">
          <ul className="main-nav__list">
            {navLinks.map((link) => (
              <li key={link.href}>
                <Link 
                  href={link.href}
                  className={pathname === link.href ? "text-(--accent) font-medium" : "flex items-center gap-1"}
                >
                  {link.label === "Demo" && (
                     <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"></path></svg>
                  )}
                  {link.label}
                </Link>
              </li>
            ))}
          </ul>
        </nav>
        <div className="desktop-nav flex items-center gap-4">
          {userEmail && (
            <span className="text-sm font-medium text-slate-500 overflow-hidden text-ellipsis max-w-37.5">
              {userEmail}
            </span>
          )}
          <button onClick={handleLogout} className="logout-button">
            Logout
          </button>
        </div>
        <div className="mobile-nav">
          <button
            onClick={() => setIsMenuOpen(!isMenuOpen)}
            className="mobile-nav__menu-button"
          >
            Menu
          </button>
        </div>
      </header>
      
      {/* Mobile Menu Panel */}
      {isMenuOpen && (
        <div className="mobile-nav__panel mt-8.25">
          <div className="mb-6 flex flex-col gap-2">
            <span className="text-xs uppercase tracking-wider text-slate-400 font-semibold">
              Logged in as
            </span>
            <span className="text-sm font-medium text-slate-700 pb-4 border-b border-slate-200">
              {userEmail || "User"}
            </span>
          </div>
          <nav>
            <ul className="space-y-4">
              {navLinks.map((link) => (
                <li key={link.href}>
                  <Link
                    href={link.href}
                    className={pathname === link.href ? "text-(--accent) font-medium flex items-center gap-2" : "text-slate-600 flex items-center gap-2"}
                    onClick={() => setIsMenuOpen(false)}
                  >
                    {link.label === "Demo" && (
                       <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"></path></svg>
                    )}
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </nav>
          <button onClick={handleLogout} className="logout-button mt-8 w-full">
            Logout
          </button>
        </div>
      )}
      
      <main className="main-content">{children}</main>
    </AuthGuard>
  );
}
