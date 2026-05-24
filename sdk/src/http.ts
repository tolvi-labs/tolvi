import {
  type ErrorEnvelope,
  TolviAbortError,
  TolviAPIError,
  TolviAuthError,
  TolviConnectionError,
  TolviEmbeddingUnavailableError,
  TolviError,
  TolviNotFoundError,
  TolviUnknownAPIError,
  TolviValidationError,
} from "./errors.js";
import { VERSION } from "./version.js";

export interface HttpOptions {
  apiKey: string;
  baseUrl: string;
  fetch?: typeof globalThis.fetch;
  userAgent?: string;
}

export interface HttpRequestInit {
  body?: unknown;
  query?: Record<string, string | number | boolean | undefined>;
  signal?: AbortSignal;
}

export class Http {
  private readonly apiKey: string;
  private readonly baseUrl: string;
  private readonly fetch: typeof globalThis.fetch;
  private readonly userAgent: string;

  constructor(opts: HttpOptions) {
    this.apiKey = opts.apiKey;
    this.baseUrl = opts.baseUrl.replace(/\/+$/, ""); // strip trailing slash
    this.fetch = opts.fetch ?? globalThis.fetch.bind(globalThis);
    this.userAgent = opts.userAgent ?? `tolvi-sdk/${VERSION}`;
  }

  async request<T>(method: "GET" | "POST" | "DELETE", path: string, init: HttpRequestInit = {}): Promise<T> {
    const url = this.buildUrl(path, init.query);
    const headers: Record<string, string> = {
      authorization: `Bearer ${this.apiKey}`,
      "user-agent": this.userAgent,
    };
    let body: string | undefined;
    if (method !== "GET" && init.body !== undefined) {
      body = JSON.stringify(init.body);
      headers["content-type"] = "application/json";
    }

    let res: Response;
    try {
      res = await this.fetch(url, { method, headers, body, signal: init.signal });
    } catch (err) {
      throw this.mapTransportError(err);
    }

    return this.handleResponse<T>(res);
  }

  private buildUrl(path: string, query?: HttpRequestInit["query"]): string {
    const base = `${this.baseUrl}${path}`;
    if (!query) return base;
    const params = new URLSearchParams();
    for (const [k, v] of Object.entries(query)) {
      if (v !== undefined) params.append(k, String(v));
    }
    const qs = params.toString();
    return qs ? `${base}?${qs}` : base;
  }

  private async handleResponse<T>(res: Response): Promise<T> {
    if (res.status === 204) return undefined as T;
    if (res.ok) return (await res.json()) as T;

    const body = await safeParseEnvelope(res);
    const requestId = res.headers.get("x-request-id") ?? undefined;
    switch (res.status) {
      case 400: throw new TolviValidationError(body, requestId);
      case 401: throw new TolviAuthError(body, requestId);
      case 404: throw new TolviNotFoundError(body, requestId);
      case 503: throw new TolviEmbeddingUnavailableError(body, requestId);
      default:  throw new TolviUnknownAPIError(res.status, body, requestId);
    }
  }

  private mapTransportError(err: unknown): TolviError {
    if (err instanceof Error && err.name === "AbortError") {
      return new TolviAbortError(err.message);
    }
    if (err instanceof Error) return new TolviConnectionError(err);
    return new TolviConnectionError(new Error(String(err)));
  }
}

async function safeParseEnvelope(res: Response): Promise<ErrorEnvelope> {
  const fallback = (): ErrorEnvelope => ({
    error: { code: "unknown", message: res.statusText || "unknown" },
  });
  try {
    const parsed = (await res.json()) as unknown;
    if (
      typeof parsed === "object" &&
      parsed !== null &&
      "error" in parsed &&
      typeof (parsed as { error: unknown }).error === "object" &&
      (parsed as { error: { code?: unknown } }).error !== null &&
      typeof (parsed as { error: { code: unknown } }).error.code === "string" &&
      typeof (parsed as { error: { message: unknown } }).error.message === "string"
    ) {
      return parsed as ErrorEnvelope;
    }
    return fallback();
  } catch {
    return fallback();
  }
}
