# Instrumentation Layers and Suppression

This document describes approach for tracing instrumentation layers, suppressing duplicate layers and unambiguously enriching spans.

## Motivation

- Provide clarity for instrumentation layers: e.g. DB calls on top of REST API
- Give mechanism to suppress instrumentation layers for the same convention: e.g. multiple instrumented HTTP clients using each other.
- Give mechanism to enrich specific spans unambiguously: e.g. HTTP server span with routing information

## Explanation

### Spec changes proposal

- Tracing Semantic Conventions: Each span MUST follow a single (besides general) convention, specific to the call it describes.
- Trace API: Add `SpanKey` API that
  - checks if similar span already exists on the context (e.g. `SpanKey.HTTP_CLIENT.exists(context)`)
  - gets span following specific convention from the context (e.g. `SpanKey.HTTP_CLIENT.fromContextOrNull(context)`).
- Tracing Semantic Conventions: instrumentation MUST back off if span of same kind and following same convention is already exists on the context by using `SpanKey` API.

- Tracing Semantic Conventions: Client libraries instrumentation MUST make context current to enable correlation with underlying layers of instrumentation
- OTel SDK SHOULD allow suppression strategy configuration
  - suppress nested by kind (e.g. only one CLIENT allowed)
  - suppress nested by kind + convention it follows (only one HTTP CLIENT allowed, but outer DB -> nested HTTP is ok)
  - suppress none

Note: some conventions may explicitly include others (e.g. [FaaS](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/faas.md) may include [HTTP](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/http.md)), in this case for the purpose of this document, we assume span follows single convention. Instrumentation should only include explicitly mentioned sub-conventions (except general).
[General](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/span-general.md) convention attributes are allowed on all spans when applicable.

#### SpanKey API

SpanKey allows to

- read/write span to context
- check if specific span is on the context
- encapsulate different suppression strategies
- define known SpanKey, shared between instrumentations (static singletons):  HTTP, RPC, DB, Messaging.

#### Suppression Example

- HTTP Client 1:
  - Check if HTTP CLIENT span exists on the context: `SpanKey.HTTP_CLIENT.exists(ctx)`
  - No HTTP client span on the context:
    - start span
    - store span in context: `SpanKey.HTTP_CLIENT.storeInContext(ctx, span)`
    - Make `ctx` current
- Http Client 2:
  - Checks if HTTP CLIENT span is already on the context: `SpanKey.HTTP_CLIENT.exists(Context.current())`
  - HTTP client span is already there: do not instrument

Suppression logic is configurable and encapsulated in `SpanKey` - specific library instrumentation should not depend on configuration, e.g.:

- suppressing by kind only - context key does not distinguish conventions within kind
  - `SpanKey.HTTP_CLIENT.exists` returns true if any CLIENT span  is on the context
  - `SpanKey.HTTP_CLIENT.fromContextOrNull` returns CLIENT span
  - `SpanKey.HTTP_CLIENT.storeInContext` stores span in CLIENT span context key
- suppressing by kind + convention - context key is per convention and kind
  - `SpanKey.HTTP_CLIENT.exists` returns true if HTTP CLIENT span is  on the context
  - `SpanKey.HTTP_CLIENT.fromContextOrNull` returns HTTP CLIENT span
  - `SpanKey.HTTP_CLIENT.storeInContext` stores span in CLIENT + convention span context key
- suppressing none
  - `SpanKey.HTTP_CLIENT.exists` returns false ignoring context
  - `SpanKey.HTTP_CLIENT.fromContextOrNull` returns innermost HTTP CLIENT span on the context
  - `SpanKey.HTTP_CLIENT.storeInContext` stores span in CLIENT + convention span context key

#### Enrichment Example

- HTTP SERVER 1 - middleware/servlet
  - HTTP INTERNAL 1 - controller
    - User code that wants to add event/attribute/status to HTTP SERVER span
    - some internal instrumentation logic that sets route info status and exception ot anything else available after controller span starts.
  
Assuming user code uses current context, it will get controller INTERNAL span. In order to enrich HTTP SERVER span, users may use `SpanKey.HTTP_SERVER.fromContextOrNull`.

## Internal details

Client libraries frequently use common protocols (HTTP, gRPC, DB drivers) to perform RPC calls, which are usually instrumented by OpenTelemetry.
At the same time, client library is rarely a thin client and may need its own instrumentation to

- connect traces to application code
- provide extra context:
  - duration of composite operations
  - overall result of all operation
  - any extra library-specific information not available on transport call span

Both, client library 'logical' and transport 'physical' spans are useful. They also rarely can be combined together because they have 1:many relationship.

So instrumentations form *layers*, where each layer follows specific convention or no convention at all. Spans that are not convention-specific (generic manual spans, INTERNAL spans) are never suppressed.

*Example*:

- HTTP SERVER span
  - DB CLIENT call - 1
    - HTTP CLIENT call - 1
      - DNS CLIENT
      - TLS CLIENT
    - HTTP CLIENT call - 2

There are two HTTP client spans under DB call, they are children of DB client spans.  DB spans follow DB semantics only, HTTP spans similarly only follow HTTP semantics. If there are other layers of instrumentation (TLS) - it happens under HTTP client spans.

### Duplication problem

Duplication is a common issue in auto-instrumentation:

- e.g. HTTP clients frequently are built on top of other HTTP clients, making multiple layers of HTTP spans
- Libraries may decide to add native instrumentation for common protocols like HTTP or gRPC:
  - to support legacy correlation protocols
  - to make better decisions failures (e.g. 404, 409)
  - give better library-specific context
  - support users that can't or don't want to use auto-instrumentation

So what happens in reality without attempts to suppress duplicates:

- HTTP SERVER span (middleware)
  - HTTP SERVER span (servlet)
    - Controller INTERNAL span
      - HTTP CLIENT call - 1 (Google HTTP client)
        - HTTP CLIENT call - 1 (Apache HTTP client)

#### Proposed solution

Disallow multiple layers of the same instrumentation, i.e. above picture translates into:

- HTTP SERVER span (middleware)
  - Controller INTERNAL span
    - HTTP CLIENT call - 1 (Google HTTP client)

To do so, instrumentation:

- checks if span with same kind + convention is registered on context already
  - yes: backs off, never starting a span
  - no: starts a span and registers it on the context

Registration is done by writing a span on the context under the key. For this to work between different instrumentations (native and auto), the API to access spans must be in Trace API.

Same mechanism can be used by users/instrumentations to enrich spans, e.g. add route info to HTTP server span (current span is ambiguous)

### Configuration

Suppression strategy should be configurable:

- backends don't always support nested CLIENT spans (extra hints needed for Application Map to show outbound connection)
- users may prefer to reduce verbosity and costs by suppressing spans of same kind

So following strategies should be supported:

- suppress all nested of same kind
- suppress all nested of same kind + convention (default?)
- suppress none (mostly for debugging instrumentation code and internal observability)

### Implementation

Here's [Instrumentation API in Java implementation](https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/instrumentation-api/src/main/java/io/opentelemetry/instrumentation/api/instrumenter/SpanKey.java) with suppression by convention.

## Trade-offs and mitigations

Trace API change is needed to support native library instrumentations - taking dependency on unstable experimental instrumentation API (or common contrib code) is not a good option. Instrumentation API is a good temporary place until we can put it in Trace API, native instrumentation can use reflection to access `SpanKey` in instrumentation API.

## Prior art and alternatives

- Terminal context: suppressing anything below
- Exposing spans stack and allowing to walk it accessing span properties
- Suppress all nested spans of same kind
- Make logical calls INTERNAL

Discussions:

- [Client library + auto instrumentation](https://github.com/open-telemetry/opentelemetry-specification/issues/1767)
- [Prevent duplicate telemetry when using both library and auto instrumentation](https://github.com/open-telemetry/opentelemetry-java-instrumentation/issues/903)
- [Generic mechanism for preventing multiple Server/Client spans](https://github.com/open-telemetry/opentelemetry-java-instrumentation/issues/465)
- [Proposal for modeling nested CLIENT instrumentation](https://github.com/open-telemetry/opentelemetry-java-instrumentation/issues/1822)
- [SpanKind with layers](https://github.com/open-telemetry/opentelemetry-specification/issues/526)
- [Should instrumentations be able to interact with or know about other instrumentations](https://github.com/open-telemetry/opentelemetry-python-contrib/issues/369)
- [Server instrumentations should look for parent spans in current context before extracting context from carriers](https://github.com/open-telemetry/opentelemetry-python-contrib/issues/445)
- [CLIENT spans should update their parent span's kind to INTERNAL](https://github.com/open-telemetry/opentelemetry-python-contrib/issues/456)

## Open questions

- Backends need hint to separate logical CLIENT spans from physical ones
- Good default (suppress by kind or kind + convention)
- Should we have configuration option to never suppress anything
