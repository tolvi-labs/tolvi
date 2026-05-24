import type { Http } from "../http.js";
import type { SearchRequest, SearchResponse } from "../index.js";

export interface RequestOptions {
  signal?: AbortSignal;
}

export class SearchResource {
  constructor(private readonly http: Http) {}

  /**
   * Semantic search across the workspace. Returns ranked results with similarity
   * scores and the matched chunk that produced each hit.
   *
   * @example
   *   const { results } = await client.search.query({
   *     query: "why did we pick postgres?",
   *     limit: 5,
   *     filters: { doc_type: ["decision"] }
   *   });
   *   for (const r of results) console.log(r.slug, r.score);
   */
  query(body: SearchRequest, opts: RequestOptions = {}): Promise<SearchResponse> {
    return this.http.request<SearchResponse>("POST", "/v1/search", {
      body,
      signal: opts.signal,
    });
  }
}
