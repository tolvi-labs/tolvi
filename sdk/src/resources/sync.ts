import type { Http } from "../http.js";
import type { SyncBatchRequest, SyncBatchResponse } from "../index.js";

export interface RequestOptions {
  signal?: AbortSignal;
}

export class SyncResource {
  constructor(private readonly http: Http) {}

  batch(body: SyncBatchRequest, opts: RequestOptions = {}): Promise<SyncBatchResponse> {
    return this.http.request<SyncBatchResponse>("POST", "/v1/sync", {
      body,
      signal: opts.signal,
    });
  }
}
