# Typed Events Specification

Add enumerated subtype and subtype-specific data structures to the existing generic tracing `TimedEvent`. Initial supported subtypes would be `ANNOTATION` (attributes only), `MESSAGE_SENT`, `MESSAGE_RECEIVED`, `ERROR`.

## Motivation

The current tracing span event specification does not provide adequate data structures for recording information from errors, faults and exceptions which impact the status of a span. A number of backend tracing systems (i.e. Stackdriver, X-Ray) provide support for detailed error info. It is important for the OpenTelemetry API to support recording this information so operators can take full advantage of backend system capabilities.

There are a high percentatge of low-volume enterprise applications which log all message content. In addition, even if there are not plans to log or otherwise store message content in production, it is often helpful to do so during the early stages of an application's development. Providing the ability to record message content will improve the utility and adoption of the OpenTelemetry API. It is anticipated that logging-only exporters may be the only exporters which ever use this data.

## Explanation

Quickly resolving issues are often aided by knowing exactly where in the code execution path the error occurred. In addition, it is often helpful to know the actual contents of local variables when edge cases trigger errors.

However this can result in substantial data. Oftentimes an exception is triggered by a different exception which is in turn triggered by another exception. Much of the resulting stacktraces is not helpful because it comes from framework methods which bear no relevance to the offending source code.

Therefore the `ERROR` event provides flexibility to record as much or as little of the details as makes sense in a particular situation with metatdata to indicate what data was omitted. With each populated event linked to the current span, root cause analysis is made easier and quicker.

## Internal details

This section specifies data format in Protocol Buffers for `TimedEvent` messages within the overall OTLP. It follows and expands on the flattened data structure in [#59 OTLP Trace Data Format](https://github.com/open-telemetry/oteps/pull/59).

### Resource

```protobuf
  // TimedEvent is a time-stamped annotation of the span, consisting of either
  // user-supplied key-value pairs, details of a message sent/received between Spans or
  // information about an error, fault or exception
  message TimedEvent {
    enum Type {
      // Unknown event type.
      TYPE_UNSPECIFIED = 0;
      // Contains only timestamp and attributes.
      ANNOTATION = 1;
      // Indicates a sent message.
      MESSAGE_SENT = 2;
      // Indicates a received message.
      MESSAGE_RECEIVED = 3;
      // Indicates an error, fault or exception occurred.
      ERROR = 4;
    }

    // The type of MessageEvent. Indicates whether the message was sent or
    // received.
    Type type = 1;

    // time_unixnano is the time the event occurred.
    int64 time_unixnano = 2;

    // name is a user-supplied description of the event.
    string name = 3;

    // attributes is a collection of attribute key/value pairs on the event.
    repeated AttributeKeyValue attributes = 4;

    // dropped_attributes_count is the number of dropped attributes. If the value is 0,
    // then no attributes were dropped.
    int32 dropped_attributes_count = 5;

    //// Fields for use only by MESSAGE_SENT and MESSAGE_RECEIVED ////

    // An identifier for the MessageEvent's message that can be used to match
    // SENT and RECEIVED MessageEvents. For example, this field could
    // represent a sequence ID for a streaming RPC. It is recommended to be
    // unique within a Span.
    uint64 message_id = 6;

    // The number of uncompressed bytes sent or received.
    uint64 uncompressed_size = 7;

    // The number of compressed bytes sent or received. If zero, assumed to
    // be the same size as uncompressed.
    uint64 compressed_size = 8;

    // The content or body of the message.
    bytes message_content = 9;

    //// Fields for use only by ERROR ////

    message Exception {
      // Unique identifier within a parent span for the exception.
      bytes id = 1;
      // The exception message.
      string messsage = 2;
      // The exception class or type.
      string type = 3;
      // Exception ID of the exception's parent, that is, the exception that caused this exception.
      bytes cause = 4;
      // The stack.
      StackTrace stack = 5;
    }

    // Collection of exceptions which triggered the error or fault.
    repeated exceptions = 10;

    // Method argument values in use when the error occurred.
    repeated AttributeKeyValue arguments = 11;

  }

  // The full details of a call stack.
  message StackTrace {
    // A single stack frame in a stack trace.
    message StackFrame {
      // The fully-qualified name that uniquely identifies the function or
      // method that is active in this frame.
      string function_name = 1;
      // The name of the source file where the function call appears.
      string file_name = 3;
      // The line number in `file_name` where the function call appears.
      int64 line_number = 4;
      // The column number where the function call appears, if available.
      // This is important in JavaScript because of its anonymous functions.
      int64 column_number = 5;
      // The binary module from where the code was loaded.
      string load_module = 6;
      // The version of the deployed source code.
      string source_version = 7;
    }

    // Stack frames in this call stack.
    repeated StackFrame frames = 1;
    // The number of stack frames that were dropped because there
    // were too many stack frames.
    // If this value is 0, then no stack frames were dropped.
    int32 dropped_frames_count = 2;

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
    uint64 stack_trace_hash_id = 3;
  }
```

## Trade-offs and mitigations

Exception info and message content can be a substantial size which may impact performance as well as application memory usage. Exporters to backend systems which do not use this info can drop it to reduce network traffic. SDKs may also want to provide configuration flags on whether this information is recorded or not when received from API calls.

## Prior art and alternatives

The error reporting APIs of AWS X-Ray, Google Stackdriver Error Reporting and Rollbar were analyzed to determine the data supported by each. This specification includes structures for providing most or all of the data these systems support. The AWS X-Ray SDKs for various languages provide an example of how this data can be populated by SDK implementations.

The annotation and sent/received message type data structures are from the OpenCensus data protocol.
