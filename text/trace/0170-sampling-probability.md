# Probability sampling of telemetry events

<!-- toc -->

- [Motivation](#motivation)
- [Examples](#examples)
  * [Span sampling](#span-sampling)
    + [Sample spans to Counter Metric](#sample-spans-to-counter-metric)
    + [Sample spans to Histogram Metric](#sample-spans-to-histogram-metric)
    + [Sample span rate limiting](#sample-span-rate-limiting)
  * [Metric sampling](#metric-sampling)
    + [Statsd Counter](#statsd-counter)
    + [Metric exemplars with adjusted counts](#metric-exemplars-with-adjusted-counts)
    + [Metric cardinality limiter](#metric-cardinality-limiter)
- [Explanation](#explanation)
  * [Model and terminology](#model-and-terminology)
    + [Sampling without replacement](#sampling-without-replacement)
    + [Adjusted sample count](#adjusted-sample-count)
    + [Sampling and variance](#sampling-and-variance)
  * [Conveying the sampling probability](#conveying-the-sampling-probability)
    + [Encoding adjusted count](#encoding-adjusted-count)
    + [Encoding inclusion probability](#encoding-inclusion-probability)
    + [Encoding base-2 logarithm of adjusted count](#encoding-base-2-logarithm-of-adjusted-count)
    + [Multiply the adjusted count into the data](#multiply-the-adjusted-count-into-the-data)
  * [Trace Sampling](#trace-sampling)
    + [Counting child spans using root span adjusted counts](#counting-child-spans-using-root-span-adjusted-counts)
	+ [Using head trace probability to count all spans](#using-head-trace-probability-to-count-all-spans)
    + [Head sampling for traces](#head-sampling-for-traces)
      - [`Parent` Sampler](#parent-sampler)
      - [`TraceIDRatio` Sampler](#traceidratio-sampler)
      - [Dapper's "Inflationary" Sampler](#dappers-inflationary-sampler)
- [Proposed specification text](#proposed-specification-text)
- [Recommended reading](#recommended-reading)
- [Acknowledgements](#acknowledgements)

<!-- tocstop -->

Objective: Specify a foundation for sampling techniques in OpenTelemetry.

## Motivation

Probability sampling allows consumers of sampled telemetry data to
collect a fraction of telemetry events and use them to estimate total
quantities about the population of events, such as the total rate of
events with a particular attribute.  Sampling is a general-purpose
facility for lowering cost at the expense of lower data quality.

These techniques enable reducing the cost of telemetry collection,
both for producers (i.e., SDKs) and for processors (i.e., Collectors),
without losing the ability to (at least coarsely) monitor the whole
system.

Sampling builds on results from probability theory, most significantly
the concept of expected value.  Estimates drawn from probability
samples are *random variables* that, when correct procedures are
followed, accurately reflect their true value, making them unbiased.
Unbiased samples can be used for after-the-fact analysis.  We can
answer questions such as "what fraction of events had property X?"
using the fraction of events in the sample that have property X.

This document outlines how producers and consumers of sample telemetry
data can convey estimates about the total count of telemetry events,
without conveying information about how the sample was computed, using
a quantity known as **adjusted count**.  In common language, a
"one-in-N" sampling scheme emits events with adjusted count equal to
N.  Adjusted count is the expected value of the number of events in
the population represented by an individual sample event.

## Examples

These examples use the proposed attribute `sampler.adjusted_count` to
convey sampling probability.  Consumers of spans, metrics, and logs
annotated with adjusted counts are able to calculate accurate
statistics about the whole population of events, at a basic level,
without knowing details about the sampling configuration.

### Span sampling

Example use-cases for probability sampling of spans
generally involve generating metrics from spans.

#### Sample spans to Counter Metric

For every complete span it receives, the example processor will synthesize
metric data as though a Counter instrument named `S.count` for span
named `S` had been incremented once per span at the original `Start()`
call site.

This processor will add the adjusted count of each span to the
instrument (e.g., `Add(adjusted_count, labels...)`)  for every span it
receives, logically taking place at the start or end time of the span.

#### Sample spans to Histogram Metric

For every span it receives, the example processor will synthesize
metric data as though a Histogram instrument named `S.duration` for
span named `S` had been observed once per span at the original `End()`
call site.

The OpenTelemetry Metric data model does not support histogram buckets
with non-integer counts, which forces the use of integer adjusted
counts here (i.e., 1-in-N sampling rates where N is an integer).

Logically speaking, this processor will observe the span's duration its
adjusted count number of times for every span it receives, at the end
time of the span.

#### Sample span rate limiting

A collector processor will introduce a slight delay in order to ensure
it has received a complete frame of data, during which time it
maintains a fixed-size buffer of complete input spans.  If the number of spans
received exceeds the size of the buffer before the end of the
interval, begin weighted sampling using the adjusted count of each
span as input weight.

This processor drops spans when the configured rate threshold is
exceeeded, otherwise it passes spans through with unmodifed adjusted
counts.

When the interval expires and the sample frame is considered complete,
the selected sample spans are output with possibly updated adjusted
counts.

### Metric sampling

Example use-cases for probability sampling of metrics
are aimed at lowering cost and addressing high cardinality.

#### Statsd Counter

A Statsd counter event appears as a line of text, describing a
number-valued event with optional attributes and inclusion probability
("sample rate").

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

Events in the example have with 0.1 inclusion probability have
adjusted count of 10.  Assuming the sample was selected using an
unbiased algorithm, we can interpret this event as having an expected
count of `100/0.1 = 1000`.

#### Metric exemplars with adjusted counts

The OTLP protocol for metrics includes a repeated exemplars field in
every data point.  This is a place where Metric aggregators (e.g., 
histograms) are able to provide example context to correlate metrics 
with traces.

OTLP exemplars support additional attributes, those that were present
on the API event and were dropped during aggregation.  Exemplars that
are selected probabilistically and recorded with their adjusted counts
make it possible to approximately count events using dimensions that
were dropped during metric aggregation.

An end-to-end pipeline of sampled metrics events can be constructed
based on exemplars with adjusted counts, one capable of supporting
approximate-count queries over sampled metric events at high
cardinality.

#### Metric cardinality limiter

A metrics processor can be configured to limit cardinality for a
single metric name, allowing no more than K distinct label sets per
export interval.  The export interval is fixed to a short interval so
that a complete set of distinct labels can be stored temporarily.

Caveats: as presented, this works for Sum and Histogram points
received with Delta aggregation temporality and where the Sum is
monotonic (see
[opentelemetry-proto/issues/303](https://github.com/open-telemetry/opentelemetry-proto/issues/303)).

Considering data points received during the interval, when the number
of points exceeds K, select a probability proportional to size sample
of points, output every point with a `sampler.adjusted_count` attribute.

## Explanation

Consider a hypothetical telemetry signal in which a stream of
data items is produced containing one or more associated numbers.
Using the OpenTelemetry Metrics data model terminology, we have two
scenarios in which sampling is common.

1. _Counter events:_ Each event represents a count, signifying the change in a sum.
2. _Histogram events:_ Each event represents an individual variable, signifying membership in a distribution.

A Tracing Span event qualifies as both of these cases simultaneously.
One span can be interpreted as at least one Counter event (e.g., one
request, the number of bytes read) and at least one Histogram event
(e.g., request latency, request size).

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
events are organized into discrete populations.  The design of a 
sampling strategy dictates how the population is framed.

For example, a simple design uses uniform probability, and a simple
framing technique is to collect one sample per distinct span name per
hour.  A different sample framing could collect one sample across all
span names every 10 minutes.

After executing a sampling design over a frame, each item selected in
the sample will have known _inclusion probability_, that determines
how likely the item was to being selected.  Implicitly, all the items
that were not selected for the sample have zero inclusion probability.

Descriptive words that are often used to describe sampling designs:

- *Fixed*: the sampling design is the same from one frame to the next
- *Adaptive*: the sampling design changes from one frame to the next based on the observed data
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

#### Sampling and variance

The use of unbiased sampling outlined above makes it possible to
estimate the population total for arbitrary subsets of the sample, as
every individual sample has been independently assigned an adjusted
count.

There is a natural relationship between statistical bias and variance.
Approximate counting comes with variance, a matter of fact which can
be controlled for by the sample size.  Variance is unavoidable in an
unbiased sample, but variance diminishes with increasing sample size.

Although this makes it sound like small sample sizes are a problem,
due to expected high variance, this is just a limitation of the
technique.  When variance is high, use a larger sample size.

An easy approach for lowering variance is to aggregate sample frames
together across time, which generally increases the size of the
subpopulations being counted.  For example, although the estimates for
the rate of spans by distinct name drawn from a one-minute sample may
have high variance, combining an hour of one-minute sample frames into
an aggregate data set is guaranteed to lower variance (assuming the
numebr of span names stays fixed).  It must, because the data remains
unbiased, so more data results in lower variance.

### Conveying the sampling probability

Some possibilities for encoding the adjusted count or inclusion
probability are discussed below, depending on the circumstances and
the protocol.  Here, the focus is on how to count sampled telemetry
events in general, not a specific kind of event.  As we shall see in
the following section, tracing comes with addional complications.

There are several ways of encoding this adjusted count or inclusion
probability:

- as a dedicated field in an OTLP protobuf message
- as a non-descriptive Attribute in an OTLP Span, Metric, or Log
- without any dedicated field.

#### Encoding adjusted count

We can encode the adjusted count directly as a floating point or
integer number in the range [0, +Inf).  This is a conceptually easy
way to understand sampling because larger numbers mean greater
representivity.

Note that it is possible, given this description, to produce adjusted
counts that are not integers.  Adjusted counts are an approximatation,
and the expected value of an integer can be a fractional count.
Floating-point adjusted counts can be avoided with the use of
integer-reciprocal inclusion probabilities.

#### Encoding inclusion probability

We can encode the inclusion probability directly as a floating point
number in the range [0, 1).  This is typical of the Statsd format,
where each line includes an optional probability.  In this context,
the probability is also commonly referred to as a "sampling rate".  In
this case, smaller numbers mean greater representivity.

#### Encoding base-2 logarithm of adjusted count

We can encode the base-2 logarithm of adjusted count (i.e., negative
base-2 logarithm of inclusion probability).  By using an integer
field, restricting adjusted counts and inclusion probabilities to
powers of two, this allows the use of small non-negative integers to
encode the adjusted count.  In this case, larger numbers mean
exponentially greater representivity.

#### Multiply the adjusted count into the data

When the data itself carries counts, such as for the Metrics Sum and
Histogram points, the adjusted count can be multipled into the data.

This technique is less desirable because, while it preserves the
expected value of the count or sum, the data loses information about
variance.  This may also lead to rounding errors, when adjusted counts
are not integer valued.

### Trace Sampling

Sampling techniques are always about lowering the cost of data
collection and analysis, but in trace collection and analysis
specifically, approaches can be categorized by whether they reduce
Tracer overhead.  Tracer overhead is reduced by not recording spans
for unsampled traces and requires making the sampling decision for a
trace before all of its attributes are known.

Traces are expected to be complete, meaning that a tree or sub-tree of
spans branching from a certain root are expected to be fully
collected.  When sampling is applied to reduce Tracer overhead, there
is generally an expectation that complete traces will still be
produced.  Sampling techniques that lower Tracer overhead and produce
complete traces are known as _Head-based trace sampling_ techniques.

The decision to produce and collect a sample trace has to be made when
the root span starts, to avoid incomplete traces.  Then, assuming
complete traces can be collected, the adjusted count of the root span
determines an adjusted count for every span in the trace.

#### Counting child spans using root span adjusted counts

The adjusted count of a root span determines the adjusted count of
each of its children based on the following logic:

- The root span is considered representative of `adjusted_count` many
  identical root spans, because it was selected using unbiased sampling
- Context propagation conveys _causation_, the fact the one span produces 
  another
- A root span causes each of the child spans in its trace to be produced
- A sampled root span represents `adjusted_count` many traces, representing
  the cause of `adjusted_count` many occurances per child span in the 
  sampled trace.

Using this reasoning, we can define a sample collected from all root
spans in the system, which allows estimating the count of all spans in
the population.  Take a simple probability sample of root spans:

1. In the `Sampler` decision for root spans, use the initial span properties 
   to determine the inclusion probability `P`
2. Make a pseudo-random selection with probability `P`, if true return
   `RECORD_AND_SAMPLE` (so that the W3C Trace Context `is-sampled`
   flag is set in all child contexts)
3. Encode a span attribute `sampler.adjusted_count` equal to `1/P` on the root span
4. Collect all spans where the W3C Trace Context `is-sampled` flag is set.

After collecting all sampled spans, locate the root span for each.
Apply the root span's adjusted count to every child in the associated
trace.  The sum of adjusted counts on all sampled spans is expected to
equal the population total number of spans.

Now, having stored the sample spans with their adjusted counts, and
assuming the source of randomness is good, we can extrapolate counts
for the population using arbitrary queries over the sampled spans.
Sampled spans can be translated into approximate metrics over the
population of spans, after their adjusted counts are known.

The cost of this analysis, using only the root span's adjusted count,
is that all root spans have to be collected before we can count
non-root spans.  The cost of indexing and looking up the root span
adjusted counts makes this analysis relatively expensive to perform in
real time.

#### Using head trace probability to count all spans

If the W3C `is-sampled` flag will be used to determine whether
`RECORD_AND_SAMPLE` is returned in a Sampler, then in order to count
sample spans without first locating the root span requires propagating
the _head trace sampling probability_ through the context.

Head trace sampling probability may be thought of as the probability
of causing a child span to be a sampled.  Propagators that maintain
this variable MUST obey the rules of conditional probability.  In this
model, the adjusted count of each span depends on the adjusted count
of its parent, not of the root in a trace.  Still, the sum of adjusted
counts of all sampled spans is expected to equal the population total
number of spans.

This applies to other forms of telemetry that happen (i.e., are
caused) within a context carrying head trace sampling probability.
For example, we may record log events and metrics exemplars with
adjusted counts equal to the inverse of the current head trace
sampling probability when they are produced.

This technique allows translating spans and logs to metrics without
first locating their root span, a significant performance advantage
compared with first collecting and indexing root spans.

Several head sampling techniques are discussed in the following
sections and evaluated in terms of their ability to meet all of the
following criteria:

- Reduces Tracer overhead
- Produces complete traces
- Spans are countable.

#### Head sampling for traces

Details about Sampler implementations that meet
the requirements stated above.

##### `Parent` Sampler

The `Parent` Sampler ensures complete traces, provided all spans are
successfully recorded.  A downside of `Parent` sampling is that it
takes away control over Tracer overhead from non-roots in the trace.
To support real-time span-to-metrics applications, this Sampler
requires propagating the sampling probability or adjusted count of
the context in effect when starting child spans.  This is expanded
upon in [OTEP 168 (WIP)](https://github.com/open-telemetry/oteps/pull/168).

When propagating head sampling probability, spans recorded by the
`Parent` sampler MAY encode the adjusted count in the corresponding
`SpanData` using a non-descriptive Span attribute named
`sampler.adjusted_count`.

##### `TraceIDRatio` Sampler

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
Sampler probability across the whole trace.  Techniques have been
developed for [analysis of partial traces that are compatible with
TraceID ratio sampling](https://arxiv.org/pdf/2107.07703.pdf).

The `TraceIDRatio` Sampler has another difficulty with testing for
completeness.  It is impossible to know whether there are missing leaf
spans in a trace without using external information.  One approach,
[lost in the transition from OpenCensus to OpenTelemetry is to count
the number of children of each
span](https://github.com/open-telemetry/opentelemetry-specification/issues/355).

Lacking the number of expected children, we require a way to know the
minimum Sampler probability across traces to ensure they are complete.

To count TraceIDRatio-sampled spans, each span MAY encode its adjusted
count in the corresponding `SpanData` using a non-descriptive Span
attribute named `sampler.adjusted_count`.

##### Dapper's "Inflationary" Sampler

Google's [Dapper](https://research.google/pubs/pub36356/) tracing
system describes the use of sampling to control the cost of trace
collection at scale.  Dapper's early Sampler algorithm, referred to as
an "inflationary" approach (although not published in the paper), is
reproduced here.

This kind of Sampler allows non-root spans in a trace to raise the
probability of tracing, using a conditional probability formula shown
below.  Traces produced in this way are complete sub-trees, not
necessarily complete.  This technique is successful especially in
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
  is in effect, independent of whether the parent context was sampled
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
trace with adjusted count `1/I`.

## Proposed specification text

The following text will be added to the semantic conventions for
recording the Sampler name and adjusted count (if known) as
OpenTelemetry Span attributes.

### Semantic conventions for Sampled spans (Proposed text)

This document defines conventions for counting system-wide span events
using sampled spans.  These conventions support accurate counting of
system-wide events using only the fraction of spans that were
collected in a probability sampling scheme.  With these conventions,
consumers of OTLP Span data are able to compute approximate metrics
about the system using only the sample data that was collected, thus
we refer to these conventions as supporting Span-to-Metrics pipelines.

The _sampling rate_, also known as _inclusion probability_, is the
probability that a Span is included in the Sample being collected.
Sampling rate is conveyed in a form known as _adjusted count_, which
tells the receiver how many events in the population are represented
by the individual Span as a result of sampling.

The adjusted count of a span is defined as follows:

- Adjusted count equals zero when inclusion probability equals zero
- Adjusted count equals the mathematical inverse (i.e., reciprocal) of inclusion probability when inclusion probability is non-zero.

Consumers of spans carrying an adjusted count attribute are able to
use the adjusted count of the span to increment a counter of matching
spans.

#### Probability Sampling Attributes (Proposed text)

The `sampler.adjusted_count` attribute, when set, MUST equal an
unbiased estimate of the number of representative spans in the
population of spans in the system.

The _exported count_ associated with a span is defined as either 1 or
0, depending on whether the span is exported and thus counted.  The
exported count is 1 if the span is exported (because it will be
counted) and 0 if the span is not exported.

To avoid recording redundent information, both the `sampler.name` and
`sampler.adjusted_count` attributes MAY be omitted when the counting
algorithm given below produces a correct result.

There are scenarios where the adjusted count is unknown, such as when
using the `ParentBased` Sampler with a W3C version-0 `traceparent`
context.

The `sampler.adjusted_count` SHOULD be omitted when its value is 1 or
unknown.  The adjusted count can be safely omitted when it is 1
because that is exactly the number of events associated the span in
that case.  In case the adjusted count is unknown, `sampler.name` MUST
be set with a Sampler name to signify an unknown adjusted count.

The presence of a `sampler.name` without a `sampler.adjusted_count`
SHOULD be taken as a signal that a span-to-metrics pipeline cannot be
established without external information.  Otherwise, `sampler.name`
SHOULD be set when the adjusted count is not equal to 1.

The algorithm for spans-to-metrics is as follows:

```
// Calls `span_to_metrics(span, C)` for the effective count `C` of 
// every `span` received.
for _, span := <-spans_received {
  if count, has := span.attributes['sampler.adjusted_count']; has {
    span_to_metrics(span, count)
  } else if name, has := span.attributes['sampler.name]; has {
    log.Error("span requires trace assembly before counting")
  } else {
    span_to_metrics(span, 1)
  }
}
```

To summarize, these two attributes convey information about the
Sampler that recorded a span.

| Attribute | Type | Description | Examples | Required |
|---------- | ---- | ----------- | -------- | -------- |
| `sampler.adjusted_count` | number | Effective count of the span. | 10       | Yes, when adjusted count is not equal to 1 |
| `sampler.name`           | string | The name of the Sampler.     | `Parent` | Yes, when adjusted count is not equal to the exported count |

For the built-in samplers, the following names are specified:

| Built-in Sampler | Sets `sampler.adjusted_count`? | `sampler.name` | Notes                     |
| ---------------- | ------------------------------ | -------------- | ------------------------- |
| AlwaysOn         | No         | Not set        | Adjusted count equals exported count          |
| AlwaysOff        | Don't care | Don't care     | Exported count is zero, spans are not counted |
| ParentBased      | Maybe      | `Parent`       | Adjusted count is known when it is propagated |
| TraceIDRatio     | Yes        | `TraceIDRatio` | Adjusted count is known                       |

## Recommended reading

[Sampling, 3rd Edition, by Steven
K. Thompson](https://www.wiley.com/en-us/Sampling%2C+3rd+Edition-p-9780470402313).

[A Generalization of Sampling Without Replacement From a Finite Universe](https://www.jstor.org/stable/2280784), JSTOR (1952)

[Performance Is A Shape. Cost Is A Number: Sampling](https://docs.lightstep.com/otel/performance-is-a-shape-cost-is-a-number-sampling), 2020 blog post, Joshua MacDonald

[Priority sampling for estimation of arbitrary subset sums](https://dl.acm.org/doi/abs/10.1145/1314690.1314696)

[Stream sampling for variance-optimal estimation of subset sums](https://arxiv.org/abs/0803.0473).

[Estimation from Partially Sampled Distributed Traces](https://arxiv.org/pdf/2107.07703.pdf), 2021 Dynatrace Research report, Otmar Ertl

## Acknowledgements

Thanks to [Neena Dugar](https://github.com/neena) and [Alex
Kehlenbeck](https://github.com/akehlenbeck) for their help
reconstructing the Dapper Sampler algorithm.
