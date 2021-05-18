# OTLP over messaging systems

Best practices for transporting OTLP data over a message system.

## Motivation

This proposal tries to bring consistency and guidelines when transporting
OTLP data over messaging systems. A non-exclusive list of examples
of products in this category are:

* Apache Kafka
* Apache Pulsar
* Google Pubsub
* AWS SNS
* RabbitMQ

Using an intermediate messaging system is in some cases preferred compared to
a direct network connection between the telemetry producer and consumer.
Reasons for using such an intermediate medium could be:

* Protecting against backend failure
* Security Policies
* Network Policies
* Buffering

An extra motivation to have a consistent definition of the payload is that
it would be easy to transfer the OTLP data from one messaging system to
another just by reading the payload from one system and writing it to
another without the need for transformations.

## Explanation

Because the OTLP payload that is sent over messaging systems is
well-defined, it’s easy to implement new systems consistently.
An implementation should at least support `otlp_proto`, meaning that the
payload is Protobuf serialized from `ExportTraceServiceRequest` for traces,
`ExportMetricsServiceRequest` for metrics, or `ExportLogsServiceRequest` for
logs.

Optionally an implementation could support `otlp_json` or other alternative
encodings like jaeger or zipkin payloads. Alternative encodings should be configured
by an `encoding` field in the configuration file.

For the user it’s beneficial that both an exporter, and a receiver are
implemented in the OpenTelemetry Collector, but SDK developers could
also implement a language specific exporter.

For log data, implementors of a receiver are encouraged to also support
receiving raw data from a topic. This enables scenarios were
non-OpenTelemetry log producers can produce a stream of logs, either
structured or unstructured, on that topic and the receiver wraps the log
data in a configurable resource.

## Internal details

The default implementation must implement `otlp_proto`, meaning that the
payload is Protobuf serialized from `ExportTraceServiceRequest` for traces,
`ExportMetricsServiceRequest` for metrics, or `ExportLogsServiceRequest` for
logs. If an implementation support other encodings an `encoding` field should
be added to the configuration to make it switchable.

| value | trace | metric | log | description |
|---|---|---|---|---|
| `otlp_proto` (default) | X | X | X | protobuf serialized from the `Export(Trace/Metric/Log)ServiceRequest` |
| `otlp_json` | X | X | X | proto3 json representation of a `Export(Trace/Metric/Log)ServiceRequest` |
| `jaeger_proto` | X | - | - | the payload is deserialized to a single Jaeger proto `Span` |
| `jaeger_json` | X | - | - | the payload is deserialized to a single Jaeger JSON Span using `jsonpb` |
| `zipkin_proto` | X | - | - | the payload is deserialized into a list of Zipkin proto spans |
| `zipkin_json` | X | - | - | the payload is deserialized into a list of Zipkin V2 JSON spans |
| `zipkin_thrift` | X | - | - | the payload is deserialized into a list of Zipkin Thrift spans |
| `raw_string` | - | - | X | see `Log specific collector receiver` |
| `raw_binary` | - | - | X | see `Log specific collector receiver` |

Above you’ll find a non-exclusive list of possible encodings. Only `otlp_proto`
is mandatory and the default.

### Log specific collector receiver

As the flow control is the hardest part of the implementations, adding a feature
for reading raw log events avoids having an additional parallel
implementation todo just that.

The encoding field should be used to indicate that the data received is not the
default `otlp_proto` data, but raw log data. The receiver should construct a valid
OTLP message for each raw message received. Valid encoding are `raw_string`
and `raw_binary`, that will control the type that the data will have when set
in the `body` of the OTLP message.

As the raw log data don’t have a resource attached to them, the receiver should add
a generic resource and instrumentation library message around the raw messages. The
instrumentation library should be set to the name of the receiver and the version to
that of the collector.  Defining the exact resource should be done in the pipeline,
using for example the `resourceprocessor`.

## Prior art and alternatives

The `kafkareceiver` and `kafkaexporter` already implement this OTLP as
described in the OpenTelemetry Collector. The description in the OTEP
makes both of them compliant without modification. Although it
doesn't implement the raw logging support but this could be added
without conflicting with the implementation.

## Open questions

This proposal doesn't take advantage of attributes that some systems
support. Should it? Would it be useful to look at the CloudEvents
spec, to leverage the conventions for attributes?

No guaranty of order could be given, because not all systems support order.
Is that a problem?
