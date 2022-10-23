# Payload Collection

Support payload collection for traces in OpenTelemetry.

## Motivation

This OTEP proposes to add support for collecting payload data in spans, by adding a non-breaking
functionality to trace API, and to OTLP. As we show in this proposal, adding such data using
the current APIs is limited and problematic.

What do we mean by payload data? While it’s hard to precisely define, the general guidelines are:

- It includes the content of a message, rather than metadata or headers
- It includes information from the “data plane”, rather than the “control plane”

Some example of payload data includes HTTP bodies, database queries (including items read or written), and messages produced or consumed from a queue.

The possible value from collecting payloads is substantial. Using payload data, OpenTelemetry users can troubleshoot applications much more effectively in many cases - they can use that to understand exact data flows in their systems, reproduce problematic requests, or search for traces using specific payload information while troubleshooting.

## Explanation

Most generally, payload data is just a binary buffer. Mostly it will have a defined encoding, used by the code pieces which work with it to decode its content. For modern applications, standard encoding is usually used. For strings, UTF-8 or ASCII is most common,
while other encodings are usually used to represent nested mappings, like JSON, YAML, Protobuf, Avro, and many more.

While collecting binary buffers may be useful in some cases, collecting decoded
objects could be much more useful as processors and backends will be able to access internal
fields in a standard and performant way. We also show other capabilities that could be
accomplished by using a standard encoding for payload data.

## Internal details

We propose adding a new type of span attributes, called '**payload attributes**', intended for storing decoded payload data. This will be implemented by additions of new fields and data types to Span proto definition, and API methods to support it. SDKs and OTel collector will be updated to support these changes as well.

The API & functionality of current Span attributes will remain the same, as they will
still be used for collecting general-purpose, non-payload data. Therefore, the proposed changes are **non-breaking**. The only potential breaking change is regarding certain semantic conventions which may fit better as payload attributes, such as `db.statement`.

### OTLP Updates

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
    int64 original_length = 7;
}

message MapValue {
    // Note - we can consider using a repeated key-value for performance
    map<string, Value> fields = 1;
    
    repeated string dropped_keys = 2; 
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

### API Example

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
    original_length=None,

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
    original_length=4; # Collected payload shortened to 3 chars
    encoded_size=8, # Of the non-shortened payload
)

span.add_payload_attribute(
    'very_long_string',
    1024 * 'x',
    original_length=2048,
)

span.add_payload_attribute(
    'long_mapping',
    {'x': 1024 * 'x', 'y': 1024 * [0], 'z': 'short'},
    original_length={'x': 2048, 'y': 2048},
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

## Trade-offs and mitigations

### Handling payloads with large size

In many cases, payloads could have very large sizes, which we should not support collecting by default.
There are a few capabilities we suggest to handle such cases:

* **Configurable limits**: We should define which limits should be set for collected payload attributes.
For example, we can limit the following (each should have a default value) -
  * String, array, and map length
  * Recursion depth
  * Optional - Total size, derived by summing the size of all internal values

* **Automatic truncating**: We can support automated truncating of payloads breaching size limits.
The challenge here is to truncate a too-large mapping value when other limits (of specific keys) are not breached.
For this case, we can define a method that truncates 'biggest' values first, to try and minimize
the loss of most relevant data. This still requires more work to define.

### Handling sensitive data

Payload data is likely to include sensitive information, such as credentials or PII.
There are already some proposals on how to handle such data, like <https://github.com/open-telemetry/oteps/pull/100> and
<https://github.com/open-telemetry/oteps/pull/187>, which are very much related.
A possible approach is to adopt changes based on this proposal for payload attributes only, hence not over-complicating
mostly simple attributes.

A further improvement is to add data classification to each `Value` object, to support the classification of
internal values. For example, an `"email"` field inside a mapping could be specifically classified.

Note that the separation itself of payload attributes from the rest is a logical classification to
be leveraged by processors and backends to apply custom logic. For example, OTel collector could support
dropping such attributes, and backends could use different encryption strategies.

For instrumentations that collect payloads, we propose using a "safe by default" strategy, as proposed in
<https://github.com/open-telemetry/oteps/pull/100> as well - payloads should support a "normalization"
which is safe to assume that removes any sensitive data (such as values in SQL statements).
When not possible, payloads will be collected only if explicitly configured by the user.

## Prior art

In Epsagon (now part of Cisco), we have already implemented payload collection in our previous (non-OTel) libraries.
This capability was leveraged by many customers, including big companies that enabled it in production (and testing) environments.
It was implemented in multiple runtimes (NodeJS, Python, Go, and others) and instrumentations (e.g. HTTP, AWS SDK, and many DB SDKs).

In OpenTelemetry, there are existing requests for supporting payload collection -

* <https://github.com/open-telemetry/opentelemetry-specification/issues/1062>
* <https://github.com/open-telemetry/opentelemetry-specification/issues/376#issuecomment-1227501082>

## Alternatives

There are other possible ways to encode payload data, using current Span attributes.
As we show, each has different drawbacks, and also missing the proposed metadata information (such as `original_encoding`, `encoded_size`
and future fields as well).

Also, explicitly separating payload attributes allows custom logic to be applied.
This is especially true while currently there isn't a structured way to classify such attributes

Potentially, we can update current Span attributes with all of the proposed functionality, though we argue
that it will introduce over-complexity where it is mostly not necessary for simple attributes.

### Encoding payload data as a serialized JSON attribute

The major drawback of this alternative is that altering the data requires expensive deserialization and serialization.
For example, this will be required by a simple processor that filters specific keys in a map.
Also, it requires backends (and potentially processors) to unnecessarily attempt deserialization of every string attribute.

### Supporting nested map values in Span attributes

There is an ongoing discussion for supporting nested map values in Spans API (see <https://github.com/open-telemetry/opentelemetry-specification/issues/376>), which will allow encoding JSON-like attributes similar to the current proposal.

This option will still lack the general advantages described, and also lacks `NULL` encoding for complete compatibility with JSON format.

### Flatteing nested JSON maps to multiple attributes

Another method for handling the lack of nested mappings in span attributes is to 'flatten' them into multiple attributes using dotted-string notation.
For example, this method is already being used to collect HTTP headers as multiple `http.request.headers.<x>` and `http.response.headers.<x>` (see
[specifications](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/http.md#http-request-and-response-headers)).

This way is problematic for various reasons:

* For complex and large payloads, the number of attributes added could be big, which will make it hard to work with for users.
  Also, especially in these cases, exploring the original nested structure is likely to be much easier for users.
* Complex payload structures could be mapped to dozens of tags, which become inefficient to encode.
  For example, consider the following payload -

    ```jsx
    {'main_key': {'sub_key0': 0, ... ,'sub_key99': 99}}
    ```

    Which will be encoded into -

    ```jsx
    {'main_key.sub_key0': 0, ..., 'main_key.sub_key99': 99}
    ```

    The repetition of `'main_key'` negatively affects the performance for handling these attributes (processing, network & storage volumes, etc.)

- Flatting the attributes may lose some data from the original mapping, when original keys include dots or when arrays are used.
  Consider the examples -

    ```jsx
    {
     'a': {
         'b.c': 100
     },
     'd': [
         {'x': 0, 'y': 10},
         {'x': 1, 'y': 11}
     ]
    }
    and
    {
     'a': {'b': {'c': 100}},
     'd': {'x': [0, 1], 'y': [10, 11]},
    }
    ```

    Both will be encoded to the same attributes:

    ```jsx
    {
     'a.b.c': 100,
     'd.x': [0, 1],
     'd.y': [10, 100]
    }
    ```

  Though a more sophisticated flattening could be used here, it will make the attributes more complicated for the user to work with.

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

### Skipping payload attributes decode**

Especially for large and complex payload attributes, decoding the OTLP data into
memory objects could be expensive, while not necessary.
For example, an OTel Collector which receives and exports OTLP data may benefit
if could copy an encoded payload attribute buffer 'as-is' instead of decoding and encoding.

We may explore methods for doing so, which could require using a custom Protobuf decoder.

### Integrating with a future columnar OTLP encoding**

As the [proposal](https://github.com/open-telemetry/oteps/pull/171) for a columnar OTLP encoding is being progressed, we should define how payload attributes could be a part of that.
