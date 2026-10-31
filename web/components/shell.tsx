"use client";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { ThemeProvider, useTheme } from "next-themes";
import { createContext, useContext, useSyncExternalStore } from "react";
import {
  Activity,
  LayoutDashboard,
  Search,
  Star,
  GitCompareArrows,
  Sun,
  Moon,
  Menu,
  ArrowUpRight,
  X,
} from "lucide-react";
import * as Dialog from "@radix-ui/react-dialog";
const empty: string[] = [];
let cached: string[] = [];
let raw: string | null = null;
function read() {
  try {
    const v = localStorage.getItem("crypto-watchlist-v1");
    if (v !== raw) {
      raw = v;
      const a = JSON.parse(v ?? "[]");
      cached = Array.isArray(a)
        ? a.filter((x: unknown) => typeof x === "string").slice(0, 250)
        : [];
    }
    return cached;
  } catch {
    return cached;
  }
}
function subscribe(fn: () => void) {
  window.addEventListener("storage", fn);
  window.addEventListener("watchlist", fn);
  return () => {
    window.removeEventListener("storage", fn);
    window.removeEventListener("watchlist", fn);
  };
}
const WatchContext = createContext<{
  ids: string[];
  toggle: (id: string) => void;
}>({ ids: [], toggle: () => {} });
export const useWatch = () => useContext(WatchContext);
export function Providers({ children }: { children: React.ReactNode }) {
  const ids = useSyncExternalStore(subscribe, read, () => empty);
  function toggle(id: string) {
    const next = ids.includes(id) ? ids.filter((x) => x !== id) : [...ids, id];
    try {
      localStorage.setItem("crypto-watchlist-v1", JSON.stringify(next));
      window.dispatchEvent(new Event("watchlist"));
    } catch {
      cached = next;
      window.dispatchEvent(new Event("watchlist"));
    }
  }
  return (
    <ThemeProvider attribute="class" defaultTheme="dark" enableSystem={false}>
      <WatchContext.Provider value={{ ids, toggle }}>
        {children}
      </WatchContext.Provider>
    </ThemeProvider>
  );
}
const nav = [
  { href: "/", label: "Overview", icon: LayoutDashboard },
  { href: "/market", label: "Markets", icon: Search },
  { href: "/watchlist", label: "Watchlist", icon: Star },
  { href: "/compare", label: "Compare", icon: GitCompareArrows },
];
function Navigation() {
  const path = usePathname();
  const { ids } = useWatch();
  return (
    <>
      <Link href="/" className="brand">
        <span className="brand-icon">
          <Activity size={24} />
        </span>
        <span>
          CRYPTO<span className="brand-sub">MARKET OBSERVATORY</span>
        </span>
      </Link>
      <div className="nav-caption">WORKSPACE</div>
      <nav aria-label="Main navigation">
        {nav.map((n) => (
          <Link
            key={n.href}
            href={n.href}
            className={`nav-link ${path === n.href ? "active" : ""}`}
          >
            <n.icon size={18} />
            {n.label}
            {n.href === "/watchlist" && (
              <span className="count">{ids.length}</span>
            )}
          </Link>
        ))}
      </nav>
      <div className="sidebar-bottom">
        <span className="eyebrow">BUILT FOR PERSPECTIVE</span>
        <p>
          Understand the market.
          <br />
          Look beyond the price.
        </p>
        <a
          href="https://github.com/viczuno/go-crypto-bot"
          target="_blank"
          rel="noreferrer"
        >
          Open source <ArrowUpRight size={14} />
        </a>
      </div>
    </>
  );
}
export function Shell({ children }: { children: React.ReactNode }) {
  const { resolvedTheme, setTheme } = useTheme();
  const mounted = useSyncExternalStore(
    () => () => {},
    () => true,
    () => false,
  );
  return (
    <>
      <a className="skip-link" href="#main">
        Skip to content
      </a>
      <aside className="sidebar">
        <Navigation />
      </aside>
      <div className="workspace">
        <header className="topbar">
          <Dialog.Root>
            <Dialog.Trigger
              className="icon-button mobile-menu"
              aria-label="Open navigation"
            >
              <Menu size={20} />
            </Dialog.Trigger>
            <Dialog.Portal>
              <Dialog.Overlay className="dialog-overlay" />
              <Dialog.Content className="mobile-drawer">
                <Dialog.Title className="sr-only">Navigation</Dialog.Title>
                <Dialog.Description className="sr-only">
                  Market workspace links
                </Dialog.Description>
                <Dialog.Close
                  className="icon-button drawer-close"
                  aria-label="Close navigation"
                >
                  <X />
                </Dialog.Close>
                <Navigation />
              </Dialog.Content>
            </Dialog.Portal>
          </Dialog.Root>
          <div className="breadcrumb">
            Workspace <span>/</span> Market analytics
          </div>
          <div className="topbar-right">
            <span className="tag">USD</span>
            <span className="schedule">
              <i /> Twice-daily observations
            </span>
            <button
              className="icon-button"
              aria-label="Toggle color theme"
              onClick={() =>
                setTheme(resolvedTheme === "dark" ? "light" : "dark")
              }
            >
              {mounted && resolvedTheme === "dark" ? (
                <Sun size={18} />
              ) : (
                <Moon size={18} />
              )}
            </button>
          </div>
        </header>
        <main id="main">{children}</main>
        <footer className="footer">
          <span>Go Crypto Bot · Independent market research</span>
          <span>Data: CoinGecko · Sampled prices, not live trading quotes</span>
        </footer>
      </div>
    </>
  );
}
export function Favorite({ id }: { id: string }) {
  const { ids, toggle } = useWatch();
  const selected = ids.includes(id);
  return (
    <button
      className={`icon-button favorite ${selected ? "selected" : ""}`}
      aria-label={`${selected ? "Remove from" : "Add to"} watchlist`}
      aria-pressed={selected}
      onClick={() => toggle(id)}
    >
      <Star size={17} fill={selected ? "currentColor" : "none"} />
    </button>
  );
}
