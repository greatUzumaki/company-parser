"use client";

import { motion } from "framer-motion";
import { exportUrl } from "@/lib/api";
import type { ExportFormat } from "@/lib/types";
import { useT } from "@/lib/i18n";

const FORMATS: { fmt: ExportFormat; label: string }[] = [
  { fmt: "json", label: "JSON" },
  { fmt: "csv", label: "CSV" },
  { fmt: "xlsx", label: "Excel" },
];

export function ExportBar({ searchId }: { searchId: number | null }) {
  const t = useT();
  const disabled = searchId === null;

  const download = (fmt: ExportFormat) => {
    if (searchId === null) return;
    // Navigate to the export endpoint to trigger a file download.
    window.location.assign(exportUrl(searchId, fmt));
  };

  return (
    <div className="flex items-center gap-2">
      <span className="mr-1 hidden text-xs font-medium uppercase tracking-wide text-slate-400 sm:inline">
        {t("export")}
      </span>
      {FORMATS.map(({ fmt, label }) => (
        <motion.button
          key={fmt}
          type="button"
          whileHover={!disabled ? { scale: 1.04 } : undefined}
          whileTap={!disabled ? { scale: 0.95 } : undefined}
          disabled={disabled}
          onClick={() => download(fmt)}
          aria-label={`Export as ${label}`}
          className="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 shadow-sm transition-colors hover:border-emerald-400 hover:text-emerald-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:border-slate-200 disabled:hover:text-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200 dark:hover:border-emerald-600"
        >
          {label}
        </motion.button>
      ))}
    </div>
  );
}
