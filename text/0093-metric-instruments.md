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
* Precomputed-sum: When the application has computed a cumulative sum
* Non-negative-rate: When a negative rate is invalid.

OTEP 88 proposes that when adding new instruments, we specify
instruments having a single purpose, each with a distinct set of
refinements, with a carefully selected name for this kind of
instrument.

OTEP 88 also also proposes to support language-specific specialization
as well, to support built-in value types (e.g., timestamps).

## Internal details

The instruments are:

1. Measure (abstract) -> Distribution (concrete)
2. Observer (abstract) -> LastValueObserver (concrete)
3. Counter (concrete) -> Counter (sum-only, non-negative)
4. UpDownCounter      (synchronous)
5. Timing             (synchronous)
6. CumulativeObserver (asynchronous)
7. DeltaObserver      (asynchronous)


## Trade-offs and mitigations

What are some (known!) drawbacks? What are some ways that they might be mitigated?

Note that mitigations do not need to be complete *solutions*, and that they do not need to be accomplished directly through your proposal. A suggested mitigation may even warrant its own OTEP!

## Prior art and alternatives

What are some prior and/or alternative approaches? For instance, is there a corresponding feature in OpenTracing or OpenCensus? What are some ideas that you have rejected?

## Open questions

What are some questions that you know aren't resolved yet by the OTEP? These may be questions that could be answered through further discussion, implementation experiments, or anything else that the future may bring.

## Future possibilities

What are some future changes that this proposal would enable?
