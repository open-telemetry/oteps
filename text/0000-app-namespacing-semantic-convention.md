# Application namespacing semantic convention

Adds a semantic convention for manual instrumentation that defines how to namespace custom attributes, starting with the `app` prefix.

## Motivation

Today, custom attributes from an app created through manual instrumentation have all kinds of names that are specific to the domain of that app. There are no conventions for how to organize this data, though.

Data organization is critical to being able to effectively analyze observability data and avoid collisions in names. Without a well-defined convention for developers to follow, their own means of organization can differ from developer to developer, leading to a need to do things like coalesce field names using `.`/`-`/`_`/etc, not to mention having a real possibility to have name collisions. to separate clauses in an attribute name. Instead, a well-documented convention for how to organize data and in what format removes a "need to think" like code formatting conventions can do.

## Explanation

Attributes created through manual instrumentation should use namespacing for organizational purposes, using `app.` as a prefix for the name. Sub-namespaces should use a period (`.`) to separate sub-namespaces.

Use as many layers of subnamespaces as needed to represent your app's hierarchy. For example, an eCommerce app may define attributes like this:

* `app.order.id`
* `app.cart.items.count`
* `app.shipping.cost.total`

The more specific the namespace, the less likely specific names can collide in your app.

DO use a period (`.`) to separate sub-namespaces.
DO NOT use characters other than a period to separate sub-namespaces.
DO NOT mix different characters to separate sub-namespaces.

## Internal details

This would be accomplished by defining the `app.` prefix as reserved for end-user apps in the specification. Additionally, a specification entry in relevant places (trace, metrics, logs, and resource semantic conventions).

Additionally, official documentation would be added:

* Code samples in documentation updated to follow the convention as needed
* Subsections in language SDK docs that explain how to add attributes will explain or link to this convention
* Dedicated section in documentation about best practices for organizing data can be written, linked to, and link to the specification

## Trade-offs and mitigations

The biggest negative to introducing this convention now is that anyone using a _different_ convention will be considered out of step with best practices. For them to follow best practices, they would need to rewrite their instrumentation names, which would likely break or complicate their analyses for a period of time. To mitigate this, several analysis tools offer ways to coalesce two field names into one, but this kind of functionality is not guaranteed to be present for everyone.

Another negative is that some developers may not like the idea of defining a convention for their own systems. It's not uncommon for developers to dislike standard coding conventions, for example, and we could expect to see a similar kind of backlash around defining this kind of convetions. There isn't a really a way to mitigate this kind of feedback; however, there is precedent that most developers tend to like conventions to follow rather than having to define their own all of the time. And so it's likely that most developers would appreciate following this convention.

Finally, like any standard convention, it's not much use if it's not documented. The proposed additions to documentation (even just having the right link the right places) are nontrivial. One potential way to mitigate this could be to use tooling to somehow suggest or enforce this convention (e.g., Analyzers in C#). However, that is a much bigger lift than defining a spec convention, and not all languages are guaranteed to have that kind of tooling available.

## Prior art and alternatives

Namespacing attributes is a documented best practice for [some vendors](https://docs.honeycomb.io/getting-data-in/data-best-practices/#namespace-custom-fields). These practices emerged before OpenTelemetry, but are equally applicable to OpenTelemetry.

This convention is also being used by the [OpenTelemetry Community Demo](https://github.com/open-telemetry/opentelemetry-demo/blob/main/src/checkoutservice/main.go#L265-L267).

## Open questions

Should the prefix `app.` be used, or should it be something else?

Should there be a limit places on number of sub-namespaces?

Should there be naming conventions for the names themselves? (e.g., `cartService` vs. `cart-service` vs. `cart_service`)

## Future possibilities

This change could also enable conventions for attribute naming within a sub-namespace.
