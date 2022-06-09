# Introducing Events and Logs API

In this OTEP we introduce an Events and Logs API that is based on the OpenTelemetry Log signal. The Events here refer to the independent Events and not to be confused with Span Events which occur only in the context of a span. In OpenTelemetry's perspective Log Records and Events are different names for the same concept - however, there is a subtle difference in how they are represented using the underlying data model that is described below. We will describe why the existing Logging APIs are not sufficient for the purpose of creating events.  It will then be evident that we will need an API in OpenTelementry for creating events.

The Logs part of the API is supposed to be used only by the Log Appenders and end-users must continue to use the logging APIs available in the languages.

## Subtle differences between Logs and Events

Logs have a mandatory severity level as a first-class parameter that events do not have, and events have a mandatory name that logs do not have. Further, logs typically have messages in string form and events have data in the form of key-value pairs. It is due to this that their API interface requirements are slightly different.

## Who requires Events API

Here are a few situations that require recording of Events, there will be more.

- RUM events (Client-side instrumentation)
- Recording kubernetes events
- Recording eBPF events
- Collector Entity events [link](https://docs.google.com/document/d/1Tg18sIck3Nakxtd3TFFcIjrmRO_0GLMdHXylVqBQmJA/edit)
- Few other event systems described in [example mappings](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#appendix-a-example-mappings) in the data model.

## Can the current Log API interfaces be used for events?

1. The log level is fundamental to the Log APIs in almost all the languages; all the methods in the Log interface are named after the log level, and there is usually no generic method to submit a log entry without log level.

  - In JavaScript for Web, the standard method of logging is to use console.log. Events can be created using [Event/CustomEvent](https://developer.mozilla.org/en-US/docs/Web/Events/Creating_and_triggering_events) interfaces.However, there is no option to define custom destination for these logs and events. Logs go only to console and event listeners are attached to the DOM element that dispatches it.
  - In Android, android.util.Log has methods  Log.v(), Log.d(), Log.i(), Log.w(), and Log.e() to write logs. These methods correspond to the severity level.
  - Swift on iOS has Logger interface that has methods corresponding to severity level too.

2. The current Log APIs do not have a standard way to pass event attributes.

  - It may be possible to use the interpolation string args as the parameter to pass event attributes. However, the logging spec seems to map the human readable message (which is obtained after replacing the args in the interpolated string)  to the Body field of LogRecord.
  - Log4j has an EventLogger interface that can be used to create structured messages with arbitrary key-value pairs, but log4j is not commonly used in Android apps as it is not officially supported on Android as per this [Stack Overflow thread](https://stackoverflow.com/questions/60398799/disable-log4j-jmx-on-android/60407849#60407849) by one of log4j’s maintainers.
  - In Python, logging.LogRecord's extra field is mapped to Otel LogRecord's attributes but this field is a hidden field and not part of the public interface.

3. The current Log APIs have a message parameter which could map to the Body field of LogRecord. However, this is restricted to String messages and does not allow for structured logs.

For the above reasons we can conclude that we will need an API for generating Events API

## Should OpenTelemetry have an API for logs?

A part of the OTel community thnks that we should not have a full-fledged logging API unless there is a language that doesn't already have a plethora of logging libraries & APIs to choose from where it might make sense to define one. Further, we will not be able to have the [rich set of configuration options](https://logging.apache.org/log4j/2.x/manual/configuration.html) that some popular logging frameworks provide so the logging API in OTel will only become yet another API. However, it was noted that the Log Appender API is very similar to the API for Events and so instead of having API for Events and API for Log Appenders separately it was agreed to have one API for Events and Logs, and that the API for Logs is targetted only to Log Appenders.

## Events and Logs API Interface

For reference, a prototype of the Events and Logs API in Java is [here](https://github.com/scheler/opentelemetry-java/pull/1/files)

Client-side telemetry is one of the initial clients that will use the Events API and so the API will be made available in JavaScript, Java and Swift first to be able to use in the SDKs for Browser, Android and iOS.  It may also be added in Go since there is a Kubernetes events receiver implemented in Collector based on Logs data-model.

The Events and Logs API consist of these main classes:

* LoggerProvider is the entry point of the API. It provides access to Loggers.
* Logger is the class responsible for creating events using Log records.


### LoggerProvider

Logger can be accessed with an LoggerProvider.

In implementations of the API, the LoggerProvider is expected to be the stateful object that holds any configuration.

Normally, the LoggerProvider is expected to be accessed from a central place. Thus, the API SHOULD provide a way to set/register and access a global default LoggerProvider.

Notwithstanding any global LoggerProvider, some applications may want to or have to use multiple LoggerProvider instances, e.g. to have different configuration (like LogRecordProcessors) for each (and consequently for the Loggers obtained from them), or because it's easier with dependency injection frameworks. Thus, implementations of LoggerProvider SHOULD allow creating an arbitrary number of Logger instances.

#### LoggerProvider operations

The LoggerProvider MUST provide the following functions:

* Get an Logger

##### Get an Logger

This API MUST accept the following parameters:

- name (required): This name SHOULD uniquely identify the instrumentation scope, such as the instrumentation library (e.g. io.opentelemetry.contrib.mongodb), package, module or class name.  If an application or library has built-in OpenTelemetry instrumentation, both Instrumented library and Instrumentation library may refer to the same library. In that scenario, the name denotes a module name or component name within that library or application. In case an invalid name (null or empty string) is specified, a working Logger implementation MUST be returned as a fallback rather than returning null or throwing an exception, its name property SHOULD be set to an empty string, and a message reporting that the specified value is invalid SHOULD be logged. A library implementing the OpenTelemetry API may also ignore this name and return a default instance for all calls, if it does not support "named" functionality (e.g. an implementation which is not even observability-related). A LoggerProvider could also return a no-op Logger here if application owners configure the SDK to suppress telemetry produced by this library.

- version (optional): Specifies the version of the instrumentation scope if the scope has a version (e.g. a library version). Example value: 1.0.0.
- schema_url (optional): Specifies the Schema URL that should be recorded in the emitted telemetry
- event_domain (optional): Specifies the domain for the events created, which should be added in the attribute `event.domain` in the instrumentation scope.
- pass_context (optional): Specifies whether the Trace Context should automatically be passed on to the events and logs created by the Logger. This SHOULD be false by default.

It is unspecified whether or under which conditions the same or different Logger instances are returned from this function.

Implementations MUST NOT require users to repeatedly obtain an Logger again with the same name+version+schema_url+event_domain+pass_context to pick up configuration changes. This can be achieved either by allowing to work with an outdated configuration or by ensuring that new configuration applies also to previously returned Loggers.

Note: This could, for example, be implemented by storing any mutable configuration in the LoggerProvider and having Logger implementation objects have a reference to the LoggerProvider from which they were obtained. If configuration must be stored per-Logger (such as disabling a certain Logger), the Logger could, for example, do a look-up with its name+version+schema_url in a map in the LoggerProvider, or the LoggerProvider could maintain a registry of all returned Loggers and actively update their configuration if it changes.

The effect of associating a Schema URL with a Logger MUST be that the telemetry emitted using the Logger will be associated with the Schema URL, provided that the emitted data format is capable of representing such association.

### Logger

The Logger is responsible for creating Events and Logs.

Note that Loggers should usually not be responsible for configuration. This should be the responsibility of the LoggerProvider instead.

#### Logger operations

The Logger MUST provide functions to:

- Function named “logEvent” to create an Event with the provided event name and attributes. The event name provided should be inserted as an attribute with key “event.name”. It will override any attribute with the same key in the attributes passed.
- Function named “recordException” to record an Exception as an Event. This is to facilitate recording an exception outside of a trace context for languages that already do not support recording an exception in a log message. This should work similar to [Record Exception](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/api.md#record-exception) in Trace API, with the following semantics
  - This is a specialized variant of logEvent function, so for anything not specified here, the same requirements as for logEvent apply.
  - The signature of the method is to be determined by each language and can be overloaded as appropriate. The method MUST record an exception as an Event with the conventions outlined in the exception semantic conventions document. The minimum required argument SHOULD be no more than only an exception object.
  - If RecordException is provided, the method MUST accept an optional parameter to provide any additional event attributes (this SHOULD be done in the same way as for the AddEvent method). If attributes with the same name would be generated by the method already, the additional attributes take precedence.
  - Note: RecordException may be seen as a variant of logEvent with additional exception-specific parameters and all other parameters being optional (because they have defaults from the [exception semantic convention](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/exceptions.md)).

A LogRecord representing exception event may look like this:

```
{
  time_unix_nano: 1653682410,
  Attributes: {
   event.name: exception,
   exception.type:  OSError,
   exception.message: Division by Zero,
   exception.stacktrace: "",
   exception.escaped: false
 }
}
```

The Logger SHOULD additionally provide the following functions for convenience:

- Create a new Event using Log Record data model.
- Create a new Log Record using LogRecord data model.

### Usage

```java
OpenTelemetry openTelemetry = OpenTelemetry.noop();
Logger logger = openTelemetry.getLogger("my-scope");

// Using the convenience method to log an event directly
logger.logEvent("network-changed", 
                 Attributes.builder().put("type", "wifi").build());

// Using the event builder to log an event
logger.eventBuilder("page-navigated").build().setAttribute("url", "http://foo.com/bar#new").emit();

// Using the logRecord builder to log a record
Logger logger = openTelemetry.getLogger("another-scope");
logger.logRecordBuilder().build().setBody("I am a log message").emit();

```

### Usage in Client-side telemetry

```java
public void addBrowserEvent(String name, Attributes attributes) {
   Logger logger = openTelemetry.getLogger("my-scope", "1.0", "browser");
   logger.logEvent(name, attributes);
}

public void addMobileEvent(String name, Attributes attributes) {
   Logger logger = openTelemetry.getLogger("my-scope", "1.0", "mobile");
   logger.logEvent(name, attributes);
}
```

### Usage in eBPF

From the eBPF [demo](https://youtu.be/F1VTRqEC8Ng?t=233), it looks like they have chosen to put the Event data in the Body field of LogRecord instead of Attributes.  They might have to make the changes to move the event data to attributes to conform to this API specification.

## Semantic Convention for event attributes

**type:** `event`

**Description:** Event attributes.

| Attribute  | Type | Description  | Examples  | Required |
|---|---|---|---|---|
| `event.name` | string | Name or type of the event. | `network-change`; `button-click`; `exception` | Yes |
| `event.domain` | string | Domain or scope for the event. | `profiling`; `browser`, `db`, `k8s` | No |

An `event.name` is supposed to be unique only in the context of an `event.domain`, so this allows for two events in different domains to have same `event.name`. No claim is made about the uniqueness of `event.name`s in the absence of `event.domain`.

## Causality on Events

For creating causality between events we can create wrapper spans that are part of the same trace. However, note that the events themselves are represented using LogRecords and not as Span Events.

```java
Span s1 = Trace.startSpan()
    addBrowserEvent(e1name, attributes)
    Span s2 = Trace.startSpan()
        addBrowserEvent(e2name, attributes)
    s2.end()
s1.end()
```

## Comparing with Span Events

- Span Events are events recorded within spans using Trace API. It is not possible to create independent Events using Trace API. The Events API must be used instead.
- Span Events were added in the Trace spec when Logs spec was in early stages. Ideally, Events should only be recorded using LogRecords and correlated with Spans by adding Span Context in the LogRecords. However, since Trace API spec is stable Span Events MUST continue to be supported.
- We may add a configuration option to the `TracerProvider` to create LogRecords for the Span Events and associate them with the Span using Span Context in the LogRecords. Note that in this case, if a noop TracerProvider is used it will not produce LogRecords for the Span Events.
