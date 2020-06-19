# Add Incremental Sampling Hooks

This document proposes adding incremental sampling hooks, which will be called
as attributes are modified on spans to update sampling decisions.

## Motivation

As of the writing of this OTEP (June 2020), the SDK sampler interface
requires the choice to sample the span to occur when the span is created.
This forces patterns such as deferred span creation or span decisions to be
inconsistent with the criteria based on the final span result.

Adding additional sampling points to make sample decisions for the span will
reduce the possible incongruence of a sampler missing spans that should or
should not be sampled, based on attributes that change through the lifetime
of the span.

Although this document proposes adding additional sampling hooks, the desire
is to solve a deeper issue around contention of making an early sampling decision,
and enabling the addition or modification of span properties after initial creation.

This section argues the following:

1. deferred span creation is a common pattern in the current spec.
2. deferred span creation is always functionally worse than creating a span,
   and is at best equal.
3. based on 1 and 2, it is better to address the issues with sampling, rather
   than modify span creation behavior for sampling.

The following sections explains the rationale behind each argument:

### Deferred span creation is a common pattern

The current specification discourages the use of trace.updateName, as the
modification of the span name results in an incongruent sampling result (one
that was based upon the old name).

The current workaround is to defer span creation: to store attributes and
fields such as start time in some context object, to be used to create a span
later.

However, most operations that involves I/O will need to defer span creation, as
it is not possible to determine the appropriate span name until some data has
been received.

Examples include:

* http(s) requests: the path and route on the server is only known after enough
  of the payload is received to read the url.
* RPC requests: RPCs typically include the desired operation in the
  payload, after some bytes have been read.

The span of the operation should have a start time that is right before the
first byte is received.

### Deferred span creation is worse than creating a span

Ensuring a span is properly closed, even in error scenarios, is a desirable
trait, and SDKs have ensured best-effort of closing of spans using language
specific features such as exception handling or deferred close routines.

Deferred span creation has not had as much development focus, which results in
a less error handling that ensures that deferred spans are being created. This
results in exceptional situations that occur before the span creation (e.g.
invalid parsing of an http request) to not be represented in a span, as the
deferred span is never created.

A way to mitigate this would be to ensure the same level of error handling
for deferred span creation as with spans themselves (create the span during a
handled exception). But that would double the amount of maintenance to ensure
robust span creation. Without span sampling as a concern, it would be desirable
to just use a span instead, and focus on improving the robustness span creation.

There is another side effect of deferred span creation, where there may be
child spans that are created before the deferred span is. Examples here include
a span to represent the receiving of data: the span that represents receiving data
would have started before the name of the span is available, since that depends
on receiving and parsing the incoming payload. The result will at worst be
multiple root spans that should actually be children of the deferred span, since
there will be no active span present when the children are created.

### It is better to address the issues with sampling

Given the need for an almost identical, most likely less robust code path
for deferred spans, and the fact that it is a very common pattern, solving
issues with sampling would ensure a more robust and maintainable code base.

## Explanation

## Additional Hooks

A sampler would consume the same arguments it does today:

- span name
- span context
- span kind
- span attributes
- span links

The sampler would incorporate additional hooks, one to match each modification
of events on an existing span:

* onUpdateName, called by updateName
* onSetAttribute, called by setAttribute
* onAddEvent, called by addEvent
* onSetStatus, called by setStatus
* onEnd, called at the end of the span

These methods will consume both the span, as well as the field or state that
was changed. This enables optimizations in the case that samplers only depend
on the data in the modification to make a decision, as well as more complex
sample checks that may require all fields of the span.

onEnd is also provided, as an optimization for samplers who may opt for
running the sample decision after no further updates to the span are possible.

## Samples return shouldRetry

Samplers will return an additional "shouldRetry" boolean value, that indicates
whether the sample should be called again. Once shouldRetry has returned false,
further sample hooks will not be called.

shouldRetry is only valid when paired with a RECORD or RECORD_AND_SAMPLED result.
A NOT_RECORD result results in no further calls to the sampler, regardless of
the shouldRetry value.

## Internal details

## Trade-offs and mitigations

### Rationale for separating shouldRetry from sampled result

## Prior art and alternatives

This implementation should address the following issues:

- [remove warning on trace.updateName](https://github.com/open-telemetry/opentelemetry-specification/issues/468)
- [remove warning on trace.updateName PR](https://github.com/open-telemetry/opentelemetry-specification/pull/506)
- [Allow samplers to be called during different moments in the Span lifetime](https://github.com/open-telemetry/opentelemetry-specification/issues/307)
- [Sampling decision is too late to gain much performance](https://github.com/open-telemetry/opentelemetry-specification/issues/620)

### Tradeoffs of additional sample hooks

- Pro: enables setting the recording decision early, which can skip additional
       processing for instrumentations that use the isRecording field.
- Pro: enables setting the recording decision early, which can be used
       to set cheaper Span implementations (such as noop spans) and save on
       processing.
- Pro: more hooks enables earlier accurate sampling decisions, which in turn
       reduces the number of child spans and propagated spans that have an incongruent
       sampling decision.

- Con: increases the complexity of the sampling logic slightly.
- Con: introduces a complex matrix of final recording / sample decisions based
       on both sample results.
- Con: increases the processing required per span marginally (need to re-run
       sampling logic multiple times)

#### Alternative: Always sample at the end

An alternative approach is to move the existing sampling decision to the end.

- Pro: reduces complexity of sampling logic
- Pro: reduces complexity of choosing a noop span (no option to do so)

- Con: every span, sampled or not, will require the same processing. This will
       result in a net increase of system resources relative to an approach that
       allows optimizations.

## Open questions

### The behavior of child spans of an unsampled span

Question reference: [spec #307](https://github.com/open-telemetry/opentelemetry-specification/issues/307#issuecomment-544912240)

If the span decision can change over time, child spans cannot rely on the existence of a parent span. This may require some enforced behavior around child spans of an unsampled span also not being sampled. This is difficult, if not impossible, to do as there is no reference to child spans from the parent span. Thus, when the parent changes it's
span decision, it can not propagate that choice to child spans.

### An amended sample result will result in lost / additional spans when propagated

Reference: [comment in spec#307](https://github.com/open-telemetry/opentelemetry-specification/issues/307#issuecomment-642954936)

As the sampled result is propagated during external calls, a sampling result
that is changed later will result in some subset of propagated spans to have the
original decision, and future spans to have the later decision.

For example, suppose that an RPC request is sent right after span start with
sampled set to false. The consumers of those RPC requests will see a sample
decision of false, which means that the spans will not be emitted.

However if the span is set to be sampled later on, the original span in question
will be emitted, but as the RPC spans mentioned above were not, the final trace
will have the RPC spans missing.

## Changes to be applied to the spec

- replace any recommendation for deferred span creation with explicit
  discouragement of the pattern
- modification of sampler API with new methods
- modification of sampler return value to include shouldRetry

Recommendations for setting span parameters as early as possible will remain,
as it will still benefit sampler accuracy to have values as early as possible,
and thus have a final decision as soon as possible.
