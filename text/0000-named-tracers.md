# Named Tracers

**Status:** `proposed`

_Creating Tracers using a factory mechanism and naming those Tracers in accordance with the library that provides the instrumentation for a traced component._

## Suggested reading

* [Proposal: Tracer Components](https://github.com/open-telemetry/opentelemetry-specification/issues/10)
* [Global Instance discussions](https://github.com/open-telemetry/opentelemetry-specification/labels/global%20instance)
* [Proposal: Add a version resource](https://github.com/open-telemetry/oteps/pull/38)

## Motivation

The mechanism of "Named Tracers" proposed here is motivated by following scenarios:

* For a consumer of OpenTelemetry instrumentation libraries, there is currently no possibility of influencing the amount of the data produced by such libraries. Instrumentation libraries can easily "spam" backend systems, deliver bogus data or - in the worst case - crash or slow down applications. These problems might even occur suddenly in production environments caused by external factors such as increasing load or unexpected input data.

* If a library hasn't implemented [semantic conventions](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/data-semantic-conventions.md) correctly or those conventions change over time, it's currently hard to interpret and sanitize these data selectively. The produced Spans cannot be associated with those instrumentation libraries later.

This proposal attempts to solve the stated problems by introducing the concept of:
 * _Named Tracers_ identified via a name (e.g. _"io.opentelemetry.contrib.mongodb"_) and a version (e.g._"semver:1.0.0"_) which is associated with the Tracer and the Spans it produces.
 * A `TracerFactory` as the only means of creating a Tracer.

Based on such an identifier, a Sampler could be implemented that discards Spans from certain libraries. Also, by providing a custom Exporter, Span data could be sanitized before it gets processed in a back-end system. However, this is beyond the scope of this proposal, which only provides the fundamental mechanisms.

## Explanation

From a user perspective, working with *Named Tracers* and *TracerFactories* is conceptually similar to how e.g. the [Java logging API](https://docs.oracle.com/javase/7/docs/api/java/util/logging/Logger.html#getLogger(java.lang.String)) and logging frameworks like [log4j](https://www.slf4j.org/apidocs/org/slf4j/LoggerFactory.html) work. In analogy to requesting Logger objects through LoggerFactories, a tracing library would create specific 'Tracer' objects through a 'TracerFactory'.

New Tracers can be created by providing the name and version of an instrumentation library. The version (following the convention proposed in https://github.com/open-telemetry/oteps/pull/38) is basically optional but *should* be supplied since only this information enables following scenarios:
* Only a specific range of versions of a given instrumentation library need to be suppressed, while other versions are allowed (e.g. due to a bug in those specific versions).
* Go modules allow multiple versions of the same middleware in a single build so those need to be determined at runtime.

Instead of using plain strings as an argument for creating new Tracers a Resource identifying an instrumentation library is used. Such resources must have a _version_ and _name_ labels (there could be semantic convention definitions for those labels). Version values will follow the conventions proposed [here](https://github.com/open-telemetry/oteps/pull/38).

```java
// Create a tracer for given instrumentation library in a specific version.
Tracer tracer = OpenTelemetry.getTracerFactory().getTracer("io.opentelemetry.contrib.mongodb", "semver:1.0.0");
```

This `TracerFactory` replaces the global `Tracer` singleton object as a ubiquitous point to request a Tracer instance.

If no tracer name (null or empty string) is specified, following the suggestions in ["error handling proposal"](https://github.com/open-telemetry/opentelemetry-specification/pull/153), a "smart default" will be applied and a default tracer implementation is returned.


## Internal details

By providing a `TracerFactory` and *Named Tracers*, a vendor or OpenTelemetry implementation gains more flexibility in providing Tracers and which attributes they set in the resulting spans that are produced.

The SpanData class is extended with a `getLibraryResource` function that returns the resource associated with the Tracer that created the span.

If there are two different instrumentation libraries for the same technology (e.g. MongoDb), these instrumentation libraries should have distinct names.

## Prior art and alternatives

This proposal originates from an `opentelemetry-specification` proposal on [components](https://github.com/open-telemetry/opentelemetry-specification/issues/10) since having a concept of named Tracers would automatically enable determining this semantic `component` property.

Alternatively, instead of having a `TracerFactory`, existing (global) Tracers could return additional indirection objects (called e.g. `TraceComponent`), which would be able to produce spans for specifically named traced components.

```java
  TraceComponent traceComponent = OpenTelemetry.Tracing.getTracer().componentBuilder(libraryResource);
  Span span = traceComponent.spanBuilder("someMethod").startSpan();
```

Overall, this would not change a lot compared to the `TracerFactory` since the levels of indirection until producing an actual span are the same.

Instead of setting the `component` property based on the given Tracer names, those names could also be used as *prefixes* for produced span names (e.g. `<TracerName-SpanName>`). However, with regard to data quality and semantic conventions, a dedicated `component` set on spans is probably preferred.

Instead of using plain strings as an argument for creating new Tracers, a `Resource` identifying an instrumentation library could be used. Such resources must have a _version_ and a _name_ label (there could be semantic convention definitions for those labels). This implementation alternative mainly depends on the availability of the `Resource` data type on an API level (see https://github.com/open-telemetry/opentelemetry-specification/pull/254).

```java
// Create resource for given instrumentation library information (name + version)
Map<String, String> libraryLabels = new HashMap<>();
libraryLabels.put("name", "io.opentelemetry.contrib.mongodb");
libraryLabels.put("version", "1.0.0");
Resource libraryResource = Resource.create(libraryLabels);
// Create tracer for given instrumentation library.
Tracer tracer = OpenTelemetry.getTracerFactory().getTracer(libraryResource);
```

## Future possibilities

Based on the Resource information identifying a Tracer these could be configured (enabled / disabled) programmatically or via external configuration sources (e.g. environment).

Based on this proposal, other "signal producers" (i.e. metrics and logs) can use the same or a similar creation approach.

## Examples (of Tracer names)

Since Tracer names describe the libraries which use the Tracers, those names should be defined in a way that makes them as unique as possible. The name of the Tracer should represent the identity of the library, class or package that provides the instrumentation. 

Examples (based on existing contribution libraries from OpenTracing and OpenCensus):

* io.opentracing.contrib.spring.rabbitmq
* io.opentracing.contrib.jdbc
* io.opentracing.thrift
* io.opentracing.contrib.asynchttpclient
* io.opencensus.contrib.http.servlet
* io.opencensus.contrib.spring.sleuth.v1x
* io.opencesus.contrib.http.jaxrs
* github.com/opentracing-contrib/go-amqp (Go)
* github.com/opentracing-contrib/go-grpc (Go)
* OpenTracing.Contrib.NetCore.AspNetCore (.NET)
* OpenTracing.Contrib.NetCore.EntityFrameworkCore (.NET)

