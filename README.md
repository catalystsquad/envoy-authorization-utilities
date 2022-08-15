# envoy-authorization-utilities

This is a library designed to work with the Envoy external authorization [GRPC protocol](https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/auth/v3/external_auth.proto).
The intent is to allow configurable auth exceptions based on either path, or graphql operation in the request body. Instantiate the utility and then pass it a request object, then do whatever auth you need to based on the result.

## Supported mechanisms
* URL Path
  * Ignore requests that match a list of url paths, including wildcards 
  * Implemented using [URL Path](https://github.com/ucarion/urlpath)
* Graphql Query
  * Ignore requests that exactly match a list of graphql operation names

## Usage
Instantiate a new struct with your settings, typically by marshalling json to allow configuration via env var or some other mechanism, but you can also use the `NewAuthorizationUtils` constructor and pass a map of host settings.
```go
hosts := map[string]pkg.HostSettings{
  "myhost.com": {
    IgnorePaths: []string{"/some/path/*"},
    IgnoreGraphqlOperations: []string["yourQueryName"],
  },
}
authUtils, err := pkg.NewAuthorizationUtils(hosts)
```
Then pass the v3 request to `ShouldIgnoreRequest`
```go
shouldIgnore := authUtils.ShouldIgnoreRequest(req)
if shouldIgnore {
    // return a v3 check response indicating the request is authenticated
    return &v3.CheckResponse{
        Status: &status.Status{Code: int32(codes.OK)},
        HttpResponse: &v3.CheckResponse_OkResponse{
            OkResponse: &v3.OkHttpResponse{},
        },
    }
}
if !shouldIgnore {
    // your auth logic here
}
```

## Benchmarks
```shell
> go test -bench=.
goos: linux
goarch: amd64
pkg: github.com/catalystsquad/envoy-authorization-utilities/test
cpu: AMD Ryzen 9 5950X 16-Core Processor            
BenchmarkShouldAuthUrlPathMatch-32             	14814903	        90.45 ns/op
BenchmarkShouldAuthGraphqlOperationMatch-32    	 1669678	       754.0 ns/op
PASS
ok  	github.com/catalystsquad/envoy-authorization-utilities/test	3.420s
```