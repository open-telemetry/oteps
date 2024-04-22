# Composite Samplers Proposal

This proposal addresses head-based sampling as described by the [Open Telemetry SDK](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/sdk.md#sampling).
It introduces additional _composite samplers_. Composite samplers use other samplers (delegates) to make sampling decisions. The composite samplers invoke the delegate samplers, but eventually make the final call.

The new samplers proposed here are mostly compatible with Consistent Probability Samplers. For verbose description of this concept see [probability sampling](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/tracestate-probability-sampling.md). However, the technical details in that document are outdated. For the current proposal see [OTEP 235](https://github.com/open-telemetry/oteps/blob/main/text/trace/0235-sampling-threshold-in-trace-state.md). Also see Draft PR 3910 [Probability Samplers based on W3C Trace Context Level 2](https://github.com/open-telemetry/opentelemetry-specification/pull/3910).

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

We present two quite different approaches to composite samplers. The first one uses only the current [sampling API](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/sdk.md#sampling). It can be applied to a large variety of samplers, but is not very efficient nor elegant.

The second approach is applicable exclusively to Consistent Probability Samplers, but is more efficient and less prone to misconfiguration.

## Approach One

The following new composite samplers are proposed.

### AnyOf

`AnyOf` is a composite sampler which takes a non-empty list of Samplers (delegates) as the argument. The intention is to make `RECORD_AND_SAMPLE` decision if __any of__ the delegates decides to `RECORD_AND_SAMPLE`.

Upon invocation of its `shouldSample` method, it MUST go through the whole list and invoke `shouldSample` method on each delegate sampler, passing the same arguments as received, and collecting the delegates' sampling Decisions.

`AnyOf` sampler MUST return a `SamplingResult` with the following elements.

- If all of the delegate Decisions are `DROP`, the composite sampler MUST return `DROP` Decision as well.
If any of the delegate Decisions is `RECORD_AND_SAMPLE`, the composite sampler MUST return `RECORD_AND_SAMPLE` Decision.
Otherwise, if any of the delegate Decisions is `RECORD_ONLY`, the composite sampler MUST return `RECORD_ONLY` Decision.
- The set of span Attributes to be added to the `Span` is the sum of the sets of Attributes as provided by those delegate samplers which produced a sampling Decision other than `DROP`. In case of conflicting attribute keys, the attribute definition from the last delegate that uses that key takes effect.
- The `TraceState` to be used with the new `Span` is obtained by cumulatively applying all the potential modifications of the parent `TraceState` by the delegate samplers. In case of conflicting entry keys, the entry definition provided by the last delegate that uses that key takes effect. However, the `th` sub-key (the sampling rejection `THRESHOLD`) for the `ot` entry gets special handling as described below.

If the final sampling Decision is `DROP` or `RECORD_ONLY`, the `th` entry MUST be removed.
If the sampling Decision is `RECORD_AND_SAMPLE`, and there's no `th` entry in any of the `TraceState` provided by the delegates that decided to `RECORD_AND_SAMPLE`, the `th` entry MUST be also removed.
Otherwise, the resulting `TraceState` MUST contain `th` entry with the `THRESHOLD` value being the minimum of all the `THRESHOLD` values as reported by those delegates that decided to `RECORD_AND_SAMPLE`.

Each delegate sampler MUST be given a chance to participate in the sampling decision as described above and MUST see the same _parent_ state. The resulting sampling Decision does not depend on the order of the delegate samplers.

### EachOf

`EachOf` is a composite sampler which takes a non-empty list of Samplers (delegates) as the argument. The intention is to make `RECORD_AND_SAMPLE` decision if __each of__ the delegates decides to `RECORD_AND_SAMPLE`.

Upon invocation of its `shouldSample` method, it MUST go through the whole list and invoke `shouldSample` method on each delegate sampler, passing the same arguments as received, and collecting the delegates' sampling Decisions.

`EachOf` sampler MUST return a `SamplingResult` with the following elements.

- If all of the delegate Decisions are `RECORD_AND_SAMPLE`, the composite sampler MUST return `RECORD_AND_SAMPLE` Decision as well.
If any of the delegate Decisions is `DROP`, the composite sampler MUST return `DROP` Decision.
Otherwise, if any of the delegate Decisions is `RECORD_ONLY`, the composite sampler MUST return `RECORD_ONLY` Decision.
- If the resulting sampling Decision is `DROP`, the set of span Attributes to be added to the `Span` is empty. Otherwise, it is the sum of the sets of Attributes as provided by the delegate samplers within their `SamplingResult`s. In case of conflicting attribute keys, the attribute definition from the last delegate that uses that key takes effect.
- The `TraceState` to be used with the new `Span` is obtained by cumulatively applying all the potential modfications of the parent `TraceState` by the delegate samplers. In case of conflicting entry keys, the entry definition provided by the last delegate that uses that key takes effect. However, the `th` sub-key (the sampling rejection `THRESHOLD`) for the `ot` entry gets special handling as described below.

If the final sampling Decision is `DROP` or `RECORD_ONLY`, the `th` entry MUST be removed.
If the sampling Decision is `RECORD_AND_SAMPLE`, and there's no `th` entry in any of the `TraceState` provided by the delegates, the `th` entry MUST be also removed.
Otherwise, the resulting `TraceState` MUST contain `th` entry with the `THRESHOLD` value being the maximum of all the `THRESHOLD` values as reported the delegates.

Each delegate sampler MUST be given a chance to participate in the sampling decision as described above and MUST see the same _parent_ state. The resulting sampling Decision does not depend on the order of the delegate samplers.

### Conjunction

`Conjunction` is a composite sampler which takes two Samplers (delegates) as the arguments. These delegate samplers will be hereby referenced as First and Second. This kind of composition forms conditional chaining of both samplers. 

Upon invocation of its `shouldSample` method, the Conjunction sampler MUST invoke `shouldSample` method on the First sampler, passing the same arguments as received, and examine the received sampling Decision. Upon receiving `DROP` or `RECORD_ONLY` decision it MUST return the SamplingResult from the First sampler, and it MUST NOT proceed with querying the Second sampler. If the sampling decision from the First sampler is `RECORD_AND_SAMPLE`, the Conjunction sampler MUST invoke `shouldSample` method on the Second sampler, effectively passing the `TraceState` received from the First sampler as the parent trace state.

If the sampling Decision from the Second sampler is `RECORD_AND_SAMPLE`, the Conjunction sampler MUST return a `SamplingResult` which is constructed as follows:

- The sampling Decision is `RECORD_AND_SAMPLE`.
- The set of span Attributes to be added to the `Span` is the union of the sets of Attributes as provided by both samplers.
- The `TraceState` to be used with the new `Span` is as provided by the Second sampler, with special handling of the `th` sub-key (the sampling rejection `THRESHOLD`) for the `ot` entry.
If the First sampler did not provide `th` entry in the returned `TraceState`, or if the value of the corresponding `THRESHOLD` is not `0`, then the `th` entry MUST be removed from the resulting `TraceState`.

If the sampling Decision from the Second sampler is `DROP` or `RECORD_ONLY`, the Conjunction sampler MUST return a `SamplingResult` which is constructed as follows:

- The sampling Decision is `DROP`.
- The set of span Attributes to be added to the `Span` is empty.
- The `TraceState` to be used with the new `Span` is the `TraceState` provided by the Second sampler, but with the `th` entry removed.

The `Conjunction` sampler can be useful in a special case where the user wants to keep a group of traces, for example belonging to an end user session, together - meaning to make the same sampling decisions for all traces belonging to the group, as much as possible.
One way of achieving this behavior for consistent probability samplers is to give all traces belonging to the group the same _randomness_ (represented by `r-value`), based on some criteria shared by all traces belonging to the group. This can be done by a special sampler which would provide the required `r-value` for all `ROOT` spans of the involved traces. When using such a sampler as the First delegate for `Conjunction`, this functionality can be encapsulated in a separate sampler, without making any changes to the current SDK specification.

### RuleBased

`RuleBased` is a composite sampler which performs `Span` categorization (e.g. when sampling decision depends on Span attributes) and sampling.
The Spans can be grouped into separate categories, and each category can use a different Sampler.
Categorization of Spans is aided by `Predicates`.

#### Predicate

The Predicates represent logical expressions which can access Span Attributes (or anything else available when the sampling decision is to be made), and perform tests on the accessible values.
For example, one can test if the target URL for a SERVER span matches a given pattern.
`Predicate` interface allows users to create custom categories based on information that is available at the time of making the sampling decision.

##### SpanMatches

This is a routine/function/method for `Predicate`, which returns `true` if a given `Span` matches, i.e. belongs to the category described by the Predicate.

##### Required Arguments for Predicates

The arguments represent the values that are made available for `ShouldSample`.

- `Context` with parent Span.
- `TraceId` of the Span to be created.
- Name of the Span to be created.
- Initial set of Attributes of the Span to be created.
- Collection of links that will be associated with the Span to be created.

#### Required Arguments for RuleBased

- `SpanKind`
- list of pairs (`Predicate`, `Sampler`)

For making the sampling decision, if the `Span` kind matches the specified kind, the sampler goes through the list in the provided order and calls `SpanMatches` on `Predicate`s passing the same arguments as received by `ShouldSample`. If a call returns `true`, the corresponding `Sampler` will be called to make the final sampling decision. If the `SpanKind` does not match, or none of the calls to `SpanMatches` yield `true`, the final decision is `DROP`.

The order of `Predicate`s is essential. If more than one `Predicate` matches a `Span`, only the Sampler associated with the first matching `Predicate` will be used.

## Summary - Approach One

### Example - sampling configuration 1

Going back to our example of sampling requirements, we can now configure the head sampler to support this particular case, using an informal notation of samplers and their arguments.
First, let's express the requirements for the ROOT spans as follows.

```
S1 = RuleBased(ROOT, {
 (http.target == /healthcheck) => AlwaysOff,
 (http.target == /checkout) => AlwaysOn,
 true => TraceIdRatioBased(0.25)
 })
```

Note: technically, `ROOT` is not a Span Kind, but is a special token matching all Spans with invalid parent context (i.e. the ROOT spans, regardless of their kind).

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

### Example - sampling configuration 2

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

### Limitations of composite samplers in Approach One

Not all samplers can participate as components of composite samplers without undesired or unexpected effects. Some samplers require that they _see_ each `Span` being created, even if the span is going to be dropped. Some samplers update the trace state or maintain internal state, and for their correct behavior it it is assumed that their sampling decisions will be honored by the tracer at the face value in all cases. A good example for this are rate limiting samplers which have to keep track of the rate of created spans and/or the rate of positive sampling decisions.

A special attention is required for CP samplers. The sampling probability they record in trace-state is later used to calculate [_adjusted count_](https://opentelemetry.io/docs/specs/otel/trace/tracestate-probability-sampling/#adjusted-count), which, in turn, is used to calculate span-based metrics. While the composite samplers presented here are compatible with CP samplers, generally, mixing CP samplers with other types of samplers may lead to undefined or sometimes incorrect adjusted counts.

## Approach Two

A principle of operation for Approach Two is that `ShouldSample` is invoked only once, on the root of the tree formed by composite samplers. All the logic provided by the composition of samplers is handled by calculating the threshold values, delegating the calculation downstream as necessary.

### Consistent Probability Sampler API

To make this approach possible, all Consistent Probability Samplers need to implement the following API, in addition to the standard Sampler API. This extension will be used by all composite samplers in Approach Two, as listed in the next sections.

#### GetThreshold

This is a routine/function/method for all Consistent Probability Samplers. Its purpose is to query the sampler to provide the rejection threshold value they would use had they been asked to make a sampling decision for a given span, however, without constructing the actual sampling Decision.

#### Required Arguments for GetThreshold:

The arguments are the same as for [`ShouldSample`](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/sdk.md#shouldsample) except for the `TraceId`.

- `Context` with parent Span.
- Name of the Span to be created.
- `SpanKind` of the `Span` to be created.
- Initial set of Attributes of the Span to be created.
- Collection of links that will be associated with the Span to be created.

#### Return value:

The THRESHOLD value from range `0` to `2^56-1` (a 56-bit unsigned integer number) if the sampler is ready to make a probability based sampling decision. Values outside of this range can be used for other situations (such as AllwaysOff decisions), or to eventually support equivalents of sampling decisions other than `DROP` or `RECORD_AND_SAMPLE`.

### ConsistentRuleBased

This composite sampler re-uses the concept of Predicates from Approach One.

#### Required Arguments for ConsistentRuleBased

- `SpanKind`
- list of pairs (`Predicate`, `ConsistentProbabilitySampler`)

For calculating the rejection THRESHOLD, if the `Span` kind matches the specified kind, the sampler goes through the list in the provided order and calls `SpanMatches` on `Predicate`s passing the same arguments as received. If a call returns `true`, the result is as returned by `GetThreshold` called on the corresponding `ConsistentProbabilitySampler`. If the `SpanKind` does not match, or none of the calls to `SpanMatches` yield `true`, the result is obtained by calling `GetThreshold` on `ConsistentAlwaysOffSampler`.

### ConsistentAnyOf

`ConsistentAnyOf` is a composite sampler which takes a non-empty list of ConsistentProbabilitySamplers (delegates) as the argument. The intention is to make a positive sampling decision if __any of__ the delegates would make a positive decision.

Upon invocation of its `GetThreshold` function, it MUST go through the whole list and invoke `GetTheshold` function on each delegate sampler, passing the same arguments as received.

`ConsistentAnyOf` sampler MUST return a THRESHOLD which is constructed as follows:

- If any of the delegates returned a threshold value from the range of `0` to `2^56-1`, the resulting threshold is the minimum value from the set of results from within that range.
- Otherwise, the result is obtained by calling `GetThreshold` on `ConsistentAlwaysOffSampler`.

Each delegate sampler MUST be given a chance to participate in calculating the threshold as described above and MUST see the same argument values. The order of the delegate samplers does not matter.

### ConsistentRateLimiting

`ConsistentRateLimiting` is a composite sampler that helps control the average rate of sampled spans while allowing another sampler (the delegate) to provide sampling hints.

#### Required Arguments for ConsistentRateLimiting

- ConsistentProbabilitySampler (delegate)
- maximum sampling (throughput) target rate

The sampler SHOULD measure and keep the average rate of incoming spans, and therefore also of the desired ratio between the incoming span rate to the target span rate.
Upon invocation of its `GetThreshold` function, the composite sampler MUST get the threshold from the delegate sampler, passing the same arguments as received.
If using the obtained threshold value as the final threshold would entail sampling more spans than the declared target rate, the sampler SHOULD increase the threshold to a value that would meet the target rate. Several algorithms can be used for threshold adjustment, no particular behavior is prescried by the specification though.

When using `ConsistentRateLimiting` in our requirements example as a replacement for `EachOf` and `RateLimiting`, we are left with no use case for any direct equivalent to `EachOf` in Approach Two.

## Summary - Approach Two

### Example - sampling configuration with Approach Two

With the samplers introduced by Approach Two, our example requirements can be coded in a very similar way as with ApproachOne. However, the work of the samplers configured this way forms a tree of `GetThreshold` invocations rather than `ShouldSample` invocations as in ApproachOne.

```
S = ConsistentRateLimiting(
     ConsistentAnyOf(
       ConsistentParentBased(
         ConsistentRuleBased(ROOT, {
             (http.target == /healthcheck) => ConsistentAlwaysOff,
             (http.target == /checkout) => ConsistentAlwaysOn,
             true => ConsistentFixedThreshold(0.25)
         }),
       ConsistentRuleBased(CLIENT, {
         (http.url == /foo) => ConsistentAlwaysOn
       }
     ),
     1000 * 60
   )
```

### Limitations of composite samplers in Approach Two

While making sampling decisions with samplers from Approach Two is more efficient and avoids dealing with non-mainstream cases, it puts some limits on the capabilities of the Consistent Probability Samplers. In particular, a custom CP sampler that wishes to add a span `Attribute` or modify TraceState will be out of luck if it is used as a delegate.

## Prior art

A number of composite samplers are already available as independent contributions
([RuleBasedRoutingSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/samplers/src/main/java/io/opentelemetry/contrib/sampler/RuleBasedRoutingSampler.java),
[LinksBasedSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/samplers/src/main/java/io/opentelemetry/contrib/sampler/LinksBasedSampler.java)).
Also, historically, some Span categorization was introduced by [JaegerRemoteSampler](https://www.jaegertracing.io/docs/1.54/sampling/#remote-sampling).

This proposal aims at generalizing these ideas, and at providing a bit more formal specification for the behavior of the composite samplers.
