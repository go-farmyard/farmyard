package fmhttp

import (
	"context"
	"github.com/go-farmyard/farmyard/fmlog"
	"github.com/go-farmyard/farmyard/fmutil"
	"github.com/gorilla/sessions"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
)

type contextWrappedType struct{}

var contextWrappedKey contextWrappedType

type contextWrappedStruct = struct {
	value any
}

type HttpServer struct {
	devMode bool

	server    *http.Server
	serverMux *http.ServeMux

	sessionStore         sessions.Store
	sessionCookieName    string
	sessionCookieOptions *sessions.Options

	assetFS    fs.FS
	tmplRender *TemplateRender

	AssetsDir     fs.FS
	AssetsWebRoot fs.FS

	commonMiddlewares handlerChain
	WrapContext       func(r *Context) AnyContext

	realIpHeader        string
	trustHttpHeaderFrom []*net.IPNet
}

type Options struct {
	DevMode  bool
	Listen   string
	AssetsFS fs.FS

	SessionDir             string
	SessionCookieSecureKey string
	SessionCookieName      string
	SessionCookieOptions   *sessions.Options

	RealIpHeader        string
	TrustHttpHeaderFrom []string
}

func NewHttpServer(opt *Options) *HttpServer {
	hs := &HttpServer{
		devMode: opt.DevMode,

		sessionCookieName:    opt.SessionCookieName,
		sessionCookieOptions: opt.SessionCookieOptions,

		assetFS:   opt.AssetsFS,
		serverMux: http.NewServeMux(),

		realIpHeader: opt.RealIpHeader,
	}

	for _, ipCidr := range opt.TrustHttpHeaderFrom {
		ipCidr = strings.TrimSpace(ipCidr)
		if ipCidr == "" {
			continue
		}
		_, ipNet, err := net.ParseCIDR(ipCidr)
		fmutil.MustNoError(err)
		hs.trustHttpHeaderFrom = append(hs.trustHttpHeaderFrom, ipNet)
	}

	hs.server = &http.Server{
		Addr:    opt.Listen,
		Handler: hs.serverMux,
	}

	hs.initAssetsDir()
	if opt.SessionCookieName != "" {
		ss := sessions.NewFilesystemStore(opt.SessionDir, []byte(opt.SessionCookieSecureKey))
		ss.Options = hs.sessionCookieOptions
		hs.sessionStore = ss
	}
	return hs
}

func (hs *HttpServer) initAssetsDir() {
	var err error

	// for static resources
	if hs.devMode {
		hs.AssetsDir = os.DirFS("assets")
	} else {
		hs.AssetsDir, err = fs.Sub(hs.assetFS, "assets")
		if err != nil {
			fmlog.Fatalf("can not open web directory for assets, err:%v", err)
		}
	}

	hs.AssetsWebRoot, err = fs.Sub(hs.AssetsDir, "web")
	if err != nil {
		fmlog.Fatalf("can not open web directory for http server, err:%v", err)
	}

	assetsTmplRoot, err := fs.Sub(hs.AssetsDir, "template")
	if err != nil {
		fmlog.Fatalf("can not open template directory for http server, err:%v", err)
	}

	hs.tmplRender = NewTemplateRender(assetsTmplRoot)
	hs.tmplRender.DevMode = hs.devMode
}

func (hs *HttpServer) TmplRender(w io.Writer, name string, data map[string]any) error {
	return hs.tmplRender.Render(w, name, data)
}

func (hs *HttpServer) UseMiddleware(handlers ...AnyHandler) {
	hs.commonMiddlewares.addMiddleware(handlers...)
}

func (hs *HttpServer) wrapHandlers(handlers ...AnyHandler) http.HandlerFunc {
	h := hs.commonMiddlewares.clone().addEndpoint(handlers...)
	return func(wOrig http.ResponseWriter, rOrig *http.Request) {
		contextWrap := &contextWrappedStruct{}
		w := &ResponseWriterWrapper{responseWriter: wOrig}
		r := rOrig.WithContext(context.WithValue(rOrig.Context(), contextWrappedKey, contextWrap))
		ctx := &Context{
			HttpServer:     hs,
			Request:        r,
			ResponseWriter: w,
		}
		ctx.WrappedContext = hs.WrapContext(ctx)
		ctx.wrappedContextType = reflect.TypeOf(ctx.WrappedContext)
		contextWrap.value = ctx.WrappedContext

		defer func() {
			if err := recover(); err != nil {
				fmlog.Infof("fmhttp: panic request: %s %s, handler err: %v\n%s\n", r.Method, r.RequestURI, err, string(debug.Stack()))
				http.Error(w, "internal error (handler panic)", http.StatusInternalServerError)
			}
		}()

		resp := h.Handle(ctx)

		if ctx.session != nil {
			err := ctx.session.AutoSave(w, r)
			if err != nil && !strings.HasPrefix(err.Error(), "remove ") {
				fmutil.MustNoError(err, "save session failed")
			}
		}

		if resp != nil && resp != responseNop {
			headers := w.Header()
			for k, v := range resp.Header() {
				headers[k] = v
			}
			_, err := resp.RespondTo(w)
			if err != nil {
				fmlog.Infof("ERROR: failed to write response %T, err:%v", resp, err)
			}
		}

		if ctx.IsResponseWritten() {
			fmlog.Infof("fmhttp: completed %03d %s %s", ctx.ResponseWriter.statusCode, r.Method, r.RequestURI)
		} else {
			fmlog.Infof("fmhttp: no response for %s %s", r.Method, r.RequestURI)
		}
	}
}

func (hs *HttpServer) HandleRequest(pattern string, handlers ...AnyHandler) {
	hs.serverMux.Handle(pattern, hs.wrapHandlers(handlers...))
}

func (hs *HttpServer) HandleAssets(pattern string) {
	hs.serverMux.Handle(pattern, http.FileServer(http.FS(hs.AssetsWebRoot)))
}

func (hs *HttpServer) ServeAssetFile(c *Context) Response {
	path := c.Request.URL.Path[1:]
	if f, err := hs.AssetsWebRoot.Open(path); err == nil {
		return c.RespondFile(path, f)
	}
	return c.Respond(404, "no static file: "+c.Request.RequestURI)
}

func (hs *HttpServer) ListenAndServe() error {
	return hs.server.ListenAndServe()
}
