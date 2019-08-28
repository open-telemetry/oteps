# API and SDK Compatibility & Deprecation Policy

**Status:** `proposed`

OpenTelemetry needs a clear policy on API and SDK compatibility for beta and major releases.

## Motivation

While we are rapidly evolving the SDK, exporter interfaces, and APIs, we need to balance usability with long term maintainability.
Our API users need to have predictability in the interfaces they will use.
We need to provide a predictable support timeframe for SDK releases, while enabling us to move onwards from decisions made early on.

## Explanation

Unless features are removed from the specification, any language's OpenTelemetry API `SHOULD` ensure that instrumentation calls against previous beta APIs still compile without modification against all future beta SDKs of that language.
This requirement can be implemented, for instance, with `@Deprecated` wrapper methods.

Users `SHOULD` compile their applications against a Beta SDK release at least every 4 weeks.
There `MUST` be an SDK release in each language at least every 2 weeks during the beta period.
Running a beta SDK older than 4 weeks `MAY` result in the SDK printing a warning message during startup and not propagating telemetry data to exporters.
Beta SDKs `MAY` break compatibility on context propagation, wire format, etc. with SDKs more than 4 weeks newer or older.

## Internal details

The `opentelemetry-beta-expiration-date` config flag is proposed (to be implemented in each language's SDK), which will be set by default to SDK release date plus 4 weeks.
If there are no changes to a language's SDK in a 2-week window prior to the SDK leaving beta, a new beta SDK will still be built updating the `opentelemetry-beta-expiration-date`.

If the current date is past `opentelemetry-beta-expiration-date`, OpenTelemetry constructors `MAY` create `NoOpSpan`s or other equivalents instead of providing real instrumentation.
If the current date is past `opentelemetry-beta-expiration-date`, an `ERROR` will be logged at least once (either on app startup, or when the date is crossed on a running process.
Operators may choose to override `opentelemetry-beta-expiration-date` if they accept the consequences of running an unsupported version.

## Trade-offs and mitigations

Alternatives considered:

* Have no firm deprecation policy.
** This ties our hands to supporting any beta release's internal binary format ~forever.
** This also doesn't formalize what our API support will be.
* Break things more vigorously.

## Prior art and alternatives

* Build Horizon at Google
* [Kubernetes Version Skew Policy](https://kubernetes.io/docs/setup/release/version-skew-policy/)

## Open questions

n/a

## Future possibilities

We should revisit this policy once we are out of beta.
