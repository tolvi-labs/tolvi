export interface ErrorEnvelope {
  error: {
    code: string;
    message: string;
    // Forward-compat: server MAY add fields like `details` later.
    [key: string]: unknown;
  };
}

export abstract class TolviError extends Error {
  abstract readonly name: string;
}

export class TolviAPIError extends TolviError {
  readonly name: string = "TolviAPIError";
  readonly status: number;
  readonly code: string;
  readonly body: ErrorEnvelope;
  readonly requestId?: string;

  constructor(status: number, body: ErrorEnvelope, requestId?: string) {
    super(body.error.message);
    this.status = status;
    this.code = body.error.code;
    this.body = body;
    this.requestId = requestId;
  }
}

export class TolviValidationError extends TolviAPIError {
  override readonly name = "TolviValidationError";
  override readonly status: 400 = 400;
  constructor(body: ErrorEnvelope, requestId?: string) {
    super(400, body, requestId);
  }
}

export class TolviAuthError extends TolviAPIError {
  override readonly name = "TolviAuthError";
  override readonly status: 401 = 401;
  constructor(body: ErrorEnvelope, requestId?: string) {
    super(401, body, requestId);
  }
}

export class TolviNotFoundError extends TolviAPIError {
  override readonly name = "TolviNotFoundError";
  override readonly status: 404 = 404;
  constructor(body: ErrorEnvelope, requestId?: string) {
    super(404, body, requestId);
  }
}

export class TolviEmbeddingUnavailableError extends TolviAPIError {
  override readonly name = "TolviEmbeddingUnavailableError";
  override readonly status: 503 = 503;
  constructor(body: ErrorEnvelope, requestId?: string) {
    super(503, body, requestId);
  }
}

export class TolviUnknownAPIError extends TolviAPIError {
  override readonly name = "TolviUnknownAPIError";
  // status is whatever the server returned
}

export class TolviConnectionError extends TolviError {
  readonly name = "TolviConnectionError";
  readonly cause: Error;

  constructor(cause: Error) {
    super(cause.message);
    this.cause = cause;
  }
}

export class TolviAbortError extends TolviError {
  readonly name = "TolviAbortError";
}
