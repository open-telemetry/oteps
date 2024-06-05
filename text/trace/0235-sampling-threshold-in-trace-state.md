# Sampling Threshold Propagation in TraceState

## Abstract

Sampling is an important lever to reduce the costs associated with collecting and processing telemetry data. It enables you to choose a representative set of items from an overall population.

There are two key aspects for sampling of tracing data. The first is that sampling decisions can be made independently for *each* span in a trace. The second is that sampling decisions can be made at multiple points in the telemetry pipeline. For example, the sampling decision for a span at span creation time could have been to **keep** that span, while the downstream sampling decision for the *same* span at a later stage (say in an external process in the data collection pipeline) could be to **drop** it.

For each of the above aspects, we want sampling decisions to be made in a **consistent** manner so that we can effectively reason about a trace. This OTEP describes a mechanism to achieve such consistent sampling decisions using a mechanism called **Consistent Probability Sampling**. To achieve this, it proposes a mechanism for a common random value (R) and a rejection threshold (T) that is based on a participant's sampling rate. This proposal describes how these values should be propagated and how participants should use them to make sampling decisions.

This mechanism will enable creating a new set of samplers (known as Consistent Probability Samplers) that will enable trace participants to choose their own sampling rates, while still achieving consistent sampling decisions. This OTEP ensures that such samplers will interoperate with existing (non consistent probability) samplers.

## Motivation

Customers want to express arbitrary sampling probabilities such as 1%, 10%, and 75%. However, the existing experimental [specification for probability sampling using TraceState](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/tracestate-probability-sampling.md) optimizes for powers of two probabilities. It supports non power of two sampling only using interpolation between powers of two. This approach is unnecessarily restrictive. Hence, we need an updated mechanism to support specifying any sampling probability.

Further, there is a need for consistent sampling in the collection path (outside of the head-based sampling paths). To achieve consistent sampling decisions, the previous experimental spec required using a custom source of randomness (`r-value`). However, in such downstream sampling decisions, it can be expensive to reference this custom value from the tracestate attribute in every span. To improve this, this proposal makes use of the inherent randomness in the traceID as a less expensive solution. However, one caveat is that the new randomness flag introduced in the W3C TraceContext Level 2 specification can potentially be reset by trace participants until they move to that Level 2 specification. Hence, there is need to still reference tracestate to check for the non-existence of this custom random value before relying on the traceid as the source of randomness.

## Explanation
Let's start with the definition for a consistent sampling decision. Consistency means that a positive sampling decision made for a particular span with probability p1 implies a positive sampling decision for any span belonging to the same trace if it is made with probability p2 >= p1.

This proposal introduces a new value with the key `th` as an alternative to the `p` value in the previous specification. The `p` value is limited to powers of two, while the `th` value in this proposal supports a large range of values.

This proposal allows for the continued expression of randomness using `r-value` as specified there using the key `r`. To distinguish the cases, this proposal uses the key `rv`.

In the general case, in order to make consistent sampling decisions for the two aspects described above, two values MUST be present in the `SpanContext`:

1. A _random_ (or pseudo-random) 56-bit value, called `R` below.
2. A 56-bit _rejection threshold_ (or just "threshold") as expressed in the TraceState, called `T` below. `T` represents the maximum threshold that was applied in all previous consistent sampling stages. If the current sampling stage applies a greater threshold value than any stage before, it MUST update (increase) the threshold correspondingly.

One way to think about _rejection threshold_ is that it is the number of spans that would be discarded out of 2^56 considered spans. This means that spans where `R >= T` will be kept.

Here is an example involving three participating operations `A`, `B`, and `C`:

`A` -> `B` -> `C`

where -> indicates a parent to child relationship.

`A` uses consistent probability sampling with a sampling probability of 0.25 (this corresponds to a rejection probability of .75).
`B` uses consistent probability sampling with a sampling probability of 0.5.
`C` uses a parent-based sampler.

When the sampling decision for `A` is to *keep* the span, its outgoing traceparent will have the 'sampled' flag SET and the 'th' in its outgoing tracestate will be set to `0xc0_0000_0000_0000`.
When the sampling decision for `A` is to *drop* the span, its outgoing traceparent will have the 'sampled' flag UNSET but the 'th' in its outgoing tracestate will still be set to `0xc0_0000_0000_0000`.
When the sampling decision for `B` is to *keep* the span, its outgoing traceparent will have the 'sampled' flag SET and the 'th' in its outgoing tracestate will be set to `0x80_0000_0000_0000`.
C (being a parent based sampler) samples a span purely based on its parent (B in this case), it will use the sampled flag to make the decision. Its outgoing 'th' value will continue to reflect what it got from B (`0x80_0000_0000_0000`), and this is useful to understand its adjusted count.

This design requires that as a given span progresses along its collection path, `th` is non-decreasing (and, in particular, must be increased at stages that apply lower sampling probabilities).
It does not, however, restrict a span's initial `th` in any way. If a parent-based consistent sampler is used, a span's initial `th` would be the same as its parent's `th` value, else it would be a new value based on the sampling rate chosen for that span. In other words, the sampling rate for each operation can be chosen independently, and this would map to having different `th` values for different spans. But for any particular span, it is not acceptable for a downstream sampler to *decrease* the `th` value in its context.

The system has the following invariant:

`(R >= T) = sampled flag`

The sampling decision is propagated with the following algorithm:

* If the `th` key is not specified, this implies that non-probabilistic sampling may be taking place.
* Else derive `T` by parsing the `th` key as a hex value as described below.
* If `T` is 0, Always Sample.
* Compare the 56 bits of `T` with the 56 bits of `R`. If `R >= T`, then set the sampling decision to *keep* else make the decision to *drop*.

The `R` value MUST be derived as follows:

* If the key `rv` is present in the Tracestate header, then `R = rv`.
* Else `R` is the lowest-order 56 bits of the trace-id.

At the root span, the `R` value must be generated as follows:

* If the new random flag in the `traceparent` is set, then there is no action required. In this case, the tracestate header will not have the `rv` key, and the last 56 bits of the traceid will be used as the source of randomness. For more info on this new flag, see [the W3C trace context specification](https://w3c.github.io/trace-context/#trace-id).
* If not, `R` MUST be generated as a random value in the range `[0, (2**56)-1]` and added to the Tracestate header with key `rv`.

Although less common, there are circumstances where trace-id randomness is inadequate (for example, when sampling a group of traces together); in these cases, an `rv` value is required.

The value of the `rv` and `th` keys MUST be expressed as up to 14 hexadecimal digits from the set `[0-9a-f]`. For `th` keys only, trailing zeros (but not leading zeros) may be omitted. `rv` keys MUST always be exactly 14 hex digits.

Examples:

- `th` value is missing: non-probabalistic sampling may be taking place.
- `th=0` -- equivalent to `th=00000000000000`, which is a 0% rejection threshold, corresponding to 100% sampling probability (Always Sample).
- `th=08` -- equivalent to `th=08000000000000`, which is a rejection threshold of 3.125%, corresponding to 96.875% sampling probability.
- `th=4` -- equivalent to `th=40000000000000`, which is a 25% rejection threshold, corresponding to 75% sampling probability.
- `th=c` -- equivalent to `th=c0000000000000`, which is a rejection threshold of 75%, corresponding to 25% sampling probability.

The `T` value MUST be derived as follows:

* If the `th` key is not present in the Tracestate header, then non-probabalistic sampling may be in use.
* Else the value corresponding to the `th` key should be interpreted as above.

Sampling Decisions MUST be propagated by setting the value of the `th` key in the Tracestate header according to the above.

## Initializing and updating T and R values

There are two categories of samplers:

- **Head samplers:** Implementations of [`Sampler`](https://github.com/open-telemetry/opentelemetry-specification/blob/v1.29.0/specification/trace/sdk.md#sampler), called by a `Tracer` during span creation.
- **Downstream samplers:** Any component that, given an ended Span, decides whether to *drop* or *keep* it by forwarding it to the next component in the system. This category is also known as "collection path samplers" or "sampling processors". _Tail samplers_ are a special class of downstream samplers that buffer spans of a trace and make a sampling decision for the trace as a whole using data from any span in the buffered trace.

This section defines behavior for each kind of sampler.

### Head samplers

A head sampler is responsible for computing the `rv` and `th` values in a new span's initial [`TraceState`](https://github.com/open-telemetry/opentelemetry-specification/blob/v1.29.0/specification/trace/api.md#tracestate). The main inputs to that computation include the parent span's trace state (if a parent span exists), the new span's trace ID, and possibly the trace flags (to know if the trace ID has been generated in a random manner).

First, a consistent probability `Sampler` may choose its own sampling rate. The higher the chosen sampling rate, the lower the rejection threshold (T). It MAY select any value of T. If a valid `SpanContext` is provided in the call to `ShouldSample` (indicating that the span being created will be a child span),

- Choosing a T greater than the parent span's T can result in partial traces. The parent span may be `kept` but it is possible that its child, the current span, may be dropped because of the lower sampling rate. At the same time, in case where the child span is `kept`, the parent span would have been to `keep` as well (meeting our consistent sampling goals) since the parent's sampling rate is greater than the child's sampling rate.
- Similarly, choosing a T less than or equal to the parent span can also result in partial traces. The parent span might have been `dropped` but it is possible that its child, the current span, may be `kept` because of the higher sampling rate. At the same time, in case where the parent span is `kept`, the child span would be `kept` as well (meeting our consistent sampling goals) since the child's sampling rate is greater than the parent's sampling rate.

For the output TraceState,

- The `th` key MUST be defined with a value corresponding to the sampling probability the sampler actually used.
- The `rv` value, if present on the input TraceState, MUST be defined and equal to the parent span's `rv`. Otherwise, `rv` MUST be defined if and only if the effective R was _generated_ during the decision, per the "derive R" algorithm given earlier.

TODO: For _new_ spans, `ShouldSample` doesn't currently have a way to know the new Span's `TraceFlags`, so it can't determine whether the Random Trace ID Flag is set, and in turn can't execute the "derive R" algorithm. Maybe it should take `TraceFlags` as an additional parameter, just like it takes `TraceId`?

### Downstream samplers

A downstream sampler, in contrast, may output a given ended Span with a _modified_ trace state, complying with following rules:

- If the chosen sampling probability is 1, the sampler MUST NOT modify any existing `th`, nor set any `th`.
- Otherwise, the chosen sampling probability is in `(0, 1)`. In this case the sampler MUST output the span with a `th` equal to `max(input th, chosen th)`. In other words, `th` MUST NOT be decreased (as it is not possible to retroactively adjust an earlier stage's sampling probability), and it MUST be increased if a lower sampling probability was used. This case represents the common case where a downstream sampler is reducing span throughput in the system.

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

Given rv and threshold as 64-bit integers, a sample should be taken if rv is greater than or equal to the threshold.

```
shouldSample = (rv >= threshold)
```

### Converting threshold to a sampling probability

The sampling probability is a value from 0.0 to 1.0, which can be calculated using floating point by dividing by 2^56:

```py
# embedded _ in numbers for clarity (permitted by Python3)
maxth = 0x100_0000_0000_0000  # 2^56
prob = float(maxth - threshold) / maxth
```

### Converting threshold to an adjusted count (sampling rate)

The adjusted count indicates the approximate quantity of items from the population that this sample represents. It is equal to `1/probability`. It is not defined for spans that were obtained via non-probabilistic sampling (a sampled span with no `th` value).

## Trade-offs and mitigations

This proposal is the result of long negotiations on the Sampling SIG over what is required and various alternative forms of expressing it. [This issue](https://github.com/open-telemetry/opentelemetry-specification/issues/3602) exhaustively covers the various formats that were discussed and their pros and cons. This proposal is the result of that decision.

## Prior art and alternatives

The existing specification for `r-value` and `p-value` attempted to solve this problem, but was limited to powers of 2, which is inadequate.

## Open questions

This specification leaves room for different implementation options. For example, comparing hex strings or converting them to numeric format are both viable alternatives for handling the threshold.

We also know that some implementations prefer to use a sampling probability (in the range from 0-1.0) or a sampling rate (1/probability); this design permits conversion to and from these formats without loss up to at least 6 decimal digits of precision.

## Future possibilities

This permits sampling systems to propagate consistent sampling information downstream where it can be compensated for.
For example, this will enable the tail-sampling processor in the OTel Collector to propagate its sampling decisions to backend systems in a standard way.
This permits backend systems to use the effective sampling probability in data presentations.
