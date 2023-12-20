# OpenTelemetry Export-pipeline metrics

Propose a uniform standard for OpenTelemetry SDK and Collector
export-pipeline metrics with three standard levels of detail.

## Motivation

OpenTelemetry has pending requests to standardize conventions for the
metrics emitted by SDKs. At the same time, the OpenTelemetry Collector
is becoming a stable and critical part of the ecosystem, and it has
different semantic conventions.  Here we attempt to unify them.

## Explanation

The OpenTelemetry Collector's pipeline metrics were derived from the
OpenCensus collector.

### Collector metrics

The core collector formerly contained a package named `obsreport`,
which has a uniform interface dedicated to each of its components.
This package has been migrated into the commonly-used helper classes
known as `receiverhelper`, `processorhelper`, and `exporterhelper.`

Obsreport is responsible for giving collector metrics a uniform
appearance.  Metric names were created using OpenCensus style, which
uses a `/` character to indicate hierarchy and a `.` to separate the
operative verb and noun.  This library creates metrics named, in
general, `{component-type}/{verb}.{noun}`, with component types
`receiver`, `processor`, and, `exporter`, and with signal-specific
nouns `spans`, `metric_points` and `logs` corresponding with the unit
of information for the tracing, metrics, and logs signals,
respectively.

Earlier adopters of the Collector would use Prometheus to read these
metrics, which does not accept `/` or `.`.  The Prometheus integration
would add a `otelcol_` prefix and replace the invalid characters with
`_`.  The same metric in the example above would appear named
`otelcol_receiver_accepted_spans`.

#### Collector: Obsreport receiver metrics

For receivers, the obsreport library counts items in two ways:

1. Receiver `accepted` items.  Items that are received and
   successfully consumed by the pipeline.
2. Receiver `refused` items.  Items that are received and fail to be
   consumed by the pipeline.

Items are exclusively counted in one of these counts.  The lifetime
average failure rate of the receiver com is defined as
`refused / (accepted + refused)`.

#### Collector: Obsreport processor metrics

For processors, the obsreport library counts items in three ways:

1. Processor `accepted` items.  Defined as the number of items that are passed to the next component and return successfully.
2. Processor `dropped` items.  This is a counter of items that are
   deliberately excluded from the output, which will be counted as accepted by the preceding pipeline component but were not transmitted.
3. Processor `refused` items.  Defined as the number of items that are passed to the next component and fail.

Items are exclusively counted in one of these counts.  The average drop rate
can be defined as `dropped / (accepted + dropped + refused)`

#### Collector: Obsreport exporter metrics

The `obsreport_exporter` interface counts spans in two ways:

1. Exporter `sent` items.  Items that are sent and succeed.
2. Receiver `send_failed` items.  Items that are sent and fail.

Items are exclusively counted in one of these counts.  The average
failure rate is defined as `send_failed / (sent + send_failed)`.

The exporterhelper package takes on many aspects of processor
behavior, including the ability to drop when a queue is full.  It uses
a separate counter for these items, known as `enqueue_failed`.

### Jaeger trace SDK metrics

Jaeger SDKs expose metrics on the "Reporter", which includes
"Success", "Failure", "Dropped" counters describing the pipeline.  See
[here](https://github.com/jaegertracing/jaeger-client-go/blob/8d8e8fcfd04de42b8482476abac6a902fca47c18/metrics.go#L22-L106).

Jaeger SDK metrics are equivalent to the three metrics produced by
OpenTelemetry Collector processor components.

### Analysis

As we can see by the examples documented above, it is a standard
practice to monitor a telemetry pipeline using three counters to count
successful, failed, and dropped items.

A central aspect of the proposed specification is to use a single
metric instrument with exclusive attribute values, as compared with
the use of separate metric instruments.

By specifying attribute dimensions for the resulting single
instrument, users can configure the level of detail and the number of
timeseries needed to convey the information they want to monitor.

#### Meaning of "dropped" telemetry

The term "Dropped" in pipeline monitoring usually refers to telemetry
that was intentionally not transmitted.  A survey of existing pipeline
components shows the following uses.

In the SDK, the standard OpenTelemetry BatchSpanProcessor will drop
spans that cannot be admitted into its queue.  These cases are
intentional, to protect the application and downstream pipeline, but
they should be considered failure because they were sampled, and not
collecting them in general will lead to trace incompleteness.

In a Collector pipeline, there are formal and informal uses:

- A sampling processor, for example, may drop spans because it was
  instructed to (e.g., due to an attribute like `sampling.priority=0`).
  In this case, drops are considered success.
- The memorylimiter processor, for example, may "drop" spans because
  it was instructed to (e.g., when it is above a hard limit).
  However, when it does this, it returns an error counts the item as
  `refused`, contradicting the documentation of that metric instrument:
  
> "Number of spans that were rejected by the next component in the pipeline."

There is already an inconsistency.  By counting its own failures as
refused, we should expect that the next component in the pipeline
handled the data.  This is a failure case drop, one where the next
component in the pipeline does not handle the item:

> "Number of spans that were dropped."

The memory limiter source code actually has a comment on this topic,

```
// TODO: actually to be 100% sure that this is "refused" and not "dropped"
// 	it is necessary to check the pipeline to see if this is directly connected
// 	to a receiver (ie.: a receiver is on the call stack). For now it
// 	assumes that the pipeline is properly configured and a receiver is on the
// 	callstack and that the receiver will correctly retry the refused data again.
```

which adds to the confusion -- it is not standard practice for
receivers to retry in the OpenTelemetry collector, that is the duty of
exporters in our current practice.  So, the memory limiter component,
to be consistent, should count "failure drops" to indicate that the
next stage of the pipeline did not see the data.

There is still another use of "dropped" in the collector, similar to
the memory limiter example and the SDK use-case, where "dropped" is a
case of failure.  In the `exporterhelper` module, the term dropped is
used in log messages to describe data that was tried at least once and
will not be retried, which matches the processor's definition of
`refused` in the sense that data was submitted to the next component
in the pipeline and failed and does not match the processor's
definition `dropped`.

As the exporter helper is not part of a processor framework, it does
not have a conventional way to count dropped items.  When the
queue-sender is enabled and the queue is full, items are dropped in
the standard sense, but they are counted using an `enqueue_failed`
metric.

## Proposed semantic conventions

### Use of a single metric name

The use of a single metric name is less confusing than the use of
multiple metric names, because the user has to know only a single name
to writing useful queries.  Users working with existing collector and
SDK pipeline monitoring metrics have to remember at least three metric
names and explicitly join them custom metric queries.  For example, to
calculate loss rate for an SDK using traditional pipeline metrics,

```
LossRate_MultipleMetrics = (dropped + failed) / (dropped + failed + success)
```

On the other hand, with a uniform boolean attribute indicating success
or failure the resulting query is simpler.

```
LossRate_SingleMetric = items{success=false} / items{success=*}
```

In a typical metric query engine, after the user has entered the one
metric name, attribute values will be automatically surfaced in the
user interface, allowing them to make sense of the data and
interactively build useful queries.  On the other hand, the user who
has to query multiple metrics has to enter each metric name
explicitly without help from the user interface.

The proposed metric instrument would be named distinctly depending on
whether it is a collector or an SDK, to prevent accidental aggregation
of these timeseries.  The specified counter names:

- `otelsdk.producer.items`: count of successful and failed items of
  telemetry produced by signal type by an OpenTelemetry SDK.
- `otelcol.receiver.items`: 
- `otelcol.processor.items`: 
- `otelcol.exporter.items`: 

### Recommended conventional attributes

- `otel.success` (boolean): This is true or false depending on whether the
  component considers the outcome a success or a failure.
- `otel.outcome` (string): This describes the outcome in a more specific
  way than `otel.success`, with recommended values specified below.
- `otel.signal` (string): This is the name of the signal (e.g., "logs",
  "metrics", "traces")
- `otel.name` (string): Name of the component in a pipeline.
- `otel.pipeline` (string): Name of the pipeline in a collector.

#### Collector

TODO: Why we count only drops for pipeline segments, but not SDKs.

## Metrics SDK special considerations

We expect that Metrics SDKs will be used to generate
pipeline-monitoring metrics reporting about themselves.

As stated above, SDKs SHOULD support configuring an alternate Meter
Provider for pipeline-monitoring metrics.  When the global Meter
Provider is used, the Metrics SDK's pipeline will receive its own
pipeline-monitoring metrics.  When a custom Meter Provider is used, a
secondary pipeline will receive the pipeline monitoring metrics, in
which case the secondary pipeline may also self-report for itself.

## Trade-offs and mitigations

The use of three-levels of metric detail may seem like more freedom
than necessary.  Implementors are expected to take advantage of Metric
View configuration in the Metrics SDK for configuring opt-out of
standard attributes (i.e., to remove `otel.signal`, `otel.name`, or
`otel.signal`).  For opt-in attributes (i.e., to configure no
`otel.reason` or `otel.scope` attribute), implementors MAY choose to
enable additional attributes only when configured.

## Prior art and alternatives

Prior work in (this PR)[https://github.com/open-telemetry/semantic-conventions/pull/184].

Issues:
- [Determine how to report dropped metrics](https://github.com/open-telemetry/opentelemetry-specification/issues/1655)
- [How should OpenTelemetry-internal metrics be exposed?](https://github.com/open-telemetry/opentelemetry-specification/issues/959)
- [OTLP Exporter must send client side metrics](https://github.com/open-telemetry/opentelemetry-specification/issues/791)
- [Making Tracing SDK metrics aware](https://github.com/open-telemetry/opentelemetry-specification/issues/381)
