---
tags: [pattern, sample, observability]
status: active
languages: [typescript, go]
---

# Tracing context propagation

A trace is only useful if it survives every hop a request makes. When a request crosses a process boundary — HTTP call, gRPC call, message queue handoff, scheduled job pickup — the trace context must travel with it, or the destination service starts a brand new trace and the two halves of the story never reconnect. OpenTelemetry standardizes this with the W3C `traceparent` and `tracestate` headers; the discipline is to make sure every outbound call carries them and every inbound entry point reads them.

For HTTP and gRPC, the official OpenTelemetry instrumentations handle propagation transparently — install them on both client and server and the headers move automatically. The traps are everywhere else: background jobs (the trace context needs to be serialized into the job payload and rehydrated on the worker), message queues (use the broker's metadata channel, not the message body), and any custom transport.

A minimal manual injection in TypeScript:

```ts
import { context, propagation } from "@opentelemetry/api";

const carrier: Record<string, string> = {};
propagation.inject(context.active(), carrier);
// carrier now contains traceparent / tracestate to send alongside the message.
```

And the Go-side extraction at the consumer:

```go
import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
)

ctx := otel.GetTextMapPropagator().Extract(
    context.Background(),
    propagation.MapCarrier(carrier),
)
```

When to use this: any time a request crosses a process boundary in code you control. When not to use it: synchronous in-process calls within the same span.
