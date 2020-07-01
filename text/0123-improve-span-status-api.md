# Improve Span Status API

Allow the Span Status API to represent more kinds of status

## Motivation

Right now OpenTelemetry Status is defined as an enumeration of gRPC status codes. Although I couldn't find design criteria written down for this API I fear it is too narrowly defined to be useful across the full breadth of scenarios OpenTelemetry targets.

OpenTelemetry allows Spans to be created to represent any operation, including those which don't involve communication with another component (Kind = Internal). These underlying operations can have native status representations from a particular domain or a language such as POSIX status codes, HRESULTs, many variations of exceptions, error messages, HTTP status, or gRPC status. However to capture the status as part of an OpenTelemetry span it must first be mapped to something in OpenTelemetry's object model and this mapping has the potential to create a few problems:

- **Inconsistency** - If the mapping from native representation to OpenTelemetry representation isn't well-defined then API users or SDK implementations are unlikely to choose the same mapping. This makes collected data hard to work with because it can no longer be treated uniformly.
- **Loss of fidelity** - Mapping from a status representation that distinguishes a large number of different results to one that only distinguishes a few is inherently lossy. This prevents users from isolating different status results they care about. It can also prevent UI from showing useful status information because users don't relate to the reduced representation and the transformation isn't reversible.
- **Conversion difficulty** - If the conversions are non-trivial then they are unlikely to be implemented correctly or perhaps at all. Past feedback suggests end users want to spend little to no effort on this task. SDK implementers may be more diligent but are constrained to native status representations that are known a-priori.

These are challenges for any design of the status API, not solely the current one. We will need to evaluate these issues as a matter of degree and make a judgement call about what is sufficient.

### Goals

In order to determine what status information is useful we first need a common understanding what we expect the collected data to be used for. These are tasks I anticipate tool vendors and end-users would like to be possible with OpenTelemetry status information:

1. **Viewing** - Developers diagnosing a specific distributed trace want to understand the status of spans that occurred while it ran. To do this they want to see status information annotated in the trace, ideally with progressive levels of detail as the focus of investigation narrows. The viewed status information should be easy to correlate back to the domain that generated it and diagnostically useful.

2. **Searching/Filtering** - Developers suspect a particular status condition might be occurring due to customer feedback, some behavior they observed locally, in another trace, or perhaps code review. They want to search collected telemetry to determine if and how often that status condition occurs. If it occurs they want to explore example traces producing it to better understand when it manifests and how it impacts their system. The search terms should be intuitive given the developers initial knowledge about the status condition they are searching for and the domain it arose from.

3. **Grouping and metrics** - Developers monitoring a service want to understand what kinds of status conditions are occurring and how frequently. To do this they want the monitoring tool to provide UI that buckets results into categories and show the top categories ranked by count of occurrence. They may also want metrics that track category counts over time to identify trends and deviations from the trend. Many useful groupings are based on sharing a common problem symptom or common root cause, but other more coarse groupings may be useful in trend analysis. At the most coarse spans can be divided into some definition of "successful" and "failed" but there is no consensus on how some status results should be bucketed (for example http 4xx results).

### Scope

   I hope to define only the tracing API here, not an SDK API nor a wire protocol. I assume that a simple SDK API will follow relatively automatically if we first have agreement on the tracing API and OpenTelemetry can still be successful in an initial release without having standardized a wire protocol.

   The operation represented by a span might be composed of sub-operations (for example a function that calls into other lower level functions) and each of those sub-operations might have a notion of status or error. If those sub-operations aren't represented with a child span then it is out of scope how their status is captured. 

## Explanation

#### Data representation of Span Status

We should aim for representations that require minimal-to-no end-developer translation effort from the native representation and capture at least the key numeric/string data developers are likely to search for or relate to. I suggest status is this set of information:

1. StatusType - The name of the type of error such as "HTTP", "gRPC", "LanguageException", "HRESULT", "POSIX", or "ErrorMessage". The list is end-user-extensible but common status type names should be standardized. (Perhaps there is already some standardization we could borrow?)
2. StatusData - A discriminated union of:
   - An integer, a string, or a tuple of integer and string. These options can be used for:
     - Enumerated status codes: For example an http status code could be represented as 404, "Not Found", or both. In the case of common status codes OpenTelemetry SDK or backend could optionally assist in filling out the remainder of a partially specified enumeration value. For enumerations that aren't well-known the community of users is responsible for determining any conventions.
     - Free-form error messages
   - Exception object - whatever the SDK language's default exception datatype is, if it has one.
   - void
3. SuccessHint - An optional boolean that represents the span author's best guess whether this status represents a successful or failed operation, however they choose to define those terms. For well-known status types I'd suggest the hint be ignored but for user-defined status types this is likely the only clue whether the span should be surfaced in a UI as being abnormal or failed.

Although UI creators are free experiment with how the data is presented I expect most presentations would either be the StatusData alone, or the StatusData qualified with the StatusType and some separator character. For example StringData alone might create names like "FileNotFoundException", "503", "E_FAIL (0x80004005)", "SyntaxError on line 405: Did you forget a semicolon?", and "BadQuery".

Exceptions could have progressive level of detail drilling into messages, stack traces, inner exceptions, links to source, etc if the exporter serialized sufficient data but how and whether that occurs is out-of-scope in this design. 

#### API to capture this data

I would suggest an API called Span.SetStatus(...) that takes all the arguments above, and optionally overloads or default parameters that make common calls easier. For example in C#:

````C#
void SetStatus(Exception spanException, string statusType = "LanguageException", bool? successHint = null);
void SetStatus(string enumNameOrMessage, string statusType = "ErrorMessage", bool? successHint = null);
void SetStatus(int enumValue, string statusType, bool? successHint = null );
void SetStatus(int enumValue, string enumName, string statusType, bool? successHint = null);
void SetStatus(string statusType, bool? successHint);
````



## Internal details

I expect a basic and reasonable implementation would be to define fields on the Span object, set them in the setter and then implement some getter for the exporters to use.

It is also possible for the SDK to destructure the Exception data into simpler serializable types though I'd expect serialization is typically the domain of the exporter and there is a fair amount of policy involved in terms of what data is captured and how it is formatted for transport. There are definite risks that the end-to-end scenario will be less functional or less performant if SDKs intercede here.

## Trade-offs and mitigations

As mentioned in the motivation section, the issues of inconsistency, loss of fidelity and conversion difficulty are all on a sliding scale. I expect this design improves each of these issues at the expensive of some increased object-model/API complexity. Making the OpenTelemetry status representation more expressive also could cause inconsistency problems for the opposite reason - now there might be multiple reasonable representations for a status and the user becomes unclear which one to pick. In general I expect this to be mitigated with documented conventions, API defaults, and the potentially canonicalizing data anywhere in the processing pipeline. 

One place I neglected to go further was defining additional types of error data format beyond string/int/Exception. This might encapsulate things such as key/value pairs (ala structured logging) or more complex or niche status types (COM IErrorInfo). There is nothing inherently problematic with them but I felt the increased expressiveness was getting diminishing returns. One mitigation to key/value pair data in particular would be to put auxilliary data in the span attributes using some convention, for example "Error.UserName"="Bob" might be used together with an error message string "Failed to find user {UserName}". Another mitigation might be adding language specific overloads that handle additional error types.

If all the data collected by the API is transmitted to the back-end this also increases the size of transmitted telemetry, but the exporter authors always retain the freedom to drop or reduce any information they don't believe is valuable with potentially a slight

## Prior art and alternatives

#### Prior discussion

There have been a few past attempts to make improvements here:

- [open-telemetry/oteps#69](https://github.com/open-telemetry/oteps/pull/69)
- [#427](https://github.com/open-telemetry/opentelemetry-specification/pull/427)
- [#432](https://github.com/open-telemetry/opentelemetry-specification/pull/432)
- [#521](https://github.com/open-telemetry/opentelemetry-specification/pull/521)
- [#599](https://github.com/open-telemetry/opentelemetry-specification/issues/599)
- https://gitter.im/open-telemetry/error-events-wg

There are also some links to further prior art within those links, sorry I didn't organize it all nicely here : )

#### Alternatives

**Status information using logging** - It is possible to collect status and error message via log messages and correlate it to a span via trace context. While I have no objection to improved logging capabilities or better integration between distributed tracing and logging I don't believe this is a problem that distributed tracing should abdicate for various reasons:

- Logging isn't a simple or cheap dependency to take in the implementation. While OpenTelemetry may solidify a logging offering for API and SDK components, an end-to-end scenario still requires snooping the log stream to determine relevant messages for a given span. Regardless whether this is done client-side, at the database or in the UI layer it potentially involves performance overhead of handling an order of magnitude more data.
- It requires establishing conventions that designate which log message represents 'the status' for a Span rather than one of potentially many results or errors that were recorded during the Span's duration. This likely means this status is a special case on the logging API if it isn't a special case on the distributed trace API.
- I expect developers both at the time they are emitting trace data and when viewing that trace data would find it is idiomatic to  include the status of the Span's workload together with the description of the Span. Identifying failed or abnormal spans is a typical APM operation.

**Closed vs. open-ended status descriptions** - We could make status represented by a fixed set of options such as the current gRPC codes or any larger/different set. It likely is simpler for tools consuming the data to only operate on fixed options and it is pleasant to imagine that all (or most) status results will be reasonably categorized into the fixed list. However I do not expect most users to take up the responsibility of mapping their status results if OpenTelemetry/back-end tools don't provide a clear representation. Instead these users will not set status or choose options that are generic and inconsistent rendering the data useless to everyone. For a fixed list to work I believe we either have to curate a very large list of options or accept that significant portions of the developer audience will not find the status representation useful.

**API using attribute semantic conventions ** - It is also possible to do this via semantic conventions on attributes and although I think the strongly typed API is preferable I don't have that strong of an opinion. Semantic conventions are likely to have higher performance overhead, higher risk of error in key names and are less discoverable/refactorable in IDEs. The advantages are that there is some past precedent for doing this specific to http codes and new conventions can be added easily. If we go semantic conventions it does imply that Exception becomes a type that can be directly passed as an argument to SetAttribute(). Requiring the user to destructure the exception into a list of key value pairs would be overly onerous and error-prone for a common usage scenario. If desired the SDK or exporter could destructure it, but that can be determined independently from API design and I'd like to keep it out of scope.

**API using Span events** - Most of the rationale for attribute semantic conventions also applies here, events are effectively another key-value store mechanism. The timestamp that is attached to an event appears to hold little value as status is probably produced at the same approximate time the span end timestamp is recorded. Similar to attribute conventions it sounds like there is precedent for storing some errors as events.

## Open questions

1. Although I specified that common status types could be given standardized names, I didn't define what that list was. We would need to define what criteria makes a status type common enough to be on the list and maintain it over time. 

2. Above the design mentioned that the SDK might fill in the name or integer value of a well-known status code when only one of the two was specified by a user. We'd have to decide if that is functionality we want, and which values are included in mapping tables.
3. I left the SDK API out-of-scope, but we will need a way to retrieve the stored data via the SDK before adding data to span has any value in an end-to-end scenario.