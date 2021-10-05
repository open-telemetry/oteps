# Scenarios and Open Questions for Tracing semantic conventions for HTTP

This document aims to capture scenarios/open questions and a road map, both of
which will serve as a basis for [stabilizing](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#stable)
the [existing semantic conventions for HTTP](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/http.md),
which are currently in an [experimental](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#experimental)
state. The goal is to declare HTTP semantic conventions stable before the
end of 2021.

## Motivation

Most observability scenarios involve HTTP communication. For Distributed Tracing
to be useful across the entire scenario, having good observability for
HTTP is critical. To achieve this, OpenTelemetry must provide stable conventions
and guidelines for instrumenting HTTP communication.

Bringing the existing experimental semantic conventions for HTTP to a
stable state is a crucial step for users and instrumentation authors, as it
allows them to rely on [stability guarantees](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#not-defined-semantic-conventions-stability),
and thus to ship and use stable instrumentation.

## Roadmap

| Description | Done By     |
|-------------|-------------|
| This OTEP, consisting of scenarios/open questions and a proposed roadmap, is approved and merged. | 09/30/2021 |
| [Stability guarantees](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#not-defined-semantic-conventions-stability) for semantic conventions are approved and merged. This is not strictly related to semantic conventions for HTTP but is a prerequisite for stabilizing any semantic conventions. | 09/30/2021 |
| Separate PRs covering the scenarios and open questions in this document are approved and merged. | 10/29/2021 |
| Proposed specification changes are verified by prototypes for the scenarios and examples below. | 11/15/2021 |
| The [specification for HTTP semantic conventions for tracing](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/http.md) is fully updated according to this OTEP and declared [stable](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#stable). | 11/30/2021 |

## General concepts

There are several general OpenTelemetry open questions exist today:

* What does a config language look like for overriding certain defaults.
  For example, what HTTP status codes count as errors?
* How to handle additional levels of detail for spans, such as retries and
  redirects?
  Should it even be designed as levels of detail or as layers reflecting logical
  or physical interactions/transactions.
* What is the data model for links? What would a reasonable storage
  implementation look like?

Answering to these questions will most likely affect the way scenarios and
open questions below will be addressed.

> NOTE. This OTEP captures a scope for changes should be done to existing
experimental semantic conventions for HTTP, but does not propose solutions.

## Scope: scenarios and open questions

> NOTE. The scope defined here is subject for discussions and can be adjusted.

Scenarios and open questions mentioned below must be addressed via separate PRs.

### Error status

Per current spec 4xx must result in span with Error status. In many cases
404/409 error criteria depends on the app though.

### Required attribute sets

> At least one of the following sets of attributes is required:
>
> * `http.url`
> * `http.scheme`, `http.host`, `http.target`
> * `http.scheme`, [`net.peer.name`](span-general.md), [`net.peer.port`](span-general.md), `http.target`
> * `http.scheme`, [`net.peer.ip`](span-general.md), [`net.peer.port`](span-general.md), `http.target`

As a result, users that write queries against raw data or Zipkin/Jaeger don't
have consistent story across instrumentations and languages. e.g. they'd need to
write queries like
`select * where (getPath(http.url) == "/a/b" || getPath(http.target) == "/a/b")`

### Optional attributes

As a library owner, I don't understand the benefits of optional attributes:
they create overhead, they don't seem to be generically useful (e.g. flavor),
and are inconsistent across languages/libraries unless unified.

### Retries, redirects and hedging policies

Each try/redirect/hedging request must have unique context to be traceable and
to unambiguously ask for support from downstream service, which implies span
per call.

Redirects: users may need observability into what server hop had an error/took
too long. E.g., was 500/timeout from the final destination or a proxy?

### Sampling

* Need to mention between pre-sampling/post-sampling attributes (all that are
required and available pre-sampling should be provided)
* To make it efficient for noop case, need a hint for instrumentation
(e.g., `GlobalOTel.isEnabled()`) that SDK is present and configured before
creating pre-sampling attributes.

### Context propagation needs explanation

* Reusing instances of client HTTP requests between tries (it’s likely, so clean
  up context before making a call).
  
### WebSockets/Long-polling and streaming

Anything we can do better here? In many cases connection has app-lifetime,
messages are independent - can we explain to users how to do manual tracing
for individual messages? Do span events per message make sense at all?
Need some real-life/expertize here.

### Request/Response body (technically out-of-scope, but we should have an idea how to let users do it)

There is a lot of user feedback that they want it, but

* We can’t read body in generic instrumentation
* We can let users collect them
* Attaching to server span is trivial
* Spec for client: we should have an approach to let users unambiguously
  associate body with http client span (e.g. outer manual span that wraps HTTP
  call and response reading and has event/log with body)
* Reading/writing body may happen outside of HTTP client API (e.g. through
  network streams) – how users can track it too?

### Security concerns

Some attributes can contain potentially sensitive information. Most likely, by
default web frameworks/http clients should not expose that.

For example, `http.target` has a query string that may contain credentials.

### Not HTTP-specific, but needs to be explained/mentioned

* Extracting/injecting context from the wire
* Always making spans current (in case of lower-level instrumentations)
  * Client HTTP spans could have children or extra events (TLS/DNS)
  * Server spans - need to pass it to user code

## Out of scope

HTTP protocol is being widely used within many different platforms and systems,
which brings a lot of intersections with a transmission protocol layer and an
application layer. However, for HTTP Semantic Conventions specification we want
to be strictly focused on HTTP-specific aspects of distributed tracing to keep
the specification clear. Therefore, the following scenarios, including but not
limited to, are considered out of scope for this workgroup:

* Batch operations.
* Fan-in and fan-out operations (e.g., GraphQL)  
* HTTP as a transport layer for other systems (e.g., Messaging system built on
  top of HTTP).

To address these scenarios, we might want to work with OpenTelemetry community
to build instrumentation guidelines going forward.
