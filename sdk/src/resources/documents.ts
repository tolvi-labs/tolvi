import type { Http } from "../http.js";
import type {
  CreateDocumentRequest,
  CreateDocumentResponse,
  ListDocumentsQuery,
  ListDocumentsResponse,
} from "../index.js";

export interface RequestOptions {
  signal?: AbortSignal;
}

export class DocumentsResource {
  constructor(private readonly http: Http) {}

  create(body: CreateDocumentRequest, opts: RequestOptions = {}): Promise<CreateDocumentResponse> {
    return this.http.request<CreateDocumentResponse>("POST", "/v1/documents", {
      body,
      signal: opts.signal,
    });
  }

  list(query: ListDocumentsQuery = {}, opts: RequestOptions = {}): Promise<ListDocumentsResponse> {
    return this.http.request<ListDocumentsResponse>("GET", "/v1/documents", {
      query: query as Record<string, string | number | boolean | undefined>,
      signal: opts.signal,
    });
  }

  get(id: string, opts: RequestOptions = {}): Promise<CreateDocumentResponse> {
    return this.http.request<CreateDocumentResponse>("GET", `/v1/documents/${encodeURIComponent(id)}`, {
      signal: opts.signal,
    });
  }

  delete(id: string, opts: RequestOptions = {}): Promise<void> {
    return this.http.request<void>("DELETE", `/v1/documents/${encodeURIComponent(id)}`, {
      signal: opts.signal,
    });
  }
}
