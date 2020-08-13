# Remove Span.Status

**Author:** Nikita Salnikov-Tarnovski, Splunk

Remove `Span.status` and related APIs and conventions.

## Motivation

It is unclear how operations from various domains (http requests, database queries, file system operations etc)
should translate their statuses into canonical status codes of gRPC operations.

## Internal details

Several areas are affected by this change and will require the coordinate effort to make this happen.

First, we have to remove all references to `Status` from Trace API. This includes `Span.setStatus` method and whole `Status` section.

Second, semantic conventions for gRPC need to replace `Status` with attributes descriptions. See below.

Third, `Status` support will be removed from OTLP->Zipkin and OTLP->Jaeger protocol translations.

Fourth, `Span.Status` field in OTLP will be renamed to `DEPRECATED_Status` and should not be used anymore.
 
## Backward compatibility

After removal of `Span.Status` neither SDKs nor Collector will produce this field. But there is a possibility
that Collector will receive messages with this field filled from older clients. We don't want to loose this information
and thus cannot just ignore it.

The simplest way to translate between "old" OTLP with status and "new" OTLP without status is to convert status
to span attributes. We can reuse current convention for this:

|Status|Tag Key| Tag Value|
|--|--|--|
|Code | `otel.status_code` | Name of the code, for example: `OK` |
|Message *(optional)* | `otel.status_description` | `{message}` |