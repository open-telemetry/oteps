# Logs: Vocabulary

This documents defines the vocabulary for logs to be used across OpenTelemetry project.

## Motivation

We need a common language and common understanding of terms that we use to
avoid the chaos experienced by the builders of the Tower of Babel.

## Proposal

OpenTelemetry specification already contains a [vocabulary](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/overview.md)
for Traces, Metrics and other relevant concepts.

This proposal is to add the following concepts to the vocabulary.

### Log Record

A recording of an event. Typically the record includes a timestamp indicating
when the event happened as well as other data that describes what happened,
where it happened, etc.

### Embedded Log 

A log record embedded inside a [Span](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/api-tracing.md#span) 
object, in the [Events](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/api-tracing.md#add-events) list.

### Standalone Log

A log record that is not embedded inside a Span and is recorded elsewhere.

### Log Attributes

Key/value pairs contained in a Log Record.

### Structured Logs

Logs that are recorded in a format that has a well-defined structure that allows
to differentiate between different elements of a Log Record (e.g. the Timestamp,
the Attributes, etc). For example [Syslog, RFC5425](https://tools.ietf.org/html/rfc5424)
protocol defines `structured-data` format.

### Flat File Logs

Logs recorded in text files, often one line per log record (although multiline
records are possible too). There is no common industry agreement whether
logs written to text files in more structured formats (e.g. JSON files)
are considered Flat File Logs or no. Where such distinction is important it is
recommended to call it out specifically.
