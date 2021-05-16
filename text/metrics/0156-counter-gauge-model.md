# Counter, UpDownCounter, and Gauge instruments explained

Counter and Gauge instruments are different in the ways they convey
meaning, and they are interpreted in different ways.  Attributes
applied to metric events enable further interpretation.  Because of
their semantics, the interpretive outcome of adding an attribute for
Counter and Gauge instruments is different.

With Counter instruments, a new attribute can be introduced along with
additional measurements to subdivide the thing being counted.

With Gauge instruments, a new attribute can be introduced along with
additional measurements to convey multiple observations of the same
variable.

The OpenTelemetry Metrics API has introduced a new kind of instrument,
the UpDownCounter, that behaves like a Counter, meaning that
attributes subdivide the thing being counted, but are interpreted like
a Gauge, meaning that users are most interested in the total sum, not
the rate of change.

## Background

OpenTelemetry has a founding principal that the interface (API) should
be decoupled from the implementation (SDK), thus the Metrics project
set out to give meaning to Metric API events.

OpenTelemetry uses the term temporality to describe how aggregations
are accumulated across time, whether they are reset to zero with each
interval ("delta") or accumulated over a sequence of intervals
("cumulative").  Both temporality forms are considered important, as
they offer a useful tradeoff between cost and reliability.  The data
model specifies that a change of temporality does not change meaning.

OpenTelemetry recognizes both synchronous and asynchronous APIs are
useful for reporting metrics, and each has unique advantages.  When
used with Counter and UpDownCounter instruments, there is an assumed
relationship between the aggregation temporality and the choice of
synchronous or asynchronous API.  Inputs to synchronous
(UpDown)Counter instruments are the changes of a Sum aggregation
(i.e., deltas).  Inputs to asynchronous (UpDown)Counter instruments
are the totals of a Sum aggregation (i.e., cumulatives).

## Meaning vs. Interpretation

The terms "Meaning" and "Interpretation" are used below to describe
how semantics are conveyed from the producer of metric events,
typically a developer working in the OpenTelemetry API, and the
consumer of metric timeseries, typically the user of an observability
product.

- Meaning is the indicated significance of a metric event, _what_ information the producer is stating through the use of an OpenTelemetry API
- Interpretation is a recipe for deriving information from a metric event, or _how_ the consumer can benefit from the stated information.

## Meaning and interpretation of Sum points

Sum points are taken to have meaning in a stream, independent of
aggregation temporality, as follows:

- Sum points are quantities that define a rate of change with respect to time
- Rate-of-change over time may be used to derive a current total.

The rate interpretation is preferred for monotonic Sum points, and the
the current total interpretation is preferred for non-monotonic Sum
points.  Both interpretations are meaningful and useful for both
monotonic and non-monotonic Sum points.

Sum points imply a linear scale of measurement.  A Sum value that is
twice the amount of another actually means twice as much of something
was counted.  Linear interpolation is considered to preserve meaning.

## Meaning and interpretation of Gauge points

Gauge points, which do not have temporality, are taken to have meaning
in a stream as follows:

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
measurements:

- No implied linear scale of measurement, therefore
- Rate-of-change may not be well defined
- Ratios are not necessarily meaningful
- Linear interpolation is not necessarily supported.

## Comparing Gauge and non-monotonic cumulative Sum

The non-monotonic Sum point, the one that naturally results from an
asynchronous UpDownCounter instrument, appears similar in nature to a
Gauge.  Both can be used to express the current value in a series.

As stated, there is a difference between Sum and Gauge points in the
data model that can be seen when adding and removing attributes from
metric events.

## Attributes are used for interpretation

The meaning of an attribute is derived from the interpretation of the
metric.  Attributes are used to logically restrict attention to a
subset of metric events, therefore the use of attributes leads to
additional interpretation.

Because attributes function logically as event selectors, we are able
to assert that the addition of an attribute does not by definition
change the meaning of a metric event.  The addition of an attribute
may create new timeseries, by producting a of greater number of
distinct attribute sets, though the meaning is unchanged.

For example, any metric event with no attributes:

```
gauge.Set(value)
```

can be extended by a new attribute, without changing its meaning or
interpretation:

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
instrumentation library with or without subdivided Sums, with no
change in the meaning conveyed.

## New measurements: Gauge instruments

Gauge instruments, unlike Counter instruments, cannot be subdivided.
Likewise, multiple Gauge measurements cannot be meaningfully combined
using addition.  As meaning is derived from individual values, new
measurements can be introduced using attributes to create distinct
streams.

By interpreting Gauge points as a distribution of current values, new
measurements can be defined to be preserve meaning.

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

This form of Gauge rewrite is generally useful when additional
measurements offer more information and are drawn from the same
distribution.

## Meaning-preserving attribute erasure

Several rules for rewriting metric events that preserve meaning have
been shown above, focused on introducing new attributes and new
measurements in ways do not change existing meaning and
interpretation.

Removing attributes from metric events does not, by definition, change
their meaning, since attributes are interpreted as event selectors.
Removing attributes from aggregated streams of OpenTelemetry Metrics
data requires attention to the meaning being conveyed.

Safe attribute erasure for OpenTelemetry Metrics streams is specified
in a way that preserves meaning while removing only the forms of
interpretation that made use of the erased attribute.

Re-aggregation describes the process of combining OpenTelemetry
metrics streams.  For re-aggregation to preserve meaning, Sum points
must be combined by adding the inputs and Gauge points must be
combined by forming a distribution of the inputs.

Note that erasure of attributes is defined so that it reverses the
effect of introducing new measurements, and meaning is preserved in
both directions.  This explains the definition for default
aggregations that should be applied when re-aggreation OpenTelemetry
metrics streams.  Sum streams are re-aggregated to preserve the
implied rate, while Gauge points are reggregated to preserve the
implied distribution.

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
