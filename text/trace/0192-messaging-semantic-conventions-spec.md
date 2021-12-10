# Stabilizing messaging semantic conventions for tracing

This document aims to describe the necessary changes for bringing the [existing semantic conventions for messaging](https://github.com/open-telemetry/opentelemetry-specification/blob/a1a8676a43dce6a4e447f65518aef8e98784306c/specification/trace/semantic_conventions/messaging.md)
from the current [experimental](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#experimental)
to a [stable](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#stable)
state.

This OTEP is based on [OTEP 0173](0173-messaging-semantic-conventions.md),
which defines basic terms and describes messaging scenarios that should be
supported by messaging semantic conventions.

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
directed at consumers or intermediaries on the message path. Messages are
transferred via one or more intermediaries.  Messages are uniquely
identifiable.

In the strict sense, a _message_ is a payload that is sent to a specific
destination, whereas an _event_ is a signal emitted by a component upon
reaching a given state. This document is agnostic of those differences and uses
the term "message" in a wider sense to cover both concepts.

#### Producer

The "producer" is a specific instance, process or device that creates and
publishes a message. "Publishing" is the process of sending a message or batch
to the intermediary.

#### Consumer

A "consumer" receives the message and acts upon it. It uses the context and
data to execute some logic, which might lead to the occurrence of new messages.

The consumer receives, processes, and settles a message. "Receiving" is the
process of obtaining a message from the intermediary, "processing" is the
process of acting on the information a message contains, "settling" is the
process of notifying an intermediary that a message was processed successfully.

#### Intermediary

An "intermediary" receives a message for the purpose of forwarding it to the
next receiver, which might be another intermediary or a consumer.

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

Two layers of context propagation need to be distinguished for messaging
workflows:

1. The _creation context layer_ allows correlating the producer with the
   consumers of a message, regardless of intermediary instrumentation. The
   creation context is created by the producer and must be propagated to the
   consumers. It must not be altered by intermediaries.

   This layer helps to model dependencies between producers and consumers,
   regardless of the underlying messaging transport mechanism and its
   instrumentation.
2. The _transport context layer_ allows correlating the producer and the
   consumer with an intermediary. It also allows to correlate multiple
   intermediaries among each other. The transport context can be changed by
   intermediaries, according to intermediary instrumentations. Intermediaries
   that are not instrumented might simply drop the transport context.

   This layer helps to gain insights into details of the message transport.

A producer MUST attach a creation context to each message. The creation context
MUST be attached in a way so that it is not changed by intermediaries. A
producer MAY propagate a transport context to an intermediary.  An
intermediary MAY propagate a transport context to a consumer.

### Span structure, names, and attributes

#### Consumer

For many use cases, it is not possible to rely on the presence of "Process"
spans for correlating producer with consumer traces: there are cases where a
dedicated processing operation cannot be identified, or where processing
happens in a different trace. Furthermore, processing operations often are not
covered by messaging libraries and SDKs, but take place in application code.
Consistently creating spans for "Processing" operations would require either
effort from the application owner to correctly instrument those operations, or
additional capabilities of messaging libraries and SDKs (e. g. hooks for
processing callbacks, which can then be instrumented by the libraries or SDKs).

While it is possible to create "Process" spans and correlate those with
consumer traces in certain cases, this is not something that can be generally
required. Therefore, it is more feasible to require the creation of "Receive"
spans to correlate producer with consumer traces. A "Receive" span must link
to the "Create" spans of all messages that are handled by the respective
"Receive" operation. Depending on the use case, "Receive" spans can correlate
with "Process" spans or other spans modelling processing operations.

The operation modelled by the "Receive" span does not strictly refer to
receiving the message from intermediaries, but instead refers to the
application receiving messages for processing. If messages are fetched from the
intermediary and forwarded to the application in one go, the whole operation
might be covered by a "Receive" span. However, clients might pre-fetch messages
from intermediaries and cache those messages, and only forward messages to the
application at a later time. In this case, the operation of pre-fetching and
caching should not be covered by the "Receive" span.

"Receive" spans SHOULD be created for all messages obtained by or passed to the
application for processing. "Receive" spans MUST NOT be created for messages
not forwarded to the application, but pre-fetched or cached by messaging
libraries or SDKs. A single "Receive" span can account for a single message,
for multiple messages (in case messages are passed for processing as batches),
or for no message at all (in it is signalled that no messages were received).
For each message it accounts for, a "Receive" span SHOULD link to the "Create"
span for the message.

### System-specific extensions

### Examples

## Future possibilities
