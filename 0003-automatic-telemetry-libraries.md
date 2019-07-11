# Automatic language/framework instrumentation in OpenTelemetry

_Cross-language requirements for automatically extracting portable telemetry data with minimal ("greybox") source code modification._

## Motivation

The purpose of OpenTelemetry is to make robust, portable telemetry a built-in feature of cloud-native software. The need to manually instrument services and wrap functions and request handlers results in friction for developers writing instrumentation. Thus, there may be a wide array of solutions for each language and each framework in the long term, but right now there's very little automatic framework instrumentation to make OpenTelemetry adoption "paste one import and add one line of code into the top of your main function".

### Why “cross-language/framework”?

There should be a _consistent_ way of adding and interacting with automatically created OpenTelemetry spans and metrics that is neither surprising to users of frameworks or languages. It should be easy for framework authors, as well as motivated users, to write automated instrumentation adapters for any framework that have similar installation methods (appropriate for the language), interoperate with other forms of instrumentation such as blackbox and whitebox, and are easy to maintain.

### Suggested reading

* https://docs.honeycomb.io/getting-data-in/beelines/ for the general philosophy on what Honeycomb would like to contribute and standardize.

Go:

* https://docs.honeycomb.io/getting-data-in/go/beeline/#wrappers-and-other-middleware which provides sets of wrappers to automatically instrument each handler and database call.
* https://github.com/open-telemetry/opentelemetry-go/blob/master/example/http/server/server.go which currently requires manually instrumenting _each_ handler.

Ruby:

* https://docs.honeycomb.io/getting-data-in/ruby/beeline/#instrumented-packages
* https://github.com/open-telemetry/opentelemetry-ruby (an empty repo)

NodeJS:

* https://docs.honeycomb.io/getting-data-in/javascript/beeline-nodejs/#instrumented-packages
* https://github.com/open-telemetry/opentelemetry-js (no automatic framework instrumentation)

## Proposed guidelines

### Requirements

Without further ado, here are a set of requirements for “official” OpenTelemetry efforts to accomplish greybox minimal-source-code-modification instrumentation (i.e., “OpenTelemetry framework adapters”) in any given language:
* No more than 50 lines of _manual_ source code modifications allowed regardless of the number of handlers/frameworks in use
* Licensing must be permissive (e.g., ASL / BSD)
* Packaging must allow vendors to “wrap” or repackage the portable (OpenTelemetry) library into a single asset that’s delivered to customers
    * That is, vendors do not want to require users to comprehend both an OpenTelemetry package and a vendor-specific package
* "Greybox" OpenTelemetry framework adapters must interoperate with both explicit, whitebox OpenTelemetry instrumentation and the “automatic” / zero-source-code-modification / blackbox instrumentation proposed in RFC 0002.
    * If the greybox instrumentation starts a Span, whitebox and blackbox instrumentation must be able to discover it as the active Span (and vice versa)
    * Relatedly, there also must be a way to discover and avoid potential conflicts/overlap/redundancy between explicit whitebox instrumentation and greybox/blackbox instrumentation of the same libraries/packages
        * That is, if a developer has already added the “official” greybox OpenTelemetry plugin for, say, gRPC, then when the blackbox instrumentation effort adds gRPC support, it should *not* “double-instrument” it and create a mess of extra spans/etc

* The code in the OpenTelemetry package must not take a hard dependency on any particular vendor/vendors (that sort of functionality should work via a plugin or registry mechanism)
    * Further, the code in the OpenTelemetry package must be isolated to avoid possible conflicts with the host application (e.g., shading in Java, etc)


### Nice-to-have properties

* Automated and modular testing of individual library/package plugins
    * Note that this also makes it easy to test against multiple different versions of any given library
* A fully pluggable architecture, where plugins can be registered at runtime without requiring changes to the central repo at github.com/open-telemetry
* Augmentation of greybox instrumentation by whitebox and blackbox instrumentation (or, perhaps, vice versa). That is, not only can the trace context be shared by these different flavors of instrumentation, but even things like in-flight Span objects can be shared and co-modified (e.g., to use runtime interposition to grab local variables and attach them to a manually-instrumented span).


## Trade-offs and mitigations

to be discussed!

## Prior art and alternatives

Honeycomb's beelines, which we propose to standardize.

Blackbox instrumentation (copied from 0002): There are many proprietary APM language agents – no need to list them all here. The Datadog APM "language agents" are notable in that they were conceived and written post-OpenTracing and thus have been built to interoperate with same. There are a number of mature JVM language agents that are pure OSS (e.g., [Glowroot](https://glowroot.org/)).
