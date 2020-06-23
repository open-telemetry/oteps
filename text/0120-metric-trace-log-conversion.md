# Metric and Trace log conversion

Specifying the structure of trace and metric data in logs.

## Motivation

There are multiple requests and issues related to converting metric/tracing data into logs:

* [Issue 398](https://github.com/open-telemetry/opentelemetry-specification/issues/398)
* [Issue 617](https://github.com/open-telemetry/opentelemetry-specification/issues/617)

The [aim here](https://gitter.im/open-telemetry/logs?at=5ee284f2ef5c1c28f0194a89) is to create a standard method to convert a trace or metric into a log for Otel exporters to reduce confusion and increase compatibility.

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
| trace.sampled  |  log.TraceFlags  |
| TRACE3  |  log.SeverityText  |
| 3 |  log.SeverityNumber  |
| trace.name |  log.Name  |
| trace.start.timestamp  |  log.Body.StartTimeStamp  |
| trace.end.timestamp  |  log.Body.EndTimeStamp  |
| trace.parentId  |  log.Body.ParentId  |
| trace semantic conventions**  |  log.Resource  |
| trace.attributes  |  log.Attributes  |
| trace.Events  |  log.Attributes.Events  |
| trace.Tracestate  |  log.Attributes.Tracestate  |
| trace.status  |  log.Attributes.status  |
| trace.kind  |  log.Attributes.kind  |

** Semantic conventions for Traces can be mapped directly to log.Resource.convention  For example `db.type` would be converted to `{Resource: {db.type: value} }`

Examples:

### Open Metric

Based on a sample metric from [Alan Storm](https://alanstorm.com/what-are-open-telemetry-metrics-and-exporters/)
The following Metric

```

{
  descriptor: {
    name: 'my-open-telemetry-counter',
    description: 'A simple counter',
    unit: '1',
    metricKind: 0,
    valueType: 1,
    labelKeys: [],
    monotonic: true
  },
  labels: {},
  aggregator: CounterSumAggregator {
    _current: 1,
    _lastUpdateTime: [ 1589046826, 890210944 ]
  }
}

```

Would be logged as:

```

{
 "timestamp": 1589046070557,
 "traceid": <sometraceid>,
 "spanid": <somespanid>,
 "traceflags": {},
 "severityText": "TRACE3",
 "severityNumber": 3,
 "name": "my-open-telemetry-counter",
 "body": {
         "value": 1,
       },
 "attributes": {
         "aggregagator": "CounterSumAggregator",
         "meterIdentifier": "my-meter",
         "description": "A simple counter",
          "unit": '1',
          "metricKind": 0,
          "valueType": 1,
          "labelKeys": [],
          "monotonic": true
  },
  "resource": {}
}

```

### Open Trace

The following [sample trace](https://opentelemetry-python.readthedocs.io/en/stable/getting-started.html):

```

{
    "name": "baz",
    "context": {
        "trace_id": "0xb51058883c02f880111c959f3aa786a2",
        "span_id": "0xb2fa4c39f5f35e13",
        "trace_state": "{}"
    },
    "kind": "SpanKind.INTERNAL",
    "parent_id": "0x77e577e6a8813bf4",
    "start_time": "2020-05-07T14:39:52.906272Z",
    "end_time": "2020-05-07T14:39:52.906343Z",
    "status": {
        "canonical_code": "OK"
    },
    "attributes": {},
    "events": [],
    "links": []
}

```

Would be logged as:

```

{
 "timestamp": 1588862393906, // 2020-05-07T14:39:52.906272Z as a timestamp
 "traceid": "0xb51058883c02f880111c959f3aa786a2",
 "spanid": "0xb2fa4c39f5f35e13",
 "traceflags": { },
 "severityText": "TRACE3",
 "severityNumber": 3,
 "name": "baz",
 "body": {
         "startTimestamp": "2020-05-07T14:39:52.906272Z",
         "endTimeStamp": "2020-05-07T14:39:52.906343Z",
         "parentId": "0x77e577e6a8813bf4",
       },
 "attributes": {
               "kind": "SpanKind.INTERNAL",
               "status": {"canonical_code: "OK"},
               "attribues": {},
               "traceState": {},
               "events": {},
               "links": {}
  },
  "resource": {}
}

```

## Internal details

Every SDK or API would implement this conversion themselves. This is merely a standard mapping to doing that conversion.

## Trade-offs and mitigations

One drawback is the limited scope of this OTEP in not handling the actual conversion of these fields. This could be mitigated by creating a Metric or Trace conversion library once the Log API/SDK is defined.


## Prior art and alternatives

* It's possible to have the metric definition, value, and label all be inserted into the log attribute or body, however, leaving the body empty except for the metric value will provide better aggregation capabilities.

* The [LogCorrelation](https://github.com/census-instrumentation/opencensus-specs/blob/master/trace/LogCorrelation.md#string-format-for-tracing-data) document in Trace has some advice for converting traces to logs. However, since the Log data model supports the TraceFlags as a bit, it's advice to turn sampling data to "true" or "false" strings, is ignored here. 

* Another suggestion from Issue 398 is to have the logs look like a sys log with some added key value pairs. This sort of output is outside the scope of this OTEP though the log data structure can easily be parsed and printed into this format. For example: `17:05:43 INFO  {sampled=true, spanId=ce751d1ad8be9d11, traceId=ce751d1ad8be9d11} [or.ac.qu.GreetingResource] (executor-thread-1) hello`


## Open questions

Is this mapping enough? Are others needed?

## Future possibilities

Once this Otep is accepted, Otel exporters can produce standarized logs for all metrics and traces increasing compatibility between Otel and reducing confusion.
We can also create further mappings for well known tracing or metric formats from other systems.
