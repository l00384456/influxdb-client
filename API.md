# Client API Requirements

The API requirements are a detailed guide to creating an official
InfluxDB library. When creating a library for a new language, please
follow these guidelines.

## Definitions

The words **MUST**, **SHOULD**, and **MAY** will be used in this
document. Something labeled **MUST** must be implemented to be
considered a capable library.  Something with **SHOULD** is recommended,
but not required if it doesn't fit in for whatever reason or is
expensive to implement without adding an additional dependency or for
whatever reason it might be excluded. A **MAY** is permission to add
something additional to the library, but is not as strong as **SHOULD**.

## Philosophy

The following are guidelines for what features the client library should
expose. Very important for a client library is feeling like the library
is native to the language it is written in. For that reason, please
exercise judgment in the API and attempt to use standard conventions for
whichever language you are using and the culture of that programming
language. Good libraries will follow the guidelines, but will feel
natural in their language.

Please do not use code generators like SWIG even if there is a possible
C library that you could bind to. While it is fine to bind to a C
library if one exists (or to make your own) for a binding, automated
code generators typically produce hard to use code in the native
language and do not feel natural.

## Guidelines

* **MUST** support writing through the HTTP endpoint and UDP endpoint.
* **MUST** support writing bytes directly to each of these endpoints.
* **MUST** support multiple line protocols through an interface.
  * **SHOULD** have a default line protocol that is easily accessible
    through a constant.
    * **MAY** use this default line protocol as a default when none is
      provided.
* **MUST** support setting the content type when writing to the HTTP
  endpoint.
* **MUST** support setting the precision and consistency when writing to the
  HTTP endpoint.
  * **SHOULD** have the precision set on the protocol object.
* **MUST** support a buffered and timed writer.
  * **MAY** combine these into a single type of writer.
* **MUST** have the writer conform to the default IO interface if the
  language has one. <sup>[1](#footnote1)</sup>
* **MUST** support setting the database and retention policy for writes in
  the HTTP API.
* **MAY** support pooling write buffers for better memory usage and fewer
  allocations.
* **MUST** support returning the raw bytes from the `/query` endpoint
  through a `raw()` function.
  * **MUST** support querying the database for any content type.
    <sup>[2](#footnote2)</sup>
    * **SHOULD** support shorter aliases that are more user friendly.
* **MUST** have a `select()` function that returns a cursor.
  * The cursor **MUST** support skipping a result or series by retrieving the
    next result or series.
  * The cursor **MUST** report any errors it encounters.
  * The cursor **MUST NOT** buffer the entire response.
    <sup>[3](#footnote3)</sup>
  * The cursor **SHOULD** use one of the streaming serialization formats.
    <sup>[4](#footnote4)</sup>
  * The cursor **MUST** support iterating through each result, series, and
    row within a series independently.
  * The cursor **SHOULD** have support utility functions designed to make
    iterating or accessing the data easier or more convenient. These
    interfaces should follow the standard method of creating iterators
    within the language.
  * The cursor **MAY** support functions that convert the iterator into an
    array. The documentation **SHOULD** include a warning about these
    functions.
  * The cursor **MAY** support additional utility functions that enhance the
    programmer's efficacy.
* **MUST** have an `execute()` function that will discard all rows, but
  throw/return an error when one is encountered.
* **MAY** have additional `select()` or `execute()` wrappers that include
  language specific functionality. <sup>[5](#footnote5)</sup>
* **MAY** split the writer and query code into separate libraries.
  <sup>[6](#footnote6)</sup>
* **MUST** have a client object that contains the connection or
  connection information.
  * **SHOULD** have a connection pool shared by those using the same
    client object.
* **MUST NOT** use code generators like SWIG.

<a name="footnote1">1</a>: Dynamic languages do not typically have a set
IO class that is inherited from everyone, but they will have a standard
interface through duck typing. A dynamic language implementation should
conform to the same interface that the string buffer and file classes
share so that either the string buffer or the file classes can be used
as substitutes when writing line protocol. For static languages, it
should be possible to use the standard IO interfaces if they exist. Some
languages, like C, do not have a standard IO interface and cannot
fulfill this requirement. In these circumstances, an alternative (like
callbacks) should be used instead so the user can implement their own
integration.

<a name="footnote2">2</a>: The query service supports many potential
output formats. The client should not need to know about all of the
possible output types so it is future compatible. For that reason, it
should support being able to set the content type directly to something
like `application/json` so the client is capable of using newer output
formats that may not have an alias in the client library yet.

<a name="footnote3">3</a>: A client is allowed to buffer decoded points
if the client finds that it is more efficient to buffer additional
points in advance, but the entire response should not be buffered. If
upcoming points are buffered, the client should consider the fact that
it's possible those points will be discarded and that the work does not
necessarily need to be done. While this option is left open to the
individual client implementation, it is not recommended. The important
part to keep in mind is that the response should be decoded iteratively.

<a name="footnote4">4</a>: While this is a **SHOULD**, functionally, it is a
**MUST**. The requirements for the cursor are not feasible with the existing
output formats in 1.3 and require the new streaming format introduced in
1.4 (tentative).

<a name="footnote5">5</a>: This allows adding additional functionality
supported by the specific language. For example, in Go, a
`context.Context` can be passed to long running processes. It is
expected and encouraged that this type of language specific behavior be
used and encouraged for a better user experience.

<a name="footnote6">6</a>: For a majority of programming languages,
please do not do this. They are not used in memory constrained
environments where a single-purpose library would be of substantial
benefit and the additional complication of splitting common code that is
shared (or creating a core library that they both inherit from)
introduces far more complexity than the benefit gained by separating the
writer and querier into separate libraries. But, for some languages,
like C, that may be used for embedded purposes, the benefit of splitting
these into separate libraries may be useful. For these purposes, feel
free to create separate libraries for only the writer or only the
querier. In these situations, it is likely most useful to use
conditional compilation.
