# Sampling Threshold Propagation in TraceState

## Motivation

Sampling is a broad topic; here it refers to the independent decisions made at points in a distributed tracing system of whether to collect a span or not. Multiple sampling decisions can be made before a span is finally consumed. When sampling is to be performed at multiple points in the process, the only way to reason about it effectively is to make sure that the sampling decisions are **consistent**.
In this context, consistency means that a positive sampling decision made for a particular span with probability p1 implies a positive sampling decision for any span belonging to the same trace, if it is made with probability p2 >= p1.

## Explanation

The existing, experimental [specification for probability sampling using TraceState](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/tracestate-probability-sampling.md) is limited to powers-of-two probabilities, and is designed to work without making assumptions about TraceID randomness.
This system can only achieve non-power-of-two sampling using interpolation between powers of two, which is unnecessarily restrictive.
In existing sampling systems, sampling probabilities like 1%, 10%, and 75% are common, and it should be possible to express these without interpolation.
There is also a need for consistent sampling in the collection path (outside of the head-sampling paths) and using inherent randomness in the traceID is a less-expensive solution than referencing a custom `r-value` from the tracestate in every span.
This proposal introduces a new value with the key `th` as a replacement for the `p` value in the previous specification.
The `p` value is limited to powers of two, while the `th` value in this proposal supports a large range of values.
This proposal allows for the continued expression of randomness using `r-value` as specified there using the key `r`.
To distinguish the cases, this proposal uses the key `rv`.

In order to make consistent sampling decisions across the entire path of the trace, two values SHOULD be propagated with the trace:

1. A _random_ (or pseudo-random) 56-bit value, called `R` below.
2. A 56-bit trace _threshold_ as expressed in the TraceState, called `T` below. `T` represents the minimum threshold that was applied in all previous consistent sampling stages. If the current sampling stage applies a lower threshold than any stage before, it has to update (decrease) the threshold correspondingly.

Here is an example involving three participants A, B, and C:

A -> B -> C

where -> indicates a parent -> child relationship.

A uses consistent probability sampling with a sampling rate of 0.25.
B uses consistent probability sampling with a sampling rate of 0.5.
C uses a parent-based sampler.

When A samples a span, its outgoing traceparent will have the 'sampled' flag SET and the 'th' in its outgoing tracestate will be set to 0x40 0000 0000 0000.
When A does not sample a span, its outgoing traceparent will have the 'sampled' flag UNSET but the 'th' in its outgoing tracestate will still be set to 0x40 0000 0000 0000.
When B samples a span, its outgoing traceparent will have the 'sampled' flag SET and the 'th' in its outgoing tracestate will be set to 0x80 0000 0000 0000.
C (being a parent based sampler) samples a span purely based on its parent (B in this case), it will use the sampled flag to make the decision. Its outgoing 'th' value will continue to reflect what it got from B (0x80 0000 0000 0000), and this is useful to understand its adjusted count.

This design requires that as a given span progresses along its collection path, `th` is non-increasing (and, in particular, must be decreased at stages that apply lower sampling probabilities).
It does not, however, restrict a span's initial `th` in any way (e.g., relating it to that of its parent, if it has one).
It is acceptable for B to have a greater initial `th` than A has. It would not be ok if some later-stage sampler increased A's `th`.

The system has the following invariant:

`(T=0) OR ((R < T) = sampled flag)`

The sampling decision is propagated with the following algorithm:

* If the `th` key is not specified, Always Sample.
* Else derive `T` by parsing the `th` key as a hex value as described below.
* If `T` is 0 and the _sampled_ flag is set, Always Sample. This implies that non-probabilistic sampling is taking place.
* Compare the 56 bits of `T` with the 56 bits of `R`. If `T <= R`, then do not sample.

The `R` value MUST be derived as follows:

* If the key `rv` is present in the Tracestate header, then `R = rv`.
* Else if the Random Trace ID Flag is `true` in the traceparent header, then `R` is the lowest-order 56 bits of the trace-id.
* Else `R` MUST be generated as a random value in the range `[0, (2**56)-1]` and added to the Tracestate header with key `rv`.

The preferred way to propagate the `R` value is as the lowest 56 bits of the trace-id.
If these bits are in fact random, the `random` trace-flag SHOULD be set as specified in [the W3C trace context specification](https://w3c.github.io/trace-context/#trace-id).
There are circumstances where trace-id randomness is inadequate (for example, sampling a group of traces together); in these cases, an `rv` value is required.

The value of the `rv` and `th` keys MUST be expressed as up to 14 hexadecimal digits from the set `[0-9a-f]`. For `th` keys only, trailing zeros (but not leading zeros) may be omitted. `rv` keys MUST always be exactly 14 hex digits.

Examples:
`th` value is missing: Always Sample (probability = 100%). The AlwaysOn sampler in the OTel SDK should do this.
`th=8` -- equivalent to `th=80000000000000`, which is 50% probability.
`th=08` -- equivalent to `th=08000000000000`, which is 3.125% probability.
`th=0` -- equivalent to `th=00000000000000`, which means Always Sample; this is outside of probabalistic sampling.

The `T` value MUST be derived as follows:

* If the `th` key is not present in the Tracestate header, then `T` is effectively 2^56 (which doesn't fit in 56 bits).
* Else the value corresponding to the `th` key should be interpreted as above.

Sampling Decisions MUST be propagated by setting the value of the `th` key in the Tracestate header according to the above.

## Changing T and R values

The T value MAY be modified.

In the case of a downstream sampler -- a tail sampler on the collection path that is attempting to reduce the volume of traffic -- the sampler MAY modify the `th` header by reducing its value.
It MUST NOT increase it, as it is not possible to retroactively adjust the sampling probability upward.

A consistent head sampler MUST set the T value corresponding to the sampling probability it actually uses. If it samples a non-root span, it MAY use the sampling probability of the parent span and use its T value.
Using different sampling probabilities for spans belonging to the same trace will lead to incomplete traces.

A sampler MUST introduce an R value to a trace that does not include one and does not have the `Random` trace-id flag set. It MUST use the `rv` key for this purpose. A sampler MUST NOT modify an existing R value or trace-id.

## Internal details

The trace state header SHOULD contain a field with the key `rv`, and a value that corresponds to a 56-bit sampling threshold.
This value will be compared to the 56-bit random value associated with the trace.

## Visual

![Sampling decision flow](../img/0235-sampling-threshold-calculation.png)

## Algorithms

The `th` and `rv` values may be represented and manipulated in a variety of forms depending on the capabilities of the processor and needs of the implementation. As 56-bit values, they are compatible with byte arrays and 64-bit integers, and can also be manipulated with 64-bit floating point with a truly negligible loss of precision.

The following examples are in Python3. They are intended as examples only for clarity, and not as a suggested implementation.

### Converting t-value to a 56-bit integer threshold

To convert a t-value string to a 56-bit integer threshold, pad it on the right with 0s so that it is 14 digits in length, and then parse it as a hexadecimal value.

```py
padded = (tvalue + "00000000000000")[:14]
threshold = int('0x' + padded, 16)
```

### Converting integer threshold to a t-value

To convert a 56-bit integer threshold value to the t-value representation, emit it as a hexadecimal value (without a leading '0x'), optionally with trailing zeros omitted:

```py
h = hex(tvalue).rstrip('0')
# remove leading 0x
tv = 'tv='+h[2:]
```

### Testing rv vs threshold

Given rv and threshold as 64-bit integers, a sample should be taken if rv is strictly less than the threshold.

```
shouldSample = (rv < threshold)
```

### Converting threshold to a sampling probability

The sampling probability is a value from 0.0 to 1.0, which can be calculated using floating point by dividing by 2^56:

```py
# embedded _ in numbers for clarity (permitted by Python3)
maxth = 0x100_0000_0000_0000  # 2^56
prob = float(threshold) / maxth
```

### Converting threshold to an adjusted count (sampling rate)

The adjusted count is an integer value, indicating the quantity of items from the population that this sample represents. It is 1/probability converted to an integer.

```py
maxth = 0x100_0000_0000_0000  # 2^56
adjcount = int((maxth / float(threshold)) + 0.5)
```

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
