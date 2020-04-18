# Explain the metric instruments

Propose and explain final names for the standard metric instruments theorized in [OTEP 98](https://github.com/open-telemetry/oteps/pull/88) and address related confusion.

## Motivation

[OTEP 88](https://github.com/open-telemetry/oteps/pull/88) introduced a logical structure for metric instruments with two foundational categories of instrument, called "synchronous" vs. "asynchronous", named "Measure" and "Observer" in the abstract.  This proposal identified four kinds of "refinement" and mapped out the space of _possible_ instruments, while not proposing which would actually be included in the standard.

[OTEP 93](https://github.com/open-telemetry/oteps/pull/93) proposed with a list of six standard instruments, the most necessary and useful combination of instrument refinements, plus one special case used to record timing measurements.  OTEP 93 was closed without merging after a more consistent approach to naming was uncovered.

This proposal finalizes the names used to describe the standard instruments above, seeking to address core confusion related to the "Measure" and "Observer" terms:

1. OTEP 88 stipulates that the terms currently in use to name synchronous and asynchronous instruments become _abstract_ terms, still it sometimes uses phrases like "Measure-like" and "Observer-like" to discuss instruments with refinements.  This proposal states that we shall prefer the adjectives, commonly abbreviated "Sync" and "Async", when describing the kind of an instrument.
2. There is inconsistency in the hypothetical naming scheme for instruments presented in OTEP 88.  Note that "Counter" and "Observer" end in "-er", a noun suffix used in the sense of "[person occupationally connected with](https://www.merriam-webster.com/dictionary/-er)", while the term "Measure" does not fit this pattern.  This proposal proposes to replace the abstract term "Measure" by "Recorder", since the associated function name (verb) is specified as `Record()`.

This proposal also repeats the current specification--and the justification--for the default aggregation of each standard instrument.

## Explanation

The following table summarizes the standard instruments resulting from this set of proposals.  The columns are described in more detail below.

| Existing name | **Standard name** | Instrument kind | Function name | Default aggregation | Measurement kind | Rate support (Monotonic) | Notes |
| ------------- | ----------------------- | ----- | ------------- | ---------- | ---- | --- | --- |
| Counter       | **Counter**             | Sync  | Add()     | Sum            | Delta         | Yes | Per-request, part of a monotonic sum |
|               | **UpDownCounter**       | Sync  | Add()     | Sum            | Delta         | No  | Per-request, part of a non-monotonic sum |
| Measure       | **Recorder**            | Sync  | Record()  | MinMaxSumCount | Instantaneous | No  | Per-request, element in a distribution |
|               | **TimingRecorder**      | Sync  | Record()  | MinMaxSumCount | Instantaneous | No  | Same as above, with automatic duration units |
| Observer      | **DeltaObserver**       | Async | Observe() | Sum            | Delta         | Yes | Per-interval, part of a monotonic sum |
|               | **UpDownDeltaObserver** | Async | Observe() | Sum            | Delta         | No  | Per-interval, part of a non-monotonic sum |
|               | **SumObserver**         | Async | Observe() | Sum            | Cumulative    | Yes | Per-interval, reporting a monotonic sum |
|               | **UpDownSumObserver**   | Async | Observe() | Sum            | Cumulative    | No  | Per-interval, reporting a non-monotonic sum |
|               | **GaugeObserver**       | Async | Observe() | MinMaxSumCount | Instantaneous | No  | Per-interval, any non-additive measurement |

### Sync vs Async instruments

Synchronous instruments are called in a request context, meaning they potentially have an associated tracing context and distributed correlation values.  Multiple metric events may occur for a synchronous instrument within a given collection interval.

Asynchronous instruments are reported by callback, lacking a request context, once per collection interval.  They are permitted to report only one value per distinct label set per period, establishing a "last value" relationship which asynchronous instruments define and synchronous instruments do not.

### Temporal quality

Measurements can be described in terms of their relationship with time.

Delta measurements are those that measure a change to a sum.  Delta instruments are usually selected because the program does not need to compute the sum and is able to measure the change.  In these cases, it would require extra state for the user to report cumulative values.

Cumulative measurements are those that report the current value of a sum.  Cumulative instruments are usually selected because the program is able to measure the sum.  In these cases, it would require extra state for the user to report delta values.

Delta and Cumulative instruments are referred to, collectively, as Additive instruments.

Instantaneous measurements are those that report a non-additive measurement, one where it is not natural to compute a sum.  Instantaneous instruments are usually chosen to when the distribution of values is of interest, not only the sum.

### Function names

Synchronous delta instruments support an `Add()` function, signifying that they add to a sum and do not report a total count.

Synchronous instantaneous instruments support a `Record` function, signifying that they capture individual events, not only a sum.

Asynchronous instruments all support an `Observe()` function, signifying that they capture only one value per measurement interval.

### Rate support

Rate aggregation is supported for Counter, DeltaObserver, and SumObserver instruments.

The other instruments either report non-additive information, where the sum is not meaningful and the distribution itself is of interest.

### Defalt Aggregations

Additive instruments use `Sum` aggregation by default, since by definition they are used when only the sum is of interest.

Instantaneous instruments use `MinMaxSumCount` aggregation by default, which is an inexpensive way to summarize a distribution.

## Detail

TODO: WIP: This section is incomplete.

### Counter

`Counter` is the most common synchronous instrument, meaning it is called in request context.  This instrument supports an `Add(delta)` function for reporting a sum, and is restricted to non-negative deltas.  The default aggregation is `Sum`, as for any additive instrument, which are those instruments with Delta or Cumulative measurement kind.

Example uses for `Counter`:
- Report a number of bytes received
- ... a number of accounts created
- ... a number of checkpoints run
- ... a number of 5xx errors

These example instruments would be useful for monitoring the rate of any of these quantities.  In these situations, it is simply more convenient to report a change of the associated sums, where typically the program has no internal need to compute a lifetime total.  

### UpDownCounter

`UpDownCounter` is similar to `Counter` except that `Add(delta)` supports negative deltas.  This makes `UpDownCounter` not useful for computing a rate aggregation.  It aggregates a `Sum`, only the sum is non-monotonic.  It is generally useful for counting changes in an amount of resources used, or any quantity that rises and falls, in a request context.

Example uses for `UpDownCounter`:
- count memory in use by instrumenting `new` and `delete`
- count queue size by instrumenting `enqueue` and `dequeue`
- count semaphore `up` and `down` operations

These example instruments would be useful for monitoring resource levels across a group of processes.
