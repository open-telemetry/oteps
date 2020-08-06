# Semantic Conventions for Timed Operations

## Motivation

Timed operations are generally fully described by Spans. Because these Spans may be
sampled, aggregating information about them downstream is subject to inaccuracies.
Computing count, duration, and error rate from _all_ operations, regardless of whether
the result in a sampled Span, results in more accurate aggregates.

## Explanation

Summarization of timed operations may be done with a single `ValueRecorder`, onto which
the duration of each instance of the timed operation is recorded.

### Naming

The semantic conventions should specify the naming of this metric instrument, following
the [Metric Instrument Naming Guidelines](./108-naming-guidelines.md#guidelines).

The goal of the semantic conventions regarding the naming of this metric instrument will
be that similar operations are rolled up together.

### Labels

The attributes described in the existing tracing semantic conventions provide guidance
for the labels to be added to this metric instrument. Some of these attributes will
result in very high cardinality; the semantic conventions should describe which attributes
should be included and which should be omitted or reformatted to reduce cardinality.

In addition, the tracing attributes allow more value types than metric instrument labels,
so semantic conventions must describe how to represent those values as labels.
