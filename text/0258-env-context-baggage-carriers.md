# Environment Variables as Context and Baggage Carriers

Add requirements for inter-process context and baggage propagation using
environment variables as the carrier.

This OTEP is based on discussion in [opentelemetry-specification #740](https://github.com/open-telemetry/opentelemetry-specification/issues/740#issue-665588273), builds on the [Context Propagators
API](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/context/api-propagators.md)
and follows [OTEP 0205: Context propagation requirements for messaging semantic conventions](https://github.com/open-telemetry/oteps/blob/main/text/trace/0205-messaging-semantic-conventions-context-propagation.md)
as a guide.

## Motivation

There are existing requirements for context propagation over HTTP,
messaging systems, and others, but there are no requirements for
inter-process context propagation, although there is considerable
ad-hoc propagation being done using environment variables today (see [Prior Art](#prior-art)).
This OTEP is to formalize a standard based on the existing informal
consensus.

These are some examples of batch and system use cases that could
benefit from this standardization:

* CI/CD systems
* Build and deployoment tooling
* ETL
* Automation tools
* System-level tooling
* Command-line tools, scripts, etc.
* Tracing helper tools

## Explanation

To propagate context from a parent process to a child process,
instrumentation can use environment variable(s) with upper-cased variable names
as a carrier for the TextMapPropagator:

```
  +----------------+
  | Parent Process |
  +----------------+
          |
          | Environment variable(s) used as the carrier
          | for context propagation, e.g.: TRACEPARENT
          v
  +----------------+
  | Child Process  |
  +----------------+
```

## Proposed addition to Process and Process Runtime Resources Semantic Conventions

This is proposed to be added to the
[Process and Process Runtime Resources](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/resource/process.md)
semantic conventions as a new section and sub-section before the exiting
"Process" section and at the same level as that section:

### Conventions

> This document does not specify semantic conventions for Spans related to
> process execution. Future versions of these conventions will give clear
> recommendations on process Spans.

#### Inter-Process Context Propagation

To propagate context from a parent process to a child process,
environment variable(s) are used as the TextMapPropagator carrier
with upper-cased key names.

## Internal details

Instrumentation implementations will likely implement support for inter-process context propagation in two pieces:

* Setting environment variable(s) as the carrier when invoking child processes.
* Reading environment variable(s) when a process starts.

### Should this be added to the specification and/or semantic conventions?

Today, we have protocol-specific context propagation requirements in
both opentelemetry-specification and in semantic-conventions. This OTEP
is specifically for context propagation between processes, and the most
appropriate place seems to be to add a context propagation section to the
[process resource semantic conventions](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/resource/process.md), which is what is proposed above.

Note that the existing semantic conventions that cover context
propagation are focused on semantic conventions for *span attributes*,
whereas the process semantic conventions are for *resource attributes*.
The process resource semantic conventions do not contain any guidance
on when Spans might be created related to processes. Adding guidance
for process Spans is considered out-of-scope for this OTEP, but would
be a good subject for a future OTEP, and is mentioned as an area for
future work in the suggested [Conventions](#conventions) section.

For reference, the existing locations for protocol-specific context
propagation requirements are:

* Specification: [API Propagators, Propagators Distribution section](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/context/api-propagators.md#propagators-distribution) -- individual propagation mechanisms are referenced from external sources
* Specification: [API Propagators, B3 Requirements section](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/context/api-propagators.md#b3-requirements) -- specific considerations for mapping B3 to/from OpenTelemetry
* Semantic Conventions: [Messaging Spans, Context Propagation section](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/messaging/messaging-spans.md#context-propagation)
* Semantic Conventions: [CloudEvents, Conventions section](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/cloudevents/cloudevents-spans.md#conventions) -- references external specification for the context propagation mechanism
* Semantic Conventions: [AWS Supplemantary Guidelines, Context Propagation section](https://github.com/open-telemetry/semantic-conventions/blob/main/supplementary-guidelines/compatibility/aws.md#context-propagation)

### Environment variable capitalization

The recommendation is to upper-cased keys in the environment variable carrier
for the TextMapPropagator to align with prior art already doing this type of
propagation, follow common programming conventions, and work on both UNIX
and Windows systems.

#### Prior art

There is significant ad-hoc context propagation being done using environment
variables today, and it almost solely uses upper-cased environment variable
names (in particular, `TRACEPARENT` and `TRACESTATE`; see [PriorArt](#prior-art)).

#### Programming conventions

Common UNIX conventions usually reserve lower-case variable names for use in
application programs and upper-case for environment variables:

> **Constants and Environment Variable Names**
> All caps, separated with underscores, declared at the top of the file.
> Constants and anything exported to the environment should be capitalized.

Source: [Google, Shell Style Guide, Variable Names](https://google.github.io/styleguide/shellguide.html#s7.3-constants-and-environment-variable-names)

> **Variable Names**
>
> As for function names.
>
> ...
>
> **Function Names**
>
> Lower-case, with underscores to separate words. Separate libraries with ::. Parentheses are required after the function name. The keyword function is optional, but must be used consistently throughout a project.

Sources:

* [Google, Shell Style Guide, Variable Names](https://google.github.io/styleguide/shellguide.html#variable-names)
* [Google, Shell Style Guide, Function Names](https://google.github.io/styleguide/shellguide.html#function-names)

#### UNIX

UNIX system utilities use upper-case for environment variables and lower-case
are reserved for applications. Using upper-case will prevent conflicts with
internal application variables.

> Environment variable names used by the utilities in the XCU specification
> consist solely of upper-case letters, digits and the "_" (underscore) from
> the characters defined in Portable Character Set. Other characters may be
> permitted by an implementation; applications must tolerate the presence of
> such names. Upper- and lower-case letters retain their unique identities and
> are not folded together. The name space of environment variable names
> containing lower-case letters is reserved for applications. Applications can
> define any environment variables with names from this name space without
> modifying the behaviour of the standard utilities.

Source: [The Open Group, The Single UNIX® Specification, Version 2, Environment Variables](https://pubs.opengroup.org/onlinepubs/7908799/xbd/envvar.html)

#### Windows

Windows appears to be case-insensitive with environment variables, so the case
of environment variables is irrelevant for our uses on Windows, and there might
be a conflict if OpenTelemetry keys match internal environment variables used
in applications.

There does not seem to be clear documentation on case-insensitivity on Windows
from Microsoft, however there are consistent third-party references.

Here is a description of an issue in CPython:

> os.environ docs don't mention that the keys get upper-cased automatically on
> e.g. Windows.

Source: [CPython issue #101754](https://github.com/python/cpython/issues/101754)

Here is a related documentation addition that was made to clarify functionality:

> On Windows, the keys are converted to uppercase. This also applies when
> getting, setting, or deleting an item. For example, environ['monty'] =
> 'python' maps the key 'MONTY' to the value 'python'.

Source: [Python 3 library, os.environ](https://docs.python.org/3/library/os.html#os.environ)

Lastly, here is a programming reference that mentions case-insensitivity:

> **(Windows) Environment Variables**
>
> Environment Variables in Windows are NOT case-sensitive (because the legacy
> DOS is NOT case-sensitive). They are typically named in uppercase, with words
> joined with underscore (_), e.g., JAVA_HOME.

Source: [yet another insignificant programming notes... Environment Variables in Windows/macOS/Linux](https://www3.ntu.edu.sg/home/ehchua/programming/howto/Environment_Variables.html#zz-2.)

### Allowed characters

The characters allowed in keys by the TextMapPropagator are all allowed as environment variable keys on UNIX and Windows.

#### OpenTelemetry TextMapPropagator

Allowed characters in keys and values, in short:

> "!" / "#" / "$" / "%" / "&" / "'" / "*"
> / "+" / "-" / "." / "^" / "_" / "`" / "|" / "~"
> / DIGIT / ALPHA

Details below.

> In order to increase compatibility, the key/value pairs MUST only consist of
> US-ASCII characters that make up valid HTTP header fields as per RFC 7230.

Source: [OpenTelemetry Specification, API Propagators, TextMapPropagator](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/context/api-propagators.md#textmap-propagator)

> header-field   = field-name ":" OWS field-value OWS
> field-name     = token

Source: [RFC 7230, section 3.2](https://datatracker.ietf.org/doc/html/rfc7230#section-3.2)

> Most HTTP header field values are defined using common syntax
> components (token, quoted-string, and comment) separated by
> whitespace or specific delimiting characters.  Delimiters are chosen
> from the set of US-ASCII visual characters not allowed in a token
> (DQUOTE and "(),/:;<=>?@[\]{}").
>
> token          = 1*tchar
>
> tchar          = "!" / "#" / "$" / "%" / "&" / "'" / "*"
> / "+" / "-" / "." / "^" / "_" / "`" / "|" / "~"
> / DIGIT / ALPHA
> ; any VCHAR, except delimiters

Source: [RFC 7230, section 3.2.6](https://datatracker.ietf.org/doc/html/rfc7230#section-3.2.6)

#### UNIX

Summary: The [Portable Character Set](https://pubs.opengroup.org/onlinepubs/7908799/xbd/charset.html#tag_001_004) lists all values allowed by `tchar` in RFC 7230.

> These strings have the form name=value; names do not contain the character =.
> For values to be portable across XSI-conformant systems, the value must be
> composed of characters from the portable character set (except NUL and as
> indicated below). There is no meaning associated with the order of strings in
> the environment. If more than one string in a process' environment has the
> same name the consequences are undefined.
>
> Environment variable names used by the utilities in the XCU specification
> consist solely of upper-case letters, digits and the "_" (underscore) from
> the characters defined in Portable Character Set. Other characters may be
> permitted by an implementation; applications must tolerate the presence of
> such names.

Source: [The Open Group, The Single UNIX® Specification, Version 2, Environment Variables](https://pubs.opengroup.org/onlinepubs/7908799/xbd/envvar.html)

#### Windows

Summary: Windows only disallows the "=" sign, which is not an allowed charater in `tchar` in RFC 7230.

> The name of an environment variable cannot include an equal sign (=).

Source: [Microsoft, Win32, Processes and Threads, Environment Variables](https://learn.microsoft.com/en-us/windows/win32/procthread/environment-variables)

> The maximum size of a user-defined environment variable is 32,767 characters.

Source: [Microsoft, Win32, Processes and Threads, Environment Variables](https://learn.microsoft.com/en-us/windows/win32/procthread/environment-variables)

### Getter and Setter implementation

In our case, environment variables names on Windows are case insensitive, so we
will not violate any of the MUSTs by **not** preserving casing in our `Set` and
`Get` implementations for environment variable propagation.

> Set
> ...
> The implementation SHOULD preserve casing (e.g. it should not transform
> Content-Type to content-type) if the used protocol is case insensitive,
> otherwise it MUST preserve casing.

-- [API Propagators - Set](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/context/api-propagators.md#set)

> Get
> ...
> The Get function is responsible for handling case sensitivity. If the getter
> is intended to work with a HTTP request object, the getter MUST be case
> insensitive.

-- [API Propagators - Get](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/context/api-propagators.md#get)

## Trade-offs and mitigations

### Case-sensitivity

On Windows, in particular, because environment variable keys are case
insensitive, there is a (low?) chance that automatically instrumented context
propagation variables could conflict with existing application environment
variables. It will be important that this instrumentation can be disabled
should such cases arise.

### Privacy and Security

Environment variables are visible to any process on the system, so it is
important to emphasize existing guidance to not place sensitive information
that could be propagated.

If an attacker on a local system can send tracing messages, they could view a
process' `TRACEPARENT` environment variable and inject traces that appear to be
part of the trace they are targeting. This is likely not much of a concern in
many deployment scenarios, but if it is, one mitigation would be to protect the
trace endpoint by requiring authentication or disabling inter-process context
propagation.

## Prior art and alternatives

### Prior art

Many existing users of `TRACEPARENT` and/or
`TRACESTATE` environment variables mentioned in
[opentelemetry-specification #740](https://github.com/open-telemetry/opentelemetry-specification/issues/740):

* [Jenkins OpenTelemetry Plugin](https://github.com/jenkinsci/opentelemetry-plugin)
* [otel-cli generic wrapper](https://github.com/equinix-labs/otel-cli)
* [Maven OpenTelemetry Extension](https://github.com/cyrille-leclerc/opentelemetry-maven-extension)
* [Ansible OpenTelemetry Plugin](https://github.com/ansible-collections/community.general/pull/3091)
* [go-test-trace](https://github.com/rakyll/go-test-trace/commit/22493612be320e0a01c174efe9b2252924f6dda9)
* [Concourse CI](https://github.com/concourse/docs/pull/462)
* [BuildKite agent](https://github.com/buildkite/agent/pull/1548)
* [pytest](https://github.com/chrisguidry/pytest-opentelemetry/issues/20)
* [Kubernetes test-infra Prow](https://github.com/kubernetes/test-infra/issues/30010)
* [hotel-california](https://github.com/parsonsmatt/hotel-california/issues/3)

### Alternatives and why they were not chosen

#### Case-sensitive variable names (not upper-casing keys in the carrier)

Case-sensitive variable names would allow use of the TextMapPropagator
without specifying custom Getter and Setter implementations and would
fully satisfy the case-sensitivity-related SHOULDs and MUSTs in the
[Propagators API](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/context/api-propagators.md)
specification. There are two overriding reasons to use upper-cased
variable names, however:

1. Windows environment variables are case-insensitive.
2. Significant prior art overwhelmingly uses upper-cased environment variables.

#### Using a file for the carrier

Using a JSON file that is stored on the filesystem and referenced
through an environment variable would eliminate the need to workaround
case-insensitivity issues on Windows, however it would introduce a
number of issues:

1. Would introduce an out-of-band file that would need to be created and reliably cleaned up.
2. Managing permissions on the file might be non-trivial in some circumstances (for example, if `sudo` is used).
3. This would deviate from significant prior art that currently uses environment variables.

## Open questions

The author has no open questions at this point.

## Future possibilities

1. Adding guidance on when to create process Spans.
