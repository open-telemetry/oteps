# Sampling Threshold Propagation in TraceState

## Motivation

Sampling can theoretically take place at nearly any point in a distributed tracing system. If sampling is to be performed at multiple points in the process, the only way to reason about it effectively is to make sure that the sampling decisions are **consistent**.
In this context, consistency means that a positive sampling decision made at a particular point with probability p1 implies a positive sampling decision made at another point that samples a different piece of information from the same trace with probability p2 >= p1.

## Explanation

The existing, experimental [specification for probability sampling using TraceState](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/tracestate-probability-sampling.md) is limited to powers-of-two probabilities, and is designed to work without making assumptions about TraceID randomness.
This system can only achieve non-power-of-two sampling using interpolation between powers of two, which is unnecessarily restrictive.
In existing sampling systems, sampling probabilities like 1%, 10%, and 75% are common, and it should be possible to express these without interpolation.
There is also a need for consistent sampling in the collection path (outside of the head-sampling paths) and using inherent randomness in the traceID is a less-expensive solution than referencing a custom `r-value` from the tracestate in every span.
This proposal introduces a new value with the key `th` as a replacement for the `p` value in the previous specification.
The `p` value is limited to powers of two, while this proposal is not.
This allows for the continued expression of randomness using `r-value` as specified there using the key `r`.
To distinguish the cases, this proposal uses the key `rv`.

In order to make consistent sampling decisions across the entire path of the trace, two values SHOULD be propagated with the trace:

1. A _random_ (or pseudo-random) 56-bit value, called `R` below.
2. A 56-bit trace _threshold_ as expressed in the TraceState, called `T` below.

The sampling decision is propagated with the following algorithm:

* If the `th` key is not specified, Always Sample.
* Else derive `T` by parsing the `th` key as a hex value as described below.
* If `T` is 0 and the _sampled_ flag is set, Always Sample. This implies that non-probabilistic sampling is taking place.
* Compare the 56 bits of `T` with the 56 bits of `R`. If `T <= R`, then do not sample.

The `R` value MUST be derived as follows:

* If the key `rv` is present in the Tracestate header, then `R = rv`.
* Else if the Random Trace ID Flag is `true` in the traceparent header, then `R` is the lowest-order 56 bits of the trace-id.
* Else `R` MUST be generated as a random value in the range `(0, (2**56)-1)` and added to the Tracestate header with key `rv`.

The preferred way to propagate the `R` value is as the lowest 56 bits of the trace-id.
If these bits are in fact random, the `random` trace-flag SHOULD be set as specified in [the W3C trace context specification](https://w3c.github.io/trace-context/#trace-id).
There are circumstances where trace-id randomness is inadequate (for example, sampling a group of traces together); in these cases, an `rv` value is required.

The value of the `rv` and `th` keys MUST be expressed as up to 14 hexadecimal characters from the set `[0-9a-f]`. Trailing zeros (but not leading zeros) may be omitted.

Examples:
`th` value is missing: Always Sample (probability = 100%). The AlwaysSample sampler in the OTel SDK should do this.
`th=8` -- equivalent to `th=80000000000000`, which is 50% probability.
`th=08` -- equivalent to `th=08000000000000`, which is 3.125% probability.
`th=0` -- equivalent to `th=00000000000000`, which means Always Sample; this is outside of probabalistic sampling.

The `T` value MUST be derived as follows:

* If the `th` key is not present in the Tracestate header, then `T` is effectively 2^56 (which doesn't fit in 56 bits).
* Else the value corresponding to the `th` key should be interpreted as above.

Sampling Decisions SHOULD be propagated by setting the value of the `th` key in the Tracestate header according to the above.

## Changing T and R values

The T value MAY be modified.

In the case of a downstream sampler -- a tail sampler on the collection path that is attempting to reduce the volume of traffic -- the sampler MAY modify the `th` header by reducing its value.
It MAY NOT increase it, as it is not possible to retroactively adjust the sampling probability upward.

A non-root head sampler MAY raise or lower the T value.
Note that changing the probability of a trace in flight introduces inconsistency and may cause the trace to be incomplete.

A sampler MUST introduce an R value to a trace that does not include one and does not have the `Random` trace-id flag set. It MUST use the `rv` key for this purpose. A sampler MUST NOT modify an existing R value or trace-id.

## Internal details

The trace state header SHOULD contain a field with the key `rv`, and a value that corresponds to a 56-bit sampling threshold.
This value will be compared to the 56-bit random value associated with the trace.

From a technical perspective, how do you propose accomplishing the proposal? In particular, please explain:

* How the change would impact and interact with existing functionality
* Likely error modes (and how to handle them)
* Corner cases (and how to handle them)

While you do not need to prescribe a particular implementation - indeed, OTEPs should be about **behaviour**, not implementation! - it may be useful to provide at least one suggestion as to how the proposal *could* be implemented. This helps reassure reviewers that implementation is at least possible, and often helps them inspire them to think more deeply about trade-offs, alternatives, etc.

## Trade-offs and mitigations

This proposal is the result of long negotiations on the Sampling SIG over what is required and various alternative forms of expressing it. [This issue](https://github.com/open-telemetry/opentelemetry-specification/issues/3602) exhaustively covers the various formats that were discussed and their pros and cons. This proposal is the result of that decision.

## Prior art and alternatives

The existing specification for `r-value` and `p-value` attempted to solve this problem, but were limited to powers of 2, which is inadequate.

## Open questions

This specification leaves room for different implementation options. For example, comparing hex strings or converting them to numeric format are both viable alternatives for handling the threshold.

We also know that some implementations prefer to use a sampling probability (in the range from 0-1.0) or a sampling rate (1/probability); this design permits conversion to and from these formats without loss up to at least 6 decimal digits of precision.

## Future possibilities

This permits sampling systems to propagate consistent sampling information downstream where it can be compensated for.
For example, this will enable the tail-sampling processor in the OTel Collector to propagate its sampling decisions to backends in a standard way.
This permits backend systems to use the effective sampling probability in data presentations.
