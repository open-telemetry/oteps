 mplegjkgStandard names for system/runtime metrics

This OTEP proposes a set of standard names and semantic conventions for common system/runtime metrics collected by OpenTelemetry. The metric names and labels proposed here are common across the supported operating systems and runtime environments. This OTEP also includes semantic conventions for metrics specific to a particular OS or runtime.

This OTEP is largely based on the existing implementation in the OpenTelemetry Collector's [Host Metrics Receiver](https://github.com/open-telemetry/opentelemetry-collector/tree/1ad767e62f3dff6f62f32c7360b6fefe0fbf32ff/receiver/hostmetricsreceiver). The proposed names aim to make system/runtime metrics easily discoverable and unambiguous. See [OTEP #108](https://github.com/open-telemetry/oteps/pull/108/files) for additional motivation.

## Explanation
There are some semantic conventions followed in this OTEP in order to keep naming consistent. Not all possible metrics are covered by these conventions, but they provide consistent names for many cases in the proposal:
- **usage** - a measure of the amount an entity is used out of a known total amount should be called `entity.usage`. For example, `system.filesystem.usage` for the amount of disk spaced used. *(I'm open to adjusting this to unit names or something else).*
- **utilization** - a measure of the *percent* usage should be called `entity.utilization`. For example, `system.memory.utilization` for the percentage of memory in use. *(I'm open to a shorter name, but wanted to distinguish between usage as an amount and a percentage).*
- A measure of the amount of an unlimited resource consumed is differentiated from *usage*. This may be time, data, etc.
  * **time** - a measure of the time taken should be called `entity.time`. For example, `system.cpu.time` with varying values of label `state` for idle, user, etc.
  * **io** - a measure of bidirectional data flow should be called `entity.io` and have labels for direction. For example, `system.net.io`.
- Other metrics that do not fit the above descriptions may be named more freely. Units do not *need* to be specified in the names, but can be added if there is ambiguity. For example, `system.swap.page_faults` and `system.net.packets`. 

## Trade-offs and mitigations

When choosing a metric name, there is a trade off between discoverability and ambiguity. For example, a metric called `system.cpu.load_average` is very discoverable, but the meaning of this metric is ambiguous. [Load average](https://en.wikipedia.org/wiki/Load_(computing)) is well defined on UNIX, but is not defined on Windows. Metric names must strike a balance between easy discovery and truthfulness of what the metric represents. 

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


## Open questions

- Should the individual runtimes have their specific naming conventions in the spec?
- Establishing consistent base units for instruments of the same family e.g. seconds vs. milliseconds.
- Should direct measurement be named for the units?

## Internal details
The following standard metric names should be used in libraries instrumenting system/runtime metrics.

### Standard System Metrics - prefix `system.`

#### `system.cpu.`

**Description:** System level processor metrics.

|Name                  |Units  |Instrument       |Label Key |Label Values|Linux/Windows Support|
|----------------------|-------|-----------------|----------|------------|---------------------|
|system.cpu.time       |seconds|SumObserver      |state     |idle        |Both                 |
|                      |       |                 |          |user        |Both                 |
|                      |       |                 |          |system      |Both                 |
|                      |       |                 |          |interrupt   |Both                 |
|                      |       |                 |          |nice        |Linux                |
|                      |       |                 |          |softirq     |Linux                |
|                      |       |                 |          |steal       |Linux                |
|                      |       |                 |          |wait        |Linux                |
|system.cpu.utilization|%      |UpDownSumObserver|(as above)|(as above)  |(as above)           |

#### `system.memory.`

**Description:** System level memory metrics.

|Name                     |Units|Instrument       |Label Key |Label Values       |Linux/Windows Support|
|-------------------------|-----|-----------------|----------|-------------------|---------------------|
|system.memory.usage      |bytes|UpDownSumObserver|state     |used               |Both                 |
|                         |     |                 |          |free               |Both                 |
|                         |     |                 |          |buffered           |Linux                |
|                         |     |                 |          |cached             |Linux                |
|                         |     |                 |          |slab\_reclaimable  |Linux                |
|                         |     |                 |          |slab\_unreclaimable|Linux                |
|system.memory.utilization|%    |UpDownSumObserver|(as above)|(as above)         |(as above)           |

#### `system.swap.`

**Description:** System level swap/paging metrics.

|Name                    |Units|Instrument       |Label Key |Label Values|Linux/Windows Support|
|------------------------|-----|-----------------|----------|------------|---------------------|
|system.swap.usage       |pages|UpDownSumObserver|state     |used        |Both                 |
|                        |     |                 |          |free        |Both                 |
|system.swap.utilization |%    |UpDownSumObserver|(as above)|(as above)  |(as above)           |
|system.swap.page\_faults|1    |SumObserver      |type      |major       |Linux                |
|                        |     |                 |          |minor       |Linux                |
|system.swap.paging\_ops |1    |SumObserver      |type      |major       |Both                 |
|                        |     |                 |          |minor       |Linux                |
|                        |     |                 |direction |in          |Both                 |
|                        |     |                 |          |out         |Both                 |


#### `system.disk.`

**Description:** System level disk performance metrics.

|Name                        |Units|Instrument |Label Key|Label Values|Linux/Windows Support|
|----------------------------|-----|-----------|---------|------------|---------------------|
|system.disk.io<!--notlink-->|bytes|SumObserver|direction|read        |Both                 |
|                            |     |           |         |write       |Both                 |
|system.disk.ops             |1    |SumObserver|direction|read        |Both                 |
|                            |     |           |         |write       |Both                 |
|system.disk.time            |s    |SumObserver|direction|read        |Both                 |
|                            |     |           |         |write       |Both                 |
|system.disk.merged          |1    |SumObserver|direction|read        |Linux                |
|                            |     |           |         |write       |Linux                |

#### `system.filesystem.`

**Description:** System level filesystem metrics. *I think usage/utilization should be consolidated into `system.disk` and the inodes metrics moved to a linux specific namespace.*

|Name                                |Units|Instrument       |Label Key |Label Values|Linux/Windows Support|
|------------------------------------|-----|-----------------|----------|------------|---------------------|
|system.filesystem.usage             |bytes|UpDownSumObserver|state     |used        |Both                 |
|                                    |     |                 |          |free        |Both                 |
|                                    |     |                 |          |reserved    |Linux                |
|system.filesystem.utilization       |%    |UpDownSumObserver|(as above)|(as above)  |(as above)           |
|system.filesystem.inodes.used       |1    |UpDownSumObserver|state     |used        |Linux                |
|                                    |     |                 |          |free        |Linux                |
|system.filesystem.inodes.utilization|%    |UpDownSumObserver|(as above)|(as above)  |(as above)           |

#### `system.net.`

**Description:** System level network metrics.

|Name                       |Units|Instrument       |Label Key|Label Values                                                                                  |Linux/Windows Support|
|---------------------------|-----|-----------------|---------|----------------------------------------------------------------------------------------------|---------------------|
|system.net.dropped\_packets|1    |SumObserver      |direction|transmit                                                                                      |Both                 |
|                           |     |                 |         |receive                                                                                       |Both                 |
|system.net.packets         |1    |SumObserver      |direction|transmit                                                                                      |Both                 |
|                           |     |                 |         |receive                                                                                       |Both                 |
|system.net.errors          |1    |SumObserver      |direction|transmit                                                                                      |Both                 |
|                           |     |                 |         |receive                                                                                       |Both                 |
|system<!--notlink-->.net.io      |bytes|SumObserver      |direction|transmit                                                                                      |Both                 |
|                           |     |                 |         |receive                                                                                       |Both                 |
|system.net.connections     |1    |UpDownSumObserver|protocol |tcp                                                                                           |Both                 |
|                           |     |                 |         |udp                                                                                           |Both                 |
|                           |     |                 |         |[others](https://en.wikipedia.org/wiki/Transport_layer#Protocols)                             |Both                 |
|                           |     |                 |state    |[e.g. for tcp](https://en.wikipedia.org/wiki/Transmission_Control_Protocol#Protocol_operation)|Both                 |

### OS Specific System Metrics - prefix `system.{os}`
Operating system specific metrics

TODO
