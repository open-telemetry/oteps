# Metric Instruments

Removes the optional semantic declarations `Monotonic` and `Absolute`
for metric instruments, declares the Measure and Observer instruments
as _foundational_, and introduces a process for standardizing new
instrument "refinements".

## Motivation

With the removal of Gauge instruments and the addition of Observer
instruments in the specification, the existing `Monotonic` and
`Absolute` options began to create confusion.  For example, a Counter
instrument is used for capturing changes in a Sum, so we could say that
values which are non-negative (absolute) determine metric events
define a monotonic Counter.  The confusion arises, in this case,
because `Absolute` refers to the captured values, whereas `Monotonic`
refers to the instrument or, more precisely, to a property of the
standard aggregation for Counter instrumenrts.

From a different perspective, Counter instruments might be treated as
as refinements of the Measure instrument.  Whereas the Measure
instrument is used for capturing all-purpose synchronous measurements,
the Counter instrument is used specifically for capturing measurements
of synchronous changes in a sum, therefore it uses `Add()` instead of
`Record()` as the action and specifies `Sum` as the standard
aggregation.

What this illustrates is that we have modeled this space poorly.  This
does not propose to change any existing metric APIs, only our
understanding of the three instruments currently part of the
specification: Measure, Observer, and Counter (a refinement).

## Explanation

The Measure and Observer instrument are defined as _foundational_
here, in the sense that any kind of metric instrument must reduce to
one of these.  The foundational instruments are unrestricted, in the
sense that metric events support any numerical value, positive or
negative, zero or infinity.

The distinction between the two foundational instruments is whether
they are synchronous.  Measure instruments are called synchronously by
the user, while Observer instruments are called asynchronously by the
implementation.  Synchronous instruments (Measure and refinements)
have three calling patterns (_Bound_, _Unbound_, and _Batch_) to
capture measurements.  Asynchronous instruments (Observer and
refinements) use callbacks to capture measurements.

All measurements produce a metric event consisting of [timestamp,
instrument descriptor, label set, and numerical
value](api-metrics.md#metric-event-format).  Synchronous instrument
events additionally have [Context](api-context.md), describing
properties of the associated trace and distributed correlation values.

Observer instruments have a well-defined _last value_ of the
measurement that can be useful in defining aggregations.  To maintain
this property, we impose this requirement: two or more calls to
`Observer()` in a single Observer callback invocation are treated as
duplicates of each other, and the last call to `Observe()` wins.

Measure instruments do not define a _last value_ relationship.

### Standard implementation of Measure and Observer

OpenTelemetry specifies how the default SDK should treat metric
events, by default, when asked to export data from an instrument.
Usually, an aggregation is specified along with the label keys used in
the aggregation.  Measure and Observer instruments use `Sum` and
`Count` aggregators by default, in the standard implementation.  This
pair of measurements, of course, defines an average value.  There are
no restrictions placed on the numerical value in an event by one of
the foundational instruments.

### Refinements to Measure and Observer

Options like `Monotonic` and `Absolute` were removed in the 0.3
specification.  Here, we propose to regain the equivalent effects
through _instrument refinements_, which declare instruments with
calling patterns like Measure and Observer, but with different
standard implementations and standard-alternative implementations.  

We have done away with options on instruments, in other words, in
favor of optional metric instruments.  Here we discuss three important
capabilities used to make up instrument refinements.

#### Non-negative

For some instruments, such as those that measure real quantities,
negative values are meaningless.  For example, it is impossible for a
person to weigh a negative amount.

A non-negative instrument refinement accepts only non-negative values.
For instruments with this property, negative values are considered
measurement errors.

#### Monotonic

A monotonic instrument is one where the user promises that successive
metric events for a given instrument definition and label set will
differ by a non-negative value.  This is defined in terms of the last
value relationship, therefore only applies to refinements of the
Observer instrument. For example, the CPU time used by a process as
read in successive collection intervals cannot change by a negative
amount, because it is impossible to use a negative amount of CPU time.

A monotonic instrument refinement accepts only values that are greater
than or equal to the last-captured value of the instrument, for a
given label set.  For instruments with this property, values that are
less than some prior value are considered a measurement error.

#### Sum-Only

A sum-only instrument is one where only the sum is considered of
interest.  For a Sum-Only instrument refinement, we have a semantic
property that two events with numeric values `M` and `N` are
semantically equivalent to a single event with value `M+N`.  For
example, in a count of users arriving by bus to an event, we are not
concerned with the number of buses that arrived.

A sum-only instrument is one where the number of events is not
counted, by default, only the `Sum`.

#### Language-level refinements

OpenTelemetry implementations may wish to add instrument refinements
to accomdate built-in types.  Languages with distinct integer and
floating point should offer instrument refinements for each, leading
to type names like `Int64Measure` and `Float64Measure`.

A language with support for unsigned integer types may wish to create
dedicated instruments to report these values, leading to type names
like `UnsignedInt64Observer` and `UnsignedFloat64Observer`.

Other uses for built-in type refinements involve the type for duration
measurements.  Where there is built-in type for the difference between
two clock measurements, OpenTelemetry languages should offer a
refinement to automatically apply the correct units.

### Counter refinement

Counter is a non-negative, sum-only refinement of the Measure
instrument.

## Internal details

This is a change of understanding.  It does not request any new
instruments be created, only specifiy how we should think about adding
new instruments.

No API changes are called for in this proposal.

## Trade-offs and mitigations

The trade-off explicitly introduced here is that we will prefer to
create new instruments for each dedicated purpose, rather than create
generic instruments with support for multiple semantic options.

## Prior art and alternatives

The optional behaviors `Monotonic` and `Absolute` were first discussed
in the August 2019 Metrics working group meeting.

## Open questions

This approach allows questions about new instruments to be addressed
on a case-by-case basis.

## Future possibilities

A future OTEP will request the introduction of several new standard
refinements.  For example, a monotonic observer instrument named
`MonotonicObserver` and a timing instrument named `TimingMeasure`.
