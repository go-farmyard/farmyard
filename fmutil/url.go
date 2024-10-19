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
		} else if ss != sp {
			raw = base + path
		} else /* ss && sp */ {
			raw = base + path[1:]
		}
	}
	if params != nil {
		alreadyHasQueryParams := strings.Contains(raw, "?")
		vals := url.Values{}
		for k, v := range params {
			vals.Add(k, AsString(v))
		}
		raw += Iif(alreadyHasQueryParams, "&", "?") + vals.Encode()
	}
	return raw
}
