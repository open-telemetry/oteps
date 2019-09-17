# Named Tracers

**Status:** `proposed`

_Creating tracers using a factory mechanism and naming those tracers in accordance with the library that provides the instrumentation for a traced component._

## Suggested reading

* [Proposal: Tracer Components](https://github.com/open-telemetry/opentelemetry-specification/issues/10)
* [Global Instance discussions](https://github.com/open-telemetry/opentelemetry-specification/labels/global%20instance)

## Motivation

Instead of having a global tracer singleton, every instrumentation library needs to create its own tracer, which also has a proper name that identifies the library / logical component that uses this tracer. If there are two different instrumentation libraries for the same technology (e.g. MongoDb), these instrumentation libraries should have distinct names. In this sense, this proposal suggests the introduction of a "semantic namespace" concept for tracers and the resulting spans.

The purpose of _Named Tracers_ is the ability to allow for associating an identifier (e.g. _"io.opentelemetry.contrib.mongodb"_) with a tracer. This further allows for logical and semantic coupling of this tracer and a certain technology or library (like _MongoDB_ in this example). Having named tracers enables support for having a dedicated property (component) depending on this name on a tracer that is different for different libraries.

This proposal originates from an `opentelemetry-specification` proposal on [components](https://github.com/open-telemetry/opentelemetry-specification/issues/10) since having a concept of named tracers would automatically enable determining this semantic `component` property.


## Explanation

From a user perspective, working with *Named Tracers* and *TracerFactories* is very similar to how e.g. the [Java logging API](https://docs.oracle.com/javase/7/docs/api/java/util/logging/Logger.html#getLogger(java.lang.String)) and logging frameworks like [log4j](https://www.slf4j.org/apidocs/org/slf4j/LoggerFactory.html) work. In analogy to requesting Logger objects through LoggerFactories, a tracing library would create specific 'Tracer' objects through a 'TracerFactory'.

```java
Tracer tracer = OpenTelemetry.getTracerFactory().getTracer("io.opentelemetry.contrib.mongodb");
```

In a way, the `TracerFactory` replaces the global `Tracer` singleton object as a ubiquitous point to request a tracer instance.


## Internal details

By providing a TracerFactory and *Named Tracers*, a vendor or OpenTelemetry implementation gains more flexibility in providing tracers and which attributes they set in the resulting spans that are produced.

In the simplest case, an OpenTelemetry implementation can return a single instance for a requested tracer, regardless of the name specified as long as the name given for the tracer 

Alternatively, an implementation can provide different tracer instances per specified tracer name, thus being able to associate this tracer with the `component` being traced.

## Prior art and alternatives

Alternatively, instead of having a `TracerFactory`, existing (global) tracers could return additional indirection objects (called e.g. `TraceComponent`), which would be able to produce spans for specifically named traced components.

```java
  TraceComponent traceComponent = OpenTelemetry.Tracing.getTracer().componentBuilder("io.opentelemetry.contrib.mongodb");
  Span span = traceComponent.spanBuilder("someMethod").startSpan();
```

Overall, this would not change a lot compared to the `TracerFactory` since the levels of indirection until producing an actual span are the same.

Instead of setting the `component` property based on the given tracer names, those names could also be used as *prefixes* for produced span names (e.g. `<TracerName-SpanName>`). However, with regard to data quality and semantic conventions, a dedicated `component` set on spans is probably preferred.

## Future possibilities

By adapting this proposal, current implementations that do not make use of the specified tracer name and provide a single global tracer, would not require much change. However they could change that behavior in future versions and provide more specific tracer implementations then. On the other side, if the mechanism of *Named Tracers* is not a part of the initial specification, such scenarios will be prevented and hard to retrofit in future version, should they be deemed necessary then. 

## Examples (of Tracer names)

Since tracer names describe the libraries which use the tracers, those names should be defined in a way that makes them as unique as possible. The name of the tracer should represent the identity of the library, class or package that provides the instrumentation. 

Examples (based on existing contribution libraries from OpenTracing and OpenCensus):

* `io.opentracing.contrib.spring.rabbitmq`
* _io.opentracing.contrib.jdbc_
* _io.opentracing.thrift_
* _io.opentracing.contrib.asynchttpclient_
* _io.opencensus.contrib.http.servlet_
* _io.opencensus.contrib.spring.sleuth.v1x_
* _io.opencensus.contrib.http.jaxrs_
* _github.com/opentracing-contrib/go-amqp_ (Go)
* _github.com/opentracing-contrib/go-grpc_ (Go)
* _OpenTracing.Contrib.NetCore.AspNetCore_ (.NET)
* _OpenTracing.Contrib.NetCore.EntityFrameworkCore_ (.NET)

