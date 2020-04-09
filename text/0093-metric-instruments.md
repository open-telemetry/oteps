# List of new metric instruments

Formalize the new metric instruments proposed in [OTEP 88](https://github.com/open-telemetry/oteps/pull/88).

## Motivation

OTEP 88 introduced a framework for reasoning about new metric
instruments with various refinements and ended with a [sample
proposal](https://github.com/open-telemetry/oteps/pull/88#sample-proposal).
This proposal uses that proposal as a starting point.

## Explanation

The four instrument refinements discussed in OTEP 88 are:

* Sum-only: When computing only a sum is the instrument's primary purpose
* Non-negative: When negative values are invalid
* Precomputed-sum: When the application has computed a cumulative sum itself
* Non-negative-rate: When a negative rate is invalid.

These refinements are not for exposing directly to users at the API
level.  These concepts are purely explanatory, used to define the
properties of the metric instruments presented in the API.  Following
OTEP 88:

* Users will select instruments based on their specified properties
* Instruments are associated with a Descriptor, that includes the instrument kind (an enumeration)
* Exported metric events include the instrument Descriptor, allowing exporters to interpret event values.

In other words, these refinements serve to define the set of
instruments.  Both users and exporters will deal in concrete kinds of
instrument, these refinements are just for explaining their
properties.

OTEP 88 describes how we are meant to compute rate information from
instruments having the Non-negative-rate refinement.  Temporal
aggregation (over time) must be treated as a special case, compared
with spatial aggregation (over labels).  The logic for computing rates
depends on whether the Precomputed-sum refinement is present or not,
which determines whether Delta or Cumulative values are being
captured.

OTEP 88 proposes that when adding new instruments, we specify
instruments having a single purpose, each with a distinct set of
refinements, and each with a carefully selected name for the
properties of the instrument.

OTEP 88 also also proposes to support language-specific specialization
as well, to support built-in value types (e.g., timestamps).

## Internal details

The existing specification includes three instruments.  In this
proposal, the two foundational instruments become abstract, in the
sense that all synchronous instruments are instances of a Measure
instrument and all asynchronous instruments are instances of an
Observer instrument.  New instrument names are given to the
unrestricted, unrefined versions of rthe foundational instruments:

1. **Distribution** is an an unrefined Measure instrument.  Distribution accepts positive or negative values and uses MinMaxSumCount aggregation by default.
2. **LastValueObserver** is an unrefined Observer instrument.  LastValueObserver accepts positive or negative values and uses MinMaxSumCount aggregation by default.

The existing Counter instrument is unchanged in this proposal.  It has
Sum-only and Non-negative refinements, which imply the
Non-negative-rate refinement.  Counter uses Sum aggregation by default

Two new synchronous instruments are introduced in this proposal.

1. **UpDownCounter** is a Sum-only instrument with no other refinements.  It supports capturing positive and negative changes to a sum (deltas).  UpDownCounter uses Sum aggregation by default.
2. **Timing** is a Non-negative instrument specialized for the native clock duration measured on the platform.  It ensures that duration values are always captured with correct units, that ensures exporters can convert duration measurements correctly.

Two new asynchronous instruments are introduced in this proposal.

1. **CumulativeObserver** is a Sum-only, Precomputed-sum, Non-negative, Non-negative-rate instrument useful when reporting precomputed sums.
2. **DeltaObserver** is a Sum-only, Non-negative, Non-negative-rate instrument useful for capturing deltas that accumulate during a collection interval.

Both new asynchronous instruments are meant to be used for aggregating rate information from a callback.

### Instruments not specified

This proposal brings the number of specified instruments to seven and leaves room for more instruments to be added in the future.  As discussed in OTEP 88, other possibilities THAT WE DO NOT propose standardizing include:

1. **CumulativeCounter** would be as synchronous instrument for reporting a cumulative value with the Non-negative-rate refinement.
2. **UpDownCumulativeCounter** would be as synchronous instrument for reporting a cumulative value with the Non-negative-rate refinement.
3. **AbsoluteDistribution** would be a synchronous instrument for reporting a distribution of non-negative values.
4. **UpDownDeltaObserver** would be an asynchronous instrument for reporting positive and negative deltas to a sum.
5. **UpDownCumulativeObserver** would be an asynchronous instrument for reporting a cumulative sum without a Non-negative-rate refinement.
6. **AbsoluteLastValueObserver** would be an asynchronous instrument for reporting non-negative values.

These could be standardized in the future if there is sufficient
demand.  Although not standard, the behavior of each of these
instruments can be obtained by configuring one of the standard
instruments with non-standard aggregation.  We will wait and see.

## Trade-offs and mitigations

There are known limitations caused by not standardizing all possible
instrument refinements.  Creating too many instruments will create
confusion of its own, so we choose to limit the set of standard
instruments.  It is possible that SDK support for configuring
alternate aggregations will avoid the need for more standard
instruments.

There are potential incompatibilities related to the input range of
existing exporters and the new instruments.  For example, a Prometheus
Histogram is used to capture distribution with non-negative values.
This proposal does not specify a standard AbsoluteDistribution
instrument, which has the corresponding input-range restriction.

We are left recommending that Prometheus Histogram users adopt an
OpenTelemetry Distribution and continue to capture non-negative
values.  Non-negative values will be reported correctly, the only
behavioral difference relates to error handling.  Whereas a Prometheus
Histogram generates an error for negative inputs, the OpenTelemetry
Distribution accepts negative inputs.  A Prometheus exporter could be
configured to work around this (e.g., by reporting negative
distributions seprately), but [metric events that were correct in the
original system will continue to be correct in OpenTelemetry](https://github.com/open-telemetry/oteps/pull/88#discussion_r404912359).
