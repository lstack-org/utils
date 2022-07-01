package k8s

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/tools/clientcmd"

	"net/http"
)

func NewFakeClient(roundTripper func(*http.Request) (*http.Response, error), fns ...ReqTransformFn) (Interface, error) {
	return newClient(newFakeRestConfig(), fake.CreateHTTPClient(roundTripper), fns...)
}

func newFakeRestConfig() *rest.Config {
	clientConfig, _ := clientcmd.NewClientConfigFromBytes([]byte(kubeConfig))
	restConfig, _ := clientConfig.ClientConfig()
	return restConfig
}

const kubeConfig = ``