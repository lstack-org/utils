package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/client-go/rest/fake"
	"net/http"
	"testing"
)

var (
	client *RESTClient
)

const (
	baseURL = "http://8.16.0.211:30802"
	getRes  = "haha"
)

func init() {
	c, err := NewRESTClientEasy("om", baseURL, fake.CreateHTTPClient(func(request *http.Request) (response *http.Response, e error) {
		switch path, method := request.URL.Path, request.Method; {
		case path == "/a/g/f/g" && method == "GET":
			return &http.Response{StatusCode: http.StatusOK, Header: DefaultHeader(), Body: StringBody(getRes)}, nil
		case path == "/n/t/m/d" && method == "POST":
			bytes, _ := ioutil.ReadAll(request.Body)
			return &http.Response{StatusCode: http.StatusOK, Header: DefaultHeader(), Body: BytesBody(bytes)}, nil
		default:
			panic("unexpected request")
		}
	}))
	if err != nil {
		panic(err)
	}
	client = c
}

func TestUrl(t *testing.T) {
	u := client.Get().
		AbsPath("/d/f/g").
		Param("sd", "df").
		Param("key1", "value1").
		URL().String()
	if u != fmt.Sprintf("%s%s", baseURL, "/d/f/g?key1=value1&sd=df") {
		t.Fatalf("unexpected output: %s", u)
	}
}

func TestGet(t *testing.T) {
	bytes, e := client.Get().
		AbsPath("/a/g/f/g").
		Param("key1", "value").
		SetHeader("token", "abcdefg").
		DoRaw(context.TODO())
	if e != nil {
		t.Fatal(e)
	}
	if getRes != string(bytes) {
		t.Fatalf("unexpected output: %s", string(bytes))
	}
}

func TestPost(t *testing.T) {
	resp := &map[string]interface{}{}
	err := client.Post().
		AbsPath("/n/t/m/d").
		Body(map[string]interface{}{
			"a": "b",
			"s": "f",
		}).DoInto(context.TODO(), resp)
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ := json.Marshal(resp)
	if `{"a":"b","s":"f"}` != string(bytes) {
		t.Fatalf("unexpected output: %s", string(bytes))
	}
}

func TestParams(t *testing.T) {
	u := client.Get().
		AbsPath("/ddd/ggg/hhh").
		Params(&query{
			A: "a",
			B: "b",
			C: 3,
		}).URL().String()
	if u != fmt.Sprintf("%s%s", baseURL, "/ddd/ggg/hhh?a=a&b=b&c=3") {
		t.Fatalf("unexpected output: %s", u)
	}
}

type query struct {
	A string `json:"a,omitempty"`
	B string `json:"b,omitempty"`
	C int    `json:"c,omitempty"`
	D bool   `json:"d,omitempty"`
}
