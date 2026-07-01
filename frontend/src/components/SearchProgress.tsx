"use client";

import { AnimatePresence, motion } from "framer-motion";
import { useT } from "@/lib/i18n";

export interface SourceProgress {
  name: string;
  status: "pending" | "done" | "failed";
  count: number;
}

/** A thin strip under the results header showing each data source's live
 *  progress: spinner while fetching, count when done, a warning if it failed. */
export function SearchProgress({
  sources,
  streaming,
}: {
  sources: SourceProgress[];
  streaming: boolean;
}) {
  const t = useT();
  if (sources.length === 0) return null;

  const done = sources.filter((s) => s.status !== "pending").length;
  const pct = Math.round((done / sources.length) * 100);

  return (
    <div className="border-b border-slate-200/70 px-4 py-2 dark:border-slate-700/70">
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
        {sources.map((s) => (
          <span key={s.name} className="flex items-center gap-1.5 text-xs">
            <StatusDot status={s.status} />
            <span className="font-medium text-slate-600 dark:text-slate-300">{s.name}</span>
            {s.status === "failed" ? (
              <span className="text-amber-600 dark:text-amber-400">{t("source.failed")}</span>
            ) : (
              <span className="tabular-nums text-slate-400">{s.count}</span>
            )}
          </span>
        ))}
      </div>

      <div className="mt-2 h-1 overflow-hidden rounded-full bg-slate-200 dark:bg-slate-800">
        <motion.div
          className="h-full rounded-full bg-emerald-500"
          initial={{ width: 0 }}
          animate={{ width: `${pct}%` }}
          transition={{ type: "spring", stiffness: 120, damping: 20 }}
        />
      </div>

      <AnimatePresence>
        {streaming && (
          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="mt-1 text-[11px] text-slate-400"
          >
            {t("search.pending")}
          </motion.p>
        )}
      </AnimatePresence>
    </div>
  );
}

function StatusDot({ status }: { status: SourceProgress["status"] }) {
  if (status === "pending") {
    return (
      <span
        className="h-3 w-3 animate-spin rounded-full border-2 border-emerald-500/40 border-t-emerald-500"
        aria-hidden
      />
    );
  }
  if (status === "failed") {
    return <span className="h-2.5 w-2.5 rounded-full bg-amber-500" aria-hidden />;
  }
  return <span className="h-2.5 w-2.5 rounded-full bg-emerald-500" aria-hidden />;
}
