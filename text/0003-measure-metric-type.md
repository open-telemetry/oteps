# Consolidate pre-aggregated and raw metrics APIs

**Status:** `proposed`

# Foreward

This propsal was originally split into three semi-related parts. Based on the feedback, they are now combined here into a single proposal. The original proposals were:

    000x-metric-pre-defined-labels
    000x-metric-measure
    000x-eliminate-stats-record

### Updated 8/27/2019

A working group convened on 8/21/2019 to discuss and debate the two metrics RFCs (0003 and 0004) and several surrounding concerns.  This document has been revised with related updates that were agreed upon during this working session.  See the (meeting notes)[https://docs.google.com/document/d/1d0afxe3J6bQT-I6UbRXeIYNcTIyBQv4axfjKF4yvAPA/edit#].

# Overview

Introduce a `Measure` kind of metric object that supports a `Record` API method.  Like existing `Gauge` and `Cumulative` metrics, the new `Measure` metric supports pre-defined labels.  A new measurement batch API is introduced for recording multiple metric observations simultaneously.

## Terminology

This RFC changes how "Measure" is used in the OpenTelemetry metrics specification.  Before, "Measure" was the name of a series of raw measurements.  After, "Measure" is the kind of a metric object used for recording a series raw measurements.

Since this document will be read in the future after the proposal has been written, uses of the word "current" lead to confusion.  For this document, the term "preceding" refers to the state that was current prior to these changes.

# Motivation

In the preceding `Metric.GetOrCreateTimeSeries` API for Gauges and Cumulatives, the caller obtains a `TimeSeries` handle for repeatedly recording metrics with certain pre-defined label values set.  This enables an important optimization for exporting pre-aggregated metrics, since the implementation is able to compute the aggregate summary "entry" using a pointer or fast table lookup. The efficiency gain requires that the aggregation keys be a subset of the pre-defined labels.

Application programs with long-lived objects and associated Metrics can take advantage of pre-defined labels by computing label values once per object (e.g., in a constructor), rather than once per call site. In this way, the use of pre-defined labels improves the usability of the API as well as makes an important optimization possible to the implementation.

The preceding raw statistics API did not specify support for pre-defined labels.  This RFC replaces the raw statistics API by a new, general-purpose kind of metric, `MeasureMetric`, generally intended for recording individual measurements like the preceding raw statistics API, with explicit support for pre-defined labels.

The preceding raw statistics API supported all-or-none recording for interdependent measurements.  This RFC introduces a `RecordBatch` API to support recording batches of measurements in a single API call, where a `Measurement` is now defined as a tuple of `MeasureMetric`, `Value` (integer or floating point), and `Labels`.

# Explanation

The common use for `MeasureMetric`, like the preceding raw statistics API, is for reporting information about rates and distributions over structured, numerical event data.  Measure metrics are the most general-purpose of metrics.  Informally, the individual metric event has a logical format expressed as one primary key=value (the metric name and a numerical value) and any number of secondary key=values (the labels, resources, and context).

    metric_name=_number_
    pre_defined1=_any_value_
    pre_defined2=_any_value_
    ...
    resource1=_any_value_
    resource2=_any_value_
    ...
    context_tag1=_any_value_
    context_tag2=_any_value_
    ...

Here, "pre_defined" keys are those captured in the metrics handle, "resource" keys are those configured when the SDK was initialized, and "context_tag" keys are those propagated via context.

Events of this form can logically capture a single update to a named metric, whether a cumulative, gauge, or measure kind of metric.  This logical structure defines a _low-level encoding_ of any metric event, across the three kinds of metric.  This establishes the separation between the metrics API and implementation required for OpenTelemetry.  An SDK could simply encode a stream of these events and the consumer, provided access to the metric definition, should be able to interpret these events according to the semantics prescribed for each kind of metric.

## Metrics API concepts

The `Meter` interface represents the metrics portion of the OpenTelemetry API.

There are three kinds of metric, `CumulativeMetric`, `GaugeMetric`, and `MeasureMetric`.

Metric objects are declared and defined independently of the SDK. They may be statically defined, as opposed to allocated through the SDK in any way.  To define a new metric, use one of the language-specific API methods (e.g., with names like `NewCumulativeMetric`, `NewGaugeMetric`, or `NewMeasureMetric`).

Each metric is declared with a list (possibly empty) of pre-defined label keys.  These pre-defined label keys declare the set of keys that are available as dimensions for efficient pre-aggregation.

To obtain a metric _handle_ from a metric object, call `getHandle` with the pre-defined label values.  There are two ways to pass the pre-defined label values:

1. As an ordered list of values.  In this case, the number of arguments in the list must match the number of pre-defined label keys.  When the number of arguments disagrees with the metric definition, the implementation may return an error or thrown an exception to synchronously indicate this condition.
2. As a list of key:value pairs.  In this case, the application is free to provide the list of label values in arbitrary order.  Values that are not passed when constructing handles in this way are marked as "not present".  Values that are not part of the pre-defined label keys are ignored when constructing handles.

Metric handles, thusly obtained with one of the `getHandle` variations, may be used to `Set()`, `Add()`, and `Record()` metrics according to their kind.  Context tags that apply when calling `Set()`, `Add()`, and `Record()` may not override values that were set in the handle as pre-defined labels.

## Selecting Metric Kind

By the "separation clause" of OpenTelemetry, we know that an implementation is free to do _anything_ in response to a metric API call.  By the low-level interpretation defined above, all metric events have the same structural representation, only their logical interpretation varies according to the metric definition.  Therefore, we select metric kinds based on two primary concerns:

1. What should be the default implementation behavior?  Unless configured otherwise, how should the implementation treat this metric variable?
1. How will the program read?  Each metric uses a different verb, which helps convey meaning and describe default behavior.  Cumulatives have an `Add()` method.  Gauges have a `Set()` method.  Measures have a `Record()` method.

To guide the user in selecting the right kind of metric for an application, we'll consider the following questions about the primary intent of reporting given data.  We use "of primary interest" here to mean information that is almost certainly useful in understanding system behavior.  Consider these questions:

- Does the measurement represent a quantity of something?  Is it also non-negative?
- Is the sum a matter of primary interest?
- Is the event count of primary interest?
- Is the distribution (p50, p99, etc.) a matter of primary interest?

The specification will be updated with the following guidance.

### Cumulative metric

Likely to be the most common kind of metric, cumulative metric events express the computation of a sum.  Choose this kind of metric when the value is a quantity, the sum is of primary interest, and the event count and distribution are not of primary interest.  To raise (or lower) a cumulative metric, call the `Add()` method.

If the quantity in question is always non-negative, it implies that the sum is strictly ascending.  When this is the case, the cumulative metric also serves to define a rate.  For this reason, cumulative metrics have an option to be declared as non-negative.  The API will reject negative updates to non-negative cumulative metrics, instead submitting an SDK error event, which helps ensure meaningful rate calculations.

For cumulative metrics, the default OpenTelemetry implementation exports the sum of event values taken over an interval of time.

### Gauge metric

Gauge metrics express a pre-calculated value that is either `Set()` by explicit instrumentation or observed through a callback.  Generally, this kind of metric should be used when the metric cannot be expressed as a sum or a rate because the measurement interval is arbitrary.  Use this kind of metric when the measurement is not a quantity, and the sum and event count are not of interest.

Only the gauge kind of metric supports observing the metric via a callback (as an option).  Semantically, there is an important difference between explicitly setting a gauge and observing it through a callback.  In case of setting the gauge explicitly, the call happens inside of an implicit or explicit context.  The implementation is free to associate the explicit `Set()` event with a context, for example.  When observing gauge metrics via a callback, there is no context associated with the event.

As a special case, to support existing metrics infrastructure, a gauge metric may be declared as a precomputed sum, in which case it is defined as strictly ascending.  The API will reject descending updates to strictly-ascending gauges, instead submitting an SDK error event.

For gauge metrics, the default OpenTelemetry implementation exports the last value that was `Set()`.  If configured for an observer callback instead, the default OpenTelemetry implementation exports  Observed at the time of metrics collection.

### Measure metric

Measure metrics express a distribution of values.  This kind of metric should be used when the count of events is meaningful and either:

1. The sum is of interest in addition to the count
1. Quantiles information is of interest.

The key property of a measure metric event is that two events cannot be trivially reduced into one, as a step in pre-aggregation.  For cumulatives and gauges, two `Add()` or `Set()` events can be replaced by a single event (for default behavior, i.e., unless the implementation is configured differently), whereas two `Record()` events must by reflected in two events. 

Like cumulative metrics, non-negative measures are an important case because they support rate calculations. As an option, measure metrics may be declared as non-negative.  The API will reject negative metric events for non-negative measures, instead submitting an SDK error event.

For measure metrics, the default OpenTelemetry implementation is left up to the implementation. The default interpretation is that the distribution should be summarized, somehow, but the specific technique used belongs to the implementation.  A low-cost policy is selected as the default behavior for export OpenTelemetry measures:

- For non-negative measure metrics, unless otherwise configured, the default implementation exports the sum, the count, and the maximum value as three separate summary variables.
- For arbitrary measure metrics, unless otherwise configured, the default implementation exports the sum, the count, the minimum, and the maximum value as four separate summary variables.

### Disable selected metrics by default

All OpenTelemetry metrics may be disabled by default, as an option.  Use this option to indicate that the default implementation should be to do nothing for events about this metric.

### RecordBatch API

Applications sometimes want to record multiple metrics in a single API call, either becase the values are inter-related or because it lowers overhead.  We agree that recording batch measurements will be restricted to measure metrics, although this support could be extended to all kinds of metric in the future.

Logically, a measurement is defined as:

- Measure metric: which metric is being updated
- Value: a floating point or integer
- Pre-defined label values: associated via metrics API handle

The batch measurement API shall be named `RecordBatch`.  The entire batch of measurements takes place within some (implicit or explicit) context.

## Prior art and alternatives

Prometheus supports the notion of vector metrics, which are those that support pre-defined labels.  The vector-metric API supports a variety of methods like `WithLabelValues` to associate labels with a metric handle, similar to `GetOrCreateTimeSeries` in OpenTelemetry.  As in this proposal, Prometheus supports a vector API for all metric types.

Statsd libraries generally report metric events individually.  To implement statsd reporting from the OpenTelemetry, a `Meter` SDK would be installed that converts metric events into statsd updates.

## Open questions

Argument ordering has been proposed as the way to pass pre-defined label values in `GetOrCreateTimeseries`.  The argument list must match the parameter list exactly, and if it doesn't we generally find out at runtime or not at all.  This model has more optimization potential, but is easier to misuse, than the alternative.  The alternative approach is to always pass label:value pairs to `GetOrCreateTimeseries`, as opposed to an ordered list of values. 

The same discussion can be had for the `MeasurementBatch` type described here.  It can be declared with an ordered list of metrics, then the `Record` API takes only an ordered list of numbers.  Alternatively, and less prone to misuse, the `MeasurementBatch.Record` API could be declared with a list of metric:number pairs.
