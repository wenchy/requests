# requests
[![GoDoc](https://godoc.org/github.com/carlmjohnson/requests?status.svg)](https://pkg.go.dev/github.com/Wenchy/requests) 
[![Go Report Card](https://goreportcard.com/badge/github.com/carlmjohnson/requests)](https://goreportcard.com/report/github.com/Wenchy/requests)
[![Coverage Status](https://codecov.io/gh/Wenchy/requests/branch/master/graph/badge.svg)](https://codecov.io/gh/Wenchy/requests) 
[![License](https://img.shields.io/github/license/Wenchy/requests?style=flat-square)](https://opensource.org/licenses/MIT)

An elegant and simple HTTP client package, which learned a lot from the well-known Python package [Requests: HTTP for Humans™](https://requests.readthedocs.io/).

## Why not just use the standard library HTTP client?

[Brad Fitzpatrick](https://github.com/bradfitz), long time maintainer of the **net/http** package, wrote [Problems with the net/http Client API](https://github.com/bradfitz/exp-httpclient/blob/master/problems.md). The four main points are:

> - Too easy to not call Response.Body.Close.
> - Too easy to not check return status codes
> - Context support is oddly bolted on
> - Proper usage is too many lines of boilerplate

**requests** solves these issues by:

- always closing the response body,
- checking status codes by default,
- optional `context.Context` parameter,
- and simplifying the boilerplate.

## Features

- [x] Keep-Alive & Connection Pooling
- [x] International Domains and URLs
- [ ] Sessions with Cookie Persistence
- [ ] Browser-style SSL Verification
- [x] Automatic Content Decoding
- [x] Basic/Digest Authentication
- [x] Elegant Key/Value Cookies
- [x] Automatic Decompression
- [x] Unicode Response Bodies
- [x] HTTP(S) Proxy Support
- [ ] Multipart File Uploads
- [ ] Streaming Downloads
- [x] Connection Timeouts
- [ ] Chunked Requests
- [ ] .netrc Support

## Examples

### Simple GET into a text string

<table>
<thead>
<tr>
<th><strong>code with net/http</strong></th>
<th><strong>code with requests</strong></th>
</tr>
</thead>
<tbody>
<tr>
<td>

```go
req, err := http.NewRequestWithContext(
    ctx, http.MethodGet,
        "http://example.com", nil)
if err != nil {
    // ...
}
res, err := http.DefaultClient.Do(req)
if err != nil {
    // ...
}
defer res.Body.Close()
b, err := io.ReadAll(res.Body)
if err != nil {
    // ...
}
s := string(b)
```
</td>
<td>

```go
var txt string
r, err := requests.Get("http://example.com",
            requests.ToText(&txt))
if err != nil {
    // ...
}
```

</td>
</tr>
<tr><td>14+ lines</td><td>5 lines</td></tr>
</tbody>
</table>


### POST a raw body

<table>
<thead>
<tr>
<th><strong>code with net/http</strong></th>
<th><strong>code with requests</strong></th>
</tr>
</thead>
<tbody>
<tr>
<td>

```go
body := bytes.NewReader(([]byte(`hello, world`))
req, err := http.NewRequestWithContext(
    ctx, http.MethodPost, 
    "http://example.com", body)
if err != nil {
    // ...
}
req.Header.Set("Content-Type", "text/plain")
res, err := http.DefaultClient.Do(req)
if err != nil {
    // ...
}
defer res.Body.Close()
_, err := io.ReadAll(res.Body)
if err != nil {
    // ...
}
```

</td>
<td>

```go
r, err := requests.Post("http://example.com",   
        requests.Data(`hello, world`))
if err != nil {
    // ...
}
```

</td>
</tr>
<tr><td>15+ lines</td><td>4+ lines</td></tr></tbody></table>

### GET a JSON object

<table>
<thead>
<tr>
<th><strong>code with net/http</strong></th>
<th><strong>code with requests</strong></th>
</tr>
</thead>
<tbody>
<tr>
<td>

```go
u, err := url.Parse("http://example.com")
if err != nil {
    // ...
}
req, err := http.NewRequestWithContext(
        ctx, http.MethodGet,
        u.String(), nil)
if err != nil {
    // ...
}
resp, err := http.DefaultClient.Do(req)
if err != nil {
    // ...
}
defer resp.Body.Close()
b, err := io.ReadAll(res.Body)
if err != nil {
    // ...
}
var res JSONResponse
err := json.Unmarshal(b, &res)
if err != nil {
    // ...
}
```
</td><td>

```go
var res JSONResponse
r, err := requests.Post("http://example.com",
            requests.ToJSON(&res))
if err != nil {
    // ...
}
```

</td>
</tr>
<tr><td>22+ lines</td><td>5 lines</td></tr></tbody></table>

### POST a JSON object and parse the response

```go
req := JSONRequest{
    Title:  "foo",
    Body:   "baz",
    UserID: 1,
}
var res JSONResponse
r, err := requests.Post("http://example.com",
            requests.JSON(&req),
            requests.ToJSON(&res))
if err != nil {
    // ...
}
```

### Set custom header and form for a request

```go
// Set headers and forms
r, err := requests.Post("http://example.com", 
            requests.HeaderPairs("martini", "shaken"),
            requests.FormPairs("name", "Jacky"))
if err != nil {
    // ...
}
```

### Easily manipulate URLs and query parameters

```go
// Set parameters
r, err := requests.Get("http://example.com?a=1&b=2", 
            requests.ParamPairs("c", "3"))
if err != nil { /* ... */ }
// URL: http://example.com?a=1&b=2&c=3
```

### Dump request and response

```go
var reqDump, respDump string
r, err := requests.Get("http://example.com", 
            requests.Dump(&reqDump, &respDump))
```
