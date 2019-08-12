# Pre-defined label support for all metric operations

*Status: proposed*

Let all Metric objects (Cumulative, Gauge, ...) and Raw statistics support pre-defined label values.

## Motivation

In the current `Metric.GetOrCreateTimeSeries` API for Gauges and Cumulatives, the caller obtains a `TimeSeries` handle for repeatedly recording metrics with certain pre-defined label values set.  This is an important optimization, especially for exporting aggregated metrics.

The use of pre-defined labels improves usability too, for working with metrics in code. Application programs with long-lived objects and associated Metrics can compute predefined label values once (e.g., in a constructor), rather than once per call site.

The current API for recording Raw statistics does not support the same optimization or usability advantage.  This RFC proposes to add support for pre-defined labels on all metrics.

## Explanation

In the current proposal, Metrics are used for pre-aggregated metric types, whereas Raw statistics are used for uncommon and vendor-specific aggregations.  The optimization and the usability advantages gained with pre-defined labels should be extended to Raw statistics because they are equally important and equally applicable. This is a new requirement.

For example, where the application wants to compute a histogram of some value (e.g., latency), there's good reason to pre-aggregate such information.  In this example, it allows an implementation to effienctly export the histogram of latencies "grouped" into individual results by label value(s).

## Internal details

This RFC is accompanied by RFC 0004-metric-measure which proposes to create a new Metric type to replace Raw statistics.  The metric type, named "Measure", would replace the existing concept and type named "Measure" in the metrics API.  The new MeasureMetric object would support a `Record` method to record measurements.

## Trade-offs and mitigations

This is a refactoring of the existing proposal to cover more use-cases and arguably reduces API complexity.

## Prior art and alternatives

Prometheus supports the notion of vector metrics, which are those with declared dimensions.  The vector-metric API supports a variety of methods like `WithLabelValues` to associate labels with a metric handle, similar to `GetOrCreateTimeSeries` in the existing proposal.  As in this proposal, Prometheus supports vector metrics for all metric types.

## Open questions

This RFC is co-dependent on several others; it's an open question how to address this concern if the other RFCs are not accepted.

## Future possibilities

This change will potentially help clarify the relationship between Metric types and Aggregation types.  In a future RFC, we will propose that MeasureMetrics can be used to support arbitrary "advanced" aggregations including histograms and distribution summaries.
