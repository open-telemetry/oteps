# Support for low-level metrics processor

**Status:** `proposed`

The low-level metrics processor API is specified to support processing metrics API events.

## Motivation

The OpenTelemetry [v1 metrics data model](https://github.com/open-telemetry/opentelemetry-proto/blob/master/opentelemetry/proto/metrics/v1/metrics.proto) supports exporting pre-aggregated metrics data and is a good choice in most cases.  We identify two cases where it is important to support exporting metric updates as individual events from the SDK.

In OpenTelemetry terminology, a "processor" API is part of the SDK that supports building export pipelines, a layered approach to building exporters.  The low-level metrics processor allows constructing bridges to other metrics systems, but it not a complete exporter by itself.

## Explanation

We consider several cases to justify this addition:

- To support third-party metrics libraries that do not match the v1 metrics data model or offer other advantages
- To support streaming SDKs which cannot aggregate inside the process.

We propose that OpenTelemetry libraries support a low-level metrics processor API to facilitate these cases.  Third-party metrics libraries and streaming metric exporters will take advantage of this support.

## Internal details

The metric processor is API is called once per metric event. The event structure includes:

- Definition: the metric instrument name, units, description, required keys, and options
- Value: the numerical value of the event (integer or floating point)
- Labels: pre-defined label set, resource scope, and current context.

The event is minimally processed, to avoid unnecessary work.  

The pre-defined label set are the set of labels associated with the metric `Handle`.  Resources are the set of labels associated with the `Meter`.  The remaining labels are associated with the current `Context`.

To construct the `Meter` SDK from a metrics processor hook, use a language-specific API method (e.g., `sdk.NewMetricsProcessor(callback)`, where callback accepts structured events as described above.

### Handling Observer gauges

Because observer gauges are non-contextual, there are no actual events for the metrics processor API.  The metrics processor API supports a method (e.g., `RecordObservers`) to periodically trigger events for `Observer` gauges, making frequency a configurable aspect of the metrics processor for observer gauge metric instruments.

## Related issues

This [CloudWatch metrics](https://github.com/open-telemetry/opentelemetry-go/issues/83) issue would be addressed with a metrics processor-based export pipeline.
