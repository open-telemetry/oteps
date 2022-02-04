# Environment Variable Resource Detection in Kubernetes

This OTEP proposes to standardize environment variables used to detect resources exposed through the [Kubernetes downward API](https://kubernetes.io/docs/tasks/inject-data-application/environment-variable-expose-pod-information/) to provide a consistent experience on Kubernetes across languages and the collector.

There is a mismatch between the format otel expects resources to be specified in for environment variables (a single, comma-deliniated list of key-value pairs in `OTEL_RESOURCE_ATTRIBUTES`), and the format Kubernetes exposes resource information in (one environment-variable for each attribute).  While it is possible to [use dependent environment variables](#alternative-using-dependent-environment-variables) to merge these resource attributes together into `OTEL_RESOURCE_ATTRIBUTES`, it lacks versioning, and is difficult to apply consistently to pods. The lack of specification has led to differing implementations across language SDKs and the collector.

## Background

Adding Kubernetes resources attributes to telemetry can be done one of three ways, depending on your telemetry collection pipeline:

1. The application detects its own resource attributes from environment variables
2. A "sidecar" collector container detects its own resource attributes from environment variables
3. A centralized collector detects resource attributes for many Kubernetes workloads by querying the kube-apiserver.

Method 3 is already implemented today as the `k8sattributesprocessor`, but requires running a collector, and adds additional load on the kube-apiserver. Methods 1 and 2 require substantial configuration today, which differs based on the language used, or if using a sidecar collector.

Not all attributes in the [k8s semantic conventions](https://github.com/open-telemetry/opentelemetry-specification/blob/main/semantic_conventions/resource/k8s.yaml) are supported by the Kubernetes downward API.  The OpenTelemetry attributes that could be supported ([api reference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#envvarsource-v1-core)) are:

* `k8s.pod.name`
* `k8s.pod.uid`
* `k8s.namespace.name`
* `k8s.node.name`

The current state of detecting downward-api-based resources differs, depending on where detection is being done:

* The collector's [`resourcedetectionprocessor`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourcedetectionprocessor#resource-detection-processor) only accepts `OTEL_RESOURCE_ATTRIBUTES` for env-var-based resource detection, which does not work with the downward API.
* The collector's [`resourceprocessor`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourceprocessor#resource-processor) can be used in conjunction with [environment variable expansion](https://opentelemetry.io/docs/collector/configuration/#configuration-environment-variables) as described in [this issue](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/570#issuecomment-1006885133).
* The [golang GKE detector](https://github.com/open-telemetry/opentelemetry-go-contrib/blob/26b22d30acb587d893b3a0fd3c8dd7895f79b352/detectors/gcp/gke.go) uses `NAMESPACE`, and `HOSTNAME` (pod name)
* The [ruby GKE detector](https://github.com/open-telemetry/opentelemetry-ruby/blob/8fe083cdf00bfe866ae89ea8f5fc06fd981ebf6a/resource_detectors/lib/opentelemetry/resource/detectors/google_cloud_platform.rb) uses `GKE_NAMESPACE_ID`, and `HOSTNAME` (pod name)
* The [javascript env var detector](https://github.com/open-telemetry/opentelemetry-js/blob/feea5167c15c41f0aeedc60959e36c18315c7ede/packages/opentelemetry-core/src/utils/environment.ts) uses `NAMESPACE`, and `HOSTNAME` (pod name)
* Others not listed do not attempt to detect Kubernetes resource information through environment variables.

## Motivation

A common set of environment variables across languages and the collector would enable a single pod `env` configuration to work correctly across languages and collector sidecars. At minimum, this would simplify our documentation and improve day-0 user experience.  But the Kubernetes API's extension points and the surrounding ecosystem of tools allow cluster administrators to go further: **They can configure these environment variables as cluster-level policy, and make OpenTelemetry resource detection "just work" on Kubernetes.**

## Explanation

The snippet of environment variables added the pod spec would look like:

```yaml
env:
- name: K8S_POD_NAME
   valueFrom:
     fieldRef:
       fieldPath: metadata.name
- name: K8S_POD_UID
   valueFrom:
     fieldRef:
       fieldPath: metadata.uid
- name: K8S_NAMESPACE_NAME
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: K8S_NODE_NAME
   valueFrom:
     fieldRef:
       fieldPath: spec.nodeName
```

### User Journey: Getting Started on Kubernetes

As a user new to Kubernetes and/or OpenTelemetry, I can copy-and-paste the above into my pod specification to make resource discovery work for me. I don't need to set up any complex infrastructure, know about the downward API or OpenTelemetry resource detection, or the specifics of the lauguage SDK I'm using to make it work.

### User Journey: Admission Webhooks

There are a number of existing Kubernetes policy engines that can enforce a consistent set of environment variables on pods (in addition to doing much more), such as [kyverno](https://github.com/kyverno/kyverno), [podpreset](https://github.com/redhat-cop/podpreset-webhook), [open policy agent](https://www.openpolicyagent.org/), etc., but they all use the same underlying Kubernetes mechanism: [Admisison Webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#admission-webhooks).

As a cluster administrator, I can use my preferred policy agent or my own mutating admission webhook to add the OpenTelemetry-specified set of environment variables to all pods subsequently created in the cluster. Application developers deploying to the cluster just need to use the OpenTelemetry SDK with the Kubernetes resource detector.

### User Journey: Kubernetes Config Management

Kubernetes templating tools, such as [Helm templates](https://helm.sh/docs/chart_template_guide/), allow users to easily re-use snippets of yaml in many places in their configuration.  A common set of environment variables would users to define the OpenTelemetry env var config once, and use it across all of their OpenTelemetry deployments.

**Note:** This pattern unfortunately isn't easy to use with kubernetes' built-in patching tools, [kustomize](https://kustomize.io/), because it doesn't [currently](https://github.com/kubernetes-sigs/kustomize/issues/1493) support patching all containers across deployments with a single patch.

## Internal details

Add an experimental k8s.md specification under specification/resource, which specifies how kubernetes resource detectors should detect the kubernetes semantic conventions:

SDK's should add a Kubernetes environment variable-based resource detector.  The Kubernetes resource detector and the collector's `resourcedetectionprocessor` should support the following environment variables, and map them to the following semantic conventions:

| Environment Variable | Semantic Convention |
| ---- | ---- |
| K8S_POD_NAME | k8s.pod.name |
| K8S_POD_UID | k8s.pod.uid |
| K8S_NAMESPACE_NAME | k8s.namespace.name |
| K8S_NODE_NAME | k8s.node.name |

Detection of the above semantic conventions using environment variables should not be done inside vendor-specific (e.g. GKE-specific) resource detectors.

## Prior art and alternatives

### Alternative: Detection required by SDKs

Similar to `OTEL_RESOURCE_ATTRIBUTES`, we could require the proposed environment variables to always be detected.  While this would further improve the ease of use, [Resource semantic conventions](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/semantic_conventions/README.md) (including Kubernetes) are still experimental, and the [SDK environment variable specification](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/sdk-environment-variables.md#general-sdk-configuration) are stable. Additionally, Kubernetes is [explicitly called out](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/sdk.md#detecting-resource-information-from-the-environment) as a custom resource detector.

### Alternative: Using HOSTNAME for k8s.pod.name

Many kubernetes detectors currently use `HOSTNAME` environment variable, which defaults to the Pod name. However, the `HOSTNAME` can be [modified in a few ways](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-hostname-and-subdomain-fields) in the pod spec. Kubernetes resource detectors may fall back to detecting the pod name using `HOSTNAME` if `K8S_POD_NAME` is not available, but this may cause user confusion in some cases.

`HOSTNAME` is also [truncated to 64 characters](https://github.com/kubernetes/kubernetes/issues/4825) on some operating systems.

### Alternative: Using Dependent Environment Variables

Kubernetes supports defining environment variables based on [dependent environment variables](https://kubernetes.io/docs/tasks/inject-data-application/define-interdependent-environment-variables/#define-an-environment-dependent-variable-for-a-container). We can use this to transform the environment variables discovered through the downward API into the `OTEL_RESOURCE_ATTRIBUTES` environment variable:

```yaml
- name: OTEL_RESOURCE_ATTRIBUTES
  value: k8s.pod.name=$(K8S_POD_NAME),k8s.pod.uid=$(K8S_POD_UID),k8s.namespace.name=$(K8S_NAMESPACE_NAME),k8s.node.name=$(K8S_NODE_NAME)
```

This approach has a few drawbacks:

* `OTEL_RESOURCE_ATTRIBUTES` may need to include additional attibutes other than those used for kubernetes. This makes it harder to apply the same environment variables across all pods.
* `OTEL_RESOURCE_ATTRIBUTES` doesn't have a way to specify the schema version it should apply to. This would conflict with the ability of [telemetry schemas](https://github.com/open-telemetry/oteps/blob/main/text/0152-telemetry-schemas.md#solution-summary) to convert from older versions of semantic conventions to newer ones.

### Alternative: Support multiple OTEL_RESOURCE_ATTRIBUTES_* variables

Proposed in [open-telemetry/opentelemetry-specification#2135](https://github.com/open-telemetry/opentelemetry-specification/issues/2135), SDKs would detect all environment variables starting with `OTEL_RESOURCE_ATTRIBUTES_`, and would add these to the detected resource attributes. The `OTEL_RESOURCE_ATTRIBUTES` prefix is trimmed, letters are lower-cased, and `_` are replaced with `.`.  For example, `OTEL_RESOURCE_ATTRIBUTES_K8S_NODE_NAME=foo` would become `k8s.node.name=foo`.

It would solve the Goals of this proposal by allowing the following to be used across all pods:

```yaml
env:
- name: OTEL_RESOURCE_ATTRIBUTES_K8S_POD_NAME
   valueFrom:
     fieldRef:
       fieldPath: metadata.name
- name: OTEL_RESOURCE_ATTRIBUTES_K8S_POD_UID
   valueFrom:
     fieldRef:
       fieldPath: metadata.uid
- name: OTEL_RESOURCE_ATTRIBUTES_K8S_NAMESPACE_NAME
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: OTEL_RESOURCE_ATTRIBUTES_K8S_NODE_NAME
   valueFrom:
     fieldRef:
       fieldPath: spec.nodeName
```

The primary drawback of this approach is that this method doesn't have a way to specify schema version it should apply to. This would conflict with the ability of [telemetry schemas](https://github.com/open-telemetry/oteps/blob/main/text/0152-telemetry-schemas.md#solution-summary) to convert from older versions of semantic conventions to newer ones. If that issue is resolved, this would be preferable to the proposed solution.

## Future possibilities

The [OpenTelemetry Operator](https://github.com/open-telemetry/opentelemetry-operator) already has a mutating admission controller, which handles automatic collector [sidecar injection](https://github.com/open-telemetry/opentelemetry-operator#sidecar-injection).  In the future, the OpenTelemetry Operator could be extended to support env var injection as well.
