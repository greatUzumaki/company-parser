"use client";

import { useMemo } from "react";
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { motion } from "framer-motion";
import type { Company } from "@/lib/types";
import { useT, type TFunc } from "@/lib/i18n";

const columnHelper = createColumnHelper<Company>();

/** Small coloured chip listing which socials a company has. */
function Socials({ company }: { company: Company }) {
  const items: { label: string; present: boolean }[] = [
    { label: "IG", present: !!company.instagram },
    { label: "FB", present: !!company.facebook },
    { label: "VK", present: !!company.vk },
    { label: "TG", present: !!company.telegram },
  ];
  const active = items.filter((i) => i.present);
  if (active.length === 0)
    return <span className="text-slate-300 dark:text-slate-600">—</span>;
  return (
    <div className="flex flex-wrap gap-1">
      {active.map((i) => (
        <span
          key={i.label}
          className="rounded bg-indigo-100 px-1.5 py-0.5 text-[10px] font-semibold text-indigo-700 dark:bg-indigo-500/20 dark:text-indigo-300"
        >
          {i.label}
        </span>
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
  columnHelper.accessor(
    (row) => row.subcategory || row.category,
    {
      id: "category",
      header: t("col.category"),
      cell: (info) => (
        <span className="capitalize text-slate-600 dark:text-slate-300">
          {info.getValue() || "—"}
        </span>
      ),
    },
  ),
  columnHelper.accessor("website", {
    header: t("col.website"),
    cell: (info) => {
      const url = info.getValue();
      if (!url) return <span className="text-slate-300 dark:text-slate-600">—</span>;
      const href = url.startsWith("http") ? url : `https://${url}`;
      return (
        <a
          href={href}
          target="_blank"
          rel="noopener noreferrer"
          className="text-emerald-600 hover:underline dark:text-emerald-400"
        >
          {url.replace(/^https?:\/\//, "")}
        </a>
      );
    },
  }),
  columnHelper.accessor("phone", {
    header: t("col.phone"),
    cell: (info) => info.getValue() || <span className="text-slate-300 dark:text-slate-600">—</span>,
  }),
  columnHelper.accessor((row) => row.addr.city, {
    id: "city",
    header: t("col.city"),
    cell: (info) => info.getValue() || <span className="text-slate-300 dark:text-slate-600">—</span>,
  }),
  columnHelper.display({
    id: "socials",
    header: t("col.socials"),
    cell: (info) => <Socials company={info.row.original} />,
  }),
];

export function ResultsTable({
  data,
  isLoading,
}: {
  data: Company[];
  isLoading: boolean;
}) {
  const t = useT();
  const columns = useMemo(() => makeColumns(t), [t]);
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
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
    <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm dark:border-slate-800 dark:bg-slate-900">
      <div className="scroll-slim overflow-x-auto">
        <table className="w-full border-collapse text-sm">
          <thead>
            {table.getHeaderGroups().map((hg) => (
              <tr
                key={hg.id}
                className="border-b border-slate-200 bg-slate-50 dark:border-slate-800 dark:bg-slate-800/60"
              >
                {hg.headers.map((header) => (
                  <th
                    key={header.id}
                    className="whitespace-nowrap px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400"
                  >
                    {flexRender(
                      header.column.columnDef.header,
                      header.getContext(),
                    )}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {table.getRowModel().rows.map((row, i) => (
              <motion.tr
                key={row.id}
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.25, delay: Math.min(i * 0.03, 0.3) }}
                className="border-b border-slate-100 transition-colors last:border-0 hover:bg-emerald-50/50 dark:border-slate-800 dark:hover:bg-emerald-500/5"
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
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-slate-200 px-4 py-3 text-sm dark:border-slate-800">
        <span className="text-slate-500 dark:text-slate-400">
          {t("pager", {
            page: table.getState().pagination.pageIndex + 1,
            pages: table.getPageCount(),
            count: data.length,
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
  );
}
