import type { Http } from "../http.js";
import type { AskRequest, AskResponse } from "../index.js";

export interface RequestOptions {
  signal?: AbortSignal;
}

export class AskResource {
  constructor(private readonly http: Http) {}

  ask(body: AskRequest, opts: RequestOptions = {}): Promise<AskResponse> {
    return this.http.request<AskResponse>("POST", "/v1/ask", {
      body,
      signal: opts.signal,
    });
  }
}
