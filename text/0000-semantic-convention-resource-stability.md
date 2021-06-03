# Resource Semantic Convention Stabiltiy Requirements

Outline what changes are "safe" to perform for resource semantic conventions
and which require major version bumps.

## Motivation

This outlines acceptable changes to semantic conventions for resources, allowing
growth and evolution over time.

While Telemetry Schemas propose a mechanism to define changes to semantic
conventions in a way that users can evolve telemetry, this document first
attempts to define what is considered a compatible change.

## Explanation

We take the assumption that `Resource`, as it exists today, is the identity of
an SDK and is leveraged as such in a lot of assumptions throughout open
telemetry. Specifically, if an SDK that used to produce a `Resource` of
attributes XYZ instead produces ABC, these metrics would be considered from a
different source. To that end, if a Metric produced on version x.y of
OpenTelemetry SDK is incompatible with one produced on version x.{y+1} because
of `Resource` changes, this would be considered breaking.

This means the following:

- `Resource` detectors should provide a comprehensive set of attributes in
  all or nothing fashion. if a resource detector produced attributes XYZ for
  a given compute, it should continue to do so for compatible versions.
- Resource attributes (e.g. `cloud.provider`) may have new possible values
  defined as long as there is zero overlap with how the previous values were
  computed.  For example, `acme-cloud` would never be inferred where previously
  `aws` was inferred.
  *Note: if `aws` was inferred incorrectly on "Acme cloud" previously, that would not be a violation.*
- Adding new Resource attribute semantic conventions is allowed, however
  existing resource detectors MUST NOT use them, only new detectors would be
  allowed.

If we assume capabilities of
[Telemetry Schema conversion](https://github.com/open-telemetry/oteps/pull/161), 
then we can relax some restrictions as follows:

- 1:1 Renaming of attribute names or labels is allowed.
- New attributes are allowed to be generated in a resource detector.

## Internal details

- Resource semantic conventions will need to be defined at the
  `Resource Detector` level to ensure stability of identity of resources.
- Attribute names and existing values cannot be renamed within stable versions
  (pending OTEP #161)

### New Semantic Conventions

Going forward, Resource semantic conventions will be outlined at a "Resource
detection" level.  While users can merge together resource detectors, each
detector will provide a stable set of Resource attributes and labels.

For Example, [k8s](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/semantic_conventions/k8s.md)
resources would be changed, such that for Node:

#### Node

A resource detector providing K8s node information MUST produce the following.

**type:** `k8s.node`

**Description:** A Kubernetes Node.

<!-- semconv k8s.node -->
| Attribute  | Type | Description  | Examples  | Required |
|---|---|---|---|---|
| `k8s.node.name` | string | The name of the Node. | `node-1` | *No* |
| `k8s.node.uid` | string | The UID of the Node. | `1eb3a0c6-0477-4080-a9cb-0cb7db65c6a2` | *YES* |
<!-- endsemconv -->

Here, the `k8s.node.uid` is marked as required for this resource detector.

## Trade-offs and mitigations

What are some (known!) drawbacks? What are some ways that they might be mitigated?

Note that mitigations do not need to be complete *solutions*, and that they do not need to be accomplished directly through your proposal. A suggested mitigation may even warrant its own OTEP!

## Prior art and alternatives

Currently, the [specification for stability states](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/versioning-and-stability.md#not-defined-semantic-conventions-stability):

```
Semantic Conventions SHOULD NOT be removed once they are added.
New conventions MAY be added to replace usage of older conventions, but the older conventions SHOULD NOT be removed.
Older conventions SHOULD be marked as deprecated when they are replaced by newer conventions.
```

This is both more restrictive and less restrictive than the current implementation. It has the consequence of potentially breaking metric/resource identity from version bumps in OpenTelemetry.

One alternative for flexibility in resource labelling is to allow resources to
contain "Identifying" and "Descriptive" attributes, where only identifying
attributees are subjected to rigorous stability limitations.

## Open questions

- Should we be more flexible in Resource identity?
  - Should Resource attributes be outlined as "identifying" and "descriptive"?
  - Should we consolidate resource identity at the detector level?

## Future possibilities

As we improve the Telemetry Schema design, we may be able to improve what is
considreed a compatible change for Resources.
