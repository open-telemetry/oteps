# Scenarios for Tracing semantic conventions for messaging

This document aims to capture scenarios and a road map, both of which will
serve as a basis for [stabilizing](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#stable)
the [existing semantic conventions for messaging](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/messaging.md),
which are currently in an [experimental](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#experimental)
state. The goal is to declare messaging semantic conventions stable before the
end of 2021.

## Motivation

Many observability scenarios involve messaging systems. For Distributed Tracing
to be useful across the entire scenario, having good observability for
messaging operations is critical. To achieve this, OpenTelemetry must provide
stable conventions and guidelines for instrumenting messaging systems.

Bringing the existing experimental semantic conventions for messaging to a
stable state is a crucial step for users and instrumentation authors, as it
allows them to rely on [stability guarantees](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#not-defined-semantic-conventions-stability),
and thus to ship and use stable instrumentation.

## Roadmap

| Description | Done By     |
|-------------|-------------|
| This OTEP, consisting of scenarios and a proposed roadmap, is approved and merged. | 09/30/2021 |
| [Stability guarantees](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#not-defined-semantic-conventions-stability) for semantic conventions are approved and merged. This is not strictly related to semantic conventions for messaging but is a prerequisite for stabilizing any semantic conventions. | 09/30/2021 |
| An OTEP proposing a set of attributes and conventions covering the scenarios in this document is approved and merged. | 10/29/2021 |
| Proposed specification changes are verified by prototypes for the scenarios and examples below. | 11/15/2021 |
| The [specification for messaging semantic conventions for tracing](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/messaging.md) are updated according to the OTEP mentioned above and are declared [stable](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#stable). | 11/30/2021 |

## Terminology

To leverage existing standards, the terminology used in this document is based
on the [CloudEvents specification](https://github.com/cloudevents/spec/blob/v1.0.1/spec.md).
CloudEvents is hosted by the CNCF and provides a specification for describing
event data in common formats to provide interoperability across services,
platforms and systems.

### Message

A "message" is a transport wrapper for the transfer of information. The
information is a combination of data and metadata.  Messages may be transferred
directly between parties or via one or more intermediaries.  A message may
carry annotations that are directed at intermediaries on the message path.
Messages are uniquely identifiable.

In the strict sense, a _message_ is an item of data that is sent to a specific
destination, whereas an _event_ is a signal emitted by a component upon
reaching a given state. This document is agnostic of those differences and uses
the term "message" in a wider sense to cover both concepts.

### Producer

The "producer" is a specific instance, process or device that creates and
publishes a message. "Publishing" is the process of sending a message or batch
to the intermediary or consumer.

### Consumer

A "consumer" receives the message and acts upon it. It uses the context and
data to execute some logic, which might lead to the occurrence of new events.

The consumer receives, processes, and settles a message. "Receiving" is the
process of obtaining a message from the intermediary, "processing" is the
process of acting on the information a message contains, "settling" is the
process of notifying an intermediary that a message was processed successfully.

### Intermediary

An "intermediary" receives a message to forward it to the next receiver, which
might be another intermediary or a consumer.

## Scenarios

Producing and consuming a message involves five stages:

```
PRODUCER

Create
  |                            CONSUMER
  v        +--------------+                   
Publish -> | INTERMEDIARY | -> Receive
           +--------------+       |
                  ^               v
                  .            Process
                  .               |
                  .               v
                  . . . . . .  Settle
```

1. The producer creates a message.
2. The producer publishes the message to an intermediary.
3. The consumer receives the message from an intermediary.
4. The consumer processes the message.
5. The consumer settles the message by notifying the intermediary that the
   message was processed. In some cases, the message is settled before it is
   processed, or it is settled automatically when it is received.

The semantic conventions need to define how to handle failures and retries in
all stages that interface with the intermediary: publish, receive and settle.

Based on this model, the following scenarios capture major requirements and
can be used for prototyping, as examples, and as test cases.

### Individual settlement

Individual settlement systems imply independent logical message flows. A single
message is created and published in the same context, and it's delivered,
consumed, and settled as a single entity. Each message needs to be settled
individually.

Transport batching can be treated as a special case: messages can be
transported together as an optimization, but are produced and consumed
individually.

```
+---------+ +---------+ +---------+ +---------+ +---------+ +---------+
|Message A| |Message B| |Message C| |Message D| |Message E| |Message F|
+---------+ +---------+ +---------+ +---------+ +---------+ +---------+
 Settled                 Settled                             Settled 
```

#### Examples

1. RabbitMQ: TODO

### Checkpoint-based settlement

Messages are processed as a stream and settled to specific checkpoints. A
checkpoint points to a position of the stream up to which messages were
processed and settled. Messages cannot be settled individually, instead, the
checkpoint needs to be forwarded.

Checkpoint-based settlement systems are designed to efficiently receive and
settle batches of messages. However, it is not possible to settle messages
independent of their position in the stream (e. g., if message B is located at
a later position in the stream than message A, then message B cannot be settled
without also settling message A).

```
                               Checkpoint
                                   | 
                                   v                                
+---------+ +---------+ +---------+ +---------+ +---------+ +---------+
|Message A| |Message B| |Message C| |Message D| |Message E| |Message F|
+---------+ +---------+ +---------+ +---------+ +---------+ +---------+
                     <---  Settled
```

#### Examples

1. The following configurations should be instrumented and tested for Kafka or
   a similar messaging system:

   * 1 producer, 2 consumers in the same consumer group
   * 1 producer, 2 consumers in different consumer groups
   * 2 producers, 2 consumers in the same consumer group

   Each of the producers produces a continuous stream of messages.

## Open questions

The following areas are considered out-of-scope of a first stable release of
semantic conventions for messaging. While not being explicitly considered for
a first stable release, it is important to ensure that this first stable
release can serve as a solid foundation for further improvements in these areas.

### Sampling

The current experimental semantic conventions rely heavily on span links as
a way to correlate spans. This is necessary, as several traces are needed to
model the complete path that a message takes through the system. With the currently
available sampling capabilities of OpenTelemetry, it is not possible to ensure
that a set of linked traces is sampled. As a result, it is unlikely to sample a
set of traces that covers the complete path a message takes.

Solving this problem requires a solution for sampling based on span links,
which is not in scope for this OTEP.

### Instrumenting intermediaries

Instrumenting intermediaries can be valuable for debugging configuration or
performance issues, or for detecting specific intermediary failures.

Stable semantic conventions for instrumenting intermediaries can be provided at
a future point in time, but are not in scope for this OTEP.

### Metrics

Messaging semantic conventions for tracing and for metrics overlap and should
be as consistent as possible. However, semantic conventions for metrics will be
handled separately and are not in scope for this OTEP.

### In-memory queues or channels

Messaging semantic conventions are not meant for instrumenting in-memory queues
and channels but are intended for inter-application systems. In-memory queues
and channels exist in many variations which can be very different from
inter-application messaging systems, furthermore, requirements for the analysis
and visualization of distributed traces are different. For those reasons, it
makes sense to treat both concepts differently in the context of distributed
tracing.

## Further reading

* [CloudEvents](https://github.com/cloudevents/spec/blob/v1.0.1/spec.md)
* [Message-Driven (in contrast to Event-Driven)](https://www.reactivemanifesto.org/glossary#Message-Driven)
* [Existing semantic conventions for messaging](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/messaging.md)
