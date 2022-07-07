# Ephemeral Resource Attributes

Define a new type of resource attribute, ephemeral resources, which are allowed to change over the lifetime of the process. Existing resources are redefined as permanent resources, which must be present at SDK initialization and cannot be changed.

## Motivation

Server applications, which opentelemetry was initially designed around, simultaneously handle many unrelated transactions. Other types of applications, such as client applications, all events and transactions within the process are associated with a single user or activity. These applications often include "global" concepts which are important to telemetry. Examples include session ID, language preference, time zone and location data. These concepts must be represented as attributes in order to correctly report the state of client applications.

Since the state being recorded is global to the process, it matches our concept of a resource attributes, as a resources are applied to all telemetry emitted by the SDK. However, unlike our current concept of a resource attribute, these attributes may change their value over the life of the application. This OTEP proposes a mechanism for extending the concept of a resource, in order to efficiently and accurately record these attributes while still preserving the immutability requirements of currently defined resource attributes.

## Explanation

There are two types of resource attributes, **permanent** and **ephemeral**. Attributed which are labeled as permanent in the semantic conventions must be present when the SDK is initialized. They cannot be added or updated at a later date.

Resources are managed via a ResourceProvider. Setting an attribute on a ResourceProvider will cause that attribute value to be included in the resource attached to any signal generated in the future. Spans which have already been started, along with any telemetry which has already been passed to the export pipeline, will not have the new attribute value. Optionally, a check can be added to ensure that permanent resources are not modified after the SDK has started


## Internal details

### ResourceProvider

#### NewResourceProvider([resource], [validator]) ResourceProvider

NewResourceProvider instantiates an implimentation of the ResourceProvider interface. As argumentes, it optionally takes an initial set of resource attributes, and a validator. 

The ResourceProvider interface has the following methods

#### MergeResource(resource)

MergeResource creates a new resource, representing the union of the resource parameter and the resource contained within the Provider. The ResourceProvider holds a reference to the nwe resource.

#### SetAttribute(key, value)

SetAttribute functions the same as MergeResource, but only adds a single attribute.

#### GetResource() Resource

GetResource returns a reference to the current resource held by the ResourceProvider.

#### FreezePermanent()

FreezePermanent is called by the SDK one it has been stared. After FreezePermanent has been called, any calls to MergeResource or SetAttributes will only be applied if the validator acceptes the input.

#### Implementation Notes

For multithreaded systems, a lock should be used to queue all calls to MergeResource and SetAttribute. But the resource reference held by the ResourceProvider should be updated atomically, to prevent calls to GetResource from being blocked.

### SDK Changes

NewTraceProvider, NewMetricsProvider, and NewLogProvider now take a ResourceProvider as a parameter. For backwards compatibility, the Resource parameter remains functional. If both a resource and a resource provider as passed as parameters, the resource is merged into the ResourceProvider, then discarded.

FreezePermanent is then called by the provider.

Internally, providers hold a reference to the ResourceProvider, rather than a specific resource. When creating a signal, such as a span, metric, or log, GetResource() is called to obtain a reference to the correct resource to attach to the signal.

## Example Usage

Pseudocode example of a ResourceProvider in use. The resource provider is loaded with all available permanent resources, then passed to a TraceProvider. The ResourceProvider is also passed to a session manager, which updates an ephemeral resource in the background.

```
var resources = {“service.name” = “example-service”};

// Example of a deny list validator.
var validator = NewDenyListValidator(PERMANENT_RESOURCE_KEYS);

// Example of an allow list validator.
// This is useful for browser environments
// where loading a deny list would be too costly.
var validator = NewAllowListValidator([“session.id”]);

// The ResourceProvider is initialized with
// a dictionary of resources and a validator.
var resourceProvider = NewResourceProvider(resources, validator);

// The resourceProvider can be passed to resource detectors 
// to populate additional permanent resources.
DetectContainerResources(resourceProvider)

// The TraceProvider now takes a ResourceProvider.
// The TraceProvider calls Freeze on the ResourceProvider.
// After this point, it is no longer possible to update or add
// additional permanent resources.
var traceProvider = NewTraceProvider(resourceProvider);

// Whenever the SessionManager starts a new session
// it updates the ResourceProvider with a new session id.
sessionManager.OnChange(
  func(sessionID){
    resourceProvider.SetAttribute(“session.id”, sessionID);
  }
);

```

## Example Implementation

Pseudocode examples for a possible Validator and ResourceProvider implementation. Attention is placed on making the ResourceProvider thread safe, without introducing any locking or synchronization overhead to `GetResource`, which is the only ResourceProvider method on the hot path for OpenTelemetry instrumentation.

```

// Example of a simple validator.
class DenyListValidator{

// Attribute keys can be stored in any
// data structure with a fast implementation 
// for detecting set membership.
  Set denyList
  
  Validate(key){
    if(this.denyList.Contains(key)){
      return false;
    }
    return true;
  }
}

// Example of a thread-safe ResourceProvider
class ResourceProvider{
  *Resource resource
  Validator validator
  bool isFrozen 
  Lock lock
  
  GetResource(){
    return this.resource;
  }
  
  
  SetAttribute(){
    // All methods on a ResourceProvider which perform mutations
    // must share a lock.
    this.lock.Aquire();

    // Only perform validation after the ResourceProvider
    // has been frozen.
    if(this.isFrozen && !this.validator.Validate(key)){
      this.lock.Release();
      return;
    }

    // Because Resource objects are immutable, it is safe to call
    // SetAttribute without locking GetResource
    var mergedResource = this.resource.SetAttribute(key, value);
    
    // Because ResourceProvider only stores a reference to
    // a Resource, that reference can be replaced using an
    // atomic swap operation. This approach allows GetResource
    // to remain lock-free.
    AtomicSwap(this.resource, mergedResource)

    this.lock.Release();
  }
  
  // MergeResource essentially has the same implementation as SetAttribute.
  MergeResource(resource){
    this.lock.Aquire();

    if(this.isFrozen){
      // Every key in the resource must be validated
      foreach(resource, key, value){
        if(!this.validator.Validate(key)){
          this.lock.Release();
          return;
        }
      }
    }

  
    var mergedResource = this.resource.Merge(resource)
    AtomicSwap(this.resource, mergedResource)

    this.lock.Release();
  }
  
  FreezePermanent(){
    this.lock.Aquire();
    this.isFrozen = true;
    this.lock.Release();  
  }
}
```

## Trade-offs and mitigations

This change should be fully backwards compatible, with one potential exception: fingerprinting. It is possible that an analysis tool which accepts OTLP may identify individual services by creating an identifier by hashing all of the resource attributes. 

In this case, it is recommended that these systems modify their behavior, and choose a subset of permanent resources to use as a hash identifier.

## Prior art and alternatives

An alternative to ephemeral resources would be to create span, metrics, and log processors which attach these ephemeral attributes to every instance of every signal. This would not require a modification to the specification.

There are two problems to this approach. One is that the duplication of attributes is very inefficient. This is a problem on clients, which have a limited newtwork bandwidth. This problem is compounded by a lack of support for gzip and other compression algorithms on the browser.

The second problem is that it becomes difficult to distinguish between emphemeral resources and other types pf attributes. 

## Open questions

The primary open question is whether any common backends are hashing the resource to obtain a service identifier.

## Future possibilities

Ephemeral resource attributes will be critical feature for implementeting RUM/client instrumentation in OpennTelemetry. 

Other application domains may discover that they have process-wide state which affects their performance or otherwise changes code execution, which would be valuable to record as an ephemeral resource. For example, applications may have a drain or shutdown phase which affects the behavior of the application. The ability to identify telemetry data which occurs during this phase may be valuable to some end users.