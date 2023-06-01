# Non-power-of-two Probability Sampling using 56 random TraceID bits

## Motivation

The existing, experimental [specification for probability sampling
using
TraceState](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/tracestate-probability-sampling.md)
supporting Span-to-Metrics pipelines is limited to powers-of-two
probabilities and is designed to work without making assumptions about
TraceID randomness. The existing mechanism could only achieve
non-power-of-two sampling using interpolation between powers of two,
which was only possible at the head sampling time.  It could not be
used with non-power-of-two sampling probabilities for span sampling in
the rest of the collection path. This proposal aims to address the
above two limitations for a couple of reasons:

1. Certain customers want support for non-powers-of-two probabilities
   (e.g., 10% sampling rate or 75% sampling rate) and it should be
   possible to do it cleanly irrespective of where the sampling is
   happening.
2. There is a need for consistent sampling in the collection path
   (outside of the head-sampling paths) and using the inherent
   randomness in the traceID is a less-expensive solution than
   referencing a custom "r-value" from the tracestate in every span.

In this proposal, we will cover how this new mechanism can be used in
both head-based sampling and different forms of tail-based sampling.

The term "Tail sampling" is in common use to describe _various_ forms
of sampling that take place after a span starts.  The term "Tail" in
this phrase distinguishes other techniques from head sampling, however
the term is only broadly descriptive.

Head sampling requires the use of TraceState to propagate context
about sampling decisions from parent spans to child spans.  With sampling
information included in the TraceState, spans can be labeled with their
effective adjusted count, making it possible to count spans as they
arrive at their destination in real time, meaning before assembling
complete traces.

Here, the term Intermediate Span Sampling is used to describe sampling
performed on individual spans at any point in their collection path.
Like Head sampling, Intermediate Span Sampling benefits from being
consistent, because it makes recovery of complete traces possible
after spans have independently sampled.  On the other hand, when "Tail
sampling" refers to sampling of complete traces, sampling consistency
is not an important property.

Intermediate Span Sampling is exemplified by the
[OpenTelemetry-Collector-Contrib's `probabilisticsampler`
processor](https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor).
This proposal is motivated by wanting to compute Span-to-Metrics from
spans that have been sampled by such a processor.

This proposal makes use of the [draft-standard W3C tracecontext
`random`
flag](https://w3c.github.io/trace-context/#random-trace-id-flag),
which is an indicator that 56 bits of true randomness are available
for probability sampler decisions.  The benefit of this is that this
inherently random value can be used by intermediate span samplers to
make _consistent_ sampling decisions. It would be a less-expensive
solution than the earlier proposal of looking up the r-value from the
tracestate of each span.

This proposes to create a specification with support for 56-bit
precision consistent Head and Intermediate Span sampling.  Because
this proposal is also meant for use with Head sampling, a new member
of the OpenTelemetry TraceState field will be defined.  Intermediate
Span Samplers will modify the TraceState field of spans they sample.

Note also there is an interest in probabilistic sampling of
OpenTelemetry Log Records on the collection path too.  This proposal
recommends the creation of a new field in the OpenTelemetry Log Record
with equivalent use and interpretation as the (W3C trace-context)
TraceState field.  It would be appropriate to name this field
`LogState`.

This proposal does makes r-value an optional 56-bit number as opposed
to a required 6-bit number.  When the r-value is supplied, it acts as
an alternative source of randomness which allows tail-samplers to
support versions of tracecontext without the `random` bit as well as
more advanced use-cases.  For example, independent traces can be
consistently sampled by starting them with identical r-values.

This proposal deprecates the experimental p-value.  For existing
stored data, the specification may recommend replacing `p:X` with an
equivalent t-value; for example, `p:2` can be replaced by `t:4` and
`p:20` can be replaced by `t:0x1p-20`.

## Explanation

This document proposes a new OpenTelemetry specific tracestate value
called t-value.  This t-value encodes either the sampling probability
(a floating point value) directly or the "adjusted count" of a span
(an integer).  The letter "t" here is a shorthand for "threshold". The
value encoded here can be mapped to a threshold value that a sampler
can compare to a value formed using the rightmost 7 bytes of the
traceID.

The syntax of the r-value changes in this proposal, as it contains 56
bits of information.  The recommended syntax is to use 14 hexadecimal
characters (e.g., `r:1a2b3c4d5e6f78`).  The specification will
recommend samplers drop invalid r-values, so that existing
implementations of r-value are not mistakenly sampled.

Like the existing specification, r-values will be synthesized as
necessary.  However, the specification will recommend that r-values
not be synthesized automatically when the W3C tracecontext `random`
flag is set.  To achieve the advanced use-case involving multiple
traces with the same r-value, users should set the `r-value` in the
tracestate before starting correlated trace root spans.

### Detailed design

Let's look at the details of how this threshold can be calculated.
This proposal defines the sampling "threshold" as a 7-byte string used
to make consistent sampling decisions, as follows.

1. When the r-value is present and parses as a 56-bit random value,
   use it, otherwise bytes 10-16 of the TraceID are interpreted as a
   56-bit random value in big-endian byte order
2. The sampling probability (range `[0x1p-56, 1]`) is multiplied by
   `0x1p+56`, yielding a unsigned Threshold value in the range `[1,
   0x1p+56]`.
3. If the unsigned TraceID random value (range `[0, 0x1p+56)`) is
   less-than the sampling Threshold, the span is sampled, otherwise it
   is discarded.

For head samplers, there is an opportunity to synthesize a new r-value
when the tracecontext does not set the `random` bit (as the existing
specification recommends synthesizing r-values for head samplers
whenever there is none).  However, this opportunity is not available
to tail samplers.

To calculate the Sampling threshold, we began with an IEEE-754
standard double-precision floating point number.  With 52-bits of
significand and a floating exponent, the probability value used to
calculate a threshold may be capable of representing more-or-less
precision than the sampler can execute.

We have many ways of encoding a floating point number as a string,
some of which result in loss of precision.  This specification dicates
exactly how to calculate a sampling threshold from a floating point
number, and it is the sampling threshold that determines exactly the
effective sampling probability.  The conversion between sampling
probability and threshold is not always reversible, so to determine
the sampling probability exactly from an encoded t-value, first
compute the exact sampling threshold, then use the threshold to derive
the exact sampling probability.

From the exact sampling probability, we are able to compute (subject
to machine precision) the adjusted count of each span.  For example,
given a sampling probability encoded as "0.1", we first compute the
nearest base-2 floating point, which is exactly 0x1.999999999999ap-04,
which is approximately 0.10000000000000000555.  The exact quantity in
this example, 0x1.999999999999ap-04, is multiplied by `0x1p+56` and
rounded to an unsigned integer (7205759403792794).  This specification
says that to carry out sampling probability "0.1", we should keep
Traces whose least-significant 56 bits form an unsigned value less
than 7205759403792794.

## T-value encoding for adjusted counts

The example used sampling probability "0.1", which is a concisely
rounded value but not exactly a power of two.  The use of decimal
floating point in this case conceals the fact that there is an integer
reciprocal, and when there is an integer reciprocal there are good
reasons to preserve it.  Rather than encoding "0.1", it is appealing
to encode the adjusted count (i.e., "10") because it conveys exactly
the user's intention.

This suggests that the t-value encoding be designed to accept either
the sampling probability or the adjusted count, depending on how the
sampling probability was derived.  Thus, the proposed t-value shall be
parsed as a floating point or integer number using any POSIX-supported
printf format specifier.  Values in the range [0x1p-56, 0x1p+56] are
valid.  Values in the range [0x1p-56, 1] are interpreted as a sampling
probability, while values in the range [1, 0x1p+56] are intepreted as
an adjusted count.  Adjusted count values must be integers, while
sampling probability values can be arbitrary floating point values.

Whether to encode sampling probabilty or adjusted count is a choice.
In both cases, the interpreted value translates into an exact
threshold, which determines the exact inclusion probability.  From the
exact inclusion probability, we can determine the adjusted count to
use in a span-to-metrics pipeline.  When the t-value is _stated_ as an
adjusted count (as opposed to a sampling probabilty), implementations
can use the integer value in a span-to-metrics pipeline.  Otherwise,
implementations should use an adjusted count of 1 divided by the
sampling probability.

## Where to store t-value in a Span and/or Log Record

As specified, t-value should be encoded in the TraceState field of the
span.  Probabilistic head samplers should set t-value in propagated
contexts so that children using ParentBased samplers are correctly
counted.

Although prepared as a solution for Head and Intermediate Span
sampling, the t-value encoding scheme could also be used to convey
Logs sampling.  This document proposes to add an optional `LogState`
string to the OTLP LogRecord, defined identically to the W3C
tracecontext `TraceState` field.

## Re-sampling with t-value

It is possible to re-sample spans that have already been sampled,
according to their t-value.  This allows a processor to further reduce
the volume of data it is sending by lowering the sampling threshold.

In such a sampler, the incoming span will be inspected for an existing
t-value.  If found, the incoming t-value is converted to a sampling
threshold and compared against the new threshold.  These are two cases:

- If the Threshold calculated from the incoming t-value is less than
  or equal to the current sampler's Threshold, the outgoing t-value is
  copied from the incoming t-value.  In this case, the span had
  already been sampled with a less-than-or-equal probability compared
  with the current sampler, so for consistency the span simply passes
  through.
- If the Threshold calculated from the incoming t-value is larger than
  the current sampler's Threshold, the current sampler's Threshold is
  re-applied; if the TraceID random value is less than the current
  Sampler's threshold, the span passes through with the current
  sampler's t-value, otherwise the span is discarded.

## S-value encoding for non-consistent adjusted counts

There are cases where sampling does not need to be consistent or is
intentionally not consistent.  Existing samplers often apply a simple
probability test, for example.  This specification recommends
introducing a new tracestate member `s-value` for conveying the
accumulation of adjusted count due to independent sampling stages.

Unlike resampling with `t-value`, independent non-consistent samplers
will multiply the effect of their sampling into `s-value`.

## Examples

### 90% consistent intermediate span sampling

A span that has been sampled at 90% by an intermediate processor will
have `ot=t:0.9` added to its TraceState field in the Span record.  The
sampling threshold is `0.9 * 0x1p+56`.

### 90% head consistent sampling

A span that has been sampled at 90% by a head sampler will add
`ot=t:0.9` to the TraceState context propagated to its children and
record the same in its Span record.  The sampling threshold is `0.9 *
0x1p+56`.

### 1-in-3 consistent sampling

The tracestate value `ot=t:3` corresponds with 1-in-3 sampling.  The
sampling threshold is `1/3 * 0x1p+56`.

### 30% simple probability sampling

The tracestate value `ot=s:0.3` corresponds with 30% sampling by one
or more sampling stages.  This would be the tracestate recorded by
`probabilisticsampler` when using a `HashSeed` configuration instead
of the consistent approach.

### 10% probability sampling twice

The tracestate value `ot=s:0.01` corresponds with 10% sampling by one
stage and then 10% sampling by a second stage.

## Trade-offs and mitigations

Support for encoding t-value as either a probability or an adjusted
count is meant to give the user control over loss of precision.  At
the same time, it can be read by humans.

Floating point numbers can be encoded exactly to avoid ambiguity, for
example, using hexadecimal floating point representation.  Likewise,
adjusted counts can be encoded exactly as integers to convey the
user's intended sampling probability without floating point conversion
loss.

## Prior art and alternatives

The existing p-value, r-value mechanism could only achieve
non-power-of-two sampling using interpolation between powers of two,
which was only possible at the head.  That specification could not be
used for Intermediate Span sampling using non-power-of-two sampling
probabilities.

There is a case to be made that users who apply simple probability
sampling with hard-coded probabilities are not asking for what they
want, which is to apply a rate-limit in their sampler.  It is true
that rate-limited sampling can be achieved confined to power-of-two
sampling probabilities, but we feel this does not diminish the case
for simply supporting non-power-of-two probabilities.
