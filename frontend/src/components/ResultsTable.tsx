"use client";

import { useMemo, useState } from "react";
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
  type SortingState,
} from "@tanstack/react-table";
import { motion } from "framer-motion";
import type { Company } from "@/lib/types";
import { useT, type TFunc } from "@/lib/i18n";
import { useHover, companyKey } from "@/store/useHover";
import { CONTACT_CHIPS, contactHref, contactValue, hasSocials } from "@/lib/contacts";
import { CampaignModal } from "@/components/CampaignModal";

const columnHelper = createColumnHelper<Company>();

/** Clickable contact chips for a company — each a safe tel:/mailto:/https link. */
function ContactChips({ company }: { company: Company }) {
  const chips = CONTACT_CHIPS.map(({ kind, label }) => {
    const href = contactHref(kind, contactValue(company, kind));
    return href ? { kind, label, href, val: contactValue(company, kind) } : null;
  }).filter((c): c is { kind: string; label: string; href: string; val: string } => c !== null);

  if (chips.length === 0)
    return <span className="text-slate-300 dark:text-slate-600">—</span>;

  return (
    <div className="flex flex-wrap gap-1">
      {chips.map((c) => (
        <a
          key={c.kind}
          href={c.href}
          target="_blank"
          rel="noopener noreferrer"
          title={c.val}
          onClick={(e) => e.stopPropagation()}
          className="rounded-full bg-emerald-50 px-1.5 py-0.5 text-[10px] font-semibold text-emerald-700 transition-colors hover:bg-emerald-100 dark:bg-emerald-500/15 dark:text-emerald-300"
        >
          {c.label}
        </a>
      ))}
    </div>
  );
}

const makeColumns = (t: TFunc) => [
  columnHelper.accessor("name", {
    header: t("col.name"),
    cell: (info) => (
      <span className="font-medium text-slate-900 dark:text-slate-100">
        {info.getValue() || "—"}
      </span>
    ),
  }),
  columnHelper.accessor((row) => row.subcategory || row.category, {
    id: "category",
    header: t("col.category"),
    cell: (info) => (
      <span className="capitalize text-slate-600 dark:text-slate-300">
        {info.getValue() || "—"}
      </span>
    ),
  }),
  columnHelper.accessor((row) => row.addr.city, {
    id: "city",
    header: t("col.city"),
    cell: (info) => info.getValue() || <span className="text-slate-300 dark:text-slate-600">—</span>,
  }),
  columnHelper.display({
    id: "contacts",
    header: t("col.contacts"),
    cell: (info) => <ContactChips company={info.row.original} />,
  }),
];

/** Contact-presence filters shown as toggle chips above the table. */
const FILTERS: { key: string; test: (c: Company) => boolean }[] = [
  { key: "email", test: (c) => !!c.email },
  { key: "phone", test: (c) => !!c.phone },
  { key: "website", test: (c) => !!c.website },
  { key: "socials", test: hasSocials },
];

export function ResultsTable({
  data,
  isLoading,
  campaignEnabled = false,
}: {
  data: Company[];
  isLoading: boolean;
  campaignEnabled?: boolean;
}) {
  const t = useT();
  const setHovered = useHover((s) => s.setHovered);
  const [sorting, setSorting] = useState<SortingState>([]);
  const [active, setActive] = useState<Record<string, boolean>>({});
  const [campaignOpen, setCampaignOpen] = useState(false);

  const columns = useMemo(() => makeColumns(t), [t]);

  const filtered = useMemo(() => {
    const on = FILTERS.filter((f) => active[f.key]);
    if (on.length === 0) return data;
    return data.filter((c) => on.every((f) => f.test(c)));
  }, [data, active]);

  // Recipients = the currently-filtered rows that have an email.
  const recipients = useMemo(
    () => filtered.filter((c) => c.email).map((c) => ({ email: c.email, name: c.name })),
    [filtered],
  );

  const table = useReactTable({
    data: filtered,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    initialState: { pagination: { pageSize: 10 } },
  });

  if (isLoading) {
    return (
      <div className="overflow-hidden rounded-2xl border border-slate-200 dark:border-slate-800">
        <div className="divide-y divide-slate-100 dark:divide-slate-800">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="flex items-center gap-4 px-4 py-3.5">
              <span className="skeleton h-4 w-1/4 rounded" />
              <span className="skeleton h-4 w-1/6 rounded" />
              <span className="skeleton h-4 w-1/5 rounded" />
              <span className="skeleton h-4 w-1/6 rounded" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (data.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center gap-2 rounded-2xl border border-dashed border-slate-300 px-6 py-16 text-center dark:border-slate-700">
        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-emerald-100 text-emerald-600 dark:bg-emerald-500/15 dark:text-emerald-400">
          <svg viewBox="0 0 24 24" fill="none" className="h-6 w-6" stroke="currentColor" strokeWidth={1.8}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-4.3-4.3m1.8-5.2a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
        <p className="text-sm font-medium text-slate-700 dark:text-slate-200">
          {t("results.emptyTitle")}
        </p>
        <p className="max-w-xs text-xs text-slate-500 dark:text-slate-400">
          {t("results.emptyDesc")}
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {/* Contact filters */}
      <div className="flex flex-wrap items-center gap-2">
        {FILTERS.map((f) => {
          const on = !!active[f.key];
          return (
            <button
              key={f.key}
              type="button"
              aria-pressed={on}
              onClick={() => setActive((prev) => ({ ...prev, [f.key]: !prev[f.key] }))}
              className={`rounded-full border px-3 py-1 text-xs font-medium transition-all focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 ${
                on
                  ? "border-emerald-500 bg-emerald-500 text-white shadow-sm"
                  : "border-slate-200 bg-white text-slate-600 hover:border-emerald-300 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-300"
              }`}
            >
              {t(`filter.${f.key}`)}
            </button>
          );
        })}
        {filtered.length !== data.length && (
          <span className="text-xs text-slate-500 dark:text-slate-400">
            {filtered.length} / {data.length}
          </span>
        )}
      </div>

      <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm dark:border-slate-800 dark:bg-slate-900">
        <div className="scroll-slim overflow-x-auto">
          <table className="w-full border-collapse text-sm">
            <thead>
              {table.getHeaderGroups().map((hg) => (
                <tr
                  key={hg.id}
                  className="border-b border-slate-200 bg-slate-50 dark:border-slate-800 dark:bg-slate-800/60"
                >
                  {hg.headers.map((header) => {
                    const canSort = header.column.getCanSort();
                    const dir = header.column.getIsSorted();
                    return (
                      <th
                        key={header.id}
                        onClick={canSort ? header.column.getToggleSortingHandler() : undefined}
                        className={`whitespace-nowrap px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400 ${
                          canSort ? "cursor-pointer select-none hover:text-slate-700 dark:hover:text-slate-200" : ""
                        }`}
                      >
                        <span className="inline-flex items-center gap-1">
                          {flexRender(header.column.columnDef.header, header.getContext())}
                          {dir === "asc" && <span aria-hidden>▲</span>}
                          {dir === "desc" && <span aria-hidden>▼</span>}
                        </span>
                      </th>
                    );
                  })}
                </tr>
              ))}
            </thead>
            <tbody>
              {table.getRowModel().rows.map((row, i) => {
                const c = row.original;
                const key = companyKey(c.osmType, c.osmId);
                return (
                  <motion.tr
                    key={row.id}
                    initial={{ opacity: 0, y: 6 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.25, delay: Math.min(i * 0.03, 0.3) }}
                    onMouseEnter={() => setHovered(key)}
                    onMouseLeave={() => setHovered(null)}
                    className="border-b border-slate-100 transition-colors last:border-0 hover:bg-emerald-50/60 dark:border-slate-800 dark:hover:bg-emerald-500/5"
                  >
                    {row.getVisibleCells().map((cell) => (
                      <td
                        key={cell.id}
                        className="max-w-[220px] truncate px-4 py-3 align-middle text-slate-700 dark:text-slate-300"
                      >
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </td>
                    ))}
                  </motion.tr>
                );
              })}
            </tbody>
          </table>
        </div>

        <div className="flex flex-wrap items-center justify-between gap-3 border-t border-slate-200 px-4 py-3 text-sm dark:border-slate-800">
          <span className="text-slate-500 dark:text-slate-400">
            {t("pager", {
              page: table.getState().pagination.pageIndex + 1,
              pages: table.getPageCount() || 1,
              count: filtered.length,
            })}
          </span>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => table.previousPage()}
              disabled={!table.getCanPreviousPage()}
              className="rounded-lg border border-slate-200 px-3 py-1.5 font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
            >
              {t("prev")}
            </button>
            <button
              type="button"
              onClick={() => table.nextPage()}
              disabled={!table.getCanNextPage()}
              className="rounded-lg border border-slate-200 px-3 py-1.5 font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
            >
              {t("next")}
            </button>
          </div>
        </div>
      </div>

      {recipients.length > 0 && (
        <button
          type="button"
          onClick={() => setCampaignOpen(true)}
          className="flex w-full items-center justify-center gap-2 rounded-2xl bg-indigo-600 px-4 py-3 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-indigo-500"
        >
          ✉ {t("campaign.sendAll", { count: recipients.length })}
        </button>
      )}

      {campaignOpen && (
        <CampaignModal
          recipients={recipients}
          enabled={campaignEnabled}
          onClose={() => setCampaignOpen(false)}
        />
      )}
    </div>
  );
}
