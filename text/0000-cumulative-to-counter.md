# Rename "Cumulative" to "Counter" in the metrics API

**Status:** `proposed`

Prefer the name "Counter" as oppose to "Cumulative".

## Motivation

Informally speaking, it seems that OpenTelemetry community members would prefer to call Cumulative metric instruments "Counters".  During conversation (e.g., in the 8/21 working session), this has become clear.

Counter is a noun, like the other kinds Gauge and Measure.  Cumulative is an adjective, so while "Cumulative instrument" makes sense, it describes a "Counter".

## Explanation

This will eliminate the cognitive cost of mapping "counter" to "cumulative" when speaking about these APIs.

## Internal details

Simply replace every "Cumulative" with "Counter", then edit for grammar. 

## Prior art and alternatives

In a survey of existing metrics libraries, Counter is far more common.
