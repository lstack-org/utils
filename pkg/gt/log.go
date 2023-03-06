// Package gt ...
package gt

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
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

const (
	tolerateTime = 5000 * time.Millisecond
)

// LogLevel 日志打印等级
type LogLevel int

const (
	// LogInfo Print source host，method， request URL
	LogInfo LogLevel = 1
	// LogDebug Print request body and response
	LogDebug LogLevel = 2
)

var (
	color = reset
)

func NewLogTrace(l LogLevel) http.RoundTripper {
	return &logTrace{
		LogLevel:              l,
		delegatedRoundTripper: &http.Transport{},
	}
}

type logTrace struct {
	LogLevel
	delegatedRoundTripper http.RoundTripper
}

func (l *logTrace) RoundTrip(request *http.Request) (*http.Response, error) {
	start := time.Now()
	var requestBodyStr string
	var responseBodyStr string
	if request.Body != nil {
		requestBody, _ := ioutil.ReadAll(request.Body)
		requestBodyStr = string(requestBody)
		request.Body = ioutil.NopCloser(bytes.NewReader(requestBody))
	}

	response, err := l.delegatedRoundTripper.RoundTrip(request)

	if err != nil {
		return response, err
	}

	if response != nil {
		responseBody, _ := ioutil.ReadAll(response.Body)
		responseBodyStr = string(responseBody)
		response.Body = ioutil.NopCloser(bytes.NewReader(responseBody))
	}

	cost := time.Now().Sub(start)
	if cost > tolerateTime {
		color = red
	}

	fmt.Printf("[SOURCE]: %s [METHOD]: %s [COST]:%s%s\u001B[0m [URL]:%s\n",
		request.Host, request.Method, color, time.Now().Sub(start), request.URL)

	if l.LogLevel > LogInfo {
		fmt.Printf("RequestBody: %s\n", requestBodyStr)
		fmt.Printf("Response: %s\n ", responseBodyStr)
	}

	return response, err
}
