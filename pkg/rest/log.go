package rest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
	"time"
)

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

type logTrace struct {
	title                 string
	delegatedRoundTripper http.RoundTripper
}

func (l *logTrace) RoundTrip(request *http.Request) (*http.Response, error) {
	var (
		requestBody    []byte
		requestBodyStr string
		start          = time.Now()
		path           = request.URL.Path
		host           = request.Host
		method         = request.Method
		raw            = request.URL.RawQuery
		log            string
	)

	if request.Body != nil {
		requestBody, _ = ioutil.ReadAll(request.Body)
		requestBodyStr = string(requestBody)
		request.Body = ioutil.NopCloser(bytes.NewReader(requestBody))
	}

	response, err := l.delegatedRoundTripper.RoundTrip(request)

	if response != nil {
		var (
			latency         = time.Now().Sub(start)
			statusCode      = response.StatusCode
			responseBody, _ = ioutil.ReadAll(response.Body)
			responseBodyStr = string(responseBody)
			colorP          = colorParams{
				StatusCode: statusCode,
				Method:     method,
			}
		)
		response.Body = ioutil.NopCloser(bytes.NewReader(responseBody))

		if raw != "" {
			path = path + "?" + raw
		}

		var (
			statusColor = colorP.StatusCodeColor()
			methodColor = colorP.MethodColor()
			resetColor  = colorP.ResetColor()
		)

		log = fmt.Sprintf("[%s] %v | %s %3d %s | %13v | %15s |%s %-7s %s %#v\n",
			strings.ToUpper(l.title),
			start.Format("2006/01/02 - 15:04:05"),
			statusColor, statusCode, resetColor,
			latency,
			host,
			methodColor, method, resetColor,
			path,
		)

		if request.Header != nil {
			if klog.V(7).Enabled() {
				log = fmt.Sprintf("%sRequest Header: %v\n", log, request.Header)
			}
		}

		if requestBodyStr != "" {
			if klog.V(7).Enabled() {
				log = fmt.Sprintf("%sRequest Body: %s\n", log, requestBodyStr)
			}
		}

		resHeader := response.Header
		if resHeader != nil {
			if klog.V(7).Enabled() {
				log = fmt.Sprintf("%sResponse Header: %v\n", log, resHeader)
			}
		}

		if responseBodyStr != "" {
			if klog.V(7).Enabled() {
				log = fmt.Sprintf("%sResponse Body: %s\n", log, responseBodyStr)
			}
		}

		if klog.V(6).Enabled() {
			fmt.Print(log)
		}
	}

	return response, err
}

type colorParams struct {
	StatusCode int
	Method     string
}

func (p *colorParams) StatusCodeColor() string {
	code := p.StatusCode

	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}

func (p *colorParams) MethodColor() string {
	method := p.Method

	switch method {
	case http.MethodGet:
		return blue
	case http.MethodPost:
		return cyan
	case http.MethodPut:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodPatch:
		return green
	case http.MethodHead:
		return magenta
	case http.MethodOptions:
		return white
	default:
		return reset
	}
}

func (p *colorParams) ResetColor() string {
	return reset
}
