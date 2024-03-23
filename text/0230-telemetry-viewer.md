# OTV: OpenTelemetry Viewer

_A local explorer for OpenTelemetry data, components, and endpoints._

## Motivation

OpenTelemetry offers a rich, highly customizable, and highly configurable
ecosystem of tooling, SDKs, APIs, and instrumentation libraries. However, with
this complexity comes a cost -- the barrier to entry for new users can be very
high, with significant cycle time required in order to understand how their
instrumentation code changes affect the instrumentation emitted by their
service. In addition, operators often find it challenging to understand
OpenTelemetry configurations at the Collector or in an SDK. While OpAMP provides
an API that can help with this, it doesn't provide a management plane or
visualization components to see and modify configurations in the browser.

To address these gaps, and others, I propose a new 'OpenTelemetry Viewer'
component that can be built into a collector and provides in-memory storage,
viewing, and modification of OpenTelemetry data and components.

## Explanation

Different users of a telemetry system have different needs and expectations for
debugging instrumentation. Currently, developers have two options for
quick feedback - using a debug exporter at the collector or SDK, or using an
existing analysis tool (open source or proprietary). Both of these options have
drawbacks - the debug exporter presents an overwhelming amount of text-based
data, and depending the characteristics of a development environment, it may be
challenging to stand up and use a local suite of open source analysis tools
(such as Jaeger, Prometheus, and OpenSearch) or use a commercial tool. Existing
options are optimized for viewing, querying, and analyzing data across hundreds
or thousands of sources, not for understanding "what attributes are my metrics
emitting?" or "what type of data are my instrumentation libraries emitting?".

Reduced cycle time (both in a DevOps sense and also in a more general reading of
the word) is a contributor to quality and resiliency of software and human
systems. Being able to quickly get feedback about if your changes are having the
desired effect or not is invaluable, especially for developers that are
beginning to instrument their services for observability.

As custom instrumentation and data transformation becomes a larger and larger
part of the OpenTelemetry story, enabling these fast feedback loops is critical
for the project. However, we must also balance this against other tools in the
ecosystem. Thus, I proprose the following set of criteria that will guide the
implementation of this OTEP:

- To be consistent with our existing stance on vendor agnosticism, any component
  developed cannot implement persistent storage. Storage must be local
  (constrained to the machine where OTV is running or accesssed from), and can
  only persist as part of a session (for example, between refreshes on a browser
  page or collector restarts).
- We will not implement a query language or semantics as part of this project.
  Data can be filtered and projected, but without a bespoke or pre-built query
  language.

With that said, the requirements of OTV are as follows:

- An in-memory data store constrained by size and time. For example, a ring
  buffer. This store can persist to local disk for persistence through collector
  reboots, but is specifically not designed for long-term storage of data.
- A web UI that displays data in the store, with user-configurable options for
  grouping, filtering, and sorting data. This UI also should be able to
  visualize metrics appropriately (as a line, bar, or big number chart).
- A web UI that displays a list of discovered OpAMP components, and allows for
  the viewing and editing of their configurations.
- A 'hot reload' function that allows for configuration changes to be applied to
  the underlying collector, so operators can tune OTTL transformations and see
  the changes immediately.

## Internal details

Broadly, the implementation of this viewer should be a collector extension that
exposes a simple web portal for viewing data along with some sort of data store
to hold the data emitted. This extension could be bundled with specific
collector releases, or brought in via the collector builder.

There should be a few options for distribution of the extension -

1. A 'default' collector distribution from the project that includes a basic
   collector configuration and the viewer extension.
2. Using the collector builder to create a custom image that includes this
   extension as well as other custom components.
3. A standalone binary that can be installed on a local machine or in a
   codespace or other development environment.

These are not an exhaustive list of deployment options, and I posit that the
community will create other strategies as well.

## Trade-offs and mitigations

There are two points I'd like to address in terms of trade-offs around this
component.

- _Jaeger, Prometheus, and other CNCF Observability Projects_

Future work on Jaeger and Prometheus promises to bring these tools more into
alignment with OpenTelemetry as a 'default choice' for data storage, query,
visualization, and workflows. This is cause for celebration, to be sure.
However, as noted above, these projects are fundamentally designed to scale out
for production uses of thousands or millions of data points a minute, and
potentially thousands of users. They are not designed for the individual
developer or operator.

This component specifically does not seek to replace either of these tools, and
I believe this OTEP helps define the boundary of tooling that we plan to build
and support going forward.

- _Why not let the community figure it out for themselves?_

Since this OTEP was originally written, there have been examples of
community-built tools (such as OTelBin, which visualizes Collector
configurations) and vertically integrated solutions like .NET Aspire which are
built on top of OpenTelemetry concepts or components. It's entirely possible
that if we do nothing, someone will come up with a solution that users coalesce
around. However, I believe we have a responsibility to define the types of tools
that are useful, and also to offer potential contributors a variety of projects
to work on. OTV would be a great place for 'traditional' application developers
and front-end developers to make contributions to OpenTelemetry, increasing our
contributor base and contributor diversity.

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

In general, both commercial tools and extant open source observability tools are
not designed for the specific use case of allowing a developer to quickly get
feedback on their instrumentation code, or their observability configuration and
pipeline.

Another example of this pattern in the cloud-native ecosystem is the Kubernetes
Dashboard. The dashboard is not a default part of a Kubernetes install, and it's
often superseded in production deployments by managed solutions or other tools
(for example, GKE provides a management UI, and command line tools like k9s
exist). However, by providing this component, Kubernetes is able to provide a
solution for developers and operators who need a simple GUI to diagnose and
visualize their cluster, its pods, etc.

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
