# Error Flagging with Status Codes
This proposal adds two status codes explicitly for use as overrides by the end user, and proposes a canonical mapping of semantic conventions to status codes. This clarifies how error reporting should work in OpenTelemetry.
 
## Motivation
Error reporting is a fundamental use case for distributed tracing. While we prefer that error flagging occurs within analysis tools, and not within instrumentation plugins, a number of currently supported analysis tools and protocols rely on the existence of an explicit error flag reported from instrumentation. In OpenTelmetry, the error flag is called "status codes."
 
However, there is confusion over the mapping of semantic conventions to status codes, and concern over the subjective nature of errors. Which network failures count as an error? Are 404s an error? The answer is dependent on the situation.
 
There is one major exception. Both application developers and operators have a deep understanding of what constitutes an error in their system. OpenTelemetry must provide a way for these users to control error flagging, and explicitly indicate both when a span should and should not be counted as an error.
A second exception is supporting analysis tools which require explicit error flagging in the data which they receive. In this case, an operator must be able to apply an error flagging schema at some point during the OTLP data processing pipeline.
 
## Explanation
The following changes add several missing features required for proper error reporting, and are completely backwards compatible with OpenTelemetry today.
 
### Status Codes
The following status codes are added to our current  schema.
 
* `DEFAULT` No status has been set. Any errors must be detected by the analysis tool. (This replaces `OK`.)
* `ERROR` Instrumentation has marked a span as an error. (This replaces `UNKNOWN`.)
* `OK_OVERRIDE` The user has provided an override. The span should NOT be flagged as an error, regardless of other analysis.
* `ERROR_OVERRIDE` The user has provided an override. The span SHOULD be flagged as an error, regardless of other analysis.
 
(Note that our  current status codes include a long list of error types. We may choose to keep  them, change them, or drop them in favor of a single `ERROR` code. How many error types we have is not relevant to this proposal .)
 
`OK_OVERRIDE` and `ERROR_OVERRIDE` are special status codes. These are explicit overrides provided by the end user, and should never be set by shared instrumentation. They should only be set by the application developer (via application code), or by the operator (via the collector).
 
Analysis tools are free to disregard status codes, in favor of their own approach to error analysis. However it is strongly suggested that analysis tools handle `OK_OVERRIDE` and `ERROR_OVERRIDE`, as these are explicitly set by the end user and contain valuable information.
 
 
### Error Mapping Schema
As part of the specification, OpenTelemetry provides a canonical mapping of semantic conventions to status codes. This removes any ambiguity as to what OpenTelemetry ships with out of the box.
 
### Error Processor
The collector will provide a processor and a configuration language to make adjustments to this error mapping schema. This provides the flexibility and customization needed for real world scenarios.
 
### Convenience methods
As a convenience, OpenTelemetry provides helper functions for adding semantic conventions and exceptions to a span. These helper functions will also set the correct status code. This simplifies the life of the instrumentation author, and helps ensure compliance and data quality.
 
 
## Internal details
Except for the renaming of two status codes, this proposal is backwards compatible with existing code, protocols, and the OpenTracing bridge.
 
**OK is renamed to DEFAULT**
 Using the term “OK” as the default implies more meaning than we intend. The span is not necessarily OK - it simply has not triggered our standard error mapping. The default status code should be renamed to `DEFAULT` to avoid confusion.
 
 Note: I intentionally avoided terms like "unset" as it may imply to users that they are required to set the status code, which is not the intention.
 
**UNKNOWN is renamed to ERROR**
 In the new schema, ERROR is the primary status code for reporting errors. Especially in a reduced list of status codes, the term "unknown" is vague and may accidentally imply a meaning to users which we do not intend.
 
 
## BUT ERRORS ARE SUBJECTIVE!! HOW CAN WE KNOW WHAT IS AN ERROR? WHO ARE WE TO DEFINE THIS?
First of all, every tracing system to-date comes with a default set of errors. No system requires that end users start completely from scratch. So... be calm!! Have faith!!
 
While flagging errors can be a subjective decision, it is true that many semantic conventions qualify as an error. By providing a default mapping of semantic conventions to errors, we ensure compatibility with existing analysis tools (e.g. Jaeger), and provide guidance to users and future implementers.
 
Obviously, all systems are different, and users will want to adjust error reporting on a case by case basis. Unwanted errors may be suppressed, and additional errors may be added. The collector will provide a processor and a configuration language to make this a straightforward process. Working from a baseline of standard errors will provide a better experience than having to define a schema from scratch.
 
Note that analysis tools are free to disregard Span Status, and do their own error analysis. For these systems, the only Status codes of import are `OK_OVERRIDE` and `ERROR_OVERRIDE`.
 
If we really hate the current canonical status codes, most may be removed and added back in later. I do suggest we keep the status codes that map to network failures, and I agree that the rest are a bit suspect for our current needs.
 
The minimal number of status codes would be `DEFAULT`, `ERROR`, `OK_OVERRIDE` and `ERROR_OVERRIDE`. `ERROR` will be used to differentiate between standard errors applied by instrumentation and overrides provided by the end user.
 
## Remind me why we need status codes again?
Status codes provide a low overhead mechanism for checking if a span counts against an error budget, without having to scan every attribute and event. This reduces overhead and is a benefit for many systems.
 
Again, the status codes may be customized by the operator during the telemetry pipeline, in order to add and suppress errors.
 
## Open questions
If we add error processing to the collector, it is unclear what the overhead would be.
 
It is also unclear what the cost is for backends to scan for errors on every span, without a hint from instrumentation that an error might be present.
 
## Prior art and alternatives
In OpenTracing, the lack of a Collector and error mapping schema proved to be unwieldy. It placed a burden on instrumentation plugin authors to set the flag correctly, and led to an explosion of non-standardized configuration options in every plugin just to adjust the default error flagging. This in turn placed a configuration burden on application developers.
 
An alternative is the `error.hint` proposal, paired with the removal of status code. This would work, but essentially provides the same mechanism provided in this proposal, only with a large number of breaking changes. It also does not address the need for user overrides.
 
## Future Work
 
The inclusion of status codes and error mappings help the opentelemetry community speak the same language in terms of error reporting. It lifts the burden on future analysis tools, and (when respected) it allows users to employ multiple analysis tools without having to synchronize an important form of configuration across multiple tools.
 
In the future, OpenTelemtry may add a control plane which allows dynamic configuration of the error mapping schema.
