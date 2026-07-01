"use client";

import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { getCategories } from "@/lib/api";
import { useSearch, type BoolFilterKey } from "@/store/useSearch";
import { useT } from "@/lib/i18n";

/** An accessible, animated on/off switch. */
function Switch({
  checked,
  onChange,
  label,
  description,
}: {
  checked: boolean;
  onChange: () => void;
  label: string;
  description: string;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={label}
      onClick={onChange}
      className="group flex w-full items-center justify-between gap-3 rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-left transition-colors hover:border-emerald-300 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 dark:border-slate-800 dark:bg-slate-900 dark:hover:border-emerald-700"
    >
      <span className="min-w-0">
        <span className="block text-sm font-medium text-slate-800 dark:text-slate-100">
          {label}
        </span>
        <span className="block truncate text-xs text-slate-500 dark:text-slate-400">
          {description}
        </span>
      </span>
      <span
        className={`relative inline-flex h-6 w-11 shrink-0 items-center rounded-full transition-colors ${
          checked ? "bg-emerald-500" : "bg-slate-300 dark:bg-slate-700"
        }`}
      >
        <motion.span
          animate={{ x: checked ? 22 : 2 }}
          transition={{ type: "spring", stiffness: 600, damping: 32 }}
          className="absolute left-0 top-0.5 h-5 w-5 rounded-full bg-white shadow"
        />
      </span>
    </button>
  );
}

const TOGGLE_KEYS: BoolFilterKey[] = ["noWebsite", "noSocials", "noPhone"];

export function FilterPanel({
  onSearch,
  isPending,
}: {
  onSearch: () => void;
  isPending: boolean;
}) {
  const t = useT();
  const region = useSearch((s) => s.region);
  const filters = useSearch((s) => s.filters);
  const toggleFilter = useSearch((s) => s.toggleFilter);
  const setCategories = useSearch((s) => s.setCategories);

  const categoriesQuery = useQuery({
    queryKey: ["categories"],
    queryFn: getCategories,
  });

  const toggleCategory = (key: string) => {
    const set = new Set(filters.categories);
    if (set.has(key)) set.delete(key);
    else set.add(key);
    setCategories([...set]);
  };

  return (
    <div className="flex h-full flex-col gap-5 rounded-2xl border border-slate-200 bg-white/70 p-4 shadow-sm backdrop-blur sm:p-5 dark:border-slate-800 dark:bg-slate-900/60">
      <div>
        <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">
          {t("contactGaps")}
        </h2>
        <div className="mt-3 flex flex-col gap-2">
          {TOGGLE_KEYS.map((key) => (
            <Switch
              key={key}
              checked={filters[key]}
              onChange={() => toggleFilter(key)}
              label={t(`f.${key}.label`)}
              description={t(`f.${key}.desc`)}
            />
          ))}
        </div>
      </div>

      <div className="min-h-0 flex-1">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">
            {t("categories")}
          </h2>
          {filters.categories.length > 0 && (
            <button
              type="button"
              onClick={() => setCategories([])}
              className="text-xs font-medium text-emerald-600 hover:underline dark:text-emerald-400"
            >
              {t("clear")}
            </button>
          )}
        </div>

        <div className="scroll-slim mt-3 flex max-h-44 flex-wrap gap-2 overflow-y-auto pr-1">
          {categoriesQuery.isLoading &&
            Array.from({ length: 8 }).map((_, i) => (
              <span
                key={i}
                className="skeleton h-7 w-20 rounded-full"
                aria-hidden
              />
            ))}

          {categoriesQuery.isError && (
            <p className="text-xs text-red-500">{t("categoriesError")}</p>
          )}

          {categoriesQuery.data?.map((c) => {
            const active = filters.categories.includes(c.key);
            return (
              <button
                key={c.key}
                type="button"
                aria-pressed={active}
                onClick={() => toggleCategory(c.key)}
                className={`rounded-full border px-3 py-1 text-xs font-medium transition-all focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 ${
                  active
                    ? "border-emerald-500 bg-emerald-500 text-white shadow-sm"
                    : "border-slate-200 bg-white text-slate-700 hover:border-emerald-300 hover:text-emerald-700 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200"
                }`}
              >
                {t(`cat.${c.key}`) || c.label}
              </button>
            );
          })}

          {categoriesQuery.data?.length === 0 && (
            <p className="text-xs text-slate-400">{t("noCategories")}</p>
          )}
        </div>
      </div>

      <motion.button
        type="button"
        whileHover={region && !isPending ? { scale: 1.02 } : undefined}
        whileTap={region && !isPending ? { scale: 0.97 } : undefined}
        disabled={!region || isPending}
        onClick={() => onSearch()}
        className="inline-flex items-center justify-center gap-2 rounded-xl bg-emerald-600 px-4 py-3 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-emerald-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:bg-slate-300 disabled:text-slate-500 dark:focus-visible:ring-offset-slate-900 dark:disabled:bg-slate-800 dark:disabled:text-slate-500"
      >
        {isPending ? (
          <>
            <span
              className="h-4 w-4 animate-spin rounded-full border-2 border-white/40 border-t-white"
              aria-hidden
            />
            {t("search.pending")}
          </>
        ) : region ? (
          t("search.idle")
        ) : (
          t("search.noRegion")
        )}
      </motion.button>
    </div>
  );
}
