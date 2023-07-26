# Experimental Semantic Convention Attribute Registration

A process for the registration of semantic convention attributes not yet included in a semantic convention.

## Motivation

Today, semantic conventions are "all or nothing."
Each instrumentation implements one or more full semantic conventions with little room for experimentation.
The lack of experimentation means there is often limited experience and evidence to guide  convention authors.
It also often leaves instrumentation authors with little or no guidance if they wish to capture telemetry not yet covered by semantic conventions.
This [Slack thread on #otel-specification](https://cloud-native.slack.com/archives/C01N7PP1THC/p1667997213067169?thread_ts=1667933561.313819&cid=C01N7PP1THC) highlights the problem in context of a specific semantic convention pull request.
What should OpenTelemetry do with semantic convention proposals when there's unclear evidence that a particular signal is necessary or ubiquitous?

## Explanation

This proposal generally follows the recommendations of [RFC6648 section 4 “Recommendations for Protocol Designers.”](https://www.rfc-editor.org/rfc/rfc6648#section-4)
RFC6648 recommends a registry of names be established which standardizes usage and prevents collisions.
This OTEP separates a dictionary of semantic convention attributes from the semantic conventions, to be referenced by semantic conventions when necessary, and establishes a procedure for registering new and experimental semantic convention attributes.

RFC6648 further recommends that X- prefix (or similar structure) MUST NOT be defined as experimental or nonstandard, and that names missing the X- prefix MUST NOT be defined as stable or standard.
This is addressed below in [alternatives considered](#alternatives-considered).

## Definitions

**Attribute Dictionary**: The attribute dictionary contains all semantic convention attributes, including their names and value definitions.

**Registered Attribute**: A registered attribute is any attribute which is in the attribute dictionary.

## Proposal

A semantic convention attribute dictionary will be established which is maintained in a separate file or directory from the semantic conventions, seeded with all existing semantic convention attributes.
All semantic conventions MUST refer to the attribute dictionary for all attributes used by that semantic convention.
Semantic conventions MUST NOT establish an attribute not in the attribute dictionary without explicit approval of the TC with the reasoning documented directly in that convention.
A procedure will be established by the technical committee (TC), or an entity designated by the TC, for registration of new semantic convention attributes.
Any attributes not intended strictly for private use SHOULD be registered in the attribute dictionary.

Instrumentation authors SHOULD assume that any attribute they create may become standardized, public, commonly deployed, or usable across multiple OpenTelemetry clients, distributions, or receivers.
They SHOULD choose attribute keys that are descriptive, relevant, and that they have reason to believe are not currently in use or already proposed.
Before creating a new attribute, the attribute author SHOULD review the attribute dictionary to see if an attribute already exists which fits the intended use case.
If such a registered attribute exists, then it SHOULD be preferred over the creation of new attributes unless there is sufficient reason to create a new attribute.
Any new attributes SHOULD be registered in the attribute dictionary before an instrumentation using the attribute is released.
If an instrumentation uses an attribute not included in the semantic conventions it is recommended that the attribute not be collected by default unless it is fundamental to the instrumentation.
For example, an instrumentation which instruments a component not yet addressed by semantic conventions.

An entry MUST NOT be refused without credible reason to believe that registration would be harmful.
Note that a disagreement about technical details are not considered grounds to refuse listing in the attribute dictionary.
Some examples of reasonable grounds for refusal: frivolous use of the dictionary such as registering nonsense words or inside jokes, a credible security or privacy risk, conflict with another attribute or proposal for a new attribute, or that the proposal is sufficiently lacking in purpose, or misleading about its purpose, that it can be held to be a waste of time and effort.
This list is non-exhaustive and final decision for refusal is at the discretion of the TC.

Once an attribute is registered it MUST NOT receive any breaking changes.
This document explicitly does NOT define a process for removal of a registered attribute from the attribute dictionary.
Because absence of evidence does not constitute evidence of absence, there is no way to know if any registered attribute is in active use.
If a registered attribute is deemed by the TC to be harmful in some way, it SHOULD be marked deprecated, sufficient reasoning for the deprecation should be included in the dictionary, and, if appropriate, a new attribute SHOULD be registered with a different name.
A registered attribute MAY ONLY be removed if, at the discretion of the TC, it is determined to be harmful or offensive to the project, its contributors, users, or some other group.
For example, if an attribute `http.headers` is registered and it is later determined by the TC to be harmful, it should be deprecated and a new attribute such as `http.message_headers` or similar created to replace it.

Consumers of telemetry SHOULD consider attributes and their values to be untrusted input.
Attribute values SHOULD be validated before assuming they conform to the format specified in the attribute dictionary.
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

## Future Possibilities

In the future, the dictionary may be extended to include other telemetry descriptors such as metric keys, span names, tracer names, et cetera.
