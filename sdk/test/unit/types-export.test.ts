import { describe, it, expect } from "vitest";
import * as sdk from "../../src/index";

describe("public type exports", () => {
  // Type-only assertions don't run at runtime; this test verifies the
  // values (constructors, VERSION) that index.ts re-exports remain present.
  // Type aliases are verified by `tsc --noEmit` in CI.
  it("exports VERSION", () => {
    expect(sdk.VERSION).toBeTypeOf("string");
  });

  it("exports the error classes", () => {
    expect(sdk.TolviError).toBeTypeOf("function");
    expect(sdk.TolviAPIError).toBeTypeOf("function");
    expect(sdk.TolviValidationError).toBeTypeOf("function");
    expect(sdk.TolviAuthError).toBeTypeOf("function");
    expect(sdk.TolviNotFoundError).toBeTypeOf("function");
    expect(sdk.TolviEmbeddingUnavailableError).toBeTypeOf("function");
    expect(sdk.TolviUnknownAPIError).toBeTypeOf("function");
    expect(sdk.TolviConnectionError).toBeTypeOf("function");
    expect(sdk.TolviAbortError).toBeTypeOf("function");
  });
});
