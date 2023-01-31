package fmhttp

import (
	"github.com/go-farmyard/farmyard/fmlog"
	"github.com/go-farmyard/farmyard/fmutil"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"time"
)

type AnyContext any

type Context struct {
	Request *http.Request

	WrappedContext     AnyContext
	wrappedContextType reflect.Type

	HttpServer     *HttpServer
	ResponseWriter *ResponseWriterWrapper

	session     *Session
	queryValues url.Values
	pathParams  []string

	HandlerData map[string]any
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.Request.Context().Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.Request.Context().Done()
}

func (c *Context) Err() error {
	return c.Request.Context().Err()
}

func (c *Context) Value(key any) any {
	return c.Request.Context().Value(key)
}

var typeRequestPtr = reflect.TypeOf(&Context{})

func (c *Context) IsMethodGet() bool {
	return c.Request.Method == "GET"
}

func (c *Context) IsMethodPost() bool {
	return c.Request.Method == "POST"
}

func (c *Context) Session() *Session {
	if c.session == nil {
		if c.HttpServer.sessionStore != nil {
			// the FilesystemStore may return error if the cookie file doesn't exist. And the session is always returned.
			lowSession, _ := c.HttpServer.sessionStore.Get(c.Request, c.HttpServer.sessionCookieName)
			c.session = NewSession(lowSession)
		}
	}
	return c.session
}

func (c *Context) Respond(vars ...any) *ResponseCommon {
	wr := &ResponseCommon{request: c}
	for _, v0 := range vars {
		switch v := v0.(type) {
		case int:
			wr.statusCode = v
		case RedirectSeeOther, RedirectTemporary, RedirectPermanent:
			wr.redirect = v
		default:
			fmutil.MustTrue(wr.respBody == nil, "response body was already set")
			wr.respBody = v
		}
	}
	return wr
}

func (c *Context) RespondRedirectSameSite(location string, params ...fmutil.Map) *ResponseCommon {
	isValid := false
	if location == "" || location[0] == '?' || location[0] == '.' {
		isValid = true
	} else if location[0] == '/' {
		if len(location) == 1 {
			isValid = true
		} else {
			isValid = location[1] != '/' && location[1] != '\\'
		}
	} else {
		u, err := url.Parse(location)
		if err == nil && u.Host == c.Request.Host {
			isValid = true
			u.Scheme = ""
			u.Opaque = ""
			location = u.RequestURI()
		}
	}
	if !isValid {
		fmlog.Debugf("location %s is not the same site as %s, redirection is denied", location, c.UriHost())
		return c.Respond(RedirectSeeOther("/"))
	}
	uri := fmutil.BuildUrl("", location, fmutil.DefZero(params))
	return c.Respond(RedirectSeeOther(uri))
}

func (c *Context) RespondJson(v any) *ResponseCommon {
	resp := c.Respond(fmutil.AsJsonDataProvider(v))
	resp.Header().Add("Content-Type", "application/json")
	return resp
}

func (c *Context) RespondTmpl(name string, data map[string]any) *ResponseCommon {
	return c.Respond(responderTmpl{
		req:  c,
		name: name,
		data: data,
	})
}

func (c *Context) RespondPathTmpl(data map[string]any) *ResponseCommon {
	return c.RespondTmpl(filepath.Clean(c.Request.URL.Path[1:]+".tmpl"), data)
}

func (c *Context) RespondFile(name string, file fs.File) *ResponseCommon {
	return c.Respond(responderFile{
		req:  c,
		name: name,
		file: file,
	})
}

func (c *Context) IsResponseWritten() bool {
	return c.ResponseWriter.statusCode != 0 || c.ResponseWriter.written != 0
}

func (c *Context) UriHost() string {
	schema := c.Request.Header.Get("X-Forwarded-Proto")
	schema = fmutil.IfZero(schema, "http")
	if c.Request.Header.Get("HTTPS") == "on" {
		schema = "https"
	}
	return schema + "://" + c.Request.Host
}

func (c *Context) RealRemoteIp() string {
	var headerRealIp string
	if c.HttpServer.realIpHeader != "" {
		headerRealIp = c.Request.Header.Get(c.HttpServer.realIpHeader)
	}

	remoteHost, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	fmutil.MustNoError(err)

	remoteIp := net.ParseIP(remoteHost)
	fmutil.MustTrue(remoteIp != nil, "remote addr must contain valid IP")

	remoteIpTrusted := false
	for _, ipNet := range c.HttpServer.trustHttpHeaderFrom {
		if ipNet.Contains(remoteIp) {
			remoteIpTrusted = true
			break
		}
	}
	if remoteIpTrusted && headerRealIp != "" {
		return headerRealIp
	} else {
		return remoteHost
	}
}
