// Public API entry for @tolvi-labs/sdk.

export { VERSION } from "./version.js";
export {
  TolviError,
  TolviAPIError,
  TolviValidationError,
  TolviAuthError,
  TolviNotFoundError,
  TolviEmbeddingUnavailableError,
  TolviUnknownAPIError,
  TolviConnectionError,
  TolviAbortError,
  type ErrorEnvelope,
} from "./errors.js";

import type { paths } from "./types.gen.js";

// ----- Document shapes -----
export type CreateDocumentRequest =
  paths["/v1/documents"]["post"]["requestBody"]["content"]["application/json"];

type CreateDocumentResponseBody =
  paths["/v1/documents"]["post"]["responses"]["200"]["content"]["application/json"];
export type CreateDocumentResponse = CreateDocumentResponseBody;
/** Document shape returned by `documents.create` — ingest metadata (content_hash, chunk count, embedded_at). For the full document with `body`/`frontmatter`/full chunks, see `FullDocument`. */
export type Document = CreateDocumentResponseBody["document"];

export type ListDocumentsQuery = {
  repo?: string;
  doc_type?: "decision" | "session" | "pattern";
  status?: string;
  limit?: number;
};
export type ListDocumentsResponse =
  paths["/v1/documents"]["get"]["responses"]["200"]["content"]["application/json"];
/** Lightweight document shape returned by `documents.list`. */
export type DocumentListItem = ListDocumentsResponse["documents"][number];

export type GetDocumentResponse =
  paths["/v1/documents/{id}"]["get"]["responses"]["200"]["content"]["application/json"];
/** Full document shape returned by `documents.get` — includes `body`, `frontmatter`, `created_at`, `date`, and the complete chunks array. */
export type FullDocument = GetDocumentResponse["document"];

// ----- Sync -----
export type SyncBatchRequest =
  paths["/v1/sync"]["post"]["requestBody"]["content"]["application/json"];
export type SyncBatchResponse =
  paths["/v1/sync"]["post"]["responses"]["200"]["content"]["application/json"];

// ----- Repos -----
export type ListReposResponse =
  paths["/v1/repos"]["get"]["responses"]["200"]["content"]["application/json"];
export type Repo = ListReposResponse["repos"][number];

// ----- Search -----
export type SearchRequest =
  paths["/v1/search"]["post"]["requestBody"]["content"]["application/json"];
export type SearchResponse =
  paths["/v1/search"]["post"]["responses"]["200"]["content"]["application/json"];
export type SearchResult = SearchResponse["results"][number];

// ----- Ask -----
export type AskRequest =
  paths["/v1/ask"]["post"]["requestBody"]["content"]["application/json"];
export type AskResponse =
  paths["/v1/ask"]["post"]["responses"]["200"]["content"]["application/json"];
export type Citation = AskResponse["citations"][number];

// ----- Client surface -----
export { Tolvi, type TolviOptions, type RequestOptions } from "./client.js";
