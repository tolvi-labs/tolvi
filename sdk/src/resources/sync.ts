import type { Http } from "../http.js";
import type { SyncBatchRequest, SyncBatchResponse } from "../index.js";

export interface RequestOptions {
  signal?: AbortSignal;
}

export class SyncResource {
  constructor(private readonly http: Http) {}

  /**
   * Batch-sync up to 500 documents into a single repo. Each document is parsed,
   * validated, and either created, updated, or marked unchanged based on content hash.
   *
   * @example
   *   const { results, summary } = await client.sync.batch({
   *     repo: "my-repo",
   *     documents: [
   *       { path: "decisions/2026-05-01-foo.md", content: "..." },
   *       { path: "decisions/2026-05-02-bar.md", content: "..." }
   *     ]
   *   });
   *   console.log(summary); // { created: 1, updated: 1, unchanged: 0, failed: 0 }
   */
  batch(body: SyncBatchRequest, opts: RequestOptions = {}): Promise<SyncBatchResponse> {
    return this.http.request<SyncBatchResponse>("POST", "/v1/sync", {
      body,
      signal: opts.signal,
    });
  }
}
