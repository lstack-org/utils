package rest

import (
	v1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"

	"net/http"
	"net/url"
)

type RESTClient struct {
	decorator *rest.RESTClient
}

func NewRESTClientEasy(clientName, baseURL string, customize *http.Client) (*RESTClient, error) {
	parsedUrl, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	if customize == nil {
		customize = &http.Client{}
	}
	return NewRESTClientWithLogTrace(clientName, parsedUrl, nil, customize), nil
}

func NewRESTClientWithLogTrace(clientName string, baseURL *url.URL, rateLimiter flowcontrol.RateLimiter, client *http.Client) *RESTClient {
	if client != nil {
		if client.Transport == nil {
			//client.Transport = transport.DebugWrappers(&http.Transport{})
			client.Transport = &logTrace{
				title:                 clientName,
				delegatedRoundTripper: &http.Transport{},
			}
		}
	}
	return NewRESTClient(baseURL, rateLimiter, client)
}

func NewRESTClient(baseURL *url.URL, rateLimiter flowcontrol.RateLimiter, client *http.Client) *RESTClient {
	gvCopy := v1.SchemeGroupVersion
	restClient, _ := rest.NewRESTClient(baseURL, "", rest.ClientContentConfig{
		ContentType:  runtime.ContentTypeJSON,
		GroupVersion: gvCopy,
		Negotiator:   runtime.NewClientNegotiator(scheme.Codecs.WithoutConversion(), gvCopy),
	}, rateLimiter, client)
	return &RESTClient{
		decorator: restClient,
	}
}

func (c *RESTClient) GetK8sRESTClient() *rest.RESTClient {
	return c.decorator
}

// GetRateLimiter returns rate limiter for a given client, or nil if it's called on a nil client
func (c *RESTClient) GetRateLimiter() flowcontrol.RateLimiter {
	if c == nil {
		return nil
	}
	return c.decorator.GetRateLimiter()
}

// Verb sets the verb this request will use.
func (c *RESTClient) Verb(verb string) *Request {
	return NewRequest(c).Verb(verb).SetHeader("Content-Type", runtime.ContentTypeJSON)
}

// Post begins a POST request. Short for c.Verb("POST").
func (c *RESTClient) Post() *Request {
	return c.Verb("POST")
}

// Put begins a PUT request. Short for c.Verb("PUT").
func (c *RESTClient) Put() *Request {
	return c.Verb("PUT")
}

// Patch begins a PATCH request. Short for c.Verb("Patch").
func (c *RESTClient) Patch() *Request {
	return c.Verb("PATCH")
}

// Get begins a GET request. Short for c.Verb("GET").
func (c *RESTClient) Get() *Request {
	return c.Verb("GET")
}

// Delete begins a DELETE request. Short for c.Verb("DELETE").
func (c *RESTClient) Delete() *Request {
	return c.Verb("DELETE")
}
