# Introduce Data Classification for Telemetry

Adding optional classification to the attributes and resources to support simplified processing of data.

## Motivation

As the scope of Observability changes to include user monitoring and analytics (Real User Monitoring for now), ensuring that data is handled correctly is becoming more problematic as the problem space changes to include more. The need for pre processing data in order that high cardinality, sensitive user data, and handling sparse data impacts the reliability and usability of the data. Furthermore, some vendors are not able to selectively remove sensitive user data or some agreements with customers mandate that sensitive data not to leave the internal edge. Moreover, attribute values that are known high cardinality (for example a container id within a cloud based environment) that are expensive to store within observability systems, making it easier to omit attributes as part of exporter or collector configuration can greatly reduce cost of the observability suite.; not all attributes and resource data hold the same value or have the same processing requirements.

Updating the SDK to include the convention of data classification would mean that users can extend their current telemetry via an opt in process to enable enhanced controls over the telemetry data without impacting the developer experience. Instrumentation authors are not limited by what attributes they can use when developing their telemetry and not a hindrance for existing standards within Open Telemetry. Organisations can define what resource classifications are allowed to be processed by an observability suite, what is subject to additional processing, and understand the types of data being sent. For example, all high cardinality data can be sent to on prem short term storage and then forwarded to a vendor with high cardinality attributes removed so that it can be viewed over a longer time period to show trends.

Having the ability to set data classification on resources means that:

Known high cardinality attributes can be filtered out to reduce the cost of Observability suites.

Attributes containing User Generated Content, Personal Data or sensitive data can be rerouted into a slower processing queue to redact, anonymise, or drop resource values.

Associate resource telemetry to Service Level Objective / Service Level Indicator (SLO/SLI) so that observability systems can automatically identify data that should be alerted on.  

Ensure that sparse data is not sampled out by mistake

This approach can also be extended by vendors to enable more advanced features in their tool such as high resolution, short retention, data residency, or an endless list of features that can be adopted.

## Explanation

A service that adopts using resource classifications allows the open telemetry exporters to route data, remove problematic attributes, or filter values being sent to downstream systems. Using classifications simplify the existing processes since it does not use a lookup table that needs to be maintained, regular expressions needing to be maintained, required user intervention of what is acceptable or not, or needing to validate the entire resource object thus meaning no performance impact.  

The following are examples of how a service owner or a on call engineer could adopt this change:

1. A cloud native system has signed an agreement with a customer that no personal data or user generated content can be exported to an external vendor and requires the ability to audit the employees that have access to that data.
2. The team responsible for configuring the collector can set rules to drop any resources with a classification of PD, UGC to be sent to their external observability vendor, but send it to an internal elastic search cluster that has RBAC enforced and an audit log that adheres to the customer agreement.
3. Site Reliability Engineers looking to understand the performance bottle neck with a service using a database that contains user data could request an exemption to allow data base queries that contain UGC to be visible temporarily as they investigate further.
4. An external observability vendor is not able to offer protections on the resource data it can accept, however the collector can be set up as a proxy to that vendor and add the required protection that the vendor can not currently offer.

An Instrumentation Author can add additional metadata can be set on the attributes that will be added to the resource object, the example shows how an to configure a middleware in Golang using classifications:

```go
// Middleware implements the semantic convention for http trace spans but
// adds classification to provide certain actions on them latter within
// either the trace exporter or the collector
func Middleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
    ctx, span := otel.GetTraceProvider().Tracer("otel-go-middleware").Start(ctx, privacy.Sanitize(r.URL.Path()),
        trace.WithAttibutes(
          attribute.String("net.peer.ip", r.SourceIP).Classify(attribute.PII, attribute.HighCardinality),
          attribute.String("http.url", r.URL).Classify(attribute.Sensitive)
          attribute.String("http.user_agent", r.UserAgent).Classify(attributes.Sensitive, attributes.HighCardinality)
          attribute.Int("retry.amount", attempts), // Classifications are considered optional by the SDK 
          // ... Other default attributes as defined by the semantic convention
        ), 
      )
    defer span.Stop()
  
    interceptor := otel.NewResponseInteceptor(rw)
    
    defer func() {
      if interceptor.IsError() {
        span.SetAttributes(
          attributes.Error(interceptor.Error()).Classify(attribute.PreserveValue)
        )
      }
    }()
    
    return next.Serve(interceptor, r)
  }))
}
```

Once classification have been added to the attribute keys, a tracing sampler can

```go
func (ts TailBasedSampler) Sample(spans trace.Spans) (export trace.Spans) {
  for _, s := range spans {
    if classification.Match(s.Classification(), classification.Ephemeral) {
      export.Append(s)
      // Always retain spans that contain ephemeral attributes
      continue
    }
    // Perform the existing sampling algorithm of choice
    if !ts.Sample(s) {
      export.Append(s)
    } 
  }
}
```

Since sampler was already filtering the spans, there was no incurred performance impact checking the classification due to the technical implementation of it, meaning that any existing sampler algorithm can be extended using classification matching without issue.

The collector can be extended to support classification based routing:

```yaml
---
# ... omitted for simplicity

processors:
  classifier/remove-sensitive:
    match:
    - pii
    - ugc
    - sensitive
    on_match:
    - drop_resource
  classifier/remove-highcardinality:
    match:
    - high_cardinality
    on_match:
    - drop_field
    
  classifier/tracesampler:
    match:
    - never_sample
    on_match:
    - keep_resource
    
service:
  pipeline:
    metrics:
      receivers:
      - otlp/metrics
      processor:
      - classifier/remove-highcardinality
      - classifier/remove-sensitive
      exporter:
      - external-metric-vendor
    traces:
      receivers:
      - otlp/traces
      processors:
      - classifier/tracesampler
      exporter:
      - external-trace-vendor
```

## Internal details

To add classifications support into the SDK, the OTLP definition would need to extend the definitions used by Attributes to include a new field that will be used to store a bit mask and the the Resource type (datapoint, span, message) would also need to be extended to add a bit mask field.

These are the proposed updates to the protobuf definition:

```proto
// Defined at https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/common/v1/common.proto#L61
message KeyValue {
  string   key   = 1;
  AnyValue value = 2;
+  // classification represents a bitwise flag 
+  // that is used to represent the key value classification type
+  int64    classification  = 3;
}

// Defined at https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/metrics/v1/metrics.proto#L160
message Metric {
// ... omitted 
+  // classification represents a bitwise flag 
+  // that is used to represent metric classification type
+  int64    classification  = 12;
}

// Defined at https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/trace/v1/trace.proto#L84
message Span {
// ... omitted 
+  // classification represents a bitwise flag 
+  // that is used to represent span classification type
+  int64    classification  = 16;
}

// Defined at https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/logs/v1/logs.proto#L113
message LogRecord {
// ... omitted 
+  // classification represents a bitwise flag 
+  // that is used to represent LogRecord classification type
+  int64    classification  = 11;
}
```

Once that has been update in the protobuf definition has been published, a classification list will need to be defined as part of the SDK with space for vendors to allow for their own classification definitions.

A classification MAY contain multiple classification, therefore a classification value MUST not overlap with other classification so that they can be combined together using bitwise inclusive or. The SDK will reserve the 32 Least Significant Bits (LSB) to allow for future additions that are yet to be considered; allow vendors to define values for the 32 Most Significant bits to define as they see fit.

A staring list can look like the following :

| Classification                     | Purpose                                                                                               | BitShift  | Hint Value (base 16)  | bit mask             |
|----------------------------------- |------------------------------------------------------------------------------------------------------ |---------- |---------------------- |--------------------- |
| No Value                           | Is the default value when un set                                                                      | 0 << 0    | 0x0000                | 0000 0000 0000 0000  |
| Ephemeral                          | The attributes that are short lived and have high potential to change over extended periods of time.  | 1 << 0    | 0x0001                | 0000 0000 0000 0001  |
| High Cardinality                   | The value is an unbounded set                                                                         | 1 << 1    | 0x0002                | 0000 0000 0000 0010  |
| Sensitive Value                    | The value MAY contain information that requires sanitisation                                          | 1 << 2    | 0x0004                | 0000 0000 0000 0100  |
| Personal Identifiable Information  | The value DOES contain Personal Identifiable Information                                              | 1 << 3    | 0x0008                | 0000 0000 0000 1000  |
| User Generated Content             | The value DOES contain User Generated Content                                                         | 1 << 4    | 0x0010                | 0000 0000 0001 0000  |
| Service Level Objective            | The value is used to track a service level objective                                                  | 1 << 5    | 0x0020                | 0000 0000 0010 0000  |

In order to support the idea, the required code to be implemented is:

```go
// Combine will take a list of classifications and combine them into 
// one final classification that represents all of them.
// Duplication of hints is a noop since the same field will be set
// repeatedly.
func Combine(classifications ...int32) (classification int32) {
  for _, c := range classifications {
    classification |= c
  }
  return classification
}

// Contains checks if the values are the same or if the hint is contained
// within the value and returns the result
func Contains(value, classification int32) bool {
  // Handles the case where expect = NoValue (0x0000)
  // and peforming a bitwise and would fail the default case
  // without the need for branching
  return value == hint || (value & hint) != 0
}

// Remove takes the value and removes the mask
// all remaining values are left unchanged and returned
func Remove(value, classification int32) int32 {
  return value ^ classification
}
```

The SDK would need to add the following methods to the modified types (mentioned earlier) in order to support classification

```go
func (t *T) AppendAttribute(attr Attribute) {
  t.attributes = append(t.attributes, attr)
  t.Classification = Combine(t.Classification, attr.GetClassification()
}

func (t T) GetClassification() Hint {
  return t.Classification
}

func (t *T) SetClassification(c Classification) {
  t.Classification = c
}
```

Once the SDK has been updated to support the definitions, along with the internal model for the collector, an extension can be used to wrap existing exporters to act on the data:

```go
type ClassifierFilterExtension struct {
  MatchOn Classification
  Action  func(t T) T
}

func (cfe *ClassifierFilterExtension) Match(list []T) (filtered []T) {
  for _, t := range list {
    if Contains(l.GetClassification(), cfe.MathchOn) {
      t = cfe.Action(t)
    }
    filtered.Append(t)
  }
}

// RemoveMathcedAttribute is an example of an Action that 
// can be set when using classification filtering extension
func RemoveMatchedAttributes(c Classification) func(t T) T {
  return func(t T) T { 
      attrs []Attribute
      for _, attr := range t.Attributes() {
        if !Match(attr.GetClassification(), c) {
          attrs.Append(attr)
        }
      }
      t.SetAttributes(attrs)
      t.SetClassification(Remove(t.GetClassification(), c))
      return t  
  }
}
```

## Trade-offs and mitigations

- Going forward with this does imply a lot of ownership on the USER developing the instrumentation code, and a lot of potential cognitive load perceived with it.
  - This only works with USERs implementing their code using OTEL instrumentation.
- Converting from non OTLP formats to OTLP would need depend on processors / converters to support this, however doing so will likely impact performance.
- Managing classifications within the semantic convention has not be explored
- The tolerance of what is considered a given classification could change between organisations

## Prior art and alternatives

There was a conversation that was related here on attributes that should be subject to data regulations Proposal: Add Sensitive Data LabelsOPEN , however, this approach would superseded what is mentioned there and be applied more broadly to the project.

The alternatives that were considered are:

- The Semantic Convention defines classification for known attributes
  - this was no considered heavily since it misses USERS (developers) whom are using the SDK natively and then would need to manage their own Semantic Convention that would also need to be understood by the down stream processors and vendors
    - Does not account for version drift with a breaking change involved
    - Does a processor need to know all versions?
    - Does that mean values can not be changed in the semantic convention?

- Appending prefixes/suffixes to attribute keys
  - String matching is expensive operation and would considerable slow down the current perform wins that the project strives for
  - Can not allow for multiple definitions to be set on one attribute
  - Has the potential to overlap with an exciting keys and require USERs to do a migration of what attributes they are already sending
    - Unneeded work that should be avoided
    - Limits what USERs can define their attributes keys
  - The ownership is then moved to processors and vendors to set the definition
    - Goes against the projects goal of being vendor neutral
    - Each vendor processor could define different values (same as above)

## Open questions

Does the semantic convention define what hints are automatically set for exported attributes?

Can a Service Owner change the definitions of classifications without needed to modify instrumentation code?

How would this work with the semantic convention?

What definitions should be used for the classifications so that the values are set subjectively?

What extra work is missing to make classifications easy to manage as part of the semantic convention?

## Future possibilities

Some ideas of future possibilities have been eluded to throughout the proposal, the potential changes that would be of great impact are:

- Vendor based classification controls
- Vendor neutral data protection mechanisms
- Enforcing Data Regulations on telemetry
- Classification based routing
  - An exporter can only with certain set of classifications

Thinking more broadly, it would mean that Open Telemetry could easily extend to support analytical based telemetry on user interactions while adhering to data regulation or organisational policies.
