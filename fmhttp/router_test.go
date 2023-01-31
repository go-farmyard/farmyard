package fmhttp

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestSplitRouteParamField(t *testing.T) {

	tests := []struct {
		name  string
		parts []string
	}{
		{
			name:  "{user}",
			parts: []string{"", "user", ""},
		},
		{
			name:  "a-{x}-b",
			parts: []string{"a-", "x", "-b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := splitRouteParamField(tt.name)
			assert.EqualValues(t, tt.parts, parts)
		})
	}
}

func TestParseRouteParamField(t *testing.T) {

	tests := []struct {
		name   string
		field  string
		params []string
	}{
		{
			name:   "{user}",
			field:  "abc",
			params: []string{"user", "abc"},
		},
		{
			name:   "a-{user}",
			field:  "a-xyz",
			params: []string{"user", "xyz"},
		},
		{
			name:   "a-{user}",
			field:  "xyz",
			params: nil,
		},
		{
			name:   "{user}-b",
			field:  "xyz",
			params: nil,
		},
		{
			name:   "a-{user}-b-{id}-c",
			field:  "a-u-b-1-c",
			params: []string{"user", "u", "id", "1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var params []string
			parts := splitRouteParamField(tt.name)
			ret := parseRouteParamField(tt.field, parts, &params)
			assert.EqualValues(t, ret, tt.params != nil)
			assert.EqualValues(t, tt.params, params)
		})
	}
}

func testResp(v int) RequestHandlerFunc {
	return func(c *Context) Response {
		return c.Respond(v)
	}
}

func testMid(v int) func(he *ChainExecutor) Response {
	return func(he *ChainExecutor) Response {
		resp := he.Next()
		resp.(*ResponseCommon).statusCode += v
		return resp
	}
}

func testCtx(req string) *Context {
	fields := strings.SplitN(req, " ", 2)
	method := fields[0]
	path := fields[1]
	u := &url.URL{Path: path}
	return &Context{
		Request: &http.Request{
			Method: method,
			URL:    u,
		},
	}
}

func TestRouter(t *testing.T) {
	r := NewRouter()
	r.Route("/", func(r Router) {
		r.Get("", testResp(200))
		r.Get("t201", testResp(201))
		r.Route("m1", func(r Router) {
			r.NotFound(testResp(444))
			r.MethodNotAllowed(testResp(445))
			r.Get("", testResp(200))
			r.Use(testMid(10))
			r.Get("t", testResp(201))
		})
		r.Route("m2", func(r Router) {
			r.Use(testMid(10))
			r.Use(testMid(25))
			r.Get("t", testResp(201))
		})
		r.Group(func(r Router) {
			r.Use(testMid(30))
			r.Get("m3", testResp(201))
		})
		r.With(testMid(10)).Group(func(r Router) {
			r.Use(testMid(30))
			r.Get("m4", testResp(201))
		})
		r.Any("/sub/t209", testMid(7), testResp(202))
		r.Pattern("pattern").Put(testResp(201)).Patch(testResp(202)).Delete(testResp(203))
		r.Get("/wildcard/**", testResp(210))
	})

	tests := []struct {
		name     string
		expected int
	}{
		{
			name:     "GET /",
			expected: 200,
		},
		{
			name:     "POST /",
			expected: 404,
		},
		{
			name:     "GET /no-such",
			expected: 404,
		},
		{
			name:     "GET /no-such/no-such",
			expected: 404,
		},
		{
			name:     "GET /t201",
			expected: 201,
		},
		{
			name:     "GET /t201/",
			expected: 201,
		},
		{
			name:     "GET /m1",
			expected: 200,
		},
		{
			name:     "GET /m1/",
			expected: 200,
		},
		{
			name:     "GET /m1/no-such",
			expected: 444,
		},
		{
			name:     "POST /m1/",
			expected: 445,
		},
		{
			name:     "GET /m1/t",
			expected: 211,
		},
		{
			name:     "GET /m2/t",
			expected: 236,
		},
		{
			name:     "GET /m3",
			expected: 231,
		},
		{
			name:     "GET /m4",
			expected: 241,
		},
		{
			name:     "POST /sub/t209",
			expected: 209,
		},
		{
			name:     "PUT /pattern",
			expected: 201,
		},
		{
			name:     "PATCH /pattern",
			expected: 202,
		},
		{
			name:     "DELETE /pattern",
			expected: 203,
		},
		{
			name:     "TRACE /pattern",
			expected: 404,
		},
		{
			name:     "GET /wildcard",
			expected: 210,
		},
		{
			name:     "GET /wildcard/test",
			expected: 210,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := r.Handle(testCtx(tt.name))
			assert.EqualValues(t, tt.expected, resp.StatusCode(), tt.name)
		})
	}
}

func TestRouterPathParams(t *testing.T) {
	r := NewRouter()
	r.Route("/", func(r Router) {
		r.Route("{a}", func(r Router) {
			r.Get("x-{b}", testResp(201))
			r.Get("x-{b}/**", testResp(211))
		})
		r.Get("foo", testResp(202))
	})

	tests := []struct {
		name       string
		statusCode int
		pathParams []string
	}{
		{
			name:       "GET /1/x-2",
			statusCode: 201,
			pathParams: []string{"a", "1", "b", "2"},
		},
		{
			name:       "GET /1/x-2/any/more",
			statusCode: 211,
			pathParams: []string{"a", "1", "b", "2", "**", "any/more"},
		},
		{
			name:       "GET /1",
			statusCode: 404,
			pathParams: []string{},
		},
		{
			name:       "GET /foo",
			statusCode: 202,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testCtx(tt.name)
			resp := r.Handle(ctx)
			assert.EqualValues(t, tt.statusCode, resp.StatusCode(), tt.name)
			assert.EqualValues(t, tt.pathParams, ctx.pathParams, tt.name)
		})
	}
}
