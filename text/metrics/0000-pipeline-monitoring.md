# OpenTelemetry Export-pipeline metrics

Propose a uniform standard for OpenTelemetry SDK and Collector
export-pipeline metrics with three standard levels of detail.

## Motivation

OpenTelemetry has pending requests to standardize the metrics emitted
by SDKs. At the same time, the OpenTelemetry Collector is becoming a
stable and critical part of the ecosystem, and it has different
semantic conventions.  Here we attempt to unify them.

## Explanation

The OpenTelemetry Collector's pipeline metrics were derived from the
OpenCensus collector.  There is no original source material explaining
the current state of metrics in the OTel collector.

### Collector metrics

The OpenTelemetry collector code base was audited for metrics usage
detail around the time of the v0.88.0 release.  Here is a summary of
the current state of the Collector regarding export-pipeline metrics.

The core collector formerly contained a package named `obsreport`,
which has a uniform interface dedicated to each of its components.
This package has been migrated into the commonly-used helper classes
known as `receiverhelper`, `processorhelper`, and `exporterhelper.`

Obsreport is responsible for giving collector metrics a uniform
appearance.  Metric names were created using OpenCensus style, which
uses a `/` character to indicate hierarchy and a `.` to separate the
operative verb and noun.  This library creates metrics named, in
general, `{component-type}/{verb}.{plural-noun}`, with component types
`receiver`, `processor`, and, `exporter`, and with signal-specific
nouns `spans`, `metric_points` and `logs` corresponding with the unit
of information for the tracing, metrics, and logs signals,
respectively.

Earlier adopters of the Collector would use Prometheus to read these
metrics, which does not accept `/` or `.`.  The Prometheus integration
would add a `otelcol_` prefix and replace the invalid characters with
`_`.  The same metric in the example above would appear named
`otelcol_receiver_accepted_spans`, for example.

#### Obsreport receiver

For receivers, the obsreport library counts items in two ways:

1. Receiver `accepted` items.  Items that are received and
   successfully consumed by the pipeline.
2. Receiver `refused` items.  Items that are received and fail to be
   consumed by the pipeline.

Items are exclusively counted in one of these counts.  The lifetime
average failure rate of the receiver com is defined as
`refused / (accepted + refused)`.

The `accepted` metric does not "lead" the `refused` metric, because
items are not counted until the end of the receiver operation.  A
single interface used by receiver components with `StartOp(...)`, and
`EndOp(..., numItems)` methods has both kinds of instrumentation.

Note there are a few well-known exporter and processor components that
return success unconditionally, preventing failures from passing back
to the producers.  With this behavior, the `refused` count becomes
unused.

#### Collector: Obsreport processor metrics

For processors, the obsreport library counts items in three ways:

1. Processor `accepted` items.  Defined as the number of items that are passed to the next component and return successfully.
2. Processor `dropped` items.  This is a counter of items that are
   deliberately excluded from the output, which will be counted as accepted by the preceding pipeline component but were not transmitted.
3. Processor `refused` items.  Defined as the number of items that are passed to the next component and fail.

Items are exclusively counted in one of these counts.  The average drop rate
can be defined as `dropped / (accepted + dropped + refused)`

Note there are a few well-known exporter and processor components that
return success unconditionally, preventing failures from passing back
to the producers.  With this behavior, the `refused` count becomes
unused.

#### Collector: Obsreport exporter metrics

The `obsreport_exporter` interface counts spans in two ways:

1. Exporter `sent` items.  Items that are sent and succeed.
2. Receiver `send_failed` items.  Items that are sent and fail.

Items are exclusively counted in one of these counts.  The average
failure rate is defined as `send_failed / (sent + send_failed)`.

### Jaeger trace SDK metrics

Jaeger SDKs expose metrics on the "Reporter", which includes
"Success", "Failure", "Dropped" (Counters), and "Queue_Length"
(UpDownCounter).  See [here](https://github.com/jaegertracing/jaeger-client-go/blob/8d8e8fcfd04de42b8482476abac6a902fca47c18/metrics.go#L22-L106).

### Analysis

#### SDK perspective

Considering the Jaeger SDK, data items are counted in exactly one of
three counters.  While unambiguous, the use of three
exclusively-counted metrics means that to compute any useful ratio
about SDK performance requires querying three tiemseries, and any pair
of these metrics tells an incomplete story.

There is no way to add varying level of detail, with three exclusive
counters.  If we wanted to omit any one of these timeseries, the other
two would have to change meaning.  While items that drop are in some
ways a failure, they are counted exclusively and so cannot be combined
with the failure count to be less detailed.

#### Collector perspective

Collector counters are exclusive.  Like for SDKs, items that enter a
processor are counted in one of three ways and to compute a meaningful
ratio requires all three timeseries.  If the processor is a sampler,
for example, the effective sampling rate is computed as
`(accepted+refused)/(accepted+refused+dropped)`.

While the collector defines and emits metrics sufficient for
monitoring the individual pipeline component--taken as a whole, there
is substantial redundancy in having so many exclusive counters.  For
example, when a collector pipeline features no processors, the
receiver's `refused` count is expected to equal the exporter's
`send_failed` count.

When there are several processors, it is primarily the number of
dropped items that we are interested in counting.  Whene there are
multiple sequential processors in a pipeline, however, counting the
total number of items at each stage in a multi-processor pipeline
leads to over-counting in aggregate.  For example, if you combine
`accepted` and `refused` for two adjacent processors, then remove the
metric attribute which distinguishes them, the resulting sum will be
twice the number of items processed by the pipeline.

The same logic applies to suggest that multiple sequential collectors
in a pipeline cannot use the same metric names, otherwise removal of
the which distinguishing metric attribute would cause over-counting of
the pipeline.

### Pipeline monitoring

The term _Stage_ is used to describe the a single component in an
export pipeline.

The term _Station_ is used to describe a location in the export
pipeline where the participating stages are part of the same logical
failure domain.  Typically each SDK or Collector is considered a
station.

#### Station integrity principles

The [OpenTelemetry library guidelines (point
4)](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/library-guidelines.md#requirements)
describes a separation of protocol-dependent ("receivers",
"exporters") and protocol-independent ("processors") parts.  We refer
to the combination of parts as a station.

The station concept is called out because within a station, we expect
that the station (software) acts responsibly by design, for the
integrity of the pipeline.  Stations allow data to enter a pipeline
only through receiver components.  Stations are never responsible for
dropping data, because only processor components drop data.  Stations
allow data to leave a pipeline only through exporter components.

Because of station integrity, we can make the following assertions:

1. Data that enters a pipeline is eventually exported or dropped.
2. No other outcomes are possible.

There is a potential for pipeline metrics to be redundant, as
described in these assertions.  In a pipeline with no fan-in or
fan-out, each stage processes as many items as the stage before it
did, minus the number of items dropped.

#### Pipeline stage-name uniqueness

The Pipeline Stage Name Uniqueness requirement developed here avoids
over-counting in an export pipeline by ensuring that no single metric
name counts items more than once in transit.  This rule prevents
counting items of telemetry sent by SDKs and Collectors in the same
metric; it also prevents counting items of telemetry sent through a
multi-tier arrangement of Collectors using the same metric.

In a standard deployment of OpenTelemetry, we expect one, two, or
three stations in a collection pipeline.  The names given to these
standard set of stations:

- `sdk`: an original source of new telemetry
- `agent`: a collector with operations "local" to the `sdk`
- `gateway`: a collector serving as a proxy to an external service.

This is not meant as an exclusive set of station names.  Users should
be given the ability to configure the station name used by particular
instances of the OpenTelemetry Collector.  It may even be desirable to
support configuring "sub-stations" within a larger pipeline, for
example when there are connectors in use; however, if so, the
collector must enforce that pipeline-stage names are unique within a
pipeline.

#### Basic detail through inclusive counters

By station integrity principles, we have several forms of detail that
may be omitted by users that want only the basic level of detail.

At a minimum, to establish information about _loss_ requires knowing
how much is received by the first station in the pipeline and how much
is exported by the last station in the pipeline.  From the total
received and the total exported, we can compute total pipeline loss.
Note that metrics about intermediate pipeline stations may be omitted,
since they implicitly factor into global pipeline loss.

For this proposal to succeed, it is necessary to use inclusive
counters as opposed to exclusive counters.  For the receivers, at a
basic level of detail, we only need to know the number of items
received (i.e., including items that succeed or fail or are dropped)
because those that fail or are dropped implicitly factor in to global
pipeline loss.

For processors, at a basic level of detail, it is not necessary to
count anything, since drops are implicit.  Any items not received by
the next stage in the pipeline must have dropped, therefore we can
infer drop counts from basic-detail metrics without any new counters.

For exporters, at a basic level of detail, the same argument applies.
Metrics describing an exporter in a pipeline ordinarily will match
those of receiver at the next stage in the pipeline, so they are not
needed at the basic level of detail, provided all receivers report
inclusive counters.

#### Pipeline failures optional

For a processor, dropping data is always on purpose, but dropped data
may be counted as failures or not, depending on the circumstances.
For an SDK batch processor, dropping data is considered failure.  For
a Collector sampler processor, dropping data is considered success.
It is up to the individual processor component whether to treat
dropped data as failures or successes.

We are also aware of a precedent for returning success at certain
stages in a pipeline, perhaps asynchronously, regardless of actual
success or not.  This is known to happen in the Collector's core
`exporterhelper` library, which provides a number of standard
features, including the ability to drop data when the queue is full.

Because failure is sometimes treated as success, it may be necessary
to monitor a point in the pipeline after failures are suppressed.

#### Pipeline signal type

OpenTelemetry currently has 3 signal types, but it may add more.
Instead of using the signal name in the metric names, we opt for a
general-purpose noun that usefully describes any signal.  

The signal-agnostic term used here is "items", referring to spans, log
records, and metric data points.  An attribute to distinguish the
`signal` will be used with name `traces`, `logs`, or `metrics`.

Users are expected to understand that the data item for traces is a
span, for logs is a record, and for metrics is a point.

#### Pipeline component name

Components are uniquely identified using a descriptive `name`
attribute which encompasses at least a short name describing the type
of component being used (e.g., `batch` for the SDK BatchSpanProcessor
or the Collector batch proessor).

When there is more than one component of a given type active in a
pipeline having the same `domain` and `signal` attributes, the `name`
should include additional information to disambiguate the multiple
instances using the syntax `<type>/<instance>`.  For example, if there
were two `batch` processors in a collection pipeline (e.g., one for
error spans and one for non-error spans) they might use the names
`batch/error` and `batch/noerror`.

#### Pipeline monitoring diagram

The relationship between items received, dropped, and exported is
shown in the following diagram.

![pipeline monitoring metrics](../images/otel-pipeline-monitoring.png)

### Proposed metrics semantic conventions

The proposed metric names are: 

`otel.{station}.received`: Inclusive count of items entering the pipeline at a station.
`otel.{station}.dropped`: Non-inclusive count of items dropped by a component in a pipeline.
`otel.{station}.exported`: Inclusive count of items exiting the pipeline at a station.

The behavior specified for SDKs and Collectors at each level of detail
is different, because SDKs do not receive items from a pipeline.

#### SDK default configuration

At the basic level of detail, SDKs are required to count spans
received by the export pipeline.  Only the `otel.sdk.received` metric
is required.  This includes items that succeed, fail, or are dropped.

At the normal level of detail, the `otel.sdk.received` metric gains an
additional boolean attribute, `success` indicating success or failure.
Also at the normal level of detail, the `otel.sdk.dropped` metric is
counted.

#### Collector default configuration

At the basic level of detail, Collectors of a given `{station}` name
are required to count `otel.{station}.received`,
`otel.{station}.dropped`, and `otel.{station}.exported`.

At the normal level of detailed, the `received` and `exported` metrics
gain an additional boolean attribute, `success` indicating success or
failure.

#### Detailed-level metrics configuration

There is one additional dimension that users may wish to opt-in to, in
order to gain information about failures at a particular pipeline
stage.  When detail-level metrics are requested, all three metric
instruments specified for pipeline monitoring gain an additional
`reason` attribute, with a short string explaining the failure.  

For example, with detailed-level metrics in use, the
`otel.{station}.received` and `otel.{station}.exported` counters will
include additional `reason` information (e.g., `timeout`,
`resource_exhausted`, `permission_denied`).

## Trade-offs and mitigations

While the use of three-levels of metric detail may seem excessive,
instrumentation authors are expected to implement the cardinality of
attributes specified here, with the use of Metric SDK View
configuration to remove unwanted attributes at runtime.  

This approach (i.e., configuration of views) can also be used in the
Collector, which is instrumented using the OTel-Go metrics SDK.

## Prior art and alternatives

Prior work in (this PR)[https://github.com/open-telemetry/semantic-conventions/pull/184].
