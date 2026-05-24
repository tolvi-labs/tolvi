import type { Http } from "../http.js";
import type { ListReposResponse } from "../index.js";

export interface RequestOptions {
  signal?: AbortSignal;
}

export class ReposResource {
  constructor(private readonly http: Http) {}

  /**
   * List all repos in the workspace.
   *
   * @example
   *   const { repos } = await client.repos.list();
   *   for (const repo of repos) console.log(repo.slug, repo.document_count);
   */
  list(opts: RequestOptions = {}): Promise<ListReposResponse> {
    return this.http.request<ListReposResponse>("GET", "/v1/repos", {
      signal: opts.signal,
    });
  }
}
