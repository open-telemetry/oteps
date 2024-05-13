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
expect telemety libraries to be harmless and safe to use, so that they
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

This document also proposes to extend the OpenTelemetry API to support
byte-slice valued attributes.  Otherwise, users have no good
alterntives to handling this kind of data, which is often how invalid
UTF-8 arises.

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
diferent protocol buffer implementation would likely be the first to
observe the invalid UTF-8, and by that time, a large batch of
telemetry could fail as a result.

### Responsibility to the API user

When a user uses a telemetry SDK to gain observability into their
application, it can lead to quite the opposite result when invalid
UTF-8 causes loss of telemetry.  When a user turns to a telemetry
system to obseve their system and makes a UTF-8 mistake, the
consequent total loss of telemetry is a definite bad outcome.  While
the above discussion addresses attempts to correct a problem,
OpenTelemetry has not given the API user what they want in this
situation.

#### Byte-slice valued attribute API

The OpenTelemetry API specification does not permit the use of
byte-slice valued attributes.  This is one reason why invalid UTF-8
arises, because OpenTelemetry users have not been given an
alternative.

When faced with a desire to record invalid UTF-8 in a stream of
telemetry, users have no good options.  They can use a base64
encoding, but they will have to do so manually, and with extra code,
and they are likely to learn this only after first making the mistake
of using a string-value to encode bytes data.  When users see
corrupted telemetry data containing � characters, they will return to
OpenTelemetry with a request: how should they report bytes data?

Ironically, in a prototype Collector processor to repair invalid
UTF-8, we have used the Zap Logger's `zap.Binary([]byte)` field
constructor to report the invalid data, and this is something an
OpenTelemetry API user cannot do.

This OTEP proposes to add an attribute-value type for byte slices, to
represent an array of bytes with unknown encoding.  Examples data
values where this attribute could be useful:

- checksum values (e.g., a sha256 checksum represented as 32 bytes)
- hash values (e.g., a 56-bit hash value represented as 7 bytes)
- register contents (e.g., a memory region ... 64 bytes)
- invalid UTF-8 (e.g., data lost because of UTF-8 validation).

### Responsibility to the Collector user

Users need to trust that the OpenTelemetry Collector will not cause
unintended data loss, and we observe this can easily result from a
mismatch of permissive receiver and strict exporter.  However, we are
aware that a strict receiver will cause whole batches of telemetry to
be lost, therefore it seems better for Collector pipelines to
explicitly handle UTF-8 validation, rather than leave it to the
protocol buffer library.

OpenTelemetry Collector SHOULD support automatic UTF-8 validation to
protect users.  There are several places this could be done:

1. Following a receiver, with data coming from an external source, 
2. Following a mutating processor, with data modified by potentially
   untrustworthy code, 
3. Prior to an exporter, with data coming from either an external
   source and/or modified by potentially untrustworhty code.

Depending on user preferences, any of these outcomes might be best.
Every mutating processor could force the pipeline to re-check for
validity, in the strictest of configurations.

#### New collector component Capabilities for UTF-8 strictness

Each collector component that strongly requires valid UTF-8 will
declare so in an Capabilities field, and this will cause the Collector
to re-validate UTF-8 before invoking that component.  By default,
exporters will require valid UTF-8.

Provided that no processors are strict about UTF-8 validation, UTF-8
validation will happen only once per pipeline segment.  This means
UTF-8 validation can be deferred until after sampling and filtering
operations have finsihed.  Each pipeline segment will have at least
one UTF-8 validation step, which will either automatically correct the
problem (default) or reject the individual item of data (optional).

### Alternatives considered

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
