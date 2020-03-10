# Metric Instruments

Removes the optional semantic declarations `Monotonic` and `Absolute`
for metric instruments, declares the Measure and Observer instruments
as _foundational_, and introduces a process for standardizing new
instrument _refinements_.

## Motivation

With the removal of Gauge instruments and the addition of Observer
instruments in the specification, the existing `Monotonic` and
`Absolute` options began to create confusion.  For example, a Counter
instrument is used for capturing changes in a Sum, and we could say
that non-negative-valued metric events define a monotonic Counter, in
the sense that its Sum is monotonic.  The confusion arises, in this
case, because `Absolute` refers to the captured values, whereas
`Monotonic` refers to the semantic output.

From a different perspective, Counter instruments might be treated as
refinements of the Measure instrument.  Whereas the Measure instrument
is used for capturing all-purpose synchronous measurements, the
Counter instrument is used specifically for synchronously capturing
measurements of changes in a sum, therefore it uses `Add()` instead of
`Record()`, and it specifies `Sum` as the standard aggregation.

What this illustrates is that we have modeled this space poorly.  This
does not propose to change any existing metrics APIs, only our
understanding of the three instruments currently included in the
specification: Measure, Observer, and Counter.

## Explanation

The Measure and Observer instrument are defined as _foundational_
here, in the sense that any kind of metric instrument must reduce to
one of these.  The foundational instruments are unrestricted, in the
sense that metric events support any numerical value, positive or
negative, zero or infinity.

The distinction between the two foundational instruments is whether
they are synchronous.  Measure instruments are called synchronously by
the user, while Observer instruments are called asynchronously by the
implementation.  Synchronous instruments (Measure and its refinements)
have three calling patterns (_Bound_, _Unbound_, and _Batch_) to
capture measurements.  Asynchronous instruments (Observer and its
refinements) use callbacks to capture measurements.

All measurement APIs produce metric events consisting of [timestamp,
instrument descriptor, label set, and numerical
value](api-metrics.md#metric-event-format).  Synchronous instrument
events additionally have [Context](api-context.md), describing
properties of the associated trace and distributed correlation values.

### Last-value relationship

Observer instruments have a well-defined _last value_ measured by the
instrument that can be useful in defining aggregations.  To maintain
this property, we impose this requirement: two or more calls to
`Observe()` in a single Observer callback invocation are treated as
duplicates of each other, and the last call to `Observe()` wins.

Measure instruments do not define a _last value_ relationship.

### Aggregating changes to a sum: Rate calculation

The former `Monotonic` option had been introduced in order to support
reporting of a current sum, such that a rate calculation is implied.
Here we defined _Rate_ as an aggregation, defined for a subset of
instruments, that may be calculated differently depending on how the
instrument is defined.  The rate aggregation outputs the amount of
change in a quantity divided by the amount of change in time.

A rate can be computed from values that are reported as differences,
referred to as _delta_ reporting here, or as sums, referred to as
_cumulative_ reporting here.  The primary goal of the instrument
refinements introduced in this proposal is to facilitate rate
calculations in more than one way.

When delta reporting, a rate is calculated by summing individual
measurements or observations.  For Measure instruments, these values
fall into a range of time, as indicated by the event timestamp.  For
Observer instruments, these values fall into a range of collection
intervals.

When cumulative reporting, a rate is calculated by computing a
difference between individual values.  For an Observer instrument, we
compute rate over a range of collection intervals, and for a Measure
instrument we compute rate over a range of timestamps.  In either
case, we are interested in subtracting the final value from the prior
value measured or observed on the instrument.

Note that rate aggregation, as illustrated above, treats the time
dimension differently than the other dimensions used for aggregation.

### Standard implementation of Measure and Observer

OpenTelemetry specifies how the default SDK should treat metric
events, by default, when asked to export data from an instrument.
Measure and Observer instruments compute `Sum` and `Count`
aggregations, by default, in the standard implementation.  This pair
of measurements, of course, defines an average value.  There are no
restrictions placed on the numerical value in an event for the two
foundational instruments.

### Refinements to Measure and Observer

The `Monotonic` and `Absolute` options were removed in the 0.3
specification.  Here, we propose to regain the equivalent effects
through instrument refinements.  Instrument refinements are added to
the foundational instruments, yielding new instruments with the same
calling patterns as the foundational instrument they refine.  These
refinements support adding either a different standard implementation
or a restriction of the input domain to the instrument.

We have done away with instrument options, in other words, in favor of
optional metric instruments.  Here we discuss four significant
instrument refinements.

#### Non-negative

For some instruments, such as those that measure real quantities,
negative values are meaningless.  For example, it is impossible for a
person to weigh a negative amount.

A non-negative instrument refinement accepts only non-negative values.
For instruments with this property, negative values are considered
measurement errors.  Both Measure and Observer instruments support
non-negative refinements.

#### Sum-only

A sum-only instrument is one where only the sum is considered to be of
interest.  For a sum-only instrument refinement, we have a semantic
property that two events with numeric values `M` and `N` are
semantically equivalent to a single event with value `M+N`.  For
example, in a sum-only count of users arriving by bus to an event, we
are not concerned with the number of buses that arrived.

A sum-only instrument is one where the number of events is not
counted, only the `Sum`.  A key property of sum-only instruments is
that they always support a Rate aggregation, whether reporting delta-
or cumulative-values.  Both Measure and Observer instruments support
sum-only refinements.

#### Precomputed-sum

A precomputed-sum refinement indicates that values reported through an
instrument are observed or measured in terms of a sum that changes
over time.  Pre-computed sum instruments support cumulative reporting,
meaning the rate aggregation is defined by computing a difference
across timestamps or collection intervals.

A precomputed sum refinement implies a sum-only refinement.  Note that
values assocaited with a precomputed sum are still sums.  Precomputed
sum values are combined using addition, when aggregating over the
spatial dimensions; only the time dimension receives special treatment.

#### Non-negative-rate

A non-negative-rate instrument refinement states that rate aggregation
produces only non-negative results.  There are non-negative-rate cases
of interest for delta reporting and cumulative reporting, as follows.

For delta reporting, any non-negative and sum-only instrument is also
a non-negative-rate instrument.

For cumulative reporting, a sum-only and pre-computed sum instrument
does not necessarily have a non-negative rate, but adding an explicit
non-negative-rate refinement makes it the equivalent of `Monotonic` in
the 0.2 specification.

For example, the CPU time used by a process, as read in successive
collection intervals, cannot change by a negative amount, because it
is impossible to use a negative amount of CPU time.  CPU time a
typical value to report through an Observer instrument, so the rate
for a specific set of labels is defined by subtracting the prior
observation from the current observation.

#### Language-level refinements

OpenTelemetry implementations may wish to add instrument refinements
to accommodate built-in types.  Languages with distinct integer and
floating point should offer instrument refinements for each, leading
to type names like `Int64Measure` and `Float64Measure`.

A language with support for unsigned integer types may wish to create
dedicated instruments to report these values, leading to type names
like `UnsignedInt64Observer` and `UnsignedFloat64Observer`.  These
would naturally apply a non-negative refinment.

Other uses for built-in type refinements involve the type for duration
measurements.  For example, where there is built-in type for the
difference between two clock measurements, OpenTelemetry APIs should
offer a refinement to automatically apply the correct unit of time to
the measurement.

### Counter refinement

Counter is a non-negative, sum-only refinement of the Measure
instrument.

### Future refinements

This leaves the potential to include other refinements by combining
the above elements.  The following are current and proposed names for
three instruments that support non-negative rate reporting:

| Foundation | Refinements | Name |
|--|--|--|
| Measure | non-negative, sum-only, non-negative-rate | Counter |
| Observer | non-negative, sum-only, non-negative-rate | DeltaObserver |
| Observer | sum-only, precomputed-sum, non-negative-rate | CumulativeObserver |

The Counter instrument is already part of the specification.  A
proposal to introduce DeltaObserver and CumulativeObserver will follow
in a future OTEP.

## Internal details

This is a change of understanding.  It does not request any new
instruments be created or APIs be changed, but it does specify how we
should think about adding new instruments.

No API changes are called for in this proposal.

## Trade-offs and mitigations

The trade-off explicitly introduced here is that we should prefer to
create new instrument refinements, each for a dedicated purpose,
rather than create generic instruments with support for multiple
semantic options.

## Prior art and alternatives

The optional behaviors `Monotonic` and `Absolute` were first discussed
in the August 2019 Metrics working group meeting.

## Future possibilities

A future OTEP will request the introduction of several standard
refinements for the 0.4 API specification.  These will be the
`DeltaObserver` and `CumulativeObserver` instruments described above
plus a synchronous timing instrument named `TimingMeasure`.
