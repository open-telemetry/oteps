# Explain the metric instruments

Propose and explain final names for the standard metric instruments theorized in [OTEP 88](https://github.com/open-telemetry/oteps/pull/88) and address related confusion.

## Motivation

[OTEP 88](https://github.com/open-telemetry/oteps/pull/88) introduced a logical structure for metric instruments with two foundational categories of instrument, called "synchronous" vs. "asynchronous", named "Measure" and "Observer" in the abstract.  This proposal identified four kinds of "refinement" and mapped out the space of _possible_ instruments, while not proposing which would actually be included in the standard.

[OTEP 93](https://github.com/open-telemetry/oteps/pull/93) proposed with a list of six standard instruments, the most necessary and useful combination of instrument refinements, plus one special case used to record timing measurements.  OTEP 93 was closed without merging after a more consistent approach to naming was uncovered.  [OTEP 96](https://github.com/open-telemetry/oteps/pull/96) made another proposal, that was closed in favor of this one.

This proposal finalizes the naming proposal for standard instruments, seeking to address core confusion related to the "Measure" and "Observer" terms:

1. OTEP 88 stipulates that the terms currently in use to name synchronous and asynchronous instruments--"Measure" and "Observer"--become _abstract_ terms.  It also used phrases like "Measure-like" and "Observer-like" to discuss instruments with refinements.  This proposal states that we shall prefer the adjectives, commonly abbreviated "Sync" and "Async", when describing the kind of an instrument.  "Measure-like" means an instrument is synchronous.  "Observer-like" means that an instrument is asynchronous.
2. There is inconsistency in the hypothetical naming scheme for instruments presented in OTEP 88.  Note that "Counter" and "Observer" end in "-er", a noun suffix used in the sense of "[person occupationally connected with](https://www.merriam-webster.com/dictionary/-er)", while the term "Measure" does not fit this pattern.  This proposal proposes to replace the abstract term "Measure" by "Recorder", since the associated function name (verb) is specified as `Record()`.

This proposal also repeats the current specification--and the justification--for the default aggregation of each standard instrument.

## Explanation

The following table summarizes the final proposed standard instruments resulting from this set of proposals.  The columns are described in more detail below.

| Existing name | **Standard name** | Instrument kind | Function name | Default aggregation | Measurement kind | Kind of data | Rate support (Monotonic) | Notes |
| ------------- | ----------------------- | ----- | --------- | -------------- | ------------- | --- | ------------------------------------ |
| Counter       | **Counter**             | Sync  | Add()     | Sum            | Delta         | Additive | Yes | Per-request, part of a monotonic sum |
|               | **UpDownCounter**       | Sync  | Add()     | Sum            | Delta         | Additive | No  | Per-request, part of a non-monotonic sum |
| Measure       | **ValueRecorder**       | Sync  | Record()  | MinMaxSumCount | Instantaneous | Event    | No  | Per-request, any non-additive measurement |
| Observer      | **DeltaObserver**       | Async | Observe() | Sum            | Delta         | Additive | Yes | Per-interval, part of a monotonic sum |
|               | **UpDownDeltaObserver** | Async | Observe() | Sum            | Delta         | Additive | No  | Per-interval, part of a non-monotonic sum |
|               | **SumObserver**         | Async | Observe() | Sum            | Cumulative    | Additive | Yes | Per-interval, reporting a monotonic sum |
|               | **UpDownSumObserver**   | Async | Observe() | Sum            | Cumulative    | Additive | No  | Per-interval, reporting a non-monotonic sum |
|               | **ValueObserver**       | Async | Observe() | MinMaxSumCount | Instantaneous | Event    | No  | Per-interval, any non-additive measurement |

The scheme proposed here uses "What you've done to it" as a naming principle.  There are three synchronous instruments and five asunchronous instruments, because synchronous cumulative instruments are excluded (see [OTEP 88]()).  In a synchronous context (i.e., in running code, with local and/or distributed Context, carrying correlation values and SpanContext), the API encourages minimally processed input data.  Hopefully, all you've "done" is measure something and captured it with an instrument.  This allows the SDK to reduce overhead by dropping measurements that are not being collected, for example.

In asynchronous contexts, there are more options because there are more ways to collect data ("what you've done") over an interval of time.  Either you've computed a delta, you're observing something cumulative, or you have another kind of measurement.  All of these cases are considered _observations_, because only one numerical value can be captured per interval, per distinct set of labels, per asynchronous instrument.  Asynchronous instruments support processed measurements, with a calling pattern that allows the SDK to limit the overhead of expensive measurements.

All additive measurements support an `UpDown-` form that allows the sum to rise and fall.  By default, `Counter`, `DeltaObserver`, and `SumObserver` support rate aggregation because they do not permit falling sums.

Synchronous cumulative instruments are excluded from the standard based on the [OpenTelemetry library guidelines]().  Simply that to report a cumulative value correctly at runtime requires a degree of synchronization that OpenTelemetry API will not incorporate itself.  We cannot block for the sake of instrumentation, therefore we should not use synchronous cumulative instruments.

With eight instruments in total, one may be curious--how does the historical Metrics API term _Gauge_ translate into this specification?  _Gauge_, in Metrics API terminology, may cover all of these instrument use-cases with the exception of `Counter`.  As defined in [OTEP 88](), the OpenTelemetry Metrics API will disambiguate these use-cases by requiring *single purpose instruments*.  The choice of instrument implies a default interpretation, a standard aggregation, and suggests how to treat Metric data in observability systems, out of the box.

Uses of `Gauge` translate into the various OpenTelemetry Metric instruments depending on what you've done to produce a single number, and whether the measurement is made synchronously or not.  The "What you've done to it" principle implies that the name refers to what you're putting in, not what you're getting out.  Historical instrument names like `Gauge`, `Histogram`, and `Summary` are suggestive of what you get out.

Summarizing the naming scheme:

- If you've measured an amount of something that adds up to a total, where you are mainly interested in that total, use an additive instrument:
  - If synchronous and monotonic, use `Counter` with non-negative values
  - If synchronous and not monotonic, use `UpDownCounter` with arbitrary values
  - If asynchronous and non-negative deltas are measured, use `DeltaObserver`
  - If asynchronous and arbitrary deltas are measured, use `UpDownDeltaObserver`
  - If asynchronous and a cumulative, monotonic sum is measured, use `SumObserver`
  - If asynchronous and a cumulative, arbitrary sum is measured, use `UpDownSumObserver`
- If the measurements are non-additive or additive with an interest in the distribution, where you are interested in individual measurements:
  - If synchronous, use `ValueRecorder` to record a value that is part of a distribution
  - if asynchronous use `ValueObserver` to record a single measurement nearing the end of a collection interval.

### Sync vs Async instruments

Synchronous instruments are called in a request context, meaning they potentially have an associated tracing context and distributed correlation values.  Multiple metric events may occur for a synchronous instrument within a given collection interval.

Asynchronous instruments are reported by a callback, once per collection interval, and lack request context.  They are permitted to report only one value per distinct label set per period.  If the application observes multiple values in a single callback, for one collection interval, the last value "wins".

### Temporal quality

Measurements can be described in terms of their relationship with time.

Delta measurements are those that measure a change to a sum.  Delta instruments are usually selected because the program does not need to compute the sum for itself, but is able to measure the change.  In these cases, it would require extra state for the user to report cumulative values and reporting deltas is natural.

Cumulative measurements are those that report the current value of a sum.  Cumulative instruments are usually selected because the program maintains a sum for its own purposes, or because changes in the sum are not instrumented.  In these cases, it would require extra state for the user to report delta values and reporting cumulative values is natural.

Delta and Cumulative instruments are referred to, collectively, as Additive instruments.  Cumulative, synchronous instruments are not included in the standard because, although they are logically sensible, there exists little demand for these instruments.

Instantaneous measurements are those that report a non-additive measurement, one where it is not natural to compute a sum.  Instantaneous instruments are usually chosen when the distribution of values is of interest, not only the sum.

### Function names

Synchronous delta instruments support an `Add()` function, signifying that they add to a sum and are not cumulative.

Synchronous instantaneous instruments support a `Record()` function, signifying that they capture individual events, not only a sum.

Asynchronous instruments all support an `Observe()` function, signifying that they capture only one value per measurement interval.

### Rate support

Rate aggregation is supported for Counter, DeltaObserver, and SumObserver instruments in the default implementation.

Non-additive instruments do not express a sum, therefore are not useful for aggregating rates.

The `UpDown-` forms of additive instrument are not suitable for aggregating rates because the up- and down-changes in state may cancel each other. 

### Defalt Aggregations

Additive instruments use `Sum` aggregation by default, since by definition they are used when only the sum is of interest.

Instantaneous instruments use `MinMaxSumCount` aggregation by default, which is an inexpensive way to summarize a distribution of values.

## Detail

Here we discuss the eight proposed instruments individually and mention other names considered for each.

### Counter

`Counter` is the most common synchronous instrument.  This instrument supports an `Add(delta)` function for reporting a sum, and is restricted to non-negative deltas.  The default aggregation is `Sum`, as for any additive instrument, which are those instruments with Delta or Cumulative measurement kind.

Example uses for `Counter`:
- count the number of bytes received
- count the number of accounts created
- count the number of checkpoints run
- count a number of 5xx errors.

These example instruments would be useful for monitoring the rate of any of these quantities.  In these situations, it is usually more convenient to report a change of the associated sums, as the change happens, as opposed to maintaining and reporting the sum.

Other names considered: `Adder`.

### UpDownCounter

`UpDownCounter` is similar to `Counter` except that `Add(delta)` supports negative deltas.  This makes `UpDownCounter` not useful for computing a rate aggregation.  It aggregates a `Sum`, only the sum is non-monotonic.  It is generally useful for counting changes in an amount of resources used, or any quantity that rises and falls, in a request context.

Example uses for `UpDownCounter`:
- count memory in use by instrumenting `new` and `delete`
- count queue size by instrumenting `enqueue` and `dequeue`
- count semaphore `up` and `down` operations.

These example instruments would be useful for monitoring resource levels across a group of processes.

Other names considered: `NonMonotonicCounter`.

### ValueRecorder

`ValueRecorder` is a non-additive synchronous instrument useful for recording any non-additive number, positive or negative.  Values captured by a `ValueRecorder` are treated as individual events belonging to a distribution that is being summarized.  `ValueRecorder` should be chosen when capturing measurements that do not contribute meaningfully to a sum.

One of the most common uses for `ValueRecorder` is to capture latency measurements.  Latency measurements are not additive in the sense that there is little need to know the latency-sum of all processed requests.  We use a `ValueRecorder` instrument to capture latency measurements typically because we are interested in knowing mean, median, and other summary statistics about individual events.

The default aggregation for `ValueRecorder` computes the minimum and maximum values, the sum of event values, and the count of events, allowing the rate and range of input values to be monitored.

Example uses for `ValueRecorder` that are non-additive:
- capture any kind of timing information.

Example _additive_ uses of `ValueRecorder` capture measurements that are cumulative or delta values, by nature.  These are recommended `ValueRecorder` applications, as opposed to the hypothetical synthronous cumulative instrument:
- capture a request size
- capture an account balance
- capture a queue length
- capture a number of board feet of lumber.

These examples show that although they are additive in nature, choosing `ValueRecorder` as opposed to `Counter` or `UpDownCounter` implies an interest in more than the sum.  If you did not care to collect information about the distribution, you would have chosen one of the additive instruments instead.  Using `ValueRecorder` makes sense for distributions that are likely to be important, in an observability setting.

Use these with caution because they naturally cost more than capturing additive measurements.

### DeltaObserver

...

Example uses for `DeltaObserver`.
- [TODO]

### UpDownDeltaObserver

...

Example uses for `UpDownDeltaObserver`.
- [TODO]

### SumObserver

...

Example uses for `SumObserver`.
- capture process user/system CPU seconds
- capture the number of cache misses

### UpDownSumObserver

...

Example uses for `SumObserver`.
- capture process heap size


### ValueObserver

...

Example uses for `SumObserver`.
- CPU fan speed
- CPU temperature

## Open Questions

Helpers:

- A timing-specific ValueRecorder?
- A synchronous cumulative?



