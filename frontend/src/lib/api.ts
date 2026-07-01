import type {
  Category,
  ExportFormat,
  Filters,
  Region,
  SearchResponse,
  SearchSummary,
  StreamEvent,
} from "./types";

/** Base URL of the Go backend. Configurable via NEXT_PUBLIC_API_URL. */
export const API_BASE =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

/** Parse a fetch Response, throwing an Error with the body text on non-2xx. */
async function parse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(text || `Request failed with status ${res.status}`);
  }
  return (await res.json()) as T;
}

/** Run a search against a region with the given filters. */
export async function searchCompanies(
  region: Region,
  filters: Filters,
): Promise<SearchResponse> {
  const res = await fetch(`${API_BASE}/api/v1/search`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ region, filters }),
  });
  return parse<SearchResponse>(res);
}

/** Run a streaming search: newly found companies and per-source progress arrive
 *  incrementally as NDJSON. `onEvent` fires for each event; pass an AbortSignal
 *  to cancel. Resolves when the stream ends. */
export async function searchCompaniesStream(
  region: Region,
  filters: Filters,
  onEvent: (event: StreamEvent) => void,
  signal?: AbortSignal,
): Promise<void> {
  const res = await fetch(`${API_BASE}/api/v1/search/stream`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ region, filters }),
    signal,
  });
  if (!res.ok || !res.body) {
    const text = await res.text().catch(() => "");
    throw new Error(text || `Stream failed with status ${res.status}`);
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  // NDJSON: one JSON object per line, flushed as the server finds results.
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    let nl: number;
    while ((nl = buffer.indexOf("\n")) >= 0) {
      const line = buffer.slice(0, nl).trim();
      buffer = buffer.slice(nl + 1);
      if (line) onEvent(JSON.parse(line) as StreamEvent);
    }
  }
  const tail = buffer.trim();
  if (tail) onEvent(JSON.parse(tail) as StreamEvent);
}

/** List past searches, most recent first. */
export async function listSearches(): Promise<SearchSummary[]> {
  const res = await fetch(`${API_BASE}/api/v1/searches`);
  const data = await parse<SearchSummary[] | null>(res);
  return data ?? [];
}

/** Fetch the list of selectable categories. */
export async function getCategories(): Promise<Category[]> {
  const res = await fetch(`${API_BASE}/api/v1/categories`);
  const data = await parse<Category[] | null>(res);
  return data ?? [];
}

/** Search for regions by free-text query (Nominatim-backed on the server). */
export async function searchRegions(q: string): Promise<Region[]> {
  const res = await fetch(
    `${API_BASE}/api/v1/regions?q=${encodeURIComponent(q)}`,
  );
  const data = await parse<Region[] | null>(res);
  return data ?? [];
}

/** Build the export download URL for a search in the requested format. */
export function exportUrl(searchId: number, format: ExportFormat): string {
  return `${API_BASE}/api/v1/searches/${searchId}/export?format=${format}`;
}
