# Event basics

## Motivation

The introduction of Events has been contentious, so we want to write down and agree on a few basics.

### What are OpenTelemetry Events?

OpenTelemetry Events are a type of log which have a required event name and a specific structure which is implied by that event name.

OpenTelemetry Events are a core concept in OpenTelemetry Semantic Conventions.

### OTLP

Since events are a type of log, they share the same OTLP data structure and the same OTLP pipeline.

### API

OpenTelemetry should have a (user-facing) Log API which includes the ability to emit OpenTelemetry Events.

### Interoperability with generic logging libraries

It should be possible to send OpenTelemetry Logs from the OpenTelemetry Log API to a generic logging library (e.g. Log4j).
This allows users to integrate OpenTelemetry Logs into an existing (non-OpenTelemetry) log stream.

It should be possible to bypass the OpenTelemetry Log API entirely and emit OpenTelemetry Logs (including events)
directly via an existing language-specific logging libraries.

OpenTelemetry will recommend that
[instrumentation libraries](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/glossary.md#instrumentation-library)
use the OpenTelemetry Log API to emit OpenTelemetry Events over emitting events via a generic logging library in order to give users a simple and consistent
onboarding story that doesn't involve mixing the two approaches.

OpenTelemetry will recommend that application developers also use the OpenTelemetry Log API to emit OpenTelemetry Events over emitting events via a generic
logging library since it avoid accidentally emitting logs which lack an event name or are unstructured.

Also, recommending the OpenTelemetry Log API for emitting OpenTelemetry Events over emitting events via a generic logging library makes for a clearer overall
OpenTelemetry API story - with first class user facing APIs for traces, metrics, and events,
all suitable for using directly in native instrumentation.

### Relationship to Span Events

Events are intended to replace Span Events in the long-term.
Span Events will be deprecated to signal that users should prefer Events.

Interoperability between Events and Span Events will be defined in the short-term.

### SDK

This is the (existing) OpenTelemetry Log SDK.

## Trade-offs and mitigations

TODO

## Prior art and alternatives

TODO

## Open questions

How to support routing logs from the Log API to a language-specific logging library,
while at the same time routing logs from the language-specific logging library to a OpenTelemetry Logging Exporter.

How do Log bodies interop with generic logging libraries?

How do Event bodies interop with Span Events?

## Future possibilities

The Event API will probably need an `IsEnabled` function based on severity level, scope name, and event name.

Ergonomic improvements to make it more attractive from the perspective of being a replacement for generic logging APIs.

Capturing raw metric events as opposed to aggregating and emitting them as OpenTelemetry Metric data
(e.g. [opentelemetry-specification/617](https://github.com/open-telemetry/opentelemetry-specification/issues/617)).

Capturing raw span events as opposed to aggregating and emitting them as OpenTelemetry Span data
(e.g. a [streaming SDK](https://github.com/search?q=repo%3Aopen-telemetry%2Fopentelemetry-specification+%22streaming+sdk%22&type=issues)).

Capturing events and computing metrics from them.
