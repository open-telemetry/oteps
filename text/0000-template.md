# A Dynamic Configuration Service for the SDK

This proposal is for an experiment to add a configuration service to dynamically configure metric collection. Tracing is also intended to be added, with details left for a later iteration.

It is related to [this pull request](https://github.com/open-telemetry/opentelemetry-proto/pull/155)

## Motivation

During normal use, a user could use sparse networking, CPU, and memory resources required to send and store telemetry data by specifying sampling of 0.1% of traces, collecting critical metrics such as SLI-related latency distributions and container CPU stats only every five minutes, and not collecting non-critical metrics. Later, while investigating a production issue, the same user could easily increase information available for debugging by reconfiguring some of their processes to sample 2% of traces, collect critical metrics every minute, and begin collecting metrics that were not deemed useful before. Because this change is centralized and does not require redeploying with new configurations, there is lower friction and risk in updating the configurations.

## Explanation

This OTEP is a request for an experimental feature [open-telemetry/opentelemetry-specification#62](https://github.com/open-telemetry/opentelemetry-specification/pull/632). It is intended to seek approval to develop a proof of concept. This means no development will be done inside either the Open Telemetry SDK or the collector.

Development is intended to be split into two phases:
1. Add remote configuration for metric collection periods. In this phase, a configuration service will be implemented as an extension to the collector. The user should be able to, for example, dynamically change the metric collection period from 1 minute to 5 minutes.
2. Add per-metric configuration. In the status quo, all metrics are exported at the same time. This phase will allow the user to, for example, configure the SDK so that it exports certain metrics every 30 seconds, and others every one minute.

Since this will be implemented mostly in the contrib repos, all of this functionality will be optional.

The user, when instrumenting their application, can configure the SDK with the endpoint of their remote configuration service, the associated Resource and a default config we revert to if we fail to read from the configuration service.

The user must then set up the config service. This MUST be done through the collector, which can be set up to expose an arbitrary configuration service implementation. Depending on implementation, this allows the collector to either act as a stand-alone configuration service, or as a bridge to remote configurations of the user's monitoring and tracing backend by 'translating' the monitoring backend's protocol to comply with the Open Telemetry configuration protocol.

## Internal details

Our remote configuration protocol will support this call:

```
service DynamicConfig {
  rpc GetConfig (ConfigRequest) returns (ConfigResponse);
}
```

A request to the config service will look like this:

```
message ConfigRequest{

  // Required. The resource for which configuration should be returned.
  opentelemetry.proto.resource.v1.Resource resource = 2;

  // Optional. The value of ConfigResponse.fingerprint for the last configuration
  // a resource received that was successfully applied.
  bytes last_known_fingerprint = 3;
}
```

While the response will look like this:

```
message ConfigResponse {

  // Optional. The fingerprint associated with this ConfigResponse. Each change
  // in configs yields a different fingerprint.
  bytes fingerprint = 2;

  // Dynamic configs specific to metrics
  message MetricConfig {


    // A Schedule is used to apply a particular scheduling configuration to
    // a metric. If a metric name matches a schedule's patterns, then the metric
    // adopts the configuration specified by the schedule.

    message Schedule {

      // A light-weight pattern that can match 1 or more
      // metrics, for which this schedule will apply. The string is used to
      // match against metric names. It should not exceed 100k characters.
      message Pattern {
        oneof match {
          string equals = 1;       // matches the metric name exactly
          string starts_with = 2;  // prefix-matches the metric name
        }
      }

      // Metrics with names that match at least one rule in the inclusion_patterns are
      // targeted by this schedule. Metrics that match at least one rule from the
      // exclusion_patterns are not targeted for this schedule, even if they match an
      // inclusion pattern.
      repeated Pattern inclusion_patterns = 1;
      repeated Pattern exclusion_patterns = 2;

      // CollectionPeriod describes the sampling period for each metric. All
      // larger units are divisible by all smaller ones.
      enum CollectionPeriod {
        NONE = 0;  // For non-periodic data (client sends points whenever)
        SEC_1 = 1;
        SEC_5 = 5;
        SEC_10 = 10;
        SEC_30 = 30;
        MIN_1 = 60;
        MIN_5 = 300;
        MIN_10 = 600;
        MIN_30 = 1800;
        HR_1 = 3600;
        HR_2 = 7200;
        HR_4 = 14400;
        HR_12 = 43200;
        DAY_1 = 86400;
        DAY_7 = 604800;
      }
      CollectionPeriod period = 3;

      // Optional. Additional opaque metadata associated with the schedule.
      // Interpreting metadata is implementation specific.
      bytes metadata = 4;
    }

    repeated Schedule schedules = 1;

  }
  MetricConfig metric_config = 3;

  // Dynamic configs specific to trace, like the sampling rate of a resource.
  message TraceConfig {
    // TODO: unimplemented
  }
  TraceConfig trace_config = 4;

  // Optional. The client is suggested to wait this long (in seconds) before
  // pinging the configuration service again.
  int32 suggested_wait_time_sec = 5;
}
```

The SDK will periodically read a config from the service using GetConfig. If it fails to do so, it will just use either the default config or the most recent successfully read config. If it reads a new config, it will apply it.

The export frequency of a metric depends on which Schedules apply to it. If the metric matches any of a Schedule's inclusion_pattern and does not match any of it's exclusion_patterns, it will be exported at that Schedule's CollectionPeriod. Metrics can match multiple Schedules.

The collector will support a new interface for a DynamicConfig service that can be used by an SDK, allowing a custom implementation of the configuration service protocol described above, to act as an optional bridge between an SDK and an arbitrary configuration service. This interface can be implemented as a shim to support accessing remote configurations from arbitrary backends. The collector is configured to expose an endpoint for requests to the DynamicConfig service, and returns results on that endpoint.

## Trade-offs and mitigations

This feature will be implemented purely as an experiment, to demonstrate its viability and usefulness. More investigation can be done after a rough prototype is demonstrated.

There are performance concerns, given that the SDK is periodically reading from a service and using this config to update itself. Everything will be implemented optionally, so this will have minimal impact on users who opt not the dynamically update their configs.

As mentioned [here](https://github.com/open-telemetry/opentelemetry-proto/pull/155#issuecomment-640582048), the configuration service can be a potential attack vector for an application instrumented with Open Telemetry, depending on what we allow in the protocol. We can highlight in the protocol comments that we should be cautious about what kind of information the SDK divulges in its request as well as the sort of behaviour changes that can come about from a config change. 

## Prior art and alternatives

One way to configure metric export schedules is to, instead of pushing the metrics, use a pull mechanism to have metrics pulled whenever wanted. This has been implemented for [Prometheus](https://github.com/open-telemetry/opentelemetry-go/pull/751).

The alternative is to stick with the status quo, where the SDK has a [fixed collection period](https://github.com/open-telemetry/opentelemetry-go/blob/34bd99896311a81cf843475779cae2e1c05e6257/sdk/metric/controller/push/push.go#L72-L76) and a fixed sampling rate.

## Open questions

- As mentioned [here](https://github.com/open-telemetry/opentelemetry-proto/pull/155#issuecomment-640582048). what happens if a malicious/accidental config change overwhelms the application/monitoring system? Is it the responsibility of the user to be cautious while making config changes? Should we automatically decrease telemetry exporting if we can detect performance problems?

## Future possibilities

If this OTEP is implemented, there is the option to remotely and dynamically configure other things. As mentioned by commenters in the associated pull request, possibilities include labels and aggregations.

In the future, we can implement per-metric configuration using our specification. It is also possible to make it so to remotely configure not from the collector, but from a different configuration service.
