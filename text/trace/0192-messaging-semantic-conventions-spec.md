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
conventions. The conventions comprise areas such as context propagation, trace
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

The consumer receives, processes, and settles a message.

* "Receiving" is the process of obtaining a message from the intermediary.
* "Processing" is the process of acting on the information a message contains.
* "Settling" is the process where intermediary and consumer agree on the state
  of the transfer.

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

A message may pass many different components and layers in one or more
intermediaries when it is propagated from the producer to the consumer. It
cannot be assumed, and in many cases it is not even desired, that all those
components and layers are instrumented and propagate context according to
OpenTelemetry requirements.

A _creation context_ allows correlating the producer with the consumers of a
message, regardless of intermediary instrumentation. The creation context is
created by the producer and must be propagated to the consumers. It must not be
altered by intermediaries.  This context helps to model dependencies between
producers and consumers, regardless of the underlying messaging transport
mechanism and its instrumentation.

Instrumentors are required to instrument producer and consumer applications
so that context is attached to messages and extraced from messages in a
coordinated way. Future versions of these conventions might recommend [context propagation according to certain industry standards](#standards-for-context-propagation).

> A producer SHOULD attach a creation context to each message. The creation context
> SHOULD be attached in a way so that it is not possible to be changed by intermediaries.

### Trace structure

#### Producer

Producers are responsible for attaching a creation context to a message.
Subsequent consumers will use this context to link consumer traces to producer
traces. Ideally, each message gets a unique and distinct creation context
assigned. However, as a context must refer to a span this would require the
creation of a distinct span for each message, which is not feasible in all
scenarios. In certain batching scenarios where many messages are created and
published in large batches, creating a span for every single message would
obfuscate traces and is not desirable. Thus having a unique and distinct
context per message is recommended, but not required.

For each producer scenario, a "Publish" span needs to be created. This span
measures the duration of the call or operation that provides messages for
sending or publishing to an intermediary. This call or operation (and the
related "Publish" span) can either refer to a single message or a batch of
multiple messages.

It is recommended to create a "Create" span for every single message. "Create"
spans can be created during the "Publish" operation as children of the
"Publish" span. Alternatively, "Create" spans can be created independently of
the "Publish" operation. In that case, SDKs may provide mechanisms to allow
attaching independent contexts with messages.

If a "Create" span exists for a message, its context must be attached to the
message. If no "Create" span exists for a message, the context of the related
"Publish" span must be attached to the message.

> "Publish" spans SHOULD be created for operations of providing messages for
> sending or publishing to an intermediary. A single "Publish" span can account
> for a single message, or for multiple messages (in case of providing
> messages in batches). "Create" spans MAY be created. A single "Create" span
> SHOULD account only for a single message.
>
> If a "Create" span exists for a message, its context SHOULD be attached to
> the message. If no "Create" span exists, the context of the related "Publish"
> span SHOULD be attached to the message.

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
required. Therefore, it is more feasible to require the creation of "Deliver"
spans (for push-based APIs) or "Receive" spans (for pull-based APIs) to
correlate producer with consumer traces.

##### Instrumenting push-based scenarios

In push-based consumer scenarios, the delivery of messages is not initiated by
the application code. Instead, callbacks or handlers are registered and then
called by messaging SDKs to forward messages to the application.

A "Deliver" span covers the call of such a callback or handler, and should link
to the "Create" spans of all messages that are forwarded via the respective
call. Depending on the use case, "Deliver" spans can correlate with "Process"
spans or other spans modelling processing operations.

> "Deliver" spans SHOULD be created for operations of passing messages to the
> application, when those operations are not initiated by the application
> code.

##### Instrumenting pull-based scenarios

In pull-based consumer scenarios, the delivery of messages is requested by the
application code. This usually involves a blocking call, which returns zero or
more messages on completion.

A "Receive" span covers such calls, and should link
to the "Create" spans of all messages that are forwarded via the respective
call. Depending on the use case, "Receive" spans can correlate with "Process"
spans or other spans modelling processing operations.

> "Receive" spans SHOULD be created for operations of passing messages to the
> application, when the those operations are initiated by the application code.

##### General considerations for both push-based and pull-based scenarios

The operations modelled by "Deliver" or "Receive" spans do not strictly refer
to receiving the message from intermediaries, but instead refer to the
application receiving messages for processing. If messages are fetched from the
intermediary and forwarded to the application in one go, the whole operation
might be covered by a "Deliver" or "Receive" span. However, clients might
pre-fetch messages from intermediaries and cache those messages, and only
forward messages to the application at a later time. In this case, the
operation of pre-fetching and caching should not be covered by the "Deliver" or
"Receive" spans.

Operations covered by "Deliver" or "Receive" can forward zero messages (e. g.
to notify the application that no message is available for processing), one
message, or multiple messages (a batch of messages). "Deliver" and "Receive"
spans should link to the "Create" span of the messages forwarded, thus those
spans can link to zero, one, or multiple "Create" spans.

> "Deliver" or "Receive" spans MUST NOT be created for messages which are not
> forwarded to the application, but are pre-fetched or cached by messaging
> libraries or SDKs.
>
> A single "Deliver" or "Receive" span can account for a single message, for
> multiple messages (in case messages are passed for processing as batches), or
> for no message at all (if it is signalled that no messages were received).  For
> each message it accounts for, the "Deliver" or "Receive" span SHOULD link to
> the "Create" span for the message.

##### Settlement

Messages can be settled in a variety of different ways. In some cases, messages
are not settled at all (fire-and-forget), in other cases settlement operations
are triggered manually by the user, in callback scenarios settlement can be
automatically triggered by messaging SDKs based on return values of callbacks.

A "Settle" spans should be created for every settlement operation, no matter if
the settlement operation was manually triggered by the user or automatically
triggered by SDKs. SDKs will in some cases auto-settle messages in
push-scenarios, when messages are delivered via callbacks. In this case it is
recommended to create a parent span, so that a "Settle" span will be a sibling
of the related "Deliver" span.

In cases where it is not feasible to create spans for settlement operations, an
event should be created instead. Events could be added to "Deliver" spans or to
ambient spans.

"Settle" spans may link to "Create" spans of the messages that are settled,
however, for some settlement scenarios this is not feasible or possible.

> "Settle" spans or events SHOULD be created for every manually or automatically
> triggered settlement operation. A single "Settle" span can account for a
> single message or for multiple messages (in case messages are passed for
> settling as batches). For each message it accounts for, the "Settle" span
> MAY link to the "Create" span for the message.

### Span names, kinds, and attributes

#### Span name

The span name should be descriptive and make it clear, what operation a span
describes. In the context of messaging systems, this means that a span should
at the very least make clear that it refers to a messaging system, in addition
it needs to make clear what particular [messaging operation](#operation-name)
it refers to.

Ideally, the span name also contains the destination name of the messages it
refers to. However, a destination name should only be added to the span name
when it is of low cardinality. This is usually the case when the destination
name is a meaningful and manually configured name (like a manually configure
queue or topic name), it is usually not the case if the destination name is an
auto-generated identifier (like a conversation id or an auto-generated name for
an anonymous destination).

> The span name SHOULD consist of the name of the messaging system followed by
> an [operation name](#operation-name). The destination name MAY be appended if
> it is of low cardinality.

##### Examples

* `kafka publish shop.orders`
* `rabbitmq receive print_jobs`
* `AmazonSQS deliver`
* `activemq settle`

#### Operation name

The following operations related to messages are covered by these semantic
conventions:

| Operation name | Description |
|----------------|-------------|
| `publish`      | One ore more messages are provided for publishing to an intermediary. |
| `create`       | A message is created. |
| `receive`      | One or more messages are requested by a consumer. |
| `deliver`      | One or more message are passed to a consumer. |
| `settle`       | One or more message are settled. |

For further details about each of those operations refer, to the [section about trace structure](#trace-structure).

#### Span kind

[Span kinds](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/api.md#spankind)
SHOULD be set according to the following table, based on the operation a span describes.

| Operation name | Span kind|
|----------------|-------------|
| `publish`      | `PRODUCER`, if no `create` spans are present. `INTERNAL` otherwise. |
| `create`       | `PRODUCER` |
| `receive`      | `CONSUMER` |
| `deliver`      | `CONSUMER` |
| `settle`       | `INTERNAL` |

Setting span kinds according to this table ensures, that span links between
consumer and producers always go from a `PRODUCER` span on the producer side to
a `CONSUMER` span on the consumer side. This allows analysis tools to interpret
linked traces without the need of additional semantic hints.

#### Attributes

### System-specific extensions

### Examples

#### Single message producer, single message push-based consumer

```
  PRODUCER                                   CONSUMER

  +------------+         (link)              +------------+
  | Publish m1 | . . . . . . . . . . . . . . | Deliver m1 |
  +------------+                             +------------+
```

#### Single message producer, single message push-based consumer with manual settlement

```
  PRODUCER                                   CONSUMER

  +------------+                             +------------------+
  | Publish m1 | . . . . . . . . . . . . . . | Deliver m1       |
  +------------+                             +-----+-----------++
                                                   | Settle m1 |
                                                   +-----------+
```

#### Single message producer, single message push-based consumer with auto-settlement

```
  PRODUCER                                   CONSUMER

                                           +--------------------------+
                                           | Ambient                  |
  +------------+                           +-+------------+-----------+
  | Publish m1 | . . . . . . . . . . . . . . | Deliver m1 |
  +------------+                             +------------+-----------+
                                                          | Settle m1 |
                                                          +-----------+
```

#### Batch message producer with "Create" spans, single message pull-based consumer

```
  PRODUCER                                        CONSUMER

  +-------------------------+                     +------------+
  | Publish                 |         . . . . . . | Receive m1 |
  +-+-----------+-----------+         .           +------------+
    | Create m1 | . . . . . . . . . . .
    +-----------+-----------+                     +------------+
                | Create m2 | . . . . . . . . . . | Receive m2 |
                +-----------+                     +------------+
```

#### Batch message producer, single message push-based consumer

```
  PRODUCER                                CONSUMER

  +---------+                             +------------+
  | Publish | . . . . . . . . . . . . . . | Deliver m1 |
  +---------+                             +------------+
        .                                 +------------+
        . . . . . . . . . . . . . . . . . | Deliver m2 |
                                          +------------+
```

#### Batch message producer with independent "Create" spans, single message pull-based consumer

```
  PRODUCER                                      CONSUMER

  +-----------------------------------+
  | Ambient                           |
  +-+-----------+---------------------+         +------------+
    | Create m1 | . . . . . . . . . . . . . . . | Receive m1 |
    +-----------+-----------+                   +------------+
                | Create m2 | . . . . . . . .
                +-----------+---------+     .   +------------+
                            | Publish |     . . | Receive m2 |
                            +---------+         +------------+
```

#### Single message producers, batch push-based consumer with process spans

```
  PRODUCER                         CONSUMER

  +------------+
  | Publish m1 | . . . . . . . . . . . . . .
  +------------+                           .
                                   +---------------------------+
                               . . | Deliver                   |
  +------------+               .   +-+------------+------------+
  | Publish m2 | . . . . . . . .     | Process m1 |
  +------------+                     +------------+------------+
                                                  | Process m2 |
                                                  +------------+
```

#### Single message producers, batch pull-based consumer with process spans

```
  PRODUCER                    CONSUMER

  +------------+
  | Publish m1 |. . . .       +-------------------------------------+
  +------------+      .       | Ambient                             |
                      .       +-+---------+-------------------------+
                      . . . . . | Receive |
  +------------+          .     +---------+------------+
  | Publish m2 |. . . . . .               | Process m1 |
  +------------+                          +------------+------------+
                                                       | Process m2 |
                                                       +------------+
```

#### Single message producers, batch pull-based consumer with manual settlement

```
  PRODUCER                    CONSUMER

  +------------+
  | Publish m1 |. . . .       +----------------------------------------+
  +------------+      .       | Ambient                                |
                      .       +-+---------+----------------------------+
                      . . . . . | Receive |
  +------------+          .     +---------+    +-----------+
  | Publish m2 |. . . . . .                    | Settle m1 |
  +------------+                               +-----------+-----------+
                                                           | Settle m2 |
                                                           +-----------+
```

## Future possibilities

### Transport context propagation

One possibility to seamlessly integrate producer/consumer and intermediary
instrumentation in a flexible and extensible way would be the introduction of a
second transport context layer in addition to the creation context layer.

1. The _creation context layer_ allows correlating the producer with the
   consumers of a message, regardless of intermediary instrumentation. The
   creation context is created by the producer and must be propagated to the
   consumers. It must not be altered by intermediaries.

   This layer helps to model dependencies between producers and consumers,
   regardless of the underlying messaging transport mechanism and its
   instrumentation.
2. An additional _transport context layer_ allows correlating the producer and
   the consumer with an intermediary. It also allows to correlate multiple
   intermediaries among each other. The transport context can be changed by
   intermediaries, according to intermediary instrumentations. Intermediaries that
   are not instrumented might simply drop the transport context.

   This layer helps to gain insights into details of the message transport.

This would keep the existing correlation between producers and consumer intact,
while allowing intermediaries to use the transport context to correlate
intermediary instrumentation with existing producer and consumer
instrumentations.

### Standards for context propagation

Currently, instrumentors have to decide how to attach and extract context from
messages in order to fulfill the [requirements for context propagation](#context-propagation).
While preserving the freedom for instrumentor to choose how to propagate
context, in the future these conventions should list recommended ways of how to
propagate context using popular messaging protocols.

Currently several attempts exist to standardize context propagation for different
messaging protocols and scenarios:

* [AMQP](https://w3c.github.io/trace-context-amqp/)
* [MQTT](https://w3c.github.io/trace-context-mqtt/)
* [CloudEvents via HTTP](https://github.com/cloudevents/spec/blob/v1.0.1/extensions/distributed-tracing.md)

Those standards are in draft states and/or are not widely adopted yet. It is
planned to drive those standards to a stable state and to make sure they cover
requirements put forth by these semantic conventions. Finally, these semantic
conventions should give a clear and stable recommendation for each protocol and
scenario.
