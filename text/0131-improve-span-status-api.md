# Improve Span Status API

Allow the Span Status API to represent more kinds of status

## Motivation

Right now OpenTelemetry Status is defined as an enumeration of gRPC status
codes. Although I couldn't find design criteria written down for this API
I fear it is too narrowly defined to be useful across the full breadth of
scenarios OpenTelemetry targets.

OpenTelemetry allows Spans to be created to represent any operation, including
those which don't involve communication with another component (Kind =
Internal). These underlying operations can have native status representations
from a particular domain or a language such as POSIX status codes, HRESULTs,
HTTP status, or gRPC status.
However to capture the status as part of an OpenTelemetry span it must first be
mapped to something in OpenTelemetry's object model and this mapping has the
potential to create a few problems:

- **Inconsistency** - If the mapping from native representation to
  OpenTelemetry representation isn't well-defined then API users or SDK
  implementations are unlikely to choose the same mapping. This makes collected
  data hard to work with because it can no longer be treated uniformly.
- **Loss of fidelity** - Mapping from a status representation that
  distinguishes a large number of different results to one that only
  distinguishes a few is inherently lossy. This prevents users from isolating
  different status results they care about. It can also prevent UI from showing
  useful status information because users don't relate to the reduced
  representation and the transformation isn't reversible.
- **Conversion difficulty** - If the conversions are non-trivial then they
  are unlikely to be implemented correctly or perhaps at all. Past feedback
  suggests end users want to spend little to no effort on this task. SDK
  implementers may be more diligent but are constrained to native status
  representations that are known a-priori.

These are challenges for any design of the status API, not solely the current
one. We will need to evaluate these issues as a matter of degree and make a
judgement call about what is sufficient.

### Goals

In order to determine what status information is useful we first need a common
understanding what we expect the collected data to be used for. These are
tasks I anticipate tool vendors and end-users would like to be possible with
OpenTelemetry status information:

1. **Viewing** - Developers diagnosing a specific distributed trace want to
   understand the status of spans that occurred while it ran. To do this they want
   to see status information annotated in the trace, ideally with progressive
   levels of detail as the focus of investigation narrows. The viewed status
   information should be easy to correlate back to the domain that generated it
   and diagnostically useful.

2. **Searching/Filtering** - Developers suspect a particular status condition
   might be occurring due to customer feedback, some behavior they observed
   locally, in another trace, or perhaps code review. They want to search
   collected telemetry to determine if and how often that status condition
   occurs. If it occurs they want to explore example traces producing it to
   better understand when it manifests and how it impacts their system. The
   search terms should be intuitive given the developers initial knowledge
   about the status condition they are searching for and the domain it arose from.

3. **Grouping and metrics** - Developers monitoring a service want to
   understand what kinds of status conditions are occurring and how frequently.
   To do this they want the monitoring tool to provide UI that buckets results
   into categories and show the top categories ranked by count of occurrence.
   They may also want metrics that track category counts over time to identify
   trends and deviations from the trend. Many useful groupings are based on
   sharing a common problem symptom or common root cause, but other more coarse
   groupings may be useful in trend analysis. At the most coarse spans can be
   divided into some definition of "successful" and "failed" but there is no
   consensus on how some status results should be bucketed (for example http
   4xx results).

### Scope

I hope to define only the tracing API here, not an SDK API nor a wire
protocol. I assume that a simple SDK API will follow relatively automatically
if we first have agreement on the tracing API and OpenTelemetry can still be
successful in an initial release without having standardized a wire protocol.

The operation represented by a span might be composed of sub-operations (for
example a function that calls into other lower level functions) and each of
those sub-operations might have a notion of status. If those sub-operations
aren't represented with a child span then it is out of scope how their status is
captured.

## Explanation

#### Data representation of Span Status

We should aim for representations that require minimal-to-no end-developer
translation effort from the native representation and capture at least the key
numeric/string data developers are likely to search for or relate to. I suggest
status is this set of information:

1. Domain - The name of the domain the status data applies to such as "HTTP",
   "gRPC", "HRESULT", "POSIX". The list is end-user-extensible but common status
   type names should be standardized.
   (Perhaps there is already some standardization we could borrow?)
2. Code - An integer status code. Can be combined with an status message. Either
   a code or message are required.
3. Message - A string status message. Can be combined with a status code. Either
   a code or message are required.

Although UI creators are free experiment with how the data is presented I
expect most presentations would either be the StatusData alone, or the
StatusData qualified with the StatusType and some separator character. For
example StatusData alone might create names like "503", "E_FAIL (0x80004005)",
"Status Code 12: Unimplemented".

#### API to capture this data

I would suggest an API called Span.SetStatus(...) that takes all the arguments
above, and optionally overloads or default parameters that make common calls
easier. For example in C#:

```C#
void SetStatus(string domain, int code, string message);
void SetStatus(string domain, string message);
void SetStatus(string domain, int code);
```

## Internal details

I expect a basic and reasonable implementation would be to define fields on the
Span object, set them in the setter and then implement some getter for the
exporters to use. A second possibility is to store the data into attributes or
events. The choice of storage may have some modest effects on memory usage
but primarily I expect the choice would be driven by the SDK API we want to
read stored data back.

## Trade-offs and mitigations

As mentioned in the motivation section, the issues of inconsistency, loss of
fidelity and conversion difficulty are all on a sliding scale. I expect this
design improves each of these issues at the expensive of some increased
object-model/API complexity. Making the OpenTelemetry status representation
more expressive also could cause inconsistency problems for the opposite reason

- now there might be multiple reasonable representations for a status and the
  user becomes unclear which one to pick. In general I expect this to be mitigated
  with documented conventions, API defaults, and the potentially canonicalizing
  data anywhere in the processing pipeline.

The current proposal only gives one string which can be used either for a
freeform message or a textual status code. Adding a 2nd string to the
StatusData would allow both to be collected side by side. This is another example of
increasing expressivity at the cost of some complexity. I'd be happy to see
this added if the community agreed.

If all the data collected by the API is transmitted to the back-end this also
increases the size of transmitted telemetry, but the exporter authors always
retain the freedom to drop or reduce any information they don't believe is
valuable with potentially a slight

## Prior art and alternatives

#### Prior discussion

There have been a few past attempts to make improvements here:

- [open-telemetry/oteps#123](https://github.com/open-telemetry/oteps/pull/123)
- https://gitter.im/open-telemetry/error-events-wg

#### Alternatives

**Status information using logging** - It is possible to collect status and
error message via log messages and correlate it to a span via trace context.
While I have no objection to improved logging capabilities or better
integration between distributed tracing and logging I don't believe this is
a problem that distributed tracing should abdicate for various reasons:

- Logging isn't a simple or cheap dependency to take in the implementation.
  While OpenTelemetry may solidify a logging offering for API and SDK components,
  an end-to-end scenario still requires snooping the log stream to determine
  relevant messages for a given span. Regardless whether this is done
  client-side, at the database or in the UI layer it potentially involves
  performance overhead of handling an order of magnitude more data.
- It requires establishing conventions that designate which log message
  represents 'the status' for a Span rather than one of potentially many results
  or errors that were recorded during the Span's duration. This likely means this
  status is a special case on the logging API if it isn't a special case on the
  distributed trace API.
- I expect developers both at the time they are emitting trace data and when
  viewing that trace data would find it is idiomatic to include the status of
  the Span's workload together with the description of the Span. Identifying
  failed or abnormal spans is a typical APM operation.

**Closed vs. open-ended status descriptions** - We could make status
represented by a fixed set of options such as the current gRPC codes
or any larger/different set. It likely is simpler for tools consuming
the data to only operate on fixed options and it is pleasant to imagine that
all (or most) status results will be reasonably categorized into the fixed
list. However I do not expect most users to take up the responsibility of
mapping their status results if OpenTelemetry/back-end tools don't provide
a clear representation. Instead these users will not set status or choose
options that are generic and inconsistent rendering the data useless to
everyone. For a fixed list to work I believe we either have to curate a very
large list of options or accept that significant portions of the developer
audience will not find the status representation useful.

**API using attribute semantic conventions** - It is also possible to do this
via semantic conventions on attributes and although I think the strongly typed
API is preferable I don't have that strong of an opinion. Semantic conventions
are likely to have higher performance overhead, higher risk of error in key
names and are less discoverable/refactorable in IDEs. The advantages are that
there is some past precedent for doing this specific to http codes and new
conventions can be added easily.

**API using Span event semantic conventions** - Most of the rationale for attribute semantic
conventions also applies here, events are effectively another key-value store
mechanism. The timestamp that is attached to an event appears to hold little
value as status is probably produced at the same approximate time the span end
timestamp is recorded.

**Move the API to a non-core package** - It is possible to have the Tracing
API expose status using Attribute or Event APIs, and then have a 2nd library
that exposes a strongly-typed API that converts the status information into
attrbute/event updates. This adds some complexity over a strongly typed API
declared directly on Span but if we identify this as an area that needs to be
more decoupled/versionable than other Span tracing APIs perhaps it would be
valuable.

## Open questions

1. Although I specified that common status types could be given standardized
   names, I didn't define what that list was. We would need to define what
   criteria makes a status type common enough to be on the list and maintain it
   over time.
2. Above the design mentioned that the SDK might fill in the name or integer
   value of a well-known status code when only one of the two was specified by a
   user. We'd have to decide if that is functionality we want, and which values
   are included in mapping tables.
3. I left the SDK API out-of-scope, but we will need a way to retrieve the
   stored data via the SDK before adding data to span has any value in an
   end-to-end scenario.
