# Standard names for system/runtime metrics

This OTEP proposes a set of standard names and semantic conventions for
common system/runtime metrics collected by OpenTelemetry. The metric names
and labels proposed here are common across the supported operating systems
and runtime environments. This OTEP also includes semantic conventions for
metrics specific to a particular OS or runtime.

This OTEP is largely based on the existing implementation in the
OpenTelemetry Collector's [Host Metrics
Receiver](https://github.com/open-telemetry/opentelemetry-collector/tree/1ad767e62f3dff6f62f32c7360b6fefe0fbf32ff/receiver/hostmetricsreceiver).
The proposed names aim to make system/runtime metrics easily discoverable and
unambiguous. See [OTEP
#108](https://github.com/open-telemetry/oteps/pull/108/files) for additional
motivation.

## Trade-offs and mitigations

TODO: find a better way to phrase this
For a given metric, there is potentially a trade-off between the breadth of its scope and its meaning.

## Prior art
There is an existing metric naming proposal
[here](https://docs.google.com/spreadsheets/d/1WlStcUe2eQoN1y_UF7TOd6Sw7aV_U0lFcLk5kBNxPsY/edit#gid=0).
In addition, there are already a few implementations of system and/or runtime
metric collection in OpenTelemetry:

- **Collector**
  * [Host Metrics
  Receiver](https://github.com/open-telemetry/opentelemetry-collector/tree/1ad767e62f3dff6f62f32c7360b6fefe0fbf32ff/receiver/hostmetricsreceiver)
  generates metrics about the host system when run as an agent.
  * Currently is the most comprehensive implementation.
  * Collects system metrics for CPU, memory, swap, disks, filesystems, network,
and load.
  * Collects process metrics for CPU, memory, and disk usage.
  * Makes good use of labels rather than defining individual metrics.
  * [Overview of collected metrics](https://docs.google.com/spreadsheets/d/11qSmzD9e7PnzaJPYRFdkkKbjTLrAKmvyQpjBjpJsR2s).

- **Go**
  * Go [has instrumentation](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/master/instrumentation/runtime) to collect runtime metrics for GC, heap use, and goroutines. 
  * This package does not export metrics with labels, instead exporting individual metrics.
  * [Overview of collected metrics](https://docs.google.com/spreadsheets/d/1YqK66oWcW1iVo1fHwEb7JZWYGI_Dv3Bu3DQ6st-JMIY/edit#gid=0).
- **Python**
  * Python [has instrumentation](https://github.com/open-telemetry/opentelemetry-python/tree/master/ext/opentelemetry-ext-system-metrics) to collect some system and runtime metrics.
  * Collects system CPU, memory, and network metrics
  * Collects runtime CPU, memory, and GC metrics.
  * Makes use of labels, similar to the Collector.
  * [Overview of collected metrics](https://docs.google.com/spreadsheets/d/1YqK66oWcW1iVo1fHwEb7JZWYGI_Dv3Bu3DQ6st-JMIY/edit#gid=0).


## Open questions

- Should the individual runtimes have their specific naming conventions in the spec?

---------
## System

**prefix:** `system.`

**Description:** System level metrics

|Name            |Units  |Instrument       |Label Key |Label Values|Linux/Windows Support|
|----------------|-------|-----------------|----------|------------|---------------------|
|system.cpu.time |seconds|SumObserver      |state     |idle        |Both                 |
|                |       |                 |          |user        |Both                 |
|                |       |                 |          |system      |Both                 |
|                |       |                 |          |interrupt   |Both                 |
|                |       |                 |          |nice        |Linux                |
|                |       |                 |          |softirq     |Linux                |
|                |       |                 |          |steal       |Linux                |
|                |       |                 |          |wait        |Linux                |
|system.cpu.usage|%      |UpDownSumObserver|(as above)|(as above)  |(as above)           |

## Telemetry SDK

**type:** `telemetry.sdk`

**Description:** The telemetry SDK used to capture data recorded by the instrumentation libraries.

The default OpenTelemetry SDK provided by the OpenTelemetry project MUST set `telemetry.sdk.name`
to the value `opentelemetry`.

If another SDK, like a fork or a vendor-provided implementation, is used, this SDK MUST set the attribute
`telemetry.sdk.name` to the fully-qualified class or module name of this SDK's main entry point
or another suitable identifier depending on the language.
The identifier `opentelemetry` is reserved and MUST NOT be used in this case.
The identifier SHOULD be stable across different versions of an implementation.

| Attribute  | Description  | Example  | Required? |
|---|---|---|---|
| telemetry.sdk.name | The name of the telemetry SDK as defined above. | `opentelemetry` | No |
| telemetry.sdk.language | The language of the telemetry SDK.<br/> One of the following values MUST be used, if one applies: "cpp", "dotnet", "erlang", "go", "java", "nodejs", "php", "python", "ruby", "webjs" | `java` | No |
| telemetry.sdk.version | The version string of the telemetry SDK as defined in [Version Attributes](#version-attributes). | `semver:1.2.3` | No |
