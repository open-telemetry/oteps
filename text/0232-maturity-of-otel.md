# Framework for defining the maturity of OpenTelemetry and its SIGs

On 08 Mar 2023, the OpenTelemetry GC and TC held an OpenTelemetry Leadership summit, discussing various topics. One of the themes we discussed was establishing standard rules for describing the maturity of the OpenTelemetry project. This OTEP summarizes what was discussed there and is intended to have the wider community provide feedback.

This OTEP builds on what was previously communicated by the project, especially the [Versioning and stability for OpenTelemetry clients](https://opentelemetry.io/docs/reference/specification/versioning-and-stability).

The Collector's [stability levels](https://github.com/open-telemetry/opentelemetry-collector#stability-levels) inspired the maturity levels.

## Motivation

Quite often, the community is faced with the question of the quality and maturity expectations of its diverse set of components. This OTEP aims to bring clarity by establishing a framework to communicate the project's maturity. As the OpenTelemetry project comprises a multitude of SIGs, and each SIG has several components of varying quality, having this framework will help set the right expectations for OpenTelemetry users. The public message we want to convey by stating that a SIG is a tier-1 SIG is that the SIG has both a healthy community and is seen as critical to the project as a whole.

Only once we reach the "stable" level at the project level will we be ready to consider graduating OpenTelemetry at the CNCF, with all of the Tier-1 SIGs being in the context of the graduation, including their Tier-1 Components. This OTEP provides a framework for describing project maturity as a whole.

## Glossary

* *SIG:* an official SIG within the OpenTelemetry project, like the Collector or Java. Given that some Working Groups (WGs) are effectively SIGs, like the Semantic Conventions WG, WGs will be accepted on an individual basis.
* *Tier-1 SIG:* a SIG within the OpenTelemetry project deemed critical by the GC/TC.
* *Candidate Tier-1 SIG:* a SIG that is deemed as critical for the future of the project, but might not be of the same maturity level yet. It's expected to reach that level eventually, and other Tier-1 SIGs should be prepared to accommodate the new component as needed.
* *Component:* module maintained by a SIG like a Collector exporter or an SDK sampler.
* *Tier-1 Component:* a component selected by project maintainers as critical for the SIG. For instance, the OTLP receiver will likely be selected as a Tier-1 Component of the Collector. See Tier-1 Components for more details.
* *Candidate Tier-1 Component:* similar to "Candidate Tier-1 SIG", but for a component within a SIG.
* *Dependency:* an OpenTelemetry SIG or component that is required by another SIG or component. For instance, the OTLP Logs Data Model from the Spec is required for the Collector's Logging support.  

## Explanation

### Maturity levels

SIGs and components of a project MUST have a declared maturity level established by SIG maintainers (SIGs). The maturity of the SIG is assessed by its maintainers, while the maturity of the components might be delegated by the maintainers to code owners. The maturity level for a SIG or component is, at most, the lowest level of its dependencies or priority components. For instance, if the Collector SIG maintainers declare the "otlpreceiver" component as a tier-1 component of the Collector core distribution, and the "otlpreceiver" declares a dependency on the OpenTelemetry Collector API "config" package which is marked with a maturity level of "beta", the "otlpreceiver" can be at most "beta". The OpenTelemetry Collector "core" distribution is also affected so that it can be at most "beta".

Only once all dependencies are marked as "stable" MAY a component be marked as "stable". Only once all Tier-1 Components of a SIG are "stable" MAY that SIG be marked as "stable".

#### Development

Not all pieces of the component are in place yet, and it might not be available for users yet. Bugs and performance issues are expected to be reported. User feedback is desired, especially regarding the user experience (configuration options, component observability, technical implementation details, ...). Configuration options might break often depending on how things evolve. The component SHOULD NOT be used in production.

#### Alpha

This is the default level: any components with no explicit maturity level should be assumed to be "Alpha". The component is ready to be used for limited non-critical production workloads, and the authors of this component welcome user feedback. Bugs and performance problems are encouraged to be reported, but component owners might not work on them immediately. The configuration options might often change without backward compatibility guarantees.

#### Beta

Same as Alpha, but the interfaces (API, configuration, generated telemetry) are treated as stable whenever possible. While there might be breaking changes between releases, component owners should try to minimize them. A component at this stage is expected to have had exposure to non-critical production workloads already during its Alpha phase, making it suitable for broader usage.

#### Stable

The component is ready for general availability. Bugs and performance problems should be reported, and there's an expectation that the component owners will work on them. Breaking changes, including configuration options and the component's output, are not expected to happen without prior notice unless under special circumstances.

#### Deprecated

Development of this component is halted, and no further support will be provided. No new versions are planned, and the component might be removed from its included distributions. Note that new issues will likely not be worked on except for critical security issues. Components that are included in distributions are expected to exist for at least two minor releases or six months, whichever happens later. They also MUST communicate in which version they will be removed.

#### Unmaintained

A component identified as unmaintained does not have an active code owner. Such components may have never been assigned a code owner, or a previously active code owner has not responded to requests for feedback within 6 weeks of being contacted. Issues and pull requests for unmaintained components SHOULD be labeled as such. After 6 months of being unmaintained, these components MAY be deprecated. Unmaintained components are actively seeking contributors to become code owners.

### Specification requirements

The specification should declare all requirements for a specified component to be declared stable. When implementations are still missing features required by the specification, they can be at most Alpha status.

### Tier-1 SIGs

This OTEP recommends the following SIGs to be included as part of the Tier-1 SIGs for OpenTelemetry:

* Spec General
* Semantic Conventions (WG)
* Collector
* Java
* Go
* .NET
* JavaScript

Once this OTEP is approved, the TC and GC members will vote on the individual SIGs from the list above and are free to make further suggestions. A final list of Tier-1 SIGs shall be in the "community" repository.

### Tier-1 Components

When discussing the components' maturity, we often debate what should be included in the evaluation criteria. To help clarify that, we introduce the concept of "Tier-1 SIGs" and "Tier-1 Components". Tier-1 Components are critical for a specific SIG: only when all Tier-1 Components are stable can a project be evaluated as stable. Example:

* The "OTLP receiver" for the OpenTelemetry Collector could potentially be a "Tier-1 Component". Only after this component is declared stable may the Collector be declared stable, provided all other Tier-1 Components are also stable.

Projects MUST list their Tier-1 Components in the main readme file. A Tier-1 Component MAY be hosted under the project's "contrib" repository. Not all components under the main repository (sometimes called the "core" repository) are automatically Tier-1 Components.

A component can be declared "Tier-1" only if they have the following characteristics:

* The component has at least the same maturity level as the SIG
* The component has a clear code owner
* The code owner is active within the project
* The maintainers of the project are comfortable supporting the component should the code owner decide to step down or leave the project

### Tier-1 Signals

Similar to a Tier-1 Component, a Tier-1 Signal must have stable support in order to declare OpenTelemetry as stable. Declaring a signal as Tier-1 means that all projects must also see that signal as required for declaring the project stable.

Not all signals will have the same requirements. For instance, Logs might not provide an instrumentation API specification.

List of Tier-1 signals and their capabilities:

* Logs
  * Data model
  * SDK specification
  * Instrumentation SDKs
  * Bridge API specification
* Metrics
  * Data model
  * API specification
  * Instrumentation APIs
  * SDK specification
  * Instrumentation SDKs
* Traces
  * Data model
  * API specification
  * Instrumentation APIs
  * SDK specification
  * Instrumentation SDKs

The lifecycle of a signal is defined in the [Versioning and stability for OpenTelemetry clients](https://opentelemetry.io/docs/reference/specification/versioning-and-stability). While the stability and versioning of the signals may relate to this OTEP, the previously linked document specifies the details on how to reach the "stable" level mentioned in this OTEP.

The "experimental" level from the signal versioning and stability document relates to the "alpha" and "beta" levels from this OTEP, "stable" and "deprecated" are equivalent to their counterparts on this document, and "removed" is non-existent here.

When a new signal reaches the "stable" level, OpenTelemetry TC and GC will ultimately decide whether to include it as part of OpenTelemetry's Tier-1 components, using the process described in [Promoting SIGs and components](#promoting-sigs-and-components).

### Promoting SIGs and components

SIGs can request to be included in the Tier-1 List by contacting a TC or GC member. The TC/GC member will bring the request to the TC/GC via Slack or Zoom call, who will vote on the matter. A simple majority suffices for including the SIG. The SIG needs to have at least equal maturity as the project. For instance, if OpenTelemetry's maturity is set to Beta, the coming SIG has to be at least Beta as well, so that it doesn't downgrade the main project's maturity.

Components can request project maintainers to be included in the Tier-1 list via a GitHub issue created on the main repository of the SIG. The SIG maintainers will then vote on the matter. A simple majority suffices for including the component. The component needs to have at least equal maturity as the SIG so that it doesn't downgrade the SIG's maturity. If SIG maintainers have strong reasons to include a component with a lower status, they can request an exception with the GC/TC. A simple majority is required to accept the request.

SIGs or components that are seen by the maintainers as critical for the future of the project, like a new signal, are added to the Candidate Tier-1 List. Once this happens, the other Tier-1 SIGs or components need to evaluate whether they need to be changed to accommodate the new candidate.

### Downgrading SIGs and components

While we don't expect SIGs and components to be downgraded, there may be situations where the evolution of the project or SIG is affected by the inaction of specific SIGs or components. For instance, if a new signal is added to the Candidate Tier-1 List and a SIG fails to adopt this new signal within a reasonable time, the GC/TC may choose to remove the SIG from the project's Tier-1 List. To do that, two TC/GC members must sponsor the removal proposal, and a supermajority vote (two-thirds) is required to accept the proposal.

Similarly, if a component becomes unmaintained or isn't of interest to the SIG maintainers anymore, it can be removed from the Tier-1 List. A simple majority vote by the SIG maintainers is required to accept the proposal.

## Trade-offs and mitigations

This approach gives both a top-down and a bottom-up approaches, allowing the GC/TC to determine what are the Tier-1 SIGs for OpenTelemetry as a whole, while still allowing SIG maintainers to determine what's in scope within their areas of expertise.

## Prior art and alternatives

* The specification status has a ["Component Lifecycle"](https://opentelemetry.io/docs/specs/status/) description, with definitions that might overlap with some of the levels listed in this OTEP.
* The same page lists the status of the different parts of the spec
* The ["Versioning and stability for OpenTelemetry clients"](https://opentelemetry.io/docs/specs/otel/versioning-and-stability/#signal-lifecycle) page has a detailed view on the lifecycle of a signal and which general stability guarantees should be expected by OpenTelemetry clients. Notably, it lacks information about maturity of the Collector. This OTEP could be seen as clashing with the last section of that page, "OpenTelemetry GA". But while that page established a point where both OpenTracing and OpenCensus would be considered deprecated, this OTEP here defines the criteria for calling OpenTelemetry "stable" and making that a requirement for a future graduation. This would also make it clear to end-users which parts of the project they can rely on.
* The OpenTelemetry Collector has its own [stability levels](https://github.com/open-telemetry/opentelemetry-collector#stability-levels), which served as inspiration to the ones here.

## Open questions

None at the moment.

## Future possibilities

This change clears the path for a future graduation, by providing an objective way to evaluate what should be included in the graduation scope.
