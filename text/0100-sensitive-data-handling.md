#  Sensitive Data Handling

By default, OpenTelemetry libraries never capture potentially-sensitive data, except for the full URL.

## Motivation

Many businesses are under regulatory or contractual constraints on how they process sensitive data (e.g., GDPR, 
PCI, HIPPA).  If we are "safe by default" and manage this issue well, it will help greatly with 
OpenTelemetry adoption in many markets.

## Explanation

By default, our java auto-instrumentation library does not capture your app's credit card numbers, taxpayer IDs, 
personal medical data, users' passwords, etc.  Unless you put them in the URL.  If you have put them in the URL, we provide a way to fix that.

How do we do that?  For example, we don't know the semantics of the SQL your app sends to its database(s).  Some of the data might be 
sensitive, and some of it might not.  What we do is remove all 
*potentially* sensitive values.  This means that sql like `SELECT * FROM USERS WHERE ID=123456789` gets turned into
`SELECT * FROM USERS WHERE ID=?` when it gets put into our `db.statement` attribute.  Since our library doesn't know anything about the 
meaning of `ID` values, we decided to be "safe by default" and drop all numeric and string values entirely from all SQL.  So what gets 
captured in our system is just table names, field names, clause structures, etc. (metadata about your data's structure and how you access it).

Yes, this also has the pleasant side effect of reducing the cardinality of these dimensions, making it easier to build metrics or indices on them. 
If you want to reduce it further, our collector has features for that.  Again, our design goal is to protect you from 
liability or exposure due to captured sensitive data.

We use similar techniques for all other protocols/query/command mechanisms.  One exception is how we treat URLs.  We will by
default capture the full URL (including the query and fragment) in `http.url`.  Note that this will not include
the contents of form submissions or XHR posts.  Since many products/features log the entire URL already 
(like in Apache's access log), developers already need to consider this.  If you are putting sensitive data in the URL,
we offer configuration options to work around this.

If your app doesn't process any sensitive data at all and you want to see the raw sql values, you can configure that.  The documentation
will have some bold/red warnings telling you to be sure you know what you're doing.

I'd like to repeat here that I'm using SQL as an *example* of how we treat potentially-sensitive data in our ecosystem.  We treat every other protocol/query format
with the same care.
 
## Internal details

I'm looking here for broad agreement of a design principle, rather than feedback on any specific low-level string manipulation technique to
achieve these goals.

That said, I have worked on sql normalization at three prior APM companies and am working on contributing a simple first version of one for the 
opentelemetry-auto-instr-java repo.  It is based on using a lexer to parse out sql numeric and string literals and replacing them with a `?`, exactly
as described above and done by many APM products on the market.

One interesting technical point to raise is what to do when our normalization/scrubbing knows it isn't working properly (for
example, encountering an ambiguous string it doesn't know how to parse).  In some cases a safe fallback approach can be found (e.g., apply a 
less-sophisticated regex that bluntly removes too much from the query), but in other cases we may need to, e.g., throw an internal exception and 
interpret that as a requirement to not add the 
attribute in the first place.  In pseudo-code:
```
  try {
      span.setAttribute("sensitive.field", normalize(sensitiveValue));
  } catch (CannotNormalizeException e) {
     // don't emit sensitive.field since we can't normalize it safely
  }
```

This work will affect all auto-instrumenting libraries, as well as any official (or recommended) manually wrapped libraries that are 
produced as part of the ecosystem.  As appropriate to each such library, features designed to properly handle sensitive data should
be highlighted in the documentation.  Sticking with the SQL example, for the Java auto library this could take the form of an additional column
on the "Supported Frameworks" table with a hyperlink to the documentation describing exactly what we do:

| Library/Framework                                                                                    | Versions                       | Notes                                      |
|------------------------------------------------------------------------------------------------------|--------------------------------|--------------------------------------------|
| [JDBC](https://docs.oracle.com/en/java/javase/11/docs/api/java.sql/java/sql/package-summary.html)    | Java 7+                        | (link to docs about sql normalization)     |


## Trade-offs and mitigations

In some cases, proper "safe by default" behavior might involve more intensive processing than can reasonably be performed inline with the application.  There have 
already been some discussions about having span (and attribute) processing logic run in-process but off-thread in the application (i.e., by a configurable span 
processing pipeline, which would use separate threads for its work).  For sql normalization (my first/example concern), I don't believe this is necessary, but 
further work on this issue may put pressure here.

One drawback of this "safe by default" approach is that it puts a configuration burden on some users who are not handling sensitive data but
want to have as much visibility as possible into their system.  Staying with the sql normalization example however, the vast majority of sql in the world doesn't 
have different performance characteristics based on the exact values passed in, but on the number of them, or the complexity of the `WHERE` clauses, 
the number of values in an `IN` list, presence of an index on fields, etc.

This "safe by default" behavior should be applied for a "default install", including the collector.  While individual instrumentation
libraries can (and are encouraged to) apply appropriate scrubbing themselves, the collector will act as a cross-language backstop, with appropriate
semantic transformations based on received data (e.g., looking at the`db.type` field to determine whether to treat a string
as sql or a mongo query).  This approach should reduce maintenance costs across the ecosystem as well as reduce pressure on needing 
the in-app span processing pipeline mentioned above.

## Prior art and alternatives

Nearly every commercial APM product on the planet has features like the sql normalization discussed above (though some take a different approach in the name of reduced
cardinality).  Most of them have landed on a similar answer for their treatment of URLs (though, again, some do different things for metrics features vs. tracing features).

## Open questions

I am not personally familiar with the OpenTelemetry stance on one value that is potentially-sensitive under some regulatory frameworks: client IP addresses
of users accessing a system.  This should be discussed, perhaps separately.

I am not an expert at all the protocols/query formats that our libraries support.  Doing this well will require a thorough audit and cooperation/discussion across
the community (again, ideally in separate issues/PRs for each such feature).

## Future possibilities

We might want to migrate some of this functionality into external libraries such that manual instrumentation (and manually-constructed library wrappers) can take advantage of these capabilities.

Additionally, work on more advanced configuration around this puts pressure on having our span processors be highly flexible, whether these are in-process or in the collector.
