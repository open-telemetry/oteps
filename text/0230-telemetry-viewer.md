# Telemetry Viewer for Developers

_A local telemetry viewer to aid in instrumentation and debugging of pipelines._

## Motivation

OpenTelemetry offers a rich, highly customizable, and highly configurable
ecosystem of tooling, SDKs, APIs, and instrumentation libraries. However, with
this complexity comes a cost -- the barrier to entry for new users can be very
high, with significant cycle time required in order to understand how their
instrumentation code changes affect the instrumentation emitted by their
service.

A 'local development' experience for OpenTelemetry would aid in reducing this
cycle time and understandability burden from developers.

## Explanation

Different users of a telemetry system have different needs and expectations for
debugging instrumentation. Currently, developers and coders have two options for
quick feedback - using a logging exporter at the collector or SDK, or using an
existing analysis tool (open source or proprietary). Both of these options have
drawbacks - the logging exporter presents an overwhelming amount of text-based
data, and depending the characteristics of a development environment, it may be
challenging to stand up and use a local suite of open source analysis tools
(such as Jaeger, Prometheus, and OpenSearch) or use a commercial tool.

Reduced cycle time (both in a DevOps sense and also in a more general reading of
the word) is a contributor to quality and resiliency of software and human
systems. Being able to quickly get feedback about if your changes are having the
desired effect or not is invaluable, especially for developers that are
beginning to instrument their services for observability.

The goal of this OTEP is to define a set of requirements for a solution to this
problem. Ultimately, the vision is that a developer would be able to use a
collector extension to view the following:

- Metrics, Trace, and Log Data collected over the last X minutes.
- The current configuration, pipelines, and operating telemetry (logs, metrics, traces) of the collector.
- A list of instrumentation libraries, agents, or other ecosystem components in
  use by the pipelines.
- All attribute and resource keys seen by the collector over the last X minutes.

## Internal details

Broadly, the implementation of this viewer should be a collector extension that
exposes a simple web portal for viewing data along with some sort of data store
to hold the data emitted. This extension could be bundled with specific
collector releases, or brought in via the collector builder.

## Trade-offs and mitigations

There are two major trade-offs this makes in terms of the ecosystem; One, it
brings this component into the purview of the OpenTelemetry organization rather
than leaving it entirely to both independent or commercial development efforts.
Two, it could seem to presage a more opinionated approach to the collector as a
product offering rather than as a component or piece of middleware.

To the former point, I believe that the project has a responsibility to define
requirements and signposts for tooling that we believe would be useful in order
to not only grow our developer community, but also our contributor community.

To the latter, I would argue that if we seek to become a 'native instrumentation
layer' for cloud-native systems, it is incumbent upon us to be opinionated about
how OpenTelemetry should be used, and to provide the development community with
tooling that makes their lives easier.

## Prior art and alternatives

The biggest piece of prior art or alternative implementation here is the
existence of open source observability and monitoring projects such as
Prometheus, Grafana, etc. The counter-argument against this proposal is that by
creating a secondary toolchain for developers, we would not be skilling them up
in popular existing tools, and these existing tools satisfy the requirements of
this proposal already.

To this, I submit the following points:

1. While Prometheus and Jaeger do provide powerful analysis tools and can both
   be run fairly easily in a local environment, they are not necessarily built
   for the purpose of this proposal as-is.
2. There is no CNCF tooling available for Dashboard or Log Storage/Querying;
   Currently, Grafana and OpenSearch seem to be the most popular and open
   source-iest options here.
3. There is a difference in effort requireed for a developer to add a local
   observability stack vs. swapping out a collector binary, and it's a pretty
   significant difference.

## Open questions

- Are we able to use existing tools (such as
  [otel-desktop-viewer](https://github.com/open-telemetry/community/issues/1515))
  as a starting point for this proposal?
- Do we have maintainers and leaders that can step up to drive this?
- What does success look like?
- Are we signing up for a long-term maintenance problem by taking this on?

## Future possibilities

With appropriate design (i.e., good APIs and interfaces), this could be used by
other local tooling such as VSCode or JetBrains to implement a language server
or other integration endpoints. This could enable integrations between
OpenTelemetry and IDEs themselves in a vendor-agnostic way.
