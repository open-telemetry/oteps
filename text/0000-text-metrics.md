# Text-valued metric data

Expand the metric semantics to include textual metric data, intended for use
with enumerations and constant-ish data sources. Define corresponding value
recorders and observers that can report text to an exporter.

## Motivation

OpenTelemetry currently supports numeric metrics (integer and floating-point),
which are a subset of instrumentation that operators of a mature distributed
system need to understand its current state. A significant missing feature is
support for text-valued metrics, which are used for ingesting system and process
properties into the monitoring system.

Text-valued metrics are primarily used for properties that do not change for
the lifetime of a process, or which change rarely. Examples:

* Properties of the build environment, such as `/build/revision` and
  `/build/branch`. These are computed at build time and baked into the binary
  as constant values.

* Properties of the current host system, such as `/sys/kernel/version` and
  `/sys/kernel/cmdline`. These are queried at process startup and will usually
  not change until the next reboot.

* Values of specific command-line flags, such as `/telemetry/log/destination`,
  which might be mutable at runtime (e.g. via a `/flagz` z-page) but are
  typically set only at process startup.

Supporting these metric types in OpenTracing would allow better integration
between OTel-instrumented libraries and monitoring systems that support text
as metrics data.

## Internal details

The Metrics API specification must be expanded to describe the semantics and
limitations of `TextValueRecorder` and `TextValueObserver` interfaces. This
should be fairly straightforward as the set of operations that can be performed
on text is limited -- for example there is no aggregation or summation.

The language-specific API specifications should have corresponding plumbing for
the new metric types. In some cases this may require renaming or restructing
existing interfaces, such as the Go `metric.SyncImpl` interface that only
supports numeric data.

Metric implementation providers that do not support text data should return an
error when attempting to construct a text metric instrument. This will apply to
any provider based on the StatsD protocol, since the StatsD datagram format only
supports numbers. Providers that do support text values, such as SignalFx,
should support the new format to provide an end-to-end test of feasibility.

The OTLP protobuf field [`Metric::data`] should be expanded to include a new
case, `TextGauge text_gauge = N;`, which reports data points in the `string`
wire format.

[`Metric::data`]: https://github.com/open-telemetry/opentelemetry-proto/blob/v0.5.0/opentelemetry/proto/metrics/v1/metrics.proto#L138-L145

## Prior art

Text-valued metrics are supported by Google's internal monitoring system
"Monarch", which has been partially described to the public. The portion
relevant to OpenTelemetry is similar to the "metricfs" patchset for the Linux
kernel described at <https://lore.kernel.org/lkml/20200806001431.2072150-8-jwadams@google.com/T/>.

There are at least two commercial monitoring services providers, Google Cloud
and SignalFx, that support ingesting text-valued metrics data. The relevant
documentation is:

* https://cloud.google.com/monitoring/api/v3/kinds-and-types
* https://github.com/signalfx/signalfx-java/blob/1.0.5/signalfx-protoc/src/main/protobuf/signal_fx_protocol_buffers.proto#L14-L18

## Alternatives

It is possible to simulate text-valued metrics using a Gauge that is either
`0` or `1` and has a label `value: "the actual metric value"`, but this is
awkward and does not permit full use of monitoring systems that support text
as metric data.

Some parts of this proposal are arbitrary and can be freely changed to fit the
idioms of OpenTelemetry:

* I used "text" instead of "string" to avoid potential confusion in languages
  that use `string` types to represent arbitrary binary data, notably C++. If
  this confusion is not a concern then it would be fine to `s/text/string/g` in
  this proposal.

* If a provider does not support text-valued metrics, it might be preferable to
  have the metric creation silently fail (returning a no-op instrument) so that
  library authors can provide text metrics without asking users to upgrade their
  monitoring integrations.
