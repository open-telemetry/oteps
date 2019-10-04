# Opening Span's Context

Remove the proposition that "`SpanContext` MUST be a final (sealed) class.".

## Motivation

Currently, SpanContext is the only class in the specification that has the limitation of being final placed upon it.
I also haven't heard a good argument for having a final `SpanContext` so this seems an arbitrary restriction.

On the other hand, a non-final SpanContext does have advantages for vendors' flexibility. For example, they could:

* Add custom routing information to the SpanContext more efficiently than by using the string-based mapping in TraceState.
* Generate Span IDs lazily (e.g., if a Span only has internal parents and children, a mechanism that is more efficient than a combined 128 + 64 bit identifier could be used).
* Count the number of times a SpanContext was injected to better estimate on the back end how long to wait for (additional) child Spans.

## Explanation

Spans remain as a non-abstract class implemented in the API, but they are non-final. Note that the proposition

> SpanContexts are immutable.

only applies to the part of the SpanContext that is publicly exposed on the API level.
Actual vendor implementations could have additional mutable properties or
actually have the first access to the `TraceId` property generate the TraceId lazily.

## Internal details

The changes in the initial summary of this OTEP shall be applied to the specification at
https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/api-tracing.md#spancontext.

Language libraries that actually have a "final" or "sealed" keyword must remove it from the API definition of SpanContext.
The reference SDK implementation will continue to use this Span implementation without deriving from it.
Additionally, all public APIs of the SpanContext must be overridable
(they must not be public data fields in languages where these are not overridable,
languages that require an explicit marker like `virtual` to make an API overridable must add it, etc.).

Existing propagators (which should be implemented in the SDK package or a separate package) can continue to directly construct a SpanContext.
API implementations that want to use a custom SpanContext implementation must either
handle having a plain SpanContext as parent
and having additional properties stripped upon extracting
or they provide their own propagator implementations too.

## Trade-offs and mitigations

The decision to have SpanContext as a non-abstract class in the API results from the goal to keep compatibility.
The cleaner alternative would be to have SpanContext be an interface like all other API classes.

Since the implementation of propagators is on the SDK level, all public APIs could be removed from the SpanContext
so that it becomes an object that is completely opaque (to everyone but the propagator implementations).
This would enable even more flexibility or at least spare vendors from adding boilerplate code that, e.g., calculates TraceIds
if there is no such concept in their own system.

## Prior art and alternatives

The topic was briefly discussed in
https://github.com/open-telemetry/opentelemetry-specification/pull/244
("Fix" contradictory wording in SpanContext spec). Before this PR, the specification contained the following sentence that directly contradicted the requirement that a SpanContext should be final (that requirement was there too):

> `SpanContext` is represented as an interface, in order to be serializable into a wider variety of trace context wire formats.

In OpenTelemetry, the SpanContext was just an interface with very few public methods
(in particular, it did not assume the existence of a thing such as Trace or Span ID):

> The `SpanContext` is more of a "concept" than a useful piece of functionality at the generic OpenTracing layer. That said, it is of critical importance to OpenTracing *implementations* and does present a thin API of its own. Most OpenTracing users only interact with `SpanContext` via [**references**](https://github.com/opentracing/specification/blob/master/specification.md#references-between-spans) when starting new `Span`s, or when injecting/extracting a trace to/from some transport protocol.

(quoted from https://github.com/opentracing/specification/blob/master/specification.md#spancontext)

In principle, the same still applies to OpenTelemetry:
SpanContext has no functionality that is useful outside of propagation or implementations of Spans and Tracers. However, we chose to expose the "internal details" of the W3C context in the API.

## Open questions

None known yet.

## Future possibilities

It is not expected that the OpenTelemetry reference implementation of the SDK or API benefits much from this change.
The reference SDK is blessed here
because if a new functionality is required in SpanContext (e.g., [IsRemote][])
it can "just" be added to the API implementation.
But for example, vendors who want to store attributes that are only useful for their special implementation in the SpanContext can now do that.

[IsRemote]: https://github.com/open-telemetry/opentelemetry-specification/pull/216
