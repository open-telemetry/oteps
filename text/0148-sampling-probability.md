# Probability sampling of telemetry events

<!-- toc -->

- [Motivation](#motivation)
- [Explanation](#explanation)
  * [Model and terminology](#model-and-terminology)
    + [Sampling without replacement](#sampling-without-replacement)
    + [Adjusted sample count](#adjusted-sample-count)
    + [Introducing variance](#introducing-variance)
  * [Conveying the sampling probability](#conveying-the-sampling-probability)
    + [Encoding adjusted count](#encoding-adjusted-count)
    + [Encoding inclusion probability](#encoding-inclusion-probability)
    + [Encoding negative base-2 logarithm of inclusion probability](#encoding-negative-base-2-logarithm-of-inclusion-probability)
    + [Multiply the adjusted count into the data](#multiply-the-adjusted-count-into-the-data)
  * [Trace Sampling](#trace-sampling)
    + [Counting spans and traces](#counting-spans-and-traces)
    + [`Parent` Sampler](#parent-sampler)
    + [`TraceIDRatio` Sampler](#traceidratio-sampler)
    + [Dapper's "Inflationary" Sampler](#dappers-inflationary-sampler)
  * [Event sampling](#event-sampling)
    + [Weighted sampling](#weighted-sampling)
    + [Example: Statsd](#example-statsd)
    + [Example: Two-pass sampling](#example-two-pass-sampling)
    + [Example: Combining samples](#example-combining-samples)
    + [Example: Downsampling](#example-downsampling)
    + [Example: Multiple samples](#example-multiple-samples)
- [Propoesed specification changes](#propoesed-specification-changes)
- [Recommended reading](#recommended-reading)
- [Acknowledgements](#acknowledgements)

<!-- tocstop -->

Objective: Specify a foundation for sampling techniques in OpenTelemetry.

## Motivation

In tracing, metrics, and logs, there are widely known techniques for
sampling a stream of events that, when performed correctly, enable
collecting a fraction of the complete data while maintaining
substantial visibility into the whole population of events.

These techniques are all forms of approximate counting.  Estimates
calculated by the forms of sampling outlined here are considered
accurate, in the sense that they are random variables with an expected
value equal to their true value.

While sampling techniques vary, it is possible to specify high-level
interoperability requirements that producers and consumers of sampled
data can follow to enable a wide range of sampling designs.

## Explanation

Consider a hypothetical telemetry signal in which an API event
produces a unit of data that has one or more associated numbers.
Using the OpenTelemetry Metrics data model terminology, we have two
scenarios in which sampling is common.

1. _Counter events:_ Each event represents a count, signifying the change in a sum.
2. _Histogram events:_ Each event represents an individual variable, signifying membership in a distribution.

A Tracing Span event qualifies as both of these cases simultaneously.
It is at least one Counter event (e.g., one request, the number of
bytes read) and at least one Histogram event (e.g., request latency,
request size).

In Metrics, [Statsd Counter and Histogram events meet this definition](https://github.com/statsd/statsd/blob/master/docs/metric_types.md#sampling).

In both cases, the goal in sampling is to estimate the count of events
in the whole population, meaning all the events, using only the events
that were selected in the sample.

### Model and terminology

This model is meant to apply in telemetry collection situations where
individual events at an API boundary are sampled for collection.  Once
the process of sampling individual API-level events is understood, we
will learn to apply these techniques for sampling aggregated data.

In sampling, the term _sampling design_ refers to how sampling
probability is decided and the term _sample frame_ refers to how
events are organized into discrete populations.

For example, a simple design uses uniform probability, and a simple
framing technique is to collect one sample per distinct span name.

After executing a sampling design over a frame, each item selected in
the sample will have known _inclusion probability_, that determines
how likely the item was to being selected.  Implicitly, all the items
that were not selected for the sample have zero inclusion probability.

Descriptive words that are often used to describe sampling designs:

- *Fixed*: the sampling design is the same from one frame to the next
- *Adaptive*: the sampling design changes from one frame to the next
- *Equal-Probability*: the sampling design uses a single inclusion probability per frame
- *Unequal-Probability*: the sampling design uses multiple inclusion probabilities per frame
- *Reservoir*: the sampling design uses fixed space, has fixed-size output.

Our goal is to support flexibility in choosing sampling designs for
producers of telemetry data, while allowing consumers of sampled
telemetry data to be agnostic to the sampling design used.

#### Sampling without replacement

We are interested in the common case in telemetry collection, where
sampling is performed while processing a stream of events and each
event is considered just once.  Sampling designs of this form are
referred to as _sampling without replacement_.  Unless stated
otherwise, "sampling" in telemetry collection always refers to
sampling without replacement.

After executing a given sampling design over a complete frame of data,
the result is a set of selected sample events, each having known and
non-zero inclusion probability.  There are several other quantities of
interest, after calculating a sample from a sample frame.

- *Sample size*: the number of events with non-zero inclusion probability
- *True population total*: the exact number of events in the frame, which may be unknown
- *Estimated population total*: the estimated number of events in the frame, which is computed from the sample.

The sample size is always known after it is calculated, but the size
may or may not be known ahead of time, depending on the design.
Probabilistic sampling schemes require that the estimated population
total equals the expected value of the true population total.

#### Adjusted sample count

Following the model above, every event defines the notion of an
_adjusted count_.

- _Adjusted count_ is zero if the event was not selected for the sample
- _Adjusted count_ is the reciprocal of its inclusion probability, otherwise.

The adjusted count of an event represents the expected contribution to
the estimated population total of a sample frame represented by the
individual event.

The use of a reciprocal inclusion probability matches our intuition
for probabilities.  Items selected with "one-out-of-N" probability of
inclusion count for N each, approximately speaking.

This intuition is backed up with statistics.  This equation is known
as the Horvitz-Thompson estimator of the population total, a
general-purpose statistical "estimator" that applies to all _without
replacement_ sampling designs.

Assuming sample data is correctly computed, the consumer of sample
data can treat every sample event as though an identical copy of
itself has occurred _adjusted count_ times.  Every sample event is
representative for adjusted count many copies of itself.

There is one essential requirement for this to work.  The selection
procedure must be _statistically unbiased_, a term meaning that the
process is required to give equal consideration to all possible
outcomes.

#### Introducing variance

The use of unbiased sampling outlined above makes it possible to
estimate the population total for arbitrary subsets of the sample, as
every individual sample has been independently assigned an adjusted
count.

There is a natural relationship between statistical bias and variance.
Approximate counting comes with variance, a matter of fact which can
be controlled for by the sample size.  Variance is unavoidable in an
unbiased sample, but it vanishes when you have enough data.

Although this makes it sounds like small sample sizes are a problem,
due to expected high variance, this is just a limitation of the
technique.  When variance is high, use a larger sample size.

An easy approach for lowering variance is to aggregate sample frames
together across time.  For example, although the estimates drawn from
a one-minute sample may have high variance, combining an hour of
one-minute sample frames into an aggregate data set is guaranteed to
lower variance.  It must, because the data remains unbiased.

### Conveying the sampling probability

Some possibilities for encoding the adjusted count or inclusion
probability are discussed below, depending on the circumstances and
the protocol.

There are several ways of encoding this information:

- as a dedicated field in an OTLP protobuf message
- as a non-descriptive Attribute in an OTLP Span, Metric, or Log
- without any dedicated field.

#### Encoding adjusted count

We can encode the adjusted count directly as a floating point or
integer number in the range [0, +Inf).  This is a conceptually easy
way to understand sampling because larger numbers mean greater
representivity.

#### Encoding inclusion probability

We can encode the inclusion probability directly as a floating point
number in the range [0, 1).  This is typical of the Statsd format,
where each line includes an optional probability.  In this context,
the probability is also commonly referred to as a "sampling rate".  In
this case, smaller numbers mean greater representivity.

#### Encoding negative base-2 logarithm of inclusion probability

We can encode the negative base-2 logarithm of inclusion probability.
This restricts inclusion probabilities to powers of two and allows the
use of small non-negative integers to encode power-of-two adjusted
counts.  In this case, larger numbers mean exponentially greater
representivity.

#### Multiply the adjusted count into the data

When the data itself carries counts, such as for the Metrics Sum and
Histogram points.

This technique is less desirable because while it preserves the
expected value of the count or sum, the data loses information about
variance.  This may also lead to rounding errors, when adjusted counts
are not integer valued.

### Trace Sampling

Sampling techniques are always about lowering the cost of data
collection and analsyis, but in trace collection and analsysis
specifically, approaches can be categorized by whether they reduce
Tracer overhead.  Tracer overhead is reduced by not recording spans
for unsampled traces and requires making the sampling decision for a
trace before all of its attributes are known.

Traces are expected to be complete, meaning that a tree or sub-tree of
spans branching from a certain root are expected to be fully
collected.  When sampling is applied to reduce Tracer overhead, there
is generally an expectation that complete traces will still be
produced.  Sampling techniques that lower Tracer overhead and produce
complete traces are known as _Head trace sampling_ techniques.

The decision to produce and collect a sample trace has to be made when
the root span starts, to avoid incomplete traces.  Using the sampling
techniques outlined above, we can approximately count finished spans
and traces, even without knowing how the head trace sampling decision
was made.

#### Counting spans and traces

When the [W3C Trace Context is-sampled
flag](https://www.w3.org/TR/trace-context/#sampled-flag) is used to
propagate a sampling decision, child spans have the same adjusted
count as their parent.  This leads to a useful optimization.

It is nice-to-have, though not a requirement, that all spans in a
trace directly encode their adjusted count.  This enables systems to
count spans upon arrival, without the work of referring to their
parent spans.  For example, knowing a span's adjusted count makes it
possible to immediately produce metric events from span events.

Several head sampling techniques are discussed in the following
sections and evaluated in terms of their ability to meet all of the
following criteria:

- Reduces Tracer overhead
- Produces complete traces
- Spans are countable.

#### `Parent` Sampler

The `Parent` Sampler ensures complete traces, provided all spans are
successfully recorded.  A downside of `Parent` sampling is that it
takes away control over Tracer overhead from non-roots in the trace.
To support counting spans, this Sampler requires propagating the
effective adjusted count of the context to use when starting child
spans.

In other head trace sampling schemes, we will see that it is useful to
propagate inclusion probability even for negative sampling decisions
(where the adjusted count is zero), therefore we prefer to use the
inclusion probability and not the adjusted count when propagating the
sampling rate via trace context.  The inclusion probability of a
context is referred to as its `head inclusion probability` for this
reason.

In addition to propagating head inclusion probability, to count
Parent-sampled spans, each span must directly encode its adjusted
count in the corresponding `SpanData`.  This may use a non-descriptive
Resource or Span attribute named `sampling.parent.adjusted_count`, for
example.

#### `TraceIDRatio` Sampler

The OpenTelemetry tracing specification includes a built-in Sampler
designed for probability sampling using a deterministic sampling
decision based on the TraceID.  This Sampler was not finished before
the OpenTelemetry version 1.0 specification was released; it was left
in place, with [a TODO and the recommendation to use it only for trace
roots](https://github.com/open-telemetry/opentelemetry-specification/issues/1413).
[OTEP 135 proposed a solution](https://github.com/open-telemetry/oteps/pull/135).

The goal of the `TraceIDRatio` Sampler is to coordinate the tracing
decision, but give each service control over Tracer overhead.  Each
service sets its sampling probability independently, and the
coordinated decision ensures that some traces will be complete.
Traces are complete when the TraceID ratio falls below the minimum
Sampler probability across the whole trace.

The `TraceIDRatio` Sampler has another difficulty with testing for
completeness.  It is impossible to know whether there are missing leaf
spans in a trace without using external information.  One approach,
[lost in the transition from OpenCensus to OpenTelemetry is to count
the number of children of each
span](https://github.com/open-telemetry/opentelemetry-specification/issues/355).

Lacking the number of expected children, we require a way to know the
minimum Sampler probability across traces to ensure they are complete.

To count TraceIDRatio-sampled spans, each span must encode its
adjusted count in the corresponding `SpanData`.  This may use a
non-descriptive Resource or Span attribute named
`sampling.traceidratio.adjusted_count`, for example.

#### Dapper's "Inflationary" Sampler

Google's [Dapper](https://research.google/pubs/pub36356/) tracing
system describes the use of sampling to control the cost of trace
collection at scale.  Dapper's early Sampler algorithm, referred to as
an "inflationary" approach (although not published in the paper), is
reproduced here.

This kind of Sampler allows non-root spans in a trace to raise the
probability of tracing, using a conditional probability formula shown
below.  Traces produced in this way are complete sub-trees, not
necessarily complete.  This technique is succesful especially in
systems where a high-throughput service on occasion calls a
low-throughput service.  Low-throughput services are meant to inflate
their sampling probability.

The use of this technique requires propagating the head inclusion
probability (as discussed for the `Parent` sampler) of the incoming
Context and whether it was sampled, in order to calculate the
probability of starting to sample a new "sub-root" in the trace.

Using standard notation for conditional probability, `P(x)` indicates
the probability of `x` being true, and `P(x|y)` indicates the
probability of `x` being true given that `y` is true.  The axioms of
probability establish that:

```
P(x)=P(x|y)*P(y)+P(x|not y)*P(not y)
```

The variables are:

- **`H`**: The head inclusion probability of the parent context that
  is in effect, independent of whether the parent context was sampled,
  the reciprocal of the parent context's effective adjusted count.
- **`I`**: The inflationary sampling probability for the span being
  started.
- **`D`**: The decision probability for whether to start a new sub-root.

This Sampler cannot lower sampling probability, so if the new span is
started with `H >= I` or when the context is already sampled, no new
sampling decisions are made.  If the incoming context is already
sampled, the adjusted count of the new span is `1/H`.

Assuming `H < I` and the incoming context was not sampled, we have the
following probability equations:

```
P(span sampled) = I
P(parent sampled) = H
P(span sampled | parent sampled) = 1
P(span sampled | parent not sampled) = D
```

Using the formula above, 

```
I = 1*H + D*(1-H)
```

solve for D:

```
D = (I - H) / (1 - H)
```

Now the Sampler makes a decision with probability `D`.  Whether the
decision is true or false, propagate `I` as the new head inclusion
probability.  If the decision is true, begin recording a sub-rooted
trace with adjusted count `1/I`.  This may use a non-descriptive
Resource or Span attribute named
`sampling.inflationary.adjusted_count`, for example.

### Working with adjusted counts

Head sampling for traces has been discussed, covering strategies to
lower Tracer overhead, ensure trace completeness, and count spans on
arrival.  Sampled spans have an added attribute to directly encode the
adjusted count, and the sum of adjusted counts for a set of spans
accurately reflects the total population count.

In systems based on collecting sample data, it is often useful to
combine samples to maintain a small data set.  For example, given 24
one-hour samples of 1000 spans each, can we combine the data into a
one-day sample of 1000 spans?  To do this without introducing bias, we
must take the adjusted count of each span into account.  Sampling
algorithms that can do this are known as weighted sampling algorithms.

TODO

#### Weighted sampling

TODO

#### Example: Statsd 

A Statsd counter event appears as a line of text, describing a
number-valued event with optional attributes and sample rate.

For example, a metric named `name` is incremented by `increment` using
a counter event (`c`) with the given `sample_rate`.

```
name:increment|c|@sample_rate
```

For example, a count of 100 that was selected for a 1-in-10 simple
random sampling scheme will arrive as:

```
counter:100|c|@0.1
```

Probability 0.1 leads to an adjusted count of 10.  Assuming the sample
was selected using an unbiased algorithm, we can interpret this event
as having an expected value of `100/0.1 = 1000`.

#### Example: Two-pass sampling

TODO

#### Example: Combining samples

TODO

#### Example: Downsampling

TODO

#### Example: Multiple samples

TODO

## Propoesed specification changes

TODO

## Recommended reading

[Sampling, 3rd Edition, by Steven K. Thompson](https://www.wiley.com/en-us/Sampling%2C+3rd+Edition-p-9780470402313).

[Performance Is A Shape. Cost Is A Number: Sampling](https://docs.lightstep.com/otel/performance-is-a-shape-cost-is-a-number-sampling), 2020 blog post, Joshua MacDonald

[Priority sampling for estimation of arbitrary subset sums](https://dl.acm.org/doi/abs/10.1145/1314690.1314696)

[A Generalization of Sampling Without Replacement From a Finite Universe](https://www.jstor.org/stable/2280784), JSTOR (1952)

## Acknowledgements

Thanks to [Neena Dugar](https://github.com/neena) and [Alex
Kehlenbeck](https://github.com/akehlenbeck) for reconstructing the
Dapper Sampler algorithm.
