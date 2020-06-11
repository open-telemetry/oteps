# Add Additional Sample Decision at Span End

This document proposes adding an additional sampling decision made at span end,
along with keeping the current sampling decision made at span creation.

## Motivation

As of the writing of this OTEP (June 2020), the SDK sampler interface
requires the choice to sample the span to occur when the span is created.
This forces patterns such as deferred span creation or span decisions to be
inconsistent with the criteria based on the final span result.

Adding an additional sampling decision to the span end will eliminate
the possible incongruence of a sampler missing spans that should or should not be
sampled, based on attributes that change through the lifetime of the span.

Although this document suggests adding an additional sampling point, the desire
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

A sampler would consume the same arguments it does today:

- span name
- span context
- span kind
- span attributes
- span links

The sampler would be called twice: once at creation, and once at span end.

The span's sampling decision would be determined by the intersection of the
sampling decision made by either ShouldSample call:

| First ShouldSample result | Second ShouldSample result | Final Result       |
|---------------------------|----------------------------|--------------------|
| NOT_RECORD                | N/A (Don't Call)           | NOT_RECORD         |
| RECORD_AND_SAMPLED        | N/A (Don't Call)           | RECORD_AND_SAMPLED |
| RECORD                    | NOT_RECORD                 | NOT_RECORD         |
| RECORD                    | RECORD                     | RECORD             |
| RECORD                    | RECORD_AND_SAMPLED         | RECORD_AND_SAMPLED |

## Internal details

### Idempotence of samplers

The calling of shouldSample multiple times can result in inconsistent results
if the sampler is truly random. As such, the samplers will need to be idempotent
based on the input passed. For example, a probability sampler would need to
use data present in the passed parameters, such as the span id, to return sampling
decisions.

### Sampling

## Trade-offs and mitigations

See alternatives.

## Prior art and alternatives

Related issues:

- [remove warning on trace.updateName](https://github.com/open-telemetry/opentelemetry-specification/issues/468)
- [remove warning on trace.updateName PR](https://github.com/open-telemetry/opentelemetry-specification/pull/506)
- [Allow samplers to be called during different moments in the Span lifetime](https://github.com/open-telemetry/opentelemetry-specification/issues/307)
- [Sampling decision is too late to gain much performance](https://github.com/open-telemetry/opentelemetry-specification/issues/620)

### Tradeoffs of multiple sampled approach

- Pro: enables setting the recording decision early, which can skip additional
       processing for instrumentations that use the isRecording field.
- Pro: enables setting the recording decision early, which can be used
       to set cheaper Span implementations (such as noop spans) and save on
       processing.

- Con: increases the complexity of the sampling logic slightly.
- Con: introduces a complex matrix of final recording / sample decisions based
       on both sample results.
- Con: increases the processing required per span marginally (need to re-run sampling logic)

#### Alternative: Always sample at the end

An alternative approach is to move the existing sampling decision to the end.

- Pro: reduces complexity of sampling logic
- Pro: reduces complexity of choosing a noop span (no option to do so)

- Con: every span, sampled or not, will require the same processing. This will
       result in a net increase of system resources relative to an approach that
       allows optimizations.

### Alternative: more hooks for the sampler to update it's decision

[spec #307](https://github.com/open-telemetry/opentelemetry-specification/issues/307) brings up
a few more hooks to modify sampling decisions. This is similar to the pros and
cons of the proposal, with an added pro of requiring more incremental effort
on determining sampler decisions, and a con of requiring more complexity on
the sampler's part.

## Open questions

### The behavior of child spans of an unsampled span

Question reference: [spec #307](https://github.com/open-telemetry/opentelemetry-specification/issues/307#issuecomment-544912240)

If the span decision can change over time, child spans cannot rely on the existence of a parent span. This may require some enforced behavior around child spans of an unsampled span also not being sampled. This is difficult, if not impossible, to do as there is no reference to child spans from the parent span. Thus, when the parent changes it's
span decision, it can not propagate that choice to child spans.

### Sampling decision is too late to gain much performance

Reference: [spec #620](https://github.com/open-telemetry/opentelemetry-specification/issues/620)

Some samplers can make the choice to sample with effectively no data,
such as a probabilistic sampler. There is currently no facility to enable
these types of samplers.
