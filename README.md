# requests

An elegant and simple HTTP library for golang, built for human beings.

This package mimics the implementation of the well-known Python package [Requests: HTTP for Humans™](https://requests.readthedocs.io/).

Click to read [documentation](https://pkg.go.dev/github.com/Wenchy/requests@master).

## Why not just use the standard library HTTP client?

Brad Fitzpatrick, long time maintainer of the net/http package, [wrote an extensive list of problems with the standard library HTTP client](https://github.com/bradfitz/exp-httpclient/blob/master/problems.md). His four main points (ignoring issues that can't be resolved by a wrapper around the standard library) are:

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

### Simple GET into a string

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
resp, err := requests.Get("http://example.com")		
if err != nil {
    // ...
}
s := resp.Text()
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
resp, err := requests.Post("http://example.com",	
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
var post placeholder
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
res, err := http.DefaultClient.Do(req)
if err != nil {
	// ...
}
defer res.Body.Close()
b, err := io.ReadAll(res.Body)
if err != nil {
	// ...
}
err := json.Unmarshal(b, &post)
if err != nil {
	// ...
}
```
</td><td>

```go
resp, err := requests.Post("http://example.com")	
if err != nil {
    // ...
}
var res JSONResponse
if err := r.JSON(&res); err != nil {
    // ...
}
```

</td>
</tr>
<tr><td>22+ lines</td><td>8 lines</td></tr></tbody></table>

### POST a JSON object and parse the response

```go
req := placeholder{
	Title:  "foo",
	Body:   "baz",
	UserID: 1,
}
resp, err := requests.Post("http://example.com",
			requests.JSON(&req))
if err != nil {
    // ...
}
var res JSONResponse
if err := r.JSON(&res); err != nil {
    // ...
}
```

### Set custom headers and forms for a request

```go
// Set headers and forms
resp, err := requests.Post("http://example.com", 
			requests.HeaderPairs("martini", "shaken"),
			requests.FormPairs("name", "Jacky"))
if err != nil {
    // ...
}
```

### Easily manipulate URLs and query parameters

```go
// Set parameters
resp, err := requests.Get("http://example.com?a=1&b=2", 
                            requests.ParamPairs("c", "3"))
if err != nil { /* ... */ }
fmt.Println(u.String()) // http://example.com?a=1&b=2&c=3
```
