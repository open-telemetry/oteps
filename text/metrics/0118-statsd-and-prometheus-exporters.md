# Specify standard treatment of OpenTelemetry aggregations in Prometheus and Statsd exporters

Specify behavior for Prometheus and Statsd exporters using standard
OpenTelemetry aggregations.

## Motivation

OpenTelemetry has specified a set of builtin Aggregators that can be
configured for use with metric instruments.  The specification also
defines the default Aggregator that will be applied when an
Aggregation method is not otherwise configured.  For some of the
Aggregators, there are multiple potential translations into existing
OSS systems.  For example, a [MinMaxLastSumCount
Aggregator](https://github.com/open-telemetry/oteps/pull/117) can be
exposed in Prometheus as a Summary or as a Gauge, with the Gauge
format preferred.

This proposal specifies how to map OpenTelemetry Aggregators into
these OSS exposition formats, in order to support migration from
Prometheus and Statsd APIs without changing metrics protocols.
Specifying the data type correspondence is a necessary prerequisite
for migration from Prometheus and Statsd instruments to OpenTelemetry
instruments.  The instrument migration is included to complete this
proposal.

As a requirement, this proposal promises that the recommended mapping
from Prometheus and Statsd instrument to OpenTelemetry instrument,
paired with the default Aggregator and then mapped back into
Prometheus or Statsd, shall produce the corresponding value type of
the original Prometheus or Statsd instrument.  This is also a
requirement when the data is forwarded over OTLP.  An exception to
this rule is associated with Prometheus Summary values, which are
discouraged from use.  Note that when the default mapping does not
work as intended (e.g., for Prometheus Summary values), the Metric
Views API or a configurable Metrics SDK will be needed to reconfigure
those OpenTelemetry instruments.

## Background

An Aggregator is an implementation of some logic to compute an
Aggregation, which is an exact or approximate summarization of a
series of metric events.  Exporters translate Aggregation values to
into an exposition format, so the choice of Aggregator decides which
exposition formats are possible by the time data reaches an Exporter.

The standard [OpenTelemetry Aggregators (TODO: WIP
document)](https://github.com/open-telemetry/opentelemetry-specification/pull/347),
listed below, each support one or more Aggregations.

| OpenTelemetry Aggregator | Aggregations Supported |
| -- | -- |
| Sum | Sum |
| LastValue | LastValue |
| MinMaxLastSumCount | Sum, Count, Min, Max, LastValue |
| Histogram | Sum, Count, Histogram |
| Sketch | Sum, Count, Quantile |
| Exact | Sum, Count, Quantile, Points |

The standard OpenTelemetry Aggregators are required to be mergeable,
meaning that two or more Aggregators can be combined using a `Merge()`
operation to form a single summarization of the combined data.  This
allows the metrics processor to generate either a _Cumulative_ value
(over all intervals) or a _Delta_ value (over one interval) of the
Aggregation on behalf of the Exporter.

The OTLP protocol has been designed as the standard exposition format
for OpenTelemetry libraries to forward data to the OpenTelemetry
collector, which is designed to process and re-export metric data.
When OpenTelemetry metric data is exposed through an Prometheus or
Statsd exporter, is it important that they produce the same result
whether data was exported directly or whether OTLP was used to forward
data to a collector.

OpenTelemetry metric instruments are classified in several ways:

- _Synchronous_: Synchronous instruments are called by the user (many times per interval), potentially have tracing context; asynchronous instruments are used through callbacks (once per interval).
- _Adding_ vs. _Grouping_: An adding instrument captures the sum of a number of measurements, while a grouping instrument captures a number of individual measurements.  Grouping instruments are more expensive by nature.
- _Monotonic_: Adding instruments can be monotonic, indicating that the sum they express logically cannot decrease.
- _Precomputed-Sum_: Asynchronous adding instruments observe a sum directly, instead of a series of changes in the sum.

These properties will help understand how to map OpenTelemetry
Aggregators into Prometheus and Statsd metric data.

| Name | Synchronous | Adding or Grouping | Monotonic | Precomputed-Sum | Default Aggregator |	
| ---- | ----------- | -------- | --------- | ---- | --- |
| Counter           | Yes | Adding   | Yes | No  | Sum |
| UpDownCounter     | Yes | Adding   | No  | No  | Sum |
| ValueRecorder     | Yes | Grouping | n/a | n/a | MinMaxLastSumCount |
| SumObserver       | No  | Adding   | Yes | Yes | Sum |
| UpDownSumObserver | No  | Adding   | No  | Yes | Sum |
| ValueObserver     | No  | Grouping | n/a | n/a | MinMaxLastSumCount |

Note that the Precomputed-Sum property places some constraints on the
combination of Aggregator and Exporter.  To compute _Delta_
aggregations of Precomputed-Sum instruments requires an aggregation
that supports subtraction (which MUST include Sum and SHOULD include
Histogram).

### Prometheus

#### Prometheus Instruments

Prometheus Counter instruments are semantically identical to
OpenTelemetry Counter instruments, including the Mnotonic property.
They are exposed as a Cumulative Sum aggregation.

Prometheus Gauge instruments have a number of uses, depending on
whether they are used as an adding or as a grouping instrument.  They
are exposed as a single data point equal to the the last value that
was set.  Because Prometheus clients are stateful, Gauges support both
`Set()` and `Add()` methods.  Generally, Prometheus Gauges used to
`Add()` map into OpenTelemetry UpDownCounter instruments, while
Prometheus Gauges used to `Set()` map into OpenTelemetry ValueRecorder
instruments.

Prometheus Histogram instruments are exposed as Cumulative
aggregations, defined by a number of buckets containing counts from
the start of the


@@@


## Trade-offs and mitigations

What are some (known!) drawbacks? What are some ways that they might be mitigated?

Note that mitigations do not need to be complete *solutions*, and that they do not need to be accomplished directly through your proposal. A suggested mitigation may even warrant its own OTEP!

## Prior art and alternatives

What are some prior and/or alternative approaches? For instance, is there a corresponding feature in OpenTracing or OpenCensus? What are some ideas that you have rejected?

## Open questions

What are some questions that you know aren't resolved yet by the OTEP? These may be questions that could be answered through further discussion, implementation experiments, or anything else that the future may bring.

## Future possibilities

What are some future changes that this proposal would enable?
