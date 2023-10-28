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

#### Station integrity principle

The [OpenTelemetry library guidelines (point
4)](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/library-guidelines.md#requirements)
describes a separation of protocol-dependent ("receivers",
"exporters") and protocol-independent ("processors") parts.  Here we
refer to this combination of parts as a station, belonging to a single
failure domain, because:

1. Logic internal to a station is presumably non-lossy.  Dropping
   within a station is presumed to be intentional, as distinct from
   the case of failures.
2. Under normal circumstances, we expect all-or-none failures for
   individual stations.

These qualities of the station will allow us to vary the level of
detail between basic and normal-level monitoring without information
loss.

#### Pipeline stage-name uniqueness

The Pipeline Stage Name Uniqueness requirement developed here avoids
over-counting in an export pipeline by ensuring that no single metric
name counts items are more than one distinct component.  This rule
prevents counting items of telemetry sent by SDKs and Collectors in
the same metric; it also prevents counting items of telemetry sent
through a multi-tier arrangement of Collectors.

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

#### Pipeline conservation principle

The station integrity principle leads to the axiom: items that are
transmitted, leading to success or failure, cannot have been dropped.

The second principle developed here establishes that items going in to
a station either succeed or fail.


---00-

From a technical perspective, how do you propose accomplishing the
proposal? In particular, please explain:

* How the change would impact and interact with existing functionality
* Likely error modes (and how to handle them)
* Corner cases (and how to handle them)

While you do not need to prescribe a particular implementation - indeed, OTEPs should be about **behaviour**, not implementation! - it may be useful to provide at least one suggestion as to how the proposal *could* be implemented. This helps reassure reviewers that implementation is at least possible, and often helps them inspire them to think more deeply about trade-offs, alternatives, etc.

## Trade-offs and mitigations

What are some (known!) drawbacks? What are some ways that they might be mitigated?

Note that mitigations do not need to be complete *solutions*, and that they do not need to be accomplished directly through your proposal. A suggested mitigation may even warrant its own OTEP!

## Prior art and alternatives

What are some prior and/or alternative approaches? For instance, is there a corresponding feature in OpenTracing or OpenCensus? What are some ideas that you have rejected?

## Open questions

What are some questions that you know aren't resolved yet by the OTEP? These may be questions that could be answered through further discussion, implementation experiments, or anything else that the future may bring.

## Future possibilities

What are some future changes that this proposal would enable?







semantic-conventions/
