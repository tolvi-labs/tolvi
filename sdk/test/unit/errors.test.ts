import { describe, it, expect } from "vitest";
import {
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
} from "../../src/errors";

const env = (code: string, message: string): ErrorEnvelope => ({
  error: { code, message },
});

describe("TolviAPIError subclasses", () => {
  it("TolviValidationError is a TolviAPIError and a TolviError", () => {
    const err = new TolviValidationError(env("validation_error", "bad input"), "req-1");
    expect(err).toBeInstanceOf(TolviValidationError);
    expect(err).toBeInstanceOf(TolviAPIError);
    expect(err).toBeInstanceOf(TolviError);
    expect(err).toBeInstanceOf(Error);
    expect(err.name).toBe("TolviValidationError");
    expect(err.status).toBe(400);
    expect(err.code).toBe("validation_error");
    expect(err.message).toBe("bad input");
    expect(err.body).toEqual(env("validation_error", "bad input"));
    expect(err.requestId).toBe("req-1");
  });

  it("TolviNotFoundError reports status 404", () => {
    const err = new TolviNotFoundError(env("not_found", "no such doc"));
    expect(err.status).toBe(404);
    expect(err.name).toBe("TolviNotFoundError");
    expect(err.requestId).toBeUndefined();
  });

  it("TolviAuthError reports status 401", () => {
    const err = new TolviAuthError(env("unauthorized", "missing api key"));
    expect(err.status).toBe(401);
  });

  it("TolviEmbeddingUnavailableError reports status 503", () => {
    const err = new TolviEmbeddingUnavailableError(env("embedding_unavailable", "ollama down"));
    expect(err.status).toBe(503);
  });

  it("TolviUnknownAPIError preserves arbitrary status codes", () => {
    const err = new TolviUnknownAPIError(429, env("rate_limited", "slow down"));
    expect(err.status).toBe(429);
    expect(err.code).toBe("rate_limited");
  });

  it("preserves the full ErrorEnvelope body for forward-compat", () => {
    const body = { error: { code: "validation_error", message: "bad", details: { field: "repo" } } };
    const err = new TolviValidationError(body as ErrorEnvelope);
    expect(err.body).toEqual(body);
  });
});

describe("TolviConnectionError", () => {
  it("wraps the underlying fetch error", () => {
    const cause = new Error("ECONNREFUSED");
    const err = new TolviConnectionError(cause);
    expect(err).toBeInstanceOf(TolviError);
    expect(err.name).toBe("TolviConnectionError");
    expect(err.cause).toBe(cause);
    expect(err.message).toContain("ECONNREFUSED");
  });
});

describe("TolviAbortError", () => {
  it("is a TolviError with the abort message", () => {
    const err = new TolviAbortError("Aborted");
    expect(err).toBeInstanceOf(TolviError);
    expect(err.name).toBe("TolviAbortError");
    expect(err.message).toBe("Aborted");
  });
});
