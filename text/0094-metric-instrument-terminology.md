# Rationalize naming of metric instruments and their default aggregations

Propose final names for the seven metric instruments introduced in [OTEP 93](https://github.com/open-telemetry/oteps/pull/93) and address related confusion.

## Motivation

[OTEP 88](https://github.com/open-telemetry/oteps/pull/88) introduced
a logical structure for metric instruments with two foundational
categories of instrument, called "synchronous" vs. "asynchronous",
named "Measure" and "Observer" in the abstract.  This proposal
identified four kinds of "refinement" and mapped out the space of
_possible_ instruments, while not proposing which would actually be
included in the standard.

[OTEP 93](https://github.com/open-telemetry/oteps/pull/93) followed
with a list of six standard instruments, the most necessary and useful
combination of instrument refinements, plus one special case used to
record timing measurements.

This proposal finalizes the names used to describe the seven
instruments above, seeking to address core confusion related to
"Measure":

1. OTEP 88 stipulates that the terms currently in use to name
synchronous and asynchronous instruments become abstract, but also
using "Measure-like" and "Observer-like" to discuss instruments with
refinements.  This proposal states that we shall prefer the
adjectives, commonly abbreviated "Sync" and "Async", when describing
instruments.
2. Prior to OTEP 88, but even with OTEPs 88 and 93 included, there is
inconsistency in the naming of instruments.  Note that "Counter" and
"Observer" end in "-er", a noun suffix used in the sense of "[person
occupationally connected
with](https://www.merriam-webster.com/dictionary/-er)", while the term
"Measure" does not fit this pattern.  This proposal proposes to
replace the abstract term "Measure" by "Recorder", since the
associated method name (verb) is specified as `Record()`.
3. The OTEP 93 asynchronous instruments ("LastValueObserver",
"DeltaObserver", and "CumulativeObserver") have the pattern
"-Observer", while the OTEP 93 synchronous instruments
("Counter", "UpDownCounter", "Distribution", "Timing") do not.  This
proposal keeps "Counter" and "UpDownCounter" for Sum-only synchronous instruments, and does the same
with "Recorder", yielding "Recorder" and "TimingRecorder".
4. Confusion over the loss of "Gauge" is addressed by replacing
"LastValueObserver" with "GaugeObserver".

This proposal also repeats the current specification of the default
Aggregator for each kind of instrument.

## Explanation

The following table summarizes the four synchronous instruments and
three asynchronous instruments that will be standardized as a result
of this set of proposals.

| Existing name | OTEP 93 name       | Final name             | Sync or Async | Default aggregation | Rate support |
| ------------- | ------------------ | ---------------------- | ----------- | ---------- | ---- |
| Counter       | Counter            | **Counter**            | Sync  | Sum | Yes | 
| Measure       | Distribution       | **Recorder**           | Sync  | MinMaxSumCount | No |
|               | UpDownCounter      | **UpDownCounter**      | Sync  | Sum | Yes |
|               | Timing             | **TimingRecorder**     | Sync  | MinMaxSumCount  | No |
| Observer      | LastValueObserver  | **GaugeObserver**      | Async | MinMaxSumCount | No |
|               | DeltaObserver      | **DeltaObserver**      | Async | Sum | Yes |
|               | CumulativeObserver | **CumulativeObserver** | Async | Sum | Yes |

The argument for "Recorder" instead of "Distribution" is that we
should prefer instrument descriptives associated with the action being
performed ("occupationally connected"), not the value being computed,
as the latter is dependent on SDK configuration.  A "Recorder" records
a value that is part of a distribution.  A "Counter" counts a value
that is part of a sum.  An "GaugeObserver" observes an instantaneous
value ("reads a gauge").  A "Recorder" records an arbitrary value.  A
"TimingRecorder" records a timing value, and so on.

## Details

This proposal consolidates OTEP 88 and OTEP 93 and proposes a consistent
pattern for naming instruments.  It will be the source of truth when
applying OTEP 88 and OTEP 93 to the OpenTelemetry metrics specification.

### Default aggregations

This [OTEP 93
conversation](https://github.com/open-telemetry/oteps/pull/93#discussion_r405852507)
raised a question about the default aggregation for GaugeObserver,
given as MinMaxSumCount.  Would "Sum" be a more appropriate default?

Note that the distinction between whether the default aggregation is
"Sum" or "MinMaxSumCount" corresponds exactly to whether the
instrument has the Sum-only refinement.  "Sum" is the default
aggregation for any Sum-only instrument since, by definition, the
Sum aggregation provides complete information.

The three instruments with a default "MinMaxSumCount" are all used to
record a value that is, by defintion, more than only a sum.  In this
case, "complete information" requires recording every value, i.e., no
aggregation.  MinMaxSumCount is applied in these cases because it
provides the maximum amount of information that can be recorded using
a fixed number of values, per time series, per collection interval.

### GaugeObserver aggregation

Why should GaugeObserver aggregate the Min, Max, Sum, and Count when
it is permitted to observe just one measurement per interval?  This
says that when observed values are aggregated they should be treated
like a distribution--we are intersted in more than a sum, by
definition.  If observing only a sum, the DeltaObserver or
CumulativeObserver should be used instead.

Clearly, when Count equals 1, the Min, Max, and Sum are equal to the
value.  Exporters may be able take advantage of this fact when
exporting data from these instruments.  In particular, since it is
known that asynchronous instruments produce only one valiue per
interval (with last-value-wins semantics), when we know in the SDK
that no spatial aggregation is configured, we can be sure that Count
equals one, and we can use the most appropriate exposition format for
the target system.

This means Prometheus and Statsd exporters SHOULD export Gauge values
for the GaugeObserver when there is no spatial aggregation being
applied, because that is the natural exposition format for
MinMaxSumCount aggregations when Count equals 1.  If there is spatial
aggregation being applied, the default MinMaxSumCount aggregation
still applies.
