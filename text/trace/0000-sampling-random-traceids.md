# Non-power-of-two Probability Sampling using 56 random TraceID bits

## Motivation

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

The existing, experimental TraceState probability sampling
specification relies on two variables known as **r-value** and
**p-value**.  The r-value carries the source of randomness and the
p-value carries the effective sampling probability.

Given this specification, a ConsistentProbabilitySampler can be
applied as a head sampler for non-power-of-two sampling probabilities
using interpolation.  For example, a neffective sampling probability
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
the `random` flag, we assumes the bits are random and do not actually
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
specified way to encode zero adjusted count (i.e., p=63).

### T-Value encoding

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

Note that the t-value encoding is not efficient for encoding
power-of-two probabilities (e.g., "ffffffffffffff" corresponds with
100% sampling).  That is why the use of p-value is recommended when
the configured sampling probability is an exact power-of-two.

## Prior art and alternatives

An earlier draft of proposal was explored [here](https://github.com/jmacd/opentelemetry-collector-contrib/pull/2925).
