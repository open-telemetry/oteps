# Remove SpanData

Remove and replace SpanData by adding span start and finish options.

## Motivation

SpanData represents an immutable span object, creating a fairly large API for all of the fields (12 to be exact). It exposes what feels like an SDK concern and implementation detail to the API surface. As a user, this is another API I need to learn how to use. As an implementer, it’s a new data type that needs to be supported. The primary motivation for removing SpanData revolves around the desire to reduce the size of the tracing API.

## Explanation

SpanData's primary use case comes from the need to construct and report out of band spans, meaning that you're creating "custom" spans for an operation you don't own. A good example of this is a program that takes in structured logs that contain correlation IDs and a duration (e.g. from splunk) and [converts them](https://github.com/lightstep/splunktospan/blob/master/splunktospan/span.py#L43) to spans for your tracing system. [Another example](https://github.com/lightstep/haproxy_log2span/blob/master/lib/lib.go#L292) is running a sidecar on an HAProxy machine, tailing the request logs, and creating spans. SpanData supports the out of band reporting case, whereas you can’t with the current Span API as you cannot set the start and end timestamp, or add any tags at span creation that the sampler might need.

I'd like to propose getting rid of SpanData and `tracer.recordSpanData()` and replacing it by allowing `tracer.startSpan()` and `span.finish()` to accept start and end timestamp options. This reduces the API surface, consolidating on a single span type. Options would meet the requirements for out of band reporting.

## Internal details

`startSpan()` would change so you can include options such as `withStartTimestamp()`, `withTags()`, `withResource()`, `withEvents()`, `withLogs()`, etc. and for `finish()` you could have `withEndTimestamp()`, `withEvents()`, `withLogs()`, etc. The exact implementation would be language specific, so some would use an options pattern with function overloading or variadic parameters, or a span builder pattern.

## Trade-offs and mitigations

From https://github.com/open-telemetry/opentelemetry-specification/issues/71: If the underlying SDK automatically adds tags to spans such as thread-id, stacktrace, and cpu-usage when a span is started, they would be incorrect for out of band spans as the tracer would not know the difference between in and out of band spans. This can be mitigated by indicating that the span is out of band to prevent attaching possibly incorrect information.

https://github.com/open-telemetry/opentelemetry-specification/issues/96 discusses the possibility of allowing SpanData to support lazy fields, preventing allocations until it is read by the exporter. With span options, I don’t believe we would be able to get a fully lazy span, but we can effectively satisify the requirements, especially by adding lazily evaluated events and logs. OpenTracing added bulk logging options when finishing a span for this reason, allowing users to attach logs should, say, an error occur and they explicitly sample the span. Some discussion should be had about the laziness questions in the Open Questions section.

## Prior art and alternatives

The OpenTracing specification for `tracer.startSpan()` includes an optional start timestamp and zero or more tags. It also calls out an optional end timestamp for `span.finish()`.

## Open questions

What does lazy mean for SpanData? When we talk about lazy logs, we normally mean lazy evaluation, where expensive formatting is deferred until they are actually needed. Is laziness a desired property of SpanData? If so, what are the other requirements for SpanData? 

There also seems to be some hidden dependency between SpanData and the sampler API. For example, given a complete SpanData object with a start and finish timestamp, I imagine there's a use case where the sampler can look at the that and decide "this took a long time" and sample it. Is this a real use case? Is there a requirement to be able to provide complete span objects to the sampler?

## Related Issues

Options would solve [open-telemetry/opentelemetry-specification#139](https://github.com/open-telemetry/opentelemetry-specification/issues/139).

By removing SpanData, [open-telemetry/opentelemetry-specification#92](https://github.com/open-telemetry/opentelemetry-specification/issues/92) can be resolved and closed.

[open-telemetry/opentelemetry-specification#68](https://github.com/open-telemetry/opentelemetry-specification/issues/68) can be closed. `withResources()` can provide a different resource for out of band spans, otherwise the tracer can provide the default resource.

[open-telemetry/opentelemetry-specification#60](https://github.com/open-telemetry/opentelemetry-specification/issues/60) can be closed due to removal of SpanData.