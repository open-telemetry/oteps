# Single Formatters for all Contexts

**Status:** `proposed`

Defining a single formatter which (de)serializes all contexts, rather than separate ones.

## Motivation

The existing propagators specification specifies that formatters should exist for each type of context (today this is SpanContext and DistributedContext). Requiring SDK authors to have to author one formatter per context type can result in error-prone and backwards-incompatible consumption and usage.

The error-prone aspect comes from the configuration of said formatters. Let's say that a user is setting up their propagators by hand: in order to do so they will need to configure both their SpanContext propagators and their DistributedContext propagators. If a user wants to use a SaaS product that requires both a custom SpanContext propagator as well as a DistributedContext propagator, they will have to know to attach both. If they forget to do so, they will end up with missing context that is hard to debug (since the user of a SaaS product would probably have little internal knowledge about how propagation should occur).

The backwards-incompatible changes come from the need to extract or inject context into new propagators. If a SaaS vendor or open source protocol previously had no need for on of the two propagators, but started to need it, they would need to notify all consumers that, as of this new version, the consumer must modify their code to add the additional formatter for the new context they require. This will probably come up often in SaaS offerings that previously only needed to propagate data in SpanContext, but then added their own fields that will then require custom DistributedContext formatters.

## Explanation

The new proposed API schema looks like:

    formatters (namespace)
        UnifiedContext.class
        BinaryAPI.class
        HTTPTextFormat.class

With the following APIs:

### UnifiedContext

UnifiedContext will be a composite object containing all contexts that are present in the OpenTelemetry implementation. As of the authoring of this RFC, this object will contain a DistributedContext and SpanContext.

### BinaryAPI

The BinaryAPI will support two methods:

  * fromBytes: convert raw bytes to a UnifiedContext or Null
    * toBytes: convert a UnifiedContext to raw bytes

### HTTPTextFormat

The HTTPTextFormat will support two methods:

    * extract: from a carrier object and getter, return a UnifiedContext
    * inject: from a UnifiedContext, a carrier object, and a setter, return Null. populate the carrier object with values from the UnifiedContext.

## Internal details

This will require medium-sized refactoring on existing propagators and integrations.

* integrations will have to switch to extracting the SpanContext and DistributedContext from the UnifiedContext.
* all existing HTTPTextFormat and BinaryAPI APIs and implementations would have to refactored to match the unified HTTPTextFormat and BinaryAPIs.

An example of the UnifiedContext at work (although not following the code organization above) exists at: 
* PR: https://github.com/open-telemetry/opentelemetry-python/pull/89/files
* file: https://github.com/open-telemetry/opentelemetry-python/blob/f1f8bb884fdcfe4ac257225e1636e1960495a51c/opentelemetry-api/src/opentelemetry/context/unified_context.py


## Trade-offs and mitigations

One drawback is this makes error cases a bit more ambiguous. For example, what happens if a format can extract a span properly, but not a distributed context? or visa versa?

This could be mitigated by defining some clear behavior around these cases. Also the current choice left to formatters is to return a full object, partial object, or none. The same could be provided here and it would be up to the consumer to error handle (although that again is prone to errors and different expectations).

## Open questions

Deeper and partial error handling as descrbied in trade-offs. For example, what if a formatter succeeds in deserializing a DistributedContext, but not a SpanContext?

## Future possibilities

This enables seamless integration of additional context objects, if that becomes a necessity.
