# Event basics

## Motivation

The introduction of Events has been contentious, so we want to write down and agree on a few basics.

### What are events?

Events are a type of log which has a required event name and a specific structure which is implied by that event name.

The event name and its implied structure make events much more suitable for observability compared to generic logs.

Because of this, the OpenTelemetry project wants to encourage the transition from generic logs to events.

### OTLP

Since events are a type of log, they share the same OTLP data structure and OTLP pipeline.

### API

OpenTelemetry should have an Event API. This will help to promote the distinction between events and generic logs,
and encourage the use of events over generic logs.

### Interoperability with generic logging libraries

It should be possible to send events from the Event API to a generic logging library (e.g. Log4j).
This allows users to integrate events from the Event API into an existing (non-OpenTelemetry) log stream.

Note: If a user chooses to send events from the Event API to a generic logging library, and they have
also chosen to send the logs from their generic logging library to the OpenTelemetry Logging SDK, then they should
avoid sending events from the Event API directly to the OpenTelemetry Logging SDK since that would lead to duplicate
capture of events that were sent from the Event API.

It should be possible to bypass the Event API entirely and emit Events directly via an existing language-specific logging libraries.
This helps reinforce the idea that events are just a specific type of log.

Note: Because of this, generic event processors should be implemented as Log SDK processors.

OpenTelemetry will recommend that
[instrumentation libraries](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/glossary.md#instrumentation-library)
use the Event API over emitting events via a generic logging library in order to give users a simple and consistent
onboarding story that doesn't involve mixing the two approaches.

OpenTelemetry will recommend that application developers also use the Event API over emitting events via a generic
logging library since it avoid accidentally emitting logs which lack an event name or are unstructured.

Also, recommending the Event API over emitting events via a generic logging library makes for a clearer overall
OpenTelemetry API story - with first class user facing APIs for traces, metrics, and events,
all suitable for using directly in native instrumentation.

### Relationship to Span Events

Events are intended to replace Span Events in the long-term.
Span Events will be deprecated to signal that users should prefer Events.

Interoperability between Events and Span Events will be defined in the short-term.

### SDK

The Event SDK needs to support two destinations for events:

* OpenTelemetry Logging SDK
* Language-specific logging libraries

## Trade-offs and mitigations

TODO

## Prior art and alternatives

TODO

## Open questions

How do Event bodies interop with generic logging libraries?

How do Event bodies interop with Span Events?

## Future possibilities

The Event API will probably need an `IsEnabled` function based on severity level, scope name, and event name.

Ergonomic improvements to make it more attractive from the perspective of being a replacement for generic logging APIs.

Capturing raw metric events as opposed to aggregating and emitting them as OpenTelemetry Metric data
(e.g. [opentelemetry-specification/617](https://github.com/open-telemetry/opentelemetry-specification/issues/617)).

Capturing raw span events as opposed to aggregating and emitting them as OpenTelemetry Span data
(e.g. a [streaming SDK](https://github.com/search?q=repo%3Aopen-telemetry%2Fopentelemetry-specification+%22streaming+sdk%22&type=issues)).

Capturing events and computing metrics from them.
