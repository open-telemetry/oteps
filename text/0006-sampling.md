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
 1. Remove `addLink` APIs from the `Span` interface, and allow recording links only during the span
 construction time.

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

 * They must only depend upon the OpenTelemetry API and not upon the SDK.
 * They are shipping source code that will be linked into others' applications.
 * They have no explicit runtime control over the application.
 * They know some signal about what traces may be interesting (e.g. unusual control plane requests)
 or uninteresting (e.g. health-checks), but have to write fully generically.

**Solution:**

 * On the start Span operation, the OpenTelemetry API will allow marking a span with one of three
 choices for the [SamplingHint](#samplinghint).
 
### Infrastructure package/binary developer
Examples: HBase, Envoy developers.

 * They are shipping self-contained binaries that may accept YAML or similar run-time configuration,
 but are not expected to support extensibility/plugins beyond the default OpenTelemetry SDK, 
 OpenTelemetry SDKTracer, and OpenTelemetry wire format exporter.
 * They may have their own recommendations for sampling rates, but don't run the binaries in
 production, only provide packaged binaries. So their sampling rate configs, and sampling strategies
 need to a finite "built in" set from OpenTelemetry's SDK.
 * They need to deal with upstream sampling decisions made by services calling them.

**Solution:**
 * Allow different sampling strategies by default in OpenTelemetry SDK, all configurable easily via
 YAML or future flags. See [default samplers](#default-samplers).

### Application developer
These are the folks we've been thinking the most about for OpenTelemetry in general.

 * They have full control over the OpenTelemetry implementation or SDK configuration. When using the
 SDK they can configure custom exporters, custom code/samplers, etc.
 * They can choose to implement runtime configuration via a variety of means (e.g. baking in feature
 flags, reading YAML files, etc.), or even configure the library in code.
 * They make heavy usage of OpenTelemetry for instrumenting application-specific behavior, beyond
 what may be provided by the libraries they use such as gRPC, Django, etc.

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
They are the people who provide an implementation for the OpenTelemetry API by using the SDK with
custom `Exporter`s, `Sampler`s, hooks, etc. or by writing a custom implementation, as well as 
running the infrastructure for collecting exported traces.

 * They care about a variety of things, including efficiency, cost effectiveness, and being able to
 gather spans in a way that makes sense for them.

**Solution:**
 * Infrastructure owners receive information attached to the span, after sampling hooks have already
 been run.

## Internal details

### Sampling flags
OpenTelemetry API has two flags/properties:
 * `RecordEvents`
   * This property is exposed in the `Span` interface.
   * If `true` the current `Span` records tracing events (attributes, events, status, etc.), 
   otherwise all tracing events are dropped.
   * Users can use this property to determine if expensive trace events can be avoided.
 * `SampledFlag` 
   * This flag is propagated via the `TraceOptions` (may be renamed to `TraceFlags` in a different
   PR) to the child Spans. For more details see the w3c definition [here][trace-flags].
   * In Dapper based systems this is equivalent to `Span` being `sampled` and exported.
   
The flag combination `SampledFlag == false` and `RecordEvents == true` means that the current `Span`
does record tracing events, but most likely the child `Span` will not. This combination is 
necessary because:
 * Allow users to control recording for individual Spans.
 * OpenCensus has this to support z-pages, so we need to keep backwards compatibility.

The flag combination `SampledFlag == true` and `RecordEvents == false` can cause gaps in the 
distributed trace, and because of this OpenTelemetry API should NOT allow this combination.

It is safe to assume that users of the API should only access the `RecordEvents` property when 
instrumenting code and never access `SampledFlag` unless used in context propagators.

### SamplingHint
This is a new concept added in the OpenTelemetry API that allows to suggest sampling hints to the
implementation of the API:
 * `UNSPECIFIED`
   * This is the default option.
   * No suggestion is made.
 * `NOT_RECORD`
   * Suggest to not `RecordEvents = false` and not propagate `SampledFlag = false`.
 * `RECORD`
   * Suggest `RecordEvents = true` and `SampledFlag = false`.
 * `RECORD_AND_PROPAGATE`
   * Suggest to `RecordEvents = true` and propagate `SampledFlag = true`.

### Sampler interface
The interface for the Sampler class that is available only in the OpenTelemetry SDK:
 * `TraceID`
 * `SpanID`
 * Parent `SpanContext` if any
 * `SamplerHint`
 * `Links`
 * Initial set of `Attributes` for the `Span` being constructed

It produces as an output called `SamplingResult`:
 * A `SamplingDecision` enum [`NOT_RECORD`, `RECORD`, `RECORD_AND_PROPAGATE`].
 * A set of span Attributes that will also be added to the `Span`.
   * These attributes will be added after the initial set of `Attributes`.
 * (under discussion in separate RFC) the SamplingRate float.

### Default Samplers
These are the default samplers implemented in the OpenTelemetry SDK:
 * ALWAYS_ON
   * Ignores all values in SamplingHint.
 * ALWAYS_OFF
   * Ignores all values in SamplingHint.
 * ALWAYS_PARENT
   * Ignores all values in SamplingHint.
   * Trust parent sampling decision (trusting and propagating parent `SampledFlag`).
   * For root Spans (no parent available) returns `NOT_RECORD`.
 * Probability
   * Allows users to configure to ignore or not the SamplingHint for every value different than 
   `UNSPECIFIED`. 
     * Default is to NOT ignore `NOT_RECORD` and `RECORD_AND_PROPAGATE` but ignores `RECORD`.
   * Allows users to configure to ignore the parent `SampledFlag`.
   * Allows users to configure if probability applies only for "root spans", "root spans and remote 
   parent", or "all spans".
     * Default is to apply only for "root spans and remote parent".
     * Remote parent property should be added to the SpanContext see specs [PR/216][specs-pr-216]
   * Sample with 1/N probability
   
**Root Span Decision:**

|Decision|ALWAYS_ON|ALWAYS_OFF|ALWAYS_PARENT|Probability|
|---|---|---|---|---|
|RecordEvents|`True`|`False`|`False`|`SamplingHint==RECORD OR SampledFlag()`|
|SampledFlag|`True`|`False`|`False`|`SamplingHint==RECORD_AND_PROPAGATE OR Probability`|

**Child Span Decision:**

|Decision|ALWAYS_ON|ALWAYS_OFF|ALWAYS_PARENT|Probability|
|---|---|---|---|---|
|RecordEvents|`True`|`False`|`ParentSampledFlag`|`SamplingHint==RECORD OR SampledFlag()`|
|SampledFlag|`True`|`False`|`ParentSampledFlag`|`ParentSampledFlag OR SamplingHint==RECORD_AND_PROPAGATE OR Probability`|

### Links
This RFC proposes that Links will be recorded only during the start `Span` operation, because:
* Link's `SampledFlag` can be used in the sampling decision.
* OpenTracing supports adding references only during the `Span` creation.
* OpenCensus supports adding links at any moment, but this was mostly used to record child Links 
which are not supported in OpenTelemetry.
* Allowing links to be recorded after the sampling decision is made will cause samplers to not 
work correctly and unexpected behaviors for sampling.

## Trade-offs
 * We considered, instead of using the `SpanBuilder`, setting the sampler on the Span constructor,
 and requiring any `Attributes` to be populated prior to the start of the span's default start time.
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

[trace-flags]: https://github.com/w3c/trace-context/blob/master/spec/20-http_header_format.md#trace-flags
[specs-pr-216]: https://github.com/open-telemetry/opentelemetry-specification/pull/216