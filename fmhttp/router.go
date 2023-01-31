package fmhttp

import (
	"github.com/go-farmyard/farmyard/fmutil"
	"net/http"
	"strings"
)

type httpMethodType uint

const (
	mAny httpMethodType = iota
	mCONNECT
	mDELETE
	mGET
	mHEAD
	mOPTIONS
	mPATCH
	mPOST
	mPUT
	mTRACE
)

var httpMethodTypeMap = map[string]httpMethodType{
	http.MethodConnect: mCONNECT,
	http.MethodDelete:  mDELETE,
	http.MethodGet:     mGET,
	http.MethodHead:    mHEAD,
	http.MethodOptions: mOPTIONS,
	http.MethodPatch:   mPATCH,
	http.MethodPost:    mPOST,
	http.MethodPut:     mPUT,
	http.MethodTrace:   mTRACE,
}

/*
var httpTypeMethodMap = map[httpMethodType]string{
	mCONNECT: http.MethodConnect,
	mDELETE:  http.MethodDelete,
	mGET:     http.MethodGet,
	mHEAD:    http.MethodHead,
	mOPTIONS: http.MethodOptions,
	mPATCH:   http.MethodPatch,
	mPOST:    http.MethodPost,
	mPUT:     http.MethodPut,
	mTRACE:   http.MethodTrace,
}
*/

type Router interface {
	RequestHandler

	// Use appends one or more middlewares onto the Router stack.
	Use(middlewares ...AnyHandler)

	With(h ...AnyHandler) *WithDelegator

	// Group adds a new inline-Router along the current routing path, with a fresh middleware stack for the inline-Router.
	Group(fn func(r Router)) Router

	// Route mounts a sub-Router along a `pattern`` string.
	Route(pattern string, fn func(r Router)) Router

	Any(pattern string, h ...AnyHandler)
	Method(method, pattern string, h ...AnyHandler)

	Connect(pattern string, h ...AnyHandler)
	Delete(pattern string, h ...AnyHandler)
	Get(pattern string, h ...AnyHandler)
	Head(pattern string, h ...AnyHandler)
	Options(pattern string, h ...AnyHandler)
	Patch(pattern string, h ...AnyHandler)
	Post(pattern string, h ...AnyHandler)
	Put(pattern string, h ...AnyHandler)
	Trace(pattern string, h ...AnyHandler)

	Pattern(pattern string, h ...AnyHandler) *PatternDelegator

	// NotFound defines a handler to respond whenever a route could not be found.
	NotFound(h ...AnyHandler)

	// MethodNotAllowed defines a handler to respond whenever a method is not allowed.
	MethodNotAllowed(h ...AnyHandler)
}

type routerMethodChainMap map[httpMethodType]*handlerChain

type PatternDelegator struct {
	router      Router
	pattern     string
	middlewares []AnyHandler
}

type routerParamDetail struct {
	field string
	parts []string
}

type WithDelegator struct {
	r           *routerImpl
	middlewares []AnyHandler
}

func (w WithDelegator) Group(fn func(r Router)) Router {
	return w.r.Group(func(r Router) {
		r.Use(w.middlewares...)
		fn(r)
	})
}

const pathPatternWildcardField = "**"

type routerImpl struct {
	parent *routerImpl

	handlerMethodMap routerMethodChainMap
	fixedFields      map[string]Router
	paramFields      map[string]Router
	paramParts       map[string]routerParamDetail

	// scoped, isolated from groups
	chain                 *handlerChain
	chainNotFound         *handlerChain
	chainMethodNotAllowed *handlerChain
}

func NewRouter() Router {
	return &routerImpl{
		chain: &handlerChain{},
	}
}

func (r *routerImpl) Use(middlewares ...AnyHandler) {
	r.chain.addMiddleware(middlewares...)
}

func (r *routerImpl) With(h ...AnyHandler) *WithDelegator {
	return &WithDelegator{r: r, middlewares: h}
}

func (r *routerImpl) Group(fn func(r Router)) Router {
	oldChain := r.chain.clone()
	oldNotFound := r.chainNotFound
	oldMethodNotAllowed := r.chainMethodNotAllowed
	fn(r)
	r.chain = oldChain
	r.chainNotFound = oldNotFound
	r.chainMethodNotAllowed = oldMethodNotAllowed
	return r
}

func (r *routerImpl) Route(pattern string, fn func(r Router)) Router {
	if len(pattern) > 0 && pattern[0] == '/' {
		pattern = pattern[1:]
	}
	r1 := r
	if pattern != "" {
		fields := strings.Split(pattern, "/")
		for _, field := range fields {
			r1 = r1.prepareSubRouter(field)
		}
	}
	fn(r1)
	return r1
}

func (r *routerImpl) newSubRouter() *routerImpl {
	return &routerImpl{
		parent:                r,
		chain:                 r.chain.clone(),
		chainNotFound:         r.chainNotFound,
		chainMethodNotAllowed: r.chainMethodNotAllowed,
	}
}

func (r *routerImpl) prepareFixedSubRouter(field string) *routerImpl {
	subRouter, ok := r.fixedFields[field].(*routerImpl)
	if !ok {
		subRouter = r.newSubRouter()
	}
	if r.fixedFields == nil {
		r.fixedFields = map[string]Router{}
	}
	r.fixedFields[field] = subRouter
	return subRouter
}

func splitRouteParamField(field string) (parts []string) {
	p := 0
	for {
		p1 := strings.IndexByte(field[p:], '{')
		p2 := strings.IndexByte(field[p:], '}')
		if p1 == -1 && p2 == -1 {
			parts = append(parts, field[p:])
			return
		}
		if p1 != -1 && p2 != -1 {
			parts = append(parts, field[p:p+p1])
			parts = append(parts, field[p+p1+1:p+p2])
			p += p2 + 1
		} else {
			fmutil.Panic("invalid field: %s", field)
		}
	}
}

func parseRouteParamField(field string, parts []string, result *[]string) bool {
	if !strings.HasPrefix(field, parts[0]) {
		return false
	}
	oldLen := len(*result)
	field = field[len(parts[0]):]
	for i := 1; i < len(parts); i += 2 {
		fixed := parts[i+1]
		pos := strings.Index(field, fixed)
		if pos == -1 {
			*result = (*result)[:oldLen]
			return false
		}
		if i == len(parts)-2 && fixed == "" {
			pos = len(field)
		}
		*result = append(*result, parts[i])
		*result = append(*result, field[:pos])
		field = field[pos+len(fixed):]
	}
	return true
}

func (r *routerImpl) prepareParamSubRouter(field string) *routerImpl {
	subRouter, ok := r.paramFields[field].(*routerImpl)
	if !ok {
		subRouter = r.newSubRouter()
	}
	if r.paramFields == nil {
		r.paramFields = map[string]Router{}
		r.paramParts = map[string]routerParamDetail{}
	}
	r.paramFields[field] = subRouter
	r.paramParts[field] = routerParamDetail{
		field: field,
		parts: splitRouteParamField(field),
	}
	return subRouter
}

func (r *routerImpl) prepareSubRouter(field string) *routerImpl {
	if strings.Contains(field, "{") {
		return r.prepareParamSubRouter(field)
	} else {
		return r.prepareFixedSubRouter(field)
	}
}

func (r *routerImpl) addMethodHandler(m httpMethodType, h []AnyHandler) {
	if r.handlerMethodMap == nil {
		r.handlerMethodMap = routerMethodChainMap{}
	}
	r.handlerMethodMap[m] = r.chain.clone().addEndpoint(h...)
}

func (r *routerImpl) handlePattern(m httpMethodType, pattern string, h []AnyHandler) {
	if len(pattern) > 0 && pattern[0] == '/' {
		pattern = pattern[1:]
	}
	posSep := strings.Index(pattern, "/")
	if posSep == -1 {
		field := pattern
		subRouter := r.prepareSubRouter(field)
		subRouter.addMethodHandler(m, h)
		return
	} else {
		field := pattern[:posSep]
		extra := pattern[posSep+1:]
		fmutil.MustTrue(field != pathPatternWildcardField, "path wildcard must be the last field")
		subRouter := r.prepareSubRouter(field)
		subRouter.handlePattern(m, extra, h)
	}
}

func (r *routerImpl) NotFound(h ...AnyHandler) {
	r.chainNotFound = r.chain.clone().addEndpoint(h...)
}

func (r *routerImpl) MethodNotAllowed(h ...AnyHandler) {
	r.chainMethodNotAllowed = r.chain.clone().addEndpoint(h...)
}

func (r *routerImpl) Any(pattern string, h ...AnyHandler) {
	r.handlePattern(mAny, pattern, h)
}

func (r *routerImpl) handleMethod(method, pattern string, h []AnyHandler) {
	if m, ok := httpMethodTypeMap[method]; ok {
		r.handlePattern(m, pattern, h)
	} else {
		fmutil.Panic("unknown http method: %s for %s", method, pattern)
	}
}

func (r *routerImpl) Method(method, pattern string, h ...AnyHandler) {
	r.handleMethod(method, pattern, h)
}

func (r *routerImpl) Connect(pattern string, h ...AnyHandler) {
	r.handleMethod(http.MethodConnect, pattern, h)
}

func (r *routerImpl) Delete(pattern string, h ...AnyHandler) {
	r.handleMethod(http.MethodDelete, pattern, h)
}

func (r *routerImpl) Get(pattern string, h ...AnyHandler) {
	r.handleMethod(http.MethodGet, pattern, h)
}

func (r *routerImpl) Head(pattern string, h ...AnyHandler) {
	r.handleMethod(http.MethodHead, pattern, h)
}

func (r *routerImpl) Options(pattern string, h ...AnyHandler) {
	r.handleMethod(http.MethodOptions, pattern, h)
}

func (r *routerImpl) Patch(pattern string, h ...AnyHandler) {
	r.handleMethod(http.MethodPatch, pattern, h)
}

func (r *routerImpl) Post(pattern string, h ...AnyHandler) {
	r.handleMethod(http.MethodPost, pattern, h)
}

func (r *routerImpl) Put(pattern string, h ...AnyHandler) {
	r.handleMethod(http.MethodPut, pattern, h)
}

func (r *routerImpl) Trace(pattern string, h ...AnyHandler) {
	r.handleMethod(http.MethodTrace, pattern, h)
}

func (r *routerImpl) Pattern(pattern string, h ...AnyHandler) *PatternDelegator {
	return &PatternDelegator{
		router:      r,
		pattern:     pattern,
		middlewares: h,
	}
}

func (pd *PatternDelegator) Delete(h ...AnyHandler) *PatternDelegator {
	pd.router.Delete(pd.pattern, append(pd.middlewares, h...)...)
	return pd
}

func (pd *PatternDelegator) Get(h ...AnyHandler) *PatternDelegator {
	pd.router.Get(pd.pattern, append(pd.middlewares, h...)...)
	return pd
}

func (pd *PatternDelegator) Head(h ...AnyHandler) *PatternDelegator {
	pd.router.Head(pd.pattern, append(pd.middlewares, h...)...)
	return pd
}

func (pd *PatternDelegator) Options(h ...AnyHandler) *PatternDelegator {
	pd.router.Options(pd.pattern, append(pd.middlewares, h...)...)
	return pd
}

func (pd *PatternDelegator) Patch(h ...AnyHandler) *PatternDelegator {
	pd.router.Patch(pd.pattern, append(pd.middlewares, h...)...)
	return pd
}

func (pd *PatternDelegator) Post(h ...AnyHandler) *PatternDelegator {
	pd.router.Post(pd.pattern, append(pd.middlewares, h...)...)
	return pd
}

func (pd *PatternDelegator) Put(h ...AnyHandler) *PatternDelegator {
	pd.router.Put(pd.pattern, append(pd.middlewares, h...)...)
	return pd
}

func (r *routerImpl) matchPatternField(c *Context, field, extra string) (subRouter *routerImpl, found bool) {
	subRouter, ok := r.fixedFields[field].(*routerImpl)
	if ok {
		return subRouter, true
	}

	for _, partDetail := range r.paramParts {
		if parseRouteParamField(field, partDetail.parts, &c.pathParams) {
			return r.paramFields[partDetail.field].(*routerImpl), true
		}
	}
	return nil, false
}

func (r *routerImpl) matchHandlerMethod(c *Context) (hc *handlerChain, found bool) {
	method, ok := httpMethodTypeMap[c.Request.Method]
	if !ok {
		fmutil.Panic("unknown method: %s", c.Request.Method)
	}
	hc, found = r.handlerMethodMap[method]
	if !found {
		hc, found = r.handlerMethodMap[mAny]
	}
	if !found {
		hc = r.chainMethodNotAllowed
	}
	return hc, found
}

func (r *routerImpl) matchHandlerChain(c *Context, path string) (hc *handlerChain, found bool) {
	fields := strings.SplitN(path, "/", 2)
	field := fields[0]
	extra := ""

	matchEndpoint := false
	if len(fields) == 1 {
		matchEndpoint = true
	} else {
		extra = fields[1]
		if extra == "" {
			// to support access "/foo/" with routes: {"/foo": ...}
			matchEndpoint = true
		}
	}

	pathParamsOldLen := len(c.pathParams)
	defer func() {
		if !found {
			var subRouter *routerImpl
			subRouter, found = r.fixedFields[pathPatternWildcardField].(*routerImpl)
			if found {
				c.pathParams = append(c.pathParams, pathPatternWildcardField, path)
				hc, found = subRouter.matchHandlerMethod(c)
			}
		}
		if c.pathParams != nil && !found {
			c.pathParams = c.pathParams[:pathParamsOldLen]
		}
	}()

	var subRouter *routerImpl
	if matchEndpoint {
		subRouter, found = r.matchPatternField(c, field, "")
		if !found {
			return r.chainNotFound, false
		}

		hc, found = subRouter.matchHandlerMethod(c)
		if !found && path != "" {
			// to support access "/foo" with routes: {"/foo":{"":...}}
			hc, found = subRouter.matchHandlerChain(c, "")
		}
		if !found && hc == nil {
			hc = r.chainMethodNotAllowed
		}
		return hc, found
	} else {
		subRouter, found = r.matchPatternField(c, field, extra)
		if !found {
			return r.chainNotFound, false
		}
		return subRouter.matchHandlerChain(c, extra)
	}
}

func (r *routerImpl) Handle(c *Context) Response {
	path := c.Request.URL.EscapedPath()
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	hc, _ := r.matchHandlerChain(c, path)
	if hc == nil {
		return c.Respond(404, "not found: "+c.Request.RequestURI)
	}
	return hc.Handle(c)
}
