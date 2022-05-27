# Introducing Events API


In this OTEP we introduce an Events API that is based on the OpenTelemetry Log signal. In OpenTelemetry's perspective Log Records and Events are different names for the same concept - however, there is a subtle difference in how they are represented using the underlying data model that is described below. Although every language has APIs for logs, they are not all capable of creating events. We will describe why the existing Logging APIs are not sufficient for the purpose of creating events.  It will then be evident that we will need an API in OpenTelementry for creating events. 

We have an option of adding API for both Logs and Events. However, there is a general consensus that we should not have an API in Otel for creating logs since each language already has multiple logging frameworks. Therefore we restrict the API specification below to Events and call it Events API. For logs, it is recommended that end-users continue to use existing Logging APIs and export the logs into OTLP using the  [appender API](https://github.com/open-telemetry/opentelemetry-java-instrumentation/tree/main/instrumentation-appender-api-internal/src/main/java/io/opentelemetry/instrumentation/api/appender/internal) and LogEmitter SDK. The Events API will offer a subset of the features of LogEmitter SDK and so it will be backed by LogEmitter SDK and the LogRecord data model.

## Subtle differences between Logs and Events
In OpenTelemetry's perspective Log Records and Events are different names for the same concept. However, there are subtle differences in how they are represented in the underlying LogRecord data model. Logs have a mandatory severity level as a first-class parameter that events do not have, and events have a mandatory name that logs do not have. Further, logs typically have messages in string form and events have data in the form of key-value pairs. It is due to this that their API interface requirements are slightly different.

## Who requires Events API

Here are a few situations that require recording of Events, there will be more.
* RUM events (Client-side instrumentation)
* Recording kubernetes events
* Recording eBPF events
* Few other event systems described in [example mappings](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#appendix-a-example-mappings) in the data model.

## Can the current Log API interfaces be used for events?

1. The log level is fundamental to the Log APIs in almost all the languages; all the methods in the Log interface are named after the log level, and there is usually no generic method to submit a log entry without log level.
* In JavaScript for Web, the standard method of logging is to use console.log. Events can be created using [Event/CustomEvent](https://developer.mozilla.org/en-US/docs/Web/Events/Creating_and_triggering_events) interfaces.However, there is no option to define custom destination for these logs and events. Logs go only to console and event listeners are attached to the DOM element that dispatches it.
* In Android, android.util.Log has methods  Log.v(), Log.d(), Log.i(), Log.w(), and Log.e() to write logs. These methods correspond to the severity level.
* Swift on iOS has Logger interface that has methods corresponding to severity level too.
2. The Log APIs do not have a standard way to pass event attributes. 
* It may be possible to use the interpolation string args as the parameter to pass event attributes. However, the logging spec seems to map the human readable message (which is obtained after replacing the args in the interpolated string)  to the Body field of LogRecord. 
* Log4j has an EventLogger interface that can be used to create structured messages with arbitrary key-value pairs, but log4j is not commonly used in Android apps as it is not officially supported on Android as per this [Stack Overflow thread](https://stackoverflow.com/questions/60398799/disable-log4j-jmx-on-android/60407849#60407849) by one of log4j’s maintainers.
* In python, logging.LogRecord’s extra field is mapped to Otel LogRecord’s attributes but this field is a hidden field and not part of the public interface.
3. The Log APIs have a message parameter which could map to the Body field of LogRecord. However, this is restricted to String messages and does not allow for a complex structure.

For the above reasons we can conclude that we will need a separate API for generating Events API

## Should OpenTelemetry have an API for logs?

There’s a general consensus in the Otel community that we should not have a full-fledged logging API unless there is a language that doesn't already have a plethora of logging libraries & APIs to choose from where it might make sense to define one. Further, we will not be able to have the [rich set of configuration options](https://logging.apache.org/log4j/2.x/manual/configuration.html) that some popular logging frameworks provide so the logging API in Otel will only become yet another API.

## Events API Interface


For reference, a prototype of the Events API in Java is [here](https://github.com/scheler/opentelemetry-java/pull/1/files)


Client-side telemetry is one of the initial clients that will use the Events API and so the API will be made available in JavaScript, Java and Swift first to be able to use in the SDKs for Browser, Android and iOS.  It may also be added in Go since there is a kubernetes events receiver implemented in Collector based on Logs data-model.


The Events API consist of these main classes:

* EventEmitterProvider is the entry point of the API. It provides access to EventEmitters..
* EventEmitter  is the class responsible for creating events using Log records.

### EventEmitterProvider
EventEmitter can be accessed with an EventEmitterProvider.

In implementations of the API, the EventEmitterProvider is expected to be the stateful object that holds any configuration.

Normally, the EventEmitterProvider is expected to be accessed from a central place. Thus, the API SHOULD provide a way to set/register and access a global default EventEmitterProvider.

Notwithstanding any global EventEmitterProvider, some applications may want to or have to use multiple EventEmitterProvider instances, e.g. to have different configuration (like LogRecordProcessors) for each (and consequently for the EventEmitters obtained from them), or because it's easier with dependency injection frameworks. Thus, implementations of EventEmitterProvider SHOULD allow creating an arbitrary number of EventEmitter instances.

#### EventEmitterProvider operations
The EventEmitterProvider MUST provide the following functions:

* Get an EventEmitter

##### Get an EventEmitter
This API MUST accept the following parameters:
* name (required): This name SHOULD uniquely identify the instrumentation scope, such as the instrumentation library (e.g. io.opentelemetry.contrib.mongodb), package, module or class name.  If an application or library has built-in OpenTelemetry instrumentation, both Instrumented library and Instrumentation library may refer to the same library. In that scenario, the name denotes a module name or component name within that library or application. In case an invalid name (null or empty string) is specified, a working EventEmitter implementation MUST be returned as a fallback rather than returning null or throwing an exception, its name property SHOULD be set to an empty string, and a message reporting that the specified value is invalid SHOULD be logged. A library implementing the OpenTelemetry API may also ignore this name and return a default instance for all calls, if it does not support "named" functionality (e.g. an implementation which is not even observability-related). A EventEmitterProvider could also return a no-op EventEmitter here if application owners configure the SDK to suppress telemetry produced by this library.
* version (optional): Specifies the version of the instrumentation scope if the scope has a version (e.g. a library version). Example value: 1.0.0.
* schema_url (optional): Specifies the Schema URL that should be recorded in the emitted telemetry
* event_domain (optional): Specifies the domain for the events created, which should be added in the attribute `event.domain` in the instrumentation scope.
* pass_context (optional): Specifies whether the Trace Context should automatically be passed on to the events created by the EventEmitter. This SHOULD be false by default.

It is unspecified whether or under which conditions the same or different EventEmitter instances are returned from this function.

Implementations MUST NOT require users to repeatedly obtain an EventEmitter again with the same name+version+schema_url+event_domain+pass_context to pick up configuration changes. This can be achieved either by allowing to work with an outdated configuration or by ensuring that new configuration applies also to previously returned EventEmitters.

Note: This could, for example, be implemented by storing any mutable configuration in the EventEmitterProvider and having EventEmitter implementation objects have a reference to the EventEmitterProvider from which they were obtained. If configuration must be stored per-EventEmitter (such as disabling a certain EventEmitter), the EventEmitter could, for example, do a look-up with its name+version+schema_url in a map in the EventEmitterProvider, or the EventEmitterProvider could maintain a registry of all returned EventEmitters and actively update their configuration if it changes.

The effect of associating a Schema URL with a EventEmitter MUST be that the telemetry emitted using the EventEmitter will be associated with the Schema URL, provided that the emitted data format is capable of representing such association.

### EventEmitter
The EventEmitter is responsible for creating Events.

Note that EventEmitters should usually not be responsible for configuration. This should be the responsibility of the EventEmitterProvider instead.

#### EventEmitter operations
The EventEmitter MUST provide functions to:

* Function named “logEvent” to create an Event with the provided event name and attributes. The event name provided should be inserted as an attribute with key “event.name”. It will override any attribute with the same key in the attributes passed.
* Function named “recordException” to record an Exception as an Event. This is to facilitate recording an exception outside of a trace context for languages that already do not support recording an exception in a log message. This should work similar to [Record Exception](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/api.md#record-exception) in Trace API, with the following semantics
  * This is a specialized variant of logEvent function, so for anything not specified here, the same requirements as for logEvent apply.
  * The signature of the method is to be determined by each language and can be overloaded as appropriate. The method MUST record an exception as an Event with the conventions outlined in the exception semantic conventions document. The minimum required argument SHOULD be no more than only an exception object.
  * If RecordException is provided, the method MUST accept an optional parameter to provide any additional event attributes (this SHOULD be done in the same way as for the AddEvent method). If attributes with the same name would be generated by the method already, the additional attributes take precedence.
  * Note: RecordException may be seen as a variant of logEvent with additional exception-specific parameters and all other parameters being optional (because they have defaults from the [exception semantic convention](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/exceptions.md)).

A LogRecord representing exception event will look like this:
```
{
  time_unix_nano: 1653682410,
  Attributes: {
   event.name: exception,
   exception.type:  OSError,
   exception.message: Division by Zero,
   exception.stacktrace: “”,
   exception.escaped: false
 }
}
```


The EventEmitter SHOULD additionally provide the following functions for convenience:
* Create a new Event using Log Record data model.


### Adding Trace Context to Events
The Event being created may not always be related to the Span in progress even though it’s created in the same execution context. 

In languages where Context is implicitly available (for eg., in Java), the API SHOULD not automatically inject the Context into the Event created. 

In languages where Context is implicitly available (for eg., in Java), new Events MUST NOT get the Context injected into them by default. However, this functionality MAY BE offered separately.

In languages where the Context must be provided explicitly, the end user must capture the Context and set it explicitly in the LogRecord.


### Usage

```java
EventEmitterProvider eventEmitterProvider = SdkLogEmitterProvider.builder()
              .addLogProcessor(BatchLogProcessor.builder(OtlpGrpcLogExporter.builder().
             build()).build()).build();

OpenTelemetry openTelemetry = OpenTelemetrySdk.builder()
    .setLogEmitterProvider(eventEmitterProvider)
    .buildAndRegisterGlobal();

EventEmitter eventEmitter = EventEmitterProvider.getEventEmitter("instrumentation-library-name", "1.0.0");

// Using the builder interface
eventEmitter.eventBuilder("network-changed").build().setAttribute("type", "wifi").emit();
eventEmitter.eventBuilder("page-navigated").build().setAttribute("url", "http://foo.com/bar#new").emit();

// Using the convenience functions
eventEmitter.logEvent(name, attributes);
```


### Usage in Client-side telemetry
For client-side instrumentation, we will choose to not use the fields Severity Number and Severity Field, and only use Attributes instead. Severity is not commonly needed in RUM events, and if needed in future we could use an attribute for it.

```java
public void addBrowserEvent(String name, Attributes attributes) {
   eventEmitter.logEvent("browser." + name, attributes).emit();
}

public void addMobileEvent(String name, Attributes attributes) {
   eventEmitter.logEvent("mobile.” + name, attributes).emit();     
}
```

### Usage in eBPF
From the eBPF [demo](https://youtu.be/F1VTRqEC8Ng?t=233), it looks like they have chosen to put the Event data in the Body field of LogRecord instead of Attributes.  They might have to make the changes to move the event data to attributes to conform to this API specification.

## Semantic Convention for event attribute

| Attribute  | Type | Description  | Examples  | Required |
|---|---|---|---|---|
| `event.name` | string | Name or type of the event. | `network-change`; `button-click`; `exception` | Yes |
| `event.domain` | string | Domain or scope for the event. | `profiling`; `browser` | No |

There are a few possible approaches - 

1. One single attribute name: _otel.event.name_
  * Pros
    * No possibility of conflict with user provided attributes since the prefix is reserved for Otel use.
    * Serves as one single attribute for event names across all domains/verticals.
  * Cons
    *  The otel namespace is typically used to express OpenTelemetry concepts not available in other formats. Event name is not an OpenTelemetry specific concept
    * The value of this attribute MUST use a namespace to avoid confusion with using the same event name across different domains/verticals and that is hard to enforce in the API.
2. Namespaced with the domain/vertical: _browser.event.name_, _mobile.event.name_, _db.event.name_, _k8s.event.name_
  * Cons
    * Requires longer API functions for recording events; instead of logEvent(eventName, attributes), we will need logEvent(eventNameattributeKey, eventNameAttributeValue, eventAttributes)
    * Condition to check for distinguishing against logs is complex now; We should at least mandate the suffix “.event.name” for the attribute key but this is not typical in OpenTelemetry and will be hard to enforce in the API.

3. Make the domain namespace a separate parameter in the logEvent function. The function signature will be: ```void logEvent(eventName, domainNamespace, eventAttributes)```. The two parameters will both go in their own attributes separately - _otel.event.name_, _otel.event.domain_
  * Pros
    * Addresses the concerns in the two approaches above.
  * Cons
    * We will be introducing domain as a first-class concept in OpenTelemetry, we will have to analyze the implications in other areas.
    * LogRecord’s setters can still be used to create events not following these conventions for event name and domain.


# Causality on Events

For creating causality between events we can create wrapper spans that are part of the same trace. However, note that the events themselves are represented using Logs and not as Span Events.

```java
Span s1 = Trace.startSpan()
    addBrowserEvent(e1name, attributes)
    Span s2 = Trace.startSpan()
        addBrowserEvent(e2name, attributes)
    s2.end()
s1.end()
```

# Open questions
2. How do we design the API to prevent automatic injection of (Span) Context in Events when not needed?
  * In Android, the event listener handlers are called in the same thread they are setup from, so if a span is in progress in that thread when the handler is called then it may be needed to not have any span context in the event created in the handler.
3. What are the implications on the Trace API to add events to a Span? Do we make it create LogRecord based Events? In this case, there will be 2 APIs to create events and another problem is that if Trace API is turned off to disable emitting of Spans, it will turn off events as well created using Trace API.

