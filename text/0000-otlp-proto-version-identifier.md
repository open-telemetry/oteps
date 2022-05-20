

# Include OTLP proto version identifier in requests

Adds OTLP request identifer to descrive the OTLP version used to create the binary or JSON payload.

## Motivation

It is useful to know the version of the OTLP protobuf definition used to create the binary or JSON.

This would enable routing to different services by inspecting the header value. This is especially useful for organisations that wish to support multiple versions where the OTLP proto definitions have changed. A currently used alternative is to append the version to the path, but this is clumsy and requires additional configuration when setting up the exporter.

## Explanation

Currently it is not possible to know the version of the OTLP proto definition before receving it. This is challening because there have been changes to the OTLP proto specification that are breaking.

This also aids organisations that receive multiple streams of OTLP data and be able to monitor and encourage users to upgrade to newer versions as appropriate.

## Internal details

* Add gRPC metadta entry or HTTP request header entry with the OTLP version used to create the binary or JSON payload

For example, we could use the key `x-otlp-version` and the string value of the OTLP proto version in the form "0.17.0".

## Trade-offs and mitigations

Currently releaed SDKs and collector versions will not include these identifiers.

## Prior art and alternatives

## Open questions

## Future possibilities
