# Composite Samplers Proposal

This proposal addresses head-based sampling as described by the [Open Telemetry SDK](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/sdk.md#sampling).
It introduces additional _composite samplers_. Composite samplers use other samplers (delegates) to make sampling decisions. The composite samplers invoke the delegate samplers, but eventually make the final call.

The new samplers proposed here are compatible with [Threshold propagation in trace state (OTEP 235)](https://github.com/open-telemetry/oteps/pull/235) as used by Consistent Probability samplers. Also see Draft PR 3910 [Probability Samplers based on W3C Trace Context Level 2](https://github.com/open-telemetry/opentelemetry-specification/pull/3910). 

## Motivation

The need for configuring head sampling has been explicitly or implicitly indicated in several discussions, both within the [Samplig SIG](https://docs.google.com/document/d/1gASMhmxNt9qCa8czEMheGlUW2xpORiYoD7dBD7aNtbQ) and in the wider community. Some of the discussions are going back a number of years, see for example

- issue [173](https://github.com/open-telemetry/opentelemetry-specification/issues/173): Way to ignore healthcheck traces when using automatic tracer across all languages?
- issue [1060](https://github.com/open-telemetry/opentelemetry-java-instrumentation/issues/1060): Exclude URLs from Tracing
- issue [1844](https://github.com/open-telemetry/opentelemetry-specification/issues/1844): Composite Sampler

## The Goal

The goal of this proposal is to help creating advanced sampling configurations using pre-defined building blocks. Let's consider the following example of sampling requirements. It is believed that many users will have requirements following the same pattern. Most notable elements here are trace classification based on target URL, some spans requiring special handling, and putting a sanity cap on the total volume of exported spans.

### Example

Head-based sampling requirements.

- for root spans:
  - drop all `/healthcheck` requests
  - capture all `/checkout` requests
  - capture 25% of all other requests
- for non-root spans
  - follow the parent sampling decision
  - however, capture all calls to service `/foo` (even if the trace will be incomplete)
- in any case, do not exceed 1000 spans/minute

# New Samplers

## AnyOf

`AnyOf` is a composite sampler which takes a non-empty list of Samplers (delegates) as the argument.

Upon invocation of its `shouldSample` method, it MUST go through the whole list and invoke `shouldSample` method on each delegate sampler, passing the same arguments as received.

`AnyOf` sampler MUST return a `SamplingResult` which is constructed as follows:

- The sampling Decision is based on the delegate sampling Decisions. If all of the delegate Decisions are `DROP`, the composite sampler MUST return `DROP` Decision as well.
If any of the delegate Decisions is `RECORD_AND_SAMPLE`, the composite sampler MUST return `RECORD_AND_SAMPLE` Decision.
Otherwise, if any of the delegate Decisions is `RECORD_ONLY`, the composite sampler MUST return `RECORD_ONLY` Decision.
- The set of span Attributes to be added to the `Span` is the sum of the sets of Attributes as provided by the delegate samplers within their `SamplingResults`s.
- The `TraceState` to be used with the new `Span` is obtained by cumulatively applying all the potential modfications of the parent `TraceState` by the delegate samplers, with special handling of the `th` sub-key (the sampling rejection `THRESHOLD`) for the `ot` entry as described below.

If the final sampling Decision is `DROP` or `RECORD_ONLY`, the `th` entry MUST be removed.
If the sampling Decision is `RECORD_AND_SAMPLE`, and there's no `th` entry in any of the `TraceState` provided by the delegates that decided to `RECORD_AND_SAMPLE`, the `th` entry MUST be also removed.
Otherwise, the resulting `TraceState` MUST contain `th` entry with the `THRESHOLD` value being the minimum of all the `THRESHOLD` values as reported by those delegates that decided to `RECORD_AND_SAMPLE`.

Each delegate sampler MUST be given a chance to participate in the sampling decision as described above and MUST see the same _parent_ state. The order of the delegate samplers does not matter, as long as there's no overlap in the Attribute Keys or the trace state keys (other than `th`) that they use.

## EachOf

`EachOf` is a composite sampler which takes a non-empty list of Samplers (delegates) as the argument.

Upon invocation of its `shouldSample` method, it MUST go through the whole list and invoke `shouldSample` method on each delegate sampler, passing the same arguments as received.

`EachOf` sampler MUST return a `SamplingResult` which is constructed as follows:

- The sampling Decision is based on the delegate sampling Decisions. If all of the delegate Decisions are `RECORD_AND_SAMPLE`, the composite sampler MUST return `RECORD_AND_SAMPLE` Decision as well.
If any of the delegate Decisions is `DROP`, the composite sampler MUST return `DROP` Decision.
Otherwise, if any of the delegate Decisions is `RECORD_ONLY`, the composite sampler MUST return `RECORD_ONLY` Decision.
- The set of span Attributes to be added to the `Span` is the sum of the sets of Attributes as provided by the delegate samplers within their `SamplingResults`s.
- The `TraceState` to be used with the new `Span` is obtained by cumulatively applying all the potential modfications of the parent `TraceState` by the delegate samplers, with special handling of the `th` sub-key (the sampling rejection `THRESHOLD`) for the `ot` entry as described below.

If the final sampling Decision is `DROP` or `RECORD_ONLY`, the `th` entry MUST be removed.
If the sampling Decision is `RECORD_AND_SAMPLE`, and there's no `th` entry in any of the `TraceState` provided by the delegates that decided to `RECORD_AND_SAMPLE`, the `th` entry MUST be also removed.
Otherwise, the resulting `TraceState` MUST contain `th` entry with the `THRESHOLD` value being the maximum of all the `THRESHOLD` values as reported by those delegates that decided to `RECORD_AND_SAMPLE`.

Each delegate sampler MUST be given a chance to participate in the sampling decision as described above and MUST see the same _parent_ state. The order of the delegate samplers does not matter, as long as there's no overlap in the Attribute Keys or the trace state keys (other than `th`) that they use.

## Conjunction

`Conjunction` is a composite sampler which takes two Samplers (delegates) as the arguments. These delegate samplers will be hereby referenced as First and Second.

Upon invocation of its `shouldSample` method, the Conjunction sampler MUST invoke `shouldSample` method on the First sampler, passing the same arguments as received, and examine the received sampling Decision. Upon receiving `DROP` or `RECORD_ONLY` decision it MUST return the SamplingResult from the First sampler, and it MUST NOT proceed with querying the Second sampler. If the sampling decision from the First sampler is `RECORD_AND_SAMPLE`, the Conjunction sampler MUST invoke `shouldSample` method on the Second sampler.
If the sampling Decision from the Second sampler is `RECORD_AND_SAMPLE`, the Conjunction sampler MUST return a `SamplingResult` which is constructed as follows:

- The sampling Decision is `RECORD_AND_SAMPLE`.
- The set of span Attributes to be added to the `Span` is the sum of the sets of Attributes as provided by the First and the Second sampler.
- The `TraceState` to be used with the new `Span` is obtained by cumulatively applying the potential modfications from the First and Second sampler, with special handling of the `th` sub-key (the sampling rejection `THRESHOLD`) for the `ot` entry as described below.

If both First and Second samplers provided `th` entry in the returned `TraceState`, the resulting `TraceState` MUST contain `th` entry with the `THRESHOLD` being maximum of the `THRESHOLD`s provided by the First and Second samplers. Otherwise, the `th` entry MUST be removed. 

If the sampling Decision from the Second sampler is `RECORD_ONLY` or `DROP`, the Conjunction sampler MUST return a `SamplingResult` which is constructed as follows:

- The sampling Decision is `DROP`.
- The set of span Attributes to be added to the `Span` is empty.
- The `TraceState` to be used with the new `Span` is the passed-in `TraceState`, but with the `th` entry removed.


## RuleBased

`RuleBased` is a composite sampler which performs `Span` categorization (e.g. when sampling decision depends on Span attributes) and sampling.
The Spans can be grouped into separate categories, and each category can use a different Sampler.
Categorization of Spans is aided by `Predicates`.

### Predicate

The Predicates represent logical expressions which can access Span Attributes (or anything else available when the sampling decision is to be made), and perform tests on the accessible values.
For example, one can test if the target URL for a SERVER span matches a given pattern.
`Predicate` interface allows users to create custom categories based on information that is available at the time of making the sampling decision.

#### SpanMatches

This is a routine/function/method for `Predicate`, which returns `true` if a given `Span` matches, i.e. belongs to the category described by the Predicate.

#### Required Arguments for Predicates

The arguments represent the values that are made available for `ShouldSample`.

- `Context` with parent Span.
- `TraceId` of the Span to be created.
- Name of the Span to be created.
- Initial set of Attributes of the Span to be created.
- Collection of links that will be associated with the Span to be created.

### Required Arguments for RuleBased

- `SpanKind`
- list of pairs (`Predicate`, `Sampler`)

For making the sampling decision, if the `Span` kind matches the specified kind, the sampler goes through the list in the provided order and calls `SpanMatches` on `Predicate`s passing the `Span` as the argument. If a call returns `true`, the corresponding `Sampler` will be called to make the final sampling decision. If the `SpanKind` does not match, or none of the calls to `SpanMatches` yield `true`, the final decision is `DROP`.

The order of `Predicate`s is essential. If more than one `Predicate` matches a `Span`, only the Sampler associated with the first matching `Predicate` will be used.

# Summary

## Example - sampling configuration 1

Going back to our example of sampling requirements, we can now configure the head sampler to support this particular case, using an informal notation of samplers and their arguments.
First, let's express the requirements for the ROOT spans as follows.

```
S1 = RuleBased(ROOT, {
 (http.target == /healthcheck) => AlwaysOff,
 (http.target == /checkout) => AlwaysOn,
 true => TraceIdRatioBased(0.25)
 })
```

In the next step, we can build the sampler to handle non-root spans as well:

```
S2 = ParentBased(S1)
```

The special case of calling service `/foo` can now be supported by:

```
S3 = AnyOf(S2, RuleBased(CLIENT, { (http.url == /foo) => AlwaysOn })
```

Finally, the last step is to put a limit on the stream of exported spans. One of the available rate limiting sampler that we can use is Jaeger [RateLimitingSampler](https://github.com/open-telemetry/opentelemetry-java/blob/main/sdk-extensions/jaeger-remote-sampler/src/main/java/io/opentelemetry/sdk/extension/trace/jaeger/sampler/RateLimitingSampler.java):

```
S4 = Conjunction(S3, RateLimitingSampler(1000 * 60))
```
## Example - sampling configuration 2

Many users are interested in [Consistent Probability Sampling](https://opentelemetry.io/docs/specs/otel/trace/tracestate-probability-sampling/#consistent-probability-sampler) (CP), as it gives them a chance to calculate span-based metrics even when sampling is active. The configuration presented above uses the traditional samplers, which do not offer this benefit.

Here is how an equivalent configuration can be put together using [CP samplers](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/tracestate-probability-sampling.md#samplers). In this example, the following implementations are used:

- [ConsistentAlwaysOffSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/consistent-sampling/src/main/java/io/opentelemetry/contrib/sampler/consistent56/ConsistentAlwaysOffSampler.java)
- [ConsistentAlwaysOnSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/consistent-sampling/src/main/java/io/opentelemetry/contrib/sampler/consistent56/ConsistentAlwaysOnSampler.java)
- [ConsistentFixedThresholdSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/consistent-sampling/src/main/java/io/opentelemetry/contrib/sampler/consistent56/ConsistentFixedThresholdSampler.java)
- [ConsistentParentBasedSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/consistent-sampling/src/main/java/io/opentelemetry/contrib/sampler/consistent56/ConsistentParentBasedSampler.java)
- [ConsistentRateLimitingSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/consistent-sampling/src/main/java/io/opentelemetry/contrib/sampler/consistent56/ConsistentRateLimitingSampler.java)

```
S = EachOf(
     AnyOf(
       ConsistentParentBased(
         RuleBased(ROOT, {
             (http.target == /healthcheck) => ConsistentAlwaysOff,
             (http.target == /checkout) => ConsistentAlwaysOn,
             true => ConsistentFixedThreshold(0.25)
         }),
       RuleBased(CLIENT, {
         (http.url == /foo) => ConsistentAlwaysOn
       }
     ),
     ConsistentRateLimiting(1000 * 60)
   )
```

## Limitations of composite samplers

Not all samplers can participate as components of composite samplers without undesired or unexpected effects. Some samplers require that they _see_ each `Span` being created, even if the span is going to be dropped. Some samplers update the trace state or maintain internal state, and for their correct behavior it it is assumed that their sampling decisions will be honored by the tracer at the face value in all cases. A good example for this are rate limiting samplers which have to keep track of the rate of created spans and/or the rate of positive sampling decisions.

A special attention is required for CP samplers. The sampling probability they record in trace-state is later used to calculate [_adjusted count_](https://opentelemetry.io/docs/specs/otel/trace/tracestate-probability-sampling/#adjusted-count), which, in turn, is used to calculate span-based metrics. While the composite samplers presented here are compatible with CP samplers, generally, mixing CP samplers with other types of samplers may lead to undefined or sometimes incorrect adjusted counts.

## Prior art

A number of composite samplers are already available as independent contributions
([RuleBasedRoutingSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/samplers/src/main/java/io/opentelemetry/contrib/sampler/RuleBasedRoutingSampler.java),
[LinksBasedSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/samplers/src/main/java/io/opentelemetry/contrib/sampler/LinksBasedSampler.java)).
Also, historically, some Span categorization was introduced by [JaegerRemoteSampler](https://www.jaegertracing.io/docs/1.54/sampling/#remote-sampling).

This proposal aims at generalizing these ideas, and at providing a bit more formal specification for the behavior of the composite samplers.
