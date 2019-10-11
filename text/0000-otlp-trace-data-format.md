# OTLP Trace Data Format

_Author: Tigran Najaryan, Splunk_

**Status:** `proposed`

OTLP Trace Data Format specification describes the structure of the trace data that is transported by OpenTelemetry Protocol (RFC0035).

## Motivation

This document is a continuation of OpenTelemetry Protocol RFC0035 and is necessary part of OTLP specification.

## Explanation

OTLP Trace Data Format is primarily inherited from OpenCensus protocol. Several changes are introduced with the goal of more efficient serialization. Notable differences from OpenCensus protocol are:

1. Removed `Node` as a concept.
2. Extended `Resource` to better describe the source of the telemetry data.
3. Replaced attribute maps by lists of key/value pairs.
4. Eliminated unnecessary additional nesting in various values.

Changes 1-2 are conceptual, changes 3-4 improve performance.

## Internal details

This section specifies data format in Protocol Buffers.

### Resource

```
// Resource information. This describes the source of telemetry data.
message Resource {
  // Set of labels that describe the resource. See OpenTelemetry specification
  // semantic conventions for standardized label names:
  // https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/data-semantic-conventions.md

  repeated AttributeKeyValue labels = 3;
  int32 dropped_labels_count = 11;
}
```

### Span

```
// A span represents a single operation within a trace. Spans can be
// nested to form a trace tree. Spans may also be linked to other spans
// from the same or different trace. And form graphs. Often, a trace
// contains a root span that describes the end-to-end latency, and one
// or more subspans for its sub-operations. A trace can also contain
// multiple root spans, or none at all. Spans do not need to be
// contiguous - there may be gaps or overlaps between spans in a trace.
//
// The next id is 17.
message Span {
  // A unique identifier for a trace. All spans from the same trace share
  // the same `trace_id`. The ID is a 16-byte array. An ID with all zeroes
  // is considered invalid.
  //
  // This field is semantically required. Receiver should generate new
  // random trace_id if empty or invalid trace_id was received.
  //
  // This field is required.
  bytes trace_id = 1;

  // A unique identifier for a span within a trace, assigned when the span
  // is created. The ID is an 8-byte array. An ID with all zeroes is considered
  // invalid.
  //
  // This field is semantically required. Receiver should generate new
  // random span_id if empty or invalid span_id was received.
  //
  // This field is required.
  bytes span_id = 2;

  // This field conveys information about request position in multiple distributed tracing graphs.
  // It is a list of Tracestate.Entry with a maximum of 32 members in the list.
  //
  // See the https://github.com/w3c/distributed-tracing for more details about this field.
  message Tracestate {
    message Entry {
      // The key must begin with a lowercase letter, and can only contain
      // lowercase letters 'a'-'z', digits '0'-'9', underscores '_', dashes
      // '-', asterisks '*', and forward slashes '/'.
      string key = 1;

      // The value is opaque string up to 256 characters printable ASCII
      // RFC0020 characters (i.e., the range 0x20 to 0x7E) except ',' and '='.
      // Note that this also excludes tabs, newlines, carriage returns, etc.
      string value = 2;
    }

    // A list of entries that represent the Tracestate.
    repeated Entry entries = 1;
  }

  // The Tracestate on the span.
  Tracestate tracestate = 3;

  // The `span_id` of this span's parent span. If this is a root span, then this
  // field must be empty. The ID is an 8-byte array.
  bytes parent_span_id = 4;

  // An optional resource that is associated with this span. If not set, this span
  // should be part of a ResourceSpan that does include the resource information, unless resource
  // information is unknown.
  Resource resource = 5;

  // A description of the span's operation.
  //
  // For example, the name can be a qualified method name or a file name
  // and a line number where the operation is called. A best practice is to use
  // the same display name at the same call point in an application.
  // This makes it easier to correlate spans in different traces.
  //
  // This field is semantically required to be set to non-empty string.
  // When null or empty string received - receiver may use string "name"
  // as a replacement. There might be smarted algorithms implemented by
  // receiver to fix the empty span name.
  //
  // This field is required.
  string name = 6;

  // Type of span. Can be used to specify additional relationships between spans
  // in addition to a parent/child relationship.
  enum SpanKind {
    // Unspecified. Do NOT use as default.
    // Implementations MAY assume SpanKind to be INTERNAL when receiving UNSPECIFIED.
    SPAN_KIND_UNSPECIFIED = 0;

    // Indicates that the span is used internally. Default value.
    INTERNAL = 1;

    // Indicates that the span covers server-side handling of an RPC or other
    // remote network request.
    SERVER = 2;

    // Indicates that the span covers the client-side wrapper around an RPC or
    // other remote request.
    CLIENT = 3;

    // Indicates that the span describes producer sending a message to a broker.
    // Unlike client and  server, there is no direct critical path latency relationship
    // between producer and consumer spans.
    PRODUCER = 4;

    // Indicates that the span describes consumer receiving a message from a broker.
    // Unlike client and  server, there is no direct critical path latency relationship
    // between producer and consumer spans.
    CONSUMER = 5;
  }

  // Distinguishes between spans generated in a particular context. For example,
  // two spans with the same name may be distinguished using `CLIENT` (caller)
  // and `SERVER` (callee) to identify queueing latency associated with the span.
  SpanKind kind = 7;

  // The start time of the span. On the client side, this is the time kept by
  // the local machine where the span execution starts. On the server side, this
  // is the time when the server's application handler starts running.
  //
  // This field is semantically required. When not set on receive -
  // receiver should set it to the value of end_time field if it was
  // set. Or to the current time if neither was set. It is important to
  // keep end_time > start_time for consistency.
  //
  // This field is required.
  int64 start_time_unixnano = 8;

  // The end time of the span. On the client side, this is the time kept by
  // the local machine where the span execution ends. On the server side, this
  // is the time when the server application handler stops running.
  //
  // This field is semantically required. When not set on receive -
  // receiver should set it to start_time value. It is important to
  // keep end_time > start_time for consistency.
  //
  // This field is required.
  int64 end_time_unixnano = 9;

  // The set of attributes. The value can be a string, an integer, a double
  // or the Boolean values `true` or `false`. Note, global attributes like
  // server name can be set as tags using resource API. Examples of attributes:
  //
  //     "/http/user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"
  //     "/http/server_latency": 300
  //     "abc.com/myattribute": true
  //     "abc.com/score": 10.239
  repeated AttributeKeyValue attributes = 10;

  // The number of attributes that were discarded. Attributes can be discarded
  // because their keys are too long or because there are too many attributes.
  // If this value is 0, then no attributes were dropped.
  int32 dropped_attributes_count = 11;

  // A time-stamped event in the Span.
  message TimedEvent {
    // The time the event occurred.
    int64 time_unixnano = 1;

    // A user-supplied name describing the event.
    string name = 2;

    // A set of attributes on the event.
    repeated AttributeKeyValue attributes = 3;

    int32 dropped_attributes_count = 4;
  }

  // A collection of `TimedEvent`s. A `TimedEvent` is a time-stamped annotation
  // on the span, consisting of either user-supplied key-value pairs, or
  // details of a message sent/received between Spans.
  message TimedEvents {
    // A collection of `TimedEvent`s.
    repeated TimedEvent timed_event = 1;

    // The number of dropped timed events. If the value is 0, then no events were dropped.
    int32 dropped_timed_events_count = 2;
  }

  // The included timed events.
  TimedEvents timed_events = 12;

  // A pointer from the current span to another span in the same trace or in a
  // different trace. For example, this can be used in batching operations,
  // where a single batch handler processes multiple requests from different
  // traces or when the handler receives a request from a different project.
  message Link {
    // A unique identifier of a trace that this linked span is part of. The ID is a
    // 16-byte array.
    bytes trace_id = 1;

    // A unique identifier for the linked span. The ID is an 8-byte array.
    bytes span_id = 2;

    // The Tracestate associated with the link.
    Tracestate tracestate = 3;

    // A set of attributes on the link.
    repeated AttributeKeyValue attributes = 4;

    int32 dropped_attributes_count = 5;
  }

  // A collection of links, which are references from this span to a span
  // in the same or different trace.
  message Links {
    // A collection of links.
    repeated Link link = 1;

    // The number of dropped links after the maximum size was enforced. If
    // this value is 0, then no links were dropped.
    int32 dropped_links_count = 2;
  }

  // The included links.
  Links links = 13;

  // An optional final status for this span. Semantically when Status
  // wasn't set it is means span ended without errors and assume
  // Status.Ok (code = 0).
  Status status = 14;

  // An optional number of child spans that were generated while this span
  // was active. If set, allows an implementation to detect missing child spans.
  google.protobuf.UInt32Value child_span_count = 15;
}

// The `Status` type defines a logical error model that is suitable for different
// programming environments, including REST APIs and RPC APIs. This proto's fields
// are a subset of those of
// [google.rpc.Status](https://github.com/googleapis/googleapis/blob/master/google/rpc/status.proto),
// which is used by [gRPC](https://github.com/grpc).
message Status {
  // The status code. This is optional field. It is safe to assume 0 (OK)
  // when not set.
  int32 code = 1;

  // A developer-facing error message, which should be in English.
  string message = 2;
}
```

### AttributeKeyValue

```
message AttributeKeyValue {
  enum ValueType {
    STRING  = 0;
    BOOL    = 1;
    INT64   = 2;
    FLOAT64 = 3;
    BINARY  = 4;
  };

  string key = 1;
  // The type of the value.
  ValueType type = 2;
  // A string up to 256 bytes long.
  string string_value = 3;
  // A 64-bit signed integer.
  int64 int_value = 4;
  // A Boolean value represented by `true` or `false`.
  bool bool_value = 5;
  // A double value.
  double double_value = 6;
  // A binary value of bytes.
  bytes binary_value = 7;
}

```

## Trade-offs and mitigations

Timestamps were changed from google.protobuf.Timestamp to a int64 representation in Unix epoch nanoseconds. This change reduces the type-safety but benchmarks show that for small spans there is 15-20% encoding/decoding CPU speed gain. This is the right trade-off to make because encoding/decoding CPU consumption tends to dominate many workloads (particularly in OpenTelemetry Service).

## Prior art and alternatives

OpenCensus and Jaeger protocol buffer data schemas were used as the inspiration for this specification. OpenCensus was the starting point, Jaeger provided performance improvement ideas.

## Open questions

A follow up RFC is required to define the data format for metrics.

## Appendix A - Benchmarking

The following shows [benchmarking of encoding/decoding in Go](https://github.com/tigrannajaryan/exp-otelproto/) using various schemas.

Legend:
- OpenCensus    - OpenCensus protocol schema.
- OTLP/AttrMap  - OTLP schema using map for attributes.
- OTLP/AttrList - OTLP schema using list of key/values for attributes and with reduced nesting for values.
- OTLP/AttrList/TimeWrapped - Same as OTLP/AttrList, except using google.protobuf.Timestamp instead of int64 for timestamps.

Suffixes:
- Attributes - a span with 3 attributes.
- TimedEvent - a span with 3 timed events.

```
BenchmarkEncode/OpenCensus/Attributes-8         	      10	 605614915 ns/op
BenchmarkEncode/OpenCensus/TimedEvent-8         	      10	1025026687 ns/op
BenchmarkEncode/OTLP/AttrAsMap/Attributes-8     	      10	 519539723 ns/op
BenchmarkEncode/OTLP/AttrAsMap/TimedEvent-8     	      10	 841371163 ns/op
BenchmarkEncode/OTLP/AttrAsList/Attributes-8    	      50	 128790429 ns/op
BenchmarkEncode/OTLP/AttrAsList/TimedEvent-8    	      50	 175874878 ns/op
BenchmarkEncode/OTLP/AttrAsList/TimeWrapped/Attributes-8         	      50	 153184772 ns/op
BenchmarkEncode/OTLP/AttrAsList/TimeWrapped/TimedEvent-8         	      30	 232705272 ns/op
BenchmarkDecode/OpenCensus/Attributes-8                          	      10	 644103382 ns/op
BenchmarkDecode/OpenCensus/TimedEvent-8                          	       5	1132059855 ns/op
BenchmarkDecode/OTLP/AttrAsMap/Attributes-8                      	      10	 529679038 ns/op
BenchmarkDecode/OTLP/AttrAsMap/TimedEvent-8                      	      10	 867364162 ns/op
BenchmarkDecode/OTLP/AttrAsList/Attributes-8                     	      50	 228834160 ns/op
BenchmarkDecode/OTLP/AttrAsList/TimedEvent-8                     	      20	 321160309 ns/op
BenchmarkDecode/OTLP/AttrAsList/TimeWrapped/Attributes-8         	      30	 277597851 ns/op
BenchmarkDecode/OTLP/AttrAsList/TimeWrapped/TimedEvent-8         	      20	 443386880 ns/op
```

The benchmark encodes/decodes 1000 batches of 100 spans, each span containing 3 attributes or 3 timed events. The total uncompressed, encoded size of each batch is around 20KBytes.

The results show OTLP/AttrList is 5-6 times faster than OpenCensus in encoding and about 3 times faster in decoding.

Using google.protobuf.Timestamp instead of int64-encoded unix timestamp results in 1.18-1.32 times slower encoding and 1.21-1.38 times slower decoding (depending on what the span contains).
