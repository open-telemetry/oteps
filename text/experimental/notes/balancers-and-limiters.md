# Two kinds of samplers
Consider two common goals in sampling:
- Avoid over-collection or under-collection of data

- Ensure that overall data collection doesn't exceed acceptable limits

***Coarse-grained adaptive sampling*** is a technique where, given a *heterogeneous* stream of traces, the sampling probabilities of each trace are determined in order to promote the preceding goals. Existing implementations of coarse-grained adaptive sampling pursue these goals simultaneously using single constructs. It is possible, however, to address them independently. Decoupling these concerns may yield a simpler and more flexible conceptual foundation for sampling.

## Balancers
Define a ***balancer*** to be a sampler that does the following: For each input trace,
1. Assign a "frequency" score to the trace.
1. Sample the trace with a probability that's inversely related to the frequency score.

One implementation of (1) is to partition input traces into strata, compute from historical data the relative frequency of each stratum among the input traces, and assign traces frequency scores equal to the relative frequency of the stratum to which they belong. For example, if a trace comes in and belongs to the "`route = /health`" stratum, and that stratum constitutes 10% of recent traces, then any trace belonging to that stratum has frequency score 0.1. This scoring algorithm will result in sampling higher-volume strata, pulling their throughput down to be closer to that of the minimal-volume stratum, reducing the [dynamic range](https://en.wikipedia.org/wiki/Dynamic_range) of strata throughputs.

One implementation of (2) is to perform ***logarithmic balancing.*** Under this scheme, the frequency ratios and throughput ratios of a pair of spans are related in a certain way. By way of example,
- 1:1 frequency => 1:1 throughput
- 10:1 frequency => 2:1 throughput
- 100:1 frequency => 3:1 throughput

Calculation details:
- Traces with minimal frequency are sampled with probability 1. Others are sampled with probability less than 1.
- For any pair of traces a and b with frequencies $f_a \geq f_b$, define $C = 1 + \log_{10}(f_a/f_b)$. Note that $C \geq 1$.
- Pick p-values s.t. for any pair of traces, ratio of trace *expected throughputs* = C. E.g., 10:1 frequency => 2:1 throughput. In this example, reduce the p of the more frequent trace by a factor of $(f_a/f_b)/C$ = 5.
- Could make the base of the logarithm in C's definition a configurable parameter: the ***squeeze factor,*** since it reduces the variance in trace throughputs (equivalently, strata throughputs, if assigning frequency scores as described above). E.g.,
	- base-2: 10:1 frequency => 4.3:1 throughput
	- base-10: 10:1 frequency => 2:1 throughput
	- base-50: 10:1 frequency => 1.6:1 throughput
	- base-∞: 10:1 frequency => 1:1 throughput
		- All traces have throughput equal to that of the minimal-frequency traces

## Limiters
A ***limiter*** is a sampler whose one job is to sample such that output throughputs are at or below some given threshold. For example,
- Per-stratum limiting: Partition input traces into strata, and sample such that each stratum's throughput does not exceed a threshold.
- Global limiting: Sample such that total throughput doesn't exceed a threshold.

Note that in addition to limiting traces per unit time, there are also use cases to support limting spans per unit time, or bytes per unit time. In such cases the limiter implementation should take care not to impart bias by systematically preferring traces comprising fewer spans, or fewer bytes, over "larger" traces.

## In practice
Existing coarse-grained adaptive sampling implementations fuse together balancing and limiting into a single construct. They can, however, be equivalently described in terms of the preceding, decoupled components.
- Jaeger `adaptive`: This attempts to sample all endpoints (pair of service and operation) at a per-endpoint target throughput. This is equivalent to partitioning traces along those two dimensions, running them through a base-∞ balancer, and finally through a per-stratum limiter with threshold equal to `--sampling.target-samples-per-second` many traces per second.
- Honeycomb Refinery: Because Refinery nodes have no shared state, their limiting is not configured in terms of total cluster throughput, nor is it in terms of per-node throughput, but rather sampling probability; in `EMADynamicSampler` samplers the knob is called `GoalSampleRate`. This sampler performs a user-configured partitioning of input traces, scores those traces according to estimated relative frequency of their respective strata, computes a per-node target throughput using per-node strata sizes and `GoalSampleRate`, and then allocates shares of that target throughput to strata in proportion to the base-10 logarithm of each stratum's size. This is equivalent to running all traces through a base-10 logarithmic balancer, followed by a global limiter whose threshold is dynamically adjusting so that some desired percentage of the input traces are included in the sample.
- AWS X-Ray: Not quite "coarse-grained" adaptive sampling, since its configuration requires individual target throughputs, and its rule semantics map each trace to exactly one target throughput.