# Propagate head trace sampling probability

Use the W3C trace context to convey consistent head trace sampling probability.

## Motivation

The head trace sampling probability is the probability associated with
the start of a trace context that determines whether child contexts
are sampled or not when using the `ParentBased` Sampler.  It is useful
to know the head trace sampling probability associated with a context
in order to build span-to-metrics pipelines when the built-in
`ParentBased` Sampler is used.

A consistent trace sampling decision is one that can be carried out at
any node in a trace, which supports collecting partial traces.
OpenTelemetry specifies a built-in `TraceIDRatioBased` Sampler that
aims to accomplish this goal but was left incomplete (see a
[TODO](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/sdk.md#traceidratiobased) 
in the specification).

We propose to propagate the necessary information alongside the [W3C
sampled flag](https://www.w3.org/TR/trace-context/#sampled-flag) using
`tracestate` with an `otel` vendor tag, which will require
(separately) [specifying how the OpenTelemetry project uses
`tracestate` itself](https://github.com/open-telemetry/opentelemetry-specification/pull/1852).

## Explanation

Two pieces of information are needed to convey consistent head trace
sampling probability:

1. The head trace sampling probability
2. Source of consistent sampling decisions.

This proposal uses 6 bits of information for each of these and does
not depend on built-in TraceID randomness, which is not sufficiently
specified for probability sampling at this time.

### Probability value

To limit the cost of this extension and for statistical reasons
documented below, we propose to limit head trace sampling probability
to powers of two.  This limits the available head trace sampling
probabilities to 1/2, 1/4, 1/8, and so on.  We can compactly encode
these probabilities as small integer values using the base-2 logarithm
of the adjusted count (i.e., inverse probability).

For example, the probability value 2 corresponds with 1-in-4 sampling,
the probability value 10 corresponds with 1-in-1024 sampling.  Using
six bits of information we can convey sampling rates as small as
2^-62.  The value 63 is reserved to mean sampling with probability 0,
which conveys an adjusted count of 0 for the associated context.

### Random value

With head trace sampling probabilities limited to powers of two, the
amount of randomness needed per trace context is limited.  A
consistent sampling decision is accomplished by propagating a specific
random variable.  The random variable is a described by a discrete
geometric distribution having shape parameter `1/2`, listed below:

| Value | Probability |
| ----- | ----------- |
| 0 | 1/2 |
| 1 | 1/4 |
| 2 | 1/8 |
| 3 | 1/16 |
| 4 | 1/32 |
| ... | ... |
| N | 1/(2^(N+1)) |

See [Estimation from Partially Sampled Distributed
Traces](https://arxiv.org/pdf/2107.07703.pdf) section 2.8 for a
detailed explanation.

Such a random variable `r` can be generated using the following
pseudocode:

```
r := 0
for {
  if nextRandomBit() {
    break // The expected value of r is 2
  }
  r++
}
```

This can be computed from a stream of random bits as the number of
leading zeros using efficient instructions on modern computer
architectures.

For example, the value 3 means there were three leading zeros and
corresponds with being sampled at probabilities 1-in-1 through 1-in-8
but not at probabilities 1-in-16 and smaller.  Using one six bits of
information we can convey a consistent sampling decision for sampling
rates as small as 2^-62.  The value 63 is reserved to mean 0
probability.

### Proposed `tracestate` syntax

The consistent sampling decision and head trace sampling probability
will be propagated using four bytes of base16 content, as follows:

```
tracestate: otel=p:PP;r:RR
```

where `PP` are two bytes of base16 probability value and `RR` are two
bytes of base16 random value.

### Examples

The following `tracestate` value:

```
tracestate: otel=r:0a;p:03
```

translates to

```
base16(probability) = 03 // 1-in-8 head probability
base16(randomness) = 0a // qualifies for 1-in-1024 sampling or greater
```

Any `TraceIDRatioBased` Sampler configured with probability 2^-10 or
greater will enable sampling this trace, whereas any
`TraceIDRatioBased` Sampler configured with probability 2^-11 or less
will stop sampling this trace.  The W3C `sampled` flag is set to true
when the probability value is less than or equal to the randomness
value.

## Internal details

The reasoning behind restricting the set of sampling rates is that it:

- Lowers the cost of propagating head sampling probability
- Limits the number of random bits required
- Makes math involving partial traces tractable.

[An algorithm for making statistical inference from partially-sampled
traces has been published](https://arxiv.org/pdf/2107.07703.pdf) that
explains how to work with a limited number of power-of-2 sampling rates.

### Behavior of the `TraceIDRatioBased` Sampler

The Sampler must be configured with a power-of-two probability
`P=2^-S`.

Using one byte to represent both the probability and randomness values
means each value is limited to 255.  As a special case, the head
probability `P=0` is represented using the probability value `S=255`,
meaning `P=0` is indistinguishable from `P=2^-255`.  We propose to
limit valid head sampling probabilities to `P=2^-254` or greater to
address this ambiguity.

If the context is a new root, the initial `tracestate` must be created
using geometrically-distributed random value `R` (as described above,
with maximum value 254) and the initial head probability value `S`.

If the context is not a new root, output a new `tracestate` with the
same `R` value as the parent context, using the Sampler's own value of
`S` for the head probability.

In both cases, set the `sampled` bit if `S<=R`.

### Behavior of the `ParentBased` sampler

The `ParentBased` sampler is unmodified by this proposal.  It honors
the W3C `sampled` flag and copies the incoming `tracestate` keys to
the child context.

### Behavior of the `AlwaysOn` Sampler

The `AlwaysOn` Sampler behaves the same as `TraceIDRatioBased` with `P=1` (i.e., `S=0`)

### Behavior of the `AlwaysOff` Sampler

The `AlwaysOff` Sampler behaves the same as `TraceIDRatioBased` with `P=0` (i.e., `S=255`).

## Trade-offs and mitigations

Restricting head sampling rates to powers of two does not limit tail
Samplers from using arbitrary probabilities.  The
`sampler.adjusted_count` attribute specified in [OTEP
170](https://github.com/open-telemetry/oteps/pull/170) is not limited
to power-of-two values.

Restricting head sampling rates to powers of two does not limit
Samplers from using arbitrary effective probabilities over a period of
time.  For example, choosing 1/2 sampling half of the time and 1/4
sampling half of the time leads to an effective sampling rate of 3/8.

## Prior art and alternatives

Google's Dapper system propagated a field in its trace context called
"inverse_probability", which is equivalent to adjusted count.  This
proposal uses the base-2 logarithm of adjusted count to save space and
limit required randomness.
