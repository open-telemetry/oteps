# OTLP/HTTP: HTTP Transport Extension for OTLP

_Author: Tigran Najaryan, Splunk._

This is a proposal to add HTTP Transport extension for
[OTLP](0035-opentelemetry-protocol.md) (OpenTelemetry Protocol).

## Table of Contents

* [Motivation](#motivation)
* [OTLP/HTTP Protocol Details](#otlphttp-protocol-details)
  * [Request](#request)
  * [Response](#response)
    * [Success](#success)
    * [Failures](#failures)
    * [Throttling](#throttling)
    * [All Other Responses](#all-other-responses)
  * [Connection](#connection)
  * [Parallel Connections](#parallel-connections)
* [Prior Art and Alternatives](#prior-art-and-alternatives)

## Motivation

OTLP can be currently communicated only via one transport: gRPC. While using
gRPC has certain benefits there are also drawbacks:

- Some users have infrastructure limitations that make gRPC-based protocol
  usage impossible. For example AWS ALB does not support gRPC connections.

- gRPC is a relatively big dependency, which some clients are not willing to
  take. Plain HTTP is a smaller dependency and is built in the standard
  libraries of many programming languages.

## OTLP/HTTP Protocol Details

This proposal keeps the existing specification of OTLP over gRPC transport
(OTLP/gRPC for short) and defines an additional way to use OTLP protocol over
HTTP transport (OTLP/HTTP for short). OTLP/HTTP uses the same ProtoBuf payload
that is used by OTLP/gRPC and defines how this payload is communicated over HTTP
transport.

OTLP/HTTP uses HTTP POST requests to send telemetry data from clients to
servers. Implementations MAY use HTTP/1.1 or HTTP/2 transports. Implementations
that use HTTP/2 transport SHOULD fallback to HTTP/1.1 transport if HTTP/2
connection cannot be established.

### Request

Telemetry data is sent via HTTP POST request. The request body is a
ProtoBuf-encoded message `TelemetryRequest` as defined below:

```protobuf
enum RequestType {
    _ = 0;
    TraceExport = 1;
    MetricExport = 2;
}
message TelemetryRequest {
    RequestType request_type = 1;

    // Only one of the following fields is supposed to contain data (determined
    // by `request_type` field). This is deliberately not using ProtoBuf `oneof`
    // for performance reasons (verified by benchmarks).
    ExportTraceServiceRequest export_trace_request = 2;
    ExportMetricsServiceRequest export_metrics_request = 3;
}
```

`TelemetryRequest` references `ExportTraceServiceRequest` and
`ExportMetricsServiceRequest` message types that are defined in
[OTLP ProtoBuf schema](https://github.com/open-telemetry/opentelemetry-proto/tree/master/opentelemetry/proto)
specification.

The client MUST set "Content-Type: application/x-protobuf" request header. The
client MAY gzip the content and in that case SHOULD include "Content-Encoding:
gzip" request header. The client MAY include "Accept-Encoding: gzip" request
header if it can receive gzip-encoded responses.

The default request URL path is `/telemetry` (for example the full URL when
connecting to "example.com" server will be `https://example.com/telemetry`). A
different URL path MAY be configured on the client and server sides.

### Response

#### Success

On success the server MUST respond with `HTTP 200 OK` and include
ProtoBuf-encoded `TelemetryResponse` message in the response body.
`TelemetryResponse` is defined as follows:

```protobuf
enum ResponseType {
    _ = 0;
    TraceExport = 1;
    MetricExport = 2;
}

message TelemetryResponse {
    ResponseType response_type = 1;

    // Only one of the following fields is supposed to contain data (determined
    // by `response_type` field). This is deliberately not using ProtoBuf
    // `oneof` for performance reasons (verified by benchmarks).
    ExportTraceServiceResponse export_trace_response = 2;
    ExportMetricsServiceResponse export_metrics_response = 3;
}
```

`TelemetryResponse` references `ExportTraceServiceResponse` and
`ExportTraceServiceResponse` message types that are defined in [OTLP Protocol
Buffers](https://github.com/open-telemetry/opentelemetry-proto/tree/master/opentelemetry/proto)
specification.

The server MUST set "Content-Type: application/x-protobuf" response header. If
the request header "Accept-Encoding: gzip" is present in the request the server
MAY gzip-encode the response and set "Content-Encoding: gzip" response header.

The server SHOULD respond with success no sooner than after successfully
decoding and validating the request.

#### Failures

If the processing of the request fails because the request contains data that
cannot be decoded or is otherwise invalid and such failure is permanent then the
server MUST respond with `HTTP 400 Bad Request`. The client MUST NOT retry the
request when it receives `HTTP 400 Bad Request` response.

#### Throttling

If the server receives more requests than the client is allowed or the server is
overloaded the server SHOULD respond with `HTTP 429 Too Many Requests` or
`HTTP 503 Service Unavailable` and MAY include "Retry-After" header with a
recommended time interval in seconds to wait before retrying. The client SHOULD
honour the waiting interval specified in "Retry-After" header if it is present.
If the client receives `HTTP 429` or `HTTP 503` response and "Retry-After"
header is not present in the response then the client SHOULD implement an
exponential backoff strategy between retries.

#### All Other Responses

All other HTTP responses that are not explicitly listed in this document should
be treated according to HTTP specification.

If the server disconnects without returning a response the client SHOULD retry
and send the same request. The client SHOULD implement an exponential backoff
strategy between retries to avoid overwhelming the server.

### Connection

If the client is unable to connect to the server the client SHOULD retry the
connection using exponential backoff strategy between retries. The interval
between retries must have a random jitter.

The client SHOULD keep the connection alive between requests. The client SHOULD
use "Connection: Keep-Alive" request header.

Server implementations MAY handle OTLP/gRPC and OTLP/HTTP requests on the same
port and multiplex the connections to the corresponding transport handler based
on "Content-Type" request header.

### Parallel Connections

To achieve higher total throughput the client MAY send requests using several
parallel HTTP connections. In that case the maximum number of parallel
connections SHOULD be configurable.

## Prior Art and Alternatives

I have also considered HTTP/1.1+WebSocket transport. Experimental implementation
of OTLP over WebSocket transport has shown that it typically has better
performance than plain HTTP transport implementation (WebSocket uses less CPU,
higher throughput in high latency connections). However WebSocket transport
requires slightly more complicated implementation and WebSocket libraries are
less ubiquitous than plain HTTP, which may make implementation in certain
languages difficult or impossible.

HTTP/1.1+WebSocket transport may be considered as a future transport for
high-performance use cases as it exhibits better performance than OTLP/gRPC and
OTLP/HTTP.
