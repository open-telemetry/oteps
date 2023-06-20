# Payload Collection APIs

Define APIs in OpenTelemetry for capturing payload data.

## Motivation

This OTEP proposes adding APIs and semantic conventions for handling payload data,
and capturing it in traces and logs, based on existing attributes.

Instrumenting payload data seems to be a feature requested by many users
[[1]](https://github.com/open-telemetry/opentelemetry-js/discussions/3446)
[[2]](https://github.com/open-telemetry/opentelemetry-specification/issues/1062)
[[3]](https://stackoverflow.com/questions/75260865/capture-request-body-using-opentelemetry-in-node-js)
[[4]](https://github.com/open-telemetry/opentelemetry-specification/issues/376#issuecomment-1227501082)

Tracing companies like Epsagon (now part of Cisco), and Aspecto (now part of SmartBear), have implemented payload collection in their custom instrumentation libraries.
This capability was leveraged by many customers, including big companies that enabled it in production (and testing) environments.
It was implemented in multiple runtimes (NodeJS, Python, Go, and others) and instrumentations (e.g. HTTP, AWS SDK, and many DB SDKs).

What do we mean by payload data? While it’s hard to precisely define, the general guidelines are:

- It includes the content of a message, rather than metadata or headers
- It includes information from the “data plane”, rather than the “control plane”
  Some example of payload data includes HTTP bodies, database queries (including items read or written), and messages produced or consumed from a queue.

Currently, the support in OpenTelemetry for collecting such data is quite limited. We propose in this OTEP to add an API that will define a standard and extendable way for collecting payloads in traces.
The implementation of this API is based on existing attributes, with proper semantic conventions.

The possible value from collecting payloads is substantial. Using payload data, OpenTelemetry users can troubleshoot applications much more effectively in many cases - they can use that to understand exact data flows in their systems, reproduce problematic requests, or search for traces by specific payload information.
Payload data could also be used to create ad-hoc analytics which helps monitoring the system.

As many could argue, there are many challenges for effieciently collecting payloads from production applications in a generic way: handling sensitive data, performance implications, telemetry vendors support and others. Nevertheless, we think that this should not block OpenTelemetry users from manually collecting payload data already, whenever it fits their requirements.

Having a standard semantic conventions for payload data can allow monitoring tools and others to use them OOTB (if they exist), without relying on custom attribute names.

Ultimately, it could be beneficial if OpenTelemetry instrumentation SDKs will have OOTB optional payload collection as well, though this OTEP does NOT propose that, or takes into account that it will be added in the future.

## Explanation

While payload data can just be a binary buffer, most of the times it has a defined structure.
For modern applications and APIs, the payload data is usually encoded as one of:

- JSON
- YAML
- Protobuf
- Avro
- Plain strings
- Blob

Whenever the payload data has some meaningful structure (everything except a blob), that structure will be reflected in the semantic attribute of the relevant span. That way, processors downstream can easily access and manipulate the payload data.

## Internal details

### Semantic conventions

We propose specifying a consistent naming for payload content attributes, and its
related metadata. These conventions could help processors, backends and users to
handle this kind of data.

In this OTEP we only give ideas and recommendations for it, though
a further discussion will be required for choosing the ideal naming when
writing the actual specifications.

The conventions we propose are added as a postfix to the base attribute name
(for example, `http.request.body`):

- **Data attribute**: `<attribute>.content`. Holds the decoded content of the payload data. Alternatives: `.data`, `.payload`.
- **Size attribute**: `<attribute>.size`. Holds the number of bytes of the encoded payload data. Alternatives: `.length`, `.bytes`.
- **Encoding attribute**: `<attribute>.encoding`. Holds the original attribute encoding type.
  Predefined values should be declared (though users may decide using custom values as well).
  For example - `json`, `protobuf`, `avro`, `utf-8`.

Adding each attribute should not be dependent on the others in any way.

### API

We propose specifying APIs the will abstract the handling and capturing of
payload data, to allow customizing this functionality and evolving it over time.

This functionality includes:

- Adding attributes with enforced semantics
- Enforcing limits - e.g. shortening long strings
- Applying central configuration over the general functionality
- Parsing raw payload buffers - converting payload data from bytes to OTel attributes, for
  supported encodings, in an extendable fashion.
- Aggregating multiple data chunks (to assist with handling asynchronous buffers)

The APIs given here are only a draft to describe general characteristics.
Like in the semantics, a further discussion will be required before creating actual specifications.

#### `Payload` class

We propose adding a class that will provide the related functionality for handling payload data.
Both traces and logs could use instances of this class, which will assist sharing related functionality.

For simplicity, we will show a basic example of this class in Python. In reality, an API and its implementation should be separated.

```python
class Payload:
    def __init__(self, content: Any, encoding=None, size=None):
        self.content = self._parse_payload_content(content)
        self.encoding = encoding
        self.size = size

    @staticmethod
    def _parse_payload_content(content: Any) -> Any:
        # Potentially altering of the content could be added, such as
        #  format conversion, shortening, filtering, etc.
        return content

    @classmethod
    def from_binary(
        cls,
        payload_buffer: bytes,
        encoding: str = 'utf-8',
    ) -> Payload:
        if encoding == 'utf-8':
            content = payload_buffer.decode('utf-8')
        elif encoding == 'utf-16':
            content = payload_buffer.decode('utf-16')
        elif encoding == 'json':
            content = json.loads(payload_buffer.decode('utf-8'))
        # ... other supported encodings
        else:
            raise ValueError("Unsupported payload encoding %s", encoding)

        return cls(content, encoding=encoding, size=len(payload_buffer))
```

### Payload attributes in Traces

In order to efficiently collect payload data in traces, this proposal depends on the [support of maps and heterogeneous arrays in spans
attributes](https://github.com/open-telemetry/opentelemetry-specification/pull/2888).

For adding payload attributes, we propose adding a dedicated method for `Span` class:

```python

def add_payload(self, key: str, payload: Payload):
  self.set_attributes(
    {
      f"{key}.content": payload.content,
      f"{key}.encoding": payload.encoding,
      f"{key}.size": payload.size
    }
  )
```

### Payload attributes in Logs

The payload content could be technically collected both as the the log record body, or as an attribute.
We think that it's needed to specify the preferred way following a discussion in the logs SIG.
[This](https://github.com/open-telemetry/opentelemetry-specification/pull/2926) PR references this issue as well.

### Null value

Representing Null values in payload contents is required, as it part of JSON and many other formats
like Protobuf and Avro. Therefore we propose supporting it by spans and logs attributes APIs (while
in OTel proto schema it is already supported).

## Trade-offs and mitigations

### Handling payloads with large size

As payloads are saved as regular attributes, they must follow the defined [limits](https://github.com/alanwest/opentelemetry-specification/blob/eb551bc9cf50d93463f20585686194d62887e044/specification/common/README.md#attribute-limits).

Capturing payloads using appropriate APIs could assist in specifying different limits in the future (either
more are less strict than of 'regular' attributes), and mechanisms for shortening large payloads as well.

### Handling sensitive data

Payload data is likely to include sensitive information, such as credentials or PII.
Nevertheless, this OTEP does not propose capturing sensitive information by OTel instrumentations, and therefore
does not change the current state where users can already manually collect sensitive data.

In this regard, it also worth mentioning the proposals on handling sensitive data (<https://github.com/open-telemetry/oteps/pull/100> and
<https://github.com/open-telemetry/oteps/pull/187>), which may prove much useful for users collecting payloads.

## Alternatives

### Adding new "Payload Attributes"

We propose adding a new type of span attributes, called '**payload attributes**', intended for storing decoded payload data. This will be implemented by additions of new fields and data types to Span proto definition, and API methods to support it. SDKs and OTel collector will be updated to support these changes as well.

The API & functionality of current Span attributes will remain the same, as they will
still be used for collecting general-purpose, non-payload data. Therefore, the proposed changes are **non-breaking**. The only potential breaking change is regarding certain semantic conventions which may fit better as payload attributes, such as `db.statement`.

#### OTLP Updates

We describe a prototype of additions to OTLP, to support encoding JSON-like objects in spans, together with some
extra metadata regarding the original plain payload. This prototype is likely to be changed during specifications and formats discussions
but hopefully could set basic characteristics.

We propose using a similar structure to the native protobuf [Struct](https://github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/struct.proto#L51) message,
(which is a general representation of a JSON object), with some embedded metadata -

```protobuf
message Value {
    // The kind of value.
    oneof kind {
      // Represents a null value.
      NullValue null_value = 1;
      // Represents a double value.
      double number_value = 2;
      // Represents a string value.
      string string_value = 3;
      // Represents a boolean value.
      bool bool_value = 4;
      // Represents a structured value.
      MapValue map_value = 5;
      // Represents a repeated `Value`.
      ListValue list_value = 6;
    }

    // Set only if the original value is shortened, for supported types -
    // > string: original number of characters
    // > ListValue: original number of items
    // > MapValue: original number of keys
    int64 original_size = 7;

    // Set only for MapValue, in case some of the keys were dropped
    repeated string dropped_keys = 8;
}

message MapValue {
    // Note - we can consider using a repeated key-value for performance
    map<string, Value> fields = 1;
}

message ListValue {
    repeated Value values = 1;
}
```

Then, we define a payload attribute which also includes extra metadata regarding the encoding -

```protobuf
message PayloadAttribute {
    string key = 1;
    Value value = 2;

    // Optional - the original bytes encoding type of this value (e.g. json/yaml/avro/csv)
    string original_encoding = 2;
    // Optional - the size of the value as bytes encoded (including dropped data)
    int64 encoded_size = 3;
}
```

And the payloads attributes are added to the Span message as -

```protobuf
repeated PayloadAttributes payload_attributes = ...
```

#### API Example

Now let's see how we can define and use an API to set payload attributes
(this example uses Python):

```python
# Added method to `Span` class
def add_payload_attribute(
    key: str,
    # JSON-like object, supports types int/double/string/bool/None, and nested
    # Iterables or Mappings
    value,

    # Optional - the original bytes encoding type of this value
    original_encoding=None,

    # Optional - the size of the value as encoded to bytes (using `original_encoding`),
    # including dropped data
    encoded_size=None,

    # Optional - set only when collecting a shortened value of type string/array/map.
    # Supports nesting (see example).
    original_size=None,

    # Optional - set only for mapping type (or array of mappings), when some of
    # the original keys are dropped.
    # Supports nesting (see example).
    dropped_keys=None,
)
```

Usage examples:

```python
span.add_payload_attribute(
    'http.request.body',
    {'a': 'test', 'b': None, 'c':{'x': [1, 2, 3.4]}},
    original_encoding='json',
    encoded_size=47
)

span.add_payload_attribute(
    'unicode_string',
    '∑∫µ',
    original_encoding='utf-16',
    original_size=4; # Collected payload shortened to 3 chars
    encoded_size=8, # Of the non-shortened payload
)

span.add_payload_attribute(
    'very_long_string',
    1024 * 'x',
    original_size=2048,
)

span.add_payload_attribute(
    'long_mapping',
    {'x': 1024 * 'x', 'y': 1024 * [0], 'z': 'short'},
    original_size={'x': 2048, 'y': 2048},
    encoded_size=4128,
)

span.add_payload_attribute(
    'filtered_keys',
    {
        'data': {'user_id': '1a2b'}
    }
    dropped_keys={'data': ['user', 'password']}
    encoded_size=134
)
```

## Next steps

We propose the following plan for adding payload collection support:

- Updating specifications with the API support
- Adding OTLP support
- Updating API and SDK libraries. Exporters should support encoding the payload attributes as serialized JSON attributes, for backward compatibility
  (in OTLP and proprietary formats)
- Update Collector to support payload attributes. Exporters should similarly support JSON serialization.

At this point, users would be able to manually instrument their applications with payload data, and backends will be able to add support for that.

The next step would be to support automated payload collection by general instrumentations.
We propose that it will be configured as an ‘advanced’ feature that is not enabled by default.
This way, users will not be exposed to the possible risks, unless they explicitly configured payload collections in their application.

We could also add more capabilities to OpenTelemetry to better support this kind of payload collection, such as -

- Automated methods for limiting the amount of collected data
- APIs for classifying sensitive data
- Defined methods and tools for accessing IO buffers handled by instrumented code

## Open questions

### Skipping payload attributes decode

Especially for large and complex payload attributes, decoding the OTLP data into
memory objects could be expensive, while not necessary.
For example, an OTel Collector which receives and exports OTLP data may benefit
if could copy an encoded payload attribute buffer 'as-is' instead of decoding and encoding.

We may explore methods for doing so, which could require using a custom Protobuf decoder.

### Integrating with a future columnar OTLP encoding

As the [proposal](https://github.com/open-telemetry/oteps/pull/171) for a columnar OTLP encoding is being progressed, we should define how payload attributes could be a part of that.
