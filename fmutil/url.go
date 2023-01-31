package fmutil

import (
	"net/url"
	"strings"
)

func BuildUrl(base, path string, params Map) string {
	var raw string

	if base == "" || path == "" {
		raw = base + path
	} else {
		ss := strings.HasSuffix(base, "/")
		sp := strings.HasPrefix(path, "/")
		if !ss && !sp {
			raw = base + "/" + path
		} else if ss || sp {
			raw = base + path
		} else {
			raw = base + path[1:]
		}
	}
	if params != nil {
		vals := url.Values{}
		for k, v := range params {
			vals.Add(k, AsString(v))
		}
		if strings.Contains(raw, "?") {
			raw += "&"
		} else {
			raw += "?"
		}
		raw += vals.Encode()
	}
	return raw
}
