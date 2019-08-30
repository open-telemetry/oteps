# Rename distributed context into tags and add support for Baggage.

**Status:** `proposed`

Rename Distributed-context to a less confusing term and restrict its use for observability.
 
 
## Motivation

Context Propagation is generalized to carry observability and non-observability contexts. This
includes SpanContext, Distributed-context, Authentication-context, etc. Distributed-context
carries key-value pairs for observability. The name Distributed-context is
very confusing. Hence it should be changed to TagMap (or other better suggestion). 
That is the first motivation.
 
The second motivation is the restrict use of Distributed-context to observability. As it is
written currently it does not specifically call out for using it only for non-observability purpose.
However, it can be used to support 'Baggage' functionality for non-observability purpose as it is
OpenTracing.
There are few reasons to have different mechanism to address observability and
non-observability use cases.
1. For observability, key-value pairs SHOULD be write-only from application perspective. SDKs can read them
   for example to apply them to measurements at the time of recording. For non-observability, 
   key-value pairs SHOULD be writable/readable for application.
2. For non-observability, it may require that the delivery is guaranteed and pairs are not dropped.
3. Semantically two are different and mechanically it could be different as well.

## Explanation

We do following to
- Rename distributed-context to TagMap.
  - It is strictly used for observability.
  - It is a write-only object from application and framework-integrations perspective.
  - It is readable by SDKs.

- Add Baggage support. Equivalent of what is there in [OpenTracing](https://opentracing.io/docs/overview/tags-logs-baggage/#baggage-items).
  - It is more of a convenient functionality to provide compatibility with Open Tracing.
  - It is propagated downstream but may require separate propagator (different from one used for 
 SpanContext and TagMap
  - It is readable by application. **Open question**: should it be mutable or immutable?
 
## Trade-offs and mitigations

There may be cases where a key-value pair is propagated as TagMap for observability and as a Baggage for 
application specific use. AB testing is one example of such use case. There is a duplication here at
call site where a pair is created and also at propagation. Solving this issue is not worth having 
semantic confusion with dual purpose.

 
## References

[OpenTracing Baggage](https://opentracing.io/docs/overview/tags-logs-baggage/#baggage-items)

[OpenCensus Tag](https://github.com/census-instrumentation/opencensus-specs/blob/master/tags/TagMap.md)

## Open questions



