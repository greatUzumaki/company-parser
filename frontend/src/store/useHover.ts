import { create } from "zustand";

/** Shared hover state linking the results list and the map: the key of the
 *  company currently hovered (`${osmType}/${osmId}`), or null. */
interface HoverState {
  hoveredId: string | null;
  setHovered: (id: string | null) => void;
}

export const useHover = create<HoverState>((set) => ({
  hoveredId: null,
  setHovered: (hoveredId) => set({ hoveredId }),
}));

/** Stable per-company key used both as the map feature id and the row id. */
export function companyKey(osmType: string, osmId: string): string {
  return `${osmType}/${osmId}`;
}
