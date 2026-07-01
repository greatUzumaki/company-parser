"use client";

import { useRef, useState } from "react";
import { motion } from "framer-motion";
import { sendCampaign } from "@/lib/api";
import type { CampaignEvent } from "@/lib/types";
import { useT } from "@/lib/i18n";

interface Progress {
  sent: number;
  failed: number;
  total: number;
}

export function CampaignModal({
  recipients,
  enabled,
  onClose,
}: {
  recipients: { email: string; name: string }[];
  enabled: boolean;
  onClose: () => void;
}) {
  const t = useT();
  const recipientCount = recipients.length;
  const [subject, setSubject] = useState("");
  const [body, setBody] = useState("");
  const [dryRun, setDryRun] = useState(true);
  const [consent, setConsent] = useState(false);
  const [sending, setSending] = useState(false);
  const [progress, setProgress] = useState<Progress | null>(null);
  const [doneMsg, setDoneMsg] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const canSend =
    enabled && !sending && consent && subject.trim() !== "" && body.trim() !== "" && recipientCount > 0;

  const run = () => {
    if (!canSend) return;
    abortRef.current?.abort();
    const ac = new AbortController();
    abortRef.current = ac;
    setSending(true);
    setDoneMsg(null);
    setError(null);
    setProgress({ sent: 0, failed: 0, total: recipientCount });

    sendCampaign(
      { subject, body, recipients, dryRun, confirm: consent },
      (e: CampaignEvent) => {
        switch (e.type) {
          case "start":
            setProgress({ sent: 0, failed: 0, total: e.total ?? recipientCount });
            break;
          case "sent":
          case "failed":
            setProgress((p) => ({
              sent: e.sent ?? p?.sent ?? 0,
              failed: e.failed ?? p?.failed ?? 0,
              total: p?.total ?? recipientCount,
            }));
            break;
          case "done":
            setSending(false);
            setDoneMsg(
              e.dryRun
                ? t("campaign.doneDry", { sent: e.sent ?? 0 })
                : t("campaign.doneSent", { sent: e.sent ?? 0, failed: e.failed ?? 0 }),
            );
            break;
          case "error":
            setSending(false);
            setError(e.message ?? "failed");
            break;
        }
      },
      ac.signal,
    ).catch((err) => {
      if (ac.signal.aborted) return;
      setSending(false);
      setError(err instanceof Error ? err.message : "failed");
    });
  };

  return (
    <div className="fixed inset-0 z-40 flex items-center justify-center p-4">
      <button
        type="button"
        aria-label={t("campaign.close")}
        onClick={onClose}
        className="absolute inset-0 bg-slate-900/40 backdrop-blur-sm"
      />
      <motion.div
        initial={{ opacity: 0, scale: 0.96, y: 10 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        className="relative z-10 flex max-h-[90vh] w-full max-w-lg flex-col overflow-hidden rounded-2xl bg-white shadow-2xl dark:bg-slate-900"
      >
        <div className="flex items-center justify-between border-b border-slate-200 px-5 py-3 dark:border-slate-800">
          <h2 className="text-base font-semibold text-slate-900 dark:text-slate-50">
            {t("campaign.title")}
          </h2>
          <button
            type="button"
            onClick={onClose}
            aria-label={t("campaign.close")}
            className="flex h-8 w-8 items-center justify-center rounded-lg text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-700 dark:hover:bg-slate-800"
          >
            ✕
          </button>
        </div>

        <div className="scroll-slim flex-1 space-y-4 overflow-y-auto px-5 py-4">
          <div className="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2.5 text-xs text-amber-800 dark:border-amber-900/50 dark:bg-amber-950/30 dark:text-amber-300">
            ⚠️ {t("campaign.warning")}
          </div>

          {!enabled && (
            <div className="rounded-xl border border-red-200 bg-red-50 px-3 py-2.5 text-xs text-red-700 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-300">
              {t("campaign.disabled")}
            </div>
          )}

          <p className="text-sm font-medium text-slate-700 dark:text-slate-200">
            {t("campaign.recipients", { count: recipientCount })}
          </p>

          <label className="block">
            <span className="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">
              {t("campaign.subject")}
            </span>
            <input
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              disabled={sending}
              className="mt-1 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 focus:border-emerald-400 focus:outline-none focus:ring-2 focus:ring-emerald-500/30 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100"
            />
          </label>

          <label className="block">
            <span className="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">
              {t("campaign.body")}
            </span>
            <textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              disabled={sending}
              rows={7}
              className="mt-1 w-full resize-y rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 focus:border-emerald-400 focus:outline-none focus:ring-2 focus:ring-emerald-500/30 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100"
            />
            <span className="mt-1 block text-xs text-slate-400">{t("campaign.bodyHint")}</span>
          </label>

          <label className="flex items-start gap-2 text-sm text-slate-700 dark:text-slate-300">
            <input
              type="checkbox"
              checked={consent}
              onChange={(e) => setConsent(e.target.checked)}
              disabled={sending}
              className="mt-0.5 h-4 w-4 accent-emerald-600"
            />
            <span>{t("campaign.consent")}</span>
          </label>

          <label className="flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300">
            <input
              type="checkbox"
              checked={dryRun}
              onChange={(e) => setDryRun(e.target.checked)}
              disabled={sending}
              className="h-4 w-4 accent-emerald-600"
            />
            <span>{t("campaign.dryRun")}</span>
          </label>

          {progress && (
            <div className="rounded-xl bg-slate-50 px-3 py-2 text-sm dark:bg-slate-800/60">
              <div className="flex justify-between text-slate-600 dark:text-slate-300">
                <span>
                  {t("campaign.progress", {
                    sent: progress.sent,
                    failed: progress.failed,
                    total: progress.total,
                  })}
                </span>
              </div>
              <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-slate-200 dark:bg-slate-700">
                <div
                  className="h-full rounded-full bg-emerald-500 transition-all"
                  style={{
                    width: `${progress.total ? Math.round(((progress.sent + progress.failed) / progress.total) * 100) : 0}%`,
                  }}
                />
              </div>
            </div>
          )}

          {doneMsg && (
            <p className="rounded-xl bg-emerald-50 px-3 py-2 text-sm font-medium text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300">
              {doneMsg}
            </p>
          )}
          {error && (
            <p className="rounded-xl bg-red-50 px-3 py-2 text-sm text-red-700 dark:bg-red-950/40 dark:text-red-300">
              {error}
            </p>
          )}
        </div>

        <div className="flex items-center justify-end gap-2 border-t border-slate-200 px-5 py-3 dark:border-slate-800">
          <button
            type="button"
            onClick={onClose}
            className="rounded-xl px-4 py-2 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800"
          >
            {t("campaign.close")}
          </button>
          <button
            type="button"
            onClick={run}
            disabled={!canSend}
            className="inline-flex items-center gap-2 rounded-xl bg-emerald-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-emerald-500 disabled:cursor-not-allowed disabled:bg-slate-300 disabled:text-slate-500 dark:disabled:bg-slate-800 dark:disabled:text-slate-500"
          >
            {sending && (
              <span className="h-4 w-4 animate-spin rounded-full border-2 border-white/40 border-t-white" aria-hidden />
            )}
            {sending ? t("campaign.sending") : dryRun ? t("campaign.sendDry") : t("campaign.send")}
          </button>
        </div>
      </motion.div>
    </div>
  );
}
