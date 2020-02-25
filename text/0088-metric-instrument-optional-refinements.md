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

Observer instruments have a well-defined _last value_ measured by the
instrument that can be useful in defining aggregations.  To maintain
this property, we impose this requirement: two or more calls to
`Observe()` in a single Observer callback invocation are treated as
duplicates of each other, and the last call to `Observe()` wins.

Measure instruments do not define a _last value_ relationship.

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
through instrument refinements.  Instrument refinements have the same
calling patterns as the foundational instrument they refine, adding
either a different standard implementation or a restriction of the
input domain.

We have done away with instrument options, in other words, in favor of
optional metric instruments.  Here we discuss three important
capabilities achievable using instrument refinements.

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
less than some prior value are considered measurement errors.

#### Sum-Only

A sum-only instrument is one where only the sum is considered to be of
interest.  For a Sum-Only instrument refinement, we have a semantic
property that two events with numeric values `M` and `N` are
semantically equivalent to a single event with value `M+N`.  For
example, in a count of users arriving by bus to an event, we are not
concerned with the number of buses that arrived.

A sum-only instrument is one where the number of events is not
counted, by default, only the `Sum`.

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
the above elements.  A `Counter`-like instrument that permits
non-negative updates could be called an `UpDownCounter`, for example.
An `Observer`-like instrument with non-descending values could be
called a `MonotonicObserver` instrument. An `Observer`-like instrument
that reports non-negative updates to a sum could be called a
`DeltaObserver` instrument.

## Internal details

This is a change of understanding.  It does not request any new
instruments be created or APIs be changed, but it does specify how we
should think about adding new instruments.

No API changes are called for in this proposal.

### Note for Cumulative Exporters

The Prometheus system collects cumulative data for its counter
instruments, meaning it exports the lifetime sum of a counter at each
collection interval, not the change in that sum over the collection
interval. 

It is important, therefore, to know when the output of an instrument
should be reflected as a cumulative value in the exporter.  The
`Counter` instrument automatically has this property, but why?

We can infer that a value is cumulative in the following circumstances:

- Non-negative and Sum-only: This is called `Counter`, if synchronous, and can be `DeltaObserver`, potentially, if asynchronous
- Monotonic: This can be `MonotonicObserver`, potentially.

We see that these refinements satisfy their intended purpose, which is
to convey additional semantics that exporters use to convert data into
their system.

## Trade-offs and mitigations

The trade-off explicitly introduced here is that we should prefer to
create new instrument refinements, each for a dedicated purpose,
rather than create generic instruments with support for multiple
semantic options.

## Prior art and alternatives

The optional behaviors `Monotonic` and `Absolute` were first discussed
in the August 2019 Metrics working group meeting.

## Open questions

This approach allows new instrument refinements to be considered on a
case-by-case basis.  For example, is a `MonotonicObserver` sufficient,
or do we also need a `DeltaObserver`?

## Future possibilities

A future OTEP will request the introduction of two new standard
refinements for the 0.4 API specification.  These will be a monotonic
Observer instrument named `MonotonicObserver` and a synchronous timing
instrument named `TimingMeasure`.
