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
standard aggregation for Counters.

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
one of these archetypes.  The foundational instruments are
unrestricted, in the sense that metric events support any numerical
value, positive or negative, zero or infinity.

The distinction between the two instrument archetypes is in their
synchronicity.  Measure instruments are called synchronously by the
user, while Observer instruments are called asynchronously by the SDK.
Synchronous instruments (Measure and refinements) have three calling
patterns (_Bound_, _Unbound_, and _Batch_) to capture measurements.
Asynchronous instruments (Observer any refinements) use callbacks to
capture measurements.

All measurements, synchronous or asynchronous, produce a metric event
([Context, timestamp, instrument descriptor, label set, and numerical
value](api-metrics.md#metric-event-format)), however there exists a
semantic distinction between synchronous and asynchronous events
related to the definition of "last value".  This is due to the
relationship with time.

Synchronous events happen concurrently, meaning we can only determine
whether one event happens before another by referring to timestamp
When querying events from synchronous instruments, you may find
multiple events for the same instrument, label set, and timestamp.

Asynchronous events are captured sequentially, meaning there is always
a well-defined _last value_.  When querying events from asynchronous
instruments, you cannot find more than one event for the same
instrument, label set, and timestamp.  Values observed asynchronously
are referred to as the _last value_ that was observed.

### Standard implementation

Sum and Count for both Measure and Observer.

### First Refinement: Counter

Captures non-negative increments (deltas).

And it matters because Prometheus needs to know.  (Duh.)

### More refinements: Monotonic, Non-negative

Measures
--------
Measure            unstricted, sumcount
Counter            non-negative, sum
UpDownCounter      unrestricted, sum
NonNegativeMeasure non-negative, sumcount

Observers
---------
Observer            unrestricted, sumcount
MonotonicObserver   unrestricted/monotonic, sum
NonNegativeObserver non-negative, sumcount
DeltaObserver       unrestricted, sum


## Explanation

Explain the proposed change as though it was already implemented and you were explaining it to a user. Depending on which layer the proposal addresses, the "user" may vary, or there may even be multiple.

We encourage you to use examples, diagrams, or whatever else makes the most sense!

## Internal details

From a technical perspective, how do you propose accomplishing the proposal? In particular, please explain:

* How the change would impact and interact with existing functionality
* Likely error modes (and how to handle them)
* Corner cases (and how to handle them)

While you do not need to prescribe a particular implementation - indeed, OTEPs should be about **behaviour**, not implementation! - it may be useful to provide at least one suggestion as to how the proposal *could* be implemented. This helps reassure reviewers that implementation is at least possible, and often helps them inspire them to think more deeply about trade-offs, alternatives, etc.

## Trade-offs and mitigations

What are some (known!) drawbacks? What are some ways that they might be mitigated?

Note that mitigations do not need to be complete *solutions*, and that they do not need to be accomplished directly through your proposal. A suggested mitigation may even warrant its own OTEP!

## Prior art and alternatives

What are some prior and/or alternative approaches? For instance, is there a corresponding feature in OpenTracing or OpenCensus? What are some ideas that you have rejected?

## Open questions

What are some questions that you know aren't resolved yet by the OTEP? These may be questions that could be answered through further discussion, implementation experiments, or anything else that the future may bring.

## Future possibilities

What are some future changes that this proposal would enable?

