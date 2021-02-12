Library to parse HTTP traffic from packet captures.

It performs preliminary processing to convert raw packets into
`akinet.HTTPRequest` and `akinet.HTTPResponse` objects. In particular, it
dechunks bodies with `Transfer-Encoding: chunked`, but keeps `Content-Encoding`
unchanged.

Note that this library returns an error for non-chunked `Transfer-Encoding`.
This is due to the limitations of the go library used to read HTTP
request/response and it's not easy to fix short of writing our own HTTP parser.
