import type { Http } from "../http.js";
import type { SearchRequest, SearchResponse } from "../index.js";

export interface RequestOptions {
  signal?: AbortSignal;
}

export class SearchResource {
  constructor(private readonly http: Http) {}

  query(body: SearchRequest, opts: RequestOptions = {}): Promise<SearchResponse> {
    return this.http.request<SearchResponse>("POST", "/v1/search", {
      body,
      signal: opts.signal,
    });
  }
}
