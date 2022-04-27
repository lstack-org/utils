package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"

	"k8s.io/apimachinery/pkg/conversion/queryparams"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
)

type Request struct {
	err       error
	decorator *rest.Request
}

func NewRequest(c *RESTClient) *Request {
	return &Request{
		decorator: rest.NewRequest(c.decorator),
	}
}

// Verb sets the verb this request will use.
func (r *Request) Verb(verb string) *Request {
	r.decorator.Verb(verb)
	return r
}

func (r *Request) AbsPath(path string) *Request {
	r.decorator.AbsPath([]string{path}...)
	return r
}

func (r *Request) AbsPathf(format string, a ...interface{}) *Request {
	r.AbsPath(fmt.Sprintf(format, a...))
	return r
}

// Param creates a query parameter with the given string value.
func (r *Request) Param(paramName, s string) *Request {
	r.decorator.Param(paramName, s)
	return r
}

//Params 会把query转换为url.Values
//query 必须为reflect.Ptr, reflect.Interface类型
//tag: json , 为空忽略：omitempty
func (r *Request) Params(query interface{}) *Request {
	values, err := queryparams.Convert(query)
	if err != nil {
		r.err = err
		return r
	}
	for k, v := range values {
		r.Param(k, v[0])
	}
	return r
}

// BackOff sets the request's backoff manager to the one specified,
// or defaults to the stub implementation if nil is provided
func (r *Request) BackOff(manager rest.BackoffManager) *Request {
	r.decorator.BackOff(manager)
	return r
}

// Throttle receives a rate-limiter and sets or replaces an existing request limiter
func (r *Request) Throttle(limiter flowcontrol.RateLimiter) *Request {
	r.decorator.Throttle(limiter)
	return r
}

// Timeout makes the request use the given duration as an overall timeout for the
// request. Additionally, if set passes the value as "timeout" parameter in URL.
func (r *Request) Timeout(d time.Duration) *Request {
	r.decorator.Timeout(d)
	return r
}

func (r *Request) SetHeader(key string, values ...string) *Request {
	r.decorator.SetHeader(key, values...)
	return r
}

// Body makes the request use obj as the body. Optional.
// If obj is a string, try to read a file of that name.
// If obj is a []byte, send it directly.
// If obj is an io.Reader, use it directly.
// If obj is a runtime.Object, marshal it correctly, and set Content-Type header.
// If obj is a runtime.Object and nil, do nothing.
// Otherwise, set an error.
func (r *Request) Body(obj interface{}) *Request {
	switch obj.(type) {
	case string, []byte, io.Reader, runtime.Object:
		r.decorator.Body(obj)
	default:
		body, err := json.Marshal(obj)
		r.decorator.Body(body)
		r.err = err
	}
	return r
}

// URL returns the current working URL.
func (r *Request) URL() *url.URL {
	return r.decorator.URL()
}

// DoRaw executes the request but does not process the response body.
func (r *Request) DoRaw(ctx context.Context) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	bytes, err := r.decorator.DoRaw(ctx)
	return bytes, ErrorConvert(bytes, err)
}

func (r *Request) DoInto(ctx context.Context, into interface{}) error {
	resp, err := r.DoRaw(ctx)
	if err != nil {
		return err
	}
	return json.Unmarshal(resp, into)
}
