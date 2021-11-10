# Guide to Uptime Monitoring for Metric Semantic convetions

A guide to best practices and use cases around uptime monitoring using metrics.

Original authors:  Keith Jordy (@kjordy) and Quentin Smith (@quentinmit), adapted for OpenTelemetry by @jsuereth.

## Motivation

Why should we make this change? What new value would it bring? What use cases does it enable?

## Explanation

Users often want to monitor the health of their long-lived tasks. However, what they mean when they use the word "up" is overloaded. 

### Use Cases

We'd like to drive a set of metric conventions around uptime metrics for processes (and possibly other process-like systems) for OpenTelemetry, based around these common use cases:

- Graphing fleet-wide process age
- Alerting on restarts using uptime
- Alerting on current process health
- Alerting on crashing processes

#### Graphing fleet-wide process age

An individual process's uptime produces an "up and to the right" graph when its absolute value is plotted. Graphing many processes' uptimes, either as lines or a heatmap, can identify when there are fleet-wide events that cause processes to restart simultaneously.

This use case can also be met with a restart count metric, but "up and to the right" uptime graphs are more comfortable for users.

#### Alerting on restarts using uptime

Some customers want to configure an alert whenever their process (re)starts. Typically, they do this by configuring an alert that fires when the absolute value of the uptime (time since process start) is below a certain threshold.

This type of alert is considered bad practice because it is, by definition, non-actionable. By the time the uptime is reported as 0, the process has already started and is now running. The alert will self-close as soon as the process has been running for a few minutes.

We recommend in general not configuring alerts on this type of metric, but it is a common practice.

#### Alerting on current process health

Some customers want to configure an alert when a process is unhealthy, or the number of healthy processes falls below a certain level. They can do this using a boolean metric exported from each task that reports if it is currently healthy.

In general, we recommend that services prefer exposing more detailed metrics instead of boolean health. For example, a user should prefer an alert defined as "successful qps > 0.1" instead of an alert defined as "healthy = TRUE". The reason for this is that boolean health metrics essentially move the alert condition definition inside the process, removing the user's ability to control those conditions.

#### Alerting on crashing processes

Users want to know if their processes are restarting frequently (for example, as part of a crash loop). This is the hardest type of metric to report, because restarting processes may not be healthy enough to report that they are crashing.

When there is some kind of supervisor process that is separate from the process being monitored, this can be tracked by having the supervisor process report the number of times a process has restarted. Then an alert can be configured when the rate of that count exceeds a threshold (e.g. "process has restarted ≥ 5 times in one hour").

Since supervisor processes like this do not often exist, other ways to partially achieve this are to configure an alert on the average uptime over a long time period (e.g. "average uptime over 60m window < 15m"), and/or alert when a metric is absent.

## Internal details

We propose the following metrics be used to track uptime within OpenTelemetry:

| Name                   | Description                  | Units | Instrument Type              |
| ---------------------- | ---------------------------- | ----- | -----------------------------|
| *.uptime               | Seconds since last restart   | s     | Asynchronous UpDownCounter   |
| *.health               | Availability flag.           | 1     | Asynchronous Gauge           |
| *.restart_count        | Number of restarts.          | 1     | Asynchronous Counter         |


### Uptime
uptime is reported as a non-monotonic sum with the value of the number of seconds that the process has been up. This is written as a non-monotonic sum because users want to know the actual value of the number of seconds since the last restart to satisfy the use cases above. Montonic sums are not a good fit for these use cases because most metric backends tend to default cumulative monotonic sums to rate-calculations, and have overflow handling that is undesired for this use case.

Cumulative monotonic sums report a total value that has accumulated over a time window; it is valid, for instance, to subtract the current value of a cumulative sum and restart the start timestamp to now. (OpenTelemetry's Prometheus receiver does this, for instance.)
An intended use case of a cumulative sum is to produce a meaningful value when aggregating away labels using sum (this operation requires a delta alignment first). Such aggregations are not meaningful in the above use cases.

### Health
health is a GAUGE which a boolean value (or 0|1) which indicates if the process is available. This satisfies the “Alerting on current process health” use case above. Health often reflects more than just whether the process is alive; e.g. a process that is in the middle of (re)loading data might affirmatively report FALSE during that time. Because metric are sampled periodically, this metric isn’t well suited for use cases of rapidly changing value (i.e. it is likely to miss a restart).

### Restart Count
restart_count is a monotonic sum of the number of times that a process has restarted. This metric should be generated from an external observer of the system.  The start timestamp of this metric is the start time of whatever process is observing restarts.
A process *may* report its own restarts, but likely this would need to be done via a DELTA sum which is aggregated by some external observer.


## Trade-offs and mitigations

The biggest tradeoff here is defining `uptime` metrics as non-montonic sums vs. either pure gauge or non-montonic sums. The fundmental question here is whether default sum-based aggregation is meaningful for this metric, in addition to the default-query-capabilities of common backends for cumulative sums. The proposal trades-off allowing an external observer to monitor uptime (with resets) in addition to common assumptions on querying rates for cumulative sums.

## Prior art and alternatives

The biggest prior art in this space is Prometheus, which has some built-in uptime-like features:


Prometheus defines some conventions around uptime tracking across its ecosystem:

- `up` is a gauge with a `{0, 1}` value indicating whether Prometheus successfully scraped metrics from the target. If a task is known but can't be scraped, up is reported as `0`. If a task is unknown (e.g. the container is not scheduled), up is not reported.
- `process_uptime` is a counter reporting the number of seconds since the process has started. The Prometheus server does not actually store or use the type of a metric, so `process_uptime` is graphed as an absolute value despite being reported as a counter.
- Restart count is not a "global" or "built-in" feature. However, in Kubernetes deployments with kube-state-metrics, `kube_pod_container_status_restarts_total` reports the number of times a container has restarted.

There are obvious differences between the proposed `health` metric and `up` metric in prometheus.  We believe these serve different use cases, but can be complementary.  Generally `up` in prometheus is done by an external observer, while `health` can be reported by the process itself.

`uptime` lines up with prometheus, but we allow non-monotonic sums so external observers can report uptime on behalf of a process.

`restart_count` lines up with what is done in kubestat metrics in the prometheus cosystem.


## Open questions

Should OpenTelemetry specify `up` metrics as "exactly what prometheus does"?

## Future possibilities

What are some future changes that this proposal would enable?
