# Composite Samplers

This proposal addresses head-based sampling as described by the [Open Telemetry SDK](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/sdk.md#sampling).
It introduces additional composite samplers. Composite samplers use other samplers (delegates) to make sampling decisions. The composite samplers invoke the delegate samplers, but eventually make the final call.

## Motivation

The need for configuring head sampling has been explicitly or implicitly indicated in several discussions, both within the [Samplig SIG](https://docs.google.com/document/d/1gASMhmxNt9qCa8czEMheGlUW2xpORiYoD7dBD7aNtbQ) and in the wider community. Some of the discussions are going back a number of years, see for example

- issue [173](https://github.com/open-telemetry/opentelemetry-specification/issues/173): Way to ignore healthcheck traces when using automatic tracer across all languages?
- issue [1060](https://github.com/open-telemetry/opentelemetry-java-instrumentation/issues/1060): Exclude URLs from Tracing
- issue [1844](https://github.com/open-telemetry/opentelemetry-specification/issues/1844): Composite Sampler

## The Goal

The goal of this proposal is to extend the set of standard SDK samplers by _Composite Samplers_ which make building advanced sampling configuration easier.

### Example

Let's assume that a user wants to configure head sampling as follows:

- for root spans:
  - drop all `/healthcheck` requests
  - capture all `/checkout` requests
  - capture 25% of all other requests
- for non-root spans
  - follow the parent sampling decision
  - however, capture all calls to service `/foo` (even if the trace will be incomplete)
- in any case, do not exceed 1000 spans/minute

It is hoped that the Composite Samplers presented here will be useful in constructing a sampling configuration satisfying the above case.

## AnyOf

`AnyOf` is a composite sampler which takes a non-empty list of Samplers (delegates) as the argument.

Upon invocation of its `shouldSample` method, it MUST go through the whole list and invoke `shouldSample` method on each delegate sampler, passing the same arguments as received.

`AnyOf` sampler MUST return a `SamplingResult` which is constructed as follows:

- The sampling Decision is based on the delegate sampling Decisions. If all of the delegate Decisions are `DROP`, the composite sampler MUST return `DROP` Decision as well.
If any of the delegate Decisions is `RECORD_AND_SAMPLE`, the composite sampler MUST return `RECORD_AND_SAMPLE` Decision.
Otherwise, if any of the delegate Decisions is `RECORD_ONLY`, the composite sampler MUST return `RECORD_ONLY` Decision.
- The set of span Attributes to be added to the `Span` is the sum of the sets of Attributes as provided by the delegate samplers within their `SamplingResults`s.
- The `TraceState` to be used with the new `Span` is obtained by cumulatively applying all the potential modfications of the parent `TraceState` by the delegate samplers.

Each delegate sampler MUST be given a chance to participate in the sampling decision as described above and MUST see the same _parent_ state. The order of the delegate samplers does not matter, as long as there's no overlap in the Attribute Keys or the trace state keys they use.

## Conjunction

`Conjunction` is a composite sampler which takes two Samplers (delegates) as the arguments. These delegate samplers will be hereby referenced as First and Second.

Upon invocation of its `shouldSample` method, the Conjunction sampler MUST invoke `shouldSample` method on the First sampler, passing the same arguments as received, and examine the received sampling Decision. Upon receiving `DROP` or `RECORD_ONLY` decision it MUST return the SamplingResult from the First sampler, and it MUST NOT proceed with querying the Second sampler. If the sampling decision from the First sampler is `RECORD_AND_SAMPLE`, the Conjunction sampler MUST invoke `shouldSample` method on the Second sampler.
If the sampling Decision from the Second sampler is `RECORD_AND_SAMPLE`, the Conjunction sampler MUST return a `SamplingResult` which is constructed as follows:

- The sampling Decision is `RECORD_AND_SAMPLE`.
- The set of span Attributes to be added to the `Span` is the sum of the sets of Attributes as provided by the First and the Second sampler.
- The `TraceState` to be used with the new `Span` is obtained by cumulatively applying the potential modfications from the First and Second sampler.

If the sampling Decision from the Second sampler is `RECORD_ONLY` or `DROP`, the Conjunction sampler MUST retrun a `SamplingResult` which is constructed as follows:

- The sampling Decision is `DROP`.
- The set of span Attributes to be added to the `Span` is empty.
- The `TraceState` to be used with the new `Span` is the passed-in `TraceState`.


## Rule based sampling

For rule-based sampling (e.g. when sampling decision should depend on Span attributes), the Spans can be grouped into separate categories, and each category can use a different Sampler. Categorization of Spans is aided by `Predicates`.
The Predicates represent logical expressions which can access Span Attributes (or anything else available when the sampling decision is to be made), and perform tests on the accessible values.
For example, one can test if the target URL for a SERVER span matches a given pattern.

### Predicate

`Predicate` interface allows users to create custom categories based on information that is availabe at the time of making the sampling decision.

#### SpanMatches

This is a function/method for `Predicate`, which returns `true` if a given `Span` matches, i.e. belongs to the category described by the Predicate.

#### Required Arguments

The arguments represent the values that are made availabe for `ShouldSample`.

- `Context` with parent Span.
- `TraceId` of the Span to be created.
- Name of the Span to be created.
- Initial set of Attributes of the Span to be created.
- Collection of links that will be associated with the Span to be created.

## RuleBased

`RuleBased` is a composite sampler which performs `Span` categorization and sampling.
It takes the following arguments:
- `SpanKind`
- list of `Predicate`s
- list of `Sampler`s

The `RuleBased` sampler SHOULD NOT accept lists of different length, i.e. it SHOULD report an error. Implementations MAY allow for replacing both lists with a list of pairs (`Predicate`, `Sampler`), if this is supported by the platform.

For making the sampling decision, if the `Span` kind matches the specified kind, the sampler goes through the lists in the declared order. If a Predicate matches the `Span` in question, the corresponding `Sampler` will be called to make the final sampling decision. If the `SpanKind` does not match, or none of the Predicates evaluates to `true`, the final decision is `DROP`.

## Limitations of composite samplers

Not all samplers can participate as components of composite samplers without undesired or unexpected effects. Some samplers require that they _see_ each `Span` being created, even if the span is going to be dropped. Some samplers update the trace state or maintain internal state, and for their correct behavior it it is assumed that their sampling decisions will be honored by the tracer at the face value in all cases. A good example for this are rate limiting samplers which have to keep track of the rate of created spans and/or the rate of positive sampling decisions.

Placing such samplers as the first or second argument for the `Conjunction` sampler can sometimes preserve their correct behavior.

A special attention is required for [consistent probability](https://opentelemetry.io/docs/specs/otel/trace/tracestate-probability-sampling/#consistent-probability-sampler) (CP) samplers. The sampling probability they record in trace-state is later used to calculate [_adjusted count_](https://opentelemetry.io/docs/specs/otel/trace/tracestate-probability-sampling/#adjusted-count), which, in turn, is used to calculate span-based metrics. Mixing CP samplers with other types of samplers in most cases will lead to incorrect adjusted counts. The family of CP samplers has its own [composition rules](https://opentelemetry.io/docs/specs/otel/trace/tracestate-probability-sampling/#composition-rules), which correctly handle composing multiple CP samplers.

## Prior art

A number of composite samplers are already available as independent contributions
([RuleBasedRoutingSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/samplers/src/main/java/io/opentelemetry/contrib/sampler/RuleBasedRoutingSampler.java),
[LinksBasedSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/samplers/src/main/java/io/opentelemetry/contrib/sampler/LinksBasedSampler.java)).
Also, historically, some Span categorization was introduced by [JaegerRemoteSampler](https://www.jaegertracing.io/docs/1.54/sampling/#remote-sampling).

This proposal aims at generalizing these ideas, and at providing a bit more formal specification for the behavior of the composite samplers.
