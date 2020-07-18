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
values by default, although this does not preclude using the
SummaryDataPoint struct in the current draft of OTLPK.

This proposal is summarized by the following change to
SummaryDataPoint.


```
index 1d45882..ace9d20 100644
--- a/opentelemetry/proto/metrics/v1/metrics.proto
+++ b/opentelemetry/proto/metrics/v1/metrics.proto
@@ -403,14 +403,38 @@ message SummaryDataPoint {
   // Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970.
   fixed64 time_unix_nano = 3;
 
-  // The total number of recorded values since start_time. Optional since
-  // some systems don't expose this.
+  // The total number of recorded values since start_time.  This field
+  // MUST be omitted when its true value is 1.  An omitted or zero
+  // value is interpreted as having `count` equal to 1, and in this
+  // case the value of `sum`, `min`, and `max` are implied to equal
+  // the value of `last`.
   uint64 count = 4;
 
-  // The total sum of recorded values since start_time. Optional since some
-  // systems don't expose this. If count is zero then this field must be zero.
+  // The last value MUST be set to the last value that was measured
+  // since start_time.
+  double last = 7;
+
+  // The total sum of recorded values since start_time.  If the count
+  // was omitted or zero because its true value was 1, this field
+  // shall be omitted by the writer and the value of `sum` shall be
+  // set by the reader to value of `last`.  Otherwise, this field MUST
+  // be set to the sum of measurements reflected in count.
   double sum = 5;
 
+  // The minimum of recorded values since start_time.  If the count
+  // was omitted or zero because its true value was 1, this field
+  // shall be omitted by the writer and the value of `min` shall be
+  // set by the reader to value of `last`.  Otherwise, this field MUST
+  // be set to the minimum of the measurements reflected in count.
+  double min = 8;
+
+  // The maximum of recorded values since start_time.  If the count
+  // was omitted or zero because its true value was 1, this field
+  // shall be omitted by the writer and the value of `max` shall be
+  // set by the reader to value of `last`.  Otherwise, this field MUST
+  // be set to the maximum of the measurements reflected in count.
+  double max = 9;
+
   // Represents the value at a given percentile of a distribution.
   //
   // To record Min and Max values following conventions are used:
```

This message type is specified so that:

- the count can be omitted when it equals 1
- the sum, minimum, and maximum MUST be omitted when count is also emitted.

This leaves a single scalar value in the encoding when a single value
was aggregated.

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
