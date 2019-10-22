# Track Current Tracer in Context

Add the requirement that the tracer used to create a span be tracked in the current context.

## Motivation

Without explicitly passing a `Tracer` variable to methods and functions any spans created do not know what tracer was used to create their parent. If tracers are meant to provide separate functionality and configuration then choosing a tracer at, for example, the root of an HTTP request is possibly the tracer the user wants to be used by any library the handler calls out to, such as a database library, without requiring the database library to modify its API to accept a tracer as an argument.

## Explanation

The tracer to start a span with is first fetched from a registry of tracers or newly created. If no explicit tracer is fetched, the tracer that is configured as the default is returned unless there is an active span in which case the tracer used to create that span is returned.

## Internal details

When setting the current span in the current context the tracer should also be stored. Fetching a tracer first checks the current context for the tracer to return. If there is no active span and no default tracer is configured and the SDK is available then the SDK tracer is returned for the default tracer. If there is no active span and no default tracer is configured and the SDK is not available then the API minimal tracer is returned by default.

## Trade-offs and mitigations

It is yet another thing to track in context. It could be optional.

## Prior art and alternatives

The alternative is passing the tracer separately or not being able to ensure the parent tracer is used by a child.

## Open questions

What does this mean for named tracers? Would it be the "Tracer Factory" that is stored in the context and named tracers need to somehow be composable with other tracers? For example, if this proposal were accepted and children created of a `NoopTracer` span were then expected to also use the `NoopTracer` by default could you create a named `NoopTracer`?

## Future possibilities

Haven't thought of anything yet.
