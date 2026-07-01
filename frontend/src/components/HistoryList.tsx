"use client";

import { useQuery } from "@tanstack/react-query";
import { AnimatePresence, motion } from "framer-motion";
import { listSearches } from "@/lib/api";
import type { SearchSummary } from "@/lib/types";
import { useLangStore, useT, type TFunc } from "@/lib/i18n";

function formatDate(iso: string, locale: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleString(locale, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function activeFilters(s: SearchSummary, t: TFunc): string[] {
  const out: string[] = [];
  if (s.filters?.noWebsite) out.push(t("hfilter.noWebsite"));
  if (s.filters?.noSocials) out.push(t("hfilter.noSocials"));
  if (s.filters?.noPhone) out.push(t("hfilter.noPhone"));
  if (s.filters?.categories?.length)
    out.push(t("hfilter.categories", { n: s.filters.categories.length }));
  return out;
}

export function HistoryList() {
  const t = useT();
  const lang = useLangStore((s) => s.lang);
  const query = useQuery({
    queryKey: ["searches"],
    queryFn: listSearches,
  });

  return (
    <div className="flex h-full flex-col rounded-2xl border border-slate-200 bg-white/70 p-4 shadow-sm backdrop-blur sm:p-5 dark:border-slate-800 dark:bg-slate-900/60">
      <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">
        {t("history.title")}
      </h2>

      <div className="scroll-slim mt-3 flex-1 space-y-2 overflow-y-auto pr-1">
        {query.isLoading &&
          Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="skeleton h-14 w-full rounded-xl" aria-hidden />
          ))}

        {query.isError && (
          <p className="text-xs text-red-500">{t("history.error")}</p>
        )}

        {query.data && query.data.length === 0 && (
          <p className="text-xs text-slate-400">{t("history.empty")}</p>
        )}

        <AnimatePresence initial={false}>
          {query.data?.map((s) => {
            const filters = activeFilters(s, t);
            return (
              <motion.div
                key={s.id}
                layout
                initial={{ opacity: 0, x: -8 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 8 }}
                transition={{ duration: 0.2 }}
                className="rounded-xl border border-slate-200 bg-white px-3 py-2.5 dark:border-slate-800 dark:bg-slate-900"
              >
                <div className="flex items-center justify-between gap-2">
                  <span className="truncate text-sm font-medium text-slate-800 dark:text-slate-100">
                    {s.regionName || t("history.region")}
                  </span>
                  <span className="shrink-0 rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-400">
                    {s.resultCount}
                  </span>
                </div>
                <div className="mt-1 flex items-center justify-between gap-2 text-xs text-slate-500 dark:text-slate-400">
                  <span className="truncate">
                    {filters.length ? filters.join(" · ") : t("history.noFilters")}
                  </span>
                  <span className="shrink-0">{formatDate(s.createdAt, lang)}</span>
                </div>
              </motion.div>
            );
          })}
        </AnimatePresence>
      </div>
    </div>
  );
}
