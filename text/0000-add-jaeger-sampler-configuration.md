# Include `jaeger` as sampling option

Enable OpenTelemetry to fetch sampling configuration from the Jaeger configuration.

## Motivation

Be able to provide an out-of-the-box solution for sampling defined in the Jaeger Collector. Auto instrumentation will benefit from this solution since you can share the same configuration (e.g: ignore `/health` and `/metrics` paths from sampling) among several services.

## Explanation

Setting the option `otel.traces.sampler` to `"jaeger"` instructs the OpenTelemetry to consider the sampling configuration defined in the Jaeger Collector endpoint.

## Internal details

Use the same value defined for the property `otel.exporter.jaeger.endpoint` in order to fetch the configuration stored in the Jaeger Collection.

## Trade-offs and mitigations

N/A

## Prior art and alternatives

N/A

## Open questions

N/A

## Future possibilities

In the future, this kind of remote sampling configuration could be a standard followed by the OpenTelemetry Collector, Jaeger, any vendor.
