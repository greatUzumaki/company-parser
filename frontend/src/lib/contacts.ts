import type { Company } from "./types";

/** Normalize a website value to a safe http(s) URL, or null. Rejects any other
 *  scheme (javascript:, data:, …) so it can never execute on click. */
export function toHttpUrl(raw?: string): string | null {
  if (!raw) return null;
  const candidate = /^https?:\/\//i.test(raw) ? raw : `https://${raw}`;
  try {
    const u = new URL(candidate);
    return u.protocol === "http:" || u.protocol === "https:" ? u.toString() : null;
  } catch {
    return null;
  }
}

/** Resolve a contact value into a safe href (tel:/mailto:/https:) or null. */
export function contactHref(kind: string, raw: string): string | null {
  const v = raw.trim();
  if (!v) return null;
  const handle = v.replace(/^@/, "");
  const isUrl = /^https?:\/\//i.test(v);
  switch (kind) {
    case "phone":
      return `tel:${v.replace(/[^\d+]/g, "")}`;
    case "email":
      return `mailto:${v}`;
    case "website":
      return toHttpUrl(v);
    case "telegram":
      return isUrl ? toHttpUrl(v) : `https://t.me/${handle}`;
    case "whatsapp": {
      const digits = v.replace(/[^\d]/g, "");
      return isUrl ? toHttpUrl(v) : digits ? `https://wa.me/${digits}` : null;
    }
    case "instagram":
      return isUrl ? toHttpUrl(v) : `https://instagram.com/${handle}`;
    case "facebook":
      return isUrl ? toHttpUrl(v) : `https://facebook.com/${handle}`;
    case "vk":
      return isUrl ? toHttpUrl(v) : `https://vk.com/${handle}`;
    default:
      return null;
  }
}

/** Contact channels in display order, with short labels for chips. */
export const CONTACT_CHIPS: { kind: string; label: string }[] = [
  { kind: "phone", label: "☎" },
  { kind: "email", label: "✉" },
  { kind: "telegram", label: "TG" },
  { kind: "whatsapp", label: "WA" },
  { kind: "instagram", label: "IG" },
  { kind: "facebook", label: "FB" },
  { kind: "vk", label: "VK" },
  { kind: "website", label: "🌐" },
];

export function hasSocials(c: Company): boolean {
  return !!(c.instagram || c.facebook || c.vk || c.telegram || c.whatsapp);
}

/** Read a contact channel's raw value off a company. */
export function contactValue(c: Company, kind: string): string {
  switch (kind) {
    case "phone":
      return c.phone;
    case "email":
      return c.email;
    case "website":
      return c.website;
    case "telegram":
      return c.telegram;
    case "whatsapp":
      return c.whatsapp;
    case "instagram":
      return c.instagram;
    case "facebook":
      return c.facebook;
    case "vk":
      return c.vk;
    default:
      return "";
  }
}
