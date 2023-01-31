package fmhttp

import (
	"io"
	"io/fs"
	"net/http"
	"os"
)

type ResponseWriterWrapper struct {
	responseWriter http.ResponseWriter
	statusCode     int
	written        int64
}

func (w *ResponseWriterWrapper) Header() http.Header {
	return w.responseWriter.Header()
}

func (w *ResponseWriterWrapper) Write(bytes []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	n, err := w.responseWriter.Write(bytes)
	w.written += int64(n)
	return n, err
}

func (w *ResponseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.responseWriter.WriteHeader(statusCode)
}

type responderTmpl struct {
	req  *Context
	name string
	data map[string]any
}

func (r responderTmpl) WriteTo(_ io.Writer) (n int64, err error) {
	var tmplData map[string]any
	if len(r.data) == 0 {
		tmplData = r.req.HandlerData
	} else if len(r.req.HandlerData) == 0 {
		tmplData = r.data
	} else {
		tmplData = map[string]any{}
		for k, v := range r.req.HandlerData {
			tmplData[k] = v
		}
		for k, v := range r.data {
			tmplData[k] = v
		}
	}

	respWriter := r.req.ResponseWriter
	err = r.req.HttpServer.TmplRender(respWriter, r.name, tmplData)
	if err != nil {
		if respWriter.written == 0 {
			if os.IsNotExist(err) {
				http.NotFound(respWriter, r.req.Request)
			} else {
				http.Error(respWriter, "internal error (render template error)", http.StatusInternalServerError)
			}
		}
		return respWriter.written, err
	}
	return respWriter.written, nil
}

type responderFile struct {
	req  *Context
	name string
	file fs.File
}

func (r responderFile) WriteTo(_ io.Writer) (n int64, err error) {
	respWriter := r.req.ResponseWriter
	st, err := r.file.Stat()
	if err != nil {
		http.Error(respWriter, "internal error (file stat error)", http.StatusInternalServerError)
		return respWriter.written, err
	}
	http.ServeContent(respWriter, r.req.Request, r.name, st.ModTime(), r.file.(io.ReadSeeker))
	return respWriter.written, nil
}
