# Probability sampling of telemetry events

Specify a foundation for sampling techniques in OpenTelemetry.

## Motivation

In both tracing and metrics, there are widely known techniques for
sampling events that, when performed correctly, enable ways to lower
collection costs.  While sampling techniques vary, it is possible
specify high-level interoperability requirements that producers and
consumers of sampled data may follow to enable a wide range of
sampling designs.

## Explanation

Consider a hypothetical telemetry signal in which an API event
produces a unit of data that has one or more associated numbers.
Using the OpenTelemetry Metrics data model terminolgy, we have two
scenarios in which sampling is common.

1. Counter events: Each event represents a count, signifying the change in a sum
2. Histogram events: Each event represents an individual variable, signifying new membership in a distribution

A Tracing Span event qualifies as both of these cases simultaneously.  It is
a Counter event (of 1 span) and a Histogram event (of 1 latency measurement).

In Metrics, [Statsd Counter and Histogram events meet this definition](https://github.com/statsd/statsd/blob/master/docs/metric_types.md#sampling).

In both cases, the goal in sampling is to estimate something about the
population of all events, using only the events that were chosen in
the sample.  Sampling theory defines various kinds of "estimator",
algorithms for calculating statistics about the population using just
the sample data.  For the broad class of telemetry sampling
application considered here, we need an estimator for the population
total represented by each individual event.

### Sample count adjustment

The estimated population total divided by the individual event value
equals the event's _adjusted count_.  The adjusted count is named
`sample_count` where it appears as a field in a telemetry data type.
The field indicates that the individual sample event is estimated to
represent `sample_count` number of identical events in the population.

A standard sampling adjustment will be defined, and for it to work
there is one essential requirement.  The selection procedure must be
_unbiased_, a statistical term meaning that the process is expected to
give equal consideration to all possible outcomes.

The specified sampling adjustment sets the `sample_count` of each
sampled event to the inverse of the event's inclusion probability.
Conveying the inverse of the inclusion probability is convenient for
several reasons:

- A `sample_count` of 1 indicates no sampling was performed
- A `sample_count` of 0 indicates an unrepresentative event
- Larger `sample_count` indicates greater representivity
- Smaller `sample_count` indicates smaller representivity
- The sum of `sample_count` in a sample equals the expected 
  value of the population size.
  
The zero `sample_count` value supports collecting values that were
rejected from a sample to use as exemplars, which supports encoding of
values that should not be counted in an estimate of the population
total.  This has applications to reservoir sampling designs, where
events may be selected for a sample only to be rejected before the end
of the frame.

### Sampling with attributes

Sampling is a powerful approach when used with event data that has
been annotated with key-value attributes.  It is possible to select
arbitrary subsets of the sampled data and use it to estimate the count
of arbitrary subsets of the population.

To summarize, there is a widely applicable procedure for sampling
telemetry data from a population:

- use an unbiased sampling algorithm to select telemetry events
- label each event in the sample with its `sample_count` (i.e., inverse inclusion probability)
- apply a predicate to events in the sample to select a subset
- apply an estimator to the subset to estimate the sub-population total.

Applied correctly, this approach provides accurate estimates for
population counts and distributions with support for ad-hoc queries
over the data.

### Changes proposed

This proposal leads to three change requests that will be carried out in
separate places in the OpenTelemetry specification.  These are:

1. For tracing, the SpanData message type should be extended with 
   the `sample_count` field defined above.
2. For metrics aggregate data: Count information aggregated from
   sample metric events will have floating point values in general.
   Histogram and Counter data must be able to support floating point
   values.
3. For metrics raw events: Exemplars should be extended with the 
   `sample_count` field defined above.

### Example: Dapper tracing

Google's [Dapper](https://research.google/pubs/pub36356/) tracing
system describes the use of sampling to control the cost of trace
collection at scale.

### Example: Statsd metrics

A Statsd counter event appears as a line of text, for example a metric
named `name` is incremented by `increment` using a counter event (`c`)
with the given `sample_rate`.

```
name:increment|c|@sample_rate
```

For example, a count of 100 that was selected for a 1-in-10 simple
random sampling scheme will arrive as:

```
counter:100|c|@0.1
```

Probability 0.1 leads to a `sample_count` of 10.  Assuming the sample
was selected using an unbiased algorithm (as by a "fair coin"), we can
interpret this event as probabilistically equal to a count of `100/0.1
= 1000`.

## Internal details

The statistical foundation of this technique is known as the
Horvitz-Thompson estimator ("A Generalization of Sampling Without
Replacement From a Finite Universe", JSTOR 1952).  The
Horvitz-Thompson technique works with _unequal probability sampling_
designs, enabling a variety of techniques for controlling properties
of the sample.  

For example, you can sample 100% of error events while sampling 1% of
non-error events, and the interpretation of `sample_count` will be
correct.

### Bias, variance, and sampling errors

There is a fundamental tradeoff between bias and variance in
 statistics.  The use of unbiased sampling leads unavoidably to
increased variance.

Estimating sampling errors and variance is out of scope for this
proposal.  We are satisfied the unbiased property, which guarantees
that the expected value of totals derived from the sample equals the
true population totals.  This means that statistics derived from
sample data in this way are always accurate, and that more data will
always improve precision.

### Non-probabilistic rate-limiters

Rate-limiting stages in a telemetry collection pipeline interfere with
sampling schemes when they operate in non-probabilistic ways.  When
implementing a non-probabilistic form of rate-limiting, processors
MUST set the `sample_count` to a NaN value.

### No Sampler configured

When no Sampler is in place and all telemetry events pass to the
output, the `sample_count` field SHOULD be set to 1 to indicate
perfect representivity, indicating that no sampling was performed.

### Applicability for tracing

When using sampling to limit span collection, there are usually
approaches under consideration.  The sampling approach covered here
dictates how to select root spans in a probabilistic way.  When
recording root spans, the `sample_count` field should be set as
described above.  The adjusted `sample_count` of the root span applies
the trace, meaning the trace should be considered as representative of
`sample_count` traces in the population.

When non-root spans are recorded because they are part of an ongoing
trace, they are considered non-probabilistic exemplars.  Non-root
spans should have `sample_count` set to zero.

### Applicability for metrics

The use of sampling in metrics makes it possible to record
high-cardinality metric events efficiently, as demonstrated by Statsd.

By pushing sampled metric events from client to server, instead of
timeseries, it is possible to defer decisions about cardinality
reduction to the server, without unreasonable cost to the client.

## Prior art and alternatives

The name `sample_count` is proposed because the resulting value is
effectively a count and may be used in place of the exact count.

Statsd conveys inclusion probability instead of `sample_count`, where
it is often called "sample rate".

Another name for the proposed `sample_count` field is
`inverse_probability`, which is considered less suggestive of the
field's purpose.

"Subset sum estimation" is the name given to this topic within the
study of computer science and engineering.
