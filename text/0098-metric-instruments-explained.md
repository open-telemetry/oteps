# Explain the metric instruments

Propose and explain final names for the standard metric instruments theorized in [OTEP 98](https://github.com/open-telemetry/oteps/pull/88) and address related confusion.

## Motivation

[OTEP 88](https://github.com/open-telemetry/oteps/pull/88) introduced a logical structure for metric instruments with two foundational categories of instrument, called "synchronous" vs. "asynchronous", named "Measure" and "Observer" in the abstract.  This proposal identified four kinds of "refinement" and mapped out the space of _possible_ instruments, while not proposing which would actually be included in the standard.

[OTEP 93](https://github.com/open-telemetry/oteps/pull/93) proposed with a list of six standard instruments, the most necessary and useful combination of instrument refinements, plus one special case used to record timing measurements.  OTEP 93 was closed without merging after a more consistent approach to naming was uncovered.

This proposal finalizes the names used to describe the standard instruments above, seeking to address core confusion related to the "Measure" and "Observer" terms:

1. OTEP 88 stipulates that the terms currently in use to name synchronous and asynchronous instruments become _abstract_ terms, still it sometimes uses phrases like "Measure-like" and "Observer-like" to discuss instruments with refinements.  This proposal states that we shall prefer the adjectives, commonly abbreviated "Sync" and "Async", when describing instruments.
2. There is inconsistency in the hypothetical naming scheme for instruments presented in OTEP 88.  Note that "Counter" and "Observer" end in "-er", a noun suffix used in the sense of "[person occupationally connected with](https://www.merriam-webster.com/dictionary/-er)", while the term "Measure" does not fit this pattern.  This proposal proposes to replace the abstract term "Measure" by "Recorder", since the associated method name (verb) is specified as `Record()`.
3. The OTEP 88 asynchronous instruments (e.g., "DeltaObserver", "CumulativeObserver") have the pattern "-Observer", while the synchronous instruments (e.g., "Counter", "Measure") do not have an obvious pattern. This proposal simplifies the pattern to create a correspondance between likewise synchronous and asynchronous instruments by adding "-Observer" to the name of the corresponding synchronous instrument (if it exists).
4. Cumulative instruments present a special naming challenge.  The "GaugeObserver" instrument is introduced to resolve an ambiguity, with special consideration given to how these measurements are aggregated.

This proposal also repeats the current specification--and the justification--for the default aggregation of each standard instrument.

## Explanation

The following table summarizes the standard instruments resulting from this set of proposals.

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

