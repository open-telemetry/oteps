# OpenTelemetry Logs Vision

The following are high-level items that define our long-term vision for 
Logs support in OpenTelemetry project, what we aspire to achieve.

This a vision document that reflects our current desires. It is not a commitment
to implement everything precisely as listed. The primary purpose of this
document is to ensure all contributors work in alignment. As our vision changes
over time maintainers reserve the right to add, modify, and remove items from
this document.

### First-class Citizen

Logs are first-class citizen in observability, along with traces and metrics.
We will aim to have best-in-class support for logs at OpenTelemetry.

### Logs Data Model

We will design a Log Data model that will aim to correctly represent all types
of logs and events. The purpose of the data model is to have a common
understanding of what a log record is, what data needs to be recorded,
transferred, stored and interpreted by a logging system.

Existing log formats can be unambiguously mapped to this data model. Reverse 
mapping from this data model is also possible to the extent that the target log 
format has equivalent capabilities.

We will produce mapping recommendations for commonly used log formats.

### Log Protocol

Armed with the Log Data model we will aim to design a high performance protocol
for logs, which will pursue the same [design goals](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/proto-design-goals.md)
as we had for traces and metrics protocol design.

Most notably the protocol will aim to be highly reliable, have low resource
consumption, be suitable for all participant nodes, ensure high throughput,
allow backpressure signalling and be load-balancer friendly.

### Unified Collection

We aim to support high-performance, unified agent and collector that supports
logs, traces and metrics in one package, symmetrically and uniformly for all 3
types of telemetry data (see also [Collector vision](https://github.com/open-telemetry/opentelemetry-collector/blob/8310e665ec1babfd56ca5b1cfec91c1f997f4f2c/docs/vision.md)).

## Cloud Native

We will have best-in-class support for logs emitted in cloud native environments
(e.g. Kubernetes, serverless, etc), including legacy applications running
in such environments.

## Support Legacy

We will produce guidelines on how legacy applications can emit logs in a
manner that makes them compatible with OpenTelemetry's approach and enables
telemetry data correlation. We will also have a reasonable story around
logs that are emitted by sources over which we may have no control and which
emit logs in pre-defined formats via pre-defined mediums (e.g. flat file logs,
Syslog, etc).

We will have technical solutions or guidelines for using popular logging
libraries in OpenTelemetry-compatible manner and we may produce logging
libraries for languages where gaps exist.

### Auto-instrumenting

To enable functionality that requires modification of how logs are emitted we
will work on auto-instrumenting solutions to reduce the adoption barrier for
existing deployments.

### Applicable to All Log Sources

Logging support at OpenTelemetry will be applicable to all sorts of log sources:
system logs, infrastructure logs, third-party and first-party application logs.

### Standalone and Embedded Logs

OpenTelemetry will support both logs embedded inside [Spans](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/api-tracing.md#span)
and standalone logs recorded elsewhere.

## Prior art and alternatives

Prior art in logging industry is nearly impossible to enumerate. There are
countless logging libraries, collection agents, network protocols, open-source
and proprietary backends.

We recognize this fact and aim to make our proposals in a manner that honours
valid legacy use-cases, while at the same time suggests better solutions
where they are due.
