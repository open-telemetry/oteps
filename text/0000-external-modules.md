# Instrumentation Ecosystem Management

Short (one sentence) summary, e.g., something that would be appropriate for a [CHANGELOG](https://keepachangelog.com/) or release notes.

## Motivation

For OpenTelemetry to become a de-facto standard in observability there must exist a vast ecosystem of OpenTelemetry components.
Including integrations with various libraries and frameworks in all languages supported by OpenTelemetry.
We cannot possibly expect that all these integrations will be provided by the core maintainers of OpenTelemetry.
We have to have a way to leverage a wider community while still providing our end-users with some way to discover
all available Otel components together with some visibility into their quality.

## Explanation

[OpenTelemetry Registry](https://opentelemetry.io/registry/) serves as a central catalogue of all known Otel components,
both provided by core maintainers of the project and any third party.

In order for a component to be included into Registry its authors have to fill a [self-assessment form](#registry-self-assessment-form).

Registry should allow a clear visibility if component came from a repository in OpenTelemetry organisation.

A component can be removed from the Registry if any declaration from the self-assessment form is violated and not remedied
in a timely manner, provisionally one or two weeks.

## Internal details

We distinguish the following sources of OpenTelemetry components and integrations.

### Native or built-in instrumentations

Any library or framework can use OpenTelemetry API to natively produce Telemetry.
We encourage library authors to submit their library for inclusion into OpenTelemetry Registry.

### Core components

Otel SIGs may provide instrumentations for any libraries that are deemed important enough by SIG’s maintainers.
By doing this the SIG maintainers commit to support these instrumentations (including future versions of the library),
provide updates/fixes, including security patches and guarantee that they comply with Otel semantic conventions and best practices.

Depending on the SIG these core instrumentations may share repository, infrastructure and maintainers with
OpenTelemetry API/SDK implementation for this language or be separate.

All core instrumentations must be included into Otel Registry following the usual process described above.

### Contrib components

Any language SIG may have one or more “contrib” repos containing components contributed by external developers.
This repository may have a separate set of approvers/maintainers than the core API/SDK repo.
SIG maintainers are encouraged to have a simplified process to promote contributors to approver/maintainer role in this repository. 
Contrib repository may leverage the CODEOWNERS functionality of GitHub to assign maintainers to individual packages
even if this means granting write permissions to the whole repo.
The goal should be to distribute the load of reviewing PRs and accepting changes as much as possible.

All components in contrib repository are expected to be included into Otel Registry following the usual process described above.

It should be very easy for external contributors to include their components into contrib repo as opposed to hosting them separately.
This way they can reuse existing infrastructure for testing, publishing, security scanning etc.
Also this will greatly simplify responsibility transfer between different maintainers if their priorities change. 

Language SIGs are encouraged to provide a testing harness to verify that component adheres to OpenTelemetry semantic conventions
and [recommendations for Otel instrumentations design](https://docs.google.com/document/d/1YNRCg9fdjJgZRs56vvf7rfFPk06mhp781sWHYypOaAk/edit#)
when OpenTelemetry starts publishing them.

It may be beneficial to provide even finer-grade separation for contrib instrumentations into “experimental” and “supported” ones.
The latter means that given instrumentation has active (external) maintainers who are committed to supporting this instrumentation on an ongoing basis.

### External components

If component authors for whatever reason want to host their contribution outside the Otel contrib repository they are free to do so.
Their submission for inclusion into OpenTelemetry Registry is still welcomed, subject to the same process described above. 

### Distribution

Whenever OpenTelemetry components are published to any repository other than OpenTelemetry Registry (such as npm registry or Maven Central)
only core and contrib components can be published under "opentelemetry" namespace.
Native and external components are to be published under their own namespace. 

In case Otel SIG provides any kind of "all-in-one" instrumentation distribution (e.g. Java and .NET do) they should include
only core and contrib packages into it.
OpenTelemetry Registry should provide a way to easily obtain a list of these components.
If possible SIG should provide a mechanism to include any external component during target application's build- or runtime.
This may mean a separate language-specific component API that all components are encouraged to implement.

## Trade-offs and mitigations

* How easy it is to get merge permission in contrib repo?
The harder it is, the larger is the maintenance burden on the core team.
The easier it is, the more uncertainty there is about the quality of contributions.
Can every language SIG decide this for themselves or should we decide together?

## Open questions

### Registry self-assessment form
The exact list should be developed separately but at least component's author should declare that
* It uses permissive OSS license
* It does not have any known security vulnerabilities
* It produces telemetry which adheres to OpenTelemetry semantic conventions
* If Otel/SIG provides a testing harness to verify produced telemetry, that test was used and passed 
* Authors commit a reasonable effort into future maintenance of this component
