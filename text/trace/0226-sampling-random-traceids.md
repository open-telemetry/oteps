# Non-power-of-two Probability Sampling using 56 random TraceID bits

## Motivation

The existing, experimental [specification for probability sampling using TraceState](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/tracestate-probability-sampling.md)
supporting Span-to-Metrics pipelines is limited to powers-of-two
probabilities and is designed to work without making assumptions about 
TraceID randomness.

Head sampling requires the use of TraceState to propagate context from
the parent for recording in child spans, in support of Span-to-Metrics
pipelines.  Tail sampling does not require context propagation
support, but it has many similar requirements:

1. Sampling should be "consistent", so that independent collection
   paths make identical sampling decisions.
2. Spans should be countable in a Span-to-Metrics pipeline, which
   requires knowing the "adjusted count" for each span directly from
   the data.

This OTEP makes use of the [draft-standard W3C tracecontext `random`
flag](https://w3c.github.io/trace-context/#random-trace-id-flag),
which is an indicator that 7 bytes of true randomness are available
for probability sampler decisions.

This proposes to create a specification with support for 56-bit
precision tail sampling.  This is seen as particularly important for
implementation of probabilistic tail samplers (e.g., in the
OpenTelemetry Collector) as explained below.

## Explanation

The existing, experimental TraceState probability sampling
specification relies on two variables known as **r-value** and
**p-value**.  The r-value carries the source of randomness and the
p-value carries the effective sampling probability.  The preceding
specification recommends the use of interpolation to achieve
non-power-of-two sampling probabilities.

This specification is proposed that aims to offer an alternative to
that r-value, p-value specification, one that is simpler to implement,
can be used in both head- and tail-samplers, and that naturally
supports non-power-of-two sampling probabilities.

This proposal uses the 7 bytes of intrinsic randomness in the TraceID,
the ones (draft-) specified [in the W3C tracecontext `random`
flag](https://w3c.github.io/trace-context/#random-trace-id-flag). With
these bits, a simple threshold test is defined to allow sampling based
on TraceID randomness.

This document proposes extending the p-value, r-value mechanism with
support for a new indicator for non-power-of-two probability sampling
known as "t-value", where "t" is chosen because it signifies a
threshold.  Tail-based sampling encoded by t-value can be combined
with p-value, in which case the adjusted count implied by t-value is
**multiplied** with the adjusted count implied by p-value because they
are independent mechanisms.

### Detailed design

Support for Span-to-Metrics pipelines requires knowing the "adjusted
count" of every collected span.  This proposal defines the sampling
"threshold" as a 7-byte string used to make consistent sampling
decisions, as follows.

1. Bytes 9-16 of the TraceID are interpreted as a 7-byte unsigned
   value in big-endian byte order.
2. If the unsigned value determined by the trace is less-than
   to the sampling threshold, the span is sampled, otherwise it is
   discarded.
   
To calculate the Sampling threshold, we begin with an IEEE-754
standard double-precision floating point number.  With 52-bits of
significand and a floating exponent, the probability value used to
calculate a threshold may be capable of representing more-or-less
precision than the sampler can execute.

We have many ways of encoding a floating point number as a string,
some of which result in loss of precision.  This specification dicates
exactly how to calculate a sampling threshold from a floating point
number, and it is the sampling threshold that determines exactly the
effective sampling probability.  The conversion between sampling
probability and threshold is not exactly reversible, so to determine
the sampling probability exactly from an encoded t-value, first
compute the exact sampling threshold, then use the threshold to derive
the exact sampling probability.

From the exact sampling probability, we are able to compute (subject
to machine precision) the adjusted count of each span.  For example,
given a sampling probability encoded as "0.1", we first compute the
nearest base-2 floating point, which is exactly 0x1.999999999999ap-04,
which is approximately 0.10000000000000000555.  The exact quantity in
this example, 0x1.999999999999ap-04, is multipled by `2^56` and
rounded to an unsigned integer (7205759403792794).  This specification
says that to carry out sampling probability "0.1", we should keep
exactly 7205759403792794 smallest unsigned values of the 56-bit random
TraceID bits.

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

Although prepared as a solution for tail sampling, the t-value
encoding scheme could also be used to convey Logs sampling.  While
tail sampling does not require the use of trace state, which is
associated with context propagation, it makes a natural place to store
t-value because it should be interpreted along with p-value, which
resides in the trace state.  However, if spans store t-value in trace
state, it is not clear how to convey logs sampling.

Here are ways to address this:

1. Store t-value in a new dedicated field in the Span or Log Record
   (as a string).  (Author's preference.)
2. Store t-value as a Span or Log Record attribute (as a string).
   This may cause confusion because the attribute, which was not
   applied by a user, can change long the collection path even though
   the data has not changed.
3. Store t-value as an optional floating point field in the Span or
   Log Record.  An optional field is required because we need a
   meaningful way to represent zero probability, for cases where spans
   are exporter due to a non-probabilistic decision.
4. Create a new field in both Spans and Log Records as a dedicated
   field for storing t-values.
   
The benefit of using TraceState is that it is an extensible field,
made for multiple vendors to place arbitrary contents.  It is not
clear whether use of tracestate to record collection-time decisions is
appropriate, or whether it is only meant for in-band context
propagation.  If this use-case is acceptable, the name Trace State
would become a legacy; in this case, a more signal-neutral name for
the field could be developed (e.g., "Collection State")

### 90% sampling 

The following header

```
tracestate: ot=t:0.9
```

### 1-in-3 sampling

The following header

```
tracestate: ot=t:3
```

corresponds with 1-in-3 sampling.

### 25% head sampling, 1-in-10 tail sampling

The following header

```
tracestate: ot=p:2;t:10
```

corresponds with 1-in-4 sampling at the head and 1-in-10 tail
sampling.  The resulting span has adjusted count 40.

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

An earlier draft of proposal was explored [here](https://github.com/jmacd/opentelemetry-collector-contrib/pull/2925).
