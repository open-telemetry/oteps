# Propagate head trace sampling probability

Propose extending the W3C trace context `traceparent` to convey head trace sampling probability.

## Motivation

The head trace probability is useful in child contexts to be able to
record the effective sampling probability in child spans.  This is
documented in [OTEP 170](TODO), which establishes semantic conventions
for conveying the adjusted count of a span via span attributes.  When
a sampling decision is based on the parent's context, the effective
sampling probability, which determines the child's adjusted count,
cannot be recorded without propagating it through the context.

We propose to propagate the trace sampling probability that is in
effect alongside the [W3C
sampled](https://www.w3.org/TR/trace-context/#sampled-flag) flag
either by extending the `traceparent` or through the use of
`tracestate` with an `otel` vendor tag.

## Explanation

Two variations of this proposal are presented.  The first, based on
`traceparent`, is the more-ideal choice because it ensures broad
support and reduces the number of bytes per request.  The second,
based on a `tracestate` key=value, is less appealing as `tracestate`
has the appearance of a vendor-specific field, when it is not.

In both cases, to limit the cost of this extension and for statistical
reasons documented below, we propose to limit head tracing probability
to powers of two.  This limits the available head sampling
probabilities to 1/2, 1/4, 1/8, and so on.  We can compactly encode
these probabilities as small integers using the base-2 logarithm of
the adjusted count.

For example, the value 2 corresponds with 1-in-4 sampling, the value
10 corresponds with 1-in-1024 sampling.

### Proposal using `traceparent`

Wheres the [version-0 W3C trace context `traceparent`
header](https://www.w3.org/TR/trace-context/#examples-of-http-traceparent-headers)
is a concatenation of four fields,

```
traceparent: (version)-(trace_id)-(span_id)-(flags)
```

This proposal would upgrade `traceparent` to version 1 with a new
field named `log_count`,

```
traceparent: (version)-(trace_id)-(span_id)-(flags)-(log_count)
```

where `log_count` is the base-2 logarithm of the adjusted count of a
child span created in this context (i.e., the logarithm of the
effective count, thus "log_count").  To compute the adjusted count of
a child span created in this context, use `2^log_count`.  A log_count
of `0` corresponds with `(2^0)=1`, thus 0 conveys a context with
probability 1.

The sampling probability of a context is independent from whether it
is sampled.  We consider it useful to convey sampling probability even
when unsampled, as shown by Dapper's "inflationary" sampler.  Note,
however, that an unsampled trace with probability 1-in-1 is illogical.
To prevent illogical interpretation and to avoid errors introduced by
downgrading `traceparent` to the version 0 format, a new flag
`probabilistic` flag is introduced to indicate when the `log_count`
field is meaningful.

This flag would use the 2nd available bit in the W3C trace flags
field (i.e., 0x2).

#### Examples using `traceparent`

These are extended [from the W3C
examples](https://www.w3.org/TR/trace-context/#examples-of-http-traceparent-headers):

```
Traceparent = 01-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-03-05
base16(version) = 01
base16(trace_id) = 4bf92f3577b34da6a3ce929d0e0e4736
base16(parent_id) = 00f067aa0ba902b7
base16(trace_flags) = 03  // sampled, probabilistic
base16(log_count) = 05  // head probability is 2^-5.
```

```
Traceparent = 01-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-02-11
base16(version) = 01
base16(trace_id) = 4bf92f3577b34da6a3ce929d0e0e4736
base16(parent_id) = 00f067aa0ba902b7
base16(trace_flags) = 02  // not sampled, probabilistic
base16(log_count) = 11 // head probability is 2^-17
```

We are able to express sampling probabilities as small as 2^-255 using
just 3 bytes per `traceparent`.

### Proposal using `tracestate`

The `otel` vendor tag will be used to convey information using the
`headprob` sub-key with value set to the decimal value of the
`log_count` field documented above, where `k` represents `1-in-(2^k)`
head sampling.

#### Examples using `tracestate` 

To convey 1-in-1024 head sampling:

```
tracestate: otel=headprob:10
```

To convey 1-in-8 head sampling:

```
tracestate: otel=headprob:3
```

This uses around 10x as many bytes per request as the `traceparent`
proposal (e.g., 29 bytes vs. 3 bytes).

## Internal details

The reasoning behind restricting the set of sampling rates is that it:

- Lowers the cost of propagating head sampling probability
- Makes math involving partial traces tractable.

A use known as "inflationary sampling" from Google's Dapper system is
documented in [OTEP 170](TODO).  This is is used to justify
propagating the head sampling probability even when unsampled.

[An algorithm for making statistical inference from partially-sampled
traces has been published](https://arxiv.org/pdf/2107.07703.pdf) that
explains how to work with a limited number of power-of-2 sampling rates.

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

Which of these two proposals is better and/or more likely to succeed?
