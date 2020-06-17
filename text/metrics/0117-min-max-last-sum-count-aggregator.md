# Replace MinMaxSumCount aggregator with MinMaxLastSumCount aggregator

Add a LastValue component to the MinMaxSumCount aggregator, name it MinMaxLastSumCount.

## Motivation

The MinMaxSumCount aggregator is specified as the default for grouping
instruments (i.e., ValueRecorder and ValueObserver).  This default
makes it impossible with these instruments to expose a Gauge
containing the last value in Prometheus and Statsd, yet a last-value
Gauge is the natural low-cost exposition format for these instruments.

This proposal adds the LastValue to the Min, Max, Sum, and Count
fields and renames the combined aggregator MinMaxLastSumCount.  Using
this default, we are able to export the default aggregation over OTLP
and expose the last value for Prometheus and Statsd exporters.

## Explanation

The Min, Max, Sum, and Count were selected as the default aggregation
for GROUPING instruments because they are an inexpensive, mergeable
summary of the value distribution and can be conveyed exactly using a
fixed amount of space.  Any quantile value other than the Min and Max
does not meet these requirements, either it will be inexact or require
non-constant space.

LastValue was not included because it does significantly add
information about the distribution of values.  LastValue could be used
to establish a tendency in the data (e.g., if the last value is
greater than the average value, the signal is rising), but it could
just be an outlier.  Knowing the _value_ part a LastValue may not tell
us much about the distribution of values, but knowing the _timestamp_
of the last value (i.e., the maximum timestamp) adds to our
understanding of the timeseries.

Adding the LastValue to MinMaxSumCount is partly pragmatic.  This
makes it possible to use the default aggregator for grouping
instruments, send data through OTLP, and expose the LastValue as a
Gauge without loss of the information commonly used by other exporters
(i.e., Min, Max, Sum, Count).

## Internal details

This proposal covers how the data should be represented in OTLP.

Prior to this proposal, the MinMaxSumCount value has been exposed as a
`Summary` value, which is a general-purpose value type meant for
conveying quantile information.  The `Summary` value type is not
mergeable in general, so it is problematic to use it with
MinMaxSumCount aggregations, which are mergeable.  In general we do
not want MinMaxSumCount values to be exposed as Prometheus Summary
values by default, although this is the current behavior in the
OpenTelemetry Collector.

An OTLP value type for the new MinMaxLastSumCount aggregation will be
added to the `Metric` as follows:


```
message Metric {
  // [descriptor and 4 existing value types fields]

  // MinMaxLastSumCount values go here:
  repeated MinMaxLastSumCountDataPoint minmaxlastsumcount_data_points = 6;
}

// Defines a metric type and its schema.
message MetricDescriptor {

  // [name, description, and unit fields]

  // Type is the type of values a metric has.
  enum Type {
    // [ ... 7 existing values ]

    // MINMAXLASTSUMCOUNT_INT64 indicates a MinMaxSumCountDataPoint of int64 values.
    MINMAXLASTSUMCOUNT_INT64 = 7;

    // MINMAXLASTSUMCOUNT_DOUBLE indicates a MinMaxSumCountDataPoint of double values.
    MINMAXLASTSUMCOUNT_DOUBLE = 8;
  }

  // [ ... ]
}

// MinMaxLastSumCountDataPoint is a mergeable summary of a timeseries of individual values.
//
// This type contains an important optimization to avoid sending min, max, sum, and count
// values when there was a single value in the stream with count equal to 1.0, in which
// case only the last-value field is needed.
message MinMaxSumCountDataPoint {
  // The set of labels that uniquely identify this timeseries.
  repeated opentelemetry.proto.common.v1.StringKeyValue labels = 1;

  // start_time_unix_nano is the time when the cumulative value was reset to zero.
  // This is used for Counter type only. For Gauge the value is not specified and
  // defaults to 0.
  //
  // The cumulative value is over the time interval (start_time_unix_nano, time_unix_nano].
  // Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970.
  //
  // Value of 0 indicates that the timestamp is unspecified. In that case the timestamp
  // may be decided by the backend.
  fixed64 start_time_unix_nano = 2;

  // time_unix_nano is the moment when this value was recorded.
  // Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970.
  fixed64 time_unix_nano = 3;

  // last_time_unix_nano is the moment when the latest value was captured during
  // the interval described by [start_time_unix_nano, time_unix_nano].
  fixed64 last_time_unix_nano = 4;

  // last_int64 is the last value as an integer.  This field MUST be set when the 
  // metric descriptor type is MINMAXLASTSUMCOUNT_INT64.
  int64 last_int64 = 5;

  // min_int64 is the minimum value as an integer, when set.  If this field is
  // not set and last_int64 is set, the minimum value equals the value of last_int64.
  int64 min_int64 = 6;

  // max_int64 is the maximum value as an integer, when set.  If this field is
  // not set and last_int64 is set, the maximum value equals the value of last_int64.
  int64 max_int64 = 7;

  // last_double is the last value as an integer.  This field MUST be set when the 
  // metric descriptor type is MINMAXLASTSUMCOUNT_DOUBLE.
  double last_double = 8;

  // min_double is the minimum value as an integer, when set.  If this field is
  // not set and last_double is set, the minimum value equals the value of last_double.
  double min_double = 9;

  // max_double is the maximum value as an integer, when set.  If this field is
  // not set and last_double is set, the maximum value equals the value of last_double.
  double max_double = 10;

  // sum is the sum of individual values captured during the interval
  // described by [start_time_unix_nano, time_unix_nano].  This is a floating
  // point value even when the metric descriptor type is MINMAXLASTSUMCOUNT_INT64
  // because it may derived from sample data with a non-integer multiplier.
  //
  // When this value is not set, it may be presumed to equal the value of
  // last_int64 or last_double, whichever field is set, times the value of count.
  double sum = 11;

  // count is the number of individual values captured during the interval
  // described by [start_time_unix_nano, time_unix_nano].  this is a floating
  // point value because it may derived from sample data with a non-integer
  // multiplier.
  //
  // This MUST be omitted when the value is exactly 1.0.  When this value
  // is not set, it MUST be assumed to equal 1.0.
  double count = 12;
}
```

This message type is specified so that:

- the count can be omitted when it equals 1.0
- the minimum and maximum can be omitted when they are equal to the last value
- the sum value can omitted when it equals the last value times the count.

The OpenTelemetry specification will be updated to specify that
MinMaxLastSumCount is the default aggregation for grouping instruments
(i.e., `ValueRecorder` and `ValueObserver`).

## Trade-offs and mitigations

This proposal adds somewhat to the computational cost of computing
this aggregation at runtime.  This is not expected to be a problem,
because both the existing MinMaxSumCount and LastValue aggregators are
already relatively inexpensive.  Capturing an extra field and carrying
through the export pipeline is expected to be negligible.

This proposal adds a new value type to OTLP, which will require work
to implement.  However, this type addresses both a theoretical need
(that Summaries are not mergable) and a practical need (need to
preserve Gauge semantics of Prometheus/Statsd).

This adds to the encoded size of OTLP, maybe or maybe not.  Existing
exporters for OTLP were using the Summary value type, which is an
inefficient way to express minimum and maximum values.

## Prior art and alternatives

When a non-OTLP exporter is configured locally for a Gauge exporter
like Prometheus or Statsd, it should be possible to downgrade from a
MinMaxLastSumCount aggregator to a LastValue aggregator automatically.

Alternatives have been discussed in an [opentelemetry-specification
issue](https://github.com/open-telemetry/opentelemetry-specification/issues/636).
