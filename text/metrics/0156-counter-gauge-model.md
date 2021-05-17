# Counter, UpDownCounter, and Gauge instruments explained

Counter and Gauge instruments are different in the ways they convey
meaning, and they are interpreted in different ways.  Attributes
applied to metric events enable further interpretation.  Because of
their semantics, the interpretive outcome of adding an attribute for
Counter and Gauge instruments is different.

With Counter instruments, a new attribute can be introduced with
additional measurements to subdivide a variable count.

With Gauge instruments, a new attribute can be introduced with
additional measurements to make multiple observations of a variable.

The OpenTelemetry Metrics API introduces a new kind of instrument, the
UpDownCounter, that behaves like a Counter, meaning that attributes
subdivide the variable being counted, but their primary interpretation
is like that of a Gauge.

## Background

OpenTelemetry has a founding principal that the interface (API) should
be decoupled from the implementation (SDK), thus the Metrics project
set out to define the meaning of metrics API events.

OpenTelemetry uses the term _temporality_ to describe how Sum
aggregations are accumulated across time, whether they are reset to
zero with each interval (_delta_) or accumulated over a sequence of
intervals (_cumulative_).  Both forms of temporality are considered
important, as they offer a useful tradeoff between cost and
reliability.  The data model specifies that a change of temporality
does not change meaning.

OpenTelemetry recognizes both synchronous and asynchronous APIs are
useful for reporting metrics, and each has unique advantages.  When
used with Counter and UpDownCounter instruments, there is an assumed
relationship between the aggregation temporality and the choice of
synchronous or asynchronous API.  Inputs to synchronous
(UpDown)Counter instruments are the changes of a Sum aggregation
(i.e., deltas).  Inputs to asynchronous (UpDown)Counter instruments
are the totals of a Sum aggregation (i.e., cumulatives).

## Glossary

_Meaning_: Metrics API events have a semantic definition that dictates
the meaning of the event, in particular how to interpret the integer
or floating point number value passed to the API.

_Interpretation_: How we extract information from metrics data using
the semantics of the API and the semantics of the OTLP data points.

_Metric instrument_ is a named instrument, belonging to an
instrumentation library, declared with one of the OpenTelemtetry
Metrics API instruments.  For the purpose of this text, it is a
Counter, an UpDownCounter, or a Gauge.

_Metric attributes_ can be applied to Metric API events, which allows
interpreting the meaning of events using different subsets of
attribute dimensions.

_Metric data stream_ is a collection of data points, written by a
writer, having an identity that consists of the instrument's name, the
instrumentation library, resource attributes, and metric attributes.

_Metric data points_ are the items in a stream, each has a point
kind. For the purpose of this text, the point kind is Sum or Gauge.
Sum points have two options: Temporality and Monotonicity.

_Metric timeseries_ is the output of aggregating a stream of data
points for a specific set of resource and attribute dimensions.

## Meaning and interpretation of Counter and UpDownCounter events

Counter and UpDownCounter instruments produce Sum metric data
points that are taken to have meaning in a metric stream, independent
of the aggregation temporality, as follows:

- Sum points are quantities that define a rate of change with respect to time
- Rate-of-change over time combined with a reset time may be used to derive a current total.

The rate interpretation is preferred for monotonic Sum points, and the
the current total interpretation is preferred for non-monotonic Sum
points.  Both interpretations are meaningful and useful for both kinds
of Sum point.

Sum points imply a linear scale of measurement.  A Sum value that is
twice the amount of another actually means twice as much of the
variable was counted.  Linear interpolation is considered to preserve
meaning.

## Meaning and interpretation of Gauge events

Gauge instruments produce Gauge metric data points are taken to
have meaning in a metric stream as follows:

- Gauge point values are individual measurements captured at an instant in time
- Gauge points record the last known value in a series of individual measurements.

Note that these two statements imply different interpretation for
synchronous and asynchronous measurements.  When recording Gauge
values through a synchronous API, the interpretation is "last known
value", and when recording Gauge values through an asynchronous API
the interpretation is "current value".

The distinction between last known value (synchronous) and current
value (asynchronous) is considered not significant in the data model.
Contrasting with Sum points, less can be assumed about the
measurements.  No implied linear scale of measurement, therefore:

- Rate-of-change may not be well defined
- Ratios are not necessarily meaningful
- Linear interpolation is not necessarily supported.

## Attributes are used for interpretation

Metric attributes enable new ways to interpret a stream of metric
data.  Metric attributes add information without changing the value of
a metric event.  Addition and removal of metric attributes can be
accomplished safely by applying transformations that preserve meaning.

Addition of attributes on a metric event can create new timeseries, by
producting a of greater number of distinct attribute sets.  However,
the meaning in the original events is preserved in the complete set of
timeseries.

Removing attributes from metric streams without changing meaning
requires re-aggregation, in general, which means applying the natural
aggregation function to merge metric streams.

For example, any metric event with no attributes:

```
gauge.Set(value)
```

can be extended by a new attribute, without changing its meaning or
altering any existing interpretation:

```
gauge.Set(value, { 'property': this.property })
```

## New measurements: Counter and UpDownCounter instruments

Sum points have been defined to have linear scale of measurement,
therefore Sum points can be subdivided.  A single Counter event can be
logically replaced by multiple Counter events having an equal sum.
This property allows the producer of metric events to introduce new
measurements, while preserving existing interpretation.

For example, it is reasonable to replace a single Counter event adding `x+y`:

```
counter.Add(x+y)
```

with separate counter events and one additional attribute:

```
counter.Add(x, { 'property': 'X' })
counter.Add(y, { 'property': 'Y' })
```

This property for Sum points makes it possible to configure an
instrumentation library with or without subdivided Sums and to
meaningfully aggregate data with a mixture of attributes.

## New measurements: Gauge instruments

Gauge instruments, unlike Counter instruments, cannot be subdivided.
Multiple Gauge measurements cannot be meaningfully combined using
addition.  In the time dimension, Gauge instrument events are
aggregated by taking the last value.

The same aggregation can be applied when removing an attributes from
metric streams forces reaggregation.  The most current value should be
selected.  In case of identical timestamps, a random value should be
selected to preserve the meaning of the Gauge.

For example, a Gauge for expressing a vehicle's speed relative to the
ground can be expressed either as the speed of its midpoint or by an
independent measurement of the speed of each wheel.

```
speedGauge.Set(vehicleSpeed)
```

This can be replaced by one Gauge per wheel, since wheel speed and
vehicle speed each define vehicle speed relative to the ground:

```
for i := 0; i < 4; i++ {
  speedGauge.Set(wheelSpeed[i], { 'wheel': i })
}
```

This form of Gauge rewrite is generally useful to capture additional
measurements by creating distinct metric streams.

## Meaning-preserving attribute erasure

Several rules for rewriting metric events that preserve meaning have
been shown above, focused on introducing new attributes and new
measurements in ways do not change existing meaning or alter existing
interpretations.

Removing attributes from metric events does not, by definition, change
their meaning, since attributes are interpreted as event selectors.
Removing attributes from aggregated streams of OpenTelemetry Metrics
data requires attention to the meaning being conveyed.

Safe attribute erasure for OpenTelemetry Metrics streams is specified
in a way that preserves meaning while removing only the forms of
interpretation that made use of the erased attribute.

_Reaggregation_ describes the process of combining OpenTelemetry
metric streams.  For reaggregation to preserve meaning, Sum points
must be combined by adding the inputs and Gauge points must be
combined by selecting the last or random value.

Note that erasure of attributes is defined so that it reverses the
effect of introducing new measurements, and meaning is preserved in
both directions.  This explains the definition for default
aggregations that should be applied when re-aggreation OpenTelemetry
metrics streams.  Sum streams are re-aggregated to preserve the
implied rate, while Gauge points are reggregated to preserve the
implied distribution of individual values.

## Conveying meaning to the user

OpenTelemetry states a requirement separating the API from the
implementation, and to do so we have defined the meaning of metrics
API events.  To preserve meaning through stages of reaggregation, we
have specified distinct default aggregation rules for Counter and
Gauge streams.

When attributes are used with Counter and Gauge instruments, every
distinct combination of attribute values determines a separate
OpenTelemetry metrics stream, and each stream conveys meaning
independently.  Because meaning is independent from the attributes
used, the user may wish to disregard some attributes when interpreting
a stream of metrics, restricting their attention to a subset of
attributes.

In database systems, this process is refered to as a performing a
"Group-By", where aggregation is used to combine streams within each
distinct set of grouped attributes.  For the benefit of OpenTelemetry
users, Metrics systems are encouraged to choose a a meaning-preserving
aggregation when grouping metric streams to convey meaning to the
user.

When conveying meaning to the user by grouping and aggregating over a
subset of attribute keys, the default aggregation selected should be
one that preserves meaning.  For monotonic Counter instruments, this
means conveying the combined rate of each group.  For UpDownCounter
instruments, this means conveying the combined total of each group.
For Gauge instruments, this means conveying the combined distribution
of each group.

## Choice of UpDownCounter or Gauge

The OpenTelemetry UpDownCounter instrument resembles the Gauge
instrument, but streams generated from these instruments apply
different aggregation rules by default.  The choice of instrument
should be made to ensure that the default aggregation rule preserves
meaning, as that is the point of these definitions.

Examining Gauge instruments in existing systems for anecdotal evidence
suggests that a significant majority of Gauges should be written as
UpDownCounters in OpenTelemetry.  Examples are given below.

### UpDownCounter measurements

UpDownCounter instruments are used for capturing quantities, where
typical examples include:

- Queue size
- Memory size
- Cache size
- Active requests
- Live object count

To test that these quantities are suitable UpDownCounter measurements,
verify that adding two inputs together logically produces another of
the same type and scale of measurement.  A queue size plus a queue
size yields a queue size, for example; add one count of live objects
with another, and you have a count of live objects.  By choosing the
UpDownCounter, developers ensure that the meaning conveyed is a sum,
which ensures the correct rate interpretation.

When interpreting total sums aggregated from UpDownCounter
instruments, it is important to consider the set of contributing
attributes, which determine the scale of measurement.  If one server
outputs UpDownCounter data in two attribute dimensions while another
uses three attribute diensions, the mean value is not a meaningful
quantity.  The process of correcting mixed attribute dimensions for
cumulative sums is referred to as _dimensional alignment_.

### Gauge measurements

Gauge instruments are used for capturing physical measurements,
calculated ratios, and results of function evaluation.  For example:

- CPU utilization
- CPU temperature
- Fan speed
- Water pressure
- Success/failure ratio

To test that these are suitable Gauge measurements, verify that adding
two inputs together does not logically produce a measurement of the
same type.

A CPU utilization plus a CPU utilization cannot meaningfully be used
as a measure of CPU utilization, it is just the sum of two CPU
utilizations.

A fan speed plus a fan speed has the correct units (a fan speed), but
the result is not a meaningful quantity.  Two fans spinning at one
speed is not the same as one fan spinning at twice the speed.

In some of these cases, it may be logical but practically impossible
to use one or more Counter instruments in place of Gauges.  CPU
utilization can be derived from a usage Counter.  Fan speed can be
derived from a revolution Counter.

## Summary

The OpenTelemetry Metrics data model supports addition and removal of
attributes in a way that preserves meaning.  This design gives
developers the ability to introduce new attributes in a safe way.

OpenTelemetry metrics developers are asked to consider whether they
want an UpDownCounter or Gauge when making asynchronous measurements,
and they should make this decision based on whether the default
aggregation rule for UpDownCounter or Gauge preserves meaning.  This
decision comes down to whether attributes are meant to subdivide a Sum
point or qualify a Gauge point.

The default aggregation rules for OpenTelemetry metrics data points
ensure that meaning is preserved when removing attributes from a
stream of metrics data.  The rules for reaggregation specify that
attributes should be safely removed before aggregating with other
metrics that are missing the same attributes, a process referred to as
dimensional alignment.

This design allows optional attributes to be included by the SDK in
metric data when it is available, such as those extracted from
TraceContext Baggage, in ways that consumers of the metrics data can
interpret correctly.

Having the ability to automatically remove attributes without changing
the meaning of Counter, UpDownCounter, and Gauge metrics API events
makes it possible for OpenTelemetry collectors to be configured with
re-aggregation rules, which can be managed by users in order to limit
collection costs.
