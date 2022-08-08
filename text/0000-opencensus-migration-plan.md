# OpenCensus To OpenTelemetry Migration Plan

Define the process of migrating applications, libraries, and agents from [OpenCensus](https://opencensus.io/) to
OpenTelemetry.

## Motivation

While OpenTelemetry has a specification for an
[OpenCensus trace bridge](https://github.com/open-telemetry/opentelemetry-specification/blob/v1.12.0/specification/compatibility/opencensus.md),
it lacks a broader context on how such bridges should be used to perform a migration from
OpenCensus to OpenTelemetry. Additionally, OpenTelemetry should provide guidance to libraries
currently using OpenCensus on how to migrate instrumentation to OpenTelemetry.

## Explanation

Migrating from OpenCensus to OpenTelemetry may require breaking changes to the telemetry produced
because of:

* Differences in semantic conventions for names and attributes
* Data model differences
* Instrumentation API feature differences
* Differences between equivalent OC and OTel exporters

Requirements:

* Give application developers control over _when_ breaking changes are introduced.
* No "long tail" of breaking changes to users when migrating.
* Make it possible to migrate instrumentation without breaking the resulting telemetry.
* Library maintainers should not need to maintain OpenCensus and OpenTelemetry instrumentation in parallel.

This proposed process groups most breaking changes into the installation of the bridge. This gives
users control over introducing the initial breaking change, and makes subsequent migrations of
instrumentation (including in third-party libraries) non-breaking.

## Migration Plans

### Migrating from the OC Agent and Protocol

Starting with a deployment of the OC agent, using the OC protocol, migrate by:

1. Deploy the OpenTelemetry collector with OpenCensus and OTLP receivers and equivalent processors and exporters.
2. **breaking**: For each workload sending the OC protocol, change to sending to the OpenTelemetry collector's OpenCensus receiver.
3. Remove the deployment of the OC Agent
4. For each workload, migrate the application from OpenCensus to OpenTelemetry, following the guidance below, and use the OTLP exporter

### Requirements for Metrics and Trace bridges

Bridges:

* MUST NOT require OpenCensus to depend on OpenTelemetry
* MUST require few or no changes to OpenCensus
* MUST convert OpenCensus semantic conventions to OpenTelemetry semantic conventions for metric and span names and attributes
* are NOT REQUIRED to support the entire OpenCensus instrumentation API Surface.

Trace Bridges:

* MUST support context propagation between OpenCensus and OpenTelemetry instrumentation
* MUST preseve span contents where an OpenTelemetry equivalent exists

Metric Bridges (to be proposed separately):

* MUST be compatible with push and pull exporters
* SHOULD support Gauges, Counters, Cumulative Histograms, and Summaries
* are NOT REQUIRED to support Gauge Histograms
* SHOULD support exemplars

### Migrating an Application using Bridges

Starting with an application using entirely OpenCensus instrumention for traces and metrics, it can be migrated by:

1. Migrate the exporters (SDK)
    1. Install the OpenTelemetry SDK, with an equivalent exporter
    2. Install equivalent OpenTelemetry resource detectors
    3. Install OpenTelemetry propagators for OpenCensus' `TextFormat` and `BinaryPropagator` formats.
    4. **breaking**: Install the metrics and trace bridges
    5. Remove initialization of OpenCensus exporters
2. Migrate the instrumentation (API)
    1. **breaking**: For OpenCensus instrumentation packages, migrate to the OpenTelemetry equivalent.
    2. For external dependencies, wait for it to migrate to OpenTelemetry, and update the dependency.
    3. For instrumentation which is part of the application, migrate it following the "library" guidance below.
3. Clean up: Remove the metrics and trace bridges

### Migrating Libraries using OC Instrumentation

#### In-place Migration

Libraries which want a simple migration can choose to replace instrumentation in-place.

Starting with a library using OpenCensus Instrumentation:

1. Annouce to users the library's transition from OpenCensus to OpenTelemetry, and recommend users adopt OC bridges.
2. Change unit tests to use the OC bridges, and use OpenTelemetry unit testing frameworks.
3. After a notification period, migrate instrumentation line-by-line to OpenTelemetry. The notification period should be long for popular libraries.
4. Remove the OC bridge from unit tests.

#### Migration via Config

Libraries which are eager to add native OpenTelemetry instrumentation sooner, and/or want to
provide extended support for OpenCensus may choose to provide users the option to use OpenCensus
instrumentation _or_ OpenTelemetry instrumentation.

Starting with a library using OpenCensus Instrumentation:

1. Add configuration allowing users to enable OpenTelemetry instrumentation and disable OpenCensus instrumentation.
2. Add OpenTelemetry instrumentation gated by the configuration.
3. After a notification period, switch to using OpenTelemetry instrumentation by default.
4. After a deprecation period, remove the option to use OpenCensus instrumentation.

## Trade-offs and mitigations

The primary downside to the proposed approach (of an in-place migration) is that it presents a
"cliff" to users when a library migrates from OpenCensus to OpenTelemetry, as OpenCensus
instrumentation can't be used after it is removed.

## Prior art and alternatives

### Alternative: Don't write bridges

This proposal lays out requirements for bridges to aid in the migration.

As an alternative, we could require libraries to implement the "Migration via Config" method of migration. For libraries which follow that guidance, users can migrate by:

1. Install the OpenTelemetry SDK, with an equivalent exporter
2. Install equivalent OpenTelemetry resource detectors
3. Install OpenTelemetry propagators for OpenCensus' `TextFormat` and `BinaryPropogator` formats.
4. **For each library, change configuration to use OpenTelemetry rather than OpenCensus.**
5. Remove initialization of OpenCensus exporters

This would not require users to install any bridge, and would cause breaking changes to occur when each library is updated.  However, this relies on libraries actively migrating to OpenTelemetry.  If any library does not follow the "Migration via Config" guidance, it would prevent the entire application from migrating.  This is because without a bridge, context propagation within the application between OpenCensus and OpenTelemetry libraries would not work properly.

### Alternative: Bridges don't map semantic conventions

This proposal groups breaking changes into the installation of the bridge by having the bridge map semantic conventions from OpenCensus to OpenTelemetry.

As an alternative, bridges could not perform that mapping, which would reduce (but not eliminate) breaking changes when installing the bridge. However, when the instrumentation is migrated from using OpenCensus conventions to OpenTelemetry conventions, it would be a breaking change.  Additionally, OpenTelemetry exporters may not work correctly when using the bridge if they rely on OpenTelemetry semantic conventions.

## Open questions

* Should we recommend one of the library migration strategies (in-place or config-based)?
* Should we require bridges to implement semantic convention mapping (MUST instead of SHOULD)?
* Is it reasonable to expect users to migrate (by installing a bridge) before libraries migrate?
