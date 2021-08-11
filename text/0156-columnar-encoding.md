# Columnar encoding for the OpenTelemetry protocol

This OTEP proposes to extend (in a compatible way) the Open Telemetry protocol with a **generic columnar representation
for metrics, logs and traces**. This extension will significantly improve the efficiency of the protocol for scenarios
such as multivariate time-series, large batches of traces and logs.

## Motivation

With the current version of the OpenTelemetry protocol, users **will have to transform multivariate time-series into a
collection of univariate time-series** resulting in a large amount of duplication and additional overhead covering the
entire chain from exporters to backends.

As Volume of metrics, traces and logs increases to meet new demands, it is important that we build further optimizations
and protocol efficiencies.

The analytical database industry and more recently the stream processing solutions have used columnar encoding methods
to optimize the processing and the storage of structured data. This proposal aims to leverage this representation to
enhance the OpenTelemetry protocol when handling columnar encoding.

> Definition: a multivariate time series has more than one time-dependent variable. Each variable depends not only on
its past values but also has some dependency on other variables. A 3 axis accelerometer reporting 3 metrics simultaneously;
a mouse move that simultaneously reports the values of x and y, a meteorological weather station reporting temperature,
cloud cover, dew point, humidity and wind speed; an http transaction chararterized by many interrelated metrics sharing
the same labels are all common examples of multivariate time-series.

This [benchmark](https://github.com/lquerel/otel-multivariate-time-series/blob/main/README2.md) illustrates in detail
the potential gain we can obtain for a multivariate time-series scenario.

## Explanation

Fundamentally metrics, logs and traces are all structured events occurring at a certain time and optionally covering a
specified time span. Creating an efficient and generic representation for events will benefit the entire OpenTelemetry
eco-system.

Currently all the OTEL entities are stored in "rows". A metric entity is a protobuf message containing timestamp, labels,
metric value and few other attributes. A batch of metrics behave as a table with multiple rows or metric messages.

Another way to represent the same data is to represent them in columns instead of rows, i.e. a column containing all the
timestamp, a distinct column per label name, a column for the metric values, and so on. This columnar representation is
a proven approach to optimize the creation, size, and processing of data batches. The main benefits of a such approach are:
* better data compression rate (group of similar data)
* faster data processing (better data locality => better use of the CPU cache lines)
* faster serialization and deserialization (less objects to process)
* faster batch creation (less memory allocation)
* better IO efficiency (less data to transmit)

The benefit of this approach increases proportionally with the size of the batches. Using the existing "row-oriented"
representation is well suited for small batch scenarios. Therefore, this proposal suggests to **extend** the protocol by
adding a new columnar event entity to better support multivariate time-series and large batches of events.

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

### Corners cases

Backends that don't support natively multivariate time-series can still automatically transform these events in multiple univariate time-series and operate as usual.

Specialized processors can be developped to group metrics, logs and traces in optimized batch of events in order to connect existing OTEL collectors with backends supporting this new protocol extension.

Specialized processors can be developped to convert batch of events into the existing entities for backends that don't support this protocol extension.

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
