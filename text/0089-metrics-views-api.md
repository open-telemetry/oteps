# Metrics View API

This OTEP addresses the [Future Work: Configurable Aggregations / View API](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/api-metrics.md#future-work-configurable-aggregations--view-api) section of [api-metrics.md](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/api-metrics.md).
It proposes a specification for a _view API_ that allows application owners and administrators to configure aggregations for individual metric instruments.

## TL;DR

The potentially contentious changes this OTEP proposes include:

- Require views to be registered: don't record measurements from API for metric instruments that don't appear in any view.
- Drop support for LastValue aggregations, don't allow multiple Observer calls per export interval.
- Don't allow views to specify how "dropped" keys are aggregated, since we can safely ignore aggregation order for measurement instruments other than observers.

## Motivation

As of [opentelemetry-specification#430](https://github.com/open-telemetry/opentelemetry-specification/pull/430), the spec describes three different metric instrument types, and a default aggregation for each.
These default aggregations are consistent with the semantics of each metric instrument, but not with the various exposition formats required by APM backends.

For example, consider the [Prometheus histogram exposition format](https://prometheus.io/docs/concepts/metric_types/#histogram).
The default aggregation for the Measure metric instrument is _MinMaxSumCount_.
This aggregation computes summary statistics about (i.e. the _min_, _max_, _sum_, and _count_ of) the set of recorded values, but doesn't capture enough information to reconstruct histogram bucket counts.

If instead the Measure instrument used an [_Exact_ aggregation](https://github.com/open-telemetry/opentelemetry-specification/pull/347/files#diff-5b01bbf3430dde7fc5789b5919d03001R254-R259) -- which stores the sequence of measurements without aggregating them -- the Prometheus exporter could compute the histogram for each recording interval at export time, but at the cost of significantly increased memory usage.

A better solution would be to use a custom _Histogram_ aggregation, which could track bucket counts and compute summary statistics without preserving the original measurements.

As the spec is written now, all metrics are aggregated following the default aggregation for the metric instrument by which they were recorded.
By design, there is no option to change these defaults:

> [Users do not have a facility in the API to select the aggregation they want for particular instruments.](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/api-metrics.md#aggregations)

This limitation prevents users from:

- Converting aggregated data to vendor-specific exposition formats
- Configuring aggregation options, e.g. histogram bucket boundaries
- Specifying multiple aggregations for a single metric

In addition, it's often desirable to aggregate metrics with respect to a specific set of label keys to reduce memory use and export data size.

A _view API_ would allow users to use custom aggregation types, associate individual metrics with one or more aggregations, configure aggregation options, and specify the set of label keys to use to aggregate each metric.

## Explanation

Metric instruments are API objects that record measurements from application code.
These measurements are aggregated by _Aggregators_, which are SDK objects.
Aggregated measurement data are passed to _Exporters_, which convert these data into APM-specific exposition formats.

Each view describes a relationship between a metric instrument and an aggregation, and specifies the set of label keys to preserve in the aggregated data.

Views are tracked in a registry, which exporters may use to get the current aggregate values for all registered aggregations at export time.
Each registered view represents a metric that is meant to be exported.
As a corollary, the SDK can choose to drop measurements from any metric instrument that does not appear in a registered view.

## From measurement to APM backend

The goal of separating metric instrument types from aggregations is that it should be possible to switch to a different APM backend without changing application code.

The default SDK should include conventional aggregators, and exporters should convert aggregated data into the appropriate exposition format based on the aggregator type.
Some exporters will require custom aggregators, either because the built-in aggregators don't capture enough information to convert aggregated data into a particular exposition format, or to improve on the efficiency of the built-in aggregators.

APM vendors should typically include custom exporters and any custom aggregators they rely on in the same distribution.

One of the design goals of OpenTelemetry metrics is that it should work seamlessly with existing metrics systems like [StatsD](https://github.com/statsd/statsd/wiki) and [Prometheus/OpenMetrics,](https://github.com/prometheus/prometheus) and provide API instruments that are familiar to these systems' users.

To check that the view API proposed here satisfies this goal, it may be helpful to compare OpenTelemetry views -- which are comprised of a metric instrument and aggregation -- to other systems' metric types:

| System     | Metric Type                                                                                        | OT Metric Instrument | OT Aggregation    |
| ---------- | -------------------------------------------------------------------------------------------------- | -------------------- | ----------------  |
| OpenCensus | Count                                                                                              | Counter              | Sum               |
| OpenCensus | Mean                                                                                               | Counter              | Mean\*            |
| OpenCensus | Sum                                                                                                | Counter              | Sum               |
| OpenCensus | Distribution                                                                                       | Measure              | Histogram         |
| OpenCensus | LastValue                                                                                          | Observer             | LastValue         |
| Datadog    | [COUNT](https://docs.datadoghq.com/developers/metrics/types/?tab=count#metric-types)               | Counter              | Sum               |
| Datadog    | [RATE](https://docs.datadoghq.com/developers/metrics/types/?tab=rate#metric-types)                 | Counter              | Rate\*            |
| Datadog    | [GAUGE](https://docs.datadoghq.com/developers/metrics/types/?tab=gauge#metric-types)               | Observer             | LastValue         |
| Datadog    | [HISTOGRAM](https://docs.datadoghq.com/developers/metrics/types/?tab=histogram#metric-types)       | Measure              | Histogram         |
| Datadog    | [DISTRIBUTION](https://docs.datadoghq.com/developers/metrics/types/?tab=distribution#metric-types) | Measure              | Sketch\* or Exact |
| Prometheus | [Counter](https://prometheus.io/docs/concepts/metric_types/#counter)                               | Counter              | Sum               |
| Prometheus | [Gauge](https://prometheus.io/docs/concepts/metric_types/#gauge)                                   | Observer             | LastValue         |
| Prometheus | [Histogram](https://prometheus.io/docs/concepts/metric_types/#histogram)                           | Measure              | Histogram         |
| Prometheus | [Summary](https://prometheus.io/docs/concepts/metric_types/#summary)                               | Measure              | Sketch\* or Exact |

A well-designed view API should not require extensive configuration for typical use cases.
It should also be expressive enough to support custom aggregations and exposition formats without requiring APM vendors to write custom SDKs.
In short, it should make easy things easy and hard things possible.

## Internal details

A view is defined by:

- A metric instrument
- An aggregation
- An optional set of label keys to preserve
- An optional unique name
- An optional description

The aggregation may be configured with options specific to its type.
For example, a _Histogram_ aggregation may be configured with bucket boundaries, e.g. `{0, 1, 10, 100, 200, 1000, inf}`.
A _Sketch_ aggregation that estimates order statistics (i.e. quantiles), may be configured with a set of predetermined quantiles, e.g. `{.5, .95, .99, 1.00}`.

Each view has a unique name, which may be automatically determined from the metric instrument and aggregation.
Implementations should refuse to register two views with the same name.

Note that a view does not describe the type (e.g. `int`, `float`) or unit of measurement (e.g. "bytes", "milliseconds") of the metric to be exported.
The unit is determined by the metric instrument, and the aggregation and exporter may preserve or change the unit.
For example, a hypothetical _MinMaxSumMeanCount_ aggregation of an int-valued millisecond latency measure may be exported as five separate metrics to the APM backend:

- An int-valued _min_ metric with "ms" units
- An int-valued _max_ metric with "ms" units
- An int-valued _sum_ metric with "ms" units
- A **float**-valued _mean_ metric with "ms" units
- An **int** valued, unitless (i.e. unit "1") _count_ metric

This OTEP doesn't propose a particular API for Aggregators, just that the API is sufficient for exporters to get all this information, including:

- That the _min_, _max_, and _sum_ metrics preserve the type and unit of the underlying Measure
- That the _mean_ metric is float-valued, but preserves the underlying measure's unit
- That the _count_ metric is int-valued with unit "1", regardless of the underlying measure's unit

### Aggregating over time

An aggregation describes how multiple measurements captured via the same metric instrument in a single collection interval are combined.
_Aggregators_ are SDK objects, and the default SDK includes an Aggregator interface (see [opentelemetry-specification#347](https://github.com/open-telemetry/opentelemetry-specification/pull/347) for details).

Aggregations are assumed to be [mergeable](https://www.cs.utah.edu/~jeffp/papers/mergeSumm-MASSIVE11.pdf): aggregating a sequence of measurements should produce the same result as partitioning the sequence into subsequences, aggregating each subsequence, and combining the results.

Said differently: given an aggregation function `agg` and sequence of measurements `S`, there should exist a function `merge` such that:

```
agg(S) = merge(agg(S_1), agg(S_2), ..., agg(S_N))
```

For every partition `{S_1, ..., S_N}` of `S`.

For example, the _min_ aggregation is trivially mergeable:

```
min([1, 2, 3, 4, 5, 6]) = min([min([1]), min([2, 3]), min([4, 5, 6])])
```

but quantile aggregations, e.g. _p95_ and _p99_ are not.
Applications that export quantile metrics should use a mergeable aggregations such as [DDSketch](https://arxiv.org/abs/1908.10693), which estimates quantile values with bounded errors, or export raw measurements without aggregation and compute exact quantiles on the backend.

We require aggregations to be mergeable so that they produce the same results regardless of the collection interval, or the number of collection events per export interval.

### Aggregating across label keys

Every measurement is associated with a _LabelSet_, a set of key-value pairs that describes the environment in which the measurement was captured.
Label keys and values may be extracted from the [correlation context](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/api-correlationcontext.md) at the time of measurement.
They may also be set by the user [at the time of measurement](https://github.com/open-telemetry/opentelemetry-specification/blob/f4aeb318a5b77c9c39132a8cbc5d995e222d124f/specification/api-metrics-user.md#direct-instrument-calling-convention) or at the time at which [the metric instrument is bound](https://github.com/open-telemetry/opentelemetry-specification/blob/f4aeb318a5b77c9c39132a8cbc5d995e222d124f/specification/api-metrics-user.md#bound-instrument-calling-convention).

For each registered view, in each collection interval, measurements with the same labelset are aggregated together.

The view API should allow the user to specify the set of label keys to track for a given view.
Other label keys should be dropped before aggregation.
By default, if the user doesn't specify a set of label keys to track, the aggregator should record all label keys.

Users have three options to configure a view's label keys:

1. Record all label keys.
    This is the default option, and the behavior of the "ungrouped" integrator in [opentelemetry-specification#347](https://github.com/open-telemetry/opentelemetry-specification/pull/347).
2. Specify a set of label keys to track at view creation time, and drop other keys from the labelsets of recorded measurements before aggregating.
    This is the behavior of the "defaultkeys" integrator in [opentelemetry-specification#347](https://github.com/open-telemetry/opentelemetry-specification/pull/347).
3. Drop all label keys, and aggregate all measurements from the metric instrument together regardless of their labelsets.
    This is equivalent to using an empty list with option 2.

Because we don't require the user to specify the set of label keys up front, and because we don't prevent users from recording measurements with missing labels in the API, some label values may be undefined.
Aggregators should preserve undefined label values, and exporters may convert them as required by the backend.

For example, consider a Sum-aggregated Counter instrument that captures four consecutive measurements:

1. `{labelset: {'k1': 'v11'}, value: 1}`
2. `{labelset: {'k1': 'v12'}, value: 10}`
3. `{labelset: {'k1': 'v11', 'k2': 'v21'}, value: 100}`
4. `{labelset: {'k1': 'v12', 'k2': 'v22'}, value: 1000}`

And consider three different views, one that tracks label key _k1_, one that tracks label key _k2_, and one that tracks both.

After each measurement, the aggregator associated with each view has a different set of aggregated values:

| time        | k1                                  | k2  | value | agg([k1])                                   | agg([k2])                                                                | agg([k1, k2]) (default)                                                                                                                  |
| ----------- | ----------------------------------- | --- | ----- | ------------------------------------------- | ------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------- |
| 1           | v11                                 | -   | 1     | `({k1: v11}, 1)`                            | `({k2: undefined}, 1)`                                                   | `({k1: v11, k2: undefined}, 1)`                                                                                                          |
| 2           | v12                                 | -   | 10    | `({k1: v11}: 1)` <br> `({k1: v12}: 10)`     | `({k2: undefined}, 11)`                                                  | `({k1: v11, k2: undefined}, 1)` <br> `({k1: v12, k2: undefined}, 10)`                                                                    |
| 3           | v11                                 | v21 | 100   | `({k1: v11}: 101)` <br> `({k1: v12}: 10)`   | `({k2: undefined}, 11)` <br> `({k2: v21}, 100)`                          | `({k1: v11, k2: undefined}, 1)` <br> `({k1: v12, k2: undefined}, 10)` <br> `({k1: v11, k2: v21}, 100)`                                   |
| 4           | v12                                 | v22 | 1000  | `({k1: v11}: 101)` <br> `({k1: v12}: 1010)` | `({k2: undefined}, 11)` <br> `({k2: v21}, 100)` <br> `({k2: v22}, 1000)` | `({k1: v11, k2: undefined}, 1)` <br> `({k1: v12, k2: undefined}, 10)` <br> `({k1: v11, k2: v21}, 100)` <br> `({k1: v12, k2: v22}, 1000)` |

Aggregated values are mergeable with respect to labelsets as well as time, as long as the intermediate aggregations preserve the label keys to be included in the final result.
Note that it's possible to reconstruct the aggregated values at each step for _agg([k1])_ and _agg([k2])_ from _agg([k1, k2])_, but not vice versa.

### Special considerations for Observers

This OTEP describes aggregation as it applies to the Measure instrument and its refinements, e.g. Counter and Timer, which do not support LastValue aggregation (see [oteps#88](https://github.com/open-telemetry/oteps/pull/88) for details).

For these instruments, aggregating across label keys is straightforward: for a particular view, label keys that do not appear in the view definition can be safely dropped before aggregation.
In the example above, the aggregated values for _agg([k1])_/_agg([k2])_ would be the same at each timestamp if none of the measurements included values for _k2_/_k1_.

Observers, which may be aggregated by last recorded value, do not have this property.

Consider an application with two jobs, A and B.
Each job runs on a separate thread, and each thread runs on a single CPU core.
Two instances of the application run on a single host with two CPU cores.
Per-core CPU usage is reported in terms of cumulative seconds by the host.

In this example, the instrumentation should emit:

- The current per-core CPU usage (_agg[core]_)
- The total CPU usage of each job, across all cores and instances (_agg[job]_)

| time | core | job | value | agg([core])                               | agg([job])                               | agg([core, job])                                                                                                         |
|------|------|-----|-------|-------------------------------------------|------------------------------------------|--------------------------------------------------------------------------------------------------------------------------|
| 1    | 1    | A   | 1     | `({core: 1}, 1)`                          | `({job: A}, 1)`                          | `({core: 1, job: A}, 1)`                                                                                                 |
| 2    | 2    | A   | 10    | `({core: 1}: 1)` <br> `({core: 2}: 10)`   | `({job: A}, 11)`                         | `({core: 1, job: A}, 1)` <br> `({core: 2, job: A}, 10)`                                                                  |
| 3    | 1    | B   | 100   | `({core: 1}: 100)` <br> `({core: 2}: 10)` | `({job: A}, 11)` <br> `({job: B}, 100)`  | `({core: 1, job: A}, 1)` <br> `({core: 2, job: A}, 10)` <br> `({core: 1, job: B}, 100)`                                  |
| 4    | 2    | B   | 1000  | `({core: 1}: 1)` <br> `({core: 2}: 1000)` | `({job: A}, 11)` <br> `({job: B}, 1100)` | `({core: 1, job: A}, 1)` <br> `({core: 2, job: A}, 10)` <br> `({core: 1, job: B}, 100)` <br> `({core: 2, job: B}, 1000)` |

Here the aggregated value of _agg([job])_ depends on the value of the "core" label, even though it's not preserved in the aggregation: it is the sum over the last-reported value for each core.

This makes configuring views for Observers significantly more complicated than for Measures.

This OTEP does not attempt to solve this problem.
We propose to make LastValue the only valid aggregation for Observer instruments, and leave it up to the user to specify label keys that produce coherent metrics for each registered Observer.

## Prior art and alternatives

One alternative is not to include a view API, and require users to configure metric instruments and aggregations individually.
This approach is partially described in the current spec.
See the motivation section for an argument against this approach.

This API borrows heavily from [OpenCensus views](https://opencensus.io/stats/view/).

See also [DataDog metric types](https://docs.datadoghq.com/developers/metrics/types/?tab=distribution#metric-types) and [Prometheus metric types](https://prometheus.io/docs/concepts/metric_types).

## Open questions

Should we ignore calls to record measurements from instruments that don't appear in any view?
This is potentially a nice optimization, but it would mean that no metrics are exported by default from instrumented code.

If we do ignore these calls, is there still a reason to have default aggregation types?

Should we treat correlation context label keys and user-specified label keys differently?

Should it be possible to create aggregator instances, and is there any use for these in the SDK?

As written, we expect exporters to infer the exposition format from the aggregation type in the view.
Some aggregations may map to multiple exposition formats for a given backend.
What should we do in this case?

Should the spec include a list of standard aggregations included in the SDK, including histogram and sketch?
The spec suggests this now:

> [Other standard aggregations are available, especially for Measure instruments, where we are generally interested in a variety of forms of statistics, such as histogram and quantile summaries.](https://github.com/open-telemetry/opentelemetry-specification/blob/ac75cfea2243ac46232cbc05c595bb0c018e2b58/specification/api-metrics.md#aggregations)

## Future Possibilities

### Automatic view creation

Configuring views for each metric instrument separately requires detailed knowledge about the instrumenting code, including the list of metric instrument names and possible label keys.
One solution for integrations, i.e. libraries that instrument other libraries, is to include a list of default views that the application owner or operator can choose to enable.
See, for example, the [list of default views for the Java OpenCensus gRPC integration](https://github.com/census-instrumentation/opencensus-java/blob/8b1fd5bbf98b27d0ad27394891e0c64c1171cb2b/contrib/grpc_metrics/src/main/java/io/opencensus/contrib/grpc/metrics/RpcViewConstants.java).

This solution doesn't apply to APM vendors, who provide backend-specific exporters and aggregators, and doesn't apply in situations where the instrumentation owner doesn't know _a priori_ which metric instruments need to be registered in views.
We should consider a configuration option that automatically creates a view for each metric instrument with a default aggregation based on its type.
