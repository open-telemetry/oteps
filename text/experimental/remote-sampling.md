# Remote Sampling for SDK

This is a proposal to add remotely configurable sampling to OpenTelemetry SDK.

## Motivation

Remote sampling configuration is a useful feature that allows users configuring sampling in a central place (e.g. collector) for
fleet of running OpenTelemetry SDKs. Apart from a central configuration place the remote sampling offers fine-grained
per-operation sampling configuration.

Remote sampling is already widely used and adopted in other open-source tracing systems such as [Jaeger](https://www.jaegertracing.io/docs/1.23/sampling/).

## Explanation

This draft is related to [0121-config-service](https://github.com/open-telemetry/oteps/blob/main/text/experimental/0121-config-service.md)

Goals:

* supply configuration from a remote source e.g. OpenTelemetry collector to SDK
* update configuration in SDK at runtime
* define remote sampler required configuration

## Internal details

```protobuf
// Sampling configuration for a single service.
message ServiceSamplingResponse {
  // Default sampler when no operation_sampling is matched.
  Sampler default_sampler = 1;
  // A list of sampling configurations for a specific operations.
  // First matching configuration is used.
  repeated OperationSampling operation_sampling = 2;
}

// Sampling configuration for an operation.
// Configuration is used only if all properties (span name, span kind, attributes)
// are matched. Matching is done at span creation time.
message OperationSampling {
  // Sampler for this operation.
  Sampler sampler = 1;
  // Span name.
  string span_name = 2;
  // Span kind.
  opentelemetry.proto.trace.v1.Span_SpanKind span_kind = 3;
  // A list of attributes.
  opentelemetry.proto.common.v1.KeyValue attributes = 4;
}

// Sampler configuration. Only build in samplers are supported.
// see https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/sdk.md#built-in-samplers
message Sampler {
  // Sampler name, equivalent to OTEL_TRACES_SAMPLER. Example always_on, traceidratio etc.
  // see https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/sdk-environment-variables.md#general-sdk-configuration
  string name = 1; // TODO this can be enum (always_on, always_off, parentbased_traceidratio...)
  // Sampler argument, equivalent to OTEL_TRACES_SAMPLER_ARG. Example 0.25 for traceidratio sampler.
  string arg = 2;
}

// Request to get sampling configuration.
message RemoteSamplingRequest {
  string service_name = 1;
  // TODO shall we add version and/or environment?
}

// Service that is used to get sampling configuration from collector.
service SamplingManager {
  // Get sampling configuration for a given workload.
  rpc GetSamplingStrategy(RemoteSamplingRequest) returns (ServiceSamplingResponse) {}
}
```

Additional SDK requirements:

* Add `RemoteSampler` to [trace/sdk.md](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/sdk.md#built-in-samplers)
* Add `remote` to `OTEL_TRACES_SAMPLER` and `poolingIntervalMs`, `initialSampler`, `initialSamplerArg` to `OTEL_TRACES_SAMPLER_ARG` [sdk-environment-variables.md](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/sdk-environment-variables.md#general-sdk-configuration)

Remote Sampler configuration:

* pooling interval - pooling interval for getting configuration from remote
* endpoint - address of running service with sampling manager
* initial sampler - initial sampler that is used before the first configuration is fetched

## Trade-offs and mitigations

## Prior art and alternatives

Jaeger client libraries support [remotely controlled sampling](https://www.jaegertracing.io/docs/1.23/sampling/).

The sampling configuration is divided into two parts: service and default strategy.
The service strategy defines sampling at a service or operation level. The default strategy is
used when no service was matched.

Follows an example JSON configuration for the Jaeger Collector:

```json
{
  "service_strategies": [
    {
      "service": "foo",
      "type": "probabilistic",
      "param": 0.8,
      "operation_strategies": [
        {
          "operation": "op1",
          "type": "probabilistic",
          "param": 0.2
        },
        {
          "operation": "op2",
          "type": "probabilistic",
          "param": 0.4
        }
      ]
    },
    {
      "service": "bar",
      "type": "ratelimiting",
      "param": 5
    }
  ],
  "default_strategy": {
    "type": "probabilistic",
    "param": 0.5,
    "operation_strategies": [
      {
        "operation": "/health",
        "type": "probabilistic",
        "param": 0.0
      },
      {
        "operation": "/metrics",
        "type": "probabilistic",
        "param": 0.0
      }
    ]
  }
}
```

## Open questions

* Should we support regex match on the span name or attributes? It can be useful for matching `http.url` attribute.
* Should be sampler name enum? Keeping it as string allows using custom samplers.
* Should the sampling request have more parameters apart from the service name. The parameters could be service version, environment name. It might be desirable to use a different sampling configuration for the same service in different environments.

## Future possibilities
