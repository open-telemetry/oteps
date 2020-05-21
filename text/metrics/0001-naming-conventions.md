# Metric naming conventions

## Purpose

Metric names and labels are the primary read-interface to metric data. The names and taxonomy need to be understandable and discoverable during routine exploration -- and this becomes critical during incidents.

## Guidelines

Namespace similar metrics together. Define top-level namespaces for common metric categories: for OS, like CPU and network; for app runtimes, like the JVM. This aids in discovery and adhoc comparison

Provide consistent names for common label. Provides discoverability and disambiguation similar to metric names.

Avoid semantic ambiguity. Use namespaced metric names in cases where similar metrics have significantly different implementations across the holistics of metrics. For example, every garbage collected runtime has a slightly different strategies and measures. Using common metric names for GC, not namespaces by the runtime, could create dissimilar comparisons and confusion for end users. Measures of operating system memory are similar.

## Conventions

All metrics have a limited set of common labels:
* `system.host`

### Operation System

#### CPU

`system.cpu` is the core metric name.

Standard usage labels include, non-exhaustively,
* `cpu.idle`
* `cpu.user`
* `cpu.sys`
* `cpu.real`
* `cpu.iowait`
* `cpu.nice`
* `cpu.interrupt`
* `cpu.softirq`
* `cpu.steal`

A user can derive total CPU capacity by summing `system.cpu` across all labels.

A user can derive CPU utilization by summing all values for the `cpu.idle` label and comparing that with all `system.cpu` values across all labels.

`system.cpu` may include labels for per-core measures:
* `cpu.core.[0-n]`, eg, `cpu.core.3`

Cores should be reported ordinally as ordered by the operating system. It is recommended that values begin at 0.

It is recommended that per-core labels should not be reported by default to reduce cardinality; a user should opt-in with via configuration.

#### Network

`system.network` is the core metric name.

Standard labels include, non-exhaustively,
* `network.sent`
* `network.received`

`system.network` may include labels for per-NIC measures:
* `network.nic.[0-n]`, eg, `network.nic.3`

NICs should be reported ordinally as ordered by the operating system. It is recommended that values begin at 0.

Interfaces may also be reported by OS name, eg, `en3` in the label `network.nic.en3`.

It is recommended that per-NIC labels should not be reported by default to reduce cardinality; a user should opt-in with via configuration.

TODO: what other network labels? dropped? how about low-level things like window size?

#### Memory

`system.memory` is the core metric name.

Standard labels include, non-exhaustively,
* `memory.free`
* `memory.resident`
* `memory.shared`
* `memory.private`

TODO: how can a user derive total memory? how about memory utilization? memory allocations that are reported in more than one label may make this difficult.

#### More system metrics

Possibilities:
* Disk
* Filesystems
* Load
* Per-process
* Processes and threads?

Note for discussion: see this [excellent reference guide](https://docs.google.com/spreadsheets/d/11qSmzD9e7PnzaJPYRFdkkKbjTLrAKmvyQpjBjpJsR2s/edit#gid=0)

### Caution

Operating systems will report different labels for common metrics based on their architecture. Queries should be scoped to a host or cluster running the same operating system to avoid aggregating dissimilar measures.


#### TODO lower-level system metrics to consider
* CPU interrupts
* System calls
* Swap and paging

### Application Runtime

All runtime metric names should be reported with a namespace that includes the name of the runtime, eg, `runtime.go.*`.

#### Go

All Go runtime metrics should be within the `runtime.go.*` namespace.

Common metrics include:
* `runtime.go.goroutines`
* `runtime.go.heap_alloc`
* `runtime.go.gc` with labels `gc.count`

#### Java

All Java runtime metrics should be within the `runtime.java.*` namespace.

* `runtime.java.threads` with optional labels for thread pools, eg, `runtime.java.thread_pool.[name]`
* `runtime.java.heap_alloc`
* `runtime.java.gc` with labels `gc.count` and `gc.time`

#### Node.js

All Node runtime metrics should be within the `runtime.nodejs.*` namespace.

* `runtime.nodejs.gc` with labels `pause.time` and `pause.count`
* `runtime.nodejs.heap_alloc` with labels `heap.total_size`, `heap.available_size`, `heap.used_heap_size` **TODO** confirm
* `runtime.nodejs.event_loop` with TBD labels **TODO**

Note: We use `nodejs` here to disambiguate from the `node` term in Kubernetes and elsewhere.

#### More runtimes

**TODO** ...et al... please contribute more runtimes as comments

## User-defined metrics

All user-defined metric conventions below are **recommendations**.

We recommend that users namespace their metrics into logical groups, eg, `shopping_cart.add_item`, `shopping_cart.remove_item`, `shopping_cart.increase_quantity`, and so forth.

We recommend that users consider common labels for their organization. For example, an organization may wish to track the performance of their systems for a specific customer organization; in this case, a common `customer.organization` label could be applied generically.

## Questions for PR review

* Separators
  * namespace separators, eg, `runtime.go`
  * word-token separtors inside a metric name, eg, `heap_alloc`

* What about things that overlap with tracing span data like upstream/downstream callers or originating systems?
