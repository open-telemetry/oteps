# Mandatory unique identifier for telemetry sources

Provide an explicit mandatory unique identifier for telemetry sources.

## Motivation

Having a way to uniquely identify a telemetry source is helpful in many ways, like in processing and storing data from that source, visualizing them in a backend UI or debugging issues with that source and it's data.

As of now `service.name` (and related attributes `service.namespace` and `service.instance_id`) are the implicit standard for that due to `service.name` being enforced as mandatory by the [Resource SDK specification](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/sdk.md#sdk-provided-resource-attributes) and [Resource Semantic Conventions](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/semantic_conventions/README.md#semantic-attributes-with-sdk-provided-default-value).

Due to the fact that those attributes are not **explicitly** available to uniquely identify a telemetry source, multiple approaches have been suggested:

1. [opentelemetry-specification/issues#1034](
https://github.com/open-telemetry/opentelemetry-specification/issues/1034) is suggesting that `service.instance.id`is poorly defined and should be removed and be replaced by something different like an `telemetry.sdk.instance_id`. An attribute like `telemetry.sdk.instance_id` could serve as the sole unique identifier.

2. [open-telemetry/opentelemetry-specification#2111](https://github.com/open-telemetry/opentelemetry-specification/pull/2111) is proposing to provide a broad definition for the term _Service_, which would mean that (almost) every telemetry source is a service and `service.name` (and `namespace` and `instance_id`) could be used as unique identifier.

3. [open-telemetry/opentelemetry-specification#2115](https://github.com/open-telemetry/opentelemetry-specification/pull/2115) is proposing to introduce `app.name` as mandatory attribute for client side telemetry sources like browser apps or mobile apps, which then would not be treated as service (and with that would not have a `service.name`). `(app|service).name` (and `namespace` and `instance_id`) could be used as unique identifier.

4. [open-telemetry/opentelemetry-specification#2192](https://github.com/open-telemetry/opentelemetry-specification/pull/2192) is proposing to introduce `telemetry.source.*` attributes as a super-set to `service.*` and `app.*`.

This OTEP is proposing to choose from those approaches to uniquely identifying a telemetry source, or to find a unifying approach, since not all proposals are mutually exclusive.)

## Explanation

As stated in the Motivation with that unique identifier in place, it can be used at different places:

* Backend developers will have certainty which attributes they can use as unique identifier for the source when storing telemetry data.
* An UI can use it for visualization, especially as fallback if no other attribute is provided for that.
* The collector (and other processors) can use that identifier while processing traces, metrics, logs.
* An end-user could use that identifier for error handling and debugging, e.g. when a telemetry source is mis-configured, it's easier to identify it among others.  

## Internal details

As stated above, there are multiple approaches to obtain that common unique identifier. Depending on the approach, there are different ways to accomplish it:

1. Introduce `telemetry.sdk.instance_id` (or similar) and make it mandatory. Make `service.name` only mandatory for backend services. Other telemetry sources can make different attributes mandatory, like `app.name`. Optionally, remove `service.instance_id` from `service.*`

2. Introduce a broad definition of the term _Service_ in the glossary. Unique identification could be achieved by (1) or making `service.name`, `service.namespace`, `service.instance_id` mandatory for all telemetry sources.

3. Narrow down the definition for the term _Service_ to backend services. Make `service.name` only mandatory for backend services. Other telemetry sources can make different attributes mandatory, like `app.name` and provide a definition for their term, like `App` in the glossary. Unique identification could be achieved by (1) or having `(service|app).instance_id` and `(service|app).namespace` made mandatory as well.

4. Introduce `telemetry.source.name`, `telemetry.source.namespace` and `telemetry.source.instance_id`. Make some or all of them mandatory for all telemetry sources. Different telemetry sources can add additional attributes in namespaces like `service.*` and `app.*`.

## Trade-offs and mitigations

All potential approaches provide different trade-offs:

1. This will not introduce any breaking changes.

2. This will not introduce any breaking changes, but end-users might get confused by calling their telemetry a service while they think of it as an app or different (see future possibilities)

3. This may introduce a breaking change with `service.name` being not mandatory anymore in that broad sense. This would need further investigation. Also, this approach might lead to further additional sets of attributes which will be used by different telemetry sources for unique identification (devices, cronjobs, bots, ...)

4. This will introduce a breaking change because `service.name` will be replaced with `telemetry.source.name`. This could be mitigated by a fallback mechanism, e.g. if `telemetry.source.name` is not provided check `service.name`.

This list is not exhaustive, There are potentially more trade-offs per approach.

## Open questions

* What approach provides the most benefit and the least breaking changes to the current specification?
* Are there further approaches missed by the author?

## Future possibilities

While the discussion right now is between backend and frontend services, in the future additional telemetry sources like different kinds of devices could be introduced and run into a similar situation that `service` is not the appropriate term.
