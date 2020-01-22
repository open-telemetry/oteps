# Remove the Metric API Gauge instrument

The [Observer instrument](./0072-metric-observer.md) is semantically
identical to the metric Gauge instrument, only it is reported via a
callback instead of synchronous API calls.  Implementation has shown
that Gauge instruments are difficult to reason about because the
semantics of a "last value" Aggregator have to address questions about
statefulness--the SDK's ability to recall old values.  Observer
instruments avoid some of these concerns because they are reported
once per collection period, making it easier to reason about "all
values" in an aggregator.

## Motivation

Observer instruments improve on our ability to define certain
aggregations, compared with the existing Gauge instrument.  Because of
callbacks, Observer instruments support reporting a complete snapshot
of the "last" values of an instrument, for all active label sets, for
a single collection period.  This is accomplished requiring new state
to be managed by the SDK, simply because the Observer instrument is
expected to support a current value for all active label sets.

## Explanation

Knowing that each collection period contains a complete set of
instrument values, it is possible to define a sum aggregation and to
assign ratios to values.  This aggregation is not possible for Gauge
instruments without imposing new state requirements on the SDK,
essentially requiring it to remember values from prior collection
periods.

## Internal details

The Gauge instrument will be removed from the specification at the
same time the Observer instrument is added.  This will make the
transition easier because in many cases, Observer instruments simply
replace Gauge instruments in the text.

## Trade-offs and mitigations

Not much is lost to the user from removing Gauge instruments.  There
may be situations where it is undesirable to be interrupted by the
Metric SDK in order to execute an Observer callback--situations where
Observer semantics are correct but a synchronous API is more
acceptable.  These cases can be addressed by Observer instruments
paired with helpers to maintain last-value state and define the
"complete" set of values.  This requires more of the user, but users
are likely to be able to optimize this more than the SDK could have,
and this forces the user to recognize that uses of Gauge-like
instruments generally should know the complete set of values when
reporting.

## Prior art and alternatives

Many existing Metric libraries support both synchronous and
asynchronous Gauge-like instruments.

See the initial discussion in [Spec issue
412](https://github.com/open-telemetry/opentelemetry-specification/issues/412).
