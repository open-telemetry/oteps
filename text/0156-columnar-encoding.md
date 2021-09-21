# Columnar encoding for the OpenTelemetry protocol

This OTEP proposes to extend (in a compatible way) the OpenTelemetry protocol with a **generic columnar representation
for metrics, logs and traces**. This extension will significantly improve the efficiency of the protocol for scenarios
such as multivariate time-series, large batches of traces and logs.

## Motivation

With the current version of the OpenTelemetry protocol, users **will have to transform multivariate time-series into a
collection of univariate time-series** resulting in a large amount of duplication and additional overhead covering the
entire chain from exporters to backends.

As Volume of metrics, traces and logs increases to meet new demands, it is important that we build further optimizations
and protocol efficiencies. CPU, memory, and network bandwidth must be optimized to minimize the impact on devices,
processes and overall cost of collecting and transmitting large streams of telemetry data.

The analytical database industry and more recently the stream processing solutions have used columnar encoding methods
to optimize the processing and the storage of structured data. This proposal aims to leverage this representation to
enhance the OpenTelemetry protocol when handling columnar encoding.

> Definition: a multivariate time series has more than one time-dependent variable. Each variable depends not only on
its past values but also has some dependency on other variables. A 3 axis accelerometer reporting 3 metrics simultaneously;
a mouse move that simultaneously reports the values of x and y, a meteorological weather station reporting temperature,
cloud cover, dew point, humidity and wind speed; an http transaction chararterized by many interrelated metrics sharing
the same attributes are all common examples of multivariate time-series.

This [benchmark](https://github.com/lquerel/otel-multivariate-time-series/blob/main/README2.md) illustrates in detail
the potential gain we can obtain for a multivariate time-series scenario. To represent different scenarios, this
benchmark has been performed with the following batch size: 10, 100, 500, 1000, 5000, 10000.

## Explanation

Fundamentally metrics, logs and traces are all structured events occurring at a certain time and optionally covering a
specified time span. Creating an efficient and generic representation for events will benefit the entire OpenTelemetry
eco-system.

Currently all the OTEL entities are stored in "rows". A metric entity is a protobuf message containing timestamp, attributes,
metric value and few other attributes. A batch of metrics behave as a table with multiple rows or metric messages.

Another way to represent the same data is to represent them in columns instead of rows, i.e. a column containing all the
timestamp, a distinct column per attribute name, a column for the metric values, and so on. This columnar representation is
a proven approach to optimize the creation, size, and processing of data batches. The main benefits of a such approach are:

* better data compression rate (group of similar data)
* faster data processing (better data locality => better use of the CPU cache lines)
* faster serialization and deserialization (less objects to process)
* faster batch creation (less memory allocation)
* better IO efficiency (less data to transmit)

The benefit of this approach increases proportionally with the size of the batches. Using the existing "row-oriented"
representation is well suited for small batch scenarios. Therefore, this proposal suggests to **extend** the protocol by
adding a new columnar event entity to better support multivariate time-series and large batches of logs and traces.

## Internal details

In addition to the 3 existing resource entities (metrics, logs and traces), we introduce the resource event entity (see [this protobuf specification](#event-proto)
for more details) encoding a batch of events into an Apache Arrow buffer. [Apache Arrow](https://arrow.apache.org/) is
a language-independent columnar memory format for flat and hierarchical data, organized for efficient analytic operations
on modern hardware like CPUs and GPUs. As demonstrated by our [benchmark](https://github.com/lquerel/otel-multivariate-time-series/blob/main/README2.md)
leveraging Apache Arrow will give us access to a mature solution optimized for our exact need as well as a broader ecosystem.

![resource-events](img/0156-resource-events.svg)

Efficient implementation of Apache Arrow exists for most of the languages (Java, Go, C++, Rust, ...). Connectors with Apache Arrow
buffer exist for well-known file format (e.g. Parquet) and for well-known backend (e.g. BigQuery). By reusing this existing infrastructure,
we accelerate the development of the OpenTelemetry protocol while expanding its field of application.

![arrow-ecosystem](img/0156-arrow-ecosystem.svg)

### OpenTelemetry entities to Arrow mapping

Apache Arrow is a general-purpose in-memory columnar format with a fast serialization and deserialization support. All
Arrow entities must be defined with a schema. For OpenTelemetry the following Arrow schemas are proposed to map the
existing entities i.e. metrics, logs and traces. **By fixing these schemas we allow all the collector components such as receivers,
processors, and exporters to encode and decode efficiently these telemetry streams.** Nothing will prevent a processor
supporting this protocol extension to filter, aggregate, project Arrow buffers. To do so, the processor will be able to
leverage the processing capability of Apache Arrow to directly process in-situ and with minimum of memory copy the
batches (some Arrow frameworks support SIMD accelerations and SQL data processing).
Alternatively, and although this is not the most optimal approach, it will be entirely possible for processors that
already have a row-oriented processing framework to convert Arrow buffers into a stream of metrics, logs, or traces to
apply their existing processing capabilities, and then rebuild the Arrow buffers before to send the message to other
intermediaries or backends.

By storing Arrow buffers in a protobuf field of type 'bytes' we can leverage the zero-copy capability of some Protobuf
implementations (e.g. C++, Java, Rust) in order to get the most out of Arrow (relying on zero-copy ser/deser framework).

Null values support can be configured per field (third parameter in the field description). When enabled, a validity map
mechanism is used.

> Please read this documentation to get a complete description of the [Arrow Memory Layout](https://arrow.apache.org/docs/format/Columnar.html#format-columnar).
> Notation: For the next 3 sections, the schema definition uses the following notation to declare fields -> Field::new(name, type, nullable)

#### OpenTelemetry metrics to Arrow mapping

This Arrow schema describes the representation of univariate and **multivariate** time-series.

Attributes are mapped to a set of dedicated column scoped by the logical struct 'attributes'. The list of attributes can be easily
determined from the schema by any participants. Metrics follow the same organization and are scoped by the logical
struct 'metrics'.

Exemplars are encoded with a list of structs instead of a struct of fields (columns). This mapping is
based on the assumption that the number of exemplars can vary greatly from one measurement point to another
for the same time-series. Arrow encodes this type representation with an offsets buffer and a child/data array.
If this assumption is not true, another option will be to declare one column per attribute and exemplar with a validity
bitmap enabled. **This should be validated by experimentation on realistic datasets.**

For more details on the Arrow Memory Layout see this [document](https://arrow.apache.org/docs/format/Columnar.html#variable-size-binary-layout).

Note: in Arrow schemas, each field declaration is a triplet following this syntax `Field::new(name, type, semantically_nullable)`.

```rust
// Multivariate time-series schema declaration
Schema::new(vec![
    // Time range
    Field::new("start_time_unix_nano", DataType::UInt64, false),
    Field::new("time_unix_nano", DataType::UInt64, false),

    // Attributes
    Field::new(
            "attributes",
            DataType::Struct(vec![
                Field::new("attribute_1", DataType::Utf8, false),
                Field::new("attribute_2", DataType::Utf8, false),
                // ...
            ]),
            true,
    ),

    // Metrics
    // Only one metric in the structure if this schema represent a univariate time-series.
    // Multiple fields in the structure to represent multivariate time-series.
    Field::new(
        "metrics",
        DataType::Struct(vec![
            Field::new("metric_1", DataType::Int64, false),
            Field::new("metric_2", DataType::Float64, false),
            // ...
        ]),
        false,
    ),

    // Exemplars
    Field::new("exemplars", DataType::List(
        Box::new(Field::new(
            "exemplar",
            DataType::Struct(vec![
                Field::new("filtered_attributes", DataType::List(
                    Box::new(Field::new(
                        "attribute",
                        DataType::Struct(vec![
                            Field::new("name", DataType::Utf8, false),
                            Field::new("value", DataType::Utf8, false),
                        ]),
                        true,
                    ))
                ), true),
                Field::new("filtered_attributes", DataType::List(
                    Box::new(Field::new(
                        "attribute",
                        DataType::Struct(vec![
                            Field::new("name", DataType::Utf8, false),
                            Field::new("value", DataType::Utf8, false),
                        ]),
                        true,
                    ))
                ), true),
                Field::new("time_unix_nano", DataType::UInt64, false),
                // Could be Float64 or Int64
                Field::new("value", DataType::Float64, false),
                Field::new("span_id", DataType::Binary, true),
                Field::new("trace_id", DataType::Binary, true),
            ]),
            true,
        ))
    ), true),
])
```

#### OpenTelemetry logs to Arrow mapping

This Arrow schema describes the representation of logs.

Attributes are mapped to a set of dedicated column scoped by the logical struct 'attributes'. The list of attributes can be easily
determined from the schema by any participants.

For more details on the Arrow Memory Layout see this [document](https://arrow.apache.org/docs/format/Columnar.html#variable-size-binary-layout).

```rust
// Log schema declaration
Schema::new(vec![
    // Timestamp
    Field::new("time_unix_nano", DataType::UInt64, false),

    Field::new("severity_number", DataType::UInt8, false),
    Field::new("severity_text", DataType::Utf8, true),
    Field::new("name", DataType::Utf8, false),
    Field::new("body", DataType::Utf8, true),
    Field::new("flags", DataType::Int32, true),
    Field::new("span_id", DataType::Binary, true),
    Field::new("trace_id", DataType::Binary, true),

    // Attributes
    Field::new(
        "attributes",
        DataType::Struct(vec![
            Field::new("attribute_1", DataType::Utf8, false),
            Field::new("attribute_2", DataType::Utf8, false),
            // ...
        ]),
        true,
    ),
])
```

#### OpenTelemetry traces to Arrow mapping

This Arrow schema describes the representation of traces.

Attributes are mapped to a set of dedicated column scoped by the logical struct 'attributes'. The list of attributes can be easily
determined from the schema by any participants.

Attributes with a list of structs instead of a struct of fields (columns). This mapping is based on the assumption that
the number of attributes can vary greatly from one log entry to another for the same log stream. Arrow encodes this type
representation with an offsets buffer and a child/data array. If this assumption is not true, another option will be to
declare one column per attribute with a validity bitmap enabled. **This should be validated by experimentation on
realistic datasets.**

Events and links follow the same representation used for the attributes.

For more details on the Arrow Memory Layout see this [document](https://arrow.apache.org/docs/format/Columnar.html#variable-size-binary-layout).

```rust
// Trace schema declaration
Schema::new(vec![
    // Time range
    Field::new("start_time_unix_nano", DataType::UInt64, false),
    Field::new("end_time_unix_nano", DataType::UInt64, false),

    Field::new("trace_id", DataType::Binary, false),
    Field::new("span_id", DataType::Binary, false),
    Field::new("trace_state", DataType::Utf8, true),
    Field::new("parent_span_id", DataType::Binary, true),
    Field::new("name", DataType::Utf8, false),
    Field::new("kind", DataType::UInt8, false),

    // Attributes
    Field::new(
        "attributes",
        DataType::Struct(vec![
            Field::new("attribute_1", DataType::Utf8, false),
            Field::new("attribute_2", DataType::Utf8, false),
            // ...
        ]),
        true,
    ),

    // Events
    Field::new("events", DataType::List(
        Box::new(Field::new(
            "event",
            DataType::Struct(vec![
                Field::new("time_unix_nano", DataType::UInt64, false),
                Field::new("name", DataType::Utf8, false),
                Field::new("attributes", DataType::List(
                    Box::new(Field::new(
                        "attribute",
                        DataType::Struct(vec![
                            Field::new("name", DataType::Utf8, false),
                            Field::new("value", DataType::Utf8, false),
                        ]),
                        true,
                    ))
                ), true),
            ]),
            true,
        ))
    ), true),

    // Links
    Field::new("links", DataType::List(
        Box::new(Field::new(
            "link",
            DataType::Struct(vec![
                Field::new("trace_id", DataType::Binary, false),
                Field::new("span_id", DataType::Binary, false),
                Field::new("trace_state", DataType::Utf8, false),
                Field::new("attributes", DataType::List(
                    Box::new(Field::new(
                        "attribute",
                        DataType::Struct(vec![
                            Field::new("name", DataType::Utf8, false),
                            Field::new("value", DataType::Utf8, false),
                        ]),
                        true,
                    ))
                ), true),
            ]),
            true,
        ))
    ), true),
])
```

### Corners cases

Backends that don't support natively multivariate time-series can still automatically transform these events in multiple univariate time-series and operate as usual.

Specialized processors can be developed to group metrics, logs and traces in optimized batch of events in order to connect existing OTEL collectors with backends supporting this new protocol extension.

Specialized processors can be developed to convert batch of events into the existing entities for backends that don't support this protocol extension.

## Trade-offs and mitigations

A columnar-oriented protocol is not necessarily desirable for all scenarios (e.g. devices that do not have the resources to accumulate data in batches). The proposed mixed solution allows to better address these different scenarios.

Implementing a new columnar format is complex and costly (multiple language implementations, tests, optimizations, industry adoption). Reusing Apache Arrow is an interesting approach to mitigate this issue.

## Prior art and alternatives

* [Column-oriented DBMS](https://en.wikipedia.org/wiki/Column-oriented_DBMS)
* [Apache Arrow](https://arrow.apache.org/)

## Open questions

More work needs to be done around examplars and histograms representation.

More discussions should happen on the processor and storage layers. This approach could simplify significantly the design of OTEL compatible stream processing and database systems.

## Future possibilities

* Leverage Apache Arrow dictionary

## Appendices

### event proto

Protobuf specification (draft) for an Arrow-based OpenTelemetry event.

```protobuf
syntax = "proto3";

package opentelemetry.proto.arrow_events.v1;

import "opentelemetry/proto/common/v1/common.proto";
import "opentelemetry/proto/resource/v1/resource.proto";

option java_multiple_files = true;
option java_package = "io.opentelemetry.proto.events.v1";
option java_outer_classname = "EventsProto";
option go_package = "github.com/open-telemetry/opentelemetry-proto/gen/go/events/v1";

// A column-oriented collection of events from a Resource.
message ResourceEvents {
  // The resource for the events in this message.
  // If this field is not set then no resource info is known.
  opentelemetry.proto.resource.v1.Resource resource = 1;

  // A list of events that originate from a resource.
  repeated InstrumentationLibraryEvents instrumentation_library_events = 2;

  // This schema_url applies to the data in the "resource" field. It does not apply
  // to the data in the "instrumentation_library_events" field which have their own
  // schema_url field.
  string schema_url = 3;
}

// A collection of Events produced by an InstrumentationLibrary.
message InstrumentationLibraryEvents {
  // The instrumentation library information for the events in this message.
  // Semantically when InstrumentationLibrary isn't set, it is equivalent with
  // an empty instrumentation library name (unknown).
  opentelemetry.proto.common.v1.InstrumentationLibrary instrumentation_library = 1;

  // A list of batch of events that originate from an instrumentation library.
  repeated BatchEvent batches = 2;

  // dropped_events_count is the number of dropped events. If the value is 0, then no
  // events were dropped.
  uint32 dropped_events_count = 3;
}

// A typed collection of events with a columnar-oriented representation.
// All the events in this collection share the same schema url.
message BatchEvent {
  // This schema_url applies to all events in this batch.
  string schema_url = 1;

  uint32 size = 2;

  bytes  arrow_buffer = 5;
}
```
