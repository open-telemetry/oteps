# Sampler data model

Define a generic, extensible data model for trace samplers.

## Motivation

Introducing the ["TraceState: Probability Sampling"](https://github.com/open-telemetry/opentelemetry-specification/blob/v1.11.0/specification/trace/tracestate-probability-sampling.md) specification[^oteps] was the largest advancement in trace signal sampling since the SDK's `Sampler` interface. However, although that work lays a foundation for statistically valid sampling, OpenTelemetry's sampling "story" still has gaps. For one, support for SDK `Sampler`s adjusting behavior based on information received from a file or network socket—termed *remote sampling*—is poor. This and other shortfalls are described in the rest of this section.

[^oteps]:And its OTEP ancestors, [168](https://github.com/open-telemetry/oteps/blob/aafcf0f4daaf027ef841197135edf2c1885afbba/text/trace/0168-sampling-propagation.md) and [170](https://github.com/open-telemetry/oteps/blob/aafcf0f4daaf027ef841197135edf2c1885afbba/text/trace/0170-sampling-probability.md).

After surveying the landscape of sampler configurations today, this OTEP proposes a sampler configuration data model. Early warning: it's not straightforward to compare separate sampling technologies features because different technologies express similar ideas differently. The OpenTelemetry Collector's `tailsampling` config, Jaeger protocol's [sampling strategies](https://www.jaegertracing.io/docs/1.33/sampling/#file-sampling), AWS X-Ray's [SamplingRule](https://docs.aws.amazon.com//xray/latest/api/API_SamplingRule.html), and Honeycomb Refinery's [rule-based](https://docs.honeycomb.io/manage-data-volume/refinery/sampling-methods/#rule-based-sampling) appear distinct, but as we'll see their differences are relatively superficial. This OTEP will propose a sampler data model that can express all that these can, and more.

### Sampling objectives

Collecting trace data is not free. Sampling trace data is a [multi-objective optimization problem](https://en.wikipedia.org/wiki/Multi-objective_optimization), trading off between objectives which can be mutually incompatible. In no particular order, the goals are:

1. Collect as little data as possible.
   1. Reduce or limit costs stemming from the construction and transmission of spans.
   2. Analytics queries are faster when searching less data.
2. Respect limits of downstream storage systems.
   1. Trace storage systems often have data ingest limits (e.g., GBs per second, spans per second, spans per calendar month). The costs of exceeding these limits can be either reduced reliability or increased hosting expenditures.
3. Keep sampling error for statistics of interest within an acceptable range.
   1. "Statistics" can be anything from [RED metrics](https://www.weave.works/blog/the-red-method-key-metrics-for-microservices-architecture/) by service, to data used to answer richer questions like "Which dimensions of trace data are correlated with higher error rate?". You want to ensure that all inferences made from the data you *do* collect are valid.
   2. Setting sampling error targets is akin to setting Service Level Objectives: just as one aspires to build *appropriately reliable* systems, so too one needs statistics which are *just accurate enough* to get valid insights from, but not so accurate that you excessively sacrifice goals #1 and #2.
4. Ensure traces are complete.
   1. "Complete" means that all of the spans belonging to the trace are collected. For more information, see ["Trace completeness"](https://github.com/open-telemetry/opentelemetry-specification/blob/v1.12.0/specification/trace/tracestate-probability-sampling.md#trace-completeness) in the trace spec.

Note: Although goals #1 and #2 can support each other, they are not redundant. Whereas #1 represents a weak constraint of "Prefer less data over more", #2 is a strong constraint: there are some limits on collection that *must not* be exceeded.

### OpenTelemetry does not balance these goals today

Sampling in OTel may happen in both SDKs or in Collector processes. The offerings available in both fall short of balancing the sampling goals.

<adjusted count: “all the following don’t yet support adjusted count, which is bad>

#### SDKs

Importantly, the `Sampler` interface's ShouldSample routine has to return a `Decision` based solely on information available at span creation-time; it cannot consider anything that is only determined after the work of the span has begun (e.g., span duration). This limited knowledge places a low ceiling on how "clever" any `Sampler` could possibly be in pursuit of balancing the goals.

Partly due to that, the two most relevant [built-in samplers](https://github.com/open-telemetry/opentelemetry-specification/blob/031630c818c60666b27764a5b7a0e4ed435f55c4/specification/trace/sdk.md#built-in-samplers), `TraceIdRatioBased` and `JaegerRemoteSampler`, have limitations.

##### TraceIdRatioBased

`TraceIdRatioBased` may be used to consistently sample or drop a certain fixed percentage of spans. The decision is based on a random value, the trace ID, rather than any of the span metadata available to ShouldSample (span name, initial attributes, etc.) As a result,

- when configured with a small ratio this sampler is effective at dropping data but ineffective at minimizing sampling error;
- when configured with a large ratio this sampler is effective at minimizing sampling error but does so at a cost of collecting lots of data

##### JaegerRemoteSampler

[`JaegerRemoteSampler`](https://github.com/open-telemetry/opentelemetry-specification/blob/v1.11.0/specification/trace/sdk.md#jaegerremotesampler), the state of the art for OTel-native remote sampling, considers a given root span's "endpoint" (pair of service name, span name) and decides whether to include the span in the sample by executing a *sampling strategy* ([data model](https://github.com/jaegertracing/jaeger-idl/blob/05fe64e9c305526901f70ff692030b388787e388/proto/api_v2/sampling.proto#L86-L104)) fetched from some out-of-process source. Typically an instance of `JaegerRemoteSampler` will obtain a sampling strategy via API request to a jaeger-collector.[^jaeger-intermediary] How jaeger-collector answers the request depends on its `SAMPLING_CONFIG_TYPE` setting:

[^jaeger-intermediary]:This request may be proxied through an intermediary such as a local jaeger-agent.

- In [`file`](https://www.jaegertracing.io/docs/1.36/sampling/#file-sampling) mode the served strategy is read from a local file containing a JSON description of the strategy. In this mode, each endpoint uses either
  - a `probablistic` sampler: each root span on the endpoint is subject to a [Bernoulli trial](https://en.wikipedia.org/wiki/Bernoulli_trial) with explicitly defined parameter *p*; or
  - a `ratelimiting` sampler: traces are included in the sample as long as an explicitly defined limit on traces per second per `JaegerRemoteSampler` is not exceeded.
- In [`adaptive`](https://www.jaegertracing.io/docs/1.36/sampling/#adaptive-sampling) mode the served strategy is computed. Multiple jaeger-collector instances may collaborate, sharing memory (cf. `SPAN_STORAGE_TYPE`) with the goal of *globally* limiting traces per second per endpoint. During its creation, each root span is subject to a Bernoulli trial with a *p* that was computed by jaeger-collector with this goal in mind.

With respect to the sampling goals, Jaeger remote sampling is strictly superior to `TraceIdRatioBased`, albeit still incomplete; Yuri Shkuro [noted](https://medium.com/jaegertracing/adaptive-sampling-in-jaeger-50f336f4334#2f6c) the following shortcomings in "Adaptive Sampling in Jaeger" (but they apply to `file` mode, too):

- Regarding goal #2: All of Jaeger's limiting is currently at the level of *trace* throughput, not span throughput or data throughput, which are the terms in which most trace store limits are expressed. This impedance mismatch reduces `JaegerRemoteSampler`'s ability to accurately enforce limits imposed by trace stores.
  - This is not a fundamental limitation. The peers that collaborate to determine sampling strategy for overall system could be updated to track trace size statistics (span count, data size) for each endpoint. At that point, alternative limiting parameters like span throughput and data throughput could be supported.
- Regarding goal #3: The sampling strategy [data model](https://github.com/jaegertracing/jaeger-idl/blob/05fe64e9c305526901f70ff692030b388787e388/proto/api_v2/sampling.proto) consumed by `JaegerRemoteSampler` cannot express many useful sampling policies. In particular, sampling behavior for a given trace is a function of service name and span name, at most.
  - This is also not a fundamental limitation. The data model could be extended to support arbitrary span or resource attributes. Note, however, that the span attributes usable in such strategies would still be limited to those whose values are known at the time the trace's root span is being created. Decisions of a Jaeger remote sampler—nor any other SDK `Sampler`—can't be influenced by information that is determined later such as root span duration, delayed root span attributes, or the durations or attributes of descendant spans.

Adopting `JaegerRemoteSampling` can also add significant complexity to a system: to use `adaptive` sampling requires exporting spans to a jaeger-collector. If multiple instances of jaeger-collector exist then a storage system like Apache Cassandra must also be run to provide a shared store to record span traffic statistics. It may make sense for the responsibilities currently carried by jaeger-collector to be moved into the OpenTelemetry Collector.

#### OpenTelemetry Collector

In contrast to the SDKs, Collectors can have access to complete traces.[^tailscale] Two opentelemetry-collector-contrib components are relevant here.

[^tailscale]:Care has to be taken to ensure that all spans in a given trace eventually reach the same Collector instance, but it is possible. See https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/4758 for a discussion of concerns.

##### probabilisticsampler processor

For our purposes, [probabilisticsampler](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/814c3a9e4a3d4d4f8bbba140fed0156616dfa765/processor/probabilisticsamplerprocessor) can be thought of as an in-Collector implementation of the SDK's `TraceIdRatioBased` sampler. It thus has the same shortcomings.

##### tailsampling processor

[tailsampling](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/814c3a9e4a3d4d4f8bbba140fed0156616dfa765/processor/tailsamplingprocessor)'s configuration comprises an array of *policies*. For each trace, every policy is evaluated and their results combined in [a particular way](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/814c3a9e4a3d4d4f8bbba140fed0156616dfa765/processor/tailsamplingprocessor/processor.go#L254-L261) to determine whether to sample or drop the trace.

The processor's most expressive policy is called `composite`. An instance of this policy is defined by a span per second limit and an array of child policies, each of which is given a token bucket with capacity equal to some portion of the span throughout limit. Rather than unconditionally evaluating all child policies (as the processor does with the top-level policies), `composite` does the following when evaluated:

1. Evaluate child policies one by one until one decides to sample. If none does, don't sample.
2. If the child policy who decided to sample has insufficient funds in its token bucket, don't sample.
3. Otherwise, deduct from the child's bucket a number of tokens equal to the number of spans in the trace, and sample.

The usability of `composite` is questionable. Its configuration asks the user to divvy up span throughput among the child policies: "I want to reserve 25% of my throughput for child policy `X`, 15% of my throughput for child policy `Y`," etc. Even though such a configuration does provide a means of better achieving goal #3, this OTEP submits that no user *intuitively* thinks about their system in these terms.

##### Requested: filterprocessor for trace

In https://github.com/open-telemetry/opentelemetry-collector/issues/2310 a user requests a Collector component that would allow them to easily select spans or traces to drop.

### External inspiration

Some other projects deserve mention for their relative effectiveness at balancing the sampling goals.

#### AWS X-Ray

At least two SDKs ([Java](https://github.com/open-telemetry/opentelemetry-java-contrib/tree/1474dff9d906328169f40b428d1816e7f9c57985/aws-xray), [Go](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/a47b4d31dd6ae604fe4cb41747979b5dd01adc65/samplers/aws/xray)) have contrib `Sampler` implementations that base their decisions on data received from [AWS X-Ray](https://github.com/awsdocs/aws-xray-developer-guide/blob/bbe425fbcefc3b8939b666100cfc0e23707e5c45/doc-source/xray-console-sampling.md#sampling-rule-options). Like Jaeger's `adaptive` remote sampling, X-Ray serves advisory sampling policies to clients. An X-Ray based sampling system behaves like so (on average):

1. Define a *[rule](https://docs.aws.amazon.com/xray/latest/api/API_SamplingRule.html)* as a triple: a predicate over span attributes, a token bucket ([e.g.](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/42818333e243682bb50e510f4f91381016f61f71/aws-xray/src/main/java/io/opentelemetry/contrib/awsxray/SamplingRuleApplier.java#L272)), and a number in \[0, 1\] called the rule's *fixed rate.*
2. Define the global sampling policy as an ordered collection of rules.
3. Given a root span in need of a sampling decision,
   1. Match the span to the first rule whose predicate it satisfies.
   2. If the token bucket contains at least 1 token, deduct 1 token from the bucket and sample the span and its descendants.
   3. Else, sample with probability equal to the matched rule's fixed rate.


As the preceding family of policies is strictly more expressive than the class of Jaeger remote sampling policies, X-Ray can more effectively solve for goal #3 (minimize sampling error for a range of statistics). However—and also like Jaeger—X-ray supports limiting data creation rate exclusively in terms of traces per second. For trace stores who impose limits in other terms such as spans per second, X-Ray is ineffective at solving goal #2.

More fundamental than that, though, is that current X-Ray cannot be considered *the* answer to OTel-native sampling on account of X-Ray being commercially-managed and closed-source software.

#### Honeycomb Refinery

OSS, trace-aware sampling proxy application developed on GitHub: https://github.com/honeycombio/refinery.

Check out [sampling types](https://docs.honeycomb.io/manage-data-volume/refinery/sampling-methods/#sampling-types) for examples of more sophisticated sampling designs that better balance the sampling goals above.

Note: A detailed comparison of dynamic samplers is out of scope of this OTEP. The intention is only to establish that dynamic samplers exist, demand for them exists, and that they balance the goals.

Shortcomings:

- Like the jaeger-collector-based solutions, this adds new infra outside of the OpenTelemetry Collector.
- Refinery supports receiving OTLP trace data but only exports via the Honeycomb Events API protocol.

### What does this have to do with sampler configurations?

The one thing all these partial solutions have in common is that they all involve *configuration:* a means of specifying the parameters of their behavior. To build a full solution, a sensible place to start is the foundation: a configuration format that can support the use cases that all the aforementioned prior art has identified.

- sampling only within a Collector (or cluster thereof)
- sampling within SDKs, with policies obtained from a file or network socket

See also:

- https://github.com/open-telemetry/opentelemetry-specification/issues/2085: feature req for remote sampling
- https://github.com/jaegertracing/jaeger/issues/425: Jaeger historical discussion of tail-based sampling.
- [Discuss the possibility of deprecating the tail-based sampling processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/1797)



Notes:

- TODO(Spencer): Look at https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/pkg/telemetryquerylanguage/tql and give feedback. Split out from transformprocessor.
  - Consider cross-platform support for consumption by head-based samplers
  - Maybe head-based sampling has access to so much less data that it doesn’t need as powerful/concise a query language
  - SDK impl: consider extensibility: using host language to write selectors