# Sensitive Data Redaction

This is a proposal for adding treatment of sensitive data to the OpenTelemetry
(OTel) Specification and semantic conventions.

## Motivation

When collecting data from an application, there is always the possibility, that
the data contains information that shouldn't be collected, because it is either
leaking (parts of) credentials (passwords, tokens, usernames, credit card information),
can be used to uniquely identify a person (name, IP, email, credit card information)
which may be protected through certain regulations. By adding OTel to a library or
instrumentation an end-user of OTel is facing exactly this challenge: emitted
telemetry may carry such sensitive data.

While it’s ultimately the responsibility of the legal entity operating an application
to protect sensitive data, end-users of OpenTelemetry (developers, operators working
for that entity) are turning to the authors of OpenTelemetry – or to those of
libraries that implement OpenTelemetry, like the Azure SDK – to have means in
place to redact/filter sensitive data. Without that capability provided, they
will raise security issues and/or will drop OpenTelemetry eventually due to it
not meeting their security/legal requirements.

In this OTEP you will find a proposal for adding treatment of sensitive data to
OpenTelemetry.

By adding the proposed features, OpenTelemetry will introduce the following
principles for treating sensitive data:

* OpenTelemetry MUST allow the end-user to meet with their security/privacy/compliance
  requirements regarding the data being collected.
* OpenTelemetry MUST avoid redaction offerings that lead to bigger security issues
  such as <https://en.wikipedia.org/wiki/ReDoS> (e.g. the redaction logic is poorly
  implemented, so a hacker could forge certain input to DDoS the redaction engine itself).
* OpenTelemetry SHOULD allow the telemetry data to apply different redaction
  logic per telemetry pipeline/exporter in a single process.

## Explanation

### Overview

This proposal aims to provide the following features:

- a consistent way to configure sensitivity requirements for end-users of the
  OpenTelemetry SDK and in instrumentation libraries, including predefined
  “configuration profiles” and ways for fine grain configuration.
- related to those a way to enrich attributes in the semantic conventions with
  sensitivity information and ways of redaction.
- methods for consumers of the OpenTelemetry API to enrich collected Attributes
  with sensitivity information and hooks to apply different levels of redaction.
- A redactor implements the logic to apply redaction and that owns predefined helpers
  for redaction in the SDK (URLParams filtering, Zeroing IPs, etc.).

The following limitations apply:

- Although it is possible that sensitive data is contained in any data generated
  by OpenTelemetry (e.g. the span name, or the instrumentation scope name could
  contain sensitive data), only the following is treated in the scope of this PR:
  [Attributes](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/common/README.md#attribute)
  and payload fields like [log bodies](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#field-body).

### Configuration-driven redaction

As stated in the motivation the responsibility for applying redaction lays with
the operator of an application, since they are the only ones knowing the specific
requirements for their environment. Because of that this proposal suggests that
OpenTelemetry should provide a solution to end users that allows them to provide
a configuration for their redaction requirements. This configuration consists of
two pieces:

* Redaction requirement "configuration profiles", that end user use to express
  their overall redaction requirements.
* Fine grained configuration that can be used to express requirements that are 
  unique to their environment.

**Example**: An end user may configure an application with "stricter" requirements
 and some custom requirements like the following:

* They provide a profile name `STRICTER` via an environment variable, e.g.
  OTEL_SENSITIVE_DATA_PROFILE="STRICTER"
* They add custom requirements in a
  [configuration file](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/configuration/file-configuration.md), e.g.
  
  ```yaml
  sensitiveData:
      - attribute: url.query
        redaction: REDACT_ALL_URL_VALUES
      - attribute: com.example.customer.name
        redaction: DROP
      - attribute: com.example.customer.email
        redaction: 's/^[^@]*@/REDACTED@/'
  ```

The above is just an example for how such a configuration may look like, different
formats are discussed below.

#### Redaction requirement profiles

An end user should be able to pick from a set of redaction requirement profiles
that cover common configurations:

- `NONE` or `NEVER`: No redaction is applied. This is useful if an end user
  chooses to run a collector nearby the application that will take care of
  redaction out of process. If this profile is selected the SDK must make sure
  that (almost) no overhead remains from the redaction methods.
- `DEFAULT`: A set of rules that especially make sure that no security-sensitive
  data is exposed, only limited PII-related redactions are applied.
- `STRICT`: A set that builds on top of `DEFAULT` and also applies commonly required
  PII-related redactions (anonymize emails, names, IPs, etc)
  `STRICTER`: A set that builds on top of `STRICT` and applies even more redactions, 
  e.g. drop all query attribute values, drop PII-related data, etc.
- `DROP` or `ALWAYS`: All values with (potentially) sensitivity details will be dropped

Those profiles will be defined through the semantic conventions (see below).

#### Fine-grained configuration

On top of the redaction requirements profiles an end user should have capabilities
to add their own local requirements, either through configuration or through code.

The configuration is a list of rules that the SDK can parse and apply. An item
in this list will consist of:

- Conditions when a redaction should be applied
- Instructions for the redaction that will be applied when the conditions are met

In an example configuration file a list could look like the following:

```yaml
  sensitiveData:
      - attribute: url.query
        redaction: REDACT_ALL_URL_VALUES
      - attribute: com.example.customer.name
        redaction: DROP
      - attribute: com.example.customer.email
        redaction: 's/^[^@]*@/REDACTED@/'
```

In this example the condition is expressed via `attribute` and the instructions
via `redaction`, where either a constant is selected that runs a predefined helper
method, or a sed-like expression that uses a regular expression to apply redaction.

Another possibility is to re-use a language like
[OTTL](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/pkg/ottl/README.md).

### Semantic Conventions

To enable the redaction requirement profiles the semantic conventions need to be
annotated with sensitivity details.

Example:

| Attribute | ... existing columns ... | sensitivity details |
|-----------|--------------------------|---------------------|
| `url.query`|                         | Rationale: Some verbatim wording why this is the way it is below<br>Type: `mixed`<br>`DEFAULT`: `REDACT_INSECURE_URL_PARAM_VALUES`<br>STRICT: `REDACT_ALL_URL_VALUES`<br>`STRICTER`: `DROP` |
| `client.address`|                    | Rationale: some reasons why dropping octets may be required<br>Type: `maybe_pii`<br>`DEFAULT`: `NONE`<br>`STRICT`: `'s/([0-9]+\.[0-9]+\.[0-9]+\.)[0-9]+/\10/'`<br>`STRICTER`: `'s/([0-9]+\.[0-9]+\.)[0-9]+\.[0-9]+/\10.0/'` |
| `enduser.creditCardNumber`**[1]** |     | Rationale: ...<br>Type: `always_pii`<br>DEFAULT: `EXTRACT_IIN`<br>`STRICT`: `DROP`|

**[1]**: _This is a made-up example for demonstration purpose, it’s not part of the current semantic conventions. It gives a more nuanced example, e.g. that extracting the IIN might be an option over dropping the number completely. It also demonstrated the value of “type”, which can enable Data lineage use cases_

The `DROP` keyword means that the value is replaced with `REDACTED` (or set to null, …).

It is the responsibility of the OpenTelemetry SDK to implement those sensitivity details provided by the semantic conventions.

### Annotate attributes with sensitivity information

As an additional building block the OpenTelemetry API should provide capabilities
to library (and application) authors to apply in-place redaction

One option is to add additional attributes to methods like `setAttribute`:

```
span.setAttribute("url.query", url.toString(), <SENSITIVITY_DETAILS>);
```

Or, if the method can not be extended, an additional method can be called:

```
span.setAttribute("url.query", OpenTelemetry.redact(url.toString(), <SENSITIVITY_DETAILS>);
```

Note that those API calls are no-op and will be implemented by the SDK (as we do it with other API methods as well), this way (almost?) no additional overhead will be created by introducing those annotations.

The `<SENSITIVY_DETAILS>` should look like the following:

```
{
  `<REDACTION_PROFILE>` => `<REDACTION_CONFIGURATION>`,
  ...
}
```

There may be multiple redaction profiles (like `DEFAULT`, `STRICT`, etc.), and
for each one a configuration (as outlined above) may be applied.

any sensitive data.

## Internal details

### Redactor

To accomplish redaction the SDK needs a component (similar to a sampler) that inspects attributes when they are set and applies the required redactions:

- `Redactor::setup(profile)` will setup the redactor with the given profile,
  maybe a constructor? depends on the language
- `Redactor::redact(value, <sensitivityDetails>)` will return the redacted value
  using the provided profile

The redactor will also have all the methods to apply predefined redactions
(`REDACT_ALL_URL_VALUES`, `REDACT_INSECURE_URL_PARAM_VALUES`, `DROP`, etc.).
If a method is not implemented (either by the SDK or by the end-user choosing
one that does not exist), it will default to apply `DROP` to avoid leakage of

## Additional context

Treating sensitive data properly is a very complex multi-dimensional topic.
Below you will find some additional context on that subject.

### Types of sensitive data

There are different kinds of “sensitivity” that may apply to data. The ones most
relevant in this proposal are “security” and “privacy”. They may overlap but we
distinguish them as follows:

- security-relevant sensitive data: any information that when exposed
  [weakens the overall security of the device/system](https://en.wikipedia.org/wiki/Vulnerability_(computing)).
- privacy-relevant sensitive data: any information that when exposed
  [can adversely affect the privacy or welfare of an individual](https://en.wikipedia.org/wiki/Information_sensitivity).

Note, that there are other kinds of sensitive data (business information,
classified information), which are not covered extensively by this proposal.

### Level of Sensitivity

The level of sensitivity of an information can also be different and that
sensitivity can be contextual, e.g.

- The password of a user without privileges is less sensitive than the password
  of an administrator
- The "client IP" in a server-to-server communication is less sensitive than the
  "client IP" in an client-to-server communication, where the client can be
  linked to a human.
- API tokens of a demo system are less sensitive than API tokens for a production
  system
- The license plate of an individual’s car is less sensitive than their social
  security number
- The full name of a user in a social network is less sensitive than the full
  name of a user in a medical research database

Depending on the sensitivity data an end-user of an observability system may
weigh up if collecting this data is worth it.

### Regulatory and other requirements

Due to the negative effects that the exposure of sensitive data can have (see
above in "Types of sensitive data"), different entities have created regulations
for the collection of sensitive data, among them:

- GDPR
- CPRA
- PIPEDA
- HIPAA
- [more…](https://en.wikipedia.org/wiki/Information_privacy)

Additionally the entities operating the applications who leverage OpenTelemetry
may have their own requirements for treating certain sensitive data.

Finally end-users may want to apply recommendations for
[Data Minimization](https://en.wikipedia.org/wiki/Data_minimization), to avoid
"unnecessary risks for the data subject".

**Note 1**: it is not (and can not be) the responsibility of the OpenTelemetry
project to provide compliance with any of those regulations, this is a
responsibility of the OTel end-user. OTel can only facilitate parts of those
requirements.

**Note 2**: Those requirements are subject of change and outside of the control
of the OpenTelemetry community.

## Trade-offs and mitigations

### Performance Impact

By adding an extra layer of processing every time an attribute value gets set,
might have an impact on the performance. There might be ways to reduce that
overhead, e.g. by only redacting values which are finalized and ready to exported
such that updated values or sampled data does not need to be handled.

## Prior art and alternatives

### OTEPS

- [OTEP 100 - Sensitive Data Handling](https://github.com/open-telemetry/oteps/pull/100)
- [OTEP 187 - Data Classification for resources and attributes](https://github.com/open-telemetry/oteps/pull/187)

### Spec & SemConv Issues

This problem has been discussed multiple times, here is a list of existing issues
in the OpenTelemetry repositories:

- [Http client and server span default collection behavior for url.full and url.query attributes](https://github.com/open-telemetry/semantic-conventions/issues/860)
- [URL query string values should be redacted by default](https://github.com/open-telemetry/semantic-conventions/pull/961)
- [Specific URL query string values should be redacted](https://github.com/open-telemetry/semantic-conventions/pull/971)
- [Allow url.path sanitization](https://github.com/open-telemetry/semantic-conventions/pull/676)
- [Guidance requested: static SQL queries may contain sensitive values](https://github.com/open-telemetry/semantic-conventions/issues/436)
- [Semantic conventions vs GDPR](https://github.com/open-telemetry/semantic-conventions/issues/128)
- [Guidelines for redacting sensitive information](https://github.com/open-telemetry/semantic-conventions/issues/877)
- [DB sanitization uniform format](https://github.com/open-telemetry/semantic-conventions/issues/717)
- [Add db.statement sanitization/masking examples](https://github.com/open-telemetry/semantic-conventions/issues/708)
- [TC Feedback Request: View attribute filter definition in Go](https://github.com/open-telemetry/opentelemetry-specification/issues/3664)

### SemConv Pages

The semantic conventions already contains notes around treating sensitive data
(search for "sensitive" on the linked pages if not stated otherwise):

- [gRPC SemConv](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/rpc/grpc.md)
- [Sensitive Information in URL SemConv](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/url/url.md#sensitive-information)
- [GraphQL Spans SemConv](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/graphql/graphql-spans.md)
- [Container Resource SemConv](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/resource/container.md)
- [Database Spans SemConv](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/database/database-spans.md)
- [HTTP Spans SemConv](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/http/http-spans.md)
- [General Attributes SemConv](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/general/attributes.md)
- [LLM Spans SemConv](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/gen-ai/llm-spans.md)
- [ElasticSearch SemConv](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/database/elasticsearch.md)
- [Redis SemConv](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/database/redis.md)
- [Connect RPC SemConv](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/rpc/connect-rpc.md)
- [Device SemConv](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/resource/device.md) (search for GDPR)

### Existing Solutions within OpenTelemetry

The following solutions for OpenTelemetry already exist:

- [MrAlias/redact](https://github.com/MrAlias/redact) for OpenTelemetry Go
- Collector processors, including
  - [Redaction Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/redactionprocessor)
  - [Transform Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/transformprocessor)
  - [Filter Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor)

### Alternative 0: Do nothing

It’s always good to analyze this option. If we do nothing end users will still 
need to satisfy their requirements for treating sensitive data accordingly:

- Instrumentation library authors are required to manage redaction before using the OpenTelemetry API
- Application developers will do the same or are forced to join the instrumented application with a collector for redaction/filtering
- Third party solutions will be implemented.

### Alternative 1: OpenTelemetry Collector

As listed above there are multiple processors available that can be used to
redact or filter sensitive data with the OpenTelemetry collector. The challenge
with that is that it is unknown to an application (owner) if data is processed
in the collector as expected. Also, the data leaving the application might
already be a risk (non-encrypted or compromised network) or may not be allowed
(collector is hosted in a different country, which may conflict with a regulation)

Ideally a combination is used.

### Alternative 2: Backend

The backend consuming the OpenTelemetry data can provide processing for filtering
and redaction as well. The same objection as for the collector apply.

### Existing Solutions outside OpenTelemetry

There are many solutions outside OpenTelemetry that help to filter or redact
sensitive data based on security and privacy requirements:

- [sanitize_field_name in Elastic Java](https://www.elastic.co/guide/en/apm/agent/java/1.x/config-core.html#config-sanitize-field-names)
- [Filter sensitive data in AppDynamics Java Agent](https://docs.appdynamics.com/appd/24.x/24.4/en/application-monitoring/install-app-server-agents/java-agent/administer-the-java-agent/filter-sensitive-data)
- [GA4 data redaction](https://support.google.com/analytics/answer/13544947?sjid=3336918779004544977-EU)
- [Configure Privacy Settings in Matamo](https://matomo.org/faq/general/configure-privacy-settings-in-matomo/)
- [DataDog Sensitive Data Redaction](https://docs.datadoghq.com/observability_pipelines/sensitive_data_redaction/)
- [Dynatrace data privacy and security configuration](https://docs.dynatrace.com/docs/shortlink/global-privacy-settings)

## Open questions

- **Question 1**: Should sensitivity details for an attribute in the semantic
  conventions be excluded from the stability guarantees? This means, updating
  them for a **stable** attribute is not a breaking change. The idea behind
  excluding them from the stability guarantees is that the requirements are
  subject of change due to changes in technology (see [#971](https://github.com/open-telemetry/semantic-conventions/pull/971), the list of query string values will evolve
  over time) or changes in regulatory requirements or both.

## Future possibilities

Attributes and payload data are most likely to carry sensitive information, but
as stated in the overview section of the explanation other user-set properties may
carry sensitive information as well. In a later iteration we might want to review
them as well.

The proposal puts the configuration of sensitivity requirements into the hands of
the person operating an application. In a future iteration we can look into
providing end-users of instrumented applications to provide their _consent_ of
which and how data related to them is tracked, see [Do Not Track](https://en.wikipedia.org/wiki/Do_Not_Track), [Global Privacy Control](https://privacycg.github.io/gpc-spec/) and the
requirements of certain local regulations.
