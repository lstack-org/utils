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
		)
		response.Body = ioutil.NopCloser(bytes.NewReader(responseBody))

		if raw != "" {
			path = path + "?" + raw
		}

		log = fmt.Sprintf("[%s] %v | %v | %13v | %15s | %-7s %#v\n",
			strings.ToUpper(l.title),
			start.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			host,
			method,
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
