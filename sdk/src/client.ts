import { Http } from "./http.js";
import { DocumentsResource } from "./resources/documents.js";
import { SyncResource } from "./resources/sync.js";
import { ReposResource } from "./resources/repos.js";
import { SearchResource } from "./resources/search.js";
import { AskResource } from "./resources/ask.js";
import type { AskRequest, AskResponse } from "./index.js";

export interface TolviOptions {
  apiKey: string;
  baseUrl: string;
  fetch?: typeof globalThis.fetch;
  userAgent?: string;
}

export interface RequestOptions {
  signal?: AbortSignal;
}

export class Tolvi {
  readonly documents: DocumentsResource;
  readonly sync: SyncResource;
  readonly repos: ReposResource;
  readonly search: SearchResource;
  private readonly askResource: AskResource;

  constructor(opts: TolviOptions) {
    const http = new Http(opts);
    this.documents = new DocumentsResource(http);
    this.sync = new SyncResource(http);
    this.repos = new ReposResource(http);
    this.search = new SearchResource(http);
    this.askResource = new AskResource(http);
  }

  ask(body: AskRequest, opts: RequestOptions = {}): Promise<AskResponse> {
    return this.askResource.ask(body, opts);
  }
}
