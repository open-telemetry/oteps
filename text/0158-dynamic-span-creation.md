# Configuration of Dynamic Span Creation vs Propagation

Add the ability for operators to dynamically configure the level of detail at which instrumentation creates new spans or simply propagates external spans.

## Motivation

The research on distributed tracing is extensive, and many proposals have been made to find a tradeoff between the performance impact of tracing and the expressiveness it delivers.
The current approach to reducing the performance impact of tracing seems to center around sampling strategies.
Operators either choose a static sampling rate, or may use adaptive sampling that changes depending on the amount of traffic their system receives.
Yet other sampling strategies may analyze traces after they have been generated to determine whether to retain them, such as to retain erroneous or anomalous traces specifically.

Sampling, however, has a number of disadvantages as we see it:

- When sampling decisions do not take contextual information into account, we may inadvertently drop traces about interesting error or corner cases.
- The longer the sampling decision is delayed, the less benefit it provides in terms of a reduced performance impact.

We propose an additional feature of OpenTelemetry, where instrumentation libraries can be configured, at runtime, to provide dynamic expressiveness of generated traces.
The motivation is that during normal operation, the system generates spans at a coarse level of detail.
In crisis situations, operators can configure the system to generate traces at an increased level of detail, to support their debugging efforts.
This will ensure a minimal performance impact in day-to-day operations, while an increased performance impact can be accepted for short periods of time during crises.
This is not unlike how logs are specified at various levels of detail, such as DEBUG, INFO, WARN, and ERROR.

We base this motivation on a number of observations:

- In our own work, we have evaluated OpenTelemetry agents with practitioners from two large companies that both operate distributed systems with millions of daily transactions. Upon experiencing the quantity of spans generated, in particular by instrumentation agents and dependency injection strategies, they have expressed a concern with regards to network overhead and data storage. While sampling is a possible remedy, the use cases of our industry partners are such that detailed trace collection is extremely useful only in rare cases, while for the most part they would rather settle on more coarse traces.
- The approach to dynamically adapting the expressiveness of traces has been touched upon in research papers, including [Mace](https://dl.acm.org/doi/10.1145/2815400.2815415), [Castanheira](https://dl.acm.org/doi/10.1145/3426746.3434058), and [Sambasivan](https://dl.acm.org/doi/10.1145/2987550.2987568).

We believe that the ability to dynamically adapt the expressiveness of traces can be beneficial for the following reasons:

- In our own work, we have measured response time overheads of up to 70% when blanket instrumenting services in a microservice application with OpenTelemetry agents (mind you, these are simple services). The impact is negligible when strongly sampling requests. This indicates that the performance overhead of dormant instrumentation should not impact existing data centers, while a performance hit can be accepted for short periods of time in crisis situations.
- [In his Master's Thesis, Carosi](https://www.semanticscholar.org/paper/Protractor%3A-Leveraging-distributed-tracing-in-for-Carosi/708e776d9440abd56006a312168773fdc1ed9667) evaluates the performance impact on tracing applications with service meshes, and finds that the performance impact (in terms of response times) is negligible. For this reason, service meshes can be relied upon for day-to-day tracing at a coarse granularity, while otherwise dormant instrumentation libraries can be activated at need to generate traces at a higher level of detail.

We hope that as OpenTelemetry continues to evolve and is implemented in a growing set of libraries, that it will be possible for operators to dynamically activate and deactivate trace points without requiring recompilation of applications or restarting services.

As mentioned above, this proposal will further benefit applications deployed on platforms that already provide tracing at an infrastructure level.
Such examples are seen in the development of Envoy and various gateways, which already attach trace and span IDs to requests.
For many situations, the level of detail provided by Envoy is sufficient for correlating activities in microservice applications.
In this case, OpenTelemetry libraries could operate in a "dormant" state where no additional spans are created by the application instrumentation or any libraries it uses.
Instead, OpenTelemetry simply forwards tracing headers and performs other activities not related to span creation, such as injecting span contexts into logging libraries.
During crisis situations, operators can dynamically activate the OpenTelemetry instrumentation to acquire additional detail on problematic services, but only for short periods of time.
As Castanheira puts it:

> Despite the increasing number of uses for distributed tracing, it remains, in the wild, focused on the more common and general use cases (such as latency profiling for troubleshooting slow requests).

For this reason, he argues, applications are typically instrumented very coarsely, and perhaps only for a few central workflows. However, he proceeds:

> During crises, operators may be willing to accept a decrease in performance in exchange for a larger wealth of information, but cannot act on this intent because of the static tracing instrumentation.

## Explanation

In order for operators to configure the expressiveness of tracing at runtime, they need a way to communicate with the instrumented applications.
As inspired by Jaeger, this could be implemented by allowing for OpenTelemetry Collectors to inform applications about the desired level of detail, perhaps using polling. We believe that the level of detail (LoD) should be configurable at at least three levels:

- `PROPAGATION_ONLY`: Only propoagate headers, and suppress the creation of all spans. This will work well with service meshes.
- `EXTERNAL_SPANS`: Propagate headers, and generate additional client, server, consumer, and producer spans. This will enhance the traces generated by service meshes, but only when the performance impact can be justified.
- `ALL_SPANS`: Propagate headers, and generate spans at any level of detail

We can imagine the following uses for the three LoDs:

| `PROPAGATION_ONLY`          | `EXTERNAL_SPANS`           | `ALL_SPANS`                  |
| --------------------------- | -------------------------- | ---------------------------- |
| Envoy generates server span |                            |                              |
| OTel span suppressed        | OTel generates server span | OTel generates server span   |
| OTel span suppressed        | OTel span suppressed       | OTel generates internal span |
| OTel span suppressed        | OTel generates client span | OTel generates client span   |
| Envoy generates client span |                            |                              |

## Internal details

This proposal would entail adding a LoD configuration that can be passed to applications via the collector.
During span creation, the instrumentation library would have to consider the current LoD, and decide whether to suppress the creation of spans, similarly to how it may decide whether a span should be sampled.
This change should be compatible with sampling strategies.
That is, the span creation strategy is only taken into consideration if the trace is to be sampled.
If no level of detail has been specified, the default setting should be `ALL_SPANS` for backward compatibility.

We believe that it would be of great value if the collector can change the LoD on at least a service-by-service level.
For instance, operators may configure the LoD as such:

- ServiceA: `PROPAGATION_ONLY`
- ServiceB: `PROPAGATION_ONLY`
- ServiceC: `ALL_SPANS`

Instrumentaiton libraries may poll the collector regularly to learn about their configured LoD.

## Trade-offs and mitigations

One potential issue is that the LoD is communicated to a distributed network of collectors, which may result in the configuration becoming out of sync.
Since this feature is mostly intended for interactive debugging scenarios, it is possible that operators could manually re-transmit their change to LoD when an out-of-sync configuration is observed.
Alternative approaches to ensure synchronized configuration among collectors may be out of scope.

## Prior art and alternatives

We have considered using sampling strategies alone to reduce the overhead of tracing.
We believe that this proposal is an interesting alternative to sampling.
Sampling reduces the overhead by tweaking the number of collected traces.
An alternative dimension is to reduce the size of traces collected.
Collecting service-to-service communication may still suffice for many clients' use cases in terms of collecting metrics and correlating logs.
By reducing the size of traces, we can sample a greater quantity, which improves our chances of having log correlation for problematic workflows, without paying the price for tracing everything.

Furthermore, we have performed a small user test on debugging of a complex system instrumented with the OpenTelemetry Java agent.
Participants felt overwhelmed by the amount of spans generated (workflows with over 300 spans), which further supports the need for dynamic expressiveness.

## Open questions

This OTEP is mostly intended as a basis for further discussion at this point.
We are very much interested in feedback on this proposal.
We are no experts on the internal details of OpenTelemetry, so we do not even know if this is feasible to implement in the current standard.

We would also like to know if this proposal is considered within our outside of the scope of OpenTelemetry.

## Future possibilities

It is possible that the LoD concept could be expanded upon, similarly to how logging APIs work.
For instance, we could distinguish between DEBUG and INFO spans, and dynamically configure their creation at runtime, without recompiling or restarting services.
