# Metric `LabelSet` specification

**Status:** `proposed`

Introduce a first-class `LabelSet` API type as a handle on a pre-defined set of labels.  

## Motivation

Labels are key-value pairs used across OpenTelemetry for categorizing spans (attributes, event fields) and metrics (labels).  Treatment of labels in the metrics API is especially important because of certain optimizations that take advantage of pre-defined labels.

This optimization currently applies to metric instrument handles, the result of `GetHandle(Instrument, { Key : Value, ... })` on a metric instrument. By allowing the SDK to pre-compute information about the pair of `(Instrument, { Key : Value, ... })`, this reasoning goes, the metrics SDK has the opportunity make individual `Add()` and `Set()` operations on handles as fast as a few machine instructions.

Adopting a first-class `LabelSet` extends the potential for optimization, significantly, for cases where a the `{ Key : Value, ... }` can be reused despite not re-using  `(Instrument, { Key : Value, ... })` pairs.

The current API specifies one way to use set of labels for more than one measurement,

```
RecordBatch({ Key1: Value1,
  	      Key2: Value2,
	      ... },
	    [ (Gauge, Value), 
	      (Cumulative, Value),
	      (Measure, Value),
	      ... ])
```

This RFC proposes the new `LabelSet` concept is returned by an API named `Meter.DefineLabels({ Key: Value, ... })` which allows the SDK to potentially sort, canonicalize, or hash the set of labels once, allowing it to be re-used in several ways.  For example, the `RecordBatch` above can be written as one call to `DefineLabels` and inidividual operations.

```
Labels := meter.DefineLabels({ Key1: Value1,
       	  		       Key2: Value2,
			       ... })
Cumulative.Add(Value, Labels)
Gauge.Set(Value, Labels)
Measure.Record(Value, Labels)
...
```

With a first-class `LabelSet`, labels can even be re-used across multiple calls to `RecordBatch`.

## Explanation

Metric instrument APIs which presently take labels in the form `{ Key: Value, ... }` will be updated to take an explicit `LabelSet`.  The `Meter.DefineLabels` API method supports getting a `LabelSet` from the SDK, allowing the programmer to pre-define labels without being required to manage handles.  This brings the number of ways to update a metric to three, via re-using a handle from `GetHandle()`:

```
cumulative := metric.NewFloat64Cumulative("my_counter", [
					    "required_key1",
					    "required_key2",
					  ])
labels := meter.DefineLabels({ "required_key1": value1,
			       "required_key2": value2 })
handle := cumulative.GetHandle(labels)
for ... {
   handle.Add(quantity)
}
```

To operate on a metric instrument directly, without requiring a handle:

```
cumulative.Add(quantity, labels)
gauge.Set(quantity, labels)
measure.Record(quantity, labels)
```

To operate on a batch of labels,

```
RecordBatch(labels, [
	(Instrument1, Value1),
	(Instrument2, Value2),
	...
      ])
```

Note that `meter.GetHandle(instrument, meter.DefineLabels())` is the same as `meter.GetDefaultHandle()`.

### Ordered `LabelSet` option

OpenCensus specified support for ordered label sets, in which values for required keys may be passed in order, allowing the implementation directly compute a unique encoding for the label set.   We recognize that providing ordered values is an option that makes sense in some languages and not others, and leave this as an option.  Support for ordered label sets might be arranged through ordered key sets, for example:

```
requiredKeys := metric.DefineKeys("key1", "key2", "key3")
labelSet := requiredKeys.DefineValues(meter, 1, 2, 3)
```

## Internal details

Metric instruments are specified as SDK-independent objects, therefore metric handles were required to bind the instrument to the SDK in order to operate. In this proposal, `LabelSet` becomes responsible for binding the SDK at any call site where it is used.  Other than knowing the `Meter` that defined it, `LabelSet` is an opaque interface.  The three ways to use `LabelSet` in the metrics API are:

```
instrument.GetHandle(labels).Action(value)       // Action on a handle 
instrument.Action(value, labels)                 // Single action, no handle
RecordBatch(labels, [(instrument, value), ...])  // Batch action, no handles
```

## Trade-offs and mitigations

Each programming language should select the names for `LabelSet` and `DefineLabels` that are most idiomatic and sensible.

In languages where overloading is a standard convenience, the metrics API may elect to offer alternate forms that elide the call to `DefineLabels`, for example this:

```
instrument.GetHandle(meter, { Key: Value, ... })
```

as opposed to this:

```
instrument.GetHandle(meter.DefineLabels({ Key: Value, ... }))
```

A key distinction between `LabelSet` and similar concepts in existing metrics libraries is that it is a _write-only_ structure, allowing the programmer to set diagnostic state while ensuring that diagnostic state does not become application-level state.

## Prior art and alternatives

There is not clear prior art like `LabelSet` as defined here.

A potential application for `DefineLabels` is to pre-compute the bytes of the statsd encoding for a label set once, to avoid repeatedly serializing this information.

## Open questions

Introducing `LabelSet` makes one more step for simply using the metrics API.  Can convenience libraries, utility classes, and overloaded functions make the simple API use-cases acceptable while supporting re-use of `LabelSet` objects for optimization?
