package fmhttp

import (
	"encoding/json"
	"github.com/go-farmyard/farmyard/fmutil"
	"net/url"
	"strconv"
)

func defString(defs []string) string {
	if len(defs) > 0 {
		return defs[0]
	}
	return ""
}

func defInt64(defs []int64) int64 {
	if len(defs) > 0 {
		return defs[0]
	}
	return 0
}

const defIntString = "__default__"

func urlValueStringWithDef(values url.Values, key string, defs ...string) string {
	if v, ok := values[key]; ok && len(v) >= 1 {
		return v[0]
	}
	return defString(defs)
}

func urlValueToInt64WithDef(val string, defIntStr string, defs ...int64) int64 {
	if val != defIntStr {
		v, _ := strconv.ParseInt(val, 10, 64)
		return v
	}
	return defInt64(defs)
}

/*
QueryString parameters
*/

func (c *Context) QueryParam(key string, defs ...string) string {
	if c.queryValues == nil {
		c.queryValues = c.Request.URL.Query()
	}
	return urlValueStringWithDef(c.queryValues, key, defs...)
}

func (c *Context) QueryParamInt64(key string, defs ...int64) int64 {
	val := c.QueryParam(key, defIntString)
	return urlValueToInt64WithDef(val, defIntString, defs...)
}

/*
PostForm parameters
*/

func (c *Context) PostParam(key string, defs ...string) string {
	if c.Request.PostForm == nil {
		_ = c.Request.ParseForm()
	}
	return urlValueStringWithDef(c.Request.PostForm, key, defs...)
}

func (c *Context) PostParamInt64(key string, defs ...int64) int64 {
	val := c.PostParam(key, defIntString)
	return urlValueToInt64WithDef(val, defIntString, defs...)
}

/*
QueryString + PostForm parameters
*/

func (c *Context) FormParam(key string, defs ...string) string {
	if c.Request.Form == nil {
		_ = c.Request.ParseForm()
	}
	return urlValueStringWithDef(c.Request.Form, key, defs...)
}

func (c *Context) FormParamInt64(key string, defs ...int64) int64 {
	val := c.FormParam(key, defIntString)
	return urlValueToInt64WithDef(val, defIntString, defs...)
}

/*
PathParam parameters
*/

func (c *Context) PathParam(key string, defs ...string) string {
	for i := 0; i < len(c.pathParams); i += 2 {
		if c.pathParams[i] == key {
			return c.pathParams[i+1]
		}
	}
	return defString(defs)
}

func (c *Context) PathParamInt64(key string, defs ...int64) int64 {
	val := c.PathParam(key, defIntString)
	return urlValueToInt64WithDef(val, defIntString, defs...)
}

func (c *Context) RequestJson() *fmutil.JsonValue {
	jv, _ := fmutil.JsonDecodeReader(c.Request.Body)
	return jv
}

func (c *Context) RequestJsonUnmarshal(v any) error {
	return json.NewDecoder(c.Request.Body).Decode(v)
}
