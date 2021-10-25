# Export SpanContext.IsRemote in OTLP

Update OTLP to indicate whether a span's parent is remote.

## Motivation

It is sometimes useful to post-process or visualise only entry-point spans: spans which either have no parent (trace roots), or which have a remote parent.
For example, the Elastic APM solution highlights entry-point spans (Elastic APM refers to these as "transactions") and surfaces these as top-level operations
in its user interface.

Currently, the only way entry-point spans can be identified is using (lack of) parent ID, and span kind. Relying on span kind can lead to invalid assumptions,
particularly with relation to `CONSUMER` messaging spans. Using the [batch receiving example](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/messaging.md#batch-receiving)
in the messaging semantic conventions, `Span Recv1` should be the only entry point into `Process C`. If we assume `CONSUMER` spans are always entry-point spans,
then this leads to `Span Proc1` and `Span Proc2` being incorrectly classified as entry-point spans. For messaging spans we might also take into account the
`messaging.operation` attribute to tell these apart, however `messaging.operation` is not required; and this would not satisfy other scenarios such as actively
polling a message queue, which would result in a `CONSUMER` span which has a non-remote parent span.

## Explanation

The OTLP encoding for spans has a boolean `parent_span_is_remote` field for identifying whether a span's parent is remote or not.
All OpenTelemetry SDKs populate this field, and backends may use it to identify a span as being an entry-point span.
A span can be considered an entry-point span if it has no parent (`parent_span_id` is empty), or if `parent_span_is_remote` is true.

## Internal details

The first part would be to update the trace protobuf, adding a `boolean parent_span_is_remote` field to the
[`Span` message](https://github.com/open-telemetry/opentelemetry-proto/blob/b43e9b18b76abf3ee040164b55b9c355217151f3/opentelemetry/proto/trace/v1/trace.proto#L84).

OpenTelemetry SDKs already track whether a span's parent is remote in [`SpanContext`](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/api.md#isremote).
The OTLP exporter in each SDK would need to be updated to record this in the new `parent_span_is_remote` field.

For backwards compatibility with older OTLP versions, the protobuf field should be `nullable` (`true`, `false`, or unspecified)
and the opentelemetry-collector protogen code should provide an API that enables backend exporters to identify whether the field is set.

```go
package pdata

// ParentSpanIsRemote indicates whether ms's parent span is remote, if known.
// If the parent span remoteness property is known then the "ok" result will be true,
// and false otherwise.
func (ms Span) ParentSpanIsRemote() (remote bool, ok bool)
```

## Trade-offs and mitigations

None identified.

## Prior art and alternatives

### Alternative 1: include entry-point span ID in other spans

As an alternative to identifying whether the parent span is remote, we could instead encode and propagate the ID of the entry-point span in all non entry-point spans.
Thus we can identify entry-point spans by lack of this field.

The entry-point span ID would be captured when starting a span with a remote parent, and propagated through `SpanContext`. We would introduce a new `entry_span_id` field to
the `Span` protobuf message definition, and set it in OTLP exporters.

This was originally [proposed in OpenCensus](https://github.com/census-instrumentation/opencensus-specs/issues/229) with no resolution.

The drawbacks of this alternative are:

- `SpanContext` would need to be extended to include the entry-point span ID; SDKs would need to be updated to capture and propagate it
- The additional protobuf field would be an additional 8 bytes, vs 1 byte for the boolean field

The main benefit of this approach is that it additionally enables backends to group spans by their process subgraph.

## Open questions

None.

## Future possibilities

No other future changes identified.
