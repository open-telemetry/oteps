# Propagate head trace sampling probability

Propose extending the W3C trace context `traceparent` to convey head trace sampling probability.

## Motivation

The head trace probability is useful in child contexts to be able to
record the effective sampling probability in child spans.  This is
documented in [OTEP 148](TODO: after merging), which establishes
semantic conventions for conveying the adjusted count of a span via
attributes recorded with the span.  When a sampling decision is based
on the parent's context, the effective sampling probability, which
determines the child's adjusted count, cannot be recorded without
propagating it through the context.

We propose to propagate the trace sampling probability that is in
effect whenever the [W3C
sampled](https://www.w3.org/TR/trace-context/#sampled-flag) flag is
set by extending the `traceparent`.

## Explanation

To limit the cost of this extension, to ensure that it is widely
supported, and for statistical reasons documented below, we propose to
limit head tracing probability to powers of two.  This limits the
available head sampling probabilities to 1/2, 1/4, 1/8, and so on, and
we can compactly encode these probabilities as small integers using
the negative base-2 logarithm of the effective probability.

For example, the value 2 corresponds with 1-in-4 sampling, the value
10 corresponds with 1-in-1024 sampling.

Wheres the [version-0 W3C trace context `traceparent`
header](https://www.w3.org/TR/trace-context/#examples-of-http-traceparent-headers)
is a concatenation of four fields, this proposal would upgrade
`traceparent` to version 1:

```
traceparent: (version)-(trace_id)-(span_id)-(flags)
```

The version 1 `traceparent` header will use a new field named `log-count`, i.e.,:

```
traceparent: (version)-(trace_id)-(span_id)-(flags)-(log-count)
```

where `log-count` is the encoded negative base-2 logarithm of
sampling probability, which is the base-2 logarithm of the adjusted
count of a child span created in this context (i.e., the logarithm of
the effective count, thus "log-count").  To compute the adjusted count
of a child span created in this context, use `2^log-count`.  A
log-count of `0` corresponds with `(2^0)=1`, thus 0 conveys a context
with probability 1.

The sampling probability of a context is independent from whether it
is sampled.  We consider it [useful to convey sampling probability
even when unsampled]() as it can be used to estimate the potential
overhead of starting new sampled traces.

### Examples

These are extended [from the W3C
examples](https://www.w3.org/TR/trace-context/#examples-of-http-traceparent-headers):

```
Value = 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01-05
base16(version) = 00
base16(trace-id) = 4bf92f3577b34da6a3ce929d0e0e4736
base16(parent-id) = 00f067aa0ba902b7
base16(trace-flags) = 01  // sampled
base16(log-count) = 05  // head probability is 2^-5.
```

```
Value = 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00-00
base16(version) = 00
base16(trace-id) = 4bf92f3577b34da6a3ce929d0e0e4736
base16(parent-id) = 00f067aa0ba902b7
base16(trace-flags) = 00  // not sampled
base16(log-count) = 10 // head probability is 2^-16
```

We are able to express sampling probabilities as small as 2^-255 using
just 3 bytes per `traceparent`.

## Internal details

A use known as "inflationary sampling" from Google's Dapper system is
documented in [OTEP 148](TODO: inflationary sampling section).  This
is is used to justify propagating the head sampling probability even
when unsampled.

[An algorithm for making statistical inferance from partially-sampled
traces has been published](https://arxiv.org/pdf/2107.07703.pdf) that
explains how to work with power-of-2 sampling rates.  The reasoning
behind restricting the set of sampling rates is:

- Lowers the cost of propagating head sampling probability
- Makes math involving partial traces tractable

## Trade-offs and mitigations

Restricting head sampling rates to powers of two does not limit tail
Samplers from using arbitrary probabilities.

Restricting head sampling rates to powers of two does not limit
Samplers from using arbitrary effective probabilities over a period of
time.  For example, choosing 1/2 sampling half of the time and 1/4
sampling half of the time leads to an effective sampling rate of 3/8.

## Prior art and alternatives

Google's Dapper system propagated a field in its trace context called
"inverse_probability", which is equivalent to adjusted count.  This
proposal uses the base-2 logarithm of adjusted count to save space

## Open questions

This OTEP suggests how to modify the W3C trace context to accomodate
sampling in OpenTelemetry.  [OTEP 148](TODO) suggests semantic
conventions for encoding adjusted count in a Span, but neither text
specifies how to modify the built-in Samplers to produce the proposed
new `traceparent` field so that the `ParentBased` Sampler can
correctly set the proposed `sampler.adjusted_count` attribute.  This
will be future work.
