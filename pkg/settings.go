package internal

import (
	"encoding/json"
	"github.com/emirpasic/gods/sets/hashset"
	v3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/tidwall/gjson"
	"github.com/ucarion/urlpath"
	"strings"
)

type AuthorizationUtils struct {
	Hosts map[string]HostSettings `json:"hostSettings"`
}

type HostSettings struct {
	IgnorePathStrings       []string       `json:"IgnorePaths"`
	IgnorePaths             []urlpath.Path `json:"-"`
	IgnoreGraphqlOperations *hashset.Set   `json:"ignoreGraphqlOperations"`
}

func (a *AuthorizationUtils) ShouldIgnoreRequest(request *v3.CheckRequest) bool {
	requestHost := request.Attributes.Request.Http.Host
	settings, ok := a.Hosts[requestHost]
	if !ok {
		// no exceptions configured for this host, return false
		return false
	}
	// request host matches ignore configurations, check the request
	return settings.shouldIgnoreRequest(request)
}

func (h *HostSettings) shouldIgnoreRequest(request *v3.CheckRequest) bool {
	requestPath := getPathFromRequest(request)
	// check paths first
	for _, path := range h.IgnorePaths {
		_, match := path.Match(requestPath)
		if match {
			// path matches an ignored path, return true
			return true
		}
	}
	// path is not ignored, check graphql operations
	body := getBodyFromRequest(request)
	operationName := getGraphqlOperationFromBody(body)
	if operationName == "" {
		// not a graphql request so can't match against graphql queries, return false
		return false
	}
	// if the host's configuration contains the operation, return true to ignore the request
	return h.IgnoreGraphqlOperations.Contains(operationName)
}

func getGraphqlOperationFromBody(body string) string {
	// get the query
	query := gjson.Get(body, "query").String()
	if query == "" {
		return ""
	}
	// strip newlines
	query = removeNewlines(query)
	firstBracket := strings.Index(query, "{")
	if firstBracket == -1 {
		return ""
	}
	secondBracket := strings.Index(query[firstBracket+1:], "{")
	if secondBracket == -1 {
		return ""
	}
	from := firstBracket + 1
	to := firstBracket + secondBracket
	segment := query[from:to]
	firstParen := strings.Index(segment, "(")
	if firstParen > -1 {
		segment = segment[0:firstParen]
	}
	return removeSpaces(segment)
}

func removeNewlines(theString string) string {
	return strings.Replace(theString, "\n", "", -1)
}

func removeSpaces(theString string) string {
	return strings.Replace(theString, " ", "", -1)
}

func getPathFromRequest(request *v3.CheckRequest) string {
	return request.Attributes.Request.Http.Path
}

func getBodyFromRequest(request *v3.CheckRequest) string {
	// attempt to return body
	body := request.Attributes.Request.Http.Body
	if body != "" {
		return body
	}

	// attempt to return raw body a string
	rawBody := request.Attributes.Request.Http.RawBody
	if len(rawBody) > 0 {
		return string(rawBody)
	}

	// neither are present, return empty string
	return ""
}

func (h *HostSettings) UnmarshalJSON(data []byte) error {
	type Alias HostSettings
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(h),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if len(h.IgnorePathStrings) > 0 {
		h.IgnorePaths = []urlpath.Path{}
		for _, pathString := range h.IgnorePathStrings {
			h.IgnorePaths = append(h.IgnorePaths, urlpath.New(pathString))
		}
	}
	return nil
}
