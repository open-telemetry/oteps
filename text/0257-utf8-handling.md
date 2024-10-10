# Consistent position towards UTF-8 validation in OpenTelemetry

The OpenTelemetry project has not taken a strong position on how to handle
invalid UTF-8 in its SDKs and Collector components.

## Motivation

When handling of invalid UTF-8 is left unspecified, anything can
happen.  As a project, OpenTelemetry should give its component authors
guidance on how to treat invalid UTF-8 so that users can confidently
rely on OpenTelemetry software/systems.

## Explanation

OpenTelemetry has existing [error handling
guidelines](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/error-handling.md)
which dictate the project's general posture towards handling.  Users
expect telemetry libraries to be harmless and safe to use, so that they
can confidently instrument their application without unnecessary risk.

> The API and SDK SHOULD provide safe defaults for missing or invalid arguments.

Generally, users should not have to check for errors from potentially
fallible instrumentation APIs, because it burdens the user, and that
instead OpenTelemetry SDKs should resort to corrective action.
Quoting the error handling guidelines,

> For instance, a name like empty may be used if the user passes in
> null as the span name argument during Span construction.

Users also make this demand of the OpenTelemetry collector, which is
expected to safely transport telemetry data.

OpenTelemetry's OTLP protocol, specifically dictates how [valid and
invalid UTF-8 strings should be mapped into attribute
values](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/common/attribute-type-mapping.md#string-values),
but there is no guidance on how to treat many other `string` fields
found in the OTLP protocol.  The Span `Name` field, for example,
SHOULD be a valid UTF-8 string according to the Protocol Buffer
`proto3`-syntax specification.  What is expected from the SDK when a
presumptive UTF-8 field is actually invalid UTF-8?  What is expected
from the Collector?  Without a clear definition, users are likely to
experience harmful loss of data.

This document proposes to update OpenTelemetry error handling
guidelines to require SDKs and Collectors to protect users from loss
of telemetry caused by string-valued fields that are not well-formed.
A number of options are available, all of which are better than doing
nothing about this problem.

## Internal details

Text handling is such a key aspect of programming languages that
practically, they all present different a approach.  While most
languages provide guardrails that maintain a well-formed character
encoding in memory at runtime, there is usually a way for a programmer
handling bytes incorrectly to yield an invalid string encoding.

These are honest mistakes.  If the telemetry will pass through an OTLP
representation and UTF-8 validation is left unchecked, there is an
opportunity for an entire batch of data to be rejected by the protocol
buffer subsystem used in both gRPC and HTTP code paths.

This is especially harmful to users, as the observability tools they
have in place suddely can't help them.  Users can't log, span, or
metric about the invalid UTF-8 data, and terminals may also have
trouble displaying the data.  For users to learn the source of their
problem, some component has to correct the invalid data before data
loss happens.

### Possible resolutions

Here are some three options:

1. Automatically correct invalid UTF-8.
2. Drop partial batches of data containing invalid UTF-8.
3. Permissively allow invalid UTF-8 to propagate (further) into the pipeline.

In this proposal, we reject the third option.  See alternatives considered, below.

#### Automatic correction

This document proposes that OpenTelemetry's default approach should be
the one taken by the Rust `String::from_utf8_lossy` method, which
converts invalid UTF-8 byte sequences to the Unicode replacement
character, `�` (U+FFFD).  This is the most preferable outcome as it is
simple and preserves what valid content can be recovered from the
data.

#### Dropping data

Among the options available, we view dropping whole batches of
telemetry as an unacceptable outcome.  OpenTelemetry SDKs and
Collectors SHOULD prefer to drop individual items of telemetry as
opposed to dropping whole batches of telemetry.

When a Collector drops individual items of telemetry as a result of
invalid UTF-8, it SHOULD return a `rejected_spans`,
`rejected_data_points`, or `rejected_log_records` count along with a
message indicating data loss.

### Survey of existing systems

The Go gRPC-Go implementation with its default protocol buffer checks
for valid UTF-8 in both the client (`rpc error: code = Internal desc =
grpc: error while marshaling: string field contains invalid UTF-8`)
and the server (`rpc error: code = Internal desc = grpc: error
unmarshalling request: string field contains invalid UTF-8`).

The Rust Hyperium gRPC implementation checks for valid UTF-8 on the
server but not on the client (`Error: Status { code: Internal,
message: "failed to decode Protobuf message: HelloRequest.name:
invalid string value: data is not UTF-8 encoded"...`).

The OpenTelemetry Collector OTLP exporter and receiver do not validate
UTF-8, as a result of choosing the "Gogofaster" protocol buffer
library option.

Given these out-of-the-box configurations, it would be possible for a
Rust SDK to send invald UTF-8 to an OpenTelemetry Collector's OTLP
receiver.  That collector could forward to another OpenTelemtry
Collector over OTLP successfully, but a non-OTLP exporter using a
different protocol buffer implementation would likely be the first to
observe the invalid UTF-8, and by that time, a large batch of
telemetry could fail as a result.

### Responsibility to the SDK user

The existing specification [dictates a safety mechanism for exporting
invalid string-valued attributes safely][SAFETY], however it only
applies to attribute strings:

[SAFETY]: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/common/attribute-type-mapping.md#string-values

```
## String Values
String values which are valid UTF-8 sequences SHOULD be converted to AnyValue's string_value field.

String values which are not valid Unicode sequences SHOULD be converted to AnyValue's bytes_value with the bytes representing the string in the original order and format of the source string.
```

To satisfy the existing error-handling requirements, the OpenTelemetry
SDK specifications will be modified for all signals with an opt-out
validation feature.

#### Proposed Collector behavior change

The SDK SHOULD in its default configuration validate all string-valued
telemetry data fields.  Each run of invalid UTF-8 (i.e., any invalid
UTF-8 sequences) will be replaced by a single Unicode replacement
character, `�` (U+FFFD).  The exact behavior of this correction is
undefined.  When possible, SDKs SHOULD use a built-in library for this
repair (for example, [Golang's
`strings.ToValidUTF8()`](https://pkg.go.dev/strings#ToValidUTF8) or
[Rust's
`String::to_utf8_lossy()`](https://doc.rust-lang.org/std/string/struct.String.html#method.from_utf8_lossy)
satisfy this requirement).

#### No byte-slice valued attribute API

As a caveat, the OpenTelemetry project has previously debated and
rejected the potential to support a byte-slice typed attribute in
OpenTelemetry APIs.

This potential feature was rejected because it allows API users a way
to record a possibly uninterpretable sequence of bytes.  Users with
invalid UTF-8 are left with a few options, for example:

- Base64-encode the invalid data wrapped in human-readable syntax
  (e.g., `base64:WNhbHF1ZWVuY29hY2hod`).
- Transmute the byte slice to an integer slice, which is supported.

### Responsibility to the Collector user

Users need to trust that the OpenTelemetry Collector will not cause
unintended data loss, and we observe this can easily result from a
mismatch of permissive receiver and strict exporter.  However, we are
aware that a strict receiver will cause whole batches of telemetry to
be lost, therefore it seems better for Collector pipelines to
explicitly handle UTF-8 validation, rather than leave it to the
protocol buffer library.

OpenTelemetry Collector should support automatic UTF-8 validation to
protect users, however there are several places this could be done:

1. Following a receiver, with data coming from an external source,
2. Following a mutating processor, with data modified by potentially
   untrustworthy code,
3. Prior to an exporter, with data coming from either an external
   source and/or modified by potentially untrustworhty code.

Each of these approaches will take significant effort and vary in cost
at runtime.

#### Proposed Collector behavior change

To reduce the cost of UTF-8 validation to a minimum, we propose:

- UTF-8 validation SHOULD be enabled by default for all Receiver components
- Users SHOULD be able to opt out of UTF-8 validation
- A `receiverhelper` library SHOULD offer a function to correct
  invalid UTF-8 in-place, with two configurable outcomes (1) reject
  individual items containing invalid UTF-8, meaning to count them as
  rejected spans/points/logs, and (2) repair invalid UTF-8 as specified
  for SDKs above.

When an OpenTelemetry collector receives telemetry data in any
protocol, in cases where the underlying RPC or protocol buffer
libraries does not guarantee valid UTF-8, the `receiverhelper` library
SHOULD be called to optionally validate UTF-8 according to the
confiuration described above.

### Alternatives considered

#### Exhaustive validation

The Collector behavior proposed above only validates data after it is
received, not after it is modified by a processor.  We consider the
risk of a malfunctioning processor to be acceptable.  If this happens,
it will be considered a bug in the Collector component.  In other
words, the Collector SHOULD NOT perform internal data validation and
it SHOULD perform external data validation.

#### Permissive transport

We observe that some protocol buffer implementations are permissive
with respect to invalid UTF-8, while others are strict.  It can be
risky to combine dissimilar protocol buffer implementations in a
pipeline, since a receiver might accept data that an exporter cannot
send, simply resulting from invalid UTF-8.

Considering whether components could support permissive data
transport, it appears to be risky and for little reward.  If
OpenTelemetry decided to allow this, it would not be a sensible
default.  On the other hand, existing OpenTelemetry Collectors have
this behavior, and changing it will require time.  In the period of
time while this situation persists, we effectively have permissive
data transport (and the associated risks).
