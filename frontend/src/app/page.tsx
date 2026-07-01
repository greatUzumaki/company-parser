"use client";

import { useRef, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { AnimatePresence, motion } from "framer-motion";
import { RegionMap } from "@/components/RegionMap";
import { FilterPanel } from "@/components/FilterPanel";
import { ResultsTable } from "@/components/ResultsTable";
import { ExportBar } from "@/components/ExportBar";
import { HistoryList } from "@/components/HistoryList";
import { SearchProgress, type SourceProgress } from "@/components/SearchProgress";
import { getCampaignStatus, searchCompaniesStream } from "@/lib/api";
import type { Company, StreamEvent } from "@/lib/types";
import { useSearch } from "@/store/useSearch";
import { useLangInit, useLangStore, useT, type Lang } from "@/lib/i18n";

type LeftTab = "filters" | "history";

export default function Home() {
  const queryClient = useQueryClient();
  const t = useT();
  useLangInit();
  const region = useSearch((s) => s.region);
  const filters = useSearch((s) => s.filters);

  const [leftOpen, setLeftOpen] = useState(true);
  const [rightOpen, setRightOpen] = useState(false);
  const [leftTab, setLeftTab] = useState<LeftTab>("filters");

  const campaignStatus = useQuery({
    queryKey: ["campaign-status"],
    queryFn: getCampaignStatus,
    staleTime: Infinity,
  });

  // Streaming search state.
  const [companies, setCompanies] = useState<Company[]>([]);
  const [searchId, setSearchId] = useState<number | null>(null);
  const [streaming, setStreaming] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sources, setSources] = useState<SourceProgress[]>([]);
  const abortRef = useRef<AbortController | null>(null);

  const runSearch = (forceArg = false) => {
    // Coerce so a MouseEvent passed by an onClick never leaks into the request.
    const force = forceArg === true;
    if (!region) return;
    abortRef.current?.abort();
    const ac = new AbortController();
    abortRef.current = ac;

    setCompanies([]);
    setSearchId(null);
    setError(null);
    setSources([]);
    setStreaming(true);
    setRightOpen(true);

    const acc: Company[] = [];
    const seen = new Set<string>();
    const appendCompanies = (list?: Company[]) => {
      if (!list?.length) return;
      for (const c of list) {
        const k = `${c.osmType}/${c.osmId}`;
        if (!seen.has(k)) {
          seen.add(k);
          acc.push(c);
        }
      }
      setCompanies([...acc]);
    };

    searchCompaniesStream(
      region,
      filters,
      (e: StreamEvent) => {
        switch (e.type) {
          case "cached":
            // Cached results render instantly; a refresh may follow.
            appendCompanies(e.companies);
            break;
          case "source_start":
            setSources((prev) =>
              prev.some((s) => s.name === e.source)
                ? prev
                : [...prev, { name: e.source ?? "", status: "pending", count: 0 }],
            );
            break;
          case "companies":
            appendCompanies(e.companies);
            setSources((prev) =>
              prev.map((s) => (s.name === e.source ? { ...s, count: e.count ?? s.count } : s)),
            );
            break;
          case "source_done":
            setSources((prev) =>
              prev.map((s) =>
                s.name === e.source
                  ? { ...s, status: e.message === "failed" ? "failed" : "done", count: e.count ?? s.count }
                  : s,
              ),
            );
            break;
          case "done":
            setSearchId(e.searchId ?? null);
            setStreaming(false);
            queryClient.invalidateQueries({ queryKey: ["searches"] });
            break;
          case "error":
            setError(e.message ?? "Search failed");
            setStreaming(false);
            break;
        }
      },
      ac.signal,
      force,
    ).catch((err) => {
      if (ac.signal.aborted) return;
      setError(err instanceof Error ? err.message : "Search failed");
      setStreaming(false);
    });
  };

  const results = companies;
  const hasSearch = streaming || companies.length > 0 || searchId !== null;

  return (
    <div className="fixed inset-0 overflow-hidden bg-slate-100 dark:bg-slate-950">
      {/* Full-screen map canvas */}
      <RegionMap fullScreen companies={companies} />

      {/* Compact floating controls: brand mark + language switch */}
      <div className="absolute right-4 top-4 z-30 flex items-center gap-2">
        <LangSwitcher />
        <span
          title="Company Parser"
          className="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-emerald-500 to-indigo-500 text-white shadow-lg ring-1 ring-black/5"
        >
          <svg viewBox="0 0 24 24" fill="none" className="h-5 w-5" stroke="currentColor" strokeWidth={1.8}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 21s-6.5-5.6-6.5-10.3A6.5 6.5 0 0112 4a6.5 6.5 0 016.5 6.7C18.5 15.4 12 21 12 21z" />
            <circle cx="12" cy="10.5" r="2.2" />
          </svg>
        </span>
      </div>

      {/* Show-left button (when collapsed) */}
      <AnimatePresence>
        {!leftOpen && (
          <motion.button
            type="button"
            initial={{ opacity: 0, x: -12 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -12 }}
            onClick={() => setLeftOpen(true)}
            aria-label="Show filters panel"
            className="absolute left-4 top-4 z-30 flex h-11 w-11 items-center justify-center rounded-xl bg-white/90 text-slate-700 shadow-lg ring-1 ring-slate-200/80 backdrop-blur transition-colors hover:text-emerald-600 dark:bg-slate-900/90 dark:text-slate-200 dark:ring-slate-700/80"
          >
            <PanelIcon />
          </motion.button>
        )}
      </AnimatePresence>

      {/* Left floating sidebar: filters + history */}
      <AnimatePresence>
        {leftOpen && (
          <motion.aside
            key="left"
            initial={{ opacity: 0, x: -40 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -40 }}
            transition={{ type: "spring", stiffness: 260, damping: 30 }}
            className="absolute bottom-4 left-4 top-4 z-20 flex w-[88vw] max-w-[360px] flex-col overflow-hidden rounded-2xl bg-white/85 shadow-xl ring-1 ring-slate-200/80 backdrop-blur dark:bg-slate-900/85 dark:ring-slate-700/80"
          >
            <div className="flex items-center justify-between border-b border-slate-200/70 px-3 py-2 dark:border-slate-700/70">
              <div className="flex gap-1 rounded-lg bg-slate-100 p-0.5 dark:bg-slate-800">
                <TabButton active={leftTab === "filters"} onClick={() => setLeftTab("filters")}>
                  {t("tab.filters")}
                </TabButton>
                <TabButton active={leftTab === "history"} onClick={() => setLeftTab("history")}>
                  {t("tab.history")}
                </TabButton>
              </div>
              <button
                type="button"
                onClick={() => setLeftOpen(false)}
                aria-label="Hide panel"
                className="flex h-8 w-8 items-center justify-center rounded-lg text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-700 dark:hover:bg-slate-800 dark:hover:text-slate-200"
              >
                <ChevronIcon dir="left" />
              </button>
            </div>

            <div className="scroll-slim min-h-0 flex-1 overflow-y-auto p-3">
              {leftTab === "filters" ? (
                <FilterPanel onSearch={runSearch} isPending={streaming} />
              ) : (
                <HistoryList />
              )}
            </div>

            <AnimatePresence>
              {error && (
                <motion.div
                  initial={{ opacity: 0, height: 0 }}
                  animate={{ opacity: 1, height: "auto" }}
                  exit={{ opacity: 0, height: 0 }}
                  className="border-t border-red-200 bg-red-50 px-3 py-2 text-xs text-red-700 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-300"
                >
                  {error}
                </motion.div>
              )}
            </AnimatePresence>
          </motion.aside>
        )}
      </AnimatePresence>

      {/* Show-right button (when results exist but panel collapsed) */}
      <AnimatePresence>
        {!rightOpen && hasSearch && (
          <motion.button
            type="button"
            initial={{ opacity: 0, x: 12 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: 12 }}
            onClick={() => setRightOpen(true)}
            aria-label="Show results panel"
            className="absolute right-4 top-16 z-30 flex items-center gap-2 rounded-xl bg-white/90 px-3 py-2.5 text-sm font-medium text-slate-700 shadow-lg ring-1 ring-slate-200/80 backdrop-blur transition-colors hover:text-emerald-600 dark:bg-slate-900/90 dark:text-slate-200 dark:ring-slate-700/80"
          >
            <PanelIcon />
            {t("results.title")}
            <span className="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-400">
              {companies.length}
            </span>
          </motion.button>
        )}
      </AnimatePresence>

      {/* Right floating panel: results + export */}
      <AnimatePresence>
        {rightOpen && (
          <motion.aside
            key="right"
            initial={{ opacity: 0, x: 40 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: 40 }}
            transition={{ type: "spring", stiffness: 260, damping: 30 }}
            className="absolute bottom-4 right-4 top-16 z-20 flex w-[92vw] max-w-[520px] flex-col overflow-hidden rounded-2xl bg-white/90 shadow-xl ring-1 ring-slate-200/80 backdrop-blur dark:bg-slate-900/90 dark:ring-slate-700/80"
          >
            <div className="flex items-center justify-between gap-3 border-b border-slate-200/70 px-4 py-3 dark:border-slate-700/70">
              <div className="flex items-center gap-2">
                <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-50">{t("results.title")}</h2>
                {hasSearch && (
                  <span className="flex items-center gap-1.5 rounded-full bg-slate-100 px-2 py-0.5 text-xs font-semibold text-slate-600 dark:bg-slate-800 dark:text-slate-300">
                    {streaming && (
                      <span className="h-3 w-3 animate-spin rounded-full border-2 border-emerald-500/40 border-t-emerald-500" aria-hidden />
                    )}
                    {companies.length}
                  </span>
                )}
              </div>
              <div className="flex items-center gap-2">
                {hasSearch && region && (
                  <button
                    type="button"
                    onClick={() => runSearch(true)}
                    disabled={streaming}
                    aria-label={t("refresh")}
                    title={t("refresh")}
                    className="flex h-8 w-8 items-center justify-center rounded-lg text-slate-400 transition-colors hover:bg-slate-100 hover:text-emerald-600 disabled:opacity-40 dark:hover:bg-slate-800"
                  >
                    <svg viewBox="0 0 24 24" fill="none" className="h-4 w-4" stroke="currentColor" strokeWidth={2}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M4 4v5h5M20 20v-5h-5M20 9a8 8 0 00-14.9-2M4 15a8 8 0 0014.9 2" />
                    </svg>
                  </button>
                )}
                <ExportBar searchId={searchId} />
                <button
                  type="button"
                  onClick={() => setRightOpen(false)}
                  aria-label="Hide results"
                  className="flex h-8 w-8 items-center justify-center rounded-lg text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-700 dark:hover:bg-slate-800 dark:hover:text-slate-200"
                >
                  <ChevronIcon dir="right" />
                </button>
              </div>
            </div>

            <SearchProgress sources={sources} streaming={streaming} />

            <div className="scroll-slim min-h-0 flex-1 overflow-y-auto p-3">
              <ResultsTable
                data={results}
                isLoading={streaming && companies.length === 0}
                campaignEnabled={campaignStatus.data?.enabled === true}
              />
            </div>
          </motion.aside>
        )}
      </AnimatePresence>

      {/* ODbL attribution */}
      <div className="pointer-events-none absolute bottom-2 left-1/2 z-10 -translate-x-1/2 rounded-full bg-white/70 px-3 py-1 text-[11px] text-slate-500 backdrop-blur dark:bg-slate-900/70 dark:text-slate-400">
        Data © OpenStreetMap contributors (ODbL)
      </div>

    </div>
  );
}

function LangSwitcher() {
  const lang = useLangStore((s) => s.lang);
  const setLang = useLangStore((s) => s.setLang);
  const options: Lang[] = ["en", "ru"];
  return (
    <div className="flex items-center gap-0.5 rounded-xl bg-white/90 p-0.5 shadow-lg ring-1 ring-slate-200/80 backdrop-blur dark:bg-slate-900/90 dark:ring-slate-700/80">
      {options.map((l) => (
        <button
          key={l}
          type="button"
          onClick={() => setLang(l)}
          aria-pressed={lang === l}
          className={`rounded-lg px-2.5 py-1.5 text-xs font-semibold uppercase transition-colors ${
            lang === l
              ? "bg-emerald-500 text-white shadow-sm"
              : "text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200"
          }`}
        >
          {l}
        </button>
      ))}
    </div>
  );
}

function TabButton({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={`rounded-md px-3 py-1.5 text-xs font-semibold transition-colors ${
        active
          ? "bg-white text-emerald-700 shadow-sm dark:bg-slate-700 dark:text-emerald-300"
          : "text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200"
      }`}
    >
      {children}
    </button>
  );
}

function PanelIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" className="h-5 w-5" stroke="currentColor" strokeWidth={1.8}>
      <rect x="3" y="4" width="18" height="16" rx="2" />
      <path d="M9 4v16" strokeLinecap="round" />
    </svg>
  );
}

function ChevronIcon({ dir }: { dir: "left" | "right" }) {
  return (
    <svg viewBox="0 0 24 24" fill="none" className="h-5 w-5" stroke="currentColor" strokeWidth={2}>
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d={dir === "left" ? "M15 18l-6-6 6-6" : "M9 6l6 6-6 6"}
      />
    </svg>
  );
}
