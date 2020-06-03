# Integrate Exemplars with Metrics

This OTEP adds exemplar support to aggregations defined in the Metrics SDK.

## Definition

Exemplars are example data points for aggregated data. They provide specific context to otherwise general aggregations. For histogram-type metrics, exemplars are points associated with each bucket in the histogram giving an example of what was aggregated into the bucket. Exemplars are augmented beyond just measurements with references to the sampled trace where the measurement was recorded and labels that were attached to the measurement.

## Motivation

Defining exemplar behaviour for aggregations allows OpenTelemetry to support exemplars in Google Cloud Monitoring.

Exemplars provide a link between metrics and traces. Consider a user using a Histogram aggregation to track response latencies over time for a high QPS server. The histogram is composed of buckets based on the speed of the request, for example, "there were 55 requests that took 400-500 milliseconds". The user wants to troubleshoot slow requests, so they would need to find a trace where the latency was high. With exemplars, the user is able to get an exemplar trace from a high latency bucket, an exemplar trace from a low latency bucket, and compare them to figure out the reason for the high latency.

Exemplars are meaningful for all aggregations where relevant traces can provide more context to the aggregation, as well as when exemplars can display specific information not otherwise shown in the aggregation (for example, the full set of labels where they otherwise might be aggregated away).

## Internal details

An exemplar is defined as:

```
message Exemplar {
  // Numerical value of the measurement that was recorded. Only one of these two fields is
  // used for the data, depending on its type
  double double_value = 0;
  int64 int64_value = 1;
  
  // Exact time that the measurement was recorded
  fixed64 time_unix_nano = 2;

  // 'label:value' map of all labels that were provided by the user recording the measurement
  repeated opentelemetry.proto.common.v1.StringKeyValue labels = 3;

  // Span ID of the current trace [Optional]
  string span_id = 4;

  // Trace ID of the current trace [Optional]
  string trace_id = 5;
}
```

Exemplar collection should be enabled through an optional parameter, and when not enabled, there should be no collection/logic performed related to exemplars. This is to ensure that when necessary, aggregations are as high performance as possible.

[#347](https://github.com/open-telemetry/opentelemetry-specification/pull/347) describes a set of standard aggregations in the metrics SDK. Here we describe how exemplars could be implemented for each aggregation.

### Exemplar behaviour for standard aggregations

#### HistogramAggregator

Every bucket in the HistogramAggregator MUST (when enabled) maintain a list of exemplars whose values are within the boundaries of the bucket. Implementations should attempt to retain at least one exemplar per bucket, with a preference for exemplars with a sampled trace context and exemplars that were recorded later in the time period. They should also not retain an unbounded number of exemplars.

#### Sketch

A Sketch aggregation should maintain a list of exemplars whose values are spaced out across the distribution. There is no specific number of exemplars that should be retained (although the amount should not be unbounded), but the implementation should pick exemplars that represent as much of the distribution as possible. Preference should be given to exemplars with a sampled trace context. (Specific details not defined, see open questions.)

#### Gauge

Most (if not all) Gauges operate asynchronously and do not ever interact with traces. Since the value of a Gauge is the last measurement (essentially the other parts of an exemplar), exemplars are not worth implementing for Gauge.

#### Exact

The Exact aggregator does not aggregate measurements. If exemplars are enabled, implementations may attach a separate exemplar to each measurement in an exact aggregation including the trace context and full set of labels.

Exemplars will always be retrieved from aggregations (by the exporter) as a list of Exemplar objects.

## Trade-offs and mitigations

Performance (in terms of memory usage and to some extent time complexity) is the main concern of implementing exemplars. However, by making recording exemplars optional, there should be minimal overhead when exemplars are not enabled.

## Prior art and alternatives

Exemplars are implemented in [OpenCensus](https://github.com/census-instrumentation/opencensus-specs/blob/master/stats/Exemplars.md#exemplars), but only for HistogramAggregator. This OTEP is largely a port from the OpenCensus definition of exemplars, but it also adds exemplar support to other aggregators.

[Cloud monitoring API doc for exemplars](https://cloud.google.com/monitoring/api/ref_v3/rpc/google.api#google.api.Distribution.Exemplar)

## Open questions

- Exemplars usually refer to a span in a sampled trace. While using the collector to perform tail-sampling, the sampling decision may be deferred until after the metric would be exported. How do we create exemplars in this case?

- We don’t have a strong grasp on how the sketch aggregator works in terms of implementation - so we don’t have enough information to design how exemplars should work properly.

- The spec doesn't yet define a standard set of aggregations, just default aggregations for standard metric instruments. Since exemplars are always attached to particular aggregations, it's impossible to fully specify the behavior of exemplars.

### Which aggregations should include exemplars?

There are other aggregations that can benefit from exemplars, but they do not have well defined exemplar implementations and they are not supported by any known exporter. Should these be included in the OTEP or should they be left out?:

#### Counter

Exemplars give value to counter aggregations by tying metric and trace data together. When enabled, the aggregator will retain a small bounded list of exemplars at each checkpoint, containing at least the minimum and maximum value measurements whose trace context was sampled. Measurements should only be retained if there is a sampled trace context when the measurement was recorded.

#### MinMaxSumCount

The aggregator should maintain a list of at least two exemplars (when enabled), one near the maximum value and one near the minimum value. Preference should be given to exemplars with sampled traces, and if those are not available then the actual min and max values should be used.
