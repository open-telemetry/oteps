# Probability sampling of telemetry events

Specify a foundation for sampling techniques in OpenTelemetry.

## Motivation

In tracing, metrics, and logs, there are widely known techniques for
sampling a stream of events that, when performed correctly, enable
collecting a tiny fraction of data while maintaining substantial
visibility into the whole population of events described by the data.

While sampling techniques vary, it is possible specify high-level
interoperability requirements that producers and consumers of sampled
data may follow to enable a wide range of sampling designs.

## Explanation

Consider a hypothetical telemetry signal in which an API event
produces a unit of data that has one or more associated numbers.
Using the OpenTelemetry Metrics data model terminology, we have two
scenarios in which sampling is common.

1. _Counter events:_ Each event represents a count, signifying the change in a sum.
2. _Histogram events:_ Each event represents an individual variable, signifying new membership in a distribution.

A Tracing Span event qualifies as both of these cases simultaneously.
It is a Counter event (of 1 span) and at lesat one Histogram event
(e.g., one of latency, one of request size).

In Metrics, [Statsd Counter and Histogram events meet this definition](https://github.com/statsd/statsd/blob/master/docs/metric_types.md#sampling).

In both cases, the goal in sampling is to estimate something about the
population of all events, using only the events that were chosen in
the sample.  Sampling theory defines various _sampling estimators_,
algorithms for calculating statistics about the population using just
the sample data.  For the broad class of telemetry sampling
application considered here, we need an estimator for the population
total represented by each individual event.

### Model and terminology

This model is meant to apply in telemetry collection situations where
individual events at an API boundary are sampled for collection.

In sampling, the term _sampling design_ refers to how sampling
probability is decided for a collection process and the term _sample
frame_ refers to how events are organized into discrete populations.

After executing a sampling design over a frame, each item selected in
the sample will have known _inclusion probability_, that determines
how likely it was to be selected.  Implicitly, all the items that were
not selected for the sample have zero inclusion probability.

Descriptive words that are often used to describe sampling designs:

- *Fixed*: the sampling design is the same from one frame to the next
- *Adaptive*: the sampling design changes from one frame to the next
- *Equal-Probability*: the sampling design uses a single inclusion probability per frame
- *Unequal-Probability*: the sampling design uses mulitple inclusion probabilities per frame
- *Reservoir*: the sampling design uses fixed space, has fixed-size output.

Our goal is to support flexibility in choosing sampling designs for
producers of telemetry data, while allowing consumers of sampled
telemetry data to be agnostic to the sampling design used.

We are interested in the common case for telemetry collection, where
sampling is performed while processing a stream of events and each
event is considered just once.  Sampling designs of this form are
referred to as _sampling without replacement_.  Unless stated
otherwise, "sampling" in telemetry collection always refers to
sampling without replacement.

After executing a given sampling design over complete frame of data,
the result is a set of selected sample events, each having known and
non-zero inclusion probability.  There are several other quantities of
interest, after calculating a sample from a sample frame.

- *Sample size*: the number of events with non-zero inclusion probability
- *True population total*: the exact number of events in the frame
- *Estimated population total*: the estimated number of events in the frame

The sample size is known after it is calculated, but the size may or
may not be known ahead of time, depending on the design.  The true
population total cannot be inferred directly from the sample, but can
(sometimes) be counted separately.  The estimated population total is
the expected value of the true population total.

### Adjusted sample count

Following the modle above, every event defines the notion of an
_adjusted count_.

- _Adjusted count_ is zero if the event was not selected for the sample
- _Adjusted count_ is the reciprocal of its inclusion probability, otherwise.

The adjusted count of an event represents the expected contribution to
the estimated population total of a sample framer represented by the
individual event.  As stated, the sample event's adjusted count is
easily derived from the Horvitz-Thompson estimator of the population
total, a general-purpose statistical estimator that applies to all
_without replacement_ sampling designs.

Assuming sample data is correctly computed, the consumer of sample
data can treat every sample event as though an identical copy of
itself has occurred _adjusted count_ times.  Every sample event is
representative for adjusted count many copies of itself.

There is one essential requirement for this to work.  The selection
procedure must be _statistically unbiased_, a term meaning that the
process is required to give equal consideration to all possible
outcomes.

### Encoding inclusion probability

Some possibilities for encoding the inclusion probability, depending
on the circumstances and the protocol, briefly discussed:

1. Encode the adjusted count directly as a floating point or integer number in the range [0, +Inf).  This is a conceptually easy way to understand sampling because larger numbers mean greater representivity.
2. Encode the inclusion probability directly as a floating point number in the range [0, 1).  This is typical of the Statsd format, where each line includes an optional probability.  In this context, the probability is commonly referred to as a "sampling rate".  In this case, smaller numbers mean greater representivity.
3. Encode the negative of the base-2 logarithm of inclusion probability.  This restricts inclusion probabilities to powers of two and allows the use of small non-negative integers to encode power-of-two adjusted counts.
4. Fold the adjusted count into the data.  This is appropriate when the data itself carries counts, such as for OTLP Metrics Sum and Histogram points encoded using delta aggregation temporality.  This may lead to rounding errors, when adjusted counts are not integer valued.

This is not an exhaustive list of approaches.  All of these techniques
are considered appropriate.

A telemetry system should be able to accurately estimate the number of
events that took place whether the events were sampled or not, which
requires being able to recognize a zero value for the adjusted count
as being distinct from a raw event where no sampling took place.

#### Recognizing zero adjusted count

An adjusted count of zero indicates an event that was recorded, where
according to the sampling design its inclusion probability is zero.
These events are may be included in a stream of sampled events as
auxiliary information, and consu

Recording events with zero adjusted count are a useful way to record
auxiliary information while sampling, events that are considered
interesting but which are accounted for by the adjusted count of other
events in the same stream.

Consider a span sampling design that applies a sampling decision only
to the roots of a trace.  Non-root spans must be recorded when their
parent span is selected for a sample.  For a class of spans that
sometimes are and sometimes are not the root of a trace, we have three
outcomes:

1. Span is part of another trace: adjusted count is zero
2. Span is a root, was selected for the sample: adjusted count is non-zero
3. Span is a root, was not selected for the sample: not recorded.

### Sampling with attributes

Sampling is a powerful approach when used with event data that has
been annotated with key-value attributes and sampled with an unbiased
design.  It is possible to select arbitrary subsets use those subsets
to estimate the count of arbitrary subsets of the population.

This application for sample data is prescribed by the statement above,
"Every sample event is representative for adjusted count many copies
of itself."  It relies on the use of an unbiased sampling design.
Readers are referred to [recommended reading](#recommended-reading)
for more resources on sampling with attributes.

### Summary: a general technique
=======
been annotated with key-value attributes.  It is possible to select
arbitrary subsets of the sampled data and use each to estimate the count
of arbitrary subsets of the population.

To summarize, there is a widely applicable procedure for sampling
telemetry data from a population:

- describe how to map telemetry events into discrete frames
- use an unbiased sampling design to select events
- encode the adjusted count or inclusion probability in the recorded events
- apply a predicate to events in the sample to select a subset of events
- sum the adjusted counts of the subset to estimate the sub-population total.

Applied correctly, this approach provides accurate estimates for
population counts and distributions with support for ad-hoc queries
over the data.

### Changes proposed

This OTEP proposes no formal changes in the OpenTelemetry
specitication.  It is meant to lay a foundation for importing sampled
telemetry events from other systems as well as to begin specifying how
OpenTelemetry SDKs that use probabilistic `Sampler` implementations
should convey inclusion probability and how consumers of this
information can use information about sampling.

### Example: Dapper tracing

Google's [Dapper](https://research.google/pubs/pub36356/) tracing
system describes the use of sampling to control the cost of trace
collection at scale.

The paper spends little time talking about Dapper's specific approach
to sampling, which evolved over time.  Dapper made use of tracing
context, similar to OpenTelemetry Baggage, to convey the probability
that the current trace was selected for sampling.  This allowed each
node in the trace to make an independent decision to begin sampling
with themselves as a new root.  This technique can ensure a minimum
rate of traces being started by every node in the system, however this
is not described by the Dapper paper.

### Example: Statsd metrics

A Statsd counter event appears as a line of text.

For example, a metric
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

Probability 0.1 leads to an adjusted count of 10.  Assuming the sample
was selected using an unbiased algorithm (as by a "fair coin"), we can
interpret this event as probabilistically equal to a count of `100/0.1
= 1000`.

## Internal details

The statistical foundation of this technique is known as the
Horvitz-Thompson estimator ("A Generalization of Sampling Without
Replacement From a Finite Universe," JSTOR 1952).  The
Horvitz-Thompson technique works with _unequal probability sampling_
designs, enabling a variety of techniques for controlling properties
of the sample.

For example, you can sample 100% of error events while sampling 1% of
non-error events, and the interpretation of adjusted count will be
correct.

### Bias, variance, and sampling errors

There is a fundamental tradeoff between bias and variance in
statistics.  The use of unbiased sampling leads unavoidably to
increased variance.

Estimating sampling errors and variance is out of scope for this
proposal.  We are satisfied with the unbiased property, which guarantees
that the expected value of totals derived from the sample equals the
true population totals.  This means that statistics derived from
sample data in this way are always accurate, and that more data will
always improve precision.

### Non-probabilistic rate-limiters

Rate-limiting stages in a telemetry collection pipeline interfere with
sampling schemes when they operate in non-probabilistic ways.  When
implementing a non-probabilistic form of rate-limiting, processors
MUST set the adjusted count to zero.

The use of zero adjusted count explicitly conveys that the events
output by non-probabilistic sampling should not be counted in a
statistical manner.

## Prior art and alternatives

The term "adjusted count" is proposed because the resulting value is
effectively a count and may be used in place of the exact count.

The term "adjusted weight" is NOT proposed to describe the adjustment
made by sampling, because the adjustment being made is that of a count.

Another term for the proposed "adjusted count" concept is
`inverse_probability`.

"Subset sum estimation" is the name given to this topic within the
study of computer science and engineering.

## Recommended reading

[Sampling, 3rd Edition, by Steven K. Thompson](https://www.wiley.com/en-us/Sampling%2C+3rd+Edition-p-9780470402313).

[Performance Is A Shape. Cost Is A Number: Sampling](https://docs.lightstep.com/otel/performance-is-a-shape-cost-is-a-number-sampling), 2020 blog post, Joshua MacDonald

[Priority sampling for estimation of arbitrary subset sums](https://dl.acm.org/doi/abs/10.1145/1314690.1314696)
