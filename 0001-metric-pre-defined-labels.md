# Pre-defined label support for all metric operations

Let all Metric objects (Cumulative, Gauge, ...) and raw statistics support pre-defined label values.

## Motivation

In the current `Metric.GetOrCreateTimeSeries` API for Gauges and Cumulatives, the caller obtains a `TimeSeries` handle for repeatedly recording metrics with certain pre-defined label values set.  This is an important optimization, especially for exporting aggregated metrics.

The use of pre-defined labels improves usability too, for working with metrics in code. Application programs with long-lived objects and associated Metrics can compute predefined label values once (e.g., in a constructor), rather than once per call site.

The current API for recording raw statistics does not support the same optimization or usability advantage.  This RFC proposes to add support for pre-defined labels on all metrics.

## Explanation

In the current proposal, Metrics are used for _common_ pre-aggregated metric types, whereas Raw statistics are used for uncommon and vendor-specific aggregations.  The optimization and the usability advantages gained with pre-defined labels should be extended to Raw statistics because they are equally important and equally applicable.

For example, where the application wants to compute a histogram of some value (e.g., latency), there's good reason to pre-aggregate such information.  In this exampler, this allows an implementation to effienctly export the histogram of latencies "grouped" into individual results by label value(s).

## Internal details

This RFC is accompanied by RFC 0002-metric-measure which proposes to create a new Metric type to replace raw statistics.  The metric type, named "Measure", would replace the existing concept and type named "Measure" in the metrics API.  The new MeasureMetric object would support a `Record` method to record measurements.

## Trade-offs and mitigations

This is a refactoring of the existing proposal to cover more use-cases and does not introduce new complexity.  The existing mechanism for metrics is extended raw statistics by a new metric type, which allows eliminating raw statistics, leaving equal complexity.

## Prior art and alternatives

This Measure Metric API is conceptually close to the Prometheus [Histogram, Summary, and Untyped metric types](https://prometheus.io/docs/concepts/metric_types/).

## Open questions

This RFC is co-dependent on several others; it's an open question how to address this concern if the other RFCs are not accepted.

## Future possibilities

This change will potentially help clarify the relationship between Metric types and Aggregation types.  In a future RFC, we will propose that Measure Metrics (as introduced here) can be used to support arbitrary "advanced" aggregations including histograms and distribution summaries.
