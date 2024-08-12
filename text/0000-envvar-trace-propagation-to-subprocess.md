# Trace Context Propagation to Subprocess using Environment Variables

Standarize usage of environment variables for propagating trace context to spawned subprocesses.
This includes execution of a different application, a new instance of the same application, or
`fork()`ing the current application.

## Motivation

There has been a long-standing want from the community to be able to propagate tracing context to
short-lived subprocesses that are spawned by its parent to perform single units of work.
([optenelemetry-specification issue #740](https://github.com/open-telemetry/opentelemetry-specification/issues/740))
This is of a particular interest to the [CI/CD Observability Semantic Conventions WG](https://github.com/open-telemetry/community/blob/main/projects/ci-cd.md).

Ad-hoc implementations enabling such propagation are already in use by many applications, and there
is existing alignment on the preferred approach.  This OTEP aims to precicsely define and standarize
on how trace context can be propagated to subprocesses using environment variables.

## Explanation

### Propagation via environment variables

The environment variables `TRACEPARENT` and `TRACESTATE` are used to propagate trace context to
subprocesses that are spwaned.  Formatting and semantics of their values are according to
[W3C's Trace Context Recommendation](https://www.w3.org/TR/trace-context/)'s `traceparent` and
`tracestate` headers, respectively, with a single modification that `TRACESTATE` does not have
provisions for being split into multiple environment variables.

### Value format

The Augented Backus-Naur Form (ABNF) of valid `TRACEPARENT` and `TRACESTATE` environment variable
values are reproduced from W3C Trace Context document, below, with changes to top-level rule names
for clarity:

```abnf
traceparent = version "-" version-format
version     = 2HEXDIGLC   ; this document assumes version 00. Version ff is forbidden

version-format   = trace-id "-" parent-id "-" trace-flags
trace-id         = 32HEXDIGLC  ; 16 bytes array identifier. All zeroes forbidden
parent-id        = 16HEXDIGLC  ; 8 bytes array identifier. All zeroes forbidden
trace-flags      = 2HEXDIGLC   ; 8 bit flags. Currently, only one bit is used. See below for details

;---

tracestate  = list-member 0*31( OWS "," OWS list-member )
list-member = (key "=" value) / OWS

key = simple-key / multi-tenant-key
simple-key = lcalpha 0*255( lcalpha / DIGIT / "_" / "-"/ "*" / "/" )
multi-tenant-key = tenant-id "@" system-id
tenant-id = ( lcalpha / DIGIT ) 0*240( lcalpha / DIGIT / "_" / "-"/ "*" / "/" )
system-id = lcalpha 0*13( lcalpha / DIGIT / "_" / "-"/ "*" / "/" )
lcalpha    = %x61-7A ; a-z

value    = 0*255(chr) nblk-chr
nblk-chr = %x21-2B / %x2D-3C / %x3E-7E
chr      = %x20 / nblk-chr

;---

HEXDIGLC = DIGIT / "a" / "b" / "c" / "d" / "e" / "f" ; lowercase hex character
```

## Internal details

### Environment variable name case sensitivty

Environment variable names are case-insensitive on Windows, while case sensitive elsewhere.
For interoperability, implementations _MUST_ use the uppercase `TRACEPARENT` and `TRACESTATE`
environment variable names when injecting for propagation.

For extraction, implementations _MAY_ also accept environment variable names with different casing
than above, (`traceparnet,`, `Tracestate`, etc), but prioritize the variant with only upper-case
characters in its name.  Otherwise, conflict resolution between multiple elgible environment
variables is implementation defined.

### Implementation using `TextMap` propagator

Using W3C's Trace Context semantic allows any existing [`TextMap` proapgator](https://opentelemetry.io/docs/specs/otel/context/api-propagators/#textmap-propagator)
to be used to serialize and deserialize `TRACEPARENT` and `TRACESTATE`.

One potential implementaiton is to provide `Getter` and `Setter` to `TextMap`'s `Extract` and
`Inject`, respectively, that perform appropriate operation on a `Carrier` capable of
interfacing with the language's environment variable facilities.  However, care should be taken to
limit their usage to a `TextMap` propagator specificalized for handling tracing context, or limit
the `Getter` and `Setter` to only interact with `TRACEPARENT` and `TRACESTATE`.  An unrestricted
propagator can set arbitrary environment variables, resulting in unintended side effects.

### Span kind

The parent process _SHOULD_ create a span of `CLIENT` or `PRODUCER` kind to cover the subprocess
launching process.  The spawned child process _SHOULD_ have a corresponding `SERVER` or `CONSUMER`
span covering its entire execution lifetime.

This OTEP provides no recommendations on if and when the child process should adopt its parent's
Trace ID as its own, or to create a new trace that is correlated using a span link.  Applications
should choose what is appropriate for their specific context, and instrumentation implementations
_SHOULD_ provide API to enable this choice.

### Limiting unintended propagation

In many language runtimes, the parent process' environment variables are wholly inherited by
spanwed child processes.  This can lead to unintended and inappropriate proapgation of the parent's
tracing context.

Implementations that use environemnt variables for proapgation to spawned child processes _SHOULD_
limit the injection of these variables to the spawning mechanism itself, and avoid modifying its own
environment variables as a way to pass them onto its children.

Applications that intend to consume `TRACEPARENT` and `TRACESTATE` _SHOULD_ consider clearing
their values after extracting the context. This eliminates the possibiilty of unintended
propagation if environment variable inheritence is not carefully controlled throughout the
application.

### Configuring subprocess' exporter endpoint

It may be necessary for the parent process to inform a spawned subprocess on where to export its
tracing data to enable correlation.  This OTEP does not specify such mechanisms, assuming
implementations can use the existing
[Configuration Environment Variable Specification](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/),
or adopt some mechanism specific to the parent and child applications in question.

## Trade-offs and mitigations

### Environment variable length limit

W3C specifies that, for `tracestate`:
> [...] _SHOULD_ proapgate at least 512 characters of a combined header.

This OTEP assumes that platforms would generally not have issues with environment variables of
length 512 or more, and do not define mechanisms to split (and join) `TRACESTATE` into (and from)
multiple environment variables.  Implementations _SHOULD_ perform truncation according to
[3.3.1.5 tracestate Limits](https://www.w3.org/TR/trace-context/#tracestate-limits) in W3C
Trace Context specification if deemed necessary.

### No `Baggage`

This OTEP intentionally omits handling of `Baggage` using
[W3C's Baggage](https://www.w3.org/TR/baggage/) propagation.

This is to defer decisions regarding handling of values that exceed platform-specific limitations
on environment variable length.

### Naive intermediary

In cases where the spawned subprocess is not tracing aware, but would spawn grandchild processes
that are tracing aware, these grandchildren will inherit `TRACEPARENT` and `TRACESTATE` unmodified.
While this would produce incomplete traces, they would still be logically coherent, as valid
ancestor-descendant relationships would still be recorded.  As such, this behaviour is considered
desirable, and do not need to be mitigated.

This applies to cases where the subprocess is started using a wrapper script or launcher.

### Lack of control over environment variables

If the parent process has no way to directly control a child process' environment variables at
launch, consider mitigating using a wrapper which forwards environment variables specified as
command line arguments, or implement such capability into the child application itself.

For example:

```sh
# `sudo` may be configured to disallow inheritance of parent process' env vars
sudo -u worker worker_app --env TRACEPARENT=00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01
```

### Conflict with existing uses of `TRACEPARENT` and `TRACESTATE`

This OTEP offers no guideance on handling cases where a spawned subprocess has preexisting
behaviour regarding `TRACEPARENT` and/or `TRACESTATE` environment variables that conflicts with
their use with trace propagation.  It is assumed that the parent process would have, or be given,
enough information regarding such subprocess to act appropriately.

## Prior art and alternatives

Many CI/CD plugins and wrappers already implement support for `TRACEPARENT` and `TRACESTATE`,
or some equivalent commandline argument:

* [Jenkins OpenTelemetry Plugin](https://github.com/jenkinsci/opentelemetry-plugin)
* [otel-cli generic wrapper](https://github.com/equinix-labs/otel-cli)
* [go-test-trace](https://github.com/rakyll/go-test-trace/commit/22493612be320e0a01c174efe9b2252924f6dda9)
* [Concourse CI](https://github.com/concourse/docs/pull/462)
* [pytest-opentelemetry](https://github.com/chrisguidry/pytest-opentelemetry/)
* [hotel-california](https://github.com/parsonsmatt/hotel-california/issues/3)

### OTEP 258

[OTEP 258: Environemnt Variable Specification for Context Baggae Propagation](https://github.com/open-telemetry/oteps/pull/258)
addresses the same problem space, and predates this OTEP.  However, OTEP 258 appears to have stalled
with unaddressed issues.  The author of this OTEP felt that attempts to address said issues, in
addition to further ambiguities discovered whilst doing so, results in an almost complete rewrite
of the OTEP.  Consequently, it was deemed more productive to move forward with a more narrowly
scoped proposal that explicitly omits decisions that are not fatal ambiguities to the de facto
standard already in use.

## Open questions

### Fallback environment variables

Is specifying fallback environment variables for `TRACEPARENT` and `TRACESTATE` desirable and
necessary?  Are there cases of such conflict in practice?  Should the decision to specify such
variables block this OTEP from moving forward?

### Client vs Producer

Should the parent process be explicit about acting in a `CLIENT` vs `PRODUCER` role?  Is it
sufficient to assume that the parent process can trust its subprocess to be accurate in its
span kind usage?

## Future possibilities

- Semantic conventions for tracing across subprocesses
- More explicit guidance on using the same span context or using a span link
