# Addressing discrepancy related to "error" span tag between opentracing and opencensus spec

Adding error span tag from opentracing specification to opentelemtery specification.  

## Motivation

Why should we make this change? 
    Tracing backends such as jaeger adhere to opentracing spec https://github.com/opentracing/specification/blob/master/semantic_conventions.md#span-tags-table. Where "error" span tag is a boolean value that helps in filtering spans those are in error state. Since opentelemetry doesn not have "error" span tag in the spec, error span tags are passed as is to the exported backend. According to opentracing spec error tag is boolean, but when zipkin span is translated to jaeger span format error tag continues to be a string carrying complete error message. This violates opentracing spec and also breaks jaeger functionality, as jaeger expects error to be boolean and error.message to be String type that carries description of the error message.    
    For example  : Zipkin spans have "error" tag as a string datatype and they carry short error message
    # Zipkin Span
    {
        "traceId": "631ef3c9f9250805","id": "631ef3c9f9250805","kind": "SERVER","name": "http:/one","timestamp": 1580830219488000,"duration": 29312,"localEndpoint": {"serviceName": "foo-service", "ipv4": "192.168.1.8", "port": 9001 },
        "tags": {
            "error": "Request processing failed; nested exception is org.springframework.web.client.HttpServerErrorException: 500",
            "http.host": "localhost", "http.method": "GET", "http.path": "/one", "http.status_code": "200", "http.url": "http://localhost:9001/one", "mvc.controller.class": "FooController", "mvc.controller.method": "one","spring.instance_id": "192.168.1.8:order-service-sleuth:9001"
        }
    }

    # Jaeger Span [Actual] OpenTelemetry exported jaeger span.
    {"traceIdLow":8380789575664951000,"traceIdHigh":0,"spanId":8380789575664951000,"parentSpanId":0,"operationName":"http:/one","flags":0,"startTime":1580830714594000,"duration":79701,
    "tags":[{"key":"http.url","vType":"STRING","vStr":"http://localhost:9001/one"},{"key":"mvc.controller.class","vType":"STRING","vStr":"FooController"},{"key":"mvc.controller.method","vType":"STRING","vStr":"one"},
    {
        "key":"error",
        "vType":"STRING",
        "vStr":"Request processing failed; nested exception is org.springframework.web.client.HttpServerErrorException: 500"
        },
        {"key":"http.method","vType":"STRING","vStr":"GET"},{"key":"http.host","vType":"STRING","vStr":"localhost"},{"key":"http.path","vType":"STRING","vStr":"/one"},{"key":"spring.instance_id","vType":"STRING","vStr":"192.168.1.8:order-service-sleuth:9001"},{"key":"http.status_code","vType":"STRING","vStr":"200"},{"key":"span.kind","vType":"STRING","vStr":"server"},{"key":"status.code","vType":"LONG","vLong":2}
        ]
    }

    # Jaeger Span [Expected]
    {"traceIdLow":8380789575664951000,"traceIdHigh":0,"spanId":8380789575664951000,"parentSpanId":0,"operationName":"http:/one","flags":0,"startTime":1580830714594000,"duration":79701,
    "tags":[{"key":"http.url","vType":"STRING","vStr":"http://localhost:9001/one"},{"key":"mvc.controller.class","vType":"STRING","vStr":"FooController"},{"key":"mvc.controller.method","vType":"STRING","vStr":"one"},
    {
        "key":"error",
        "vType":"BOOL",
        "vBool":true
        },
        {"key":"error.message",
        "vType":"STRING",
        "vStr":"Request processing failed; nested exception is org.springframework.web.client.HttpServerErrorException: 500"
        },
        {"key":"http.method","vType":"STRING","vStr":"GET"},{"key":"http.host","vType":"STRING","vStr":"localhost"},{"key":"http.path","vType":"STRING","vStr":"/one"},{"key":"spring.instance_id","vType":"STRING","vStr":"192.168.1.8:order-service-sleuth:9001"},{"key":"http.status_code","vType":"STRING","vStr":"200"},{"key":"span.kind","vType":"STRING","vStr":"server"},{"key":"status.code","vType":"LONG","vLong":2}
        ]
    }

What new value would it bring? 
    Adding below tags to opentelemtery spec would help in non lossy translation of spans
    error	            bool	true if and only if the application considers the operation represented by the Span to have failed
    error.message       String  Short description of the error
    
What use cases does it enable?
    This is in accordance with opentracing spec and also fix issues w.r.t lossy translations as explained above.

## Explanation
    
    Adding below tags to opentelemtery spec would help in non lossy translation of spans
    error	            bool	true if and only if the application considers the operation represented by the Span to have failed
    error.message       String  Short description of the error

Refer above motivation section for examples and more details.
    

## Internal details

From a technical perspective, how do you propose accomplishing the proposal? In particular, please explain:

* How the change would impact and interact with existing functionality
  This would need a minor change in translation logic accross all the formats.