# RFC: Data Formats for OpenTelemetry

# Overview
OpenTelemetry is an observability system, which generates observations, and then sends that data to remote systems for analysis. Describing how this observability data is represented and transmitted is critical part of the OpenTelemetry specification.

**Proposed Layers**
*   Metadata and Semantic Conventions
*   Logical Structure
*   Encodings
*   Exchange Protocols


# Metadata and Semantic Conventions

Spans represent specific operations in and between systems. Some of these operations represent calls that use well-known protocols like HTTP or database calls. Depending on the protocol and the type of operation, additional information is needed to represent and analyze a span correctly in monitoring systems. It is also important to unify how this attribution is made in different languages. This way, the operator will not need to learn specifics of a language and telemetry collected from multi-language micro-service can still be easily correlated and cross-analyzed.

## Requirements


### Handle Complex Data

Semantic information is often complex and multi-valued. How we represent semantic data is related to how attributes are defined at the logical layer.

*   At minimum, nested lists of key/value pairs are necessary to represent them.
*   Is it advantageous to also have strongly typed data structures?
*   Are there issues with having strongly typed data structures in certain environment, such as dynamic languages?


### Compose Multiple Semantics

Objects in OpenTelemetry may describe more than one type of operation. For example, a span may simultaneously represent a MySQL transaction and a TCP/IP request. Pieces of metadata must compose well, without losing the structure integrity of each piece or risking a conflict in meaning.


### Easily Extensible

Not all semantics conventions and common operations are known at this time. It must be possible to add more semantic definitions without creating to versioning or compatibility issues with OpenTelemetry implementations.


## Examples of Common Operations

#### RPC calls

Modeling distributed transactions is a core usecase, it is common to analyze errors and latency for RPC calls. Because multiple individual network requests may be involved in a single logical RPC operation, it must be possible to describe RPC calls as logical operations, independent from the descriptions of transport and application protocols which they contain.

Examples:
*   A client sending a request to a remote server.
*   A server receiving a request from a remote client.
*   Logical vs physical networking: a request for “[https://www.wikipedia.org/wiki/Rabbit](https://www.wikipedia.org/wiki/Rabbit)” results in two physical HTTP requests - one 301 redirection, and another 200 okay. The  logical RPC operation represents both requests together.

#### Network protocols

The internet suite of network protocols is thoroughly standardized, and the details of network interactions calls are critical to analysing the behavior of distributed systems.

Examples:
*   Application Protocols: HTTP/1.1, HTTP/2, gRPC
*   Transport Protocols: TCP, UDP
*   Internet Protocols: IPv4, IPv6

#### Database protocols

Database requests often contain information which is key to root causing errors and latency.

Examples:
*   SQL: MySQL, Postrgres
*   NoSQL: MongoDB


# Logical Structure

OpenTelemetry observability model, defined as data structures instead of API calls. A logical layer describes the default data structures expected to be produced by OpenTelemetry APIs for transport.

## Requirements

### Support Metadata

The logical structure must be able to support all of the requirements for the metadata layer.


### Support both an API and a Streaming Representation

OpenTelemetry data structures must work well with the APIs and SDKs which produce them. It must be possible to stream data from services without resorting to an encoding or exchange protocol which bears little resemblance to the API. 


# Encoding Formats

While there is a single set of OpenTelemetry data structures, they can be encoded in a variety of serialization formats.

## Requirements

### Prefer pre-existing formats

Codecs must be commonly available in all major languages. It is outside the bounds of the OpenTelemetry project to define new serialization formats.

**Examples:**
*   Protobuf
*   JSON


# Exchange Protocols

Exchange protocols describe how to transmit serialized OpenTelemetry data. Beyond just defining which application and transport protocol should be used, the exchange protocol defines transmission qualities such as streaming, retry, acks, backpressure, and throttling.

## Requirements

### Support Common Topologies

Be suitable for use between all of the following node types: instrumented applications, telemetry backends, local agents, stand-alone collectors/forwarders.

*   Be load-balancer friendly (do not hinder re-balancing).
*   Allow backpressure signalling.


### Be Reliable and Observable

*   Prefer transports which have high reliability of data delivery.
*   When data must be dropped, have visibility into what was not delivered.


### Be Efficient

*   Have low CPU usage for serialization and deserialization.
*   Impose minimal pressure on memory manager, including pass-through scenarios, where deserialized data is short-lived and must be serialized as-is shortly after and where such short-lived data is created and discarded at high frequency (think telemetry data forwarders).
*   Support ability to efficiently modify deserialized data and serialize again to pass further. This is related but slightly different from the previous requirement.
*   Ensure high throughput (within the available bandwidth) in high latency networks (e.g. scenarios where telemetry source and the backend are separated by high latency network).

---

# Appendix


## Currently Open Issues and PRs

*   Exchange Protocol: [https://github.com/open-telemetry/opentelemetry-specification/pull/193](https://github.com/open-telemetry/opentelemetry-specification/pull/193)


## Prior art

*   [https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/data-semantic-conventions.md](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/data-semantic-conventions.md)
*   [https://github.com/open-telemetry/opentelemetry-proto/](https://github.com/open-telemetry/opentelemetry-proto/pull/21)
*   [https://github.com/open-telemetry/opentelemetry-specification/tree/master/work_in_progress/typedspans](https://github.com/open-telemetry/opentelemetry-specification/tree/master/work_in_progress/typedspans)
*   [https://github.com/dynatrace-innovationlab/TracingApiDatamodel](https://github.com/dynatrace-innovationlab/TracingApiDatamodel)


## Questions, TODOs

*   Should “span status” be at the logical or semantic layer?
    *   Already overlaps with some semantics, like `http.status`
    *   Separate PR for this
*   Are transports separate from the exchange protocol?
    *   Supported Transport protocols, such as HTTP and UDP, may be part of the exchange protocol, or they may be a separate layer.
    *   [https://github.com/open-telemetry/opentelemetry-specification/pull/193#issuecomment-516325059](https://github.com/open-telemetry/opentelemetry-specification/pull/193#issuecomment-516325059)
