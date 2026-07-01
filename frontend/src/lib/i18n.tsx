"use client";

import { useEffect } from "react";
import { create } from "zustand";

export type Lang = "en" | "ru";

interface LangStore {
  lang: Lang;
  setLang: (lang: Lang) => void;
}

/** Language store. Defaults to "en" so server and first client render match;
 *  the saved choice is applied after mount via useLangInit (no hydration clash). */
export const useLangStore = create<LangStore>((set) => ({
  lang: "en",
  setLang: (lang) => {
    if (typeof window !== "undefined") {
      window.localStorage.setItem("lang", lang);
      document.documentElement.lang = lang;
    }
    set({ lang });
  },
}));

/** Read the persisted language once on mount. */
export function useLangInit(): void {
  const setLang = useLangStore((s) => s.setLang);
  useEffect(() => {
    const saved = window.localStorage.getItem("lang");
    if (saved === "ru" || saved === "en") setLang(saved);
  }, [setLang]);
}

type Dict = Record<string, string>;

const en: Dict = {
  tagline: "OpenStreetMap business finder",
  "tab.filters": "Filters",
  "tab.history": "History",

  contactGaps: "Contact gaps",
  "f.noWebsite.label": "No website",
  "f.noWebsite.desc": "Keep companies missing a website",
  "f.noSocials.label": "No socials",
  "f.noSocials.desc": "Keep companies with no social links",
  "f.noPhone.label": "No phone",
  "f.noPhone.desc": "Keep companies missing a phone number",

  categories: "Categories",
  clear: "Clear",
  categoriesError: "Could not load categories.",
  noCategories: "No categories available.",

  "search.idle": "Search companies",
  "search.pending": "Searching…",
  "search.noRegion": "Select a region first",
  "source.failed": "unavailable",

  "cat.shop": "Shops & retail",
  "cat.amenity": "Amenities (cafés, clinics, banks…)",
  "cat.office": "Offices",
  "cat.craft": "Craft & trades",
  "cat.tourism": "Tourism (hotels, attractions…)",

  "col.name": "Name",
  "col.category": "Category",
  "col.website": "Website",
  "col.phone": "Phone",
  "col.city": "City",
  "col.socials": "Socials",
  "col.contacts": "Contacts",
  "filter.email": "Has email",
  "filter.phone": "Has phone",
  "filter.website": "Has website",
  "filter.socials": "Has socials",

  "results.title": "Results",
  refresh: "Refresh (re-parse)",
  "results.emptyTitle": "No results yet",
  "results.emptyDesc":
    "Pick a region on the map, choose your filters, and run a search to see companies here.",
  pager: "Page {page} of {pages} · {count} companies",
  prev: "Previous",
  next: "Next",
  export: "Export",

  "history.title": "Search history",
  "history.error": "Could not load history.",
  "history.empty": "No searches yet.",
  "history.region": "Region",
  "history.noFilters": "no filters",
  "hfilter.noWebsite": "no website",
  "hfilter.noSocials": "no socials",
  "hfilter.noPhone": "no phone",
  "hfilter.categories": "{n} categories",

  "region.unknown": "Unknown region",

  "campaign.button": "Email campaign",
  "campaign.sendAll": "Email all filtered with an address ({count})",
  "campaign.title": "Email campaign",
  "campaign.warning":
    "Sends from your own SMTP. Only email recipients who have consented — unsolicited bulk email is regulated (CAN-SPAM / GDPR) and can get your domain blacklisted.",
  "campaign.recipients": "{count} recipients with email",
  "campaign.subject": "Subject",
  "campaign.body": "Message",
  "campaign.bodyHint": "Use {{name}} to insert the company name.",
  "campaign.consent": "I have consent to email these recipients and comply with anti-spam law.",
  "campaign.dryRun": "Dry run (preview, don't send)",
  "campaign.send": "Send campaign",
  "campaign.sendDry": "Preview (dry run)",
  "campaign.sending": "Sending…",
  "campaign.progress": "{sent} sent · {failed} failed / {total}",
  "campaign.doneSent": "Done — {sent} sent, {failed} failed.",
  "campaign.doneDry": "Dry run complete — would send to {sent}.",
  "campaign.close": "Close",
  "campaign.disabled": "Email sending is not configured on the server.",
};

const ru: Dict = {
  tagline: "Поиск бизнесов по OpenStreetMap",
  "tab.filters": "Фильтры",
  "tab.history": "История",

  contactGaps: "Контактные пробелы",
  "f.noWebsite.label": "Без сайта",
  "f.noWebsite.desc": "Оставить компании без сайта",
  "f.noSocials.label": "Без соцсетей",
  "f.noSocials.desc": "Оставить компании без соцсетей",
  "f.noPhone.label": "Без телефона",
  "f.noPhone.desc": "Оставить компании без телефона",

  categories: "Категории",
  clear: "Очистить",
  categoriesError: "Не удалось загрузить категории.",
  noCategories: "Нет доступных категорий.",

  "search.idle": "Искать компании",
  "search.pending": "Поиск…",
  "search.noRegion": "Сначала выберите регион",
  "source.failed": "недоступен",

  "cat.shop": "Магазины и ритейл",
  "cat.amenity": "Заведения (кафе, клиники, банки…)",
  "cat.office": "Офисы",
  "cat.craft": "Ремёсла и услуги",
  "cat.tourism": "Туризм (отели, достопримечательности…)",

  "col.name": "Название",
  "col.category": "Категория",
  "col.website": "Сайт",
  "col.phone": "Телефон",
  "col.city": "Город",
  "col.socials": "Соцсети",
  "col.contacts": "Контакты",
  "filter.email": "Есть почта",
  "filter.phone": "Есть телефон",
  "filter.website": "Есть сайт",
  "filter.socials": "Есть соцсети",

  "results.title": "Результаты",
  refresh: "Обновить (перепарсить)",
  "results.emptyTitle": "Пока нет результатов",
  "results.emptyDesc":
    "Выберите регион на карте, задайте фильтры и запустите поиск, чтобы увидеть компании здесь.",
  pager: "Стр. {page} из {pages} · {count} компаний",
  prev: "Назад",
  next: "Вперёд",
  export: "Экспорт",

  "history.title": "История поиска",
  "history.error": "Не удалось загрузить историю.",
  "history.empty": "Пока нет запросов.",
  "history.region": "Регион",
  "history.noFilters": "без фильтров",
  "hfilter.noWebsite": "без сайта",
  "hfilter.noSocials": "без соцсетей",
  "hfilter.noPhone": "без телефона",
  "hfilter.categories": "{n} категорий",

  "region.unknown": "Неизвестный регион",

  "campaign.button": "Email рассылка",
  "campaign.sendAll": "Разослать всем с почтой ({count})",
  "campaign.title": "Email рассылка",
  "campaign.warning":
    "Отправка с вашего SMTP. Пишите только тем, кто дал согласие — массовая рассылка без согласия регулируется законом (ФЗ «О рекламе», GDPR) и грозит блокировкой домена.",
  "campaign.recipients": "{count} получателей с почтой",
  "campaign.subject": "Тема",
  "campaign.body": "Сообщение",
  "campaign.bodyHint": "Вставьте {{name}} для имени компании.",
  "campaign.consent": "У меня есть согласие получателей, соблюдаю закон о рекламе/спаме.",
  "campaign.dryRun": "Тест (превью, без отправки)",
  "campaign.send": "Отправить рассылку",
  "campaign.sendDry": "Превью (тест)",
  "campaign.sending": "Отправка…",
  "campaign.progress": "{sent} отправлено · {failed} ошибок / {total}",
  "campaign.doneSent": "Готово — отправлено {sent}, ошибок {failed}.",
  "campaign.doneDry": "Тест завершён — отправили бы {sent}.",
  "campaign.close": "Закрыть",
  "campaign.disabled": "Отправка email не настроена на сервере.",
};

const messages: Record<Lang, Dict> = { en, ru };

export type TFunc = (key: string, vars?: Record<string, string | number>) => string;

/** Hook returning a translate function bound to the current language. */
export function useT(): TFunc {
  const lang = useLangStore((s) => s.lang);
  return (key, vars) => {
    let out = messages[lang][key] ?? messages.en[key] ?? key;
    if (vars) {
      for (const [k, v] of Object.entries(vars)) {
        out = out.replace(`{${k}}`, String(v));
      }
    }
    return out;
  };
}
