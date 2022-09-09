# Sampler survey
This doc compares the capabilities of popular telemetry sampling systems. The dimensions that are compared:

## Dimensions related to limiting throughput

Temporal resolution: The time range that the limiting occurs on. E.g., limit the number of...
- Spans per second
- Spans per calendar month

Degree of limiting: In a [steady state](https://en.wikipedia.org/wiki/Steady_state) with spans created at a rate R span/s that is *greater than* the desired limit,
- hard limiting: throughput = limit
- soft limiting: E\[throughput\] = limit

Horizontally scalable: Is the desired limit enforced per-sampler, or is it a global limit?
- Yes: Global
- No: Per-sampler

Responsiveness: How quickly does the system return to steady state when perturbed (i.e., when R changes)?

## Other dimensions

Supports statistical estimation: Modifies span metadata such that post hoc analysis can compute unbiased estimates from the data ("count the spans").
- Yes
- No

## Sampling systems

### otelcol's tailsampling processor
- supports estimation: No
- limiting:
	- temporal resolution: Spans per second
	- degree of limiting: Hard
	- horizontally scalable: No
	- responsiveness: < 1 s (token buckets are replenished each second)

The [tailsampling](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/tailsamplingprocessor) processor implements a `ratelimiting` policy ([src](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.47.0/processor/tailsamplingprocessor/internal/sampling/rate_limiting.go#L44-L56)) equivalent to a token bucket with capacity of `spans_per_second` many tokens, replenished every second. Sampling a trace costs `trace.SpanCount` many tokens. Support for updating span p-values has been requested in [#7962](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/7962).

It also has a `composite` policy which is characterized by a sequence of sub-policies, each of which are subject to individual token bucket limiting. Each bucket's capacity is [computed](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.47.0/processor/tailsamplingprocessor/composite_helper.go#L41-L55) as a share of an overall `max_total_spans_per_second`, but otherwise the decisions are identical to those done by `ratelimiting` ([src](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.47.0/processor/tailsamplingprocessor/internal/sampling/composite.go#L89-L134)).

Takes a concept of "allocating bandwidth" (span throughput) to different families of traces. See design doc linked from https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/1410.

If there's more than one otelcol instance in the system, in order to guarantee complete traces you need to somehow guarantee that all spans in a given trace are routed to a given otelcol instance. One way to do that is with the [loadbalancing](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/loadbalancingexporter) exporter.

References
- https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/4758
- [aggregate](https://github.com/grafana/opentelemetry-collector/tree/scale-tail-sampling/processor/aggregateprocessor) processor, described in https://grafana.com/blog/2020/06/18/how-grafana-labs-enables-horizontally-scalable-tail-sampling-in-the-opentelemetry-collector/
- [Issue](https://github.com/open-telemetry/opentelemetry-collector/issues/1724) associated with the loadbalancing exporter
- [Issue](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/1410) associated with the tailsampling processor's `composite` policy

### Jaeger
- supports estimation: Yes, in cases where sampling was probabilistic (including `adaptive` mode); sampling probability is recorded in the span tag `sampler.param`
- limiting (`sampler.type == 'ratelimiting'`):
	- temporal resolution: Traces per second
	- degree of limiting: Hard
	- horizontally scalable: No
	- responsiveness: < 1 s (token buckets are replenished each second)
- limiting (`SAMPLING_CONFIG_TYPE == 'adaptive'`)
	- temporal resolution: Traces per second
	- degree of limiting: Soft (typically) or none (if data is generated at a high enough volume for `--sampling.min-sampling-probability` to overtake `--sampling.target-samples-per-second`)
	- horizontally scalable: Yes
	- responsiveness: Configurable (at most jaeger-client's polling interval + jaeger-collector's `--sampling.calculation-interval`)

Jaeger SDKs (jaeger-client) get sampling policy various ways:
- local: hardcoded `AlwaysOn`, `AlwaysOff`, `probability` (static p), `ratelimiting` (token bucket, parameter: maximum samples per sec). No stratification.
- remote, `file`: per-stratum `probability` or `ratelimiting`. jaeger-collector reloads from filesystem or URL; clients polls jaeger-agent, who proxies requests to jaeger-collector.
- remote, `adaptive`: each stratum as a target throughput + some minimums. jaeger-collector maintains policy based on spans it's *received*; client polls jaeger-agent, who proxies requests to jaeger-collector.
- First two options use local memory for `ratelimiting`. Third option has cluster-level coordination.
- Spans are stratified by a list of priority-ordered rules: (Service name, Span name) > Span name default > (Service name) > global default.
- In `adaptive`, many jaeger-collectors write strata statistics to shared memory. From this data, every jaeger-collector can independently calculate the whole-system stats needed to adjust sampling probabilities. A collector reads statistics (from a configurable number of epochs back; 1 by default), combines them to get whole-cluster strata stats, and recalculates new per-strata sampling probabilities. Defaults:
	- stratum sampling probability: initial (1 in 1,000), minimum (1 in 100,000)
	- stratum throughput: target (1 /s), minimum (1 /min)
- Because collectors *receive* spans, clients don't need to explicitly send statistics themselves (contrast w/ X-Ray, whose sampling and collection APIs are independent)

### AWS X-Ray
- supports estimation: No
- limiting:
	- temporal resolution: Traces per second
	- degree of limiting: Soft
	- horizontally scalable: Yes
	- responsiveness: < 10 s (token buckets are replenished via [GetSamplingTargets](https://docs.aws.amazon.com/xray/latest/devguide/xray-api-sampling.html) requests, which occur every 10 s by default)

Each actor performing sampling sends statistics to a central API describing how many spans it's seen in a period. At least two SDKs ([Java](https://github.com/open-telemetry/opentelemetry-java-contrib/tree/1474dff9d906328169f40b428d1816e7f9c57985/aws-xray), [Go](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/a47b4d31dd6ae604fe4cb41747979b5dd01adc65/samplers/aws/xray)) have contrib `Sampler` implementations that obtain sampling configuration from [AWS X-Ray](https://github.com/awsdocs/aws-xray-developer-guide/blob/bbe425fbcefc3b8939b666100cfc0e23707e5c45/doc-source/xray-console-sampling.md#sampling-rule-options). Like Jaeger's `adaptive` remote sampling, X-Ray serves advisory sampling policies to clients. An X-Ray based sampling system behaves like so (on average):
1. Define a *[rule](https://docs.aws.amazon.com/xray/latest/api/API_SamplingRule.html)* as a triple: a predicate over span attributes, a token bucket ([e.g.](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/42818333e243682bb50e510f4f91381016f61f71/aws-xray/src/main/java/io/opentelemetry/contrib/awsxray/SamplingRuleApplier.java#L272)), and a number in \[0, 1\] called the rule's *fixed rate.*
2. Define the global sampling policy as an ordered collection of rules.
3. Given a root span in need of a sampling decision,
	1. Match the span to the first rule whose predicate it satisfies.
	2. If the token bucket contains at least 1 token, deduct 1 token from the bucket and sample the span and its descendants.
	3. Else, sample with probability equal to the matched rule's fixed rate.

Docs refer to "reservoirs", which are per-rule token buckets ([src](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/42818333e243682bb50e510f4f91381016f61f71/aws-xray/src/main/java/io/opentelemetry/contrib/awsxray/SamplingRuleApplier.java#L272)). Actors doing sampling are dynamically allotted portions of the desired reservoir size (token bucket capacity) called `ReservoirQuota` in the GetSamplingTargets API response ([docs](https://docs.aws.amazon.com/xray/latest/devguide/xray-api-sampling.html)).

References:
- [Design doc for X-Ray `Sampler` implementations](https://docs.google.com/document/d/11V1CDr6eLoMq_3cK1bUm0GzYEyhEiI_83QEJMUFfu9g/edit#)

### Honeycomb Refinery
- supports estimation: Yes, via span attribute `SampleRate` value = N in "1-in-N" (feature request to support p-value [here](https://github.com/honeycombio/husky/issues/77))
- limiting (`EMADynamicSampler`):
	- temporal resolution: Spans per second
	- degree of limiting: Soft
	- horizontally scalable: No (limiting is per Refinery node)
	- responsiveness: Configurable as `AdjustmentInterval`
- limiting (`TotalThroughputSampler`):
	- temporal resolution: Spans per second
	- degree of limiting: Soft, if used correctly (# of strata is small; [src](https://github.com/honeycombio/dynsampler-go/blob/53d08de30228dea6c8b9fcba9fa0253a52de05a8/totalthroughput.go#L16-L19))
	- horizontally scalable: No (limiting is per Refinery node)
	- responsiveness: Configurable as `ClearFrequencySec`

Horizontally scales by forwarding spans to the appropriate node as necessary. The node which ought to handle a given trace is determined via consistent hashing of trace ID ([src](https://github.com/honeycombio/refinery/blob/v1.14.0/route/route.go#L443-L444)). Peers are discovered via either Redis or specified in Refinery's configuration file ([docs](https://docs.honeycomb.io/manage-data-volume/refinery/configuration/#peer-management)).

Not set-it-and-forget-it: as one's system's rate of telemetry production increases over time, either `GoalSampleRate` or their Honeycomb events-per-month quota will need to be adjusted.

## Opinion: Ideal state
- limiting: Support all of both spans per second, spans per month, GB per month (approximation is ok)
- degree of limiting: Soft is ok
- horizontally scalable: Yes
- Prioritize tail sampling in Collector over head sampling in SDK
- Strive for a configuration that is "set it and forget it" (notwithstanding ad hoc changes to aid in investigation or incident response)