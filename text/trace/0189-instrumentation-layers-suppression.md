# Instrumentation Layers and Suppression

This document describes approach for tracing instrumentation layers and suppressing duplicate layers.

## Motivation

- Provide clarity for instrumentation layers: e.g. DB calls on top of REST API
- Give mechanism to suppress instrumentation layers for the same convention: e.g. multiple instrumented HTTP clients using each other.

## Explanation

### Spec changes proposal

- Tracing Semantic Conventions: Each span MUST follow a single (besides general and potentially composite), convention, specific to the call it describes.
- Tracing Semantic Conventions: Client libraries instrumentation MUST make context current to enable correlation with underlying layers of instrumentation
- Trace API: Add `InstrumentationType` expandable enum with predefined values for known trace conventions (HTTP, RPC, DB, Messaging (see open questions)). *Type* is just a convention name.
- Trace SDK: Add span creation option to set `InstrumentationType`
  - During span creation, checks if span should be suppressed (there is another one for kind + type on the parent `Context`) and returns a *suppressed* span, which is
    - non-recording
    - propagating (carries parent context)
    - does not become current (i.e. `makeCurrent` call with it is noop)
- OTel SDK SHOULD allow suppression strategy configuration
  - suppress nested by kind (e.g. only one CLIENT allowed)
  - suppress nested by kind + convention it follows (only one HTTP CLIENT allowed, but outer DB -> nested HTTP is ok)
  - suppress none

Note: some conventions may explicitly include others (e.g. [FaaS](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/faas.md) may include [HTTP](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/http.md)). For the purpose of this document, we assume span follows single convention and instrumentation MUST include explicitly mentioned sub-conventions (except general) only.
[General](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/span-general.md) convention attributes are allowed on all spans when applicable, pure generic instrumentation are not suppressed.

#### Suppression Example

- Instrumentation - HTTP Client 1:
  - starts span
    - SDK checks if there is a span of the same kind + type on the parent context
    - there is no similar span
    - SDK returns new span
  - stores it in the context (presumably current)
    - SDK stores span in context behind `ContextKey` which is combination of kind and type (depending on configuration)
- Instrumentation - HTTP Client 2 (in scope of span1):
  - [Optional suggestion]: add API to check if span should be started (optimization for power-users)
  - starts span
    - SDK checks if there is a span of the same kind + type on the parent context
    - there is one already
    - SDK returns suppressed (non-recording, propagating span, it's also never becomes current).
  - HTTP client continues instrumentation with suppressed span
    - adds attributes: noop
    - injects context: duplicate work, since parent context is already injected and suppressed span carries the same context
    - makes suppressed span current: current is not modified, scope is noop
- Instrumentation - TLS (imaginary, in scope of span2)
  - works as if it's in scope of span1
  - current is span1
  - if explicit parent was used (span2), it carries span1 context

Suppression logic is configurable and encapsulated in span builder SDK implementation - specific library instrumentation should not depend on configuration, e.g.:

- suppressing by kind only - context key does not distinguish types within kind
- suppressing by kind + type - context key is per type and kind
- suppressing none

## Internal details

Client libraries frequently use common protocols (HTTP, gRPC, DB drivers) to perform RPC calls, which are usually instrumented by OpenTelemetry.
At the same time, client library is rarely a thin client and may need its own instrumentation to

- connect traces to application code
- provide extra context:
  - duration of composite operations
  - overall result of composite operation
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

- HTTP clients frequently are built on top of other HTTP clients, making multiple layers of HTTP spans.
- Web frameworks have multiple layers, and auto-instrumentation is applied to many of them, causing duplicate instrumentation, depending on the app configuration.
- Libraries may decide to add native instrumentation for common protocols like HTTP or gRPC:
  - to support legacy correlation protocols
  - to make better decisions on failures (e.g. 404, 409)
  - give better library-specific context
  - support users that can't use auto-instrumentation

So what happens in reality without attempts to suppress duplicates:

- HTTP SERVER span (middleware)
  - HTTP SERVER span (servlet)
    - Controller INTERNAL span
      - HTTP CLIENT call - 1 (Google HTTP client)
        - HTTP CLIENT call - 1 (Apache HTTP client)

#### Proposed solution

Suppress inner layers of the same instrumentation, i.e. above picture translates into:

- HTTP SERVER span (middleware)
  - Controller INTERNAL span
    - HTTP CLIENT call - 1 (Google HTTP client)

To do so, instrumentation declares convention (`InstrumentationType`) when starting a span and SDK:

- checks if span with same kind + type is present on context already
  - yes: backs off, never starting a span
  - no: starts a span and sets it on the context, e.g. by writing a span on the context under the key (where key is function of kind and type).

For this to work between different instrumentations (native and auto), the API to set type must be in Trace API.

### Configuration

Suppression strategy should be configurable on SDK side - instrumentation does not need to know about it.

Configuration is needed since:

- backends don't always support nested CLIENT spans (extra hints needed for Application Map to show outbound connection)
- users may prefer to reduce verbosity and costs by suppressing spans of same kind

So following strategies should be supported:

- suppress all nested of same kind
- suppress all nested of same kind + type (default?)
- suppress none (mostly for debugging instrumentation code and internal observability)

#### Suppression examples: kind and type

- HTTP SERVER
  - HTTP CLIENT - ok

- HTTP SERVER
  - HTTP SERVER - suppressed

- HTTP SERVER
  - MESSAGING CONSUMER - ok // open questions around receive/process

- MESSAGING PRODUCER // open questions around receive/process
  - HTTP CLIENT - ok

- MESSAGING CLIENT
  - HTTP CLIENT - ok

#### Suppression examples: kind

- HTTP SERVER
  - HTTP CLIENT - ok

- HTTP SERVER
  - HTTP SERVER - suppressed

- HTTP SERVER
  - MESSAGING CONSUMER - ok

- MESSAGING PRODUCER // open questions around client/producer uncertainty
  - HTTP CLIENT - ok

- MESSAGING CLIENT
  - HTTP CLIENT - suppressed

### Implementation

#### Trace API

Proof of concept in Java: https://github.com/lmolkova/opentelemetry-java/pull/1

- introduces a new `SuppressedSpan` implementation. It's almost the same as `PropagatingSpan` (i.e. sampled-out), with the difference that `makeCurrent` is noop
  - `PropagatingSpan` can't be reused since sampling out usually assumes to sample out this and all following downstream spans. Sampled out span becomes current and causes side-effects (span.current().setAttribute()) not compatible with suppression.
- adds extendable `InstrumentationType` and `SpanBuilder.setType(InstrumentationType type)`
- adds optional optimization `Tracer.shouldStartSpan(name, kind, type)`

#### Instrumentation API

[Instrumentation API in Java implementation](https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/instrumentation-api/src/main/java/io/opentelemetry/instrumentation/api/instrumenter/SpanKey.java) with suppression by type.

## Trade-offs and mitigations

Trace API change is needed to support native library instrumentations - taking dependency on unstable experimental instrumentation API (or common contrib code) is not a good option.

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

- Should we suppress by direction (inbound/outbound) instead of kind? Suggestion: let's fix current messaging quirks that cause concern here
  - Messaging CONSUMER spans (CONSUMER *receive* is a parent of CONSUMER *process* and seem to violate [kind definition](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/api.md#spankind))
    - I'm going to drive changing it in messaging spec (*receive* is normal CLIENT span based on the )
  - Messaging PRODUCER/CLIENT uncertainty: it's changing, see [Creation and publishing](https://github.com/open-telemetry/oteps/pull/173#discussion_r704737276)
    - Proposal: They are two different things
      - Creation is PRODUCER
      - Publish is CLIENT

- Should we have `Tracer.shouldStart(spanName, kind, type, ?)` or `SpanBuilder.shouldStart()` methods to optimize instrumentation. If it's not called, everything works, just not too efficient

- Backends need hint to separate logical CLIENT spans from physical ones
- Good default (suppress by kind or kind + type). Up to distro + user.
- Should we have configuration option to never suppress anything
