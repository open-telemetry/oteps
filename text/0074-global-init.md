# Global SDK initialization

Specify the behavior of OpenTelemetry APIs and implementations at startup.

## Motivation

OpenTelemetry is designed with a separation between the API and the
SDK, allowing an application to configure and bind an SDK at runtime.
OpenTelemetry is designed to support "zero touch" instrumentation for
third party libraries through the use of a global Tracer and Meter
provider (factory) instances.  This OTEP proposes a specification for how to
initialize the global provider (factory) instances.

In some programming environments, it is possible for libraries of code
to auto-initialize static variables, allowing them to begin operation
concurrently with the main program, while initializing static program
state.  This presents a set of opposing requirements: (1) the API
supports a configurable SDK; (2) third party libraries can use
OpenTelemetry before the SDK is configured.

The current specification discusses a global provider (factory) for
named Tracers and Meters as well as a global Propagators instance for
injectors and extractors, but does not discuss how these are
initialized or whether their values can be modified during the process
lifetime.

Global variables face significant opposition from some developers,
which forces the question: "why support globals at all?".  In
languages with automatic dependency injection support, then
conceivably we do not need global variables.  In languges without
automatic dependency injection, without globals we could not have the
"zero touch" instrumentation [given as a requirement for the
project](https://github.com/open-telemetry/oteps/blob/master/text/0001-telemetry-without-manual-instrumentation.md).
If a third-party library is to be instrumented and integrated without
modification into an application, either dependency injection or
global variables are the solution.

Global variables **are a dangerous programming pattern**, but they
also enable easy integration in languages without automatic dependency
injection.  To address this risk, this proposal specifies strict
limits on their initialization.  We propose at-most-once
initialization for the three global variables in OpenTelemetry.
Specifically, this specification says that the global Tracer provider,
the global Meter provider, and the global Propagators instance can
only be initialized once per process lifetime (except possibly in
test-only scenarios).

This OTEP also specifies the behavior of the global instances when
they are used before the SDK is configured and installed, in case this
cannot be performed by automatic dependency injection.

## Explanation

There are two acceptable ways to provide default instances in
OpenTelemetry: (1) through dependency injection, (2) through global
variables initialized at most once.  The feasibility of each approach
varies by language.  The implementation MUST select one of the
following strategies.

### Service provider mechanism

Where the language provides a commonly accepted way to inject SDK
components, it should be preferred.  The Java SPI (Service Provider
Interface) supports loading and configuring the SDK before it is first
used.  This kind of support is the preferred choice in languages with
common support for automatic dependency injection.

In this case, there is no use-before-configuration question to
address, as there is in some languages with support for static
initialization.

### Explicit initializer mechanism

When it is not possible to ensure the SDK is installed and configured
before the API is first used, initializing the default, global SDK
instances becomes the user's responsibility.

Methods to set the global instances shall be independent, allowing
each SDK component to be intialized separately when the process
starts.  The methods shall be declared in a separate API package,
e.g., named `global`:

```golang
package global

import (
    "go.opentelemetry.io/otel/api/context/propagation"
    "go.opentelemetry.io/otel/api/metric"
    "go.opentelemetry.io/otel/api/trace"
)

// SetMeterProvider initializes the global Meter provider.  May only
// be called once per process lifetime.  Subsequent calls will panic.
//
// Prior to setting the global Meter provider, the default global
// Meter provider acts as a "forwarding" SDK.  The default global
// Meter provider will begin forwarding to the installed Meter
// provider once this is called.
func SetMeterProvider(metric.Provider) { ... }

// SetPropagators initializes the global Propagators instance.  May only
// be called once per process lifetime.  Subsequent calls will panic.
//
// Prior to setting the global Meter provider, the default global
// Propagators instance performs pass-through W3C Traceparent and
// Correlation-Context propagation.
func SetPropagators(propagation.Propagators) { ... }

// SetPropagators initializes the global Tracer provider.  May only
// be called once per process lifetime.  Subsequent calls will panic.
//
// Prior to setting the global Tracer provider, the default global
// Tracer provider acts as a "forwarding" SDK.  The default global
// Tracer provider will begin forwarding to the installed Tracer
// provider once this is called.
func SetTraceProvider(trace.Provider) { ... }
```

#### Requirements

We anticipate third party libraries using the global instances before
they are installed, and we wish for references obtained through these
instances to become functional once the corresponding implementation
is initialized.  The default instances returned by the global getters
for Tracer provider, Meter provider, and Propagators must "forward" to
the real SDK implementation once it is installed.

#### Tracer

Tracers obtained through the provider will become functional when the
user's Tracer SDK is installed as the global instance.

Spans started prior to installing the Tracer SDK will be No-op spans.
Installing a Tracer SDK after starting a span via the default global
instance does not change this behavior.

#### Meter

Meters obtained through the provider will become functional when the
users's Meter SDK is installed as the global instance.

Metric events will be dropped until the Meter SDK is installed.

#### Propagators

The default global Propagators instance will by default perform
pass-through W3C Traceparent and Correlation-Context propagation.

The default global Propagators instance will begin forwarding to the
user's Propagators when it is installed as the global instance.

## Trade-offs and mitigations

### Testing support

Testing should be performed without depending on the global SDK, if
possible.  A convenience method may be provided for tests to reset the
global state to the initial, default conditions.

### Efficiency concern

Since the global instances are required to begin working once the real
implementations are installed, there is some implied synchronization
overhead and cost.  This overhead SHOULD be minimal.

We recommend to explicitly install a No-op instance to lower the cost
of instrumentation when no SDK will be installed, as opposed to
leaving the default global instances in place, perpetually waiting to
begin forwarding.  True No-op instances will be slightly less
expensive than the default global instances.
