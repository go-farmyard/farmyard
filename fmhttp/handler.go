package fmhttp

import (
	"github.com/go-farmyard/farmyard/fmutil"
	"net/http"
	"reflect"
)

type AnyHandler any

type RequestHandlerFunc func(*Context) Response

type RequestHandler interface {
	Handle(c *Context) Response
}

var typeMiddleRequestPtr = reflect.TypeOf(&ChainExecutor{})
var typeResponseWriter = reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
var typeGoHttpRequestPtr = reflect.TypeOf(&http.Request{})

type HandlerCaller struct {
	p        AnyHandler
	pv       reflect.Value
	pt       reflect.Type
	numIn    int
	numOut   int
	argTypes []reflect.Type
}

func NewHandlerCaller(p AnyHandler) *HandlerCaller {
	hc := &HandlerCaller{p: p}
	hc.pv = reflect.ValueOf(p)
	hc.pt = hc.pv.Type()
	fmutil.MustTrue(hc.pt.Kind() == reflect.Func, "handler must be a func, but: %T", p)
	hc.numIn = hc.pt.NumIn()
	hc.numOut = hc.pt.NumOut()
	fmutil.MustTrue(1 <= hc.numIn && hc.numIn <= 3, "handler must be: func([*ChainExecutor], [*XxxRequest] ...) Response")
	fmutil.MustTrue(hc.pt.NumOut() <= 1, "handler must be: func([*ChainExecutor], [*XxxRequest] ...) Response")
	hc.argTypes = make([]reflect.Type, hc.numIn)
	for i := 0; i < hc.numIn; i++ {
		hc.argTypes[i] = hc.pt.In(i)
	}
	return hc
}

func (hc *HandlerCaller) Call(ce *ChainExecutor, isMiddleware bool) Response {
	argValues := make([]reflect.Value, hc.numIn)
	hasMiddleArg := false
	for i := 0; i < hc.numIn; i++ {
		switch hc.argTypes[i] {
		case typeMiddleRequestPtr:
			argValues[i] = reflect.ValueOf(ce)
			hasMiddleArg = true
		case typeRequestPtr:
			argValues[i] = reflect.ValueOf(ce.context)
		case typeGoHttpRequestPtr:
			argValues[i] = reflect.ValueOf(ce.context.Request)
		case typeResponseWriter:
			argValues[i] = reflect.ValueOf(ce.context.ResponseWriter)
		case ce.context.wrappedContextType:
			argValues[i] = reflect.ValueOf(ce.context.WrappedContext)
		default:
			fmutil.Panic("unsupported handler argument type: %T", hc.argTypes[i])
		}
	}
	if isMiddleware != hasMiddleArg {
		fmutil.Panic("only one endpoint handler is allowed for a request, middleware handler must accept the argument ChainExecutor")
	}
	ret := hc.pv.Call(argValues)
	if hc.numOut == 1 {
		if ret[0].IsNil() {
			return nil
		} else {
			return ret[0].Interface().(Response)
		}
	} else {
		if ce.context.IsResponseWritten() {
			return responseNop
		}
		return nil
	}
}

type ChainExecutor struct {
	chain   *handlerChain
	context *Context
	nextIdx int
}

func (ce *ChainExecutor) Next() Response {
	// TODO: support change *Request and ResponseWriter when calling Next
	if ce.nextIdx >= len(ce.chain.middlewares) {
		return ce.chain.endpoint.Call(ce, false)
	} else {
		p := ce.chain.middlewares[ce.nextIdx]
		ce.nextIdx++
		return p.Call(ce, true)
	}
}

type handlerChain struct {
	middlewares []*HandlerCaller
	endpoint    *HandlerCaller
}

func (ch *handlerChain) addMiddleware(h ...AnyHandler) *handlerChain {
	for _, p := range h {
		if m2, ok := p.(*handlerChain); ok {
			ch.middlewares = append(ch.middlewares, m2.middlewares...)
		} else {
			ch.middlewares = append(ch.middlewares, NewHandlerCaller(p))
		}
	}
	return ch
}

func (ch *handlerChain) addEndpoint(h ...AnyHandler) *handlerChain {
	fmutil.MustTrue(ch.endpoint == nil, "one handler chain can only have one endpoint")
	for _, p := range h {
		switch h := p.(type) {
		case *handlerChain:
			ch.middlewares = append(ch.middlewares, h.middlewares...)
		case RequestHandler:
			ch.middlewares = append(ch.middlewares, NewHandlerCaller(h.Handle))
		default:
			ch.middlewares = append(ch.middlewares, NewHandlerCaller(p))
		}
	}
	cnt := len(ch.middlewares)
	fmutil.MustTrue(cnt != 0, "no endpoint handler")
	ch.endpoint = ch.middlewares[cnt-1]
	ch.middlewares = ch.middlewares[:cnt-1]
	return ch
}

func (ch *handlerChain) Handle(c *Context) Response {
	mr := &ChainExecutor{
		chain:   ch,
		context: c,
	}
	return mr.Next()
}

func (ch *handlerChain) clone() *handlerChain {
	ch1 := &handlerChain{}
	if ch != nil {
		ch1.middlewares = append(ch1.middlewares, ch.middlewares...)
	}
	return ch1
}
