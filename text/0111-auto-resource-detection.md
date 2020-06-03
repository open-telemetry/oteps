# Automatic Resource Detection

Describes a mechanism to support auto-detection of resource information.

## Motivation

Resource information, i.e. attributes associated with the entity producing telemetry, can currently be supplied to tracer and meter providers or appended in custom exporters. In addition to this, it would be useful to have a mechanism to automatically detect resource information from the host (e.g. from an environment variable or from aws, gcp, etc metadata) and apply this to all kinds of telemetry. This will in many cases prevent users from having to manually configure resource information.

Note there are some existing implementations of this already in the SDKs (see [below](#prior-art-and-alternatives)), but nothing currently in the specification.

## Explanation

In order to apply auto-detected resource information to all kinds of telemetry, a user will need to:

- Configure which resource detector(s) they would like to run (e.g. AWS EC2 detector)
- Provide the resource information returned by the configured detector(s) to the relevant tracer or meter provider(s)

Note this means that if the user wants to make use of auto-detected resource information, they will be required to explicitly pass the resource to the trace and/or metric provider(s).

If multiple detectors are configured, and more than one of these successfully detects a resource, the resources will be merged according to the Merge interface already defined in the specification, i.e. the earliest matched resource's attributes will take precedence. Each detector may be run in parallel, but to ensure deterministic results, the resources must be merged in the order the detectors were added.

A default implementation of a detector that reads resource data from the `OTEL_RESOURCE` environment variable will be included in the SDK. The environment variable will contain of a list of key value pairs, and these are expected to be represented in a format similar to the [W3C Correlation-Context](https://github.com/w3c/correlation-context/blob/master/correlation_context/HTTP_HEADER_FORMAT.md#header-value), i.e.: `key1=value1,key2=value2`. This detector must always be configured as the first detector by default.

Custom resource detectors related to specific environments (e.g. specific cloud vendors) must be implemented outside of the SDK, and users will need to import these separately.

In order to allow users to be able to use different detectors for different trace or meter providers, the detector(s) must be configured against a resource provider, and users can create multiple providers if desired.

## Internal details

As described above, the following will be added to the Resource SDK specification:

- An interface for "detectors", to retrieve resource information
- Specification for the resource provider, which can have custom detectors added, and will return a merged resource
- Details of the default "From Environment Variable" detector implementation described above

This is a relatively small proposal so is easiest to explain the details with a code example:

### Usage

The following example in Go creates a trace provider that uses resource information automatically detected from AWS or GCP:

Assumes a dependency has been added on the `otel api`, `otel sdk`, `otel awsdetector`, and `otel gcpdetector` packages.

```go
rp := resource.NewMustProvider() // or NewProvider (see below)
rp.AddDetectors(awsdetector.EC2, gcpdetector.GCE, gcpdetector.GKE)
resource := rp.AutoDetect(ctx)
tp := sdktrace.NewProvider(sdktrace.WithResource(resource))
```

Or more simply:

```go
tp := sdktrace.NewProvider(sdktrace.MustDetectResource(awsdetector.EC2, gcpdetector.GCE, gcpdetector.GKE)) // or DetectResource (see below)
```

In the case of both `WithResource` & `DetectResource` being supplied, the detected resource will be merged with the supplied resource (with the supplied resource taking precedence).

### Components

#### Detector

The Detector interface will simply return a Resource:

```go
type Detector interface {
    Detect(ctx context.Context) (*Resource, error)
}
```

If a detector is not able to detect a resource, it must return an uninitialized resource such that the result of each call to `Detect` can be merged.

#### Provider

A Provider will have a function to add detectors & to retrieve the auto-detected reosurce information:

```go
type Provider struct {
    AddDetectors(detectors ...Detector) { ... }
    AutoDetect(ctx context.Context) (*Resource, error) { ... }
}
```

### Error Handling

In the case of one or more detectors raising an error, there are two reasonable options:

1. Ignore that detector, and continue with a warning (likely meaning we will continue without expected resource information)
2. Crash the application (raise a panic)

These options will be provided as separate interfaces to let the user decide how to recover from failure, e.g. `Provider` & `ProviderMust`

## Trade-offs and mitigations

- The resource provider adds a small amount of complexity that may not be necessary if we can't think of any use cases where a user would want to configure different sets of detectors, but omitting it enforces that restriction. We could add a `global.SetResourceProvider(rp)` function similar to the convention for metrics & traces, but there isn't much value in doing this as the resource provider would only be used during initialization.
- In this proposal, no resource detection will happen by default. It may be preferable to make trace / metric providers use the default resource provider if no resource is supplied. This would be in line with the behaviour of the current Java implementation - see [below](#prior-art-and-alternatives)
- In the case of an error at resource detection time, another alternative would be to start a background thread to retry following some strategy, but it's not clear that there would be much value in doing this, and it would add considerable unnecessary complexity.

## Prior art and alternatives

This proposal is largely inspired by the existing OpenCensus specification, the OpenCensus Go implementation, and the OpenTelemetry JS implementation. For reference, see the relevant section of the [OpenCensus specification](https://github.com/census-instrumentation/opencensus-specs/blob/master/resource/Resource.md#populating-resources)

### Existing OpenTelemetry implementations

- Resource detection implementation in JS SDK [here](https://github.com/open-telemetry/opentelemetry-js/tree/master/packages/opentelemetry-resources): The JS implementation is very similar to this proposal. This proposal adds a resource provider instead of just having a global `DetectResources` function. In addition, vendor specific resource detection code is currently in the resource package, so this would need to be separated.
- Environment variable resource detection in Java SDK [here](https://github.com/open-telemetry/opentelemetry-java/blob/master/sdk/src/main/java/io/opentelemetry/sdk/resources/EnvVarResource.java): This implementation does not currently include a detector interface, and this detector is used by default for trace and meter providers (so this proposal would introduce a breaking change in its current form)

## Open questions

- Does this interfere with any other upcoming specification changes related to resources?
- If custom detectors need to live outside the core repo, what is the expectation regarding where they should be hosted?

## Future possibilities

When the Collector is run as an agent, the same interface, shared with the Go SDK, could be used to append resource information detected from the host to all kinds of telemetry in a Processor (probably as an extension to the existing Resource Processor). This would require a translation from the SDK resource to the collector's internal representation of a resource.
