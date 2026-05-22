// Shared error handler. Installed by both buildApp (production) and
// buildTestApp (integration tests) so the request → 400 path is identical
// in both. Without this, integration tests pin a different surface than
// production and validation-error contracts drift silently.

import type { FastifyInstance } from 'fastify';
import { ZodError } from 'zod';

// installErrorHandler wires a single global setErrorHandler on the
// given Fastify instance. Detects validation errors from multiple
// shapes (fastify-type-provider-zod's Fastify-wrapped errors, raw
// ZodError instances, errors with a `validation` array or `issues`
// array, errors with a ZodError cause) and emits the canonical
// `{ error: { code, message } }` envelope with status 400.
//
// Other errors are passed through to Fastify's default rendering.
export function installErrorHandler(app: FastifyInstance): void {
  app.setErrorHandler((error, request, reply) => {
    const e = error as {
      validation?: unknown;
      validationContext?: string;
      code?: string;
      issues?: unknown;
      cause?: unknown;
      statusCode?: number;
    };
    const isZodError = error instanceof ZodError;
    const isValidation =
      isZodError ||
      Array.isArray(e.validation) ||
      e.code === 'FST_ERR_VALIDATION' ||
      (typeof e.code === 'string' && e.code.startsWith('FST_ERR_VALIDATION')) ||
      Array.isArray(e.issues) ||
      e.cause instanceof ZodError;

    if (isValidation) {
      const message = isZodError
        ? error.issues.map((i) => `${i.path.join('.') || '<root>'}: ${i.message}`).join('; ')
        : e.cause instanceof ZodError
          ? e.cause.issues.map((i) => `${i.path.join('.') || '<root>'}: ${i.message}`).join('; ')
          : (error as Error).message || 'validation failed';
      return reply.code(400).send({
        error: { code: 'validation_failed', message },
      });
    }
    request.log.error({ err: error }, 'unhandled route error');
    return reply.send(error);
  });
}
