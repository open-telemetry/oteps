# Instrumentation API

API providing support to library instrumentation authors.

## Motivation

The OpenTelemetry API provides a means for data collection, for example starting / stopping
spans or recording metrics. However, this is only half the story when it comes to instrumenting and
misses out on common patterns required during instrumentation of libraries. For example, server
instrumentation will always need to extract context from incoming headers and use it as the parent
of a created span. There are also ongoing discussions on how to model nested spans. Both of these
are telemetry concepts that do not actually tie directly to any one library, and abstracting it into
an instrumentation API allows library instrumentation to only deal with attribute mapping of library
objects for the most part.

Without defining these
patterns through an instrumentation API, it is difficult to ensure consistency of quality and user
experience of different instrumentation libraries.

The goals of the instrumentation API are

- Provide a good user experience for authors of library instrumentation that minimizes logic and cognitive
load.

- Codify quality requirements of good library instrumentation, for example completeness of semantic
conventions.

- Provide a consistent user experience for users of library instrumentation to ensure similar
configurability and usability regardless of the library being used.

## Explanation

The instrumentation API is a higher level API built on top of the OpenTelemetry API. The key notion
is that instrumentation of libraries almost always involves intercepting a library at two points, the
start and end of an operation, for example an HTTP request or a DB query. This means that the primary
instrumentation API only needs to provide two methods, `start` and `end` which are called from
the instrumentation at the beginning and end of the operation. This API is called the `Instrumenter`.

In addition to the operation lifecycle, the other library-specific aspect of instrumentation is
mapping a library domain object into OpenTelemetry semantic conventions. The instrumentation API
allows instrumentation to implement `Extractor`s, which map from a request or response object into
attributes. Extractors enforce a contract, for example by defining an abstract class, which includes
all the semantic conventions of a given namespace in OpenTelemetry. The library instrumentation just
implements extractors relevant to the library - for example, an HTTP framework will implement the
HTTP extractor. This allows all HTTP attributes to be filled without the potential to miss
attributes.

An `InstrumenterBuilder` stitches together `Instrumenter` and `Extractor`s, as well as consistently
defining a set of well known configuration knobs for library users. The instrumentation library will
initialize a builder with defaults registering the implemented `Extractor`s and return it to the
user of the instrumentation, which can further configure it to their needs.

## Internal details

These details are based on a [prototype](https://github.com/open-telemetry/opentelemetry-java-instrumentation/tree/9f31a057b6878619157aff93e27fdb15fad89958/instrumentation-api/src/main/java/io/opentelemetry/instrumentation/api/instrumenter)
in Java.

### Instrumenter

The `Instrumenter` is the main entry point for the author of library instrumentation. It has two
primary methods.

- `Context start(Context parentContext, REQUEST request)`: The method to invoke at the beginning of an
operation. The parent context, usually the current context, and the library-specific request object
are passed in. The implementation creates a span, invokes all extractors and request listeners,
starts the span, and returns a `Context` with the instrumentation state.

- `void end(Context context, REQUEST request, RESPONSE response, Throwable error)`: The method to
invoke at the end of an operation. The `Context` returned from `start` is passed in, as well as the
library-specific request and response objects and any error if present. The implementation retrieves
a span from the context, invokes all extractors and request listeners, records the error and status,
and ends the span.

Many libraries offer a means of intercepting operations, for example a servlet filter, gRPC
interceptor, etc. In these cases, a library author will simply create an implementation of this
interceptor with an `Instrumenter` and call the methods. For example, in a servlet filter, the code
would simply be something like

```java
class TracingServletFilter implements Filter {
  private final Instrumenter instrumenter;

  TracingServletFilter(Instrumenter instrumenter) {
    this.instrumenter = instrumenter;
  }

  public void doFilter(ServletRequest request, ServletResponse response, FilterChain chain) {
    Context context = instrumenter.start(Context.current(), request);
    Throwable error = null;
    try (Scope ignored = context.makeCurrent()) {
      chain.doFilter(request, response);
    } catch (Throwable t) {
      error = t;
    }
    instrumenter.end(context, request, response, error);
  }
}
```

### AttributesExtractor

`AttribuetsExtractor` allows populating attributes from a library's request and response domain
objects.

The `AttributesExtractor` interface itself is two callbacks that are called from `start` and `end`
of the `Instrumenter`.

- `onStart(AttributesBuilder attributes, REQUEST request)`: Populates attributes with information
from the domain specific request object at the beginning of a request.

- `onEnd(AttributesBuilder attributes, REQUEST request, RESPONSE response)`: Populates attributes
with information from the domain specific request and response object at the end of a request.

The main reason we must pass `request` in `onEnd` as well is because for many frameworks, information
is contained in the request object but not available until the end, for example the `net.peer.ip`
remote address for a client request.

Additional interfaces are provided for `AttributesExtractor` corresponding to each semantic convention
namespace.

For example, `HttpAttributesExtractor` defines an interface with one method per HTTP semantic convention.
The implementation of `onStart` and `onEnd` populate the attributes with the result of these methods,
which operate on the request domain object. In Java it looks like [this](https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/instrumentation-api/src/main/java/io/opentelemetry/instrumentation/api/instrumenter/http/HttpAttributesExtractor.java).

An implementation of each semantic convention method is required, i.e., no default implementation of the
interface methods is provided. This is to ensure it is a conscience decision by the instrumentation library
author to avoid populating a semantic convention because it is not available in that framework. The
expectation is these interfaces will generally lead instrumentation authors into populating all of
OpenTelemetry's semantic conventions.

There are cases when a single framework object does not provide all the information about supported
semantic conventions, several objects are required. For example, in database instrumentation, it is
common to have the SQL statement and connection information only available separately. For these cases,
the instrumentation author can define a wrapper type to hold both, for example `RedisRequestWrapper`
which contains connection information and statement information. The extractors will operate on the
wrapper. This allows extractors to avoid memory allocations in best cases, but still provide a
mechanism when it is not possible.

### `RequestListener`

A `RequestListener` is a simple callback called at the start and end of an operation after resolving
attributes with the registered `AttributesExtractor`s.

- `Context start(Context context, Attributes requestAttributes)`: Called at the start of a request
- `void end(Context context, Attributes responseAttributes, Throwable error)`: Called at the end of a request

One major use case we have found for these listeners is to compute metrics. A `RequestListener` is
implemented to populate HTTP server metrics [here](https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/instrumentation-api/src/main/java/io/opentelemetry/instrumentation/api/instrumenter/http/HttpServerMetrics.java).
An HTTP framework simply registers this with the `Instrumenter` to enable metric collection of
requests.

In addition, users often need to run custom logic within the lifecycle of a request and will
implement their own request listeners.

### Additional knobs

Additional interfaces are provided to map library domain objects to tracing information.

- `SpanNameExtractor`: Maps the request object to a span name. Span names are generally defined by
semantic conventions, and implementations can be provided for them. For example, HTTP frameworks
will use `HttpSpanNameExtractor` which wraps an `HttpAttributesExtractor` to compose the name based
on the HTTP method or path, as [here](https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/instrumentation-api/src/main/java/io/opentelemetry/instrumentation/api/instrumenter/http/HttpSpanNameExtractor.java).

- `SpanKindExtractor`: Maps the request object to a span kind, generally needed for frameworks that
can handle multiple kinds of spans based on the request i.e., `SERVER` or `CONSUMER`.

- `SpanLinkExtractor`: Maps the parent context and request object to links.

- `StartTimeExtractor`: Maps the request object to a start time. Rarely used but sometimes necessary.

- `SpanStatusExtractor`: Maps the request, response, and error objects to a span status. Semantic
conventions often define this, so for example `HttpSpanStatusExtractor` would wrap an `HttpAttributesExtractor`
to determine the status based on the HTTP status code if available, or the error otherwise.

- `EndTimeExtractor`: Maps the response object to an end time. Rarely used but sometimes necessary.

### InstrumenterBuilder

`InstrumenterBuilder` accepts all the implementations of the interfaces for a library instrumentation
and constructs an `Instrumenter`.

Mutators add or override the implementation of one of the above interfaces

- `addAttributesExtractor`
- `setSpanStatusExtractor`
- `addSpanLinkExtractor`
- `setTimeExtractors`
- `addRequestListener`

An instrumentation author will provide an entry point to the instrumentation that preconfigures the
builder based on the semantic conventions that apply to that framework.

```
InstrumenterBuilder newBuilder() {
  HttpAttributesExtractor httpAttributes = new MyLibraryHttpAttributes();
  return new InstrumenterBuilder(INSTRUMENTATION_NAME, HttpSpanNameExtractor.create(httpAttributes))
    .addAttributesExtractor(httpAttributes, new MyLibraryNetAttributes())
    .setSpanStatusExtractor(HttpSpanStatusExtractor.create(httpAttributes))
    .addRequestListener(HttpServerMetrics.create())
}
```

The builder is returned to the library user to allow them to customize as they need. They can register
additional extractors, reconfigure the span name, or add listeners.

Note that library users do not need to use `Instrumenter`, therefore it is recommended this entry
point only provides the knobs for configuring the `InstrumenterBuilder` without returning the
`Instrumenter` itself. Instead, use the configuration to return the library-specific registration
mechanism, for example the gRPC interceptor.

The entry point for gRPC in Java looks like [this](https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/instrumentation/grpc-1.6/library/src/main/java/io/opentelemetry/instrumentation/grpc/v1_6/GrpcTracing.java#L16).

## Trade-offs and mitigations

In certain languages like Java, there may be some performance implications of enforcing semantic
convention contracts through interfaces. Especially if an `AttributeExtractor` must be implemented
for a wrapper type, there could be some additional allocation. Instrumentation authors that need
the absolutely best performance possible may choose to avoid the instrumenter API and use the
OpenTelemetry API directly. The overhead should not be significant in most cases though, and most
authors will likely prefer the significant ease of use. In particular, for instrumentation written
inside a library's codebase itself (not OpenTelemetry's codebase), we can expect authors to not be
as familiar with instrumentation. They will almost always want to trade off a few cycles using our
high level instrumentation API instead of trying to implement it correctly without.

## Prior art and alternatives

Much of the API, especially mapping request/response objects to attributes, and the instrumentation
entry point is inspired by Brave's [instrumentation API](https://github.com/openzipkin/brave/blob/master/instrumentation/http/src/main/java/brave/http/HttpRequestParser.java).
One key difference is that extractors can be written to avoid the wrapper object in many cases.

## Open questions

This design is based on prototyping in Java - most do not seem to be Java specific but it is not clear
how generally applicable this is across languages.

Should this be an enforced specification or just guidelines? One of the goals of this proposal is to
provide a consistent experience when using instrumentation, something that ideally stretches across
languages.

## Future possibilities

A sound instrumentation API will increase the ability for library owners, not OpenTelemetry maintainers,
to write high quality instrumentation.
