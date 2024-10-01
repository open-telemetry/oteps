# Structured Logs

## Motivation

With the definition of Events there has been additional contentious and confusion around the difference between what IS and Event vs a Structured Log when they are both represented and transported as a `type` of OpenTelemetry Log.

This also avoids overloading of the "Events" definition to include Structured Logs, as they have different intents and purposes.

And this avoids the "Hack" approach like that used for Zero length Spans to simulate Events, before Events had been defined. With the introduction of Events and support in the OpenTelemetry Semantic Convention tooling it is an extremely trivial task to define a new `type` definition for Structured Logs whch can be used to define the structure of the log within Semantic Conventions yaml and avoiding the multiple needs of overloading the term "Event", overloading the usage of the `event.name` and to support the Semantic differences between Events and Structured Logs.

Like supporting "Unnamed" Structured Logs, which would not be possible with the current Event definition as events MUST have a `name` field and all events with the same `name` are comparable.

### What are OpenTelemetry Structured Logs?

As with Events, Structured Logs are a type of OpenTelemetry Log that has a specific structure implied by some semantic conventions that apply the semantic meaning of the log.

As Logs are a core concept in OpenTelemetry Semantic Conventions, Structured Logs are a specific definition of how those logs should be represented and structured to enable interoperability and consistency across the OpenTelemetry ecosystem.

They also provide a way to define a common structure for logs that are emitted by a specific library or system. This allows for interoperability and consistency across the OpenTelemetry ecosystem, and enables users to easily understand and analyze logs emitted by different libraries or systems.

### OTLP

Since OpenTelemetry Structured Logs are a type of OpenTelemetry Log, they share the same OTLP log data structure and pipeline.

### Semantic Conventions

As part of describing the structure of OpenTelemetry Structured Logs, we will define a new OpenTelemetry Semantic Convention `type` that will be used to define the structure of the log within Semantic Conventions yaml. One possible option is to use the `log` type to define the structure of the log, just as the Event uses `event` for both Log based and Span based Events.

### API

To support OpenTelemetry Structured Logs, the OpenTelemetry Logs API will include a new `StructuredLog` method that will allow users to emit OpenTelemetry Structured Logs conforming to the required Semantic Conventions.

### Interoperability with other logging libraries

The emitting of Structured Logs via the OpenTelemetry Logs API is not intended to replace the emitting of logs via other logging libraries. Instead, OpenTelemetry SHOULD provide a way to send OpenTelemetry Structured Logs from the OpenTelemetry Logs API to other logging libraries (e.g., Log4j). This allows users to integrate OpenTelemetry Logs into an existing (non-OpenTelemetry) log stream.

As with Events it is permissible for logging libraries to emit OpenTelemetry Structured Logs directly via the `Bridge` API, as long as they conform to the Semantics Conventions of the identified Structured Log Record.

### Relationship to Events

The intent of Events are to describe [characteristics of a specific event that has occurred in the system](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/general/events.md) which may be used to support interactive querying for observability operations like Dashboards, User Monitor, Alerting, Tracking of how an Application / system is used. And they SHOULD be small and concise as they are not intended to provide Detailed Information.

As with standard (Unstructured) Logging the intent of a Structured Logs is to provide additional context or Detailed Information but in a well defined "known" format, these may be used to dive deeper into a system to understand why or what occurred. You would use Structured Logs whenever you are defining an logging log level of details in a manner that requires a common known structure such as Logs being emitted from a specific library or system.

And while both Events and Structured Logs shared the common OpenTelemetry Log representation, they each MUST to retain their original Semantic Meaning (intent) and structure so that while being processed through the entire OpenTelemetry pipeline (API, SDK, Processors, Exporters, Collectors and at Rest) they can be deterministically evaluated to obtain their original intent without loss of information.

### OpenTelemetry Logs and the Relationships to Structured Logs

At this point it is important to clarify the relationship between OpenTelemetry Logs, OpenTelemetry Event and Structured Logs from their explicit common representation as a `LogRecord` in the OpenTelemetry API.

Stating what are the `types` of OpenTelemetry Logs when they are packaged and represented as their required `LogRecord` representations, using their defined Semantic Conventions to understand their structure:

* An OpenTelemetry Log IS and should be treated as an event if the type of `LogRecord` (the representation) includes the `event.name` attribute.
* An OpenTelemetry Log IS and should be treated as a Structured Log if the type `LogRecord` (the representation) includes the TBD??? (Open Questions)
  * Is could be determined by, Overloading the existing Event definition to include Structured Logs.
    * By defining a single specific `event.name` for ALL Structured Logs. (eg. `log.structured`), which would make ALL Structured Logs a sub-type of Event, and deterministically identifiable, but it would break the Event Semantic Conventions that all Events with the same `event.name` are comparable.
    * By defining a explicit convention for Structured Logs within the `event.name` field where they MUST have a specific prefix such as `log.`, along with the other agreed upon restrictions and fields. Which would also make Structured Logs a type of Event, and deterministically identifiable, but would then require that every Structured Long have a specific and unique name, so that when combined with the prefix it would continue to follow the Event Semantic Conventions, that all `event.name` MUST be namespaced and unique to avoid confusion.
  * Defined by a new `log.name` field that is used to identify the Structured Log, with a default `name` of "log", "unnamed" or "unknown" if not provided. Keeping both the `event.name` and `log.name` fields separate and distinct. This will also have a similar set of issues as identified above for `event.name` in terms of comparability and uniqueness.
    * While this would then allow for a Detailed Structured Log and a concise Event definition to "use" the same logical name which could lead to further confusion, it would also allow for more explicit and easier back-end query joins between the two types of logs.
    * Alternatives may include `log.record.name` etc.
* An OpenTelemetry Log IS NOT and should not be treated as either a Structured Log or an Event if the type of `LogRecord` (the representation) does not include the required attributes to identify it as a Structured Log or an Event.

So to close the loop and explicitly define the relationships this means that:

* While ALL OpenTelemetry Structured Logs are a `type` of OpenTelemetry Log, this does NOT imply that all OpenTelemetry Logs are Structured Logs.
* While ALL OpenTelemetry Events are also a `type` of OpenTelemetry Logs, this does NOT imply that all OpenTelemetry Logs are Structured Events. (depending on the above Open Questions)
* Just as two different OpenTelemetry Events are not comparable (unless they shared the same `event.name`) the fact that both Structure Logs and Events are both `types` of OpenTelemetry Logs this does not imply that ALL OpenTelemetry Structured Logs are also OpenTelemetry Events and therefore can and should be treated the same. (again depending on the Open Questions above).

### Explicit Differences from Events

* Structured Logs DO NOT require a specific `Name` field / attribute, they may be unnamed and they may use other attributes to Identify their structure.
* Each OpenTelemetry Structured Log MUST provide the required identifying characteristics of the Structured Logs so that anyone receiving the evaluate the values.
  * Name, schema, severity, comination of fields, etc.
* Structured Logs SHOULD provide a `body` field.
* Structured Logs SHOULD provide a `severity` field.
* Structured Logs MUST not use the `event.name` attribute, as this would identify it as an Event, which would confuse / lose the original intent of the generated OpenTelemetry Log (`LogRecord`).
* OpenTelemetry Events have a unique requirement and explicit formatting on the `Name` (`event.name`), structured logs do NOT have this limitation. ?? (Open Question)

### SDK

This refers to the public facing OpenTelemetry Logs SDK.

## Alternatives

With the introduction of Events and the definition of the public Logs API we are still in the early stages of identifying how Structured Logs can retain their Semantic Meaning and definition.

Alternatives may include:

* What is the new Semantic Convention `type` for Structured Logs (`log`).
* Overloading the existing Event definition to include Structured Logs.
  * By defining a single specific `event.name` for Structured Logs. (`log.structured`)
  * By defining a explicit convention for Structured Logs within the `event.name` field where they MUST have a specific prefix such as `log.`, along with the other agreed upon restrictions and fields.
* Defining a new top level Attribute to identify / name the Structured Log, in a similar manner as the `event.name` field.
* Extending the OpenTelemetry Logs definition to a new top-level field (Also an open question for Events) to identify the Structured Log.

## Open questions

This is a similar set of Open Questions as with Events as they are closely related and are both represented as `types` of OpenTelemetry Logs.

* How to represent the structure of OpenTelemetry Structured Logs in the OpenTelemetry Semantic Conventions? (new `type` field in `yaml`)
* How to represent the structure of OpenTelemetry Structured Logs in the OpenTelemetry Logs API?
* How to identify the intent of the OpenTelemetry Structured Logs vs an Unstructured Log or an Event?
* How to support routing logs from the Logs API to a language-specific logging library
  while simultaneously routing logs from the language-specific logging library to an OpenTelemetry Logging Exporter?
* Should the Logs API have an `Enabled` function based on severity level and event name?
