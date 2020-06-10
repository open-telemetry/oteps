# Conventions for Trace and Resource Association in Logs

To set out standards for correlating traces and resources in existing logs.

## Motivation

In order to correlate standalone log entries with the spans that were occurring 
when the log entry was created, we need to emit certain information in the logs 
to correlate the data entries. This document attempts to lay out the information 
that needs to be emitted, and sets forth several approaches for doing so.

## Explanation

Two types of correlation need to happen to tie a log record into its full 
execution context. The first is Request Correlation, which ties the record to 
operations that were occurring when the record was created. The second is 
Resource Correlation, which associates the entry with where the event occurred 
such as a host, a pod, or a virtual machine.

### Request Correlation

Trace correlation is achieved primarily with two values- a trace identifier and 
a span identifier. A trace may contain multiple spans, arranged as a tree, and 
may also contain links to other related spans. The combination of a trace 
identifier and a span identifier corresponds to a specific scope of work. In the 
tracing API, that scope can also contain attributes and events that describe 
what work the program being traced was currently performing.

In most cases when traces and logs are used in tandem, the attributes of the 
current scope do not need to be added to the log entry- doing so would duplicate 
transmission of those values. As one span is likely associated with several (or 
many) log entries, it is more efficient to transmit those attributes with the 
span once rather than many times with each log entry. As a result, for most 
purposes the goal is to set the three values of traceid, spanid, and traceflags 
into each log entry that should be associated with that span.

| Field      | Required | Format  
| :--------- | :------- | :--------------------------------------------------
| traceid    | Yes      | 16-byte numeric value as Base16-encoded string
| spanid     | Yes      | 8-byte value represented as Base16-encoded string
| traceflags | No       | 8-bit numeric value as a Base16-encoded string

### Resource Correlation
As important as what was occurring in a program’s execution is, where it was 
occurring is just as important. Resource correlation allows a log entry to be 
associated with an infrastructure resource, and in turn system and program 
metrics that describe the wider program state. The form that the resource takes 
is also more diverse than the tracing scope- the resource may be a pod running 
in a Kubernetes cluster, a virtual machine running in a cloud, a serverless 
lambda, or an old-school server sitting in a data center. An application 
environment may be an orchestration of multiple types of resource working 
together.

From a logging standpoint, a resource is also almost always a constant within an 
application process- a container may not have the same identifier on every run, 
but it does keep that identifier while it’s running, and that resource 
identifier is constant for every log entry created on that resource. As a 
result, resource correlation may not happen at the log entry level, so we may or 
may not put the resource correlation information in the log entry itself.

Resource information may be managed as part of the log ingestion process- for 
example a Docker logging driver will know which container logs came from. 
Full resource information may also not be available, as when logs are 
aggregated by syslog or a similar system that has less context available to it. 
As a result resource correlation information in the logs entries themselves 
should be considered optional. If included, the resource information should 
follow the semantic conventions for resources.

If the log entry is expressed in key-value pairs, any resource keys should be 
prepended with ‘resource’- for example, `resource.service.name=”shoppingcart”`. 
If the entry is expressed in JSON, the resource key-values should be placed in 
an object named “resource” at the top level of the object.

### Correlation Context

A Correlation Context is a set of key-value pairs that is shared amongst the 
spans of a distributed trace. Like span attributes, the correlation context may 
already be carried with span information, so duplicating this information may be 
redundant. In certain cases it may be important to associate this context with 
log entries. When the context is embedded in a log entry, the key-value
pairs should be placed in a ‘correlation’ namespace. Where key-value pairs are 
supported, embed the correlation key as “correlation.key_name”. In JSON or 
other formats that allow nested structures, the key-value pairs should be 
placed in an object named ‘correlation’ at the top of the object.


## Internal details

### Encoding
#### Key-Value Pairs
    2020-05-20 20:13:31 INFO Message logged. resource.hostname=myhost correlation.user=djones traceid=0354af75138b12921 spanid=14c902d73a traceflags=01
#### JSON
    {
      "time": "2020-05-20 20:13:31",
      "msg": "Message logged.",
      "level": "INFO",
      "correlation": {
        “user”: “djones”
      },
      "resource": {
        "hostname":"myhost"
      }
      "traceid": "0354af75138b12921",
      "spanid": "14c902d73a",
      "traceflags": "01"
    }

### Custom Format

Custom formats that don’t allow for automatic parsing of key-value pairs can be 
used, but they will require synchronization between the output format and the 
extraction mechanism. These types of extractions may have the advantage of being 
less verbose, but they also have the disadvantage of requiring setting up a 
custom extraction process, and may be more fragile. Since this approach is 
vendor-dependent, there is little guidance that can be provided by 
OpenTelemetry.

## Trade-offs and mitigations

As a convention, adherence to these standards may have some variation, and some
of this advice may be difficult to implement in certain circumstances.

## Prior art and alternatives

Elastic Common Schema [has standards](https://www.elastic.co/guide/en/ecs/current/ecs-tracing.html#ecs-tracing) 
for adding trace information to JSON log formats, but they do not support the 
full OpenTelemetry correlation model.

## Open questions


## Future possibilities

A stronger specification should be created for logs that are generated by
OpenTelemetry instrumentation and adapters that support conversion from and 
to OpenTelemetry's internal logging models. As that specification is created,
they should be kept in sync with these conventions.
