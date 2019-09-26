# OpenTelemetry
[![Gitter chat][gitter-image]][gitter-url]
[![Build Status][circleci-image]][circleci-url]

# OpenTelemetry RFCs
## Evolving OpenTelemetry at the speed of Markdown

OpenTelemetry uses an "RFC" (request for comments) process for proposing changes to the [OpenTelemetry specification](https://github.com/open-telemetry/opentelemetry-specification).

### Table of Contents

* [What changes require an RFC?](#what-changes-require-an-rfc)
* [RFC scope](#rfc-scope)
* [Writing an RFC](#writing-an-rfc)
* [Submitting the RFC](#submitting-the-rfc)
* [Integrating the RFC into the Spec](#integrating-the-rfc-into-the-spec)
* [Implementing the RFC](#implementing-the-rfc)
* [Changes to the RFC process](#changes-to-the-rfc-process)
* [Background on the OpenTelemtry RFC process](#background-on-the-opentelemetry-rfc-process)

### What changes require an RFC?

The OpenTelemetry RFC process is intended for changes that are **cross-cutting** - that is, applicable across *languages* and *implementations* - and either **introduce new behaviour**, **change desired behaviour**, or otherwise **modify requirements**.

In practice, this means that RFCs should be used for such changes as:

* New tracer configuration options
* Additions to span data
* New metric types
* Modifications to extensibility requirements

On the other hand, they do not need to be used for such changes as:

* Bug fixes
* Rephrasing, grammatical fixes, typos, etc.
* Refactoring
* Things that affect only a single language or implementation

**Note:** The above lists are intended only as examples and are not meant to be exhaustive. If you don't know whether a change requires an RFC, please feel free to ask!

#### Extrapolating cross-cutting changes

Sometimes, a change that is only immediately relevant within a single language or implementation may be indicative of a problem upstream in the specification. We encourage you to add an RFC if and when you notice such cases.

### RFC scope

While RFCs are intended for "significant" changes, we recommend trying to keep each RFC's scope as small as makes sense. A general rule of thumb is that if the core functionality proposed could still provide value without a particular piece, then that piece should be removed from the proposal and used instead as an *example* (and, ideally, given its own RFC!).

For example, an RFC proposing configurable sampling *and* various samplers should instead be split into one RFC proposing configurable sampling as well as an RFC per sampler.

### Writing an RFC

* First, [fork](https://help.github.com/en/articles/fork-a-repo) this [repo](https://github.com/open-telemetry/oteps).
* Copy [`0000-template.md`](./0000-template.md) to `text/0000-my-rfc.md`, where `my-rfc` is a title relevant to your proposal, and `0000` is the RFC ID. Leave the number as is for now. Once a Pull Request is made, update this ID to match the PR ID.
* Fill in the template. Put care into the details: It is important to present convincing motivation, demonstrate an understanding of the design's impact, and honestly assess the drawbacks and potential alternatives.

### Submitting the RFC
* An RFC is `proposed` by posting it as a PR. Once the PR is created, update the RFC file name to use the PR ID as the RFC ID.
* An RFC is `approved` when four reviewers github-approve the PR. The RFC is then merged.
* If an RFC is `rejected` or `withdrawn`, the PR is closed. Note that these RFCs submissions are still recorded, as Github retains both the discussion and the proposal, even if the branch is later deleted.
* If an RFC discussion becomes long, and the RFC then goes through a major revision, the next version of the RFC can be posted as a new PR, which references the old PR. The old PR is then closed. This makes RFC review easier to follow and participate in.

### Integrating the RFC into the Spec
* Once an RFC is `approved`, an issue is created in the [specification repo](https://github.com/open-telemetry/opentelemetry-specification) to integrate the RFC into the spec.
* When reviewing the spec PR for the RFC, focus on whether the spec is written clearly, and reflects the changes approved in the RFC. Please abstain from relitigating the approved RFC changes at this stage.
* An RFC is `integrated` when four reviewers github-approve the spec PR. The PR is then merged, and the spec is versioned.

### Implementing the RFC
* Once an RFC is `integrated` into the spec, an issue is created in the backlog of every relevant OpenTelemetry implementation.
* PRs are made until the all the requested changes are implemented.
* The status of the OpenTelemetry implementation is updated to reflect that it is implementing a new version of the spec.

## Changes to the RFC process

The hope and expectation is that the RFC process will **evolve** with the OpenTelemetry. The process is by no means fixed.

Have suggestions? Concerns? Questions? **Please** raise an issue or raise the matter on our [community](https://github.com/open-telemetry/community) chat.

## Background on the OpenTelemetry RFC process

Our RFC process borrows from the [Rust RFC](https://github.com/rust-lang/rfcs) and [Kubernetes Enhancement Proposal](https://github.com/kubernetes/enhancements) processes, the former also being [very influential](https://github.com/kubernetes/enhancements/blob/master/keps/0001-kubernetes-enhancement-proposal-process.md#prior-art) on the latter; as well as the [OpenTracing RFC](https://github.com/opentracing/specification/tree/master/rfc) process. Massive kudos and thanks to the respective authors and communities for providing excellent prior art 💖

[circleci-image]: https://circleci.com/gh/open-telemetry/rfcs.svg?style=svg 
[circleci-url]: https://circleci.com/gh/open-telemetry/rfcs
[gitter-image]: https://badges.gitter.im/open-telemetry/opentelemetry-specification.svg 
[gitter-url]: https://gitter.im/open-telemetry/opentelemetry-specification?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge
