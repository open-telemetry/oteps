# Consistent position towards UTF-8 validation in OpenTelemetry

The OpenTelemetry project has not taken a strong position on how to handle
invalid UTF-8 in its SDKs and collector components.

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
be experience harm through data loss.

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

These are honest mistakes.  The OpenTelemetry project should declare
invalid UTF-8 as a non-error condition, consistent with its error
handling principles.

> The API and SDK SHOULD provide safe defaults for missing or invalid arguments.

This document proposes that OpenTelemetry's default approach should be
the one taken by the Rust `String::from_utf8_lossy` method, which
converts invalid UTF-8 byte sequences to the Unicode replacement
character, `�` (U+FFFD).  Taking this approach by default will avoid
data loss for users when the inevitable happens.

### Scenarios covered

In an example scenario, a user creates a span with an invalid name.
Without further intervention, the trace SDK then exports this through
OTLP over gRPC.  What happens next depends on which language, RPC
client library, and protocol buffer library is used.  For this
document, we examined the official [Golang
gRPC](https://github.com/grpc/grpc-go) implementation and the [Rust
hyperium/tonic](https://github.com/hyperium/tonic) implementation.

For demonstration purposes, gRPC libraries usually provide a Hello
World client and server program.  This study modifies the Hello World
client available in each language.

#### Golang client

With a Golang gRPC client modified to send an invalid string, e.g.,

```
	var buf bytes.Buffer
	buf.WriteString(*name)
	buf.Write([]byte{0xff, 0xff})
    // ...
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: buf.String()})
 ```

the client program exits:

```
2024/05/10 15:09:23 could not greet: rpc error: code = Internal desc = grpc: error while marshaling: string field contains invalid UTF-8
```

In this case, validation is performed on the client side.  If a client
somehow bypasses UTF-8 validation, the server will respond with an
error.

```
2024/05/10 15:28:13 could not greet: rpc error: code = Internal desc = grpc: error unmarshalling request: string field contains invalid UTF-8
```

It appears possible to configure gRPC and/or the protocol buffer
library to bypass UTF-8 validation, but by default gRPC-Go will
validate on both sides.

#### Rust client

With a Rust gRPC client modified to send an invalid string, e.g.,

```
    let bytes = vec![255, 255];
    
    let request = tonic::Request::new(HelloRequest {
        name: unsafe { String::from_utf8_unchecked(bytes) },
    });
```

the client program exits:

```
Error: Status { code: Internal, message: "failed to decode Protobuf message: HelloRequest.name: invalid string value: data is not UTF-8 encoded", metadata: MetadataMap { headers: {"content-type": "application/grpc", "date": "Fri, 10 May 2024 16:10:46 GMT", "content-length": "0"} }, source: None }
```

In this case, validation is only performed on the server side.

#### OpenTelemetry Collector `pdata`

The OpenTelemetry Collector uses Gogo protobuf package to unmarshal
the structs that underly its `pdata` objects.  With the "gogofaster"
compiler option is selected, its `Unmarshal()` logic does not check
for valid UTF-8 strings.

The potential for user harm is clear at this point.  gRPC libraries
may intentionally bypass UTF-8 validation in some cases, and this may
go unnoticed until a small amount of invalid UTF-8 data leads to data
loss in a different part of the system where validation is strictly
performed.  When this happens in a telemetry pipeline, it can lead
loss of valid data that was inadvertently combined with invalid data,
possibly through batching.

### Responsibility to the user

When a user uses a telemetry SDK to gain observability into their
application, it can lead to quite the opposite result when invalid
UTF-8 causes loss of telemetry.  When a user turns to a telemetry
system to obseve their system and makes a UTF-8 mistake, the
consequent total loss of telemetry is a definite bad outcome.  For the
user to be well-served in this situation, OpenTelemetry SDKs and
Collectors SHOULD correct invalid UTF-8 using the Unicode replacement
character, `�` (U+FFFD) as a replacement for all invalid substrings.

We take the position that OpenTelemety SDKs and Collectors SHOULD
explicitly make an effort to preserve telemetry data when
presumed-UTF-8 data is invalid, rather than allowing it to drop with
an error as a result.  Accepting this proposal means this policy will
be added to the OpenTelemetry error handling guidelines.

#### Give users what they want

The OpenTelemetry API specification does not permit the use of
byte-slice valued attributes.  This is one reason why invalid UTF-8
arises, because OpenTelemetry users have not been given an
alternative.

When faced with a desire to use record invalid UTF-8 in a stream of
telemetry, users have no good options.  They can use a base64
encoding, but they will have to do so manually, and with extra code,
and they are likely to learn this only after first making the mistake
of using a string-value to encode bytes data.  When users see
corrupted telemetry data containing � characters, they will return to
OpenTelemetry and with a request, how should they report bytes data.

Ironically, in a prototype Collector processor to repair invalid
UTF-8, we have used the Zap Logger's `zap.Binary([]byte)` field
constructor to report the invalid data, and this is something an
OpenTelemetry API user cannot do.

### Alternatives considered

There are scenarios, such as in the Rust SDK, where there is a strong
expectation of safety and correctness.  In these settings, where the
use of `unsafe` is possible but rare and invalid UTF-8 is unlikely, it
can make sense to bypass UTF-8 validation in the client to save a few
CPU cycles.

We could imagine a more permissive OpenTelemetry policy, in which SDKs
and Collectors are advised to treat all String-valued fields in the
OTLP protocol buffer as "presumptive" UTF-8, meaning data that SHOULD
be encoded as UTF-8 but that has not been checked for validity.  Under
this policy, SDKs and Collectors would propagate invalid data and
leave it to the consumer to decide how to handle it.

## Trade-offs and mitigations

TODO

## Prior art and alternatives

TODO

## Open questions

TODO

## Future possibilities

TODO
