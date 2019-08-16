# Sampling API

*Status: proposed*

## TL;DR
This section tries to summarize all the changes proposed in this RFC:
 1. Move the `Sampler` interface from the API to SDK package. Apply some minor changes to the
 `Sampler` API.
 1. Add a new `SamplerHint` concept to the API package.
 1. Add capability to record `Attributes` that can be used for sampling decision during the `Span`
 creation time.
 1. Add capability to start building a `Span` with a delayed `build` method. This is useful for
 cases where some `Attributes` that are useful for sampling are not available when start building
 the `Span`. As an example in Java the current `Span.Builder` will use as a start time for the
 `Span` the moment when the builder is created and not the moment when the `build()` method is
 called.

## Motivation

Different users of OpenTelemetry, ranging from library developers, packaged infrastructure binary
developers, application developers, operators, and telemetry system owners, have separate use cases
for OpenTelemetry that have gotten muddled in the design of the original Sampling API. Thus, we need
to clarify what APIs each should be able to depend upon, and how they will configure sampling and
OpenTelemetry according to their needs.

![Personas](https://i.imgur.com/w1H0CfH.png)

## Explanation

We outline five different use cases (who may be overlapping sets of people), and how they should
interact with OpenTelemetry:

### Library developer
Examples: gRPC, Express, Django developers.

 * They must only depend upon the OTel API and not upon the SDK.
 * They are shipping source code that will be linked into others' applications.
 * They have no explicit runtime control over the application.
 * They know some signal about what traces may be interesting (e.g. unusual control plane requests)
 or uninteresting (e.g. health-checks), but have to write fully generically.

**Solution:**

 * On the start Span operation, the OpenTelemetry API will allow marking a span with one of three
 choices for the SamplingHint, with "don't care" as the default: [`don't care`, `suggest keeping`,
 `suggest discarding`]
 
### Infrastructure package/binary developer
Examples: HBase, Envoy developers.

 * They are shipping self-contained binaries that may accept YAML or similar run-time configuration,
 but are not expected to support extensibility/plugins beyond the default OTel SDK, OTel SDKTracer,
 and OTel wire format exporter.
 * They may have their own recommendations for sampling rates, but don't run the binaries in
 production, only provide packaged binaries. So their sampling rate configs, and sampling strategies
 need to a finite "built in" set from OpenTelemetry's SDK.
 * They need to deal with upstream sampling decisions made by services calling them.

**Solution:**
 * Allow different sampling strategies by default in OTel SDK, all configurable easily via YAML or
 future flags, etc.:
   * Trust parent sampling decision (trusting & propagating parent SpanContext SampleBit)
   * Always keep
   * Never keep
   * Keep with 1/N probability

### Application developer
These are the folks we've been thinking the most about for OTel in general.

 * They have full control over the OTel implementation or SDK configuration. When using the SDK they
 can configure custom exporters, custom code/samplers, etc.
 * They can choose to implement runtime configuration via a variety of means (e.g. baking in feature
 flags, reading YAML files, etc.), or even configure the library in code.
 * They make heavy usage of OTel for instrumenting application-specific behavior, beyond what may be
 provided by the libraries they use such as gRPC, Django, etc.

**Solution:**
 * Allow application developers to link in custom samplers or write their own when using the
 official SDK.
   * These might include dynamic per-field sampling to achieve a target rate
   (e.g. https://github.com/honeycombio/dynsampler-go)
 * Sampling decisions are made within the start Span operation, after attributes relevant to the
 span have been added to the Span start operation but before a concrete Span object exists (so that
 either a NoOpSpan can be made, or an actual Span instance can be produced depending upon the
 sampler's decision).
 * Span.IsRecording() needs to be present to allow costly span attribute/log computation to be
 skipped if the span is a NoOp span.
 
### Application operator
Often the same people as the application developers, but not necessarily
 
 * They care about adjusting sampling rates and strategies to meet operational needs, debugging,
 and cost.
 
**Solution:**
 * Use config files or feature flags written by the application developers to control the
 application sampling logic.
 * Use the config files to configure libraries and infrastructure package behavior.

### Telemetry infrastructure owner
They are the people who provide an implementation for the OTel API by using the SDK with custom
`Exporter`s, `Sampler`s, hooks, etc. or by writing a custom implementation, as well as running the
infrastructure for collecting exported traces.

 * They care about a variety of things, including efficiency, cost effectiveness, and being able to
 gather spans in a way that makes sense for them.

**Solution:**
 * Infrastructure owners receive information attached to the span, after sampling hooks have already
 been run.

## Internal details
The interface for the Sampler class takes in:
 * `TraceID`
 * `SpanID`
 * Parent `SpanContext` if any
 * `Links`
 * Initial set of `Attributes` for the `Span` being constructed

It produces as an output:
* A boolean indicating whether to sample or drop the span.
* The new set of initial span Attributes (or passes along the SpanAttributes unmodified)
* (under discussion in separate RFC) the SamplingRate float.

## Trade-offs
 * We considered, instead of using the `SpanBuilder`, setting the sampler on the Span constructor, and
 requiring any `Attributes` to be populated prior to the start of the span's default start time.
 * We considered, instead of using the `SpanBuilder`, setting the `Sampler` and the `Attributes`
 used for the sampler before running an explicit MakeSamplingDecision() on the span. Attempts to
 create a child of the span would fail if MakeSamplingDecision() had not yet been run.
 * We considered allowing the sampling decision to be arbitrarily delayed.

## Prior art and alternatives
Prior art for Zipkin, and other Dapper based systems: all client-side sampling decisions are made at
head. Thus, we need to retain compatibility with this.

## Open questions
This RFC does not necessarily resolve the question of how to propagate sampling rate values between
different spans and processes. A separate RFC will be opened to cover this case.

## Future possibilities
In the future, we propose that library developers may be able to defer the decision on whether to
recommend the trace be sampled or not sampled until mid-way through execution;

## Related Issues
 * [opentelemetry-specification/189](https://github.com/open-telemetry/opentelemetry-specification/issues/189)
 * [opentelemetry-specification/187](https://github.com/open-telemetry/opentelemetry-specification/issues/187)
 * [opentelemetry-specification/164](https://github.com/open-telemetry/opentelemetry-specification/issues/164)
 * [opentelemetry-specification/125](https://github.com/open-telemetry/opentelemetry-specification/issues/125)
 * [opentelemetry-specification/87](https://github.com/open-telemetry/opentelemetry-specification/issues/87)
 * [opentelemetry-specification/66](https://github.com/open-telemetry/opentelemetry-specification/issues/66)
 * [opentelemetry-specification/65](https://github.com/open-telemetry/opentelemetry-specification/issues/65)
 * [opentelemetry-specification/53](https://github.com/open-telemetry/opentelemetry-specification/issues/53)
 * [opentelemetry-specification/33](https://github.com/open-telemetry/opentelemetry-specification/issues/33)
 * [opentelemetry-specification/32](https://github.com/open-telemetry/opentelemetry-specification/issues/32)
 * [opentelemetry-specification/31](https://github.com/open-telemetry/opentelemetry-specification/issues/31)
