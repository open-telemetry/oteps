# Allowing experimental extensions to Semantic Conventions

A process that allows experimental / non-stable semantic conventions to exist within stable versions and how they evolve into stable semantic conventions.

## Motivation

Today, semantic conventions are released as a large bundle of "all or nothing".
OpenTelemetry struggles with ensuring semantic conventions represent true "generic" need in observability in addition to providing a mechanism to "discover" or "experiment" with a standard.
This [Slack thread on #otel-specification](https://cloud-native.slack.com/archives/C01N7PP1THC/p1667997213067169?thread_ts=1667933561.313819&cid=C01N7PP1THC) highlights the problem in context of a specific semantic convention pull request.
What should OpenTelemetry do with semantic convention proposals when there's unclear evidence that a particular signal is necessary or ubiquitous?

## Explanation

This proposal generally follows the recommendations of [RFC6648 section 4 “Recommendations for Protocol Designers.”](https://www.rfc-editor.org/rfc/rfc6648#section-4)
RFC6648 recommends a registry of names be established which standardizes usage and prevents collisions.
We already have such a registry in the form of our semantic conventions.
RFC6648 further recommends that X- prefix (or similar structure) MUST NOT be defined as experimental or nonstandard, and that names missing the X- prefix MUST NOT be defined as stable or standard.

## Definitions

**Attribute**: For the purposes of this document, an attribute may refer to any component of any signal emitted by OpenTelemetry instrumentation.
This may include span or metric attribute keys and values, span or metric names, log or event attribute names and values, or any other value emitted by OTel instrumentation.

**Private Attribute**: An attribute is private if it is meant only for the exclusive private use of the instrumentation author, and is not expected to be emitted by any public library or understood by any public backend tools.

**Registered Attribute**: A registered attribute is any attribute which is in the provisional registry outlined in the proposal or the semantic conventions.

## Proposal

For the purposes of this document only attributes are discussed, but it should be understood that the TC may decide to make other telemetry components such as metric names, event names, or other components registerable.
This proposal SHOULD also apply to any such components at the discretion of the TC.

A procedure will be established by the technical committee (TC), or an entity designated by the TC, for registration of provisional semantic convention attributes.
Any non-private attributes which can be registered SHOULD be registered in the provisional registry or an official OpenTelemetry semantic convention.
The provisional registry MAY be separate from the semantic conventions, a status applied to some entries in the existing semantic conventions, or some other mechanism specified by the TC.

Instrumentation authors SHOULD assume that any attribute they create may become standardized, public, commonly deployed, or usable across multiple OpenTelemetry clients, distributions, or receivers.
Instrumentation authors SHOULD choose attribute keys that are descriptive, relevant, and that they have reason to believe are not currently in use or reserved by the TC for future use.
Before creating a new attribute, the attribute author SHOULD review the semantic conventions and provisional registry to see if an attribute already exists which fits the intended use case.
If such a registered attribute exists, then it SHOULD be preferred over the creation of new attributes.

An entry MUST NOT be refused without credible reason to believe that registration would be harmful.
Some reasonable grounds for refusal might be frivolous use of the registry, TC consensus that the attribute may be a security or privacy risk, conflict with another attribute or a name reserved by the TC for future use, or that the proposal is sufficiently lacking in purpose, or misleading about its purpose, that it can be held to be a waste of time and effort.
Note that a disagreement about technical details are not considered grounds to refuse listing in the provisional registry.

A provisional registry entry which has gained enough support, use, and market penetration SHOULD be promoted to the semantic conventions following existing procedures for review and inclusion.
When proposing a new semantic convention, existing provisional registry entries SHOULD be considered, and a semantic convention proposal SHOULD NOT break compatibility with an existing provisional registry entry without sufficient reasoning.
If a semantic convention is added which is not compatible with existing provisional registry entries, a new name SHOULD be created.
For example, if a provisional entry exists for http.headers, a semantic convention proposal which breaks compatibility might choose the name http.message_headers in order to avoid a conflict.

This document explicitly does NOT define a process for removal of a registered attribute from the provisional registry or semantic conventions.
Because absence of evidence does not constitute evidence of absence, there is no way to know if any registered attribute is in active use.
If a registered attribute is deemed by the TC to be harmful in some way, it SHOULD be marked deprecated, and sufficient reasoning for the deprecation should be included in the registry.

Consumers of telemetry SHOULD consider attributes and their values to be untrusted input.
Attribute values SHOULD be validated before assuming they conform to the format specified in the semantic conventions or the provisional registry.
Consumers SHOULD consider any security and privacy implications of displaying unhidden attribute values.

## Alternatives Considered

### "x." prefix

This alternative involves establishing an “x.” prefix convention for experimental conventions.
For example, "x.mydatabase.attribute" may be an experimental version of the eventual "mydatabase.attribute" attribute key.
See [X- convention for HTTP message headers](#x--http-header-convention) below for more details about why this alternative was discarded.

## Prior Art

### HTTP Header field and media type registration procedures

[RFC6648](https://www.rfc-editor.org/rfc/rfc6648), which serves as the primary inspiration for this OTEP, recommends [RFC3864](https://www.rfc-editor.org/rfc/rfc3864) "Registration Procedures for Message Header Fields" and [RFC4288](https://www.rfc-editor.org/rfc/rfc4288) "Media Type Specifications and Registration Procedures" as positive examples of "simpler registration rules."

Much of this OTEP draws inspiration from [RFC6864 Section 4](https://www.rfc-editor.org/rfc/rfc3864#section-4), which is itself an implementation suggested RFC6648 section 4.

### X- HTTP header convention

* Deprecated in [RFC6648](https://datatracker.ietf.org/doc/html/rfc6648)

For many years, HTTP headers had a convention of using an X- prefix for headers which were either not yet standardized or not meant to ever be standardized.
The practice, which began as an informal suggestion, and has since permeated much of web practices and technologies, was officially deprecated in RFC6648.
A more full accounting of the history can be found in [RFC6648 Appenix A](https://www.rfc-editor.org/rfc/rfc6648#appendix-A), but a short version is included here for completeness.

In 1975, Brian Harvey suggested “an initial letter X to be used for really local idiosyncrasies [sic]” with regard to FTP parameters.
The convention was later adopted for several standard and nonstandard uses including email as user extension fields, SIP as P- prefix headers, iCalendar x- tokens, and HTTP X- headers.
The practice has since been deprecated in FTP fields, email, SIP, and HTTP.

The inclusion of X- prefixed headers, and similar constructs in other standards, introduced a series of problems.
The first, and most severe, issue is that X- prefixed names leaked into the set of standardized headers.
An example of this can be seen in the HTTP media type standard where x-gzip and x-compress are considered equivalent to gzip and compress respectively, or the X-Archived-At message header field which MAY be parsed but MUST NOT be generated.
An exhaustive list does not exist and is not provided here, but an engineer who has been familiar with web technologies for a sufficiently long time is likely to be familiar with the phenomenon which unnecessarily complicates specifications and standards.

The second problem is that including X- prefixes in a name encodes, either explicitly or implicitly, an understanding that the field in question is experimental or nonstandard in nature.
It implies a level of instability which may or may not be present.
Even a field which starts as an experiment or nonstandard may gain enough popularity or market use where it becomes a de facto standard.
When this happens, the understanding of the experimental nature of the field becomes a false assumption.
Subsequent efforts to formally standardize the field must take such use into account and often result in standards which either codify the X- version as the standard, or state that it is to be treated as equivalent, leading to the first issue again.

#### Examples

##### Email extension fields

* Implemented in [RFC822](https://www.rfc-editor.org/rfc/rfc822)
* Removed in [RFC2822](https://www.rfc-editor.org/rfc/rfc2822)

##### SIP P- headers

[RFC5727](https://www.rfc-editor.org/rfc/rfc5727)

##### iCalendar x-token

[RFC5545](https://www.rfc-editor.org/rfc/rfc5545)

## Open Questions

## Future Possibilities
