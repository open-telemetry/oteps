# Telemetry Schema and Resource

Specify a mechnism for users of Resource to interact with Resource leveraging
TelemetrySchema version for a stable API.

## Motivation

In the [current definition](https://github.com/open-telemetry/opentelemetry-specification/pull/1692)
of Resource + Telemetry schema interaction, the specification focuses mostly on
how to provide a consistent "storage" of Resource within the SDK. However,
from a usage standpoint, how do consumers gain stability across changes to
schema?

- Telemetry Schema allows the notion of "renaming" attributes in a schema.
  Can a consumer of Resource "lock" in a specific version of the schema to
  ensure no renaming occurs?
- If resource detection is provided by libraries external to open-telemetry, do
  we require ALL resource detectors to be against the same telemetry schema
  version number?  This would inhibit upgrading versions as you'd need to wait
  for all resource detectors to upgrade, even if a (stable) conversion exists
  from the resource version.

Specifically, we need a way for exporters that rely on Resource labels to ensure
various version differences across OpenTelemetry components do not cause
breakages by leveraging the powers provided by telemetry schema.

As an example, prior to schema version and stability, the PR
[#1495](https://github.com/open-telemetry/opentelemetry-specification/pull/1495)
renamed `cloud.zone` to `cloud.availability_zone`.  In a Schema world, this
change could have been defined with the following schema rule:

```yaml
  next.version:
    resources:
      changes:
        - rename_attributes:
            attribute_map:
              cloud.zone: cloud.availability_zone
  previous.version:
```

If an exporter is defined expecting to find `cloud.zone` on `previous_version`,
then when the SDK is bumped to `next.version` it will no longer be able to see
`cloud.zone`.

Specifically, we assume the following components could have independent
versioning + release cycles:

```
+-------------------+     +-----+     +----------+
| Resource Detector | <-> | SDK | <-> | Exporter |
+-------------------+     +-----+     +----------+
```

## Explanation

We propose a new set of SDK features:

- A mechanism to "migrate" a resource up/down compatible schema changes.
- A mechanism to determine if two schema URLs allow compatible changes.

### Consuming resources at specific schema version

The first is done by providing a new method against `Resource` in the SDK that
will transform the attributes as necessary to match a given schema version.

e.g. In Java you might have the following for resource:

```java
public abstract class Resource {
    ... existing methods ...
    /** The current schema version. */
    public String getSchemaUrl();
    /**
     * Returns a new instance of Resource with labels compatible with the
     * given schema url.
     *
     * <p> Returns an invalid Resource if conversion is not possible.
     */
    public Resource convertToSchema(String schemaUrl);
}
```

This would allow a consumer of resources to ensure a stable view of resource
labels, e.g.

```java
public static String MY_SCHEMA_VERSION="https://opentelemetry.io/schemas/1.1.0";
public void myMetricMethod(MetricData metric) {
    useStableResource(metric.getResource().convertToSchema(MY_SCHEMA_VERSION));
}
```

### Understanding schema compatibility

The second will be done via a new (public) SDK method that allows understanding
whether or not two schemas are compatible, and which is the "newer" version.

e.g. in Java you would have:

```java
public abstract class Schemas {
    public static boolean isCompatible(String oldUrl, String newUrl);
    public static boolean isNewer(String testUrl, String baseUrl);
}
```

This could be used in the `Resource.merge` method to determine if resources
can be converted to a compatible schema.  Specfically resource can be merged
IFF:

- The schemas are compatible
- Converting all `Resource` instances to the "newest" schema version returns
  valid `Resource` objects.

## Internal details

There are three major changes to SDKs for this proposal.

### Understanding schema compatibility

Here we propose simply providing an implementation of
[SemVer](https://semver.org/) for isCompatible, and using the
the [ordering rules](https://semver.org/#spec-item-11) for isNewer.  This would
pull the version number out of the schema url.

### Resource merge logic

Resource merge logic is updated as follows (new mechanisms in bold):

The resulting resource will have the Schema URL calculated as follows:

- If the old resource's Schema URL is empty then the resulting
  resource's Schema URL will be set to the Schema URL of the updating resource
- Else if the updating resource's Schema URL is empty then the resulting
  resource's Schema URL will be set to the Schema URL of the old resource,
- Else if the Schema URLs of the old and updating resources are the same then
  that will be the Schema URL of the resulting resource,
- **Else if the Schema urls of the old and updating resources are compatible**
  **then the older resource will be converted to newer schema url and merged.**
- Else this is a merging error (this is the case when the Schema URL of the old
  and updating resources are not empty and are not compatible).

### Resource conversion logic

SDKs will be expected to provide a mechanism to apply transformations
(both backwards and forwards) listed in `changes` for any compatible schemas.

**Note: Until conversion logic is available in SDKs, any change to telemetry**
**schema that relies on `changes` will be considered a breaking, i.e.**
**they will lead to a major version bump.**

## Trade-offs and mitigations

The major drawback to this proposal is the requirement of making all schema
migration implementations available in all SDKs.  This is mitigated by the
following:

- Initially, we could prevent any `rename` alterations to schema for compatible
  version numbers. `convertToSchema` would check for semver compatibility and
  just return the original resource if compatible.
- Alternatively, we can check the newer schema for any transformations in the
  `all` or `resource` section before failing to convert.

This should allow a slow-rollout of implementation by SDKs by increasing the
number of successful `convertToSchema` operations they allow.

## Prior art and alternatives

It's a common practice in protocols to negotiate version semantics and for
newer code to provide backwards-compatible "views" of older APIs for
compatibility.   This is intended as a stop-gap to prevent immediate breakages
as the ecosystem of components evolves to the new protocol on independent
timelines.  Specifically "core" components can update without forcing all
downstream to components to also update.

### Alternatives

One alternative we can take here is to force all exporters to always be on the
same version of telemetry schema as the SDK and find a mechanism to warn/prevent
older exporters from using a newer SDK.

Another alternative is for SDKs to continue propagating resources as-is and
have exporters issue errors when resource versions don't align with
expectations.

## Open questions

- What kinds of operations on Resource should be considered compatible?
- Is "downgrading" a schema version going to be "safe" for exporters?

## Future possibilities

The notion of telemetry-schema conversion could be expanded and adapted into
its own set of functionality that SDKs or SDK-extensions (like resource
detectors and exporters) can make use of.
