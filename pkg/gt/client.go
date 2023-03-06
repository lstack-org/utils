package gt

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/http2"
	"gopkg.in/yaml.v3"
)

// Client wrap http.Client, used for Do request
type Client struct {
	ht       *http.Client
	Method   string
	URL      string
	Body     io.Reader
	Header   http.Header
	Resp     *http.Response
	LogLevel LogLevel
	Err      error
	Option
}

const (
	Timeout10 time.Duration = 10
	Timeout20 time.Duration = 20
)

const (
	contentType     = "Content-Type"
	xmlContentType  = "application/xml"
	jsonContentType = "application/json"
)

// Option http request option
type Option struct {
	Timeout time.Duration
}

var (
	defaultHeader = http.Header{
		contentType: []string{jsonContentType},
	}
)

// NewClient return http default client
func NewClient() *Client {
	return &Client{
		ht:     http.DefaultClient,
		Header: http.Header{},
	}
}

// NewDefaultClient return http default client and set Json header
func NewDefaultClient() *Client {
	return &Client{
		ht:     http.DefaultClient,
		Header: defaultHeader,
		Option: Option{
			Timeout: time.Second * Timeout10,
		},
		LogLevel: LogInfo,
	}
}

func (c *Client) GET(url string) *Client {
	c.Method = "GET"
	c.URL = url
	return c
}

func (c *Client) POST(url string) *Client {
	c.Method = "POST"
	c.URL = url
	return c
}

func (c *Client) DELETE(url string) *Client {
	c.Method = "DELETE"
	c.URL = url
	return c
}

func (c *Client) PUT(url string) *Client {
	c.Method = "PUT"
	c.URL = url
	return c
}

func (c *Client) UPDATE(url string) *Client {
	c.Method = "UPDATE"
	c.URL = url
	return c
}

func (c *Client) HEADER(url string) *Client {
	c.Method = "HEADER"
	c.URL = url
	return c
}

func (c *Client) OPTIONS(url string) *Client {
	c.Method = "OPTIONS"
	c.URL = url
	return c
}

func (c *Client) SetTimeout(duration time.Duration) *Client {
	c.Option.Timeout = duration
	return c
}

// EnableLog 重写RoundTrip方法，打印Request和Resp等信息
func (c *Client) EnableLog(l LogLevel) *Client {
	c.ht.Transport = NewLogTrace(l)
	return c
}

func transformResponse(resp *http.Response) error {
	var body []byte
	if resp.Body != nil {
		data, err := io.ReadAll(resp.Body)
		switch err.(type) {
		case nil:
			body = data
		case http2.StreamError:
			return fmt.Errorf("stream error when reading response body, may be caused by closed connection. Please retry. Original error: %w", err)
		default:
			return fmt.Errorf("unexpected error when reading response body. Please retry. Original error: %w", err)
		}
	}
	if len(body) == 0 {
		return errors.New(resp.Status)
	}
	return errors.New(string(body))
}

func (c *Client) Do() *Client {
	req, err := http.NewRequest(c.Method, c.URL, c.Body)
	if err != nil {
		c.Err = err
		return c
	}
	req.Header = c.Header
	c.ht.Timeout = c.Option.Timeout

	c.Resp, err = c.ht.Do(req)
	if err != nil {
		c.Err = err
		return c
	}

	switch {
	case c.Resp.StatusCode == http.StatusSwitchingProtocols:
		// no-op, we've been upgraded
	case c.Resp.StatusCode >= http.StatusInternalServerError:
		c.Err = errors.New("服务器内部错误")
		return c
	case c.Resp.StatusCode < http.StatusOK || c.Resp.StatusCode > http.StatusPartialContent:
		c.Err = transformResponse(c.Resp)
		return c
	}

	return c
}

// SetBody set request body
func (c *Client) SetBody(body io.Reader) *Client {
	c.Body = body
	return c
}

// AddHeader add http.Header
func (c *Client) AddHeader(header http.Header) *Client {
	for key, values := range header {
		for _, v := range values {
			c.Header.Add(key, v)
		}
	}
	return c
}

func (c *Client) SetHeader(key string, values ...string) *Client {
	for _, v := range values {
		c.Header.Add(key, v)
	}
	return c
}

func (c *Client) InTo(object interface{}, format DecodeFormat) error {
	if c.Err != nil {
		return c.Err
	}
	if object == nil {
		return nil
	}
	switch format {
	case JSON:
		decode := json.NewDecoder(c.Resp.Body)
		return decode.Decode(object)
	case YAML:
		decode := yaml.NewDecoder(c.Resp.Body)
		return decode.Decode(object)
	case BODY:
		decode := NewBodyDecode(c.Resp.Body)
		return decode.Decode(object)
	}
	return DecoderTypeNotSupport
}
