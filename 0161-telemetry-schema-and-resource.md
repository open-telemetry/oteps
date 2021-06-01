# Telemetry Schema and Resource

Specify a mechnism for users of Resource to interact with Resource leveraging
TelemetrySchema version for a stable API.

## Motivation

In the [current defintion](https://github.com/open-telemetry/opentelemetry-specification/pull/1692)
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

### Understanding schema compatiblity

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

TODO: Outline details on resource convertTo.

### Understanding schema compatiblity

Here we propose simply providing an impelmentation of
[SemVer](https://semver.org/) for isCompatible, and using the
the [ordering rules](https://semver.org/#spec-item-11) for isNewer.  This would
pull the version number out of the schema url.

## Trade-offs and mitigations

The major drawback to this proposal is the requirement of making all schema
migration implementations available in all SDKs.  This is mitigated by the
following:

- Initially, we could prevent any `rename` alterations to schema for compatible
  version numbers. `convertToSchema` would check for semver compatibiltiy and
  just return the original resource if compatible.
- Alternatively, we can check the newer schema for any transformations in the
  `all` or `resource` section before failing to convert.

This should allow a slow-rollout of implementation by SDKs by increasing the
number of successful `convertToSchema` operations they allow.

## Prior art and alternatives

TODO: Look these up.

## Open questions

- What kinds of operations on Resource should be considered compatible?
- Is "downgrading" a schema version going to be "safe" for exporters?

## Future possibilities

TODO: outline these.
