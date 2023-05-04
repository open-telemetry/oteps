# Non-power-of-two Probability Sampling using 56 random TraceID bits

## Motivation

**Status*: CURRENT

The existing, experimental [specification for probability sampling using TraceState](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/tracestate-probability-sampling.md)
supporting Span-to-Metrics pipelines is limited to powers-of-two
probabilities and is designed to work without making assumptions about
TraceID randomness.

This proposes to extend that specification with support for 56-bit
precision sampling probability.  This is seen as particularly
important for implementation of probabilistic tail samplers (e.g., in
the OpenTelemetry Collector) as explained below.

This OTEP makes use of the [draft-standard W3C tracecontext `random`
flag](https://w3c.github.io/trace-context/#random-trace-id-flag),
which is an indicator that 7 bytes of true randomness are available
for probability sampler decisions.

## Explanation

**Status*: CURRENT

The existing, experimental TraceState probability sampling
specification relies on two variables known as **r-value** and
**p-value**.  The r-value carries the source of randomness and the
p-value carries the effective sampling probability.

Given this specification, a ConsistentProbabilitySampler can be
applied as a head sampler for non-power-of-two sampling probabilities
using interpolation.  For example, an effective sampling probability
of 1-in-3 can be achieved by alternating between 25% and 50% sampling.
However, interpolation only works for trace roots, otherwise
"consistent" sampling can only be achieved at the next smaller power
of two.  In the example, sampling at 1-in-3 using interpolation means
traces are only guaranteed **consistent** at 25% and smaller sampling
probabilities.

The major downside of the r-value, p-value approach is that r-value
must be encoded even for unsampled contexts.  Ideally, building
Span-to-Metrics pipelines should be low overhead which means not
adding additional data to unsampled contexts.

This proposal avoids r-value by using 7 bytes of intrinsic randomness
in the TraceID, the ones (draft-) specified [in the W3C tracecontext
`random` flag](https://w3c.github.io/trace-context/#random-trace-id-flag).
Since this Sampler is expected to behave consistently with or without
the `random` flag, we assume the bits are random and do not actually
check the W3C random flag.

This document propose extending the existing p-value, r-value
mechanism with support for a new indicator for non-power-of-two
probability sampling known as "t-value", where "t" is chosen because
it signifies a threshold.  If widely adopted, the tracestate r-value
can be deprecated, as it is not needed when randomness is provided in
the TraceID.

As proposed, t-value and p-value are mutually exclusive; p-value
remains the preferred encoding for probability sampling when a
power-of-two sampling probability is used.  P-value also remains the
specified way to encode zero adjusted count (i.e., p=63).  T-value MAY
be used to encode power-of-two probabilities, although typically the
equivalent p-value uses fewer bytes.

### T-Value encoding: Requirements

**Status*: NEW-DRAFT

#### Exactness

This proposal is required to be support precision Span-to-Metrics
pipelines.  This means that effective sampling probabilities are
limited to discrete values that can be exactly represented. The number
of discrete steps between powers of two is limited by the number of
remaining bits of randomness in the TraceID.

To achieve exactly 1-in-2^56 sampling, a sampler can select all traces
with 56 `0`s of TraceID randomness.  It is not possible to achieve a 
smaller sampling probability than 1-in-2^56.

The next larger, exactly representable sampling probability is
1-in-2^55.  At this probability, a sampler can select all traces with
55 leading `0`s of TraceID randomness (i.e., 55 `0`s followed by a `1`
and 55 `0`s followed by a `0`).  There are no exact probabilities
representable between 1-in-2^55 and 1-in-2^56.

The next larger, exactly representable power-of-two sampling
probability is 1-in-2^54.  At this probability, a sampler can select
all traces with 54 leading `0`s of TraceID randomness.  At 4 out of
2^56, this sampling probability includes the two TraceID-randomess
values selected at smaller powers-of-two (i.e., 1-in-2^55 and
1-in-2^56) plus two new TraceID-randomness values.  One of the two new
TraceID-randomness values corresponds with exactly 1-in-2^54 sampling,
the other of these is the smallest exactly-representable
non-power-of-two sampling probability according to this scheme.  It
lies halfway between 1-in-2^54 and 1-in-2^55; in binary floating point
representation, this value is displayed as `0x1.8p-55`.

Continuing this pattern, the next larger power-of-two sampling
probability is 1-in-2^53, which is 8 out of 2^56, 4 of which were
covered above and 4 of which are new.  Of the four new, 1 is the exact
power-of-two and there are three available non-power-of-two
probabilities in this range.  These probabilities are (exactly)
`0x1.Cp-54`, `0x1.8p-54`, and `0x1.4p-54`.

In the pattern developed here, the number of sampling probabilities in
the open interval `(2^-N, 2^-(N+1))` equals `(2^(56-N))-1`.

Note we are disregarding the fact that a TraceID with all zeros (i.e.,
128 `0` bits) is specified invalid by OpenTelemetry, which makes the
all-zeros TraceID-randomness value slightly less probable than other
values.

#### Correspondence with R-value

**Status*: NEW-DRAFT

There are reasons to maintain compatibility with r-values in the range
[0, 56] as developed in the earlier specification, particularly
because it enables intentionally-consistent sampling across multiple
traces.  We require that when r-value is used, r-value takes
precendece over builtin TraceID-randomness.

In this specification, the use of r-values greater than 56 is deprecated.

We require the correspondence with non-power-of-two sampling
probabilities exact to be exact.  This can be achieved as follows by
calculating an *effective TraceID-randomness value* from the r-value
combined with the original randomness.

When r-value is set to the value `x` (where `x < 56`), the effective
TraceID-randomness value used is calculated as `x` leading `0`s,
followed by a `1`, followed by the original `56-x-1` trailing bits of
TraceID-randomness.

R-value propgation rules are unmodified.  R-value consistency-checking
rules will be updated to detect inconsistent t-values, similar to the
current specification's rules for detecting inconsistent p-values..

#### Sampling decision logic

**Status*: NEW-DRAFT

An implementation of a head or tail sampler is expected to perform a
simple comparison between the 56 bits of TraceID-randomness value and
a threshold value.  The encoded t-value will correspond with one of
the exactly representable values of TraceID-randomness, such that a
simple less-than-or-equal comparison achieves exactly the correct
sampling probability.

#### Consistency between head and tail sampling

**Status*: NEW-DRAFT

The correspondence with r-value is meant to ensure that head samplers
and tail samplers will make a consistent decision at non-power-of-two
sampling probabilities.  Whereas the existing specification states
that head samplers should use random interpolation between
powers-of-two, the updated consistent sampling specification will use
the deterministic algorithm for head and tail developed above.

#### Deterministic mapping to integer adjusted counts

**Status*: NEW-DRAFT

One requirement remains to be developed.  A nice-to-have feature
developed in the earlier specification is that when interpolating
between power-of-two sampling probabilities, the final p-value would
nevertheless be output with one of the nearby power-of-two adjusted
counts.

Using the smallest representable non-power-of-two sampling probability
`0x1.8p-55` as an example--this value lies exactly half-way between
two powers-of-two so we require a deterministic, unbiased way to
select `0x1p-54` 1-out-of-3 times and `0x1p-55` 2-out-of-3 times.

Can we use the SpanID bits to make this selection consistently at the
consumer for each Span?  This would allow an exactly-encoded
non-power-of-two `t-value` to nevertheless be mapped into integer
(power-of-two) adjusted counts.

TODO: This is an ongoing investigation.

#### Summary of sampling algorithm

**Status*: NEW-DRAFT

The steps to perform a sampling decision are the same for both head
and tail samplers.

First, select an exactly representable sampling probability.  If the
input is an arbitrary floating point value, it will have to be rounded
to a nearby exact probablity.  Then, the probability is converted in
two ways: 

1. The t-value is calculated that encodes the exact effective samping
   probability.
2. The 56-bit threshold for comparing against TraceID-randomness is
   calculated as described above.

For each span, the sampler extracts 56 bits of presumed randomness
from the TraceID, the so-called TraceID-randomness value.

When r-value is set to `x` in the span's context, the sampler modifies
the leading `x+1` bits of TraceID-randomness value with `x` `0`s and
followed by a `1`.

A simple comparison is made between the threshold and the effective
TraceID-randomness value.  If the effective TraceID-randomess value is
less than or equal to the threshold, the span is selected with the
calculated t-value.  Otherwise, the span is not selected.

### T-Value encoding: Original draft

**Status*: OUT-OF-DATE

Since we have 7 bytes, or 56 bits of randomness available, there are
2^56 non-zero sampling probabilities that can be encoded.  These
probabilities can be expressed as a 56-bit number in the range [0,
0xffffffffffffff], where 0 corresponds with sampling 1 span out of
2^56 and 0xffffffffffffff corresponds with sampling 100% of spans.

The proposal is summarized as follows.  T-value is encoded as a
hexadecimal string containing between 1 and 14 hex digits.  When the
T-value is less than 14 hex digits, it is extended to 14 bytes using
by padding with 0s.  For example, the t-value string "003f"
corresponds with a the 14-hex-digit string "003f0000000000".

Head samplers and tail samplers alike can be implemented simply by
tersting whether the least-significant 7 bytes of the TraceID are
lexicgraphically less-than-or-equal to the sampling threshold.  Note
that this comparison may be carried out directly on hex digits or on
binary data using simple string or bytes comparison.

Modifying an in-SDK Sampler to perform this calculation is a simple
change relative to setting p-value for sampled spans.  For tail
samplers, a span processor can simply pass through all spans where the
least-significant 7-bytes of TraceID are less-than-or-equal to the
configured threshold.  When the span passes, it has its TraceState
t-value set to the configured threshold for use in Span-to-Metrics
pipelines.

### Converting between Thresholds and Probabilities

**Status*: OUT-OF-DATE

Sampling probabilities in the range (0, 1] can be mapped onto 56-bit
encoded t-values in the range [0, 0xffffffffffffff].  For a given
sampling threshold, the corresponding probability is expressed as a
fraction `(T+1)/2^56` (i.e., sampling threshold plus one divided by
2^56).

Note that IEEE double-width floating point numbers use 52 bits of
significand, so not all sampling thresholds have corresponding
floating point values that the user might be able to express.

For SDKs and Span processors to implement consistent probability
sampling, OpenTelemetry should define how to compute a sampling
threshold from a floating point number and in the reverse direction,
how to compute a floating point number from a threshold.  Combined,
these rules allow simple sampling logic to be easily translated into
probabilities or adjusted counts for use in a Span-to-Metrics
pipeline.

#### Probability to Hex Threshold

**Status*: OUT-OF-DATE

Note that the procedure here only works for probabilities greater than
or equal to 2^-52.

To convert from a floating point number to the nearest threshold as a
14-byte hex string:

```
func ProbabilityToThreshold(prob float64) string {
	return fmt.Sprintf("%.14x", math.Nextafter(prob+1, 0))[4:18]
}
```

Note that this can be truncated after one or more non-zero digits,
leaving a more-compact encoding of a sampling probability that is
nearby.

Note that the threshold is rounded down, it will be slightly smaller
than the configured probabilty in cases where the probability cannot
be exactly represented in 56 bits.

#### Hex Threshold to Probability

**Status*: OUT-OF-DATE

To convert a hex threshold string to the corresponding probability, we
perform that opposite of the above.

```
func ThresholdToProbability(thresh string) float64 {
    parsed, _ := strconv.ParseFloat("0x1."+thresh[:13]+"p+00", 64)
	return math.Nextafter(parsed, 2) - 1
}
```

Note that these transformations are not always reversible, since
floating point numbers have less precision.  Note that only 13 bytes
of the hex string are used to form the floating point value, since
that is all the precision a double-wide floating point number has.

## Examples

**Status*: OUT-OF-DATE

### 90% sampling 

The following header

```
tracestate: ot=t:e66
```

contains a sampling threshold "e66", which is extended to
"e6600000000000".  The corresponding TraceID's least-significant 7
bytes are expected to be less than or equal to "e6600000000000".

The corresponding sampling probability, calculated using the equation
above, is 0.9.  The adjusted count of this span in a Span-to-Metrics
pipeline is 1.11.

### 0.33333% sampling

The following header

```
tracestate: ot=t:00da7
```

corresponds with 0.33333% sampling.

## Trade-offs and mitigations

**Status*: OUT-OF-DATE

Note that the t-value encoding is not efficient for encoding
power-of-two probabilities (e.g., "ffffffffffffff" corresponds with
100% sampling).  That is why the use of p-value is recommended when
the configured sampling probability is an exact power-of-two.

## Prior art and alternatives

**Status*: OUT-OF-DATE

An earlier draft of proposal was explored [here](https://github.com/jmacd/opentelemetry-collector-contrib/pull/2925).
