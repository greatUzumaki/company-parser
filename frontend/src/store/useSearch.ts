import { create } from "zustand";
import type { Filters, Region } from "@/lib/types";

/** Keys of the boolean "gap" filters. */
export type BoolFilterKey = "noWebsite" | "noSocials" | "noPhone";

interface SearchState {
  region: Region | null;
  filters: Filters;
  setRegion: (region: Region) => void;
  toggleFilter: (key: BoolFilterKey) => void;
  setCategories: (categories: string[]) => void;
}

const defaultFilters: Filters = {
  noWebsite: false,
  noSocials: false,
  noPhone: false,
  categories: [],
};

/** Global store holding the selected region and active filters. */
export const useSearch = create<SearchState>((set) => ({
  region: null,
  filters: defaultFilters,
  setRegion: (region) => set({ region }),
  toggleFilter: (key) =>
    set((s) => ({ filters: { ...s.filters, [key]: !s.filters[key] } })),
  setCategories: (categories) =>
    set((s) => ({ filters: { ...s.filters, categories } })),
}));
