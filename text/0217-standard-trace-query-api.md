# Standard Telemetry Data Query API

Vendor agnostic API definition to query telemetry data (traces, logs and metrics) from observability backends.

## Motivation

- The Observability space is getting mature, more telemetry producers (cloud services, libraries, platforms, frameworks, etc.) align with the OTel standard to produce telemetry signals.

  On the other side, telemetry consumer platforms (platforms that use telemetry signals to enrich and correlate existing events.

### Some examples for Telemetry consumers

- [Kiali](https://kiali.io/) consume telemetry signals and configurations from different sources to correlate, combine and offer added value to their end user in the Service Mesh domain.
- Any existing observability backends like [Tempo](https://grafana.com/oss/tempo/), [Jaeger](https://www.jaegertracing.io/), [SigNoz](https://signoz.io/) and [Zipkin](https://zipkin.io/) which may support integration of different telemetry sources, which will need to implement multiple proxy services and APIs (for each observability backend integrated to their platform) to consume and process these telemetry data.

**Sooner than later, at the same rhythm that the adoption of OTel producers (instrument, collect and export) grows, the need for a standard backend to consume these telemetry signals will grow as well.**

## Explanation

- OTel focus on a **vendor-agnostic** standard for producing telemetry signals, but there is a gap (or no standard) in a common API to query the stored telemetry signals.

  Unfortunately, if you want to query traces for a specific platform you'd require some technical dependency on that platform (in the Kiali case, this one is the Jaeger API) which results in **vendor-specific** solution.

- This suggested specification should prioritize Traces API definition as this is the most mature capability of OTel and consist of great variety of backend platforms.

## Internal details

This change will be an additional API and SDK where a specific set of common Telemetry Data query and search capabilities will be defined.
Once a set of these capabilities will defined, a technical specification describing the data exchange between the telemetry backends and the consumers will be defined along with its delivery protocols (gRPS/HTTP) and schema (Protbuf i.e).

- This API could leverage the existing protbuf schema, in the case of a Trace - the response object can be derived from OTel [Trace Definition](https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/trace/v1/trace.proto)
- The API route can complement the existing [OTLP Exporter API](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md)

### Example of existing Trace Query APIs

- [Jaeger Trace Query API Definition](https://github.com/jaegertracing/jaeger-idl/blob/main/proto/api_v2/query.proto)
- [Tempo Trace Query API](https://grafana.com/docs/tempo/latest/api_docs/#query)
- [Tempo Search API (experimental)](https://grafana.com/docs/tempo/latest/api_docs/#query)
- [Zipkin Trace Query API](https://zipkin.io/zipkin-api/#/default/get_traces)

### Comparison Between Different Implementations of Response Object

> _The objects below were simplified (comments and types were removed). Complete objects are linked._

- [Zipkin Span and Trace Objects](https://github.com/openzipkin/zipkin-api/blob/main/zipkin-jsonv2.proto#L30)

  ```protobuf
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

  // Demonstration of a Trace object as described in the API referance
  message Trace {
    repeated Span = 1;
  }
  ```

- [Jaeger Span and Trace Objects](https://github.com/jaegertracing/jaeger-idl/blob/main/proto/api_v2/model.proto)

  ```protobuf
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
  }

  message Trace {
    message ProcessMapping {
        string process_id = 1;
        Process process = 2;
    }
    repeated Span spans = 1;
    repeated ProcessMapping process_map = 2;
    repeated string warnings = 3;
  }
  ```

- [Tempo Span and Trace Objects](https://grafana.com/docs/tempo/latest/api_docs/) - From [Tempo docs](https://grafana.com/docs/tempo/latest/api_docs/#query) it is not clear what is exactly the structure of a Trace, though we can see a [Go implementation of a proto definition](https://github.com/grafana/tempo/blob/main/pkg/tempopb/trace/v1/trace.pb.go#L307).

From looking at these different, opinionated Span and Trace objects we can immediately see that the burden for mapping these objects to a common format falls on the consumer side.

## Trade-offs and mitigations

### Standard API vs. Backend Proxy Library

- In order to make the shift toward this standard more pleasant from the telemetry consumers side, while reducing the dependency on external platforms, a dedicated contrib repo with a proxy client that will abstract the details of specific implementations (Jaeger, Tempo, etc.) so that the user will be able to change from one observability backend to another while enabling telemetry consumers to query any supported observability backend in a common format.

## Prior art and alternatives

- The proxy library mentioned above can be inspired by this [POC](https://github.com/lucasponce/jaeger-proto-client) made by @lucasponce that can plug any Jaeger solution.
- [Tempo TraceQL](https://github.com/grafana/tempo/blob/main/docs/design-proposals/2022-04%20TraceQL%20Concepts.md), a language for selecting traces that Tempo will implement. This language is currently only focused on trace selection. This is an example for a specific implementation for querying syntax.

## Open questions

1. This initiative is not very well aligned with the current vision of OTel. See [https://opentelemetry.io/docs/](https://opentelemetry.io/docs/) where they say:

   > _OpenTelemetry, also known as OTel for short, is a vendor-neutral open-source Observability framework for instrumenting, generating, collecting, and exporting telemetry data such as traces, metrics, logs._

   - Does the OpenTelemetry organization is the right place for such an initiative? any other suggestions?

2. One of the challenges we see is different vendors having different capabilities and APIs. For example, one vendor support searching by arbitrary attributes and aggregations between telemetry signals, etc. Probably some parts of the API should be mandatory, and some optional.
3. We need to find more downstream telemetry consumers to validate this need (some other platforms/users/organizations).

## Future possibilities

- Define APIs that combine multiple telemetry signals.
- Probably a common set of features can be "standardized," and the OTel group may foster a "standard backend" for these needs.
