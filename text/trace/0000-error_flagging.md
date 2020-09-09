# Error Flagging with Status Codes
This proposal adds two status codes explicitly for use as overrides by the end user, and proposes a canonical mapping of semantic conventions to status codes. This clarifies how error reporting should work in OpenTelemetry.
 
## Motivation
Error reporting is a fundamental use case for distributed tracing. While we prefer that error flagging occurs within analysis tools, and not within instrumentation plugins, a number of currently supported analysis tools and protocols rely on the existence of an explicit error flag reported from instrumentation. In OpenTelemetry, the error flag is called "status codes."
 
However, there is confusion over the mapping of semantic conventions to status codes, and concern over the subjective nature of errors. Which network failures count as an error? Are 404s an error? The answer is often dependent on the situation, but without even a baseline of suggested status codes for each convention, the instrumentation author is placed under the heavy burden of making the decision. Worse, the decisions will not be in sync across different instrumentation packages.
 
There is one other missing piece, required for proper error flagging. Both application developers and operators have a deep understanding of what constitutes an error in their system. OpenTelemetry must provide a way for these users to control error flagging, and explicitly indicate that it is the end user setting the status code, and not instrumentation plugins. In these specific cases, the error flagging is known to be correct: the end user has decided the status of the span, and they do not want another interpretation. 

While generic instrumentation can only provide a generic schema, end users are capable of making subjective decisions about their systems. And, as the end user, they should get to have the final call in what consitutes an error. In order to accomplish this, there must be a way to differntiate between errors flagged by instrumentation, and errors flagged by the end user.
 
## Explanation
The following changes add several missing features required for proper error reporting, and are completely backwards compatible with OpenTelemetry today.
 
### `span.user_override(boolean)`
The `user_override` method indicates that the end user has confirmed that the status code is correct. When using OTLP, this will set the `user_override` field. When setting status codes via the collector or application code, `user_override` can be set to ensure that the span status is not re-interpreted by further analysis.

Analysis tools MAY disregard status codes, in favor of their own approach to error analysis. However, it is strongly suggested that analysis tools SHOULD pay attention to the status code when `user_override` is set, as it is a communication from the end-user and contains valuable information.

### Status Codes
Note that our current status codes include a long list of error types. We may choose to keep them, change them, or drop them in favor of a single `ERROR` code. How many error types we have is not relevant to this proposal.
 
### Status Mapping Schema
As part of the specification, OpenTelemetry provides a canonical mapping of semantic conventions to status codes. This removes any ambiguity as to what OpenTelemetry ships with out of the box.
 
### Status Processor
The collector will provide a processor and a configuration language to make adjustments to this status mapping schema. This provides the flexibility and customization needed for real world scenarios.
 
### Convenience methods
As a convenience, OpenTelemetry provides helper functions for adding semantic conventions and exceptions to a span. These helper functions will also set the correct status code. This simplifies the life of the instrumentation author, and helps ensure compliance and data quality.

Note that these convenience methods simply wire together multiple API calls. They should live in a helper package, and should not be directly added to existing API interfaces. Given how many semantic conventions we have, there will be a pile of them.
 
 
## Internal details
This proposal is backwards compatible with existing code, protocols, and the OpenTracing bridge.
 
 
## BUT ERRORS ARE SUBJECTIVE!! HOW CAN WE KNOW WHAT IS AN ERROR? WHO ARE WE TO DEFINE THIS?
First of all, every tracing system to-date comes with a default set of errors. No system requires that end users start completely from scratch. So... be calm!! Have faith!!
 
While flagging errors can be a subjective decision, it is true that many semantic conventions qualify as an error. By providing a default mapping of semantic conventions to errors, we ensure compatibility with existing analysis tools (e.g. Jaeger), and provide guidance to users and future implementers.
 
Obviously, all systems are different, and users will want to adjust error reporting on a case by case basis. Unwanted errors may be suppressed, and additional errors may be added. The collector will provide a processor and a configuration language to make this a straightforward process. Working from a baseline of standard errors will provide a better experience than having to define a schema from scratch.
 
Note that analysis tools MAY disregard Span Status, and do their own error analysis. There is no requirement that the status code is respected, even when `user_override` is set. However, it is strongly suggested that analysis tools SHOULD pay attention to the status code when `user_override` is set, as it represents a subjective decision made by either the operator or application developer.
 
If we really hate the current canonical status codes, most may be removed and added back in later. I do suggest we keep the status codes that map to network failures, and I agree that the rest are a bit suspect for our current needs. The minimal number of status codes would be `OK` and `ERROR`.
 
## Remind me why we need status codes again?
Status codes provide a low overhead mechanism for checking if a span counts against an error budget, without having to scan every attribute and event. This reduces overhead and is a benefit for many systems.
 
Again, the status codes may be customized by the operator during the telemetry pipeline, in order to add and suppress errors.
 
## Open questions
If we add error processing to the Collector, it is unclear what the overhead would be.
 
It is also unclear what the cost is for backends to scan for errors on every span, without a hint from instrumentation that an error might be present.
 
## Prior art and alternatives
In OpenTracing, the lack of a Collector and status mapping schema proved to be unwieldy. It placed a burden on instrumentation plugin authors to set the error flag correctly, and led to an explosion of non-standardized configuration options in every plugin just to adjust the default error flagging. This in turn placed a configuration burden on application developers.
 
An alternative is the `error.hint` proposal, paired with the removal of status code. This would work, but essentially provides the same mechanism provided in this proposal, only with a large number of breaking changes. It also does not address the need for user overrides.
 
## Future Work
 
The inclusion of status codes and status mappings help the OpenTelemetry community speak the same language in terms of error reporting. It lifts the burden on future analysis tools, and (when respected) it allows users to employ multiple analysis tools without having to synchronize an important form of configuration across multiple tools.
 
In the future, OpenTelemetry may add a control plane which allows dynamic configuration of the status mapping schema.
