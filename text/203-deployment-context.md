New Standard Attribute: `deployment.annotation`

Authors: Dan Bode (@bodepd), Bob Johnson (@nosnhojbob), Fredrick Roberts (@froberts71)

## Motivation

In our organization, `deployment.environment` is chosen from a small set
of known values, namely: `prod`, `test`, `nonprod`, `sbox`, `local`. This allows
for reports grouped by these values. For example: "Show me the number of spans
being sent to `deployment.environment=prod`." We keep this list small because
each of these environment types has specific meaning with regard to what types
of data and network traffic are available, security requirements, etc.

Our organization-wide has over 200+ distinct values set for `deployment.environment`.
This is because individual teams supporting applications (developers, SREs) think about
deployments in more fine-grained details than the organization does.

For example, `sbox` environments are deployed by individuals. Those individuals
need to be able to query for observability data for each of those deployments.
For example, "Show me the average latency for a recent deployment where
`deployment.environment = sbox and deployment.user = dan`."

For deployments into `test`, there is often additional context for what is being
tested. Some of these values are static (UAT, perf, integration) and others
vary for each deployment (feature being testing, what pull request is being tested).

To summarize, teams think of deployments using multiple dimensions. A high level
`environment` that maps to organizational definitions and more fine grained information
related to why something is deployed into that environment.

## Explanation

The standard attribute `deployment.annotation` is used to provide
additional context to the `deployment.environment`.

For example, you may wish to indicate which developer performed a `deployment` or
which test an `environment` is for.

| Attribute  | Type | Description  | Examples  | Required |
|---|---|---|---|---|
| `deployment.annotation` | string | An annotation to provide more information about why a deployment was made, such as what developer provisioned it or what test it is for | `bob` | No |

## Internal details

Since the proposed change is only for a standard attribute, this section is not relevant.

## Trade-offs and mitigations

Since the proposed change is only for a standard attribute, this section is not relevant.

## Prior art and alternatives

For these examples, consider a service with the following attributes:

| Application  | Team | Services  | Team Members | Environments |
|---|---|---|---|---|
| Shop | Shop team | cart, checkout | Alice, Bob, Charlie | sbox, prod, non-prod, test |

This example will consider an environment type: `development` in which users
deploy their own environments for testing.

### Use service.namespace

Instead of setting our service.namespace to Shop, we could instead have separate
`service namespace` for Shop-Alice, Shop-Bob, Shop-Charlie.

Using the namespace for each user means that we can't run reports across an
entire application, for example, if I want resource count across an entire
application, Shop, I want those to include sbox environments provisioned by
developers.

### Use deployment.environment

Instead of having a deployment.environment called `development`, have many environments called:
`development-alice`, `development-bob`, `development-charlie`. As mentioned above,
the organization needs a way to query across a higher level, and pre-defined
concept of an environment.

### Create attributes for each use case

Instead of creating a generic annotation field, instead capture each use case
and create specific attributes for them, ie:
* deployment.environment.dev.user
* deployment.environment.test.scenario

While this is much more clear as to intent of each environment, it has the
following issues:

* these two attributes do not capture all of the use cases that we have
* teams may already have conflicting terms for environment types, can OTEL be
prescriptive about what types of environments there are?

## Open questions

If deployment.environment is already an attribute, can I add another level to it? `deployment.environment.annotation`,
or do I need to come up with another name?

Provided that it makes sense to have a generic annotation, what should we name it?

Ideas:
* deployment.annotation
* deployment.details
* deployment.environment.annotation
* deployment.environment.details
* deployment.scenario
* deployment.environment.scenario

## Future possibilities

### Create specific attributes for each use case

Even if this proposal is accepted, the question around making this attribute
more specific for the use cases should continue to be considered.
