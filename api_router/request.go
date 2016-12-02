package api_router

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type contextKey struct {
	name string
}

var requestContextCtxKey = &contextKey{"RequestContext"}

func (self *contextKey) String() string {
	return "go-api-controller context value " + self.name
}

// Satisfies context.Context interface
type RequestContext struct {
	context.Context
	request             *http.Request
	writer              ResponseWriter
	currentRoute        *Route
	routeVars           map[string]string
	statusHeaderWritten bool
	startTime           time.Time
}

func (self *RequestContext) Value(key interface{}) interface{} {
	if key == requestContextCtxKey {
		return self
	}
	return self.Context.Value(key)
}

func (self *RequestContext) Body() io.ReadCloser {
	return self.request.Body
}

func (self *RequestContext) BodyCopy() (buf []byte, err error) {
	body := self.request.Body
	buf, err = ioutil.ReadAll(body)
	if err == nil {
		defer body.Close()
		self.request.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	}
	return
}

func (self *RequestContext) SetBody(body io.ReadCloser) {
	self.request.Body = body
}

func (self *RequestContext) HTTPRequest() *http.Request {
	return self.request
}

func (self *RequestContext) WithContext(ctx context.Context) *RequestContext {
	self.request = self.request.WithContext(ctx)
	return self
}

func (self *RequestContext) CurrentRoute() *Route {
	return self.currentRoute
}

func (self *RequestContext) StartTime() time.Time {
	return self.startTime
}

func (self *RequestContext) TimeElapsed() time.Duration {
	return time.Since(self.startTime)
}

func (self *RequestContext) RouteVar(name string) (string, bool) {
	val, ok := self.routeVars[name]
	return val, ok
}

func (self *RequestContext) Header(name string) string {
	return self.request.Header.Get(name)
}

func (self *RequestContext) SetStatus(status int) {
	self.writer.SetStatus(status)
}

func (self *RequestContext) SetResponseHeader(name, val string) {
	self.writer.Header()[name] = []string{val}
}

func (self *RequestContext) ResponseWriter() ResponseWriter {
	return self.writer
}

func (self *RequestContext) WriteStatusHeader() {
	self.writer.WriteStatusHeader()
}

func (self *RequestContext) WriteResponse(data []byte) (err error) {
	_, err = self.writer.Write(data)
	return
}

func (self *RequestContext) WriteResponseString(data string) (err error) {
	_, err = io.WriteString(self.writer, data)
	return
}

func NewContextForRequest(w ResponseWriter, r *http.Request, cur_route *Route) *RequestContext {
	vars := cur_route.RouteVars(r)
	if vars == nil {
		vars = make(map[string]string)
	}

	req_ctx := &RequestContext{
		Context:      r.Context(),
		writer:       w,
		currentRoute: cur_route,
		routeVars:    vars,
		startTime:    time.Now(),
	}

	req_ctx.request = r.WithContext(req_ctx)
	return req_ctx
}

func RequestContextFromContext(ctx context.Context) *RequestContext {
	val := ctx.Value(requestContextCtxKey)
	if val == nil {
		return nil
	}
	return val.(*RequestContext)
}
