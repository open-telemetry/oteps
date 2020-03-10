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
values associated with a precomputed sum are still sums.  Precomputed
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

Counter is a sum-only, non-negative, thus non-negative-rate refinement
of the Measure instrument.

### Standardizing new instruments

With these refinements we can exhaustively list each distinct kind of
instrument.  There are a total of twelve hypothetical instruments
listed in the table below, of which only one has been standardized.
Hypothetical future instrument names are _italicized_.

| Foundation instrument | Sum-only? | Precomputed-sum? | Non-negative? | Non-negative-rate? | Instrument name _(hyptothetical)_ |
|--|--|--|--|--|--|
| Measure  | sum-only |                 | non-negative  | non-negative-rate | Counter |
| Measure  | sum-only | precomputed-sum |               | non-negative-rate | _CumulativeCounter_ |
| Measure  | sum-only |                 |               |                   | _UpDownCounter_ |
| Measure  | sum-only | precomputed-sum |               |                   | _UpDownCumulativeCounter_ |
| Measure  |          |                 | non-negative  |                   | _AbsoluteMeasure_ |
| Measure  |          |                 |               |                   | _NonAbsoluteMeasure_ |
| Observer | sum-only |                 | non-negative  | non-negative-rate | _DeltaObserver_ |
| Observer | sum-only | precomputed-sum |               | non-negative-rate | _CumulativeObserver_ |
| Observer | sum-only |                 |               |                   | _UpDownDeltaObserver_ |
| Observer | sum-only | precomputed-sum |               |                   | _UpDownCumulativeObserver_ |
| Observer |          |                 | non-negative  |                   | _AbsoluteObserver_ |
| Observer |          |                 |               |                   | _NonAbsoluteObserver_ |

To arrive at this listing, several assumptions have been made.  For
example, the precomputed-sum and non-negative-rate refeinments are only
applicable in conjunction with a sum-only refinement.

For the precomputed-sum instruments, we technically do not care
whether the inputs are non-negative, because rate aggregation computes
differences.  However, it is useful for other aggregations to assume
that precomputed sums start at zero, and we will ignore the case where
a precomputed sum has an initial value other than zero.

## Internal details

This is a change of understanding.  It does not request any new
instruments be created or APIs be changed, but it does specify how we
should think about adding new instruments.

No API changes are called for in this proposal.

## Example

Suppose you wish to capture the CPU usage of a process broken down by
the CPU core ID.  The operating system provides a mechanism to read
the current usage from the `/proc` file system, which will be reported
once per collection interval using an Observer instrument.  Because
this is a precomputed sum with a non-negative rate, use a
_CumulativeObserver_ to report this quantity with a metric label
indicating the CPU core ID.

It will be common to compute a rate of CPU usage over this data.  The
rate can be calculated for an individual CPU core by computing a
difference between the value of two metric events.  To compute the
aggregate rate across all cores–a spatial aggregation–these
differences are added together.

## Open question

Eleven instruments have been given hyptothetical names in the table
above, but only a subset of these should be included in the
specification.

An open question is whether the foundational instruments should be
considered "abstract", meaning that users can only create refined
instruments.

An argument in favor of treating the foundation instruments as
abstract goes like this: users will be confused because sometimes the
documentation and specification discusses Measure and Observer
instruments generally, and sometimes it discusses them specifically.
If the foundation instruments are abstract, this confusion is
eliminated.

An argument against treating the foundation instruments as abstract
goes like this: by excluding these short, well-understood names from
use in the API, we force long, less-well understood names on the user,
which will leave them confused.  For example, _NonAbsoluteObserver_ is
a completely unrefined Observer, and wouldn't you rather read and
write "Observer" in code?  (Likewise for _NonAbsoluteMeasure_ vs
Measure.)

## Trade-offs and mitigations

The trade-off explicitly introduced here is that we should prefer to
create new instrument refinements, each for a dedicated purpose,
rather than create generic instruments with support for multiple
semantic options.

## Prior art and alternatives

The optional behaviors `Monotonic` and `Absolute` were first discussed
in the August 2019 Metrics working group meeting.

## Future possibilities

A future OTEP will request the introduction of two standard
refinements for the 0.4 API specification.  This will be the
`CumulativeObserver` instrument described above plus a synchronous
timing instrument named `TimingMeasure` that is equivalent to
_AbsoluteMeasure_ with the correct unit and a language-specific
duration type for measuring time.

If the above open question is decided in favor of treating the
foundational instruments as abstract, instrument names like
_NonAbsoluteMeasure_ and _NonAbsoluteCounter_ will need to be
standardized.
