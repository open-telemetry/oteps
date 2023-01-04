# OpenTelemetry Proposal: CI/CD Observability Support by OpenTelemetry

OpenTelemetry project can serve Continuous Integration & Continuous Delivery (CI/CD) observability use cases.

## Motivation

OpenTelemetry is already known for DevOps use cases around monitoring production systems and reducing mean time to identification and resolution/recovery (MTTI/MTTR).
However, the project can also bring value for pre-production DevOps use cases, by enabling monitoring of the Continuous Integration & Continuous Delivery (CI/CD) pipelines. CI/CD observability helps to reduce the Lead Time for Changes, which is another crucial [DORA metric](https://horovits.medium.com/improving-devops-performance-with-dora-metrics-918b9604f8e2) measuring how much time it takes a commit to get into production.
 
This enhancement will broaden the target audience of the project also to Release Engineering teams, and will unleash a whole new value proposition of OpenTelemetry in the software release process, in close collaboration and integration with the CI/CD ecosystem, specifications and tooling.

## Explanation

Lack of CI/CD observability results in unnecessarily long Lead Time for Changes, which is a crucial metric measuring how much time it takes a commit to get into production.

CI/CD tools today emit various telemetry data, whether logs, metrics or trace data to report on the release pipeline state, to help pinpoint flakyness, and accelerate root cause analysis of failures, whether stemming from the application code, a configuration, or from the CI/CD environment. However, these tools do not follow any particular standard, specification, or semantic conventions. This makes it hard to use observability tools for monitoring these pipelines. Some of these tools provide some observability visualization and analytics capabilities out of the box, but in addition to the tight coupling the offered capabilities are oftentime not enough, especially when one wishes to monitor aggregated information across different tools and different stages of the release process.

Some tools have started adopting OpenTelemetry, which is an important step in creating standardization. A good example is [Jenkins](https://github.com/jenkinsci/jenkins), a popular CI OSS project, which offers the [Jenkins OpenTelemetry plugin](https://plugins.jenkins.io/opentelemetry/) for emitting telemetry data in order to:
1. Visualize jobs and pipelines executions as distributed traces
2. Visualize Jenkins and pipeline health indicators
3. Troubleshoot Jenkins performances with distributed tracing of HTTPs requests

Building CI/CD observability involves four stages: Collect → Store → Visualize → Alert. OpenTelemetry provides a unified way for the first step, namely collecting and ingesting the telemetry data in an open and uniform manner. 

If you are a CI/CD tool builder, the specification and instrumentation will enable you to properly structure your telemetry, package and emit it over OTLP. OpenTelemetry specification will determine which data to collect, the semantic convention of the data, and how different signal types can be correlated based on that, to support downstream analytics of that data by various tools.

If you are an end user looking to gain observability over your pipelines, you will be able to collect OpenTelemetry-formatted telemetry using the OpenTelemetry Collector, ingest, process and then export to a rich ecosystem of observability analytics backend tools, independent of your CI/CD tools in use.     

Here are some examples of potential resulting observability visualization over popular backend tools such as Jaeger, Grafana and OpenSearch:

Monitoring Jenkins metrics for nodes, queues, jobs and executors with Grafana dashboard:
![Monitoring Jenkins metrics for nodes, queues, jobs and executors with Grafana dashboard](https://dytvr9ot2sszz.cloudfront.net/wp-content/uploads/2022/05/image6.png "Monitoring Jenkins metrics for nodes, queues, jobs and executors with Grafana dashboard")

Jenkins pipeline run visualized as a trace in the Timeline View in Jaeger UI:
![Jenkins pipeline run visualized as a trace in the Timeline View in Jaeger UI](https://dytvr9ot2sszz.cloudfront.net/wp-content/uploads/2022/05/image9.png)

OpenSearch dashboard for monitoring Jenkins pipelines:
![OpenSearch dashboard for monitoring Jenkins pipelines](https://dytvr9ot2sszz.cloudfront.net/wp-content/uploads/2022/05/image7.png)

For more examples, see [this article](https://logz.io/learn/cicd-observability-jenkins/) on CI/CD observability using currently available open source tools.

## Internal details

OpenTelemetry specification should be enhanced to cover semantics relevant to pipelines, such as the branch, build, step (ID, duration, status), commit SHA (or other UUID), run (type, status, duration). These should be geared for observability into issues in the released application code. 
In addition, oftentimes release issues are not code-based but rather environmental, stemming from issues in the build machines, garbage collection issues of the tool or even a malstructured pipeline step. In order to provide observability into CI/CD environment, especially one with distributed execution mechanism, there's need to monitor various entities such as nodes, queues, jobs and executors (using the Jenkins terms, other tools having respective equivalents, which the specification should abstract with the semantic convention).

The CDF (Continuous Delivery Foundation) has the Events Special Interest Group ([SIG Events](https://github.com/cdfoundation/sig-events)) which explores standardizing on CI/CD event to facilitate interoperability (it is a work-stream within the CDF SIG Interoperability.). The group is working on [CDEvents](https://cdevents.dev/), a standardized event protocol that caters for technology agnostic machine-to-machine communication in CI/CD systems. It makes sense to evaluate alignment between the standards.

OpenTelemetry instrumentation should then support in collecting and emitting the new data. 

OpenTelemetry Collector can then offer designated processors for these payloads, as well as new exporters for designated backend analytics tools, as such prove useful for release engineering needs beyond existing ecosystem.   

## Trade-offs and mitigations

Today’s tools already emit some telemetry, some of which may not easily fit into vendor-agnostic unified semantic conventions. These can be accommodated within extra baggage payload, which may be parsed on a tool-specific fashion. 

## Prior art and alternatives

Today’s tools already emit some telemetry, which can be visualized by the tool’s designated backend, or by general-purpose tools with custom-built dashboards and queries for this specific data. These, however, use proprietary specifications.

## Open questions

Open questions include:
- Which entity model should be supported to best represent CI/CD domain and pipelines?
- What are the common CI/CD workflows we aim to support? 
- What are the primary tools that should be supported with instrumentation in order to gain critical mass on CI/CD coverage?
- Is CDEvents a good fit of a specification to integrate with? what is the aligmment, overlap and gaps? and if so, how to establish the cross-foundation and cross-group collaboration in an effective manner?
- How can we bring the existing ecosystem players, both open source and others, to form a concensus and leverage existing knowledge and experience?
- Which receivers are needed beyond OTLP to support the use cases and workflows?
- Which exporters are needed to support common backends?
- Which processors are needed to support the defined workflows?

## Future possibilities

This OTEP will enable customized instrumentation options, as well as processing within the Collector, which will be designated to the capabilities and evolution of the CI/CD tools and domain. 
