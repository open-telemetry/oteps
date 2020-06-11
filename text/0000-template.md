# A Dynamic Configuration Service for the SDK

This proposal adds a configuration service to dynamically configure metric collection and tracing.

It is related to [this pull request](https://github.com/open-telemetry/opentelemetry-proto/pull/155)

## Motivation

During normal use, a user could use sparse networking, CPU, and memory resources required to send and store telemetry data by specifying sampling of 0.1% of traces, collecting critical metrics such as SLI-related latency distributions and container CPU stats only every five minutes, and not collecting non-critical metrics. Later, while investigating a production issue, the same user could easily increase information available for debugging by reconfiguring some of their processes to sample 2% of traces, collect critical metrics every minute, and begin collecting metrics that were not deemed useful before. Because this change is centralized and does not require redeploying with new configurations, there is lower friction and risk in updating the configurations.

## Explanation

All of this functionality will be optional, so the user does not need to do extra work if they do not need a dynamic config service. If they do, the user, when instrumenting their application, would need to configure the open telemetry agent with the endpoint of their remote configuration service, as well as a Resource and a default config in case we fail to read from the service.

The user must then set up the config service. This can be done throught the collector, which can set up either a stand-alone service, or as something that is an interface to the remote configurations of the user's monitoring and tracing backend, "translating" them to comply with the Open Telemetry configuration protocol

## Internal details

Our protocol will support this call:

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

      // Metrics with names that match a rule in the inclusion_patterns are
      // targeted by this schedule. Metrics that match the exclusion_patterns
      // are not targeted for this schedule, even if they match an inclusion
      // pattern.
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

  // Dynamic configs specific to trace, like sampling rate of a resource.
  message TraceConfig {
    // TODO: unimplemented
  }
  TraceConfig trace_config = 4;

  // Optional. The client is suggested to wait this long (in seconds) before
  // pinging the configuration service again.
  int32 suggested_wait_time_sec = 5;
}
```

On the agent/SDK side, everything will be implemented so that all of this functionality will be optional, and the user is not required to add any new configurations. As stated before, to read from a configuration service, the agents needs to be configured with the service endpoint, a Resource and a default config. The agent will periodically read a config from the service. If it fails to do so, it will just use the default config, or the most recent successfully read config. If it reads a new config, it will apply it.

Export frequency from the SDK depends on Schedules. Each Schedule has inclusion_patterns, exclusion_patterns and a CollectionPeriod. Any metrics that match any of the inclusion_patterns and do not match any of the exclusion_patterns will be exported every CollectionPeriod (ie. every minute).

On the metrics side, the agent implementation will be changed so that it can export metrics that match one certain pattern at a certain interval, while exporting metrics that match another pattern at a different interval. On the tracing side, a sampler will be implemented that can change its sampling rate according to new configs.

The collector will support a new interface for a DynamicConfig service that can be used by an agent, allowing a custom implementation of the configuration service protocol described above, to act as an optional bridge between an agent and an arbitrary configuration service. This interface can be implemented as a shim to support accessing remote configurations from arbitrary backends. The collector is configured to expose an endpoint for requests to the DynamicConfig service, and returns results on that endpoint.

The collector will support both implementing a standalone DynamicConfig service, and combining
DynamicConfig service and Exporter implementations for monitoring and tracing backends that
have integrated remote configurations.

## Trade-offs and mitigations

There are performance concerns, given that the agent is periodically reading from a service and using this config to update itself. Everything will be implemented optionally, so this will have minimal impact on users who opt not the dynamically update their configs.

As mentioned [here](https://github.com/open-telemetry/opentelemetry-proto/pull/155#issuecomment-640582048), the configuration service can be a potential attack vector for an application instrumented with Open Telemetry, depending on what we allow in the protocol. We can highlight in the protocol comments that we should be cautious about what kind of information the agent divulges in its request as well as the sort of behaviour changes that can come about from a config change. 

## Prior art and alternatives

The alternative is to stick with the status quo, where the agent has a [fixed collection period](https://github.com/open-telemetry/opentelemetry-go/blob/34bd99896311a81cf843475779cae2e1c05e6257/sdk/metric/controller/push/push.go#L72-L76) and a fixed sampling rate.

## Open questions

- As mentioned [here](https://github.com/open-telemetry/opentelemetry-proto/pull/155#issuecomment-640582048). what happens if a malicious/accidental config change overwhelms the application/monitoring system? Is it the responsibility of the user to be cautious while making config changes? Should we automatically decrease telemetry exporting if we can detect performance problems?

## Future possibilities

If this OTEP is implemented, there is the option to remotely and dynamically configure other things. As mentioned by commenters in the associated pull request, possibilities include labels and aggregations.
