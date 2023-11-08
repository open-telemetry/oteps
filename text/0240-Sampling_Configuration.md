# Sampling Configuration

An attempt to provide a framework for defining sampling configurations.

Calling this proposal half-baked would be too generous. At this time it just an idea, with many questions unanswered and without any working prototype. Its purpose is to start a discussion on the needs and wants the Open Telemetry community might have on the subject of sampling configuration and a possible way to accomplish them.

The focus is on head-based sampling, but a similar approach might be used for later sampling stages as well.

## Motivation

The need for sampling configuration has been explicitly or implicitly indicated in several discussions, some of them going back a number of years, see for example
issue [173](https://github.com/open-telemetry/opentelemetry-specification/issues/173),
issue [1060](https://github.com/open-telemetry/opentelemetry-java-instrumentation/issues/1060),
discussion [3725](https://github.com/open-telemetry/opentelemetry-specification/discussions/3725),
issue [3205](https://github.com/open-telemetry/opentelemetry-specification/issues/3205).

A number of custom samplers are already available as indpendent contributions or just as ideas. They all share the same pain point, which is lack of easy to use configuration mechanism. The goal of this proposal is to create a configuration schema which supports not only the current set of SDK standard samplers, but also non-standard ones, and even samplers that will be added in the future.

In contrast, the existing configuration schemas, such as [Jaeger sampling](https://www.jaegertracing.io/docs/1.50/sampling/#file-based-sampling-configuration), or Agent [Configuration proposal](https://github.com/open-telemetry/oteps/pull/225) address sampling configuration with a limited known set of samplers only.

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
The mechanism to map the sampler TYPE to the sampler implementation will be platform specific. For Java, the TYPE can be a simple class name, or a part of class name, and the implementing class can be found using reflection.

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
   spanKind: SERVER | CONSUMER | ...   # optional
   parameters:
    - - RULE1
      - RULE2
      ...
      - RULEn
    - FALLBACK_SAMPLER
```
where each RULE is
```yaml
    predicate: PREDICATE
    sampler: SAMPLER
```
The Predicates represent logical expressions which can access Span Attributes (or anything else available when the sampling decision is to be taken), and perform tests on the accessible values.
For example, one can test if the target URL for a SERVER span matches a given pattern.

## Open Issues
- How to encode Predicates so they will be readable and yet powerfull and efficiently calculated?
- How to handle RecordOnly (_do not export_) sampling decisions?
 
