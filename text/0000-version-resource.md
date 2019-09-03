# Version Resource Type

**Status:** `proposed`

Add a composable 'Version' resource type.

## Motivation

When creating trace data or metrics, it can be extremely useful to know the specific version that
emitted the iota of span or measurement being viewed. However, versions can mean different things
to different systems and users, leading to a situation where a semantic `Version` type that can
encapsulate these identifiers would be useful for analysis systems in order to allow for slicing
data by service, component, or other version.

## Explanation

A `Version` is a semantic resource that can be composed with other resources, such as a `Service`, 
`Component`, `Library`, `Device`, `Platform`, etc. A `Version` is optional, but reccomended.
The definition of a `Version` is as follows:

| Key     | Value  | Description                                  |
|---------|--------|----------------------------------------------|
| semver  | string | version string in semver format              |
| git_sha | string | version string as a git sha                  |

Only one field should be specified in a single `Version`. 

## Internal details

The exact implementation of this resource would vary based on language, but it would ultimately need to be represented in the data format so that it could be commmunicated to analysis backends. A JSON representation of a version resource follows:

```
{
    "version": {
        "type": "semver",
        "value": "1.0.0",
    }
}
```

## Trade-offs and mitigations

The largest drawback to this proposal is that there is a wide variety of things that constitute a 'version string'. By design, we do not attempt to solve for all of them in this proposal - instead, we focus on versions strings that are well-understood to have some semantic meaning attached to them if properly used.

## Prior art and alternatives

Tagging service resources with their version is generally suggested by analysis tools -- see [JAEGER_TAGS](https://www.jaegertracing.io/docs/1.8/client-features/) for an example -- but lacks standardization.

## Open questions

What should the exact representation of the version object be?

Would it make more sense to just have a `version` key that any arbitrary string can be applied to?
