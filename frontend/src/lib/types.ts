// Frontend mirrors of the backend DTOs. Keep these in lockstep with the Go
// structs in backend/internal/domain and backend/internal/store.

/** A search area selected on the map. bbox is [minLon, minLat, maxLon, maxLat].
 *  polygon (a drawn circle/rectangle/freehand zone) is a ring of [lon, lat]. */
export interface Region {
  name: string;
  osmAreaId: number;
  bbox: [number, number, number, number];
  polygon?: [number, number][];
}

/** Search filters. The booleans are "gap" filters: true keeps companies that
 *  LACK the corresponding contact info. Empty categories matches everything. */
export interface Filters {
  noWebsite: boolean;
  noSocials: boolean;
  noPhone: boolean;
  categories: string[];
}

/** Structured postal address from addr:* tags. */
export interface Address {
  country: string;
  city: string;
  street: string;
  housenumber: string;
  postcode: string;
}

/** A business found in the data source. */
export interface Company {
  osmType: string;
  osmId: string;
  name: string;
  category: string;
  subcategory: string;
  website: string;
  phone: string;
  email: string;
  instagram: string;
  facebook: string;
  vk: string;
  telegram: string;
  whatsapp: string;
  addr: Address;
  lat: number;
  lon: number;
  openingHours: string;
}

/** Response from POST /api/v1/search. */
export interface SearchResponse {
  searchId: number;
  count: number;
  results: Company[];
}

/** One row of search history. */
export interface SearchSummary {
  id: number;
  regionName: string;
  regionAreaId: number;
  filters: Filters;
  resultCount: number;
  createdAt: string;
}

/** A selectable company category. */
export interface Category {
  key: string;
  label: string;
}

/** One event in a streaming email campaign (NDJSON from /campaign/send). */
export interface CampaignEvent {
  type: "start" | "sent" | "failed" | "done" | "error";
  email?: string;
  name?: string;
  sent?: number;
  failed?: number;
  total?: number;
  dryRun?: boolean;
  message?: string;
}

/** One event in a streaming search (NDJSON line from /search/stream). */
export interface StreamEvent {
  type: "cached" | "source_start" | "source_done" | "companies" | "done" | "error";
  source?: string;
  companies?: Company[];
  count?: number; // running total after dedup
  done?: number; // providers finished
  total?: number; // providers total
  searchId?: number; // set on "done"
  cached?: boolean; // served from cache without a refresh
  message?: string;
}

/** Supported export formats. */
export type ExportFormat = "json" | "csv" | "xlsx";
