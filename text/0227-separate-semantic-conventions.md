# Separate Semantic Conventions

Move Semantic Conventions outside of the main Specifications and version them
separately.

## Motivation

We need to allow semantic conventions to evolve mostly independent of the
overall OpenTelemetry specification. Today, any breaking change in a semantic
convention would require bumping the version number of the entirety of the
OpenTelemetry specification.

## Explanation

A new github repository called `semantic-conventions` would be created in the
OpenTelemetry organization.

This would *initially* have the following structure:

- Boilerplate files, e.g. `README.md`, `LICENSE`, `CODEOWNERS`, `CONTRIBUTING.md`
- `Makefile` that allows automatic generation of documentation from model.
- `model/` The set of YAML files that exist in
  `{specification}/semantic_conventions` today.
- `docs/` A new directory that contains human readable documentation for how to
  create instrumentation compliant with semantic conventions.
  - `resource/` - Contents of `{specification}/resource/semantic_conventions`
  - `trace/` - Contents of `{specification}/trace/semantic_conventions`
  - `metrics/` - Contents of `{specification}/metrics/semantic_conventions`
  - `logs/`- Contents of `{specification}/logs/semantic_conventions`
  - `schemas/` - A new location for [Telemetry Schemas](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/schemas/README.md)
    to live. This directory will be hosted at
    `https://opentelemetry.io/semantic_conventions/schemas/`

Existing semantic conventions in the specification would be marked as
deprecated, with documentation denoting the move.

Additionally, if the semantic conventions eventually move to domain-specific
director structure (e.g. `docs/{domain}/README.md`, with trace, metrics, events
in the same file), then this can be refactored in the new repository, preserving
git history.

There will also be the following exceptions in the specification:

- Semantic convetions used to implement API/SDK details will be fully specified
  and will not be allowed to change in the Semantic Convention directory.
  - Error/Exception handling will remain in the specification.
  - SDK configuration interaction w/ semantic convention will remain in the
    specification. Specifically `service.name`.
- The specification may elevate some semantic conventions as necessary for
  compatibility requirements, e.g. `service.instance.id` and
  [Prometheus Compatibility](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/compatibility/prometheus_and_openmetrics.md).

These exceptions exist because:

- Stable portions of the specification already rely on these conventions
- These conventions are required to implement an SDK today.

As such, the Specification should define the absolute minimum of reserved or
required attribute names and their interaction with the SDK.

## Internal details

The following process would be used to ensure semantic conventions are
seamlessly moved to their new location.

- A new repository `opentelemetry/semantic_conventions` will be constructed with
  the proposed format and necessary `Makefile` / tooling.
- A moratorium will be placed on Semantic Convention PRs to the specififcation
  repository. (Caveat that PRs related to this proposal would be allowed).
- All semantic conventions in the Specification will be copied to the new
  repository.
- Semantic conventions in the Specification will be marked as deprecated with
  links to the new location.
- The Specification will be updated to require "special" conventions, like
  `service.name` and configuration interaction.
- Instrumentation authors will update their `SchemaURL` to
   `https://opentelemetry.io/semantic_conventions/schemas/{semconv_version}`
   from
   `https://opentelemetry.io/schemas/{spec_version}`

## Trade-offs and mitigations

This proposal has a few drawbacks:

- The semantic conventions will no longer be easily referencable form the specification.
  - This is actually a benefit. We can ensure isolation of convention from
    specification and require the Specification to use firm language for
    attributes it requires, like `service.name`.
  - We will provide x-links from existing location to the new location.
- Semantic Convention version will no longer match the specification version.
  - Instrumentation authors will need to consume a separate semantic-convention
    bundle from Specification bundle. What used to be ONE upgrade effort will
    now be split into two (hopefully smaller) efforts.
  - We expect changes from Semantic Conventions and Specification to be
    orthogonal, so this should not add significant wall-clock time.

Initially this repository would have the following ownership:

- Approvers
  - [Liudmila Molkova](github.com/lmolkova)
- Approvers (HTTP Only)
  - [Trask Stalnaker](github.com/trask)
- Maintainers
  - [Josh Suereth](github.com/jsuereth)
  - [Armin Reuch](github.com/arminru)
  - [Reiley Yang](github.com/reyang)

That is, Maintenance would initially continue to fall on (a subset of) the
Technical committee. Approvers would start targeted at HTTP semantic convention
stability and expand rapidly as we build momentum on semantic conventions.

## Prior art and alternatives

When we evaluate equivalent communities and efforts we see the following:

- `OpenTracing` - had specification and [semantics](https://github.com/opentracing/specification/blob/master/semantic_conventions.md)
  merged.
- `OpenCensus` - had specification and semantics merged. However, OpenCensus
  merged with OpenTelemetry prior to mass adoption or stable release of its
  specification.
- `Elastic Common Schema` - the schema is its own project / document.
- `Prometheus` - Prometheus does not define rigid guidelines for telemetry, like
  semantic conventions, instead relying on naming conventions and
  standardization through mass adoption.

## Open questions

This OTEP doesn't address the full needs of tooling and codegen that will be
needed for the community to shift to a separate semantic convention directory.
This will require each SiG that uses autogenarated semantic conventions to
adapt to the new location.

## Future possibilities

This OTEP paves way for the following desirable features:

- Semantic Conventions can decide to bump major version numbers to accommodate
  new signals or hard-to-resolve new domains without breaking the Specification
  version number.
- Semantic Conventions can have dedicated maintainers and approvers.
- Semantic Conventions can restructure to better enable subject matter experts
  (SMEs) to have approver/CODEOWNER status on relevant directories.

There is a desire to move semantic conventions to domain-specific directories
instead of signal-specific. This can occur after a separation of the repository.

For example:

- `docs/`
  - `signals/` - Conventions for metrics, traces + logs
    - `http/`
    - `db/`
    - `messaging/`
    - `client/`
  - `resource/` - We still need resource-specific semantic conventions
