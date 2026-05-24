import type { Http } from "../http.js";
import type {
  CreateDocumentRequest,
  CreateDocumentResponse,
  GetDocumentResponse,
  ListDocumentsQuery,
  ListDocumentsResponse,
} from "../index.js";

export interface RequestOptions {
  signal?: AbortSignal;
}

export class DocumentsResource {
  constructor(private readonly http: Http) {}

  /**
   * Create a document in a repo. Synchronous ingest — returns when the document
   * is parsed, validated, chunked, embedded, and searchable.
   *
   * @example
   *   const result = await client.documents.create({
   *     repo: "my-repo",
   *     path: "decisions/2026-05-01-adopt-postgres.md",
   *     content: "---\ntags: [decision]\n---\n# Adopt Postgres\n..."
   *   });
   *   console.log(result.document.id);
   */
  create(body: CreateDocumentRequest, opts: RequestOptions = {}): Promise<CreateDocumentResponse> {
    return this.http.request<CreateDocumentResponse>("POST", "/v1/documents", {
      body,
      signal: opts.signal,
    });
  }

  /**
   * List documents in the workspace, optionally filtered by repo / doc_type / status.
   *
   * @example
   *   const { documents } = await client.documents.list({ repo: "my-repo", doc_type: "decision" });
   */
  list(query: ListDocumentsQuery = {}, opts: RequestOptions = {}): Promise<ListDocumentsResponse> {
    return this.http.request<ListDocumentsResponse>("GET", "/v1/documents", {
      query: query as Record<string, string | number | boolean | undefined>,
      signal: opts.signal,
    });
  }

  /**
   * Fetch a single document by its UUID. Returns the full document including
   * `body`, `frontmatter`, `created_at`, `date`, and the complete chunks array
   * — richer than the create/list shapes.
   *
   * @example
   *   const { document } = await client.documents.get("11111111-1111-1111-1111-111111111111");
   *   console.log(document.body);
   */
  get(id: string, opts: RequestOptions = {}): Promise<GetDocumentResponse> {
    return this.http.request<GetDocumentResponse>("GET", `/v1/documents/${encodeURIComponent(id)}`, {
      signal: opts.signal,
    });
  }

  /**
   * Soft-delete a document by its UUID. Returns once the deletion has been recorded.
   *
   * @example
   *   await client.documents.delete("11111111-1111-1111-1111-111111111111");
   */
  delete(id: string, opts: RequestOptions = {}): Promise<void> {
    return this.http.request<void>("DELETE", `/v1/documents/${encodeURIComponent(id)}`, {
      signal: opts.signal,
    });
  }
}
