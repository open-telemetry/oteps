# Sampling Configuration

An attempt to provide a framework for defining sampling configurations.

Calling this proposal half-baked would be too generous. At this time it just a vision, with many questions unanswered and without any working prototype. Its purpose is to start a discussion on the needs and wants the Open Telemetry community might have on the subject of sampling configuration and a possible way to accomplish them.

The focus is on head-based sampling, but a similar approach might be used for later sampling stages as well. Most of the technical details presented here assume Java as the platform, but should be general enough to have corresponding concepts and solutions available for other platforms too.

## Motivation

The need for sampling configuration has been explicitly or implicitly indicated in several discussions, both within the [Samplig SIG](https://docs.google.com/document/d/1gASMhmxNt9qCa8czEMheGlUW2xpORiYoD7dBD7aNtbQ) and in the wider community. Some of the discussions are going back a number of years, see for example

- issue [173](https://github.com/open-telemetry/opentelemetry-specification/issues/173): Way to ignore healthcheck traces when using automatic tracer across all languages?
- issue [679](https://github.com/open-telemetry/opentelemetry-specification/issues/679): Configuring a sampler from an environment variable
- issue [1060](https://github.com/open-telemetry/opentelemetry-java-instrumentation/issues/1060): Exclude URLs from Tracing
- issue [1844](https://github.com/open-telemetry/opentelemetry-specification/issues/1844): Composite Sampler
- issue [2085](https://github.com/open-telemetry/opentelemetry-specification/issues/2085): Remote controlled sampling rules using attributes
- otep [213](https://github.com/open-telemetry/oteps/pull/213): Add Sampling SIG research notes
- issue [3205](https://github.com/open-telemetry/opentelemetry-specification/issues/3205): Equivalent of "logLevel" for spans
- discussion [3725](https://github.com/open-telemetry/opentelemetry-specification/discussions/3725): Overriding Decisions of the Built-in Trace Samplers via a "sample.priority" Like Span Attribute

A number of custom samplers are already available as independent contributions
([RuleBasedRoutingSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/samplers/src/main/java/io/opentelemetry/contrib/sampler/RuleBasedRoutingSampler.java),
[LinksBasedSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/samplers/src/main/java/io/opentelemetry/contrib/sampler/LinksBasedSampler.java),
and, of course, the latest and greatest [Consistent Probability Sampling](https://github.com/open-telemetry/opentelemetry-java-contrib/tree/main/consistent-sampling)) or just as ideas. They all share the same pain point, which is lack of easy to use configuration mechanism.

Even when the code for these samplers is available, their use is not very simple. In case of Java, they require writing a custom agent [extension](https://opentelemetry.io/docs/instrumentation/java/automatic/extensions/). This can become a hurdle, especially for the Open Telemetry users with no hands-on coding experience.

## The Goal

The goal of this proposal is to create an open-ended configuration schema which supports not only the current set of SDK standard samplers, but also non-standard ones, and even samplers that will be added in the future. Furthermore the samplers should be composable together, as it is often required by the users.

In contrast, the existing configuration schemas, such as [Jaeger sampling](https://www.jaegertracing.io/docs/1.50/sampling/#file-based-sampling-configuration), or Agent [OTEP 225 - Configuration proposal](https://github.com/open-telemetry/oteps/pull/225) address sampling configuration with a limited known set of samplers only.

## Use cases

- I want to use some of the samplers from the `opentelemetry-java-contrib` repository, but I do not want to build my own agent extension. I prefer to download one or more jarfiles containing the samplers and configure their use without writing any additional code.
- I want to apply a sampling strategy that combines different samplers depending on the span attributes, such as the URL of the incoming request, and I expect to update the configuration frequently, so I prefer that it is file based (rather than hardcoded), and better yet, applied dynamically.
- I want to write my own sampler with some unique logic, but I want to focus entirely on the sampling algorithm and avoid writing any boilerplate code for instantiating, configuring, and wrapping it up as an agent extension.

## The basics

It is assumed that the sampling configuration will be a YAML document, in most cases available as a file. Remote configuration remains an option, as well as dynamic changes to the configuration.

The configuration file will contain an actual configuration of the sampler to use. It may also optionally contain definitions of custom samplers.

```yaml
---
sampler_definitions:   # optional
  <CUSTOM_SAMPLER_DEF1>
  <CUSTOM_SAMPLER_DEF2>
   ...
  <CUSTOM_SAMPLER_DEF3>
sampler:
  <SAMPLER>
```

Each sampler is identified by a unique _sampler_key_, which takes a form of string.
A <SAMPLER> is described by the sampler_key followed by the sampler arguments, if applicable.


```yaml
  <SAMPLER_KEY>:
    <PARAM_NAME1>: <ARGUMENT_VALUE1>
    <PARAM_NAME2>: <ARGUMENT_VALUE2>
     ...
```

A sampler can be passed as an argument for another sampler:

```yaml
    parent_based:
      root:
        trace_id_ratio_based:
          ratio: 0.75
      remote_parent_sampled:
        always_on:
      remote_parent_not_sampled:
        always_off:
      local_parent_sampled:
        always_on:
      local_parent_not_sampled:
        always_off:
```

There's no limit on the depth of nesting samplers, which hopefully allows to create complex configurations addressing most of the sampling needs.

## Composite Samplers

New composite samplers are proposed to make group sampling decisions. They always ask the child samplers for their decisions, but eventually make the final call.

### Logical-Or Sampling

```yaml
   any_sampler:
     from:
      - <SAMPLER1>
      - <SAMPLER2>
      ...
      - <SAMPLERn>
```

The `any_sampler` is a composite sampler which takes a non-empty list of Samplers as the argument. When making a sampling decision, it goes through the list to find a sampler that decides to sample. If found, this sampling decision is final. If none of the samplers from the list wants to sample the span, the composite sampler drops the span.
If the first child which decided to sample modified the trace state, the effect of the modification remains in effect.

### Logical-And Sampling

```yaml
   all_samplers:
     from:
      - <SAMPLER1>
      - <SAMPLER2>
      ...
      - <SAMPLERn>
```

The `all_samplers` composite sampler takes a non-empty list of Samplers as the argument. When making a sampling decision, it goes through the list to find a sampler that decides not to sample. If found, the final sampling decision is not to sample. If all of the samplers from the list want to sample the span, the composite sampler samples the span.

If all of the child samplers agreed to sample, and some of them modified the trace state, the modifications are cumulative as performed in the given order. If the final decision is not to sample, the trace state remains unchanged.

### Rule based sampling

For rule-based sampling (e.g. decision depends on Span attributes), we need a `rule_based` sampler, which will take a list of Rules as an argument, and an optional Span Kind. Each rule will consist of a Predicate and a Sampler. For a sampling decision, if the Span kind matches the optionally specified kind, the list will be worked through in the declared order. If a Predicate holds, the corresponding Sampler will be called to make the final sampling decision. If the Span Kind does not match, or none of the Predicates evaluates to True, the final decision is not to sample.

Note: The `opentelemetry-java-contrib` repository contains [RuleBasedRoutingSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/samplers/src/main/java/io/opentelemetry/contrib/sampler/RuleBasedRoutingSampler.java), with similar but a bit different functionality.

```yaml
   rule_based:
     span_kind: SERVER | CONSUMER | ...   # optional
     rules:
       - <RULE1>
       - <RULE2>
        ...
```

where RULE is a pair of PREDICATE and SAMPLER:

```yaml
  predicate: PREDICATE
  sampler: SAMPLER
```

The Predicates represent logical expressions which can access Span Attributes (or anything else available when the sampling decision is to be taken), and perform tests on the accessible values.
For example, one can test if the target URL for a SERVER span matches a given pattern.

## Example

Let's assume that a user wants to configure head sampling as follows:

- for root spans:
  - drop all `/healthcheck` requests
  - capture all `/checkout` requests
  - capture 25% of all other requests
- for non-root spans
  - follow the parent sampling decision
  - however, capture all calls to service `/foo` (even if the trace will be incomplete)
- in any case, do not exceed 1000 spans/second

Such sampling requirements can be expressed as:

```yaml
sampler:
  all_samplers:
    from:
     - any_sampler:
         from:
          - parent_based:
              root:
                rule_based:
                  span_kind: SERVER
                  rules:
                   - predicate: http.target == /healthcheck
                     sampler: always_off
                   - predicate: http.target == /checkout
                     sampler: always_on
                   - predicate: true     # works as a fallback sampler
                     sampler: trace_id_ratio_based
                       ratio: 0.25
              remote_parent_sampled:
                always_on:
              remote_parent_not_sampled:
                always_off:
              local_parent_sampled:
                always_on:
              local_parent_not_sampled:
                always_off:
          - rule_based:
              span_kind: CLIENT
              rules:
               - predicate: http.url == /foo
                 sampler: always_on
     - rate_limiting:
         max_rate: 1000   # spans/second
```

The above example uses plain text representation of predicates. Actual representation for predicates is TBD. The example also assumes that a hypothetical `rate_limiting` sampler is available.

## Custom samplers

The YAML parser will have built-in knowledge about the standard SDK samplers as well as about the composite samplers described above.
However, one of the goals is to support also other samplers. Before such a custom sampler can be used in the configuration it needs to be introduced to the parser by providing the sampler definition.
Not all platforms are expected to accept custom samplers.

For example, let's assume that the `rate_limiting` sampler from the previous example is not known to the parser. It's definition can look like follows

```yaml
sampler_definitions:
  rate_limiting:   # this is the sampler key
    parameters:
     - max_rate: number  # spans per second
    implementation:
      class_name: java/io/opentelemetry/contrib/sampler/RateLimitingSampler
      instantiation: constructor
```

The platform dependant `implementation` part will specify the details allowing the parser to access the sampler code and create the sampler object.
There will be a need for each supported custom sampler to have a documented canonical way of instantiation or initialization, which takes a known list of parameters. The order of the parameters specified in the definition section will have to match exactly that list.

Specifying the parameter type within the sampler definition can be used to both guide the parser with the correct interpretation of the argument values, or to provide additional verification of the configuration correctness. Definition of more complex types remains to be designed. 

## Strong and Weak Typing

Constructing a sampler instance may require using some data types which are not strings, numbers, samplers, or lists of those types. Obviously, the beforementioned rule-based sampler needs to get `SpanKind` and `Rules`, which, in turn, will reference `Predicates`. The knowledge of these types can be built-in to the YAML parser, to ensure proper support.

If there's a need to support other complex types, the supported custom parsers may be required take maps of (key, value) pairs, which will be provided directly by the YAML parser. The samplers will be responsible for converting these values into a suitable internal representation.

## Suggested deployment pattern

Using the file-based configuration for sampling as described in this proposal does not require any changes to the OpenTelemetry Java Agent. The YAML file parser and the code to instantiate and configure requested samplers can be provided as an Extension jarfile (`config_based_sampler.jar` below). All samplers from the `opentelemetry-java-contrib` repository can be also made available as a separate jarfile (`all_samplers.jar` below).

```bash
$ java -javaagent:path/to/opentelemetry-javaagent.jar \
     -Dotel.javaagent.extensions=path/to/config_based_sampler.jar \
     -Dotel.sampling.config.file=path/to/sampling_config.yaml \
     -Dotel.sampling.classpath=path/to/all_samplers.jar \
     ... \
     -jar myapp.jar
```

If the feature proves popular, other options can be considered.

## Compatibility with existing standards and proposals

There already exists a [specification](https://github.com/open-telemetry/oteps/pull/225) for agent configuration which addresses sampling configuration. As long as the standard SDK samplers are used, the YAML representation of sampling configuration in that specification is identical to that of this proposal.
However, here the following extended capabilities are offered:

- additional composite samplers are built-in
- new (custom) samplers can be defined without adding any code to the SDK or SDK extension
- sampler configuration changes can be applied dynamically upon detecting configuration file modification, and supporting remote configuration via a wire protocol is an option

## Open Issues

- How to encode Predicates so they will be readable and yet powerful, and at the same time calculated efficiently?
- How to describe sampler parameter types to arrive at the _right_ balance of complexity and expressiveness?
- How to handle RecordOnly (_do not export_) sampling decisions by the composite samplers?

