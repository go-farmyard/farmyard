package fmhttp

import (
	"encoding/json"
	"errors"
	"github.com/go-farmyard/farmyard/fmutil"
	"io"
	"net/http"
)

type Response interface {
	StatusCode() int
	Header() http.Header
	RespondTo(w http.ResponseWriter) (int64, error)
}

type RedirectSeeOther string
type RedirectTemporary string
type RedirectPermanent string

type ResponseCommon struct {
	request    *Context
	statusCode int
	redirect   any
	headers    http.Header
	respBody   any
}

var responseNop = &ResponseCommon{}

func (wr *ResponseCommon) StatusCode() int {
	return wr.statusCode
}

func (wr *ResponseCommon) Header() http.Header {
	if wr.headers == nil {
		wr.headers = http.Header{}
	}
	return wr.headers
}

func (wr *ResponseCommon) SetStatusCode(v int) *ResponseCommon {
	wr.statusCode = v
	return wr
}

const headerContentType = "Content-Type" // canonical header

func (wr *ResponseCommon) respondJson(v any) (int64, error) {
	w := wr.request.ResponseWriter
	if len(w.Header()[headerContentType]) == 0 {
		w.Header().Add(headerContentType, "application/json")
	}
	buf, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(buf)
	return int64(n), err
}

func (wr *ResponseCommon) RespondTo(w http.ResponseWriter) (int64, error) {
	if wr.redirect != nil {
		statusCode := 0
		redirectUrl := ""
		switch s := wr.redirect.(type) {
		case RedirectSeeOther:
			statusCode = http.StatusSeeOther
			redirectUrl = string(s)
		case RedirectTemporary:
			statusCode = http.StatusTemporaryRedirect
			redirectUrl = string(s)
		case RedirectPermanent:
			statusCode = http.StatusPermanentRedirect
			redirectUrl = string(s)
		}
		if statusCode == 0 {
			fmutil.Panic("unknown redirect type: %T", wr.redirect)
		}
		http.Redirect(w, wr.request.Request, redirectUrl, statusCode)
		return 0, nil
	}

	if wr.statusCode != 0 {
		w.WriteHeader(wr.statusCode)
	}

	switch v := wr.respBody.(type) {
	case []byte:
		n, err := w.Write(v)
		return int64(n), err
	case string:
		n, err := w.Write([]byte(v))
		return int64(n), err
	case io.Reader:
		return io.Copy(w, v)
	case io.WriterTo:
		return v.WriteTo(w)
	case map[string]any, map[any]any:
		return wr.respondJson(v)
	case fmutil.JsonDataProvider:
		return wr.respondJson(v.JsonData())
	case error:
		if wr.statusCode == 0 {
			wr.statusCode = http.StatusInternalServerError
			w.WriteHeader(wr.statusCode)
		}
		return wr.respondJson(fmutil.Map{"Error": v.Error()})
	default:
		if v != nil {
			fmutil.Panic("unknown response body: %T", v)
			return 0, errors.New("invalid response body type")
		}
		return 0, nil
	}
}
