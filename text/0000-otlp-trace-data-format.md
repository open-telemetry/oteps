# OTLP Trace Data Format

Author: Tigran Najaryan, Omnition Inc.

Status: Draft.

OTLP Trace Data Format specification describes the structure of the trace data that is transported by OpenTelemetry Protocol (RFCNNN).

## Motivation

This document is a continuation of OpenTelemetry Protocol RFCNNN and is necessary part of OTLP specification.

## Explanation

OTLP Trace Data Format is primarily inherited from OpenCensus protocol. Several changes are introduced with the goal of more efficient serialization. Notable differences from OpenCensus protocol are:

1. Removal of `Node` as a concept.
2. Extending `Resource` to better describe the source of the telemetry data.
3. Replacing attribute maps by lists of key/value pairs.
4. Eliminating unnecessary additional nesting in various values.

Changes 1-2 are conceptual, changes 3-4 improve performance.

## Internal details

This section specifies data format in Protocol Buffers.

### Resource

```
// Resource information. This describes the source of telemetry data.
message Resource {
  // Identifier that uniquely identifies a process within a VM/container.
  Process process = 1;

  // Name of the service.
  string service_name = 3;

  // Set of labels that describe the resource.
  repeated AttributeKeyValue attributes = 2;
}

// Identifies a process within a VM/container.
message Process {

  // The host name. Usually refers to the machine/container name.
  // For example: os.Hostname() in Go, socket.gethostname() in Python.
  string host_name = 1;

  // Process id.
  uint32 pid = 2;

  // Start time of this process. Represented in epoch time.
  int64 start_time_unixnano = 3;
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
  string name = 5;

  // Type of span. Can be used to specify additional relationships between spans
  // in addition to a parent/child relationship.
  enum SpanKind {
    // Unspecified.
    SPAN_KIND_UNSPECIFIED = 0;

    // Indicates that the span covers server-side handling of an RPC or other
    // remote network request.
    SERVER = 1;

    // Indicates that the span covers the client-side wrapper around an RPC or
    // other remote request.
    CLIENT = 2;
  }

  // Distinguishes between spans generated in a particular context. For example,
  // two spans with the same name may be distinguished using `CLIENT` (caller)
  // and `SERVER` (callee) to identify queueing latency associated with the span.
  SpanKind kind = 6;

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
  int64 start_time_unixnano = 7;

  // The end time of the span. On the client side, this is the time kept by
  // the local machine where the span execution ends. On the server side, this
  // is the time when the server application handler stops running.
  //
  // This field is semantically required. When not set on receive -
  // receiver should set it to start_time value. It is important to
  // keep end_time > start_time for consistency.
  //
  // This field is required.
  int64 end_time_unixnano = 8;

  // The set of attributes. The value can be a string, an integer, a double
  // or the Boolean values `true` or `false`. Note, global attributes like
  // server name can be set as tags using resource API. Examples of attributes:
  //
  //     "/http/user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"
  //     "/http/server_latency": 300
  //     "abc.com/myattribute": true
  //     "abc.com/score": 10.239
  repeated AttributeKeyValue attributes = 9;

  // A stack trace captured at the start of the span.
  StackTrace stack_trace = 10;

  // A time-stamped annotation or message event in the Span.
  message TimeEvent {
    // The time the event occurred.
    int64 time_unixnano = 1;

    // A text annotation with a set of attributes.
    message Annotation {
      // A user-supplied message describing the event.
      string description = 1;

      // A set of attributes on the annotation.
      repeated AttributeKeyValue attributes = 2;
    }

    // An event describing a message sent/received between Spans.
    message MessageEvent {
      // Indicates whether the message was sent or received.
      enum Type {
        // Unknown event type.
        TYPE_UNSPECIFIED = 0;
        // Indicates a sent message.
        SENT = 1;
        // Indicates a received message.
        RECEIVED = 2;
      }

      // The type of MessageEvent. Indicates whether the message was sent or
      // received.
      Type type = 1;

      // An identifier for the MessageEvent's message that can be used to match
      // SENT and RECEIVED MessageEvents. For example, this field could
      // represent a sequence ID for a streaming RPC. It is recommended to be
      // unique within a Span.
      uint64 id = 2;

      // The number of uncompressed bytes sent or received.
      uint64 uncompressed_size = 3;

      // The number of compressed bytes sent or received. If zero, assumed to
      // be the same size as uncompressed.
      uint64 compressed_size = 4;
    }

    // A `TimeEvent` can contain either an `Annotation` object or a
    // `MessageEvent` object, but not both.
    oneof value {
      // A text annotation with a set of attributes.
      Annotation annotation = 2;

      // An event describing a message sent/received between Spans.
      MessageEvent message_event = 3;
    }
  }

  // A collection of `TimeEvent`s. A `TimeEvent` is a time-stamped annotation
  // on the span, consisting of either user-supplied key-value pairs, or
  // details of a message sent/received between Spans.
  message TimeEvents {
    // A collection of `TimeEvent`s.
    repeated TimeEvent time_event = 1;

    // The number of dropped annotations in all the included time events.
    // If the value is 0, then no annotations were dropped.
    int32 dropped_annotations_count = 2;

    // The number of dropped message events in all the included time events.
    // If the value is 0, then no message events were dropped.
    int32 dropped_message_events_count = 3;
  }

  // The included time events.
  TimeEvents time_events = 11;

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

    // The relationship of the current span relative to the linked span: child,
    // parent, or unspecified.
    enum Type {
      // The relationship of the two spans is unknown, or known but other
      // than parent-child.
      TYPE_UNSPECIFIED = 0;
      // The linked span is a child of the current span.
      CHILD_LINKED_SPAN = 1;
      // The linked span is a parent of the current span.
      PARENT_LINKED_SPAN = 2;
    }

    // The relationship of the current span relative to the linked span.
    Type type = 3;

    // A set of attributes on the link.
    repeated AttributeKeyValue attributes = 4;

    // The Tracestate associated with the link.
    Tracestate tracestate = 5;
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
  Links links = 12;

  // An optional final status for this span. Semantically when Status
  // wasn't set it is means span ended without errors and assume
  // Status.Ok (code = 0).
  Status status = 13;

  // An optional resource that is associated with this span. If not set, this span
  // should be part of a ResourceSpan that does include the resource information, unless resource
  // information is unknown.
  Resource resource = 14;

  // A highly recommended but not required flag that identifies when a
  // trace crosses a process boundary. True when the parent_span belongs
  // to the same process as the current span. This flag is most commonly
  // used to indicate the need to adjust time as clocks in different
  // processes may not be synchronized.
  google.protobuf.BoolValue same_process_as_parent_span = 15;

  // An optional number of child spans that were generated while this span
  // was active. If set, allows an implementation to detect missing child spans.
  google.protobuf.UInt32Value child_span_count = 16;
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

// The call stack which originated this span.
message StackTrace {
  // A single stack frame in a stack trace.
  message StackFrame {
    // The fully-qualified name that uniquely identifies the function or
    // method that is active in this frame.
    string function_name = 1;
    // An un-mangled function name, if `function_name` is
    // [mangled](http://www.avabodh.com/cxxin/namemangling.html). The name can
    // be fully qualified.
    string original_function_name = 2;
    // The name of the source file where the function call appears.
    string file_name = 3;
    // The line number in `file_name` where the function call appears.
    int64 line_number = 4;
    // The column number where the function call appears, if available.
    // This is important in JavaScript because of its anonymous functions.
    int64 column_number = 5;
    // The binary module from where the code was loaded.
    Module load_module = 6;
    // The version of the deployed source code.
    string source_version = 7;
  }

  // A collection of stack frames, which can be truncated.
  message StackFrames {
    // Stack frames in this call stack.
    repeated StackFrame frame = 1;
    // The number of stack frames that were dropped because there
    // were too many stack frames.
    // If this value is 0, then no stack frames were dropped.
    int32 dropped_frames_count = 2;
  }

  // Stack frames in this stack trace.
  StackFrames stack_frames = 1;

  // The hash ID is used to conserve network bandwidth for duplicate
  // stack traces within a single trace.
  //
  // Often multiple spans will have identical stack traces.
  // The first occurrence of a stack trace should contain both
  // `stack_frames` and a value in `stack_trace_hash_id`.
  //
  // Subsequent spans within the same request can refer
  // to that stack trace by setting only `stack_trace_hash_id`.
  //
  // TODO: describe how to deal with the case where stack_trace_hash_id is
  // zero because it was not set.
  uint64 stack_trace_hash_id = 2;
}

// A description of a binary module.
message Module {
  // TODO: document the meaning of this field.
  // For example: main binary, kernel modules, and dynamic libraries
  // such as libc.so, sharedlib.so.
  string module = 1;

  // A unique identifier for the module, usually a hash of its
  // contents.
  string build_id = 2;
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

The following shows [benchmarking](https://github.com/tigrannajaryan/exp-otelproto/) using various schemas.

Legend:
- OpenCensus    - OpenCensus protocol schema.
- OTLP/AttrMap  - OTLP schema using map for attributes.
- OTLP/AttrList - OTLP schema using list of key/values for attributes and with reduced nesting for values.
- OTLP/AttrList/TimeWrapped - Same as OTLP/AttrList, except using google.protobuf.Timestamp instead of int64 for timestamps.

```
BenchmarkEncode/OpenCensus-8         	      10	 622843375 ns/op
BenchmarkEncode/OTLP/AttrMap-8       	      10	 528652501 ns/op
BenchmarkEncode/OTLP/AttrList-8      	      50	 132795894 ns/op
BenchmarkEncode/OTLP/AttrList/TimeWrapped-8         	      50	 154429659 ns/op
BenchmarkDecode/OpenCensus-8                        	      10	 656730934 ns/op
BenchmarkDecode/OTLP/AttrMap-8                      	      10	 567621380 ns/op
BenchmarkDecode/OTLP/AttrList-8                     	      30	 240150634 ns/op
BenchmarkDecode/OTLP/AttrList/TimeWrapped-8         	      30	 293842943 ns/op
```

Benchmarks show OTLP/AttrList is 4.7 times faster than OpenCensus in encoding and 2.7 times faster in decoding.

Using google.protobuf.Timestamp results in 1.16 times slower encoding and 1.22 times slower decoding.

