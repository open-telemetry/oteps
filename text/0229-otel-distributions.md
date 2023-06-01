# OpenTelemetry Collector Distributions

Design and specification of OpenTelemetry Collector distributions.

## Motivation

The OpenTelemetry Collector is currently distributed by the OpenTelemetry project through two distributions, core and contrib.

In numerous discussions with the community, maintainers and a few SIG meetings, it has emerged that those distributions
emerged organically. We discuss them in detail below.

This OTEP aims to clarify what distributions the OpenTelemetry project should distribute. The OTEP will elicit all requirements
and dependencies attached to the current distributions, discuss the tooling in place, and propose a way forward.

This OTEP also aims to become the center of discussions for the requirements and needs of the community, rather than
seeing this issue being rehashed.

### Core distribution
The core distribution is defined in the [opentelemetry-collector-releases](https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol) repository.

The core distribution maps to well established technologies that predate OpenTelemetry, such as Jaeger, Zipkin, Prometheus and OpenCensus.

There are conflicting views on the usage of the core distribution:
* The core distribution is aimed to represent what "[otelcol contributors] expect most users to use daily". Some have made the [point](https://github.com/open-telemetry/opentelemetry-collector-releases/issues/337#issuecomment-1557855969).
* Taking the previous point, the core distribution needs to satisfy core use cases offered by OpenTelemetry, such as supporting reading log files. However, that is not supported at this time. [See](https://github.com/open-telemetry/opentelemetry-collector-releases/issues/337).
* The core distribution is aimed to only include stable components.
* The core distribution should [only include components in the opentelemetry-collector repository](https://github.com/open-telemetry/opentelemetry-collector-releases/issues/337#issuecomment-1561532445).

It has been proposed to let go of the core distribution.

### Contrib distribution

The contrib distribution aims to contain all the components currently hosted in the opentelemetry-collector-contrib repository.

This is however incomplete. As of [2021](https://github.com/open-telemetry/opentelemetry-collector-releases/issues/46), it was already noted a number of components were missing.

The delta has increased since as new components are added one by one to the contrib manifest. An example is this [discussion](https://github.com/open-telemetry/opentelemetry-collector-contrib/discussions/21483), where the user is expecting the component to be present.

There is no consensus for sure when a component should be included in contrib. Components use [stability levels](https://github.com/open-telemetry/opentelemetry-collector#stability-levels) to help end users understand the maturity of the component.

Typically, a component is not included in the contrib distribution before it's at least considered alpha.
There is no automatic promotion of the component from alpha status to be listed in the contrib distribution.
It is unclear if components which are deprecated or unmaintained must be removed from the contrib distribution.

### Core and contrib repositories

The [opentelemetry-collector repository](https://github.com/open-telemetry/opentelemetry-collector) (Core) 
and [opentelemetry-collector-contrib repository](https://github.com/open-telemetry/opentelemetry-collector-contrib) (Contrib) follow
different processes and have different maintainers, approvers and triagers.

The core repository uses code coverage checks to continously ratchet up code quality.

The core repository doesn't have a process to add a new component. The core repository has less activity. The core repository is more mature and aims to release 1.0.

The core repository issues a release every 2 weeks. It only releases components that have changed rather than release all the contents of the repository.

The contrib repository used code coverage and will use it again. It has over 170 different components.

There is a set of processes by which codeowners who are outside of the committers group can manage changes specific to their component.

Issues filed against a component get routed to the codeowners via a comment pinging them.

The maintainers of the repository use Dependabot and now Renovate to make sure all the dependencies of all the components are always patched to the latest versions.

The maintainers make a release every 2 weeks during which they tack on to the core release, upgrading all components.

There is confusion between the manifest defined in the opentelemetry-collector-releases repository and the manifest used for integration testing in cmd/otelcontribcol.

It is becoming increasing difficult to continue to support such a wide set of components, and new components continue to be added at a fast clip.

This [issue](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21494) tracks some of the pains involved and makes the point that the repository cannot continue to grow indefinitely.

### Existing tooling

#### ocb

The OpenTelemetry project has built a [tool](https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder) 
to distribute a complete distribution.

The tool relies on a configuration file that represents the distribution content as a manifest represented in YAML format.

The tool generates the go code of the executable.

#### Status metadata

The contrib repository is adopting a convention of relying on a metadata.yaml file in the component folder.

This file contains status information, such as stability levels per pipeline type. This information could be used
programmatically to determine if a component should be listed in a distribution.

#### OpenTelemetry Collector Releases repository

This repository contains the manifests of the core and contrib distributions.

It contains the default configuration files for both distributions.

It contains the Dockerfile for each distro as well.

### Other distributions

Per [this issue](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21383), it is now supported to list vendor-supported distributions of OpenTelemetry.
Those distributions are [listed](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/cmd/mdatagen/statusdata.go#L14) as part of the mdatagen tool so users can simply mention the distribution name in the metadata.yaml status section.

== TODO ==
## Explanation

Explain the proposed change as though it was already implemented and you were explaining it to a user. Depending on which layer the proposal addresses, the "user" may vary, or there may even be multiple.

We encourage you to use examples, diagrams, or whatever else makes the most sense!

## Internal details

From a technical perspective, how do you propose accomplishing the proposal? In particular, please explain:

* How the change would impact and interact with existing functionality
* Likely error modes (and how to handle them)
* Corner cases (and how to handle them)

While you do not need to prescribe a particular implementation - indeed, OTEPs should be about **behaviour**, not implementation! - it may be useful to provide at least one suggestion as to how the proposal *could* be implemented. This helps reassure reviewers that implementation is at least possible, and often helps them inspire them to think more deeply about trade-offs, alternatives, etc.

## Trade-offs and mitigations

What are some (known!) drawbacks? What are some ways that they might be mitigated?

Note that mitigations do not need to be complete *solutions*, and that they do not need to be accomplished directly through your proposal. A suggested mitigation may even warrant its own OTEP!

## Prior art and alternatives

What are some prior and/or alternative approaches? For instance, is there a corresponding feature in OpenTracing or OpenCensus? What are some ideas that you have rejected?

## Open questions

What are some questions that you know aren't resolved yet by the OTEP? These may be questions that could be answered through further discussion, implementation experiments, or anything else that the future may bring.

## Future possibilities

What are some future changes that this proposal would enable?
