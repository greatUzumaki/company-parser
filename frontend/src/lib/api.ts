import type {
  CampaignEvent,
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

/** Read an NDJSON response body, invoking onEvent for each JSON line. */
async function readNDJSON<T>(res: Response, onEvent: (event: T) => void): Promise<void> {
  if (!res.ok || !res.body) {
    const text = await res.text().catch(() => "");
    throw new Error(text || `Stream failed with status ${res.status}`);
  }
  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    let nl: number;
    while ((nl = buffer.indexOf("\n")) >= 0) {
      const line = buffer.slice(0, nl).trim();
      buffer = buffer.slice(nl + 1);
      if (line) onEvent(JSON.parse(line) as T);
    }
  }
  const tail = buffer.trim();
  if (tail) onEvent(JSON.parse(tail) as T);
}

/** Run a streaming search: newly found companies and per-source progress arrive
 *  incrementally as NDJSON. `onEvent` fires for each event; pass an AbortSignal
 *  to cancel. Resolves when the stream ends. */
export async function searchCompaniesStream(
  region: Region,
  filters: Filters,
  onEvent: (event: StreamEvent) => void,
  signal?: AbortSignal,
  force = false,
): Promise<void> {
  const res = await fetch(`${API_BASE}/api/v1/search/stream`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ region, filters, force }),
    signal,
  });
  await readNDJSON<StreamEvent>(res, onEvent);
}

/** Whether email campaigns are configured (server has SMTP set up). */
export async function getCampaignStatus(): Promise<{ enabled: boolean }> {
  const res = await fetch(`${API_BASE}/api/v1/campaign/status`);
  return parse<{ enabled: boolean }>(res);
}

/** Send an email campaign to the emails collected in a search, streaming
 *  per-recipient progress. `confirm` is the required consent acknowledgement;
 *  `dryRun` previews without sending. */
export async function sendCampaign(
  params: {
    subject: string;
    body: string;
    recipients: { email: string; name: string }[];
    dryRun: boolean;
    confirm: boolean;
  },
  onEvent: (event: CampaignEvent) => void,
  signal?: AbortSignal,
): Promise<void> {
  const res = await fetch(`${API_BASE}/api/v1/campaign/send`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(params),
    signal,
  });
  await readNDJSON<CampaignEvent>(res, onEvent);
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
