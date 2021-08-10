# Columnar encoding for the Open Telemetry protocol

This OTEP proposes to extend (in a compatible way) the Open Telemetry protocol with a **generic columnar representation 
for metrics, logs and traces**. This extension will significantly improve the efficiency of the protocol for scenarios 
such as multivariate time-series, large batches of traces and logs.

## Motivation

With the current version of the OpenTelemetry protocol, users are **forced to transform multivariate time-series into a 
collection of univariate time-series** resulting in a large amount of duplication and additional overhead covering the 
entire chain from exporters to backends.

The volume of metrics, traces and logs is constantly increasing to meet new scenarios where accuracy and completeness
of measurements are important. Therefore, the efficiency of the protocol capturing these large telemetry streams is crucial.

The analytical database industry and more recently the stream processing solutions have used columnar encoding methods 
to optimize the processing and the storage of structured data. This proposal aims to leverage this representation to 
make the Open Telemetry protocol more efficient end-to-end.

> Definition: a multivariate time series has more than one time-dependent variable. Each variable depends not only on 
its past values but also has some dependency on other variables. A 3 axis accelerometer reporting 3 metrics simultaneously; 
a mouse move that simultaneously reports the values of x and y, a meteorological weather station reporting temperature, 
cloud cover, dew point, humidity and wind speed; an http transaction chararterized by many interrelated metrics sharing 
the same labels are all common examples of multivariate time-series.

This [benchmark](https://github.com/lquerel/otel-multivariate-time-series/blob/main/README2.md) illustrates in detail
the potential gain we can obtain for a multivariate time-series scenario.

## Explanation

Fundamentally metrics, logs and traces are all structured events occurring at a certain time and optionally covering a 
specified time span. Creating an efficient and generic representation for events will benefit the entire Open Telemetry 
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

A new type of entity 

## Trade-offs and mitigations

## Prior art and alternatives

## Open questions

## Future possibilities

## Appendices

### event.proto
Protobuf specification (draft) for an Arrow-based Open Telemetry event.

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
