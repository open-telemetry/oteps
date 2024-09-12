# Event Basics

## Motivation

The introduction of Events has been contentious, so we want to document and agree on a few basics.

### What are OpenTelemetry Events?

OpenTelemetry Events are a type of OpenTelemetry Log that requires an event name and follows a specific structure implied by that event name.

They are a core concept in OpenTelemetry Semantic Conventions.

### OTLP

Since OpenTelemetry Events are a type of OpenTelemetry Log, they share the same OTLP log data structure and pipeline.

### API

OpenTelemetry should provide a (user-facing) Log API that includes the capability to emit OpenTelemetry Events.

### Interoperability with Generic Logging Libraries

It should be possible to send OpenTelemetry Logs from the OpenTelemetry Log API to a generic logging library (e.g., Log4j).
This allows users to integrate OpenTelemetry Logs into an existing (non-OpenTelemetry) log stream.

It should also be possible to bypass the OpenTelemetry Log API entirely and emit OpenTelemetry Logs (including Events)
directly via existing language-specific logging libraries, if that library has the capability to do so.

OpenTelemetry will recommend that
[instrumentation libraries](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/glossary.md#instrumentation-library)
use the OpenTelemetry Log API to emit OpenTelemetry Events rather than using a generic logging library. This recommendation aims to provide users with a simple and consistent
onboarding experience that avoids mixing approaches.

OpenTelemetry will also recommend that application developers use the OpenTelemetry Log API to emit OpenTelemetry Events instead of using a generic
logging library, as this helps prevent accidentally emitting logs that lack an event name or are unstructured.

Recommending the OpenTelemetry Log API for emitting OpenTelemetry Events, rather than using a generic logging library, contributes to a clearer overall
OpenTelemetry API story. This ensures a unified approach with first-class user-facing APIs for traces, metrics, and events,
all suitable for direct use in native instrumentation.

### Relationship to Span Events

Events are intended to replace Span Events in the long term.
Span Events will be deprecated to signal that users should prefer Events.

Interoperability between Events and Span Events will be defined in the short term.

### SDK

This refers to the existing OpenTelemetry Log SDK.

## Trade-offs and mitigations

TODO

## Prior art and alternatives

TODO

## Open questions

* How to support routing logs from the Log API to a language-specific logging library
  while simultaneously routing logs from the language-specific logging library to an OpenTelemetry Logging Exporter?
* How do log bodies interoperate with generic logging libraries?
* How do event bodies interoperate with Span Events?

## Future possibilities

* The Event API will likely need an `IsEnabled` function based on severity level, scope name, and event name.
* Ergonomic improvements to make it more attractive as a replacement for generic logging APIs.
* Capturing raw metric events as opposed to aggregating and emitting them as OpenTelemetry Metric data
  (e.g. [opentelemetry-specification/617](https://github.com/open-telemetry/opentelemetry-specification/issues/617)).
* Capturing raw span events as opposed to aggregating and emitting them as OpenTelemetry Span data
  (e.g. a [streaming SDK](https://github.com/search?q=repo%3Aopen-telemetry%2Fopentelemetry-specification+%22streaming+sdk%22&type=issues)).
* Capturing events and computing metrics from them.
