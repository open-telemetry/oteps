# Propagate head trace sampling probability

Use the W3C trace context to convey consistent head trace sampling probability.

## Motivation

The head trace sampling probability is the probability associated with
the start of a trace context that was used to determine whether the
W3C `sampled` flag is set, which determines whether child contexts
will be sampled by a `ParentBased` Sampler.  It is useful to know the
head trace sampling probability associated with a context in order to
build span-to-metrics pipelines when the built-in `ParentBased`
Sampler is used.  Further motivation for supporting span-to-metrics
pipelines is presented in [OTEP
170](https://github.com/open-telemetry/oteps/pull/170).

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
specified for probability sampling at this time.  This proposal closely 
follows [research by Otmar Ertl](https://arxiv.org/pdf/2107.07703.pdf).

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
2^-61.  The value 62 is reserved to mean sampling with probability 0,
which conveys an adjusted count of 0 for the associated context.

When propagated the probability value will be interpreted as shown in
the following table:

| Probability Value | Head Probability |
| ----- | ----------- |
| 0 | 1 |
| 1 | 1/2 |
| 2 | 1/4 |
| 3 | 1/8 |
| ... | ... |
| N | 2^-N |
| ... | ... |
| 61 | 2^-61 |
| 62 | 0 |
| 63 | _Reserved_ |

The value 63 is reserved for use in encoding adjusted count in Span
data.  [Described in OTEP
170](https://github.com/open-telemetry/oteps/pull/170), Span data
would encode the probability value described here offset by +1, when
the adjusted count is known, and would encode 0 when the adjusted
count is unknown.

### Randomness value

With head trace sampling probabilities limited to powers of two, the
amount of randomness needed per trace context is limited.  A
consistent sampling decision is accomplished by propagating a specific
random variable denoted `R`.  The random variable is a described by a
discrete geometric distribution having shape parameter `1/2`, listed
below:

| `R` Value | Selection Probability |
| ---------------- | --------------------- |
| 0 | 1/2 |
| 1 | 1/4 |
| 2 | 1/8 |
| 3 | 1/16 |
| 4 | 1/32 |
| ... | ... |
| 0 <= `R` <= 62 | 1/(2^(`R`+1)) |
| ... | ... |
| 62 | 2^-63 |
| `R` >= 63 | Reject |

Such a random variable `R` can be generated using the following
pseudocode.  Note there is a tiny probability that the code has to
reject the calculated result and start over, since the value 62 is
defined to have adjusted count 0, not 2^62.

```golang
func nextRandomness() int {
  // Repeat until a valid result is produced.
  for {
    R := 0
    for {
      if nextRandomBit() {
        break
      }
      R++
    }
    // The expected value of R is 2.
	if R < 63 {
	  return R
    }
	// Reject, try again.
  }
}
```

This can be computed from a stream of random bits as the number of
leading zeros using efficient instructions on modern computer
architectures.

For example, the value 3 means there were three leading zeros and
corresponds with being sampled at probabilities 1-in-1 through 1-in-8
but not at probabilities 1-in-16 and smaller.

### Proposed `tracestate` syntax

The consistent sampling decision and head trace sampling probability
will be propagated using four bytes of base16 content, as follows:

```
tracestate: otel=p:PP;r:RR
```

where `PP` are two bytes of base16 probability value and `RR` are two
bytes of base16 random value.  These values are omitted when they are
unknown.

This proposal should be taken as a recommendation and will be modified
to [match whatever format OpenTelemtry specifies for its
`tracestate`](https://github.com/open-telemetry/opentelemetry-specification/pull/1852).
The choice of base16 encoding is therefore just a recommendation,
chosen because `traceparent` uses base16 encoding.

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
- Avoids floating-point to integer rounding errors
- Makes math involving partial traces tractable.

[An algorithm for making statistical inference from partially-sampled
traces has been published](https://arxiv.org/pdf/2107.07703.pdf) that
explains how to work with a limited number of power-of-2 sampling rates.

### Behavior of the `TraceIDRatioBased` Sampler

The Sampler must be configured with a power-of-two probability
`P=2^-S` except for the special case of `P=0`, which is handled
specially.

If the context is a new root, the initial `tracestate` must be created
using geometrically-distributed random value `R` (as described above,
with maximum value 61) and the initial head probability value `S`.  If
the head probability is zero (i.e., `P=0`) use `S=62`, the specified
value for zero probability.

If the context is not a new root, output a new `tracestate` with the
same `R` value as the parent context, and this Sampler's value of `S`
for the outgoing context's probability value (i.e., as the value for
`P`).

In both cases, set the `sampled` bit if `S<=R` and `S<62`.

### Behavior of the `ParentBased` sampler

The `ParentBased` sampler is unmodified by this proposal.  It honors
the W3C `sampled` flag and copies the incoming `tracestate` keys to
the child context.

### Behavior of the `AlwaysOn` Sampler

The `AlwaysOn` Sampler behaves the same as `TraceIDRatioBased` with `P=1` (i.e., `S=0`)

### Behavior of the `AlwaysOff` Sampler

The `AlwaysOff` Sampler behaves the same as `TraceIDRatioBased` with `P=0` (i.e., `S=62`).

## Prototype

[This proposal has been prototyped in the OTel-Go
SDK.](https://github.com/open-telemetry/opentelemetry-go/pull/2177) No
changes in the OTel-Go Tracing SDK's `Sampler` or `tracestate` APIs
were needed.

## Trade-offs and mitigations

### Not using TraceID randomness

It would be possible, if TraceID were specified to have at least 62
uniform random bits, to compute the randomness value described above
as the number of leading zeros among those 62 random bits.

This proposal requires modifying the W3C traceparent specification,
therefore we do not propose to use bits of the TraceID.

[This issue has been filed with the W3C trace context group.](https://github.com/w3c/trace-context/issues/463)

### Not using TraceID hashing

It would be possible to make a consistent sampling decision by hashing
the TraceID, but we feel such an approach is not sufficient for making
unbiased sampling decisions.  It is seen as a relatively difficult
task to define and specify a good enough hashing function, much less
to have it implemented in multiple languages.

Hashing is also computationally expensive. This proposal uses extra
data to avoid the computational cost of hashing TraceIDs.

### Restriction to power-of-two 

Restricting head sampling rates to powers of two does not limit tail
Samplers from using arbitrary probabilities.  The companion [OTEP
170](https://github.com/open-telemetry/oteps/pull/170) has discussed
the use of a `sampler.adjusted_count` attribute that would not be
limited to power-of-two values.  Discussion about how to represent the
effective adjusted count for tail-sampled Spans belongs in [OTEP
170](https://github.com/open-telemetry/oteps/pull/170), not this OTEP.

Restricting head sampling rates to powers of two does not limit
Samplers from using arbitrary effective probabilities over a period of
time.  For example, choosing 1/2 sampling half of the time and 1/4
sampling half of the time leads to an effective sampling rate of 3/8.

## Prior art and alternatives

Google's Dapper system propagated a field in its trace context called
"inverse_probability", which is equivalent to adjusted count.  This
proposal uses the base-2 logarithm of adjusted count to save space and
limit required randomness.
