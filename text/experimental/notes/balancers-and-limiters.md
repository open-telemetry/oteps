# Two kinds of samplers
Consider two common goals in sampling:
- Avoid over-collection or under-collection of data

- Ensure that overall data collection doesn't exceed acceptable limits

***Coarse-grained adaptive sampling*** is a technique where, given a *heterogeneous* stream of traces, sampling probabilities are chosen on a trace-by-trace basis in order to promote the preceding goals. Existing implementations of coarse-grained adaptive sampling pursue these goals simultaneously using single constructs. It is possible, however, to address them independently. Decoupling these concerns may yield a simpler and more flexible conceptual foundation for sampling.

## Balancers
Define a ***balancer*** to be a sampler that does the following: For each input trace,
1. Assign a "frequency" score to the trace.
1. Sample the trace with a probability that's inversely related to the frequency score.

One implementation of (1) is to partition input traces into strata, compute from historical data the relative frequency of each stratum among the input traces, and assign traces frequency scores equal to the relative frequency of the stratum to which they belong. For example, if a trace comes in and belongs to the "`route = /health`" stratum, and that stratum constitutes 10% of recent traces, then any trace belonging to that stratum has frequency score 0.1. This scoring algorithm will result in sampling higher-volume strata, pulling their throughput down to be closer to that of the minimal-volume stratum, reducing the [dynamic range](https://en.wikipedia.org/wiki/Dynamic_range) of strata throughputs.

One implementation of (2) is to perform ***logarithmic balancing.*** Under this scheme, the frequency ratios and throughput ratios of a pair of spans are related in a certain way. By way of example,

- 1:1 frequency => 1:1 throughput
- 10:1 frequency => 2:1 throughput
- 100:1 frequency => 3:1 throughput
- $10^{n-1}$:1 frequency => $n$:1 throughput

Calculation details:
- Traces with minimal frequency $f_0$ are sampled with probability 1. Others are sampled with probability less than 1.
- For any other trace with frequency $f$, define $C = 1 + \log_{10}(f/f_0)$. Note that $C \geq 1$.
- Pick that trace's sampling probability such that the ratio of *expected throughputs* of traces with frequency $f$ and traces with frequency $f_0$ equals C. For example, 10:1 frequency => 2:1 throughput. In this example, reduce the sampling probability of the more frequent trace by a factor of $(f/f_0)/C$ = 5.

If there are use cases where "10:1 frequency => 2:1 throughput" is not quite right—say, more than 2x as much data should be collected for the higher-frequency traces—then we could define a parameter: the ***doubling point* $D \geq 2$, such that:

- $D$ is the base of the logarithm in $C$'s definition.
- Note that when $D = \infty$, $C = 1$, meaning all classes of traces are collected with the same throughput. This produces a sample with maximum diversity, which can be good or bad depending on one's goals.
- Its meaning is described by the statements:
  - Given traces with frequencies $f_0$ and $f = D \times f_0$, the higher-frequency trace is $D/2$ times less likely to be included in the sample than the minimal-frequency trace.
  - (If doing stratum-based scoring as described above) Given a stratum B with $D$ times as much volume as the minimal-volume stratum A, twice as much "B" data will be collected as "A" data.

Another implementation of (2) could be ***slowed exponential balancing:***

- 1:1 frequency => 1:1 throughput
- 10:1 frequency => 2:1 throughput
- 100:1 frequency => 4:1 throughput
- $10^n$:1 frequency => $2^n$:1 throughput

This is simpler to describe: rather than pegging all traces to $f_0$, there is a simple relation between *any pair* of traces: 10x-ing frequency leads to doubling throughput.

## Limiters
A ***limiter*** is a sampler whose one job is to sample such that output throughputs are at or below some given threshold. For example,
- Per-stratum limiting: Partition input traces into strata, and sample such that each stratum's throughput does not exceed a threshold.
- Global limiting: Sample such that total throughput doesn't exceed a threshold.

Note that in addition to limiting traces per unit time, there are also use cases to support limting spans per unit time, or bytes per unit time. In such cases the limiter implementation should take care not to impart bias by systematically preferring traces comprising fewer spans, or fewer bytes, over "larger" traces.

## In practice
Existing coarse-grained adaptive sampling implementations fuse together balancing and limiting into a single construct. They can, however, be equivalently described in terms of the preceding, decoupled components.
- Jaeger `adaptive`: This attempts to sample all endpoints (pair of service and operation) at a per-endpoint target throughput. This is equivalent to partitioning traces along those two dimensions, running them through a $D = \infty$ logarithmic balancer, and finally through a per-stratum limiter with threshold equal to `--sampling.target-samples-per-second` many traces per second.
- Honeycomb Refinery: Because Refinery nodes have no shared state, their limiting is not configured in terms of total cluster throughput, nor is it in terms of per-node throughput, but rather sampling probability; in `EMADynamicSampler` samplers the knob is called `GoalSampleRate`. This sampler performs a user-configured partitioning of input traces, scores those traces according to estimated relative frequency of their respective strata, computes a per-node target throughput using per-node strata sizes and `GoalSampleRate`, and then allocates shares of that target throughput to strata in proportion to the base-10 logarithm of each stratum's size. This is equivalent to running all traces through a $D = 10$ logarithmic balancer, followed by a global limiter whose threshold is dynamically adjusting so that some desired percentage of the input traces are included in the sample.
- AWS X-Ray: Not quite "coarse-grained" adaptive sampling, since its configuration requires individual target throughputs, and its rule semantics map each trace to exactly one target throughput.