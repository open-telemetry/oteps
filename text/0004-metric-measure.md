# Replace Raw statistics with Measure-type Metric

**Status:** `proposed`

Define a new Metric type named "Measure" to cover existing "Raw" statistics uses.

## Motivation

The primary motivation is that Raw statistics should support the optimization and usability improvements associated with pre-defined label values (0003-metric-pre-defined-labels).  By elevating non-Cumulative, non-Gauge statistics to the same conceptual level as Metrics in the API, we effectively make the type of a metric independent from whether it supports pre-defined labels.

This also makes it possible to eliminate the low-level `stats.Record` interface from the API specification entirely (0005-eliminate-stats-record).

## Explanation

This proposal suggests we think about which aggregations apply to a metric independently from its type.  A MeasureMetric could be used to aggregate a Histogram, or a Summary, or _both_ of these aggregations simultaneously.  This proposal makes metric type independent of aggregation type, whereas there is a precedent for combining these types into one.

The proposal here suggests that we think of the metric type in terms of the _action performed_ (i.e., which _verb_ applies?).  Gauges support the `Set` action. Cumulatives support an `Inc` action. Measures support a `Record` action.

This extends the `GetOrCreateTimeSeries` (pre-defined labels) functionality supported by Metrics to what has been known as Raw statistics, satisfying the change in capability requested in RFC 0003-metric-pre-defined-labels.  This allows programmers to predefine labels for all metrics.  This is not only an important potential optimization for the programmer, it is a usability improvement in the code.

There are no new requirements stated in this RFC.

## Internal details

The type known as `MeasureMetric` is a direct replacement for Raw statistics.  The `MeasureMetric.Record` method records a single observation of the metric.  The `MeasureMetric.GetOrCreateTimeSeries` supports pre-defined keys as discussed in 0003-metric-pre-defined-labels.

## Trade-offs and mitigations

This change, while it eliminates the need for a Raw statistics concept, potentially introduces new required concepts.  Whereas Raw statistics have no directly-declared aggregations, introducing MeasureMetric raises the question of which aggregations apply.  We will propose how a programmer can declare recommended aggregations (and good defaults) in RFC 0006-configurable-aggregation.

## Prior art and alternatives

This Measure Metric API is conceptually close to the Prometheus [Histogram, Summary, and Untyped metric types](https://prometheus.io/docs/concepts/metric_types/).

## Open questions

With this proposal accepted, there would be three Metric types: Gauge, Cumulative, and Measure.  This proposal does not directly address what to do over the existing, conflicting uses of "Measure".

## Future possibilities

This change enables metrics to support configurable aggregation types, which allows the programmer to provide recommended aggregations at the point where Metrics are defined.  This will allow support for good out-of-the-box behavior for metrics defined by third-party libraries, for example.

Without Raw statistics in the API, it becomes possible to elimiante the low-level `stats.Record` API, which may also be desirable.
