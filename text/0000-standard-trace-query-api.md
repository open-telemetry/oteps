# Standardization for Telemetry Data Query API

Vendor agnostic API definition to query telemetry data from observability backends.

## Motivation

- The Observability space is getting mature, more and more telemetry producers (cloud services, libraries, platforms, frameworks, etc.) align with the OpenTelemetry standard to produce telemetry signals.

  On the other side, telemetry consumers platforms (platforms that use telemetry signals to enrich and correlate existing events, i.e [Kiali](https://kiali.io/) [1] or any existing observability platforms that supports integration for different telemetry sources), need to implement multiple proxy services and APIs (each for each observability backend integrated to their platform) to consume and process these telemetry signals.

- Sooner than later, at the same rhythm that the producers side of OpenTelemetry (instrument, collect, export) adoption grows, the need for a standard backend to consume these telemetry signals will grow as well.

_[1] Telemetry consumption platforms like Kiali consume telemetry signals and configurations from different sources to correlate, combine and offer added value to their end user in the Service Mesh domain._

## Explanation

- OpenTelemetry focused on a **vendor agnostic** standard for producing telemetry signals, but there is a gap (or no standard) in a common API to be to query the stored telemetry signals.

  Unfortunately if you want to query traces for a specific platform you'd require some technical dependency on that platform (in the Kiali case, this one is the Jaeger API) which results in **vendor specific** solution.

- This specification should prioritize traces api defenition as this is the most mature capability of OpenTelemetry and has a diverse backend platforms.

## Internal details

This change will be an additional API and SDK where a specific set of common telemetry data query and search capabilites will be defined.
Once a set of common query and search capabilities is defined, a technical specification describing the data exchange options between the telemetry backends and the consumers will be defined along with its delivery protocols.

- This API could leverage the existing protbuf schema, in the case of a trace - the response object can be derived from the [trace definition](https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/trace/v1/trace.proto)
- The API route can complement the existing [OTLP exporter API](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md)

### Example of existing Traces Query API

- [Jaeger Trace Query API Definition](https://github.com/jaegertracing/jaeger-idl/blob/main/proto/api_v2/query.proto)
- [Tempo Trace Query API](https://grafana.com/docs/tempo/latest/api_docs/#query)
- [Tempo Search API (experimental)](https://grafana.com/docs/tempo/latest/api_docs/#query)
- [Zipkin Trace Query API](https://zipkin.io/zipkin-api/#/default/get_traces)

### Comparison Between Different Implementations of Span Response Object

> **_NOTE:_** The objects below were simplified (comments and types were removed)

- [Zipkin Span Object](https://github.com/openzipkin/zipkin-api/blob/7692ca7be4dc3be9225db550d60c4d30e6e9ec59/zipkin-jsonv2.proto#L30)

  ```json
  message Span {
    string trace_id = 1;
    string parent_id = 2;
    string id = 3;
    string kind = 4;
    string name = 5;
    fixed64 timestamp = 6;
    uint64 duration = 7;
    Endpoint local_endpoint = 8;
    Endpoint remote_endpoint = 9;
    repeated Annotation annotations = 10;
    map<string, string> tags = 11;
    bool debug = 12;
    bool shared = 13;
  }
  ```

- [Jaeger Span Object](https://github.com/jaegertracing/jaeger-idl/blob/main/proto/api_v2/model.proto)

  ````json
  message Span {
    bytes trace_id = 1;
    bytes span_id = 2;
    string operation_name = 3;
    repeated SpanRef references = 4;
    uint32 flags = 5;
    google.protobuf.Timestamp start_time = 6;
    google.protobuf.Duration duration = 7;
    repeated KeyValue tags = 8;
    repeated Log logs = 9;
    Process process = 10;
    string process_id = 11;
    repeated string warnings = 12;
  }```
  ````

- Tempo Span Object is not supported (as for the time writing)

From looking at those different and opinionated span obects we can imidiatly see the gap created on the consumer side for trying to map these objects to a common one.

## Trade-offs and mitigations

### Standard API vs. Backend Proxy Library

- In order to make the shift toward this standard mor pleasent from the telemetry consumers side, while reducing the dependency on external platforms, a dedicated contrib repo (i.e. opentelemetry-backend) can be created with a proxy client for that will abstract the details of specific implementations (Jaeger, Tempo, etc.) so that user can change from one observability backend to another while enabling telemetry consumers to query any supported observability backend in a common format.

## Prior art and alternatives

- The proxy library mentioned above can be inspired by this POC made by @lucasponce [https://github.com/lucasponce/jaeger-proto-client](https://github.com/lucasponce/jaeger-proto-client) that can plug any jaeger solution (jaeger/tempo) for consumers.

## Open questions

1. This initiative is not very well aligned with the current vision of OpenTelemetry. Otel in the past mostly stayed away from backend functionality. See [https://opentelemetry.io/docs/](https://opentelemetry.io/docs/) where they say:
   > OpenTelemetry, also known as OTel for short, is a vendor-neutral open-source Observability framework for instrumenting, generating, collecting, and exporting telemetry data such as traces, metrics, logs.
   > See how it stops at "exporting". With some small exceptions OTel consider their job done once the telemetry is delivered to the backend.
   - Does the OpenTelemetry organization is the right place for such an initiative? any other suggestion?
2. One of the challenges we see, is different vendors having different capabilities and APIs. For example, one vendor support searching by arbitrary attributes and aggregations between telemetry signals, etc. Probably some parts of the API should be mandatory, and some optional.
3. We need to find more downstream telemetry consumers to validate this need (some other platforms / users / organizations) other then Kiali.

## Future possibilities

- Potentially extended not only for tracing but other signals (logs, metrics).
- Probably a common set of features can be "standardized," and the OpenTelemetry group may foster a "standard client" for these needs.

```

```
