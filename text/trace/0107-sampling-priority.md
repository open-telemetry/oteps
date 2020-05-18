# Associating sampling priority with the trace using tracestate

Propagating `sampling.priority` (calculated by probability sampler) in a
`tracestate` will enable a better experience for system with the independent
stateless configuration of components sampling.

## Motivation

Consistent sampling decision made in each app of a distributed trace is
important for better user experience of traces analysis. Consistency is achieved
by aligning stateless hashing algorithm used to make a decision based on
trace-id or explicitly propagating sampling flag.

Increasingly more apps participate in distributed tracing. With the
standardized wire format for context propagation, there is bigger chance that
approach for sampling chosen by one app may not match another app sampling
policy. Since coordination of sampling policies across many apps not always
possible, OpenTelemetry must provide a way to share sampling hints between apps.

Beyond sampling hints infrastructure, it is beneficial for the community to
agree on standard behavior and hints exposed by probability sampler.

The `sampling.priority` property is used by many vendors for various purposes.
Propagating this field alongside the trace will allow for many improvements and
as a very minimum will simplify transition of customers from SDKs using
different `trace-id` based hash functions to OpenTelemetry SDK.

Also see discussion here: <https://github.com/open-telemetry/opentelemetry-specification/pull/570>

## Explanation

### Sampling hints exchange

OpenTelemetry SDK has an infra to expose sampling hints as span attributes today
via
[SamplingResult](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/trace/sdk.md#shouldsample)
class.

This recommendation suggests OpenTelemetry will allow to return an updated
`tracestate` that will be used for a newly created `Span`. This will allow
samplers from a single vendor coordinate sampling decisions across the
different components of the trace.

### Probability Sampler default behavior

The default behavior of probability sampler is to calculate a stateless hash
function of `trace-id` to make a sampling decision.

It is suggested to extend the sampler behavior to expose a `tracestate` field
called `sampling.priority` with the integer value `[0-1000]` that will indicate
sampling priority of the current span.

This will allow to align sampling algorithms between various components.
Especially for the transition scenarios from the different SDKs.

Beyond the transition scenarios, propagation of `sampling.priority` will allow
to build more sophisticated logic of probability calculation that may be
customized to be consistent across properties beyond `trace-id`. Like
session-aware samplers. It also allows to use more secure probability
calculation that is NOT based on incoming parameters like `trace-id`. Especially
if there is an option to configure sampler to NOT respect `sampling.priority`
from the incoming `tracestate`.

## Internal details

Options on `sampling.priority` default behavior. Exposing `sampling.priority` is
not always needed. It also may be desired to not respect incoming
`sampling.priority`. So `ProbabilitySampler` may be configured out of the box to
respect incoming `sampling.priority`. And inserting it into `tracestate` when
not present. A setting can be exposed to NOT write `sampling.priority` to
`tracestate`.

## Trade-offs and mitigations

## Prior art and alternatives

## Open questions

## Future possibilities
