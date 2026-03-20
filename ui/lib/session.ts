export interface UserSession {
  email: string;
  expiresAt: number;
}

const SESSION_KEY = "auth_session";
const SESSION_DURATION_MS = 60 * 60 * 1000; // 1 hour

export function setSession(email: string) {
  const session: UserSession = {
    email,
    expiresAt: Date.now() + SESSION_DURATION_MS,
  };
  localStorage.setItem(SESSION_KEY, JSON.stringify(session));
}

export function getSession(): UserSession | null {
  if (typeof window === "undefined") return null;

  const stored = localStorage.getItem(SESSION_KEY);
  if (!stored) return null;

  try {
    const session: UserSession = JSON.parse(stored);
    
    // Check expiry
    if (Date.now() > session.expiresAt) {
      clearSession();
      return null;
    }
    
    return session;
  } catch (e) {
    clearSession();
    return null;
  }
}

export function clearSession() {
  localStorage.removeItem(SESSION_KEY);
  // Also clean up old dummy state if exists
  localStorage.removeItem("isLoggedIn");
}

export function isSessionValid(): boolean {
  return getSession() !== null;
}
