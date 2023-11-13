# Sampling Configuration

An attempt to provide a framework for defining sampling configurations.

Calling this proposal half-baked would be too generous. At this time it just a vision, with many questions unanswered and without any working prototype. Its purpose is to start a discussion on the needs and wants the Open Telemetry community might have on the subject of sampling configuration and a possible way to accomplish them.

The focus is on head-based sampling, but a similar approach might be used for later sampling stages as well. Most of the technical details presented here assume Java as the platform, but should be general enough to have corresponding concepts and solutions available for other platformstoo.

## Motivation

The need for sampling configuration has been explicitly or implicitly indicated in several discussions, some of them going back a number of years, see for example
- issue [173](https://github.com/open-telemetry/opentelemetry-specification/issues/173): Way to ignore healthcheck traces when using automatic tracer across all languages?
- issue [679](https://github.com/open-telemetry/opentelemetry-specification/issues/679): Configuring a sampler from an environment variable
- issue [1060](https://github.com/open-telemetry/opentelemetry-java-instrumentation/issues/1060): Exclude URLs from Tracing
- issue [1844](https://github.com/open-telemetry/opentelemetry-specification/issues/1844): Composite Sampler
- issue [2085](https://github.com/open-telemetry/opentelemetry-specification/issues/2085): Remote controlled sampling rules using attributes
- issue [3205](https://github.com/open-telemetry/opentelemetry-specification/issues/3205): Equivalent of "logLevel" for spans
- discussion [3725](https://github.com/open-telemetry/opentelemetry-specification/discussions/3725): Overriding Decisions of the Built-in Trace Samplers via a "sample.priority" Like Span Attribute


A number of custom samplers are already available as indpendent contributions
([RuleBasedRoutingSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/samplers/src/main/java/io/opentelemetry/contrib/sampler/RuleBasedRoutingSampler.java),
[LinksBasedSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/samplers/src/main/java/io/opentelemetry/contrib/sampler/LinksBasedSampler.java),
and, of course, the latest and greatest [Consistent Probability Sampling](https://github.com/open-telemetry/opentelemetry-java-contrib/tree/main/consistent-sampling)) or just as ideas. They all share the same pain point, which is lack of easy to use configuration mechanism. 

Even when the code for these samplers is available, their use is not very simple. In case of Java, they require writing a custom agent [extension](https://opentelemetry.io/docs/instrumentation/java/automatic/extensions/). This can become a hurdle, especially for the Open Telemetry users with no hands-on coding experience.

## The Goal

The goal of this proposal is to create an open-ended configuration schema which supports not only the current set of SDK standard samplers, but also non-standard ones, and even samplers that will be added in the future. Furthermore the samplers should be composable together, as it is often required by the users.

In contrast, the existing configuration schemas, such as [Jaeger sampling](https://www.jaegertracing.io/docs/1.50/sampling/#file-based-sampling-configuration), or Agent [OTEP 225 - Configuration proposal](https://github.com/open-telemetry/oteps/pull/225) address sampling configuration with a limited known set of samplers only.

## Use cases

- I want to use some of the samplers from the `opentelemetry-java-contrib` repository, but I do not want to build my own agent extension. I prefer to download one or more jarfiles containing the samplers and configure their use without writing any additional code.
- I want to apply a sampling strategy that combines different samplers depending on the span attributes, such as the URL of the incoming request, and I expect to update the configuration frequently, so I prefer that it is file-based (rather than hardcoded), and better yet, applied dynamically
 - I want to write my own sampler with some unique logic, but I want to focus entirely on the sampling algorithm and avoid writing any boilerplate code for instantiating, configuring, and wrapping it up as an agent extension

## The basics

It is assumed that the sampling configuration will be a YAML document, in most cases available as a file. Remote configuration remains an option, as well as dynamic changes to the configuration.

The configuration file will contain a definition of the sampler to use:

```yaml
---
sampler:
   SAMPLER
```
Additional information could be placed there as well. For example, for Java, there could be a location of a jarfile containing any non-standard samplers which have been configured.

A SAMPLER is described by a structure with two fields:
```yaml
  samplerType: TYPE
  parameters:    # an optional list of parameters for the sampler
   - PARAM1
   - PARAM2
   ...
   - PARAMn
```
The mechanism to map the sampler TYPE to the sampler implementation will be platform specific. For Java, the TYPE can be a simple class name, or a part of class name, and the implementing class can be found using reflection. Specifying a fully qualified class name should be an option.

There will be a need for each supported sampler to have a documented canonical way of instantiation or initialization, which takes a known list of parameters. The order of the parameters specified in the configuration file will have to match exactly that list.

A sampler can be passed as an argument for another sampler:
```yaml
   samplerType: ParentBased
   parameters:
    - samplerType: TraceIdRatioBased  # for root spans
      parameters:
       - 0.75
    - samplerType: AlwaysOn    # remote parent sampled
    - samplerType: AlwaysOff   # remote parent not sampled
    - samplerType: AlwaysOn    # local parent sampled
    - samplerType: AlwaysOff   # local parent not sampled
```

There's no limit on the depth of nesting samplers, which hopefully allows to create complex configurations addressing most of the sampling needs.

## Composite Samplers

New composite samplers are proposed to make group sampling decisions. They always ask the child samplers for their decisions, but eventually make the final call.

### Logical-Or Sampling
```yaml
   samplerType: AnyOf
   parameters:
    - - SAMPLER1
      - SAMPLER2
      ...
      - SAMPLERn
```
The AnyOf sampler takes a list of Samplers as the argument. When making a sampling decision, it goes through the list to find a sampler that decides to sample. If found, this sampling decision is final. If none of the samplers from the list wants to sample the span, the AnyOf sampler drops the span.
If the first child which decided to sample modified the trace state, the effect of the modification remains in effect.

### Logical-And Sampling
```yaml
   samplerType: AllOf
   parameters:
    - - SAMPLER1
      - SAMPLER2
      ...
      - SAMPLERn
```
The AllOf sampler takes a list of SAMPLERs as the argument. When making a sampling decision, it goes through the list to find a sampler that decides not to sample. If found, thie final sampling decision is not to sample. If all of the samplers from the list want to sample the span, the AllOf sampler samples the span.

If all of the child samplers agreed to sample, and some of them modified the trace state, the modifications are cumulative as performed in the given order. If the final decision is not to sample, the trace state remains unchanged.

### Rule based sampling

For rule-based sampling (e.g. decision depends on Span attributes), we need a RuleBasedSampler, which will take a list of Rules as an argument, an optional Span Kind and a fallback Sampler. Each rule will consist of a Predicate and a Sampler. For a sampling decision, if the Span kind matches the optionally specified kind, the list will be worked through in the declared order. If a Predicate holds, the corresponding Sampler will be called to make the final sampling decision. If the Span Kind does not match the final decision is not to sample. If the Span Kind matches, but none of the Predicates evaluates to True, the fallback sampler makes the final decision.

Note: The `opentelemetry-java-contrib` repository contains [RuleBasedRoutingSampler](https://github.com/open-telemetry/opentelemetry-java-contrib/blob/main/samplers/src/main/java/io/opentelemetry/contrib/sampler/RuleBasedRoutingSampler.java), with similar functionality.
```yaml
   samplerType: RuleBased
   parameters:
    - spanKind: SERVER | CONSUMER | ...   # optional
    - - RULE1
      - RULE2
      ...
      - RULEn
    - FALLBACK_SAMPLER
```
where RULE is an extension of SAMPLER, providing additional predicate 
```yaml
  predicate: PREDICATE
  samplerType: TYPE
  parameters:
   - PARAM1
   - PARAM2
   ...
   - PARAMn
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
  samplerType: AllOf
  parameters:
  - - samplerType: AnyOf
      parameters:
       - - samplerType: ParentBased
           parameters:
            - samplerType: RuleBased  # for root spans
              parameters:
               - spanKind: SERVER
               - - predicate: http.target == /healthcheck
                   samplerType: AlwaysOff
                 - predicate: http.target == /checkout
                   samplerType: AlwaysOn
               - samplerType: TraceIdRatioBased  # fallback sampler
                 parameters:
                  - 0.25
            - samplerType: AlwaysOn    # remote parent sampled
            - samplerType: AlwaysOff   # remote parent not sampled
            - samplerType: AlwaysOn    # local parent sampled
            - samplerType: AlwaysOff   # local parent not sampled
         - samplerType: RuleBased
           parameters:
            - spanKind: CLIENT
            - - predicate: http.url == /foo
                samplerType: AlwaysOn
            - samplerType: AlwaysOff   # fallback sampler
    - samplerType: RateLimiting
      parameters:
      - 1000   # spans/second
```

The above example uses plain text representation of predicates. Actual representation for predicates is TBD. The example also assumes that a hypothetical `RateLimitingSampler` is available.

## Strong and Weak Typing

Constructing a sampler instance may require using some data types which are not strings, numbers, samplers, or lists of those types. Obviously, the beforementioned rule-based sampler needs to get `Rules`, which, in turn, will reference `Predicates`. The knowledge of these types can be built-in to the YAML parser, to ensure proper support.

If there's a need to support other complex types, the supported parsers can take maps of (key, value) pairs, which will be provided directly by the YAML parser. The samplers will be responsible for converting these values into a suitable internal representation.

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

## Compatibility with existing standards and proposals

Generally, the standard SDK samplers as well as those from the `opentelemetry-java-contrib` repository, with few exceptions, are not prepared to be used directly by this proposal. Even the standard samplers do not have a uniform way of instantiation. For example `ParentBasedSampler` offers only a constructor, while `TraceIdRatioBasedSampler` is typically instantiated using static `create` method.

However, adding some uniformity there, as well as to the samplers from `opentelemetry-java-contrib` should be quite easy, and hopefully not very controversial. It is also possible to demand a uniform instantiation mechanism only for non-standard samplers; the knowledge about the standard samplers can be built-in.

Another point of contention is that the existing configuration practices and proposals (see [JSON Schema Definitions for OpenTelemetry File Configuration](https://github.com/open-telemetry/opentelemetry-configuration/blob/main/schema/tracer_provider.json)) expect very specific knowledge about the samplers, while in this proposal responsibility for matching the samplers' arguments with the samplers signature becomes user's responsibility. However, to decrease the risk of misconfiguration, this proposal can be extended by introducing another configuration section which would describe the number of arguments and their type for non-standard samplers, thus providing some level of consistency checking.

## Open Issues
- How to encode Predicates so they will be readable and yet powerfull and efficiently calculated?
- How to handle RecordOnly (_do not export_) sampling decisions?
 
