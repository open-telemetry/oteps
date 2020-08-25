# Remove Span.Status

**Author:** Nikita Salnikov-Tarnovski, Splunk

Remove `Span.status` and related APIs and conventions.

## Motivation

During numerous heated discussions in various pull requests (see
<https://github.com/open-telemetry/opentelemetry-specification/issues/599>
<https://github.com/open-telemetry/opentelemetry-specification/pull/697>
<https://github.com/open-telemetry/opentelemetry-specification/pull/427>
<https://github.com/open-telemetry/oteps/pull/69>
<https://github.com/open-telemetry/oteps/pull/86>
<https://github.com/open-telemetry/oteps/pull/123>)
and during Errors Working Group meetings a lot of concerns were raised about suitability of the current `Span.Status`
field encoded as Google RPC code for recording a result of arbitrary monitored operations from various domains
(http requests, database queries, file system operations etc). It is hard to define clear and umambigous mapping
from those domain to a fixed enum originated from RPC.

Therefore the proposal is to remove the `Span.Status` field from Trace API and OTLP.

## Internal details

Several areas are affected by this change and will require a coordinated effort to make this happen.

### Trace API Specification

* Remove all mentions of `Status` and `Set Status` API from Span section.
* Remove whole `Status` section.

### Trace SDK Specification

No changes needed.

### Trace SDK exporters

Remove translation rules pertaining to `Span.Status`.

### Span semantic conventions

* From HTTP conventions, remove `Status` section.
* In gRPC conventions, replace the requirement to set span status with the requirement to set span attribute `grpc.status`
to the canonical status code of the gRPC response.

### OTLP data definition

Rename `Span.status` field to `Span.deprecated_status` field.

### Language API and SDK

* Every language must remove status related code from their implementations of Trace API and SDK.
* Every language must remove all code in their span exporters which serializes span status.
* All instrumentation libraries, both manual and automated, must remove all calls to the removed `Span.setStatus` API.
* All instrumentation libraries, both manual and automated, must adapt new semantic conventions for gRPC spans.

### Collector

Collector will have a transition period during which it has a special handling both for old data,
which has `Span.Status` present, and new data without Status.
This will allow for older pipelines, still relying on `Span.Status` to continue to work with as little disruption as possible.
The exact duration of this transition period will be left for Collector's maintainers to decide.

* If incoming OTLP data has `Span.Status` present, that will NOT be removed and will be passed on as-is.
* If incoming OTLP data has `Span.Status` present and it is not equal to `OK`,
Collector will translate that to temporary semantic attributes as follows:

|Status|Tag Key| Tag Value|
|--|--|--|
|StatusCanonicalCode | `otel.deprecated_status_code` | Name of the code|
|Message *(optional)* | `otel.deprecated_status_message` | `{message}`|

* When receiving data in Zipkin format, no special handling will take place.
All Zipkin tags present will be translated verbatim to OTLP attributes.
* When exporting data into Zipkin format, no special handling will take place.
All OTLP attributes will be translated verbatim to Zipkin tags.
`Span.Status` if present, will NOT be translated.
* When receiving data in Jaeger format, no special handling will take place.
Jaeger `error` tag will be translated in the usual way to an OTLP error attribute
(see [Open questions](#open-questions)).
* When exporting data into Jaeger format, if a new temporary attribute `otel.deprecated_status_code` is present
and does indicate a non-OK status,
then Jaeger `error` tag will be set to `true` (see [Open questions](#open-questions)).

## Open questions

Neither OpenTelemetry Specification nor Protocol currently assigns any special meaning to the `error` attribute mentioned above
nor do they have any comparable functionality.
There are still ongoing discussions about that and this issue will be addressed later and separately.
