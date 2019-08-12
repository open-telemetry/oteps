# Eliminate stats.Record functionality

*Status: proposed*

Remove `stats.Record` from the specification, following the MeasureMetric type (RFC 0004-metric-measure).

## Motivation

`stats.Record` is no longer a necessary interface. There are conceivable reasons to support it, but they are outweighed by the cost of implementing and supporting two interfaces for recording metrics and statistics.

## Explanation

In RFC 0004-metric-measure, a new MeasureMetric type is introduced to replace raw statistics, with support for pre-defined label values.  With the new type introduced, it's now possible to record formerly-raw statistics through a higher-level Metric interface.

## Internal details

This simply involves removing the low-level `stats.Record` API from the specification, as it is no longer required.

## Trade-offs and mitigations

There are two reasons to maintain a low-level API that we know of:

1. For _generality_.  An application that forwards metrics from another source may need to handle metrics in generic code.  For these applications, having type-specific Metric handles could actually require more code to be written, whereas the low-level `stats.Record` API is more amenable to generic use.
1. For _atomicity_.  An application that wishes to record multiple statistics in a single operation can feel confident computing formulas based on multiple metrics, not worry about inconsistent views of the data.

## Prior art and alternatives

Raw statistics were a solution to confusion found in existing metrics APIs over Metric types vs. Aggregation types.  This proposal accompanies RFC 0003-metric-pre-defined-labels and RFC 0004-metric-measure.md in proposing that we think about Metric _type_ as independent of which aggregations apply.  Once we have a Metric to support histogram and summary aggregations, we no longer need raw statistics, and we no longer need `stats.Record`.  This avoids introducing new concepts (Raw statistics), at the same time departs from prior art in letting one Metric type support both Histogram and Summary aggregations.

## Open questions

Are either of the trade-offs described above important enough to keep the low-level `stats.Record` API?

## Future possibilities

This restricts future possibilities for the benefit of a smaller, simpler specification.

This leaves open the possibility of adding `stats.Record` functionality later, when the need is more clearly recognized.