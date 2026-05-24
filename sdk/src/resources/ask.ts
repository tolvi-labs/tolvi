import type { Http } from "../http.js";
import type { AskRequest, AskResponse } from "../index.js";

export interface RequestOptions {
  signal?: AbortSignal;
}

export class AskResource {
  constructor(private readonly http: Http) {}

  /**
   * Ask a question of the workspace. Returns a synthesized answer grounded
   * in retrieved documents, plus citations.
   *
   * @example
   *   const { answer, citations } = await client.ask({ query: "why postgres?" });
   *   console.log(answer);
   *   for (const c of citations) console.log(c.slug, c.document_id);
   */
  ask(body: AskRequest, opts: RequestOptions = {}): Promise<AskResponse> {
    return this.http.request<AskResponse>("POST", "/v1/ask", {
      body,
      signal: opts.signal,
    });
  }
}
