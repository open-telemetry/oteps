# Stabilizing messaging semantic conventions for tracing

This document aims to describe the necessary changes for bringing the [existing semantic conventions for messaging](https://github.com/open-telemetry/opentelemetry-specification/blob/a1a8676a43dce6a4e447f65518aef8e98784306c/specification/trace/semantic_conventions/messaging.md)
from the current [experimental](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#experimental)
to a [stable](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#stable)
state.

This OTEP is based on [OTEP 0173](0173-messaging-semantic-conventions.md),
which defines basic terms and describes messaging scenarios that should be
supported by the semantic conventions.

## Motivation

This OTEP serves as a foundation for a first stable version of [messaging semantic conventions for tracing](https://github.com/open-telemetry/opentelemetry-specification/blob/a1a8676a43dce6a4e447f65518aef8e98784306c/specification/trace/semantic_conventions/messaging.md).
It aims to define clear, consistent, and extensible conventions, it also
describes reasons and motivations that led to the formulation of those
conventions. The conventions comprise areas such as context propagation, span
structure, names, and attributes.

After this OTEP is merged, the changes it proposes will be merged into the
messaging semantic conventions, which will subsequently be declared
[stable](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#stable).

## Proposed stable messaging semantic conventions for tracing

### Terminology

The terminology used in this document is based on the [CloudEvents specification](https://github.com/cloudevents/spec/blob/v1.0.1/spec.md).
CloudEvents is hosted by the CNCF and provides a specification for describing
event data in common formats to provide interoperability across services,
platforms and systems.

#### Message

A "message" is a transport envelope for the transfer of information. The
information is a combination of a payload and metadata. Metadata can be
directed at consumers or at intermediaries on the message path. Messages are
transferred via one or more intermediaries.  Messages are uniquely
identifiable.

In the strict sense, a _message_ is a payload that is sent to a specific
destination, whereas an _event_ is a signal emitted by a component upon
reaching a given state. This document is agnostic of those differences and uses
the term "message" in a wider sense to cover both concepts.

#### Producer

The "producer" is a specific instance, process or device that creates and
publishes a message. "Publishing" is the process of sending a message or batch
to the intermediary or consumer.

#### Consumer

A "consumer" receives the message and acts upon it. It uses the context and
data to execute some logic, which might lead to the occurrence of new events.

The consumer receives, processes, and settles a message. "Receiving" is the
process of obtaining a message from the intermediary, "processing" is the
process of acting on the information a message contains, "settling" is the
process of notifying an intermediary that a message was processed successfully.

#### Intermediary

An "intermediary" receives a message to forward it to the next receiver, which
might be another intermediary or a consumer.

### Stages of producing and consuming messages

Producing and consuming a message involves five stages:

```
                               CONSUMER

                   . . . . . . Settle
PRODUCER           .              ^
                   .              |
Create             .           Process
  |                v              ^
  v        +--------------+       |            
Publish -> | INTERMEDIARY | -> Receive
           +--------------+
```

1. The producer creates a message.
2. The producer publishes the message to an intermediary.
3. The consumer receives the message from an intermediary.
4. The consumer processes the message.
5. The consumer settles the message by notifying the intermediary that the
   message was processed. In some cases (fire-and-forget), the settlement stage
   does not exist.

The semantic conventions described below define how to model those stages in
traces, how to propagate context, and how to enrich traces with attributes.

### Context propagation

Two layers of context propagation are required for messaging workflows:

1. The _creation context layer_ allows to correlate the producer and consumers of
   a message, regardless of intermediary instrumentation. The creation context
   is created by the producer and must be propagated to the consumers. It must not
   be altered by intermediaries.

   This layer helps to model dependencies between producers and consumers,
   regardless of the underlying messaging transport mechanism and its
   instrumentation.
2. The _transport context layer_ allows to correlate the producer and the
   consumer with an intermediary. If there are more than one intermediaries,
   it allows to correlate intermediaries among each other. The transport context
   might be changed by intermediaries, according to intermediary instrumentations.
   Intermediaries that are not instrumented might simply drop the transport
   context.

   This layer helps to gain insights into details of the message transport.

A producer MUST attach a creation context to each message. The creation context
MUST be attached in a way so that it is not changed by intermediaries. A
producer MAY propagate a transport context to an intermediary.  An
intermediary MAY propagate a transport context to a consumer.

### Span structure, names, and attributes

### System specific extensions

### Examples

## Future possibilities
