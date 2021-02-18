# Scenarios for Metrics API/SDK Prototyping

With the stable release of the tracing specification, the OpenTelemetry
community is willing to spend more energy on metrics API/SDK. The goal is to get
the metrics API/SDK specification to
[`Experimental`](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#experimental)
state by end of 6/2021, and make it
[`Stable`](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#stable)
before end of 2021:

* By end of 6/2021, we should have a good confidence that we can recommend
  language client owners to work on metrics preview release. This means starting
  from 7/1/2021 the specification should not have major surprise or big changes
  which would thrust the language client maintainers.

* By end of 10/2021, we should mark the metrics API/SDK specification as
  [`Feature-freeze`](https://github.com/open-telemetry/opentelemetry-specification/blob/1afab39e5658f807315abf2f3256809293bfd421/specification/document-status.md#feature-freeze),
  and focusing on bug fixing or editorial changes.

* By end of 2021, we want to have a stable release of metrics API/SDK
  specification, with multiple language SIGs providing RC (release candidate) or
  GA (general available) clients.

In this document, we will focus on two scenarios that we use for prototyping
metrics API/SDK. The goal is to have two scenarios which clearly capture the
major requirements, so we can work with language client SIGs to prototype,
gather the learnings, determine the scopes and stages. Later the scenarios can
be used as examples and test cases for all the language clients.

Here are the languages we've agreed to use during the prototyping:

* C#
* Java
* Python

Instead of boiling the ocean, we will need to divide the work into multiple
stages:

1. Do the end-to-end prototype to get the overall understanding of the problem
   domain. We should also clarify the scope and be able to articulate it
   precisely during this stage, here are some examples:

    * Why do we want to introduce brand new metrics APIs versus taking a well
      established API (e.g. Prometheus and Micrometer), what makes OpenTelemetry
      metrics API different (e.g. Baggage)?
    * Do we need to consider OpenCensus Stats API shim, or this is out of scope?

2. Focus on a core subset of API, cover the end-to-end library instrumentation
   scenario. At this stage we don't expect to cover all the APIs as some of them
   might be very similar (e.g. if we know how to record an integer, we don't
   have to work on float/double as we can add them later by replicating what
   we've done for integer).

3. Focus on a core subset of SDK. This would help us to get the end-to-end
   application.

4. Replicate stage 2 to cover the complete set of APIs.

5. Replicate stage 4 to cover the complete set of SDKs.

## Scenario 1: Grocery

The **Grocery** scenario covers how a developer could use metrics API and SDK in
a final application. It is a self-contained application which covers:

* How to instrument the code in a vendor agnostic way
* How to configure the SDK and exporter

Considering there might be multiple grocery stores, the metrics we collect will
have the store name as a dimension - which is fairly static (not changing while
the store is running).

The store has plenty supply of potato and tomato, with the following price:

* Potato: $1.0 / ea
* Tomato: $3.0 / ea

Each customer has a unique name (e.g. customerA, customerB), a customer could
come to the store multiple times. Here goes the Python snippet:

```python
store = GroceryStore("Portland")
store.process_order("customerA", {"potato": 2, "tomato": 3})
store.process_order("customerB", {"tomato": 10})
store.process_order("customerC", {"potato": 2})
store.process_order("customerA", {"tomato": 1})
```

When the store is closed, we will report the following metrics:

### Order info

| Store    | Customer  | Number of Orders | Amount (USD) |
| -------- | --------- | ---------------- | ------------ |
| Portland | customerA | 2                | 14.0         |
| Portland | customerB | 1                | 30.0         |
| Portland | customerC | 1                | 2.0          |

### Items sold

| Store    | Customer  | Item   | Count |
| -------- | --------- | ------ | ----- |
| Portland | customerA | potato | 2     |
| Portland | customerA | tomato | 4     |
| Portland | customerB | tomato | 10    |
| Portland | customerC | potato | 2     |

## Scenario 2: HTTP Server

The _HTTP Server_ scenario covers how a library developer X could use metrics
API to instrument a library, and how the application developer Y can configure
the library to use OpenTelemetry SDK in a final application. X and Y are working
for different companies and they don't communicate. The demo has two parts - the
library (HTTP lib owned by X) and the server app (owned by Y):

* How developer X could instrument the library code in a vendor agnostic way
  * Performance is critical for X
  * X doesn't know which metrics and which dimensions will Y pick
  * X doesn't know the aggregation time window, nor the final destination of the
    metrics
* How developer Y could configure the SDK and exporter
  * How should Y hook up the metrics SDK with the library
  * How should Y configure the time window(s) and destination(s)
  * How should Y pick the metrics and the dimensions

### Library Requirements

The library developer (developer X) will expose the following metrics out of
box:

### Pull Metrics

These are pull metrics - the value is always available, and is only reported and
collected based on the ask from consumer(s). If there is no ask from the
consumer, the value will not be reported at all (e.g. there is no API call to
fetch the room temperature unless someone is asking for the room temperature).

#### Process CPU Usage

Note: the **Host Name** should leverage [`OpenTelemetry
Resource`](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/sdk.md),
so it should be covered by the metrics SDK rather than API, and strictly
speaking it is not considered as a "dimension" from the SDK perspective.

| Host Name | Process ID | CPU% [0.0, 100.0] |
| --------- | ---------- | ----------------- |
| MachineA  | 1234       | 15.3              |

#### System CPU Usage

| Host Name | CPU% [0, 100] |
| --------- | ------------- |
| MachineA  | 30            |

#### Server Room Temperature

| Host Name | Temperature (F) |
| --------- | --------------- |
| MachineA  | 65.3            |

### Push Metrics

These are the push metrics - the value is reported (via the API) when it is
available, and collected (via the SDK) based on the ask from consumer(s). If
there is no ask from the consumer, the API will be no-op and the data will be
dropped on the floor.

#### Received HTTP Requests

Note: the **Client Type** is passed in via the [`OpenTelemetry
Baggage`](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/baggage/api.md),
strictly speaking it is not part of the metrics API, but it is considered as a
"dimension" from the metrics SDK perspective.

| Host Name | Process ID | Client Type | HTTP Method | HTTP Host | HTTP Flavor | Peer IP   | Peer Port | Host IP   | Host Port |
| --------- | ---------- | ----------- | ----------- | --------- | ----------- | --------- | --------- | --------- | --------- |
| MachineA  | 1234       | Android     | GET         | otel.org  | 1.1         | 127.0.0.1 | 51327     | 127.0.0.1 | 80        |
| MachineA  | 1234       | Android     | POST        | otel.org  | 1.1         | 127.0.0.1 | 51328     | 127.0.0.1 | 80        |
| MachineA  | 1234       | iOS         | PUT         | otel.org  | 1.1         | 127.0.0.1 | 51329     | 127.0.0.1 | 80        |

#### HTTP Server Duration

Note: the server duration is only available for **finished HTTP requests**.

| Host Name | Process ID | Client Type | HTTP Method | HTTP Host | HTTP Status Code | HTTP Flavor | Peer IP   | Peer Port | Host IP   | Host Port | Duration (ms) |
| --------- | ---------- | ----------- | ----------- | --------- | ---------------- | ----------- | --------- | --------- | --------- | --------- | ------------- |
| MachineA  | 1234       | Android     | GET         | otel.org  | 200              | 1.1         | 127.0.0.1 | 51327     | 127.0.0.1 | 80        | 8.5           |
| MachineA  | 1234       | Android     | POST        | otel.org  | 304              | 1.1         | 127.0.0.1 | 51328     | 127.0.0.1 | 80        | 100.0         |

### Application Requirements

The application owner (developer Y) would only want the following metrics:

* [System CPU Usage](#system-cpu-usage) reported every 5 seconds
* [Server Room Temperature](#server-room-temperature) reported every minute
* [HTTP Server Duration](#http-server-duration), reported every 5 seconds, with
  a subset of the dimensions:
  * Host Name
  * HTTP Method
  * HTTP Host
  * HTTP Status Code
  * Client Type
  * 90%, 95%, 99% and 99.9% server duration
* HTTP request counters, reported every 5 seconds:
  * Total number of received HTTP requests
  * Total number of finished HTTP requests
  * Number of currently-in-flight HTTP requests (concurrent HTTP requests)

  | Host Name | Process ID | HTTP Host | Received Requests | Finished Requests | Concurrent Requests |
  | --------- | ---------- | --------- | ----------------- | ----------------- | ------------------- |
  | MachineA  | 1234       | otel.org  | 630               | 601               | 29                  |
  | MachineA  | 5678       | otel.org  | 1005              | 1001              | 4                   |
* Exception samples (exemplar) - in case HTTP 5xx happened, developer Y would
  want to see a sample request with all the dimensions (IP, Port, etc.)
