# Automatic Peer Service Name propagation

Automatic propagation of `peer.service` through `TraceState``.

## Motivation

Knowing the service name on the other side of a remote call is valuable
troubleshooting information. The semantic conventions represent this via
`peer.service`, which needs to be manually populated. In a deployment scenario,
when a new service is added, all the existing services interacting with it
need to update `peer.service`, which is error-prone and may
become unreliable, making it eventually obsolete.

This information can be effectively derived in the backend using the
`Resource` of the parent `Span`, but is otherwise not available
at Collector processing time, where it could be used for transformation
purposes or sampling (e.g. adaptive sampling based on the calling service).

As metrics and logs do not have defined a parent-child relationship, using
`peer.service` could help gaining insight into the remote service as well.

Defining (optional) automated population of `peer.service` will greatly help
adoption of this attribute by users and vendors explicitly interested in this
scenario.

## Explanation

SDKs will define an optional feature, disabled by default,
to read the `service.name` attribute of the related `Resource` and set it
in the spans' `TraceState` as described in
[trace state handling](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/tracestate-handling.md)
specifically using the `us` subkey (denoting **upstream service**):

```
ot=us:MyService
```

Instrumentation and processors are then free to use this information to set
`peer.service` and perform other related processing.

## Internal details

SDKs will disable by default this option, maintaining the current behavior.
When the feature is explicitly enabled by the user, spans will include
an additional entry in `TraceState` as described above. By doing this,
the user acknowledges the additional cost in memory and bandwidth.

Span creation will be updated like this:

```java
//
// SpanBuilder.startSpan()
//
if (tracerSharedState.propagateServiceName) {
  String serviceName = tracerSharedState.getResource().getAttribute(PEER_SERVICE);
  traceState = addServiceNameToTracerState(traceState, serviceName);
}
// Use the updated `traceState` to create the new SpanContext.
```

Server-side instrumentation (e.g. http servers, gRPC on the receiver side)
can then be updated to use the propagated context to look for the `us` subkey
in `TraceState`, and if it exists, use it to set `peer.service` on the local `Span`:

```java
// 
// Incoming request handling.
//
try (Scope scope = extractRemoteContext(headers).makeCurrent()) {
  SpanBuilder spanBuilder = tracer.spanBuilder("server-span");
  
  TraceState remoteTraceState = Span.current()
      .getSpanContext()
      .getTraceState();
  String peerServiceName = getUpstreamServiceName(remoteTraceState);
  if (peerServiceName != null) {
    spanBuilder.setAttribute(PEER_SERVICE, peerServiceName);
  }
}
```

With `peer.service` present in server spans, further processing, filtering and sampling can
then be accomplished in the Collector, e.g. a preview of the dependency map of a service,
similar in spirit to zPages could be created.

### Sampling scenarios

A specially interesting case is sampling depending on the calling service, specifically:

* An adaptive sampler may decide to sample or not based on the calling service, e.g.
  given Service A amounting to 98% of requests, and Service B amounting to 2% only,
  more traces could be sampled for the latter.
* In cases where a parent `Span ` is **not** sampled **but** its child (or linked-to `Span`)
  wants to sample, knowing the calling service **may** help with the sampling decision.
* In deployment scenarios where context is properly propagated through all the services,
  but not all of them are actually traced, it would be helpful to know what services
  were part of the request, even if no traces/spans exist for them, see
  https://github.com/w3c/trace-context/issues/550 as an example.

## Trade-offs and mitigations

Given the `TraceState` [lenght contrains](https://www.w3.org/TR/trace-context/#tracestate-header)
we may decide to trim the service name up to a given length.

In case propagating `peer.service` ever represents a privacy or security concern,
consider hashing the `peer.service` values, and provide a dictionary to interpret them
by the Collector and backends.

## Prior art and alternatives

 Using `Baggage` to **automatically** propagate `service.name` was explored.
It would consist of two parts:

* Explicit `Resource`'s `service.name` propagation using `Baggage`
  at **request** time. Instrumentation libraries would need to include
  an option to perform such propagation (potentially false by default,
  in order to keep the current behavior). Caveat is that `Resource` is an SDK item, while
  instrumentation is expected to solely rely on the API.
* Either explicit handling on the server-side instrumentation (similar to how
  it's proposed using `TraceState` above, but relying on `Baggage ` instead), or
  specialized processors that automatically enrich `Spans` with `Baggage` values,
  as shown below:

```java
public class BaggageDecoratingSpanProcessor implements SpanProcessor {
  public BaggageDecoratingSpanProcessor(SpanProcessor processor, Predicate<Span> predicate, Set<String> keys) {
    this.processor = processor;
    this.predicate = predicate;
    this.keys = keys;
  }

  public void onStart(Context context, ReadWriteSpan span) {
    this.processor.onStart(context, span);
    if (predicate.test(span)) {
      Baggage baggage = Baggage.current();
      keys.forEach(key -> {
        String value = baggage.getEntryValue(key);
        if (value != null) {
          span.setAttribute(key, baggage.getEntryValue(key))
        }
      })
    }
  }
}
```

The `TraceState` alternative was preferred as `Baggage` has general,
application-level propagation purposes, whereas `TraceState` can be used
by observability purposes, along with the fact that accessing `Resource`
from instrumentation is not feasible.

## Open questions

* At the moment of writing this OTEP only `peer.service` is defined (which
  relies on `service.name`). However, semantic conventions also define
  `service.version`, `service.instance.id` and `service.namespace`,
  which may provide additional details. Given the contraints of memory
  and bandwidth (for both `TraceState` and `Baggage`) we will decide
  in the future whether to propagate these additional values or not.

## Future possibilities

Logging and metrics can be augmented by using automatic `peer.service` propagation,
in order to hint at a parent-child (or client-server) relationship, given they do not
include such information as part of their data models:

* Logs can optionally be converted to traces if a hierarchy or dependency map is desired,
  but augmenting them with `peer.service` could be done as an intermediate step.
* Metrics could use `peer.service` as an additional dimension that helps performing filtering
  based on related services, for example.

