# Context Propagation: A Layered Approach 

* [Motivation](#Motivation)
* [OpenTelemetry layered architecture](#OpenTelemetry-layered-architecture)
  * [Cross-Cutting Concerns](#Cross-Cutting-Concerns)
    * [Observability API](#Observability-API)
    * [Baggage API](#Baggage-API)
  * [Context Propagation](#Context-Propagation)
    * [Context API](#Context-API)
    * [Propagation API](#Propagation-API)
* [Prototypes](#Prototypes)
* [Examples](#Examples)
  * [Global initialization](#Global-initialization)
  * [Extracting and injecting from HTTP headers](#Extracting-and-injecting-from-HTTP-headers)
  * [Simplfy the API with automated context propagation](#Simplfy-the-API-with-automated-context-propagation)
  * [Implementing a propagator](#Implementing-a-propagator)
  * [Implementing a concern](#Implementing-a-concern)
  * [The scope of current context](#The-scope-of-current-context)
  * [Referencing multiple contexts](#Referencing-multiple-contexts)
  * [Falling back to explicit contexts](#Falling-back-to-explicit-contexts)
* [Internal details](#Internal-details)
* [FAQ](#faq)

Refactor OpenTelemetry into a set of separate cross-cutting concerns which 
operate on a shared context propagation mechanism.

# Motivation

This RFC addresses the following topics:

**Separatation of concerns**  
* Remove the Tracer dependency from context propagation mechanisms.
* Handle user data (Baggage) and observability data (SpanContext, etc) 
  separately.  

**Extensibility**
* Allow developers to create new applications for context propagation. For 
  example: A/B testing, encrypted or authenticated data, and new, experimental 
  forms of observability.

## Explanation


# OpenTelemetry layered architecture

The design of OpenTelemetry is based on the priciples of [aspect-oriented 
programming](https://en.wikipedia.org/wiki/Aspect-oriented_programming), 
adopted to the needs of distributed systems.

OpenTelemetry is separated into two layers. The top layer contains a set of 
independent **cross-cutting concerns**, which intertwine with a program's 
application logic, and cannot be cleanly encapsulated. All concerns share an 
underlying distributed **context propagation** layer, for storing state and 
accessing data across the lifespan of a distributed transaction.


# Cross-Cutting Concerns

## Observability API
Distributed tracing is one example of a cross-cutting concern. Tracing code is 
interleaved with regular code, and ties together independent code modules which 
would otherwise remain encapsulated. Tracing is also distributed, and requires 
non-local, transaction-level context propagation in order to execute correctly.

The various observability APIs are not described here directly. However, in this new 
design, all observability APIs would be modified to make use of the generalized 
context propagation mechanism described below, rather than the tracing-specific 
propagation system it uses today.

## Baggage API

In addition to observability, OpenTelemetry provides a simple mechanism for 
propagating arbitrary user data, called Baggage. This mechanism is not related
to tracing or observability, but it uses the same context propagation layer. 

Baggage may be used to model new concerns which would benefit from the same 
transaction-level context as tracing, e.g., for identity, versioning, and 
network switching. 

To manage the state of these cross-cutting concerns, the Baggage API provides a 
set of functions which read, write, and propagate data.

**`GetBaggage(context, key) -> value`**  
To access the distributed state of a concern, the Baggage API provides a 
function which takes a context and a key as input, and returns a value.

**`SetBaggage(context, key, value) -> context`**  
To record the distributed state of a concern, the Baggage API provides a 
function which takes a context, a key, and a value as input, and returns an 
updated context which contains the new value.

**`RemoveBaggage(context, key) -> context`**  
To delete distributed state from a concern, the Baggage API provides a function 
which takes a context, a key, and a value as input, and returns an updated 
context which contains the new value.

**`ClearBaggage(context) -> context`**  
To avoid sending baggage to an untrusted process, the Baggage API provides a 
function to remove all baggage from a context.

**`GetBaggagePropagator() -> (HTTP_Extractor, HTTP_Injector)`**  
To deserialize the state of the system sent from the the prior process, and to 
serialize the the current state of the system and send it to the next process, 
the Baggage API provides a function which returns a baggage-specific 
implementation of the HTTPExtract and HTTPInject functions.

# Context Propagation

## Context API

Cross-cutting concerns access data in-process using the same, shared context 
object. Each concern uses its own namespaced set of keys in the context, 
containing all of the data for that cross-cutting concern.

**`GetValue(context, key) -> value`**  
To access the local state of an concern, the Context API provides a function 
which takes a context and a key as input, and returns a value.

**`SetValue(context, key, value) -> context`**  
To record the local state of a cross-cutting concern, the Context API provides a 
function which takes a context, a key, and a value as input, and returns an 
updated context which contains the new value.


### Optional: Automated Context Management
When possible, the OpenTelemetry context should automatically be associated 
with the program execution context. Note that some languages do not provide any 
facility for setting and getting a current context. In these cases, the user is 
responsible for managing the current context.  

**`GetCurrent() -> context`**  
To access the context associated with program execution, the Context API 
provides a function which takes no arguments and returns a Context.

**`SetCurrent(context)`**  
To associate a context with program execution, the Context API provides a 
function which takes a Context.



## Propagation API

Cross-cutting concerns send their state to the next process via propagators: 
functions which read and write context into RPC requests. Each concern creates a 
set of propagators for every type of supported medium - currently only HTTP 
requests.

**`Extract(context, []http_extractor, headers) -> context`**  
In order to continue transmitting data injected earlier in the transaction, 
the Propagation API provides a function which takes a context, a set of 
HTTP_Injectors, and a set of HTTP headers, and returns a new context which 
includes the state sent from the prior process.

**`Inject(context, []http_injector, headers) -> headers`**  
To send the data for all concerns to the next process in the transaction, the 
Propagation API provides a function which takes a context and a set of 
HTTP_Extractors, and adds the contents of the context in to HTTP headers to 
include an HTTP Header representation of the context.

**`HTTP_Extractor(context, headers) -> context`**  
Each concern must implement an HTTP_Extractor, which can locate the headers 
containing the http-formatted data, and then translate the contents into an 
in-memory representation, set within the returned context object. 

**`HTTP_Injector(context, headers) -> headers`**  
Each concern must implement an HTTP_Injector, which can take the in-memory 
representation of its data from the given context object, and add it to an 
existing set of HTTP headers.

### Optional: Global Propagators
It may be convenient to create a list of propagators during program 
initialization, and then access these propagators later in the program. 
To facilitate this, global injectors and extractors are optionally available. 
However, there is no requirement to use this feature.

**`GetExtractors() -> []http_extractor`**  
To access the global extractor, the Propagation API provides a function which 
returns an extractor.

**`SetExtractors([]http_extractor)`**  
To update the global extractor, the Propagation API provides a function which 
takes an extractor.

**`GetInjectors() -> []http_injector`**  
To access the global injector, the Propagation API provides a function which 
returns an injector.

**`SetInjectors([]http_injector)`**  
To update the global injector, the Propagation API provides a function which 
takes an injector.

# Prototypes

**Erlang:** https://github.com/open-telemetry/opentelemetry-erlang-api/pull/4  
**Go:** https://github.com/open-telemetry/opentelemetry-go/pull/297  
**Java:** https://github.com/open-telemetry/opentelemetry-java/pull/655  
**Python:** https://github.com/open-telemetry/opentelemetry-python/pull/278  
**Ruby:** https://github.com/open-telemetry/opentelemetry-ruby/pull/147  

# Examples

It might be helpful to look at an example, written in pseudocode. Let's describe 
a simple scenario, where `service A` responds to an HTTP request from a `client` 
with the result of a request to `service B`.

```
client -> service A -> service B
```

Now, let's assume the `client` in the above system is version 1.0. With version 
v2.0 of the `client`, `service A` must call `service C` instead of `service B` 
in order to return the correct data.

```
client -> service A -> service C
```

In this example, we would like `service A` to decide on which backend service 
to call, based on the client version. We would also like to trace the entire 
system, in order to understand if requests to `service C` are slower or faster 
than `service B`. What might `service A` look like?

## Global initialization 
First, during program initialization, `service A` configures baggage and tracing 
propagation, and include them in the global list of injectors and extractors. 
Let's assume this tracing system is configured to use B3, and has a specific 
propagator for that format. Initializating the propagators might look like this:

```php
func InitializeOpentelemetry() {
  // create the propagators for tracing and baggage.
  bagExtract, bagInject = Baggage::HTTPPropagator()
  traceExtract, traceInject = Tracer::B3Propagator()
  
  // add the propagators to the global list.
  Propagation::SetExtractors(bagExtract, traceExtract)
  Propagation::SetInjectors(bagInject, traceInject)
}
```

## Extracting and injecting from HTTP headers
These propagators can then be used in the request handler for `service A`. The 
tracing and baggage concerns use the context object to handle state without 
breaking the encapsulation of the functions they are embedded in.

```php
func ServeRequest(context, request, project) -> (context) {
  // Extract the context from the HTTP headers. Because the list of 
  // extractors includes a trace extractor and a baggage extractor, the 
  // contents for both systems are included in the  request headers into the 
  // returned context.
  extractors = Propagation::GetExtractors()
  context = Propagation::Extract(context, extractors, request.Headers)

  // Start a span, setting the parent to the span context received from 
  // the client process. The new span will then be in the returned context.
  context = Tracer::StartSpan(context, [span options])
  
  // Determine the version of the client, in order to handle the data 
  // migration and allow new clients access to a data source that older 
  // clients are unaware of.
  version = Baggage::GetBaggage( context, "client-version")

  switch( version ){
    case "v1.0":
      data, context = FetchDataFromServiceB(context)
    case "v2.0":
      data, context = FetchDataFromServiceC(context)
  }

  context = request.Response(context, data)

  // End the current span
  context = Tracer::EndSpan(context)

  return context
}

func FetchDataFromServiceB(context) -> (context, data) {
  request = NewRequest([request options])
  
  // Inject the contexts to be propagated. Note that there is no direct 
  // reference to tracing or baggage.
  injectors = Propagation::GetInjectors()
  context = Propagation::Inject(context, injectors, request.Headers)

  // make an http request
  data = request.Do()

  return data
}
```

## Simplify the API with automated context propagation
In this version of pseudocode above, we assume that the context object is 
explict,and is pass and returned from every function as an ordinary parameter. 
This is cumbersome, and in many languages, a mechanism exists which allows 
context to be propagated automatically.

In this version of pseudocode, assume that the current context can be stored as 
a thread local, and is implicitly passed to and returned from every function.

```php
func ServeRequest(request, project) {
  extractors = Propagation::GetExtractors()
  Propagation::Extract(extractors, request.Headers)
  
  Tracer::StartSpan([span options])
  
  version = Baggage::GetBaggage("client-version")
  
  switch( version ){
    case "v1.0":
      data = FetchDataFromServiceB()
    case "v2.0":
      data = FetchDataFromServiceC()
  }

  request.Response(data)
  
  Tracer::EndSpan()
}

func FetchDataFromServiceB() -> (data) {
  request = newRequest([request options])
  
  injectors = Propagation::GetInjectors()
  Propagation::Inject(request.Headers)
  
  data = request.Do()

  return data
}
```

## Implementing a propagator
Digging into the details of the tracing system, what might the internals of a 
span context propagator look like? Here is a crude example of extracting and 
injecting B3 headers, using an explicit context.

```php
  func B3Extractor(context, headers) -> (context) {
    context = Context::SetValue( context, 
                                 "trace.parentTraceID", 
                                 headers["X-B3-TraceId"])
    context = Context::SetValue( context,
                                "trace.parentSpanID", 
                                 headers["X-B3-SpanId"])
    return context
  }

  func B3Injector(context, headers) -> (headers) {
    headers["X-B3-TraceId"] = Context::GetValue( context, "trace.parentTraceID")
    headers["X-B3-SpanId"] = Context::GetValue( context, "trace.parentSpanID")

    return headers
  }
```

## Implementing a concern
Now, have a look at a crude example of how StartSpan might make use of the 
context. Note that this code must know the internal details about the context 
keys in which the propagators above store their data. For this pseudocode, let's 
assume again that the context is passed implicitly in a thread local.

```php
  func StartSpan(options) {
    spanData = newSpanData()
    
    spanData.parentTraceID = Context::GetValue( "trace.parentTraceID")
    spanData.parentSpanID = Context::GetValue( "trace.parentSpanID")
    
    spanData.traceID = newTraceID()
    spanData.spanID = newSpanID()
    
    Context::SetValue( "trace.parentTraceID", spanData.traceID)
    Context::SetValue( "trace.parentSpanID", spanData.spanID)
    
    // store the spanData object as well, for in-process propagation. Note that 
    // this key will not be propagated, it is for local use only.
    Context::SetValue( "trace.currentSpanData", spanData)

    return
  }
```

## The scope of current context
Let's look at a couple other scenarios related to automatic context propagation.

When are the values in the current context available? Scope managemenent may be 
different in each langauge, but as long as the scope does not change (by 
switching threads, for example) the current context follows the execuption of 
the program. This includes after a function returns. Note that the context 
objects themselves are immutable, so explict handles to prior contexts will not 
be updated when the current context is changed.

```php
func Request() {
  emptyContext = Context::GetCurrent()
  
  Context::SetValue( "say-something", "foo") 
  secondContext = Context::GetCurrent()
  
  print(Context::GetValue("say-something")) // prints "foo"
  
  DoWork()
  
  thirdContext = Context::GetCurrent()
  
  print(Context::GetValue("say-something")) // prints "bar"

  print( emptyContext.GetValue("say-something") )  // prints ""
  print( secondContext.GetValue("say-something") ) // prints "foo"
  print( thirdContext.GetValue("say-something") )  // prints "bar"
}

func DoWork(){
  Context::SetValue( "say-something", "bar") 
}
```

## Referencing multiple contexts
If context propagation is automantic, does the user ever need to reference a 
context object directly? Sometimes. Even when automated context propagation is 
an available option, there is no restriction which says that concerns must only 
ever access the current context. 

For example, if a concern wanted to merge the data beween two contexts, at 
least one of them will not be the current context.

```php
mergedContext = MergeBaggage( Context::GetCurrent(), otherContext)
Context::SetCurrent(mergedContext)
```

## Falling back to explicit contexts
Sometimes, suppling an additional version of a function which uses explict 
contexts is necessary, in order to handle edge cases. For example, in some cases 
an extracted context is not intended to be set as current context. An 
alternate extract method can be added to the API to handle this.

```php
// Most of the time, the extract function operates on the current context.
Extract(headers)

// When a context needs to be extracted without changing the current 
// context, fall back to the explicit API.
otherContext = ExtractWithContext(Context::GetCurrent(), headers)
```


# Internal details

![drawing](img/context_propagation_details.png)

## Context details
OpenTelemetry currently intends to implement three context types of context 
propagation.

**Span Context -** The serializable portion of a span, which is injected and 
extracted. The readable attributes are defined to match those found in the 
[W3C Trace Context specification](https://www.w3.org/TR/trace-context/). 

**Baggage -** Transaction-level application data, meant to be shared with 
all services. This data is readable, and must be propagated in-band. 
Because of this, Baggage should be used sparingly, to avoid ballooning the size 
of every request.

Note that OpenTelemetry APIs calls should *always* be given access to the entire 
context object, and never just a subset of the context, such as the value in a 
single key. This allows the SDK to make improvements and leverage additional 
data that may be available, without changes to all of the call sites.


## Context Management and in-process propagation

In order for Context to function, it must always remain bound to the execution 
of code it represents. By default, this means that the programmer must pass a 
Context down the call stack as a function parameter. However, many languages 
provide automated context management facilities, such as thread locals. 
OpenTelemetry should leverage these facilities when available, in order to 
provide automatic context management.

## Pre-existing Context implementations

In some languages, a single, widely used Context implementation exists. In other 
languages, there many be too many  implementations, or none at all. For example, 
Go has a the `context.Context` object, and widespread conventions for how to 
pass it down the call stack.

In the cases where an extremely clear, pre-existing option is not available, 
OpenTelemetry should provide its own Context implementation.

## Default Propagators

When available, OpenTelemetry defaults to propagating via HTTP header 
definitions which have been standardized by the W3C.


# FAQ

## What about complex propagation behavior?

Some OpenTelemetry proposals have called for more complex propagation behavior. 
For example, falling back to extracting B3 headers if W3C Trace-Context headers 
are not found. "Fallback propagators" and other complex behavior can be modeled as 
implementation details behind the Propagator interface. Therefore, the 
propagation system itself does not need to provide an mechanism for chaining 
together propagators or other additional facilities.


## Did you add a context parameter to every API call because Go has infected your brain?

No. The concept of an explicit context is fundamental to a model where 
independent cross-cutting concerns share the same context propagation layer. 
How this context appears or is expressed is language specific, but it must be 
present in some form.

# Prior art and alternatives

Prior art:  
* OpenTelemetry distributed context
* OpenCensus propagators
* OpenTracing spans
* gRPC context

# Open questions

Related work on HTTP propagators has not been completed yet.

* [W3C Trace-Context](https://www.w3.org/TR/trace-context/) candidate is not yet 
  accepted.
* Work on [W3C Correlation-Context](https://w3c.github.io/correlation-context/) 
  has begun, but was halted to focus on Trace-Context. 
* No work has begun on a theoretical W3C Baggage-Context.

Given that we must ship with working propagators, and the W3C specifications are 
not yet complete, how should we move forwards with implementing context 
propagation?

# Future possibilities

Cleanly splitting OpenTelemetry into Apects and Context Propagation layer may 
allow us to move the Context Propagation layer into its own, stand-alone 
project. This may facilitate adoption, by allowing us to share Context 
Propagation with gRPC and other projects.
