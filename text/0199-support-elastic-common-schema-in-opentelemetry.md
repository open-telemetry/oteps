# Support Elastic Common Schema in OpenTelemetry

Collaborators: Alexander Wert (Elastic), Jamie Hynds (Elastic), Alolita Sharma (OTel GC, Apple), Christian Beedgen (Sumo), Jonah Kowall (Logz.io), Tigran N. (Splunk), Jamie Hynds (Elastic) .

## Introduction

This proposal is to add support for the Elastic Common Schema (ECS) in the OpenTelemetry specification and provide full interoperability for ECS in OpenTelemetry component implementations. We propose to implement this support by enriching OpenTelemetry Semantic Conventions with ECS fields through the donation of complete [ECS FieldSets](https://www.elastic.co/guide/en/ecs/current/ecs-field-reference.html#ecs-fieldsets). The goal is to contribute ECS into OTel Semantic Conventions.

## Proposed process to contribute ECS to OpenTelemetry Semantic Conventions

This constitutes a contribution of ECS fieldsets into OpenTelemetry Semantic Conventions whereby both ECS and Opentelemetry will continue to maintain their own specification and governance bodies. This will minimize breaking changes in both ecosystems. Although existing separately, this OTEP establishes both specifications as extremely closely related specifications.

We propose to integrate the contribution of ECS into OpenTelemetry Semantic Conventions as part of a joint review process of existing Semantic Conventions:
1. Open issue to discuss prioritization of [ECS fieldsets](https://www.elastic.co/guide/en/ecs/current/ecs-field-reference.html#ecs-fieldsets) contribution
2. Open issue for each individual ECS fieldsets for discussion
   * See this [Geo fieldset example issue](https://github.com/open-telemetry/opentelemetry-specification/issues/2834)
3. Open draft PR’s to drive implementation (see [Geo example](https://github.com/open-telemetry/opentelemetry-specification/pull/2835))
   * Identifying overlaps/breaking changes.
   * Map ECS data types to OpenTelemetry data types.
   * Identify high cardinality fields.
   * Include 2-3 use cases of how these fieldsets are used today as part of the documentation

### Dealing with conflicts
There are fields or fieldsets that conflict between ECS and OpenTelemetry Semantic Conventions.
In some cases it might be possible and reasonable to introduce breaking changes on ECS or OpenTelemetry Semantic Conventions to resolve the conflict.
However, there will also be fields for which it will not be feasible to introduce breaking-changes, both on the ECS side and on the OpenTelemetry Semantic Conventions side.
For these cases, we propose to introduce OpenTelemetry Collector Processors that would provide automated mapping between ECS-specific fields and OpenTelemetry Semantic Conventions atributes.

### Schema evolution
As described above, the contribution of ECS into OpenTelemetry Semantic Conventions will result in two highly aligned schemas, yet, both specifications will be separate.
Once the initial contribution of ECS into OpenTelemetry Semantic Conventions is completed, an important goal is to maintain an aligned evolution process for both specifications to avoid drift over time.
Therefore, we propose to establish and collaborate on a bidirectional synchronization processes. Concretely this would mean:
* An open invitation to OpenTelemetry to participate and review ECS RFC’s
* This could be facilitated through automated github review requests.
* ECS’s RFC process would explicitly call contribution back to OpenTelemetry out as a necessary step.
* Vice versa, new contributions to the OpenTelemetry Semantic Conventions will be reviewed and adopted by ECS.
* Exploring open source community generated conversion tooling as needed.

Acceptance of this OTEP registers the intent to kick off this process.

## Motivation

Adding the Elastic Common Schema (ECS) to OpenTelemetry (OTel) is a great way to accelerate the integration of vendor-created logging and OTel component logs (ie OTel Collector Logs Receivers). The goal is to define vendor neutral semantic conventions for most popular types of systems and support vendor-created or open-source components (for example HTTP access logs, network logs, system access/authentication logs) extending OTel correlation to these new signals.

Adding the coverage of ECS to OTel would provide guidance to authors of OpenTelemetry Collector Logs Receivers and help establish the OTel Collector as a de facto standard log collector with a well-defined schema to allow for richer data definition.

In addition to the use case of structured logs, the maturity of ECS for SIEM (Security Information and Event Management) is a great opportunity for OpenTelemetry to expand its scope to the security use cases.

Another significant use case is providing first-class support for Kubernetes application logs, system logs as well as application introspection events. We would also like to see support for structured events (e.g. [k8seventsreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/k8seventsreceiver)) and using 'content-type' to identify event types.

We'd like to see different categories of structured logs being well-supported in the [OTel Log Data Model](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md), presumably through [semantic conventions for log attributes](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#field-attributes). For example, NGINX access logs and Apache access logs should be processed the same way as structured logs. This would help in trace and metric correlation with such log data as well as it would help grow the ecosystem of curated UIs provided by observability backends and monitoring dashboards (e.g. one single HTTP access log dashboard benefiting Apache httpd, Nginx, and HAProxy).

## Customer Motivation

Adoption of OTel logs will accelerate greatly if ECS is leveraged as the common standard, using this basis for normalization. OTel Logs adoption will be accelerated by this support. For example, ECS can provide the unified structured format for handling vendor-generated along with open source logs.

Customers will benefit from turnkey logs integrations that will be fully recognized by OTel-compatible observability products and services.

OpenTelemetry logging is today mostly structured when instrumentation libraries are used. However, most of the logs which exist today are generated by software, hardware, and cloud services which the user cannot control. OpenTelemetry provides a limited set of "reference integrations" to structure logs: primarily the [OpenTelemetry Collector Kubernetes Events Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/k8seventsreceiver) and an example of a regexp based parsing of Tomcat access logs with OpenTelemetry Collector File Receiver ([here](https://github.com/open-telemetry/opentelemetry-log-collection/blob/30807b96b2f0771e7d11452ebf98fe5e211ed6d7/examples/tomcat/config.yaml#L20)).
By expanding the OTel semantic conventions with further namespaces already defined in ECS, a broader coverage of such mappings from different sources can be defined and implemented in the OTel collector.
This, for example, includes logs from network appliances (mapping to the `network` and `interface` namespaces in ECS).

The semantic conventions of a log are a challenge. What is a specific component defined in a log and how does it relate to other logs which have the same semantic component defined differently. ECS has already done some heavy-lifting on defining a unified set of semantic conventions which can be adopted in OTel.

OpenTelemetry has the potential to grow exponentially if the data from these other services can be correlated with instrumented code and components. In order to do this, industry stakeholders should leverage a common and standard logging data model which allows for the mapping of these different data types. The OpenTelemetry data protocol can provide this interoperable open standard. This unlocks countless use cases, and ensures that OpenTelemetry can work with other technologies which are not OpenTelemetry compliant.

## Background

### What is ECS?

The [Elastic Common Schema (ECS)](https://github.com/elastic/ecs) is an open source specification, developed with support from Elastic's user community. ECS defines a common set of fields to be used when storing data in Elasticsearch, such as logs, metrics, and security and audit events. The goal of ECS is to enable and encourage users of Elasticsearch to normalize their event data, so that they can better analyze, visualize, and correlate the data represented in their events. Learn more at: [https://www.elastic.co/guide/en/ecs/current/ecs-reference.html](https://www.elastic.co/guide/en/ecs/current/ecs-reference.html)

The coverage of ECS is very broad including in depth support for logs, security, and network events such as "[logs.* fields](https://www.elastic.co/guide/en/ecs/current/ecs-log.html)" , "[geo.* fields](https://www.elastic.co/guide/en/ecs/current/ecs-geo.html)", "[tls.* fields](https://www.elastic.co/guide/en/ecs/current/ecs-tls.html)", "[dns.* fields](https://www.elastic.co/guide/en/ecs/current/ecs-dns.html)", or "[vulnerability.* fields](https://www.elastic.co/guide/en/ecs/current/ecs-vulnerability.html)".

ECS has the following guiding principles:

* ECS favors human readability in order to enable broader adoption as many fields can be understood without having to read up their meaning in the reference,
* ECS events include metadata to enable correlations across any dimension (host, data center, docker image, ip address...),
  * ECS does not differentiate the metadata fields that are specific to each event of the event source and the metadata that is shared by a¬ll the events of the source in the way OTel does, which differentiates between Resource Attributes and Log/Span/Metrics Attributes,
* ECS groups fields in namespaces in order to:
  * Offer consistency and readability,
  * Enable reusability of namespaces in different contexts,
    * For example, the "geo" namespace is nested in the "client.geo", "destination.geo", "host.geo" or "threat.indicator.geo" namespaces
  * Enable extensibility by adding fields to namespaces and adding new namespaces,
  * Prevent field name conflicts
* ECS covers a broad spectrum of events with 40+ namespaces including detailed coverage of security and network events. It's much broader than simple logging use cases.

### Example of a log message structured with ECS: NGINX access logs

Example of a Nginx Access Log entry structured with ECS

```json
{
   "@timestamp":"2020-03-25T09:51:23.000Z",
   "client":{
      "ip":"10.42.42.42"
   },
   "http":{
      "request":{
         "referrer":"-",
         "method":"GET"
      },
      "response":{
         "status_code":200,
         "body":{
            "bytes":2571
         }
      },
      "version":"1.1"
   },
   "url":{
      "path":"/blog"
   },
   "user_agent":{
      "original":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36",
      "os":{
         "name":"Mac OS X",
         "version":"10.14.0",
         "full":"Mac OS X 10.14.0"
      },
      "name":"Chrome",
      "device":{
         "name":"Other"
      },
      "version":"70.0.3538.102"
   },
   "log":{
      "file":{
         "path":"/var/log/nginx/access.log"
      },
      "offset":33800
   },
   "host": {
     "hostname": "cyrille-laptop.home",
     "os": {
       "build": "19D76",
       "kernel": "19.3.0",
       "name": "Mac OS X",
       "family": "darwin",
       "version": "10.15.3",
       "platform": "darwin"
     },
     "name": "cyrille-laptop.home",
     "id": "04A12D9F-C409-5352-B238-99EA58CAC285",
     "architecture": "x86_64"
   }
}
```

## Comparison between OpenTelemetry Semantic Conventions for logs and ECS

## Principles

| Description | [OTel Logs and Event Record](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#log-and-event-record-definition) | [Elastic Common Schema (ECS)](https://www.elastic.co/guide/en/ecs/current/ecs-reference.html) |
|-------------|-------------|--------|
| Metadata shared by all the Log Messages / Spans / Metrics of an application instance | Resource Attributes | ECS fields |
| Metadata specific to each Log Message / Span / Metric data point | Attributes | ECS Fields |
| Message of log events | Body | [message field](https://www.elastic.co/guide/en/ecs/current/ecs-base.html#field-message) |
| Naming convention | Dotted names | Dotted names |
| Reusability of namespaces | Namespaces are intended to be composed | Namespaces are intended to be composed |
| Extensibility | Attributes can be extended by either adding a user defined field to an existing namespaces or introducing new namespaces. | Extra attributes can be added in each namespace and users can create their own namespaces |

## Data Types

| Category | <a href="https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#log-and-event-record-definition">OTel Logs and Event Record</a> (all or a subset of <a href="https://developers.google.com/protocol-buffers/docs/proto3">GRPC data types</a>) | <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/mapping-types.html">ECS Data Types</a> |
|---|---|---|
| Text | string | <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/text.html#text-field-type">text</a>, <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/text.html#match-only-text-field-type">match_only_text</a>, <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/keyword.html#keyword-field-type">keyword</a> <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/keyword.html#constant-keyword-field-type">constant_keyword</a>, <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/keyword.html#wildcard-field-type">wildcard</a> |
| Dates | uint64 nanoseconds since Unix epoch | <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/date.html">date</a>, <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/date_nanos.html">date_nanos</a> |
| Numbers | number | <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/number.html">long</a>, <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/number.html">double</a>, <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/number.html">scaled_float</a>, <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/boolean.html">boolean</a>… |
| Objects | uint32, uint64… | <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/object.html">object</a> (JSON object), <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/flattened.html">flattened</a> (An entire JSON object as a single field value) |
| Structured Objects | No complex semantic data type specified for the moment (e.g. string is being used for ip addresses rather than having an "ip" data structure in OTel). <br/> Note that OTel supports arrays and nested objects. | <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/ip.html">ip</a>, <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/geo-point.html">geo_point</a>, <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/geo-shape.html">geo_shape</a>, <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/version.html">version</a>, <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/range.html">long_range</a>, <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/range.html">date_range</a>, <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/range.html">ip_range</a> |
| Binary data | Byte sequence | <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/binary.html">binary</a> |

## Known Differences

Some differences exist on fields that are both defined in OpenTelemetry Semantic Conventions and in ECS. In this case, it would make sense for overlapping ECS fields to not be integrated in the new specification.

<!-- 
As the markdown code of the tables is hard to read and maintain with very long lines, we experiment maintaining this one as an HTML table 
-->

<table>
  <tr>
   <td><strong><a href="https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#log-and-event-record-definition">OTel Logs and Event Record</a></strong>
   </td>
   <td><strong><a href="https://www.elastic.co/guide/en/ecs/current/ecs-reference.html">Elastic Common Schema (ECS)</a></strong>
   </td>
   <td><strong>Description</strong>
   </td>
  </tr>
  <tr>
   <td><a href="https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#log-and-event-record-definition">Timestamp</a> (uint64 nanoseconds since Unix epoch)
   </td>
   <td><a href="https://www.elastic.co/guide/en/ecs/current/ecs-base.html#field-timestamp">@timestamp</a> (date)
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td><a href="https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#log-and-event-record-definition">TraceId</a> (byte sequence), <a href="https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#log-and-event-record-definition">SpanId</a> (byte sequence)
   </td>
   <td><a href="https://www.elastic.co/guide/en/ecs/current/ecs-tracing.html#field-trace-id">trace.id</a> (keyword), <a href="https://www.elastic.co/guide/en/ecs/current/ecs-tracing.html#field-trace-id">span.id</a> (keyword)
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td>N/A
   </td>
   <td><a href="https://www.elastic.co/guide/en/ecs/current/ecs-tracing.html#field-transaction-id">Transaction.id</a> (keyword)
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td><a href="https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#log-and-event-record-definition">SeverityText</a> (string)
   </td>
   <td><a href="https://www.elastic.co/guide/en/ecs/current/ecs-log.html#field-log-syslog-severity-name">log.syslog.severity.name</a> (keyword), <a href="https://www.elastic.co/guide/en/ecs/current/ecs-log.html#field-log-level">log.level</a> (keyword)
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td><a href="https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#log-and-event-record-definition">SeverityNumber</a> (number)
   </td>
   <td><a href="https://www.elastic.co/guide/en/ecs/current/ecs-log.html#field-log-syslog-severity-code">log.syslog.severity.code</a>
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td><a href="https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#log-and-event-record-definition">Body</a> (any)
   </td>
   <td><a href="https://www.elastic.co/guide/en/ecs/current/ecs-base.html#field-message">message</a> (match_only_text)
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td>process.cpu.load (not specified but collected by OTel Collector)
<br/>
<a href="https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/metrics/semantic_conventions/process-metrics.md">process.cpu.time</a> (async counter)
<br/>
<a href="https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/metrics/semantic_conventions/system-metrics.md">system.cpu.utilization</a>
   </td>
   <td><a href="https://www.elastic.co/guide/en/ecs/current/ecs-host.html#field-host-cpu-usage">host.cpu.usage</a> (scaled_float) with a slightly different measurement than what OTel metrics measure
   </td>
   <td>Note that most metrics have slightly different names and semantics between ECS and OpenTelemetry
   </td>
  </tr>
</table>

## How would OpenTelemetry users practically use the new OpenTelemetry Semantic Conventions Attributes brought by ECS
The concrete usage of ECS-enriched OpenTelemetry Semantic Conventions Attributes depends on the use case and the fiedset.
In general, OpenTelemetry users would transparently upgrade to ECS and benefit from the alignment of attributes for new use cases.
The main goal of this work is to enable producers of OpenTelemetry signals (collectors/exporters) to create enriched uniform signals for existing and new use cases.
The uniformity allows for easier correlation between signals originating from different producers. The richness ensures more options for Root Cause Analysis, correlation and reporting.

While ECS covers many different use cases and scenarios, in the following, we outline two examples:

### Example: OpenTelemetry Collector Receiver to collect the access logs of a web server

The author of the "OTel Collector Access logs file receiver for web server XXX" would find in the OTel Semantic Convention specifications all
the guidance to map the fields of the web server logs, not only the attributes that the OTel Semantic Conventions has specified today for
[HTTP calls](https://github.com/open-telemetry/opentelemetry-specification/blob/v1.9.0/specification/trace/semantic_conventions/http.md),
but also attributes for the [User Agent](https://www.elastic.co/guide/en/ecs/current/ecs-user_agent.html)
or the [Geo Data](https://www.elastic.co/guide/en/ecs/current/ecs-geo.html).

This completeness of the mapping will help the author of the integration to produce OTel Log messages that will be compatible with access logs
of other web components (web servers, load balancers, L7 firewalls...) allowing turnkey integration with observability solutions
and enabling richer correlations.

### Other Examples
- [Logs with sessions (VPN Logs, Network Access Sessions, RUM sessions, etc.)](https://github.com/elastic/ecs/blob/main/rfcs/text/0004-session.md#usage)
- [Logs from systems processing files](https://www.elastic.co/guide/en/ecs/current/ecs-file.html)

## Alternatives / Discussion

### Prometheus Naming Conventions

Prometheus is a de facto standard for observability metrics and OpenTelemetry already provides full interoperability with the Prometheus ecosystem.

It would be useful to get interoperability between metrics collected by [official Prometheus exporters](https://prometheus.io/docs/instrumenting/exporters/) (e.g. the [Node/system metrics exporter](https://github.com/prometheus/node_exporter) or the [MySQL server exporter](https://github.com/prometheus/mysqld_exporter)) and their equivalent OpenTelemetry Collector receivers (e.g. OTel Collector [Host Metrics Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver) or [MySQL Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/mysqlreceiver)).

Note that one of the challenges with Prometheus metrics naming conventions is that these are implicit conventions defined by each integration author which doesn't enable correlation due to the lack of consistency across integrations. For example, this inconsistency increases the complexity that an end-user has to deal with when configuring and monitoring alerts.

Prometheus' conventions are restricted to the style of the name of the metrics (see [Prometheus Metric and label naming](https://prometheus.io/docs/practices/naming/)) but don't specify unified metric names.

## Other areas that need to be addressed by OTel (the project)

Some areas that need to be addressed in the long run as ECS is integrated into OTel include defining the innovation process,
providing maintainer rights to Elastic contributors who maintain ECS today, ensuring the OTel specification incorporates the changes to
accommodate ECS, and a process for handling breaking changes if any (the proposal
[Define semantic conventions and instrumentation stability #2180](https://github.com/open-telemetry/opentelemetry-specification/pull/2180)
should tackle this point). Also, migration of existing naming (e.g. Prometheus exporter) to standardized convention (see
[Semantic Conventions for System Metrics](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/metrics/semantic_conventions/system-metrics.md) ,
[Semantic Conventions for OS Process Metrics](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/metrics/semantic_conventions/process-metrics.md)).
