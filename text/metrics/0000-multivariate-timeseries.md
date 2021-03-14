# Multivariate time-series

Generalization of the metrics data model (currently univariate) to support multivariate time-series.

## Motivation

A Multivariate time series has more than one time-dependent variable. Each variable depends not only on its past values but also has 
some dependency on other variables. A 3 axis accelerometer reporting 3 metrics simultaneously; a meteorological weather station reporting 
temperature, cloud cover, dew point, humidity and wind speed; an http transaction chararterized by many related metrics sharing the same 
labels are all common examples of multivariate time-series. 

With the current version of the OpenTelemetry protocol we force users to transform these multivariate time-series into a collection of 
univariate time-series instead of keeping end-to-end an efficient multivariate representation.

A transport layer should not constrain such a radical change of representation as the implications are so important, to name a few:
* It's straighforward to translate a multivariate time-series into a collection of univariate time-series but the opposite is much more 
complex to do at scale. To rebuild a multivariate from a set of univariate time-series you need to join potentially dozens of time-series 
together. On a distributed backend environment, it's involve a lot of data movement, complex logic to deal with missing data points, some 
form of join computation, ... 
* This transformation enforced by the protocol has also strong implications on the data usage. Each query on a multivariate time-series will
involve a set of joins. The discoverability of data becomes more complex as every univariate time-series are independent. A data catalog
will not magically discover these relationships.
* This transformation involves a lot of redundant informations. All the labels must be redeclared for each univariate time-series. It's
not only more work for the developer but it's also less efficient. Even with a good compression algorithm, a such transformation will be 
much more computationally intensive and slightly less efficient in term of compression ratio.
* This transformation makes simple filtering and aggregation on multivariate time-series much more complex in the OpenTelemetry processor 
layer. This has the potential to transform stateless processors into statefull/complex processors. For example, a user wants to report
http transaction where dns-latency + tcp-con-latency + ssl-handshake-latency + content-transfer-latency + server-processing-latency greater
than 2s. Defining this expression on a multivariate time-series and implementing it with a processor multivariate compatible will be 
straightforward and stateless. With a collection of univariate time-series it's a different story at every level and it's not going in
the direction of a stateless architecture.

By generalizing the existing metric data model we can get rid of all these limitations without adding much complexity to the protocol.
We simplify the implementation of multivariate forecasting and anomaly detection mechanisms. We minimize the risk of seing again a new
telemetry protocol with the only purpose to add a better support for multivariate time-series. Finally, for backends that do not support 
multivariate time-series, a simple transformation to univariate time-series will be simple to implement in their server-side 
OpenTelemetry endpoints.

## Explanation

[TBD]
Explain the proposed change as though it was already implemented and you were explaining it to a user. Depending on which layer the proposal addresses, the "user" may vary, or there may even be multiple.

We encourage you to use examples, diagrams, or whatever else makes the most sense!

Inefficiency, redundant definition of labels
Stateless architecture
Multivariate to univariate is easy but univariate to multivariate is hard
Units
Description

## Internal details

[TBD]
From a technical perspective, how do you propose accomplishing the proposal? In particular, please explain:

* How the change would impact and interact with existing functionality
* Likely error modes (and how to handle them)
* Corner cases (and how to handle them)

While you do not need to prescribe a particular implementation - indeed, OTEPs should be about **behaviour**, not implementation! - it may be useful to provide at least one suggestion as to how the proposal *could* be implemented. This helps reassure reviewers that implementation is at least possible, and often helps them inspire them to think more deeply about trade-offs, alternatives, etc.

## Trade-offs and mitigations

[TBD]
What are some (known!) drawbacks? What are some ways that they might be mitigated?

Note that mitigations do not need to be complete *solutions*, and that they do not need to be accomplished directly through your proposal. A suggested mitigation may even warrant its own OTEP!

## Prior art and alternatives

[TBD]
What are some prior and/or alternative approaches? For instance, is there a corresponding feature in OpenTracing or OpenCensus? What are some ideas that you have rejected?

## Open questions

[TBD]
What are some questions that you know aren't resolved yet by the OTEP? These may be questions that could be answered through further discussion, implementation experiments, or anything else that the future may bring.

## Future possibilities

[TBD]
What are some future changes that this proposal would enable?
