# SDK Proposal: Default and Configurable Aggregations for Instruments

This proposal builds off the work of John in
[OTEP-126](https://github.com/open-telemetry/oteps/pull/126) and Chris in
[OTEP-89](https://github.com/open-telemetry/oteps/pull/89).

The unique part of this proposal is that all metric exports are based on
views. Each `Instrument` has a default `View` created when the `Instrument` is
created.

## Definitions

* Measurement: The value recorded along with a set of label keys and values.
* Instrument: The thing a `Measurement` is associated with when recorded. Each
  has a name, description, instrument kind and unit.
* Aggregation: Defines how a set of measurements for an instrument are to be
  combined for export.
* View: An aggregation on an `Instrument` and some metadata (a name,
  description, etc).

## View

* name: The name used when exporting metrics. 
* description: String value describing what the aggregated measurements represent.
* aggregation: The implementation of combining measurements to use.
* labels: The labels to filter down to from the labels included in the measurement.
* exporter: Each `view` can be associated with an exporter. By default it is set
  to the default SDK exporter, if it is overridden as some form of `undefined`
  value (depending on the language) this `view` is not exported and does not
  need to be updated when a `measurement` is recorded for an `instrument`. This
  leaves open the possibility of more complex export piplines in the SDK, as
  opposed to simply being a boolean value.

### Default Views

An `Instrument` is more than simply the metadata for a metric that
will be recorded. `Instruments` have default `views` defined for them. This
eases the use by an end user as opposed to OpenCensus where a `View` was
required to be created by the user for any measurement the user wants exported.

The default view for an instrument uses the same name and description as the
instrument and the default aggregation defined for the kind of instrument it
is. The table in the [Metrics API
specification](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/metrics/api.md#interpretation)
provides the default aggregation per instrument kind:

| **Name** | Instrument kind | Function(argument) | Default aggregation | Notes |
| ----------------------- | ----- | --------- | ------------- | --- |
| **Counter**             | Synchronous additive monotonic | Add(increment) | Sum | Per-request, part of a monotonic sum |
| **UpDownCounter**       | Synchronous additive | Add(increment) | Sum | Per-request, part of a non-monotonic sum |
| **ValueRecorder**       | Synchronous  | Record(value) | MinMaxSumCount  | Per-request, any non-additive measurement |
| **SumObserver**         | Asynchronous additive monotonic | Observe(sum) | Sum | Per-interval, reporting a monotonic sum |
| **UpDownSumObserver**   | Asynchronous additive | Observe(sum) | Sum | Per-interval, reporting a non-monotonic sum |
| **ValueObserver**       | Asynchronous | Observe(value) | MinMaxSumCount  | Per-interval, any non-additive measurement |

### Overriding Default Aggregation

Through the SDK the aggregation used by a kind of instrument can be overridden
for all instruments of that kind or by name and kind of the instrument:

```
\\ for all instruments of a specific kind
Meter.set_default_aggregation(ValueRecorder, TDigest)

\\ for a specific instrument
Meter.set_default_aggregation("http.latency", ValueRecorder, Distribution{Buckets=[0,100,200,500,1000]})
```

### Additional Views

An instrument can have multiple views associated with it:

```
// Meter.new_view(InstrumentName, InstrumentKind, ViewName, ViewDescription, Aggregation, Labels)
Meter.register_view(InstrumentName, InstrumentKind, "http.latency_histogram", "Histogram of HTTP request
latencies", HdrHistogram, [HttpMethod])
```

When `ViewName` is the same as `InstrumentName` it acts the same as
`set_default_aggregation` but with a new description and a set of label keys to
filter down to. 

## Open Questions

- How is aggregation output configured to be deltas or cumulative? Is this an
  exporter configuration that each view inherits in order to know how to
  aggregate for the particular exporter?

