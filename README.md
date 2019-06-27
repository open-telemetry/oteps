# OpenTelemetry RFCs
## Evolving OpenTelemetry at the speed of Markdown

The OpenTelemetry "RFC" (request for comments) process aims to track the **rationale** behind *new functionality*, including trade-offs and alternatives that have been considered.

Much like how designers may improve user research feedback cycles by using prototypes instead of production-ready implementations, we hope that RFCs will help the OpenTelemetry consider functionality more quickly and thoroughly than, say, jumping directly into the standard pull request workflow.

Bug fixes and changes with no external impact, such as refactors, can skip the RFC process and instead stick with the standard GitHub tooling :)

### Table of Contents

* [Opening](#opentelemetry-rfcs)
* [What changes require an RFC?](#what-changes-require-an-rfc)
  * [Extrapolating cross-cutting changes](#extrapolating-cross-cutting-changes)
* [RFC scope](#rfc-scope)
* [The RFC life cycle](#the-rfc-life-cycle)
* [Submitting a new RFC](#submitting-a-new-rfc)
* [Reviewing an RFC](#reviewing-an-rfc)
* [Implementing an RFC](#implementing-an-rfc)
* [RFCs vs GitHub issues](#rfcs-vs-github-issues)
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

### The RFC life cycle

An RFC may go through the following states:

* `proposed`: The starting state. The RFC is still pending approval and should be actively refined and discussed
* `approved`: The RFC has been approved and is ready to be implemented
* `implemented`: The RFC has been completed
* `deferred`: The RFC's approval has been deferred until late, e.g., because significant open questions cannot be answered yet
* `rejected`: The approvers and/or authors (hopefully together!) have decided against implementing the RFC
* `withdrawn`: The authors have withdrawn the RFC. Either the original authors or other community members are still welcome to re-open it in the future
* `replaced`: The RFC has been superseded by a new one

If the rationale behind a change of state cannot be easily inferred - e.g., if and when it becomes `deferred` - the reason should be clearly documented in the relevant commit.

### Submitting a new RFC

1. If you haven't done so yet, [fork](https://help.github.com/en/articles/fork-a-repo) the [repo](https://github.com/open-telemetry/rfcs)
1. Copy [`0000-template.md`](./0000-template.md) to `text/0000-my-rfc.md`, where `my-rfc` is a title relevant to your proposal. Leave the number as is for now.
1. Fill in the template. Do not omit things you think may be obvious to readers; it's better to be too explicit than to leave ambiguity! In particular, assume that no one else cares yet about the proposal, and that your are writing this to **convince** them that they should :)
1. Submit a [pull request](https://github.com/open-telemetry/compare). Assuming that your RFC follows the template and there are no significant omissions, the pull request will be accepted and your RFC will be assigned a number
    * TODO: Should this be automated?
1. A new "work-in-progress" (WIP) pull request will be created with that updates the RFC's status to `approved`
    * TODO: This should probably be automated
1. In the WIP pull request, your proposed RFC will now be up for feedback from the OpenTelemetry community. Others may suggest changes, raise concerns, or add new pull requests building on top of yours
1. Aim to address the discussion! While you are the champion of the RFC, the community collectively owns the result. We're all in this together :)
    * Please make changes as new commits, documenting the reasoning for each in the commit message. Try to avoid squashing, rebasing, force-pushing, etc. at this point, as doing so may mess with the historical discussion on GitHub
1. Once the value and trade-offs of the change have been sufficiently discussed, a non-author member of the relevant SIG will remove "WIP" from the PR's title and propose a "motion for final comment period" (FCP)
    * By this point, no **new** points should have been raised in the discussion for **at least three business days**
    * If substantial new points (e.g., a significant trade-off) are raised *during* the FCP, the FCP should be cancelled and discussion renewed
1. During the FCP, SIG members should each vote to `approve`, `reject`, or `defer` the RFC
    * If the discussion has been contentious, members should explain the rationale behind their vote
1. The FCP will be considered closed and the RFC either approved, rejected, or approved after the following conditions have all been met:
    * At least two business days have passed since the FCP began
    * Either:
      * A majority of non-author members have voted
      * At least three non-author members have voted, and the vote is unanimous
1. The RFC will be merged with the appropriate state

### Reviewing an RFC

Discussion should primarily take place in the pull request aimed at approving the RFC. Any discussion that takes place outside of GitHub - e.g., in meetings or community chat - should be summarized and added to the pull request.

### Implementing an RFC

Once an RFC has been approved, a corresponding issue should be created and prioritized accordingly in the relevant repository.

The original author is welcome but not obligated to implement it. If you are interested in implementing an RFC, please mention it in either the RFC pull request or the new issue (depending on which is active by then). While authors who are interested in implementation generally get priority, we encourage anyone who's interested to collaborate together!

If neither an author nor another community member asks to implement the RFC, it will be assigned once it gets high enough in the backlog.

### RFCs vs GitHub issues

The RFC process primarily aims to document, and facilitate discussion of, the **motivation** and **known drawbacks** of a suggested change. While issues (and pull requests) can definitely accomplish some of this, experience from prior projects has demonstrated that they quickly become too unwieldy, e.g., because the search tools available for issues are much more limited than what's available for a repository's code (or "code", in this case!).

### Changes to the RFC process

The hope and expectation is that the RFC process will **evolve** with the OpenTelemetry. The process is by no means fixed.

Have suggestions? Concerns? Questions? **Please** raise an issue or raise the matter on our [community](https://github.com/open-telemetry/community) chat.

### Background on the OpenTelemetry RFC process

Our RFC process borrows (very!) heavily from the [Rust RFC](https://github.com/rust-lang/rfcs) and [Kubernetes Enhancement Proposal](https://github.com/kubernetes/enhancements) processes, the former also being [very influential](https://github.com/kubernetes/enhancements/blob/master/keps/0001-kubernetes-enhancement-proposal-process.md#prior-art) on the latter; as well as the [OpenTracing RFC](https://github.com/opentracing/specification/tree/master/rfc) process. Massive kudos and thanks to the respective authors and communities for providing excellent prior art 💖
