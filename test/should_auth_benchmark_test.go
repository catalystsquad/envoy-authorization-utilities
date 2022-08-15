package test

import (
	"encoding/json"
	"github.com/catalystsquad/envoy-authorization-utilities/pkg"
	v3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"testing"
)

func BenchmarkShouldAuthUrlPathMatch(b *testing.B) {
	// set configuration
	settingsJson := []byte(`{"hostSettings": {"test.com": {"ignorePaths": ["/some/path/here/*"]}}}`)
	var authUtils pkg.AuthorizationUtils
	err := json.Unmarshal(settingsJson, &authUtils)
	if err != nil {
		panic(err)
	}
	req := &v3.CheckRequest{
		Attributes: &v3.AttributeContext{
			Request: &v3.AttributeContext_Request{
				Http: &v3.AttributeContext_HttpRequest{
					Host: "test.com",
					Path: "/some/path/here/do/some/stuff",
				},
			},
		},
	}
	for i := 0; i < b.N; i++ {
		result := authUtils.ShouldIgnoreRequest(req)
		if !result {
			panic("result should be true")
		}
	}
}

func BenchmarkShouldAuthGraphqlOperationMatch(b *testing.B) {
	// set configuration
	settingsJson := []byte(`{"hostSettings": {"test.com": {"ignorePaths": ["/some/path/here/*"], "ignoreGraphqlOperations": ["doThing"]}}}`)
	body := "{\n  \"operationName\": \"DoThing\",\n  \"variables\": {},\n  \"query\": \"query DoThing {\\n  doThing {\\n    result {\\n      name\\n      place\\n}\\n}\\n}\"\n}"
	var authUtils pkg.AuthorizationUtils
	err := json.Unmarshal(settingsJson, &authUtils)
	if err != nil {
		panic(err)
	}
	// create request that should be skipped based on matching ignore paths
	req := &v3.CheckRequest{
		Attributes: &v3.AttributeContext{
			Request: &v3.AttributeContext_Request{
				Http: &v3.AttributeContext_HttpRequest{
					Host: "test.com",
					Path: "/path/not/matched",
					Body: body,
				},
			},
		},
	}
	for i := 0; i < b.N; i++ {
		result := authUtils.ShouldIgnoreRequest(req)
		if !result {
			panic("result should be true")
		}
	}
}
