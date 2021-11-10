# Per Tracer Configuration

Support setting configuration (Sampler, IdGenerator, Span Limits, etc) on
individual Tracers that overrides the shared configuration provided by the
Tracer Provider.

## Motivation

- Allows Sampler decisions based on Instrumentation Library without adding it as
  an explicit argument to `Sample` calls. This resolves [Issue
  1588](https://github.com/open-telemetry/opentelemetry-specification/issues/1588).
  
- Provides a way to disable Tracing in a library or for any Named Tracer.

This is already mentioned in the specification:

> If configuration must be stored per-tracer (such as disabling a certain
> tracer), the tracer could, for example, do a look-up with its
> name+version+schema_url in a map in the TracerProvider, or the TracerProvider
> could maintain a registry of all returned Tracers and actively update their
> configuration if it changes.

But with no requirement or API related to actually providing this feature.

## Explanation

Tracer instances are created through a Tracer Provider which stores the
configuration used for each Tracer (i.e., SpanProcessors, IdGenerator,
Span Limits and Sampler). Creating or retrieving a Tracer from the Tracer
Provider requires a name (and optional version) that identify the
instrumentation library.

In addition to the default configuration, the Tracer Provider can have separate
configuration settings for individual Tracers based on the name of the Named
Tracer. Adding a configuration does not require the Named Tracer to exist, the
name and/or version, is the only requirement.

For example, a configuration with Sampler of `always_off` can be added under the
name `io.opentelemetry.contrib.mongodb`. Then, when a Tracer is created through
the Tracer Provider with name `io.opentelemetry.contrib.mongodb` it will use the
`always_off` Sampler, but the rest of its configuration is the same as the
default Tracer Provider configuration.

## Internal details

Where possible the configuration should remain shared between Tracer instances
as a reference to avoid allocations.

Configuration should be able to be done both through static configuration and at
the time a Tracer Provider is created.

## Trade-offs and mitigation

- Additional memory usage and allocation.

Before this change each Tracer could hold a reference to a shared configuration
provided by the Tracer Provider. But this will only be the case when a different
configuration is used for a Tracer, all others can remain references.

## Prior art and alternatives

- Multiple Tracer Providers

The specification as of this writing hints at the use of multiple Tracer
Providers in order to allow the creation of Tracers with different
configurations:

> some applications may want to or have to use multiple TracerProvider
> instances, e.g. to have different configuration (like SpanProcessors) for each
> (and consequently for the Tracers obtained from them)

While this proposal does not require removing this capability it does replace
the necessity of using multiple Tracer Providers in order to have different
configurations. Most existing implementations do not offer the
capability of creating multiple Tracer Providers and may essentially prohibit
it. Allowing configuration to be done through the Tracer Provider based on the
Tracer name allows for easier adoption and making additional use of the existing
concept of Named Tracers that fits this well.

I started this proposal with the intention of the solution being through
expanding on the idea of multiple Tracer Providers. Aside from the issues
implementing that as mentioned above, the idea began to turn into Named Tracer
Providers -- in order to support defining what Tracer Provider to use for what
Named Tracer/Instrumentation Library a name for the Provider itself became
needed, or at least useful. This felt clunky and to be adding an unnecessary
abstraction when Named Tracers already provided a method for users to
distinguish Tracers from one another.

Lastly, multiple Tracer Providers makes the configuration of the Exporter
pipeline more complicated. I didn't see a clean way to use the same Batch
Processor and Exporter for multiple Tracer Providers, since the Tracing spec is
1.0 I didn't want to propose breaking Exporting out from the Span Processors.

## Open questions

- Span Processors
  - The API for setting a specific Sampler for a Tracer is simple. But for Span
    Processors the common case will be a user wanting to use different or
    additional Processors except for the one calling the exporter. Because the
    exporting pipeline is tied to the Span Processors it makes it complicated to
    provide a simple API for updating how the user expects -- instead the spec
    to implement this OTEP should likely discourage, or even disallow, Span
    Processor configuration on a per-Tracer basis. This is further explored in
    the following section and a future OTEP.
- Can a Tracer's configuration (whether it is shared or unique) be updated after
  it is retrieved from a Tracer Provider? e.g. should something like `Tracer =
  TracerProvider.getTracer("name").set_sampler(...)` be possible.
  - And should a Tracer not given a name be able to be configured
    `TracerProvider.getTracer().set_sampler(...)`.
- Or should configuration only be predefined by name of a Named Tracer or
  Instrumentation Library?

## Future possibilities

- Configuring the propagators to use for individual Instrumentation Libraries.
  - Not immediately related because they are not part of the Tracer, but having
    configuration be related to the Instrumentation Library as it is in this
    OTEP does at least open up the idea of configuring parts of the SDK based on
    the Instrumentation Library.
- Span Processors V2
  - Investigating this OTEP led to realizations about the tight coupling of Span
    Processors with the Exporter and Tracer Provider. This OTEP does not address
    these concerns but will be built on in a follow up OTEP to provide more
    flexibility to configure how Spans are processed.
