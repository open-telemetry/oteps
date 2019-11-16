# Proposal: Separate Layer for Context Propagation

* [OpenTelemetry layered architecture](#openTelemetry-layered-architecture)
  * [Aspects](#Aspects)
    * [Observability API](#Observability-API)
    * [Baggage API](#Baggage-API)
  * [Context Propagation](#Context-Propagation)
    * [Context API](#Context-API)
    * [Propagation API](#Propagation-API)
* [Prototypes](#Prototypes)
* [Examples](#Examples)
* [Internal details](#Internal-details)
* [FAQ](#faq)

Refactor OpenTelemetry into a set of separate cross-cutting concerns which 
operate on a shared context propagation mechanism.

## Motivation

This RFC addresses the following topics:

**Separatation of concerns**  
* Remove the Tracer dependency from context propagation mechanisms.
* Handle user data (Baggage) and observability data (SpanContext, Correlations) 
  separately.  

**Extensibility**
* Allow developers to create new applications for context propagation. For 
  example: A/B testing, encrypted or authenticated data, and new, experimental 
  forms of observability.

## Explanation

# OpenTelemetry layered architecture

The design of OpenTelemetry is based on the priciples of aspect-oriented 
programming, adopted to the needs of distributed systems.

OpenTelemetry is separated into two layers: **aspects** which crosscut and 
intertwine with a program's functional concerns, and cannot be encapsulated. In 
this architecture, each cross-cutting concern is modeled as an independent 
subsystem. Multiple aspects - including the tracing and baggage systems provided 
by OpenTelemetry - then share the same underlying **context propagation** 
system, which allows these cross-cutting concerns to store and access their 
data across the lifespan of a distributed transaction.

# Aspects

## Observability API
Distributed tracing is one example of an aspect. Tracing code is interleaved 
with regular code, and ties together indepentent code modules which would 
otherwise remain encapsulated. Tracing is also distributed, and requires 
non-local, transaction-level context propagation in order to execute correctly.

A second observability aspect is correlations. Correlations are labels for 
metrics, where the value of the label may be defined anywhere in the 
transaction. Correlations are like Baggage, but are write only - the values can 
only be used for observability.

The observability APIs are not described here directly. However, in this new 
design, all observability APIs would be modified to make use of the generalized 
context propagation mechanism described below, rather than the tracing-specific 
propagation system it uses today.

## Baggage API

In addition to observability, OpenTelemetry provides a simple mechanism for 
propagating arbitrary user data, called Baggage. This mechanism is not related
to tracing or observability, but it uses the same context propagation layer. 

Baggage may be used to model new aspects which would benefit from the same 
transaction-level context as tracing, e.g., for identity, versioning, and 
network switching. 

To manage the state of these cross-cutting concerns, the Baggage API provides a 
set of functions which read, write, and propagate data.

**SetBaggage(context, key, value) -> context**  
To record the distributed state of an aspect, the Baggage API provides a 
function which takes a context, a key, and a value as input, and returns an 
updated context which contains the new value.

**GetBaggage(context, key) -> value**  
To access the distributed state of an aspect, the Baggage API provides a 
function which takes a context and a key as input, and returns a value.

**RemoveBaggage(context, key) -> context**  
To delete distributed state from an aspect, the Baggage API provides a function 
which takes a context, a key, and a value as input, and returns an updated 
context which contains the new value.

**ClearBaggage(context) -> context**  
To avoid sending baggage to an untrusted downstream process, the Baggage API 
provides a function remove all baggage from a context.

**GetBaggageHTTPPropagator() -> (HTTPExtract, HTTPInject)**  
To deserialize the state of the system sent from the the prior upstream process, 
and to serialize the the current state of the system and send it to the next 
downstream process, the Baggage API provides a function which returns a 
baggage-specific implementation of the HTTPExtract and HTTPInject functions.

# Context Propagation

## Context API

Aspects access data in-process using a shared context object. Each aspect uses 
its own namespaced set of keys in the context, containing all of the data for 
that cross-cutting concern.

**SetValue(context, key, value) -> context**  
To record the local state of an aspect, the Context API provides a function 
which takes a context, a key, and a value as input, and returns an updated 
context which contains the new value.

**GetValue(context, key) -> value**  
To access the local state of an aspect, the Context API provides a function 
which takes a context and a key as input, and returns a value.

### Optional: Automated Context Management
When possible, the OpenTelemetry context should automatically be associated 
with the program execution context. Note that some languages do not provide any 
facility for setting and getting a current context. In these cases, the user is 
responsible for managing the current context.  

**SetCurrent(context)**  
To associate a context with program execution, the Context API provides a 
function which takes a Context.

**GetCurrent() -> context**  
To access the context associated with program execution, the Context API 
provides a function which takes no arguments and returns a Context.


## Propagation API

Aspects send their state to downstream processes via propagators: 
functions which read and write context into RPC requests. Each aspect creates a 
set of propagators for every type of supported medium - currently only HTTP 
requests.

**HTTPExtract(context, request) -> context**  
To receive data injected by prior upstream processes, the Propagation API 
provides a function which takes a context and an HTTP request, and returns 
context which represents the state of the upstream system.

**HTTPInject(context, request)**  
To send the data for all aspects downstream to the next process, the 
Propagation API provides a function which takes a context and an HTTP request, 
and mutates the HTTP request to include an HTTP Header representation of the 
context.

**ChainHTTPInjector(injector, injector) -> injector**  
To allow multiple aspects to inject their context into the same request, the 
Propagation API provides a function which takes two injectors, and returns a 
single injector which calls the two original injectors in order.

**ChainHTTPExtractor(extractor, extractor) -> extractor**  
To allow multiple aspects to extract their context from the same request, the 
Propagation API provides a function which takes two extractors, and returns a 
single extractor which calls the two original extractors in order.

### Optional: Global Propagators
It is often convenient to create a chain of propagators during program 
initialization, and then access these combined propagators later in the program. 
To facilitate this, global injectors and extractors are optionally available. 
However, there is no requirement to use this feature.

**SetHTTPExtractor(extractor)**  
To update the global extractor, the Propagation API provides a function which 
takes an extractor.

**GetHTTPExtractor() -> extractor**  
To access the global extractor, the Propagation API provides a function which 
returns an extractor.

**SetHTTPInjector(injector)**  
To update the global injector, the Propagation API provides a function which 
takes an injector.

**GetHTTPInjector() -> injector**  
To access the global injector, the Propagation API provides a function which 
returns an injector.

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

In , we would like `service A` to decide on which backend service to call, 
based on the client version. We would also like to trace the entire system, in 
order to understand if requests to `service C` are slower or faster than 
`service B`. What might `service A` look like?

First, during program initialization, `service A` might set a global 
extractor and injector which chains together baggage and tracing 
propagation. Let's assume this tracing system is configured to use B3, 
and has a specific propagator for that format. Initializating the propagators 
might look like this:

```php
func InitializeOpentelemetry() {
  // create the propagators for tracing and baggage.
  bagExtract, bagInject = Baggage::HTTPPropagator()
  traceExtract, traceInject = Tracer::B3Propagator()
  
  // chain the propagators together and make them globally available.
  extract = Propagation::ChainHTTPExtractor(bagExtract,traceExtract)
  inject = Propagation::ChainHTTPInjector(bagInject,traceInject)

  Propagation::SetHTTPExtractor(extract)
  Propagation::SetHTTPInjector(inject)
}
```

These propagators can then be used in the request handler for `service A`. The 
tracing and baggage aspects use the context object to handle state without 
breaking the encapsulation of the functions they are embedded in.

```php
func HandleUpstreamRequest(context, request, project) -> (context) {
  // Extract the span context. Because the extractors have been chained,
  // both a span context and any upstream baggage have been extracted 
  // from the request headers into the returned context.
  extract = Propagation::GetHTTPExtractor()
  context = extract(context, request.Headers)

  // Start a span, setting the parent to the span context received from 
  // the upstream. The new span will then be in the returned context.
  context = Tracer::StartSpan(context, [span options])
  
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
  inject = Propagation::GetHTTPInjector()
  context = inject(context, request.Headers)

  // make an http request
  data = request.Do()

  return data
}
```

In this version of pseudocode above, we assume that the context object is 
explict,and is pass and returned from every function as an ordinary parameter. 
This is cumbersome, and in many languages, a mechanism exists which allows 
context to be propagated automatically.

In this version of pseudocode, assume that the current context can be stored as 
a thread local, and is implicitly passed to and returned from every function.

```php
func HandleUpstreamRequest(request, project) {
  extract = Propagation::GetHTTPExtractor()
  extract(request.Headers)
  
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
  
  inject = Propagation::GetHTTPInjector()
  inject(request.Headers)
  
  data = request.Do()

  return data
}
```

Digging into the details of the tracing system, what might some of the details 
look like? Here is a crude example of extracting and injecting B3 headers, 
using an explicit context.

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

  func B3Injector(context, headers) -> (context) {
    headers["X-B3-TraceId"] = Context::GetValue( context, "trace.parentTraceID")
    headers["X-B3-SpanId"] = Context::GetValue( context, "trace.parentSpanID")

    return context
  }
```

Now, have a look at a crude example of how StartSpan might then make use of the 
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
    this key will not be propagated downstream.
    Context::SetValue( "trace.currentSpanData", spanData)

    return
  }
```

Let's look at a couple other scenarios related to automatic context propagation.

When are the values in the current contexdt available? Scope managemenent may be different in each langauge, but as long as the scope does not change (by switching threads, for example) the 
current context follows the execuption of the program. This includes after a 
function returns. Note that the context objects themselves are immutable, so 
explict handles to prior contexts will not be updated when the current context 
is changed.

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

If context propagation is automantic, does the user ever need to reference a 
context object directly? Sometimes. Ehen automated context propagation is 
available, there is no restriction that aspects must only ever access the 
current context. 

For example, if an aspect wanted to merge the data beween two contexts, at 
least one of them will not be the current context.

```php
mergedContext = MergeBaggage( Context::GetCurrent(), otherContext)
Context::SetCurrent(mergedContext)
```

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

**Correlation Context -** Correlation Context contains a map of labels and 
values, to be shared between metrics and traces. This allows observability data 
to be indexed and dimensionalized in a variety of ways. Note that correlations 
can quickly add overhead when propagated in-band. But because this data is 
write-only, it may be possible to optimize how it is transmitted.

**Baggage -** Transaction-level application data, meant to be shared with 
downstream components. This data is readable, and must be propagated in-band. 
Because of this, Baggage should be used sparingly, to avoid ballooning the size 
of all downstream requests.

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

## Why separate Baggage from Correlations?

Since Baggage Context and Correlation Context appear very similar, why have two? 

First and foremost, the intended uses for Baggage and Correlations are 
completely different. Secondly, the propagation requirements diverge 
significantly.

Correlation values are solely to be used as labels for metrics and traces. By 
making Correlation data write-only, how and when it is transmitted remains 
undefined. This leaves the door open to optimizations, such as propagating the 
correlation data out-of-band.

Baggage values, on the other hand, are explicitly added in order to be accessed 
by downstream by other application code. Therefore, Baggage Context must be 
readable, and reliably propagated in-band in order to accomplish this goal.

There may be cases where a key-value pair is propagated as a Correlation for 
observability and as a Baggage item for application-specific use. AB testing is 
one example of such use case. This would result in extra overhead, as the same 
key-value pair would be present in two separate headers.  

Solving this edge case is not worth having the semantic confusion of a single 
implementation with a dual purpose.

## What about complex propagation behavior?

Some OpenTelemetry proposals have called for more complex propagation behavior. 
For example, falling back to extracting B3 headers if W3C Trace-Context headers 
are not found. Chained propagators and other complex behavior can be modeled as 
implementation details behind the Propagator interface. Therefore, the 
propagation system itself does not need to provide chained propagators or other 
additional facilities.


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
