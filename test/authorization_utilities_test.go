package test

import (
	"encoding/json"
	"github.com/catalystsquad/envoy-authorization-utilities/pkg"
	v3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AuthorizationSuite struct {
	suite.Suite
}

func (s *AuthorizationSuite) SetupSuite() {
}

func (s *AuthorizationSuite) TearDownSuite() {
}

func (s *AuthorizationSuite) SetupTest() {
}

func TestAuthorizationSuite(t *testing.T) {
	suite.Run(t, new(AuthorizationSuite))
}

func (s *AuthorizationSuite) TestConstructor() {
	expectedHost := "test.com"
	expectedIgnorePaths := []string{"/*"}
	expectedIgnoreGraphqlOperations := []string{"myOperation"}
	hosts := map[string]pkg.HostSettings{
		expectedHost: {
			IgnorePaths:             expectedIgnorePaths,
			IgnoreGraphqlOperations: expectedIgnoreGraphqlOperations,
		},
	}
	authUtils, err := pkg.NewAuthorizationUtils(hosts)
	require.NoError(s.T(), err)
	require.Len(s.T(), authUtils.Hosts, 1)
	hostSettings := authUtils.Hosts["test.com"]
	require.NotNil(s.T(), hostSettings)
	require.Len(s.T(), hostSettings.IgnoreUrlPaths, 1)
	require.Len(s.T(), hostSettings.IgnoreUrlPaths[0].Segments, 1)
	require.Equal(s.T(), hostSettings.IgnoreGraphqlOperationsSet.Size(), 1)
	require.True(s.T(), hostSettings.IgnoreGraphqlOperationsSet.Contains(expectedIgnoreGraphqlOperations[0]))
}
func (s *AuthorizationSuite) TestShouldIgnoreRequestNoHostMatched() {
	// set configuration
	settingsJson := []byte(`{"hostSettings": {"test.com": {"ignorePaths": ["/some/path/here/*"]}}}`)
	var authUtils pkg.AuthorizationUtils
	err := json.Unmarshal(settingsJson, &authUtils)
	require.NoError(s.T(), err)
	// create request with path that would match the ignorePaths, but on a different host
	req := &v3.CheckRequest{
		Attributes: &v3.AttributeContext{
			Request: &v3.AttributeContext_Request{
				Http: &v3.AttributeContext_HttpRequest{
					Host: "trains.test.com",
					Path: "/some/path/here/do/some/stuff",
				},
			},
		},
	}
	require.False(s.T(), authUtils.ShouldIgnoreRequest(req))
}

func (s *AuthorizationSuite) TestShouldIgnoreRequestHostMatchedIgnorePathsMatched() {
	// set configuration
	settingsJson := []byte(`{"hostSettings": {"test.com": {"ignorePaths": ["/some/path/here/*"]}}}`)
	var authUtils pkg.AuthorizationUtils
	err := json.Unmarshal(settingsJson, &authUtils)
	require.NoError(s.T(), err)
	// create request that should be skipped based on matching ignore paths
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
	require.True(s.T(), authUtils.ShouldIgnoreRequest(req))
}

func (s *AuthorizationSuite) TestShouldIgnoreRequestHostMatchedIgnorePathsNotMatched() {
	// set configuration
	settingsJson := []byte(`{"hostSettings": {"test.com": {"ignorePaths": ["/some/path/here/*"]}}}`)
	var authUtils pkg.AuthorizationUtils
	err := json.Unmarshal(settingsJson, &authUtils)
	require.NoError(s.T(), err)
	// create request that should be skipped based on matching ignore paths
	req := &v3.CheckRequest{
		Attributes: &v3.AttributeContext{
			Request: &v3.AttributeContext_Request{
				Http: &v3.AttributeContext_HttpRequest{
					Host: "test.com",
					Path: "/oh/noes/you/must/auth",
				},
			},
		},
	}
	require.False(s.T(), authUtils.ShouldIgnoreRequest(req))
}

func (s *AuthorizationSuite) TestShouldIgnoreRequestHostMatchedGraphqlMatchedStringBody() {
	// set configuration
	settingsJson := []byte(`{"hostSettings": {"test.com": {"ignorePaths": ["/some/path/here/*"], "ignoreGraphqlOperations": ["doThing"]}}}`)
	body := "{\n  \"operationName\": \"DoThing\",\n  \"variables\": {},\n  \"query\": \"query DoThing {\\n  doThing {\\n    result {\\n      name\\n      place\\n}\\n}\\n}\"\n}"
	var authUtils pkg.AuthorizationUtils
	err := json.Unmarshal(settingsJson, &authUtils)
	require.NoError(s.T(), err)
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
	require.True(s.T(), authUtils.ShouldIgnoreRequest(req))
}

func (s *AuthorizationSuite) TestShouldIgnoreRequestHostMatchedGraphqlMatchedBytesBody() {
	// set configuration
	settingsJson := []byte(`{"hostSettings": {"test.com": {"ignorePaths": ["/some/path/here/*"], "ignoreGraphqlOperations": ["doThing"]}}}`)
	body := "{\n  \"operationName\": \"DoThing\",\n  \"variables\": {},\n  \"query\": \"query DoThing {\\n  doThing {\\n    result {\\n      name\\n      place\\n}\\n}\\n}\"\n}"
	var authUtils pkg.AuthorizationUtils
	err := json.Unmarshal(settingsJson, &authUtils)
	require.NoError(s.T(), err)
	// create request that should be skipped based on matching ignore paths
	req := &v3.CheckRequest{
		Attributes: &v3.AttributeContext{
			Request: &v3.AttributeContext_Request{
				Http: &v3.AttributeContext_HttpRequest{
					Host:    "test.com",
					Path:    "/path/not/matched",
					RawBody: []byte(body),
				},
			},
		},
	}
	require.True(s.T(), authUtils.ShouldIgnoreRequest(req))
}

func (s *AuthorizationSuite) TestShouldIgnoreRequestHostMatchedGraphqlNotMatchedStringBody() {
	// set configuration
	settingsJson := []byte(`{"hostSettings": {"test.com": {"ignorePaths": ["/some/path/here/*"], "ignoreGraphqlOperations": ["doAnotherThing"]}}}`)
	body := "{\n  \"operationName\": \"DoThing\",\n  \"variables\": {},\n  \"query\": \"query DoThing {\\n  doThing {\\n    result {\\n      name\\n      place\\n}\\n}\\n}\"\n}"
	var authUtils pkg.AuthorizationUtils
	err := json.Unmarshal(settingsJson, &authUtils)
	require.NoError(s.T(), err)
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
	require.False(s.T(), authUtils.ShouldIgnoreRequest(req))
}

func (s *AuthorizationSuite) TestShouldIgnoreRequestHostMatchedGraphqlNotMatchedBytesBody() {
	// set configuration
	settingsJson := []byte(`{"hostSettings": {"test.com": {"ignorePaths": ["/some/path/here/*"], "ignoreGraphqlOperations": ["doAnotherThing"]}}}`)
	body := "{\n  \"operationName\": \"DoThing\",\n  \"variables\": {},\n  \"query\": \"query DoThing {\\n  doThing {\\n    result {\\n      name\\n      place\\n}\\n}\\n}\"\n}"
	var authUtils pkg.AuthorizationUtils
	err := json.Unmarshal(settingsJson, &authUtils)
	require.NoError(s.T(), err)
	// create request that should be skipped based on matching ignore paths
	req := &v3.CheckRequest{
		Attributes: &v3.AttributeContext{
			Request: &v3.AttributeContext_Request{
				Http: &v3.AttributeContext_HttpRequest{
					Host:    "test.com",
					Path:    "/path/not/matched",
					RawBody: []byte(body),
				},
			},
		},
	}
	require.False(s.T(), authUtils.ShouldIgnoreRequest(req))
}
