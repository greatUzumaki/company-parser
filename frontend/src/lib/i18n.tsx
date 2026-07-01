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

  "results.title": "Results",
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

  "results.title": "Результаты",
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
