# Spec in spans

Specifying the structure of trace and metric data in logs.

## Motivation

There are multiple requests and issues related to converting metric/tracing data into logs:
* [Issue 398](https://github.com/open-telemetry/opentelemetry-specification/issues/398)
* [Issue 617](https://github.com/open-telemetry/opentelemetry-specification/issues/617)

This OTEP was created based on the following comment from Joshua MacDonald in the gitter channel on Jun 11 2020
> HI all, I've been thinking about how this Logging sub-project could help more with tracing and metrics.
I was thinking that before beginning to work on Logging API design, we could try to standardize the translation
from Span and Metric events into structured logging events. This was requested as a feature for the metric
SDK (open-telemetry/opentelemetry-specification#617), and I think the same could be said of spans.
Right now many OTel repos are including standard-output exporters for both trace and metrics, but there are no
standard conventional data formats for them to use, so it's a bit arbitrary what you get. I think it would be great
if we had a standard structured log form for both metrics and spans.
I would add that the OTel metrics API was conceived as a logging API from the start. The data model describes metric
events, which can be recorded as a faithful way to capture metrics, so we would just need to standardize how they are
structured as log records. 
Here's the event structure:
https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/metrics/api.md#metric-event-format
@jmacd do you mean a standardized way to encapsulate span data and metric data inside a log record or something else?



Why should we make this change? What new value would it bring? What use cases does it enable?

## Explanation

Sometimes, a system needs to log tracing or metric data into an easy to parse log format. In those cases, the data should be logged using the following conversion tables.

|Open Telemetry Metric field | Open Telemetry Log field|
|--- |--- |
| time of metric collection  |  log.Timestamp   |
| correlation Context   |  log.TraceId   |
| span Context   |  log.SpanId  |
| sampled  |  log.TraceFlags  |
| TRACE3  |  log.SeverityText  |
| 3  |  log.SeverityNumber  |
| metric.label  |  log.Name  |
| metric.value  |  log.Body.value  |
| metric.resources |  log.Resource  |
| metric.definition  |  log.Attributes  |
| metric.meterIdentifier  |  log.Attributes.meterIdentifier  |
| metric.AggregationType  |  log.Attributes.AggregationType  |
| metric.Instrumentation  |  log.Attributes.Instrumentation  |


|Open Telemetry Trace field| Open Telemetry Log field|
|--- |--- |
| trace.start.timestamp   |  log.Timestamp   |
| trace.TraceId   |  log.TraceId   |
| trace.SpanId  |  log.SpanId  |
| trace flags  |  log.TraceFlags  |
| TRACE3  |  log.SeverityText  |
| 3 |  log.SeverityNumber  |
| trace.name |  log.Name  |
| trace.start.timestamp  |  log.Body.StartTimeStamp  |
| trace.end.timestamp  |  log.Body.EndTimeStamp  |
| trace semantic conventions**  |  log.Resource  |
| trace.attributes  |  log.Attributes  |
| trace.Events  |  log.Attributes.Events  |
| trace.Tracestate  |  log.Attributes.Tracestate  |

**Semantic conventions for Traces can be mapped directly to log.Resource.convention  For example `db.type` would be converted to `{Resource: {db.type: value} }`

Examples:
### Open Metric
### Open Trace
### Zipkin Trace
### Jaeger Trace
### Prometheus Metric
### Suggested example from Issue 398
17:05:43 INFO  {sampled=true, spanId=ce751d1ad8be9d11, traceId=ce751d1ad8be9d11} [or.ac.qu.GreetingResource] (executor-thread-1) hello


## Internal details

Every SDK or API would implement this conversion themselves. This is merely a standard mapping to doing that conversion.

## Trade-offs and mitigations

One drawback is the limited scope of this OTEP in not handling the actual conversion of these fields. This could be mitigated by creating a Metric or Trace conversion library once the Log API/SDK is defined.


## Prior art and alternatives

* It's possible to have the metric definition, value, and label all be inserted into the log attribute or body, however, leaving the body empty except for the metric value will provide better aggregation capabilities.
What are some prior and/or alternative approaches? For instance, is there a corresponding feature in OpenTracing or OpenCensus? What are some ideas that you have rejected?

## Open questions

What are some questions that you know aren't resolved yet by the OTEP? These may be questions that could be answered through further discussion, implementation experiments, or anything else that the future may bring.

## Future possibilities

What are some future changes that this proposal would enable?
