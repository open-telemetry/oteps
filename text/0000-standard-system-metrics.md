# Standard names for system/runtime metrics

This OTEP proposes a set of standard names, labels, and semantic conventions for common system/runtime metrics collected by OpenTelemetry. The metric names proposed here are common across the supported operating systems and runtime environments. Also included are general semantic conventions for system/runtime metrics including those not specific to a particular OS or runtime.

This OTEP is largely based on the existing implementation in the OpenTelemetry Collector's [Host Metrics Receiver](https://github.com/open-telemetry/opentelemetry-collector/tree/1ad767e62f3dff6f62f32c7360b6fefe0fbf32ff/receiver/hostmetricsreceiver). The proposed names aim to make system/runtime metrics unambiguous and easily discoverable. See [OTEP #108](https://github.com/open-telemetry/oteps/pull/108/files) for additional motivation.

## Trade-offs and mitigations

When choosing a metric name, there is a trade off between discoverability and ambiguity. For example, a metric called `system.cpu.load_average` is very discoverable, but the meaning of this metric is ambiguous. [Load average](https://en.wikipedia.org/wiki/Load_(computing)) is well defined on UNIX, but is not a standard metric on Windows. While discoverability is important, metric names must be unambiguous. 

## Prior art
There is an existing metric naming proposal [here](https://docs.google.com/spreadsheets/d/1WlStcUe2eQoN1y_UF7TOd6Sw7aV_U0lFcLk5kBNxPsY/edit#gid=0). In addition, there are already a few implementations of system and/or runtime metric collection in OpenTelemetry:

- **Collector**
  * [Host Metrics Receiver](https://github.com/open-telemetry/opentelemetry-collector/tree/1ad767e62f3dff6f62f32c7360b6fefe0fbf32ff/receiver/hostmetricsreceiver) generates metrics about the host system when run as an agent.
  * Currently is the most comprehensive implementation.
  * Collects system metrics for CPU, memory, swap, disks, filesystems, network, and load.
  * Collects process metrics for CPU, memory, and disk usage.
  * Makes good use of labels rather than defining individual metrics.
  * [Overview of collected metrics](https://docs.google.com/spreadsheets/d/11qSmzD9e7PnzaJPYRFdkkKbjTLrAKmvyQpjBjpJsR2s).

- **Go**
  * Go [has instrumentation](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/master/instrumentation/runtime) to collect runtime metrics for GC, heap use, and goroutines. 
  * This package does not export metrics with labels, instead exporting individual metrics.
  * [Overview of collected metrics](https://docs.google.com/spreadsheets/d/1r50cC9ass0A8SZIg2ZpLdvZf6HmQJsUSXFOu-rl4yaY/edit#gid=0).
- **Python**
  * Python [has instrumentation](https://github.com/open-telemetry/opentelemetry-python/tree/master/ext/opentelemetry-ext-system-metrics) to collect some system and runtime metrics.
  * Collects system CPU, memory, and network metrics
  * Collects runtime CPU, memory, and GC metrics.
  * Makes use of labels, similar to the Collector.
  * [Overview of collected metrics](https://docs.google.com/spreadsheets/d/1r50cC9ass0A8SZIg2ZpLdvZf6HmQJsUSXFOu-rl4yaY/edit#gid=0).
- **TODO: Java**
- **TODO: Opencensus**


## Open questions

- Should the individual runtimes have their specific naming conventions in the spec?
- Establishing consistent base units for instruments of the same family e.g. seconds vs. milliseconds. [Prometheus recommends using base units for compatibility](https://prometheus.io/docs/practices/naming/#base-units).

## Semantic Conventions
The following semantic conventions aim to keep naming consistent across different metrics. Not all possible metrics are covered by these conventions, but they provide guidelines for most of the cases in this proposal:
- **usage** - an instrument that measures an amount used out of a known total amount should be called `entity.usage`. For example, `system.filesystem.usage` for the amount of disk spaced used. A measure of the amount of an unlimited resource consumed is differentiated from **usage**. This may be time, data, etc. *(I'm open to adjusting this to unit names or something else).*
- **utilization** - an instrument that measures percent usage should be called `entity.utilization`. For example, `system.memory.utilization` for the percentage of memory in use. *(I'm open to a shorter name, but wanted to distinguish between usage as an amount and a percentage).*
- **time** - an instrument that measures passage of time should be called `entity.time`. For example, `system.cpu.time` with varying values of label `state` for idle, user, etc.
- **io** - an instrument that measures bidirectional data flow should be called `entity.io` and have labels for direction. For example, `system.net.io`.
- Other metrics that do not fit the above descriptions may be named more freely. For example, `system.swap.page_faults` and `system.net.packets`. Units do not need to be specified in the names, but can be added if there is ambiguity.


## Internal details
The following standard metric names should be used in libraries instrumenting system/runtime metrics.

### Standard System Metrics - `system.`
---

#### `system.cpu.`

**Description:** System level processor metrics.
|Name                  |Units  |Instrument       |Label Key|Label Values                       |
|----------------------|-------|-----------------|---------|-----------------------------------|
|system.cpu.time       |seconds|SumObserver      |state    |idle, user, system, interrupt, etc.|
|                      |       |                 |cpu      |1 - #cores                         |
|system.cpu.utilization|%      |UpDownSumObserver|state    |idle, user, system, interrupt, etc.|
|                      |       |                 |cpu      |1 - #cores                         |

#### `system.memory.`

**Description:** System level memory metrics.
|Name                     |Units|Instrument       |Label Key|Label Values            |
|-------------------------|-----|-----------------|---------|------------------------|
|system.memory.usage      |bytes|UpDownSumObserver|state    |used, free, cached, etc.|
|system.memory.utilization|%    |UpDownSumObserver|state    |used, free, cached, etc.|

#### `system.swap.`

**Description:** System level swap/paging metrics.
|Name                    |Units|Instrument       |Label Key|Label Values|
|------------------------|-----|-----------------|---------|------------|
|system.swap.usage       |pages|UpDownSumObserver|state    |used, free  |
|system.swap.utilization |%    |UpDownSumObserver|state    |used, free  |
|system.swap.page\_faults|1    |SumObserver      |type     |major, minor|
|system.swap.paging\_ops |1    |SumObserver      |type     |major, minor|
|                        |     |                 |direction|in, out     |

#### `system.disk.`

**Description:** System level disk performance metrics.
|Name                        |Units|Instrument |Label Key|Label Values|
|----------------------------|-----|-----------|---------|------------|
|system.disk.io<!--notlink-->|bytes|SumObserver|device   |(identifier)|
|                            |     |           |direction|read, write |
|system.disk.ops             |1    |SumObserver|device   |(identifier)|
|                            |     |           |direction|read, write |
|system.disk.time            |s    |SumObserver|device   |(identifier)|
|                            |     |           |direction|read, write |
|system.disk.merged          |1    |SumObserver|device   |(identifier)|
|                            |     |           |direction|read, write |

#### `system.filesystem.`

**Description:** System level filesystem metrics. *I think usage/utilization should be consolidated into `system.disk` and the inodes metrics moved to a linux specific namespace.*
|Name                         |Units|Instrument       |Label Key|Label Values        |
|-----------------------------|-----|-----------------|---------|--------------------|
|system.filesystem.usage      |bytes|UpDownSumObserver|device   |(identifier)        |
|                             |%    |                 |state    |used, free, reserved|
|system.filesystem.utilization|     |UpDownSumObserver|device   |(identifier)        |
|                             |     |                 |state    |used, free, reserved|

#### `system.net.`

**Description:** System level network metrics.
|Name                       |Units|Instrument       |Label Key|Label Values                                                                                  |
|---------------------------|-----|-----------------|---------|----------------------------------------------------------------------------------------------|
|system.net.dropped\_packets|1    |SumObserver      |device   |(identifier)                                                                                  |
|                           |     |                 |direction|transmit, receive                                                                             |
|system.net.packets         |1    |SumObserver      |device   |(identifier)                                                                                  |
|                           |     |                 |direction|transmit, receive                                                                             |
|system.net.errors          |1    |SumObserver      |device   |(identifier)                                                                                  |
|                           |     |                 |direction|transmit, receive                                                                             |
|system<!--notlink-->.net.io|bytes|SumObserver      |device   |(identifier)                                                                                  |
|                           |     |                 |direction|transmit, receive                                                                             |
|system.net.connections     |1    |UpDownSumObserver|device   |(identifier)                                                                                  |
|                           |     |                 |protocol |tcp, udp, [others](https://en.wikipedia.org/wiki/Transport_layer#Protocols)                   |
|                           |     |                 |state    |[e.g. for tcp](https://en.wikipedia.org/wiki/Transmission_Control_Protocol#Protocol_operation)|

#### OS Specific System Metrics - `system.{os}.`
System level metrics specific to a certain operating system should be prefixed with `system.{os}.` and follow the hierarchies listed above for different entities like CPU, memory, and network.

TODO: example

### Standard Runtime Metrics - `runtime.`
---

Runtime environments vary widely in their terminology, implementation, and relative values for a given metric. For example, Go and Python are both garbage collected languages, but comparing heap usage between the two runtimes directly is not meaningful. For this reason, this OTEP does not propose any standard top-level runtime metrics. See [OTEP #109](https://github.com/open-telemetry/oteps/pull/108/files) for additional discussion.

#### Runtime Specific Metrics - `runtime.{environment}.`
Runtime level metrics specific to a certain runtime environment should be prefixed with `runtime.{environment}.` and follow the semantic conventions outlined in [Semantic Conventions](#semantic%20conventions).