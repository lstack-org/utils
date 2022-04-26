package k8s

import (
	"github.com/lstack-org/utils/pkg/rest"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/clientcmd"

	"net/http"
	"testing"
	"time"
)

var (
	svcGroupVersionResource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
)

const (
	testName = "wjf-test"
)

func TestClient(t *testing.T) {
	rest.SetLogLevel(0, 0)
	clientConfig, err := clientcmd.NewClientConfigFromBytes([]byte(kubeConfig))
	if err != nil {
		t.Fatal(err)
	}
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		t.Fatal(err)
	}

	restConfig.Timeout = 10 * time.Second

	dynamicInterface, err := NewClient(restConfig)
	if err != nil {
		t.Fatal(err)
	}

	resourceInterface := dynamicInterface.Resource(svcGroupVersionResource).Namespace("")

	err = resourceInterface.Create(&v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: testName,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "port-1",
					Protocol:   v1.ProtocolTCP,
					Port:       9090,
					TargetPort: intstr.FromInt(9090),
				},
			},
			Selector: map[string]string{
				"app": testName,
			},
			Type: v1.ServiceTypeClusterIP,
		},
	}, nil, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	svc := &v1.Service{}
	err = resourceInterface.Get(testName, svc, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if svc.Name != testName {
		t.Fatal("expect svc name = wjf-test,but got " + svc.Name)
	}

	svc.Spec.Type = v1.ServiceTypeNodePort
	err = resourceInterface.Apply(svc, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = resourceInterface.Delete(testName, metav1.NewDeleteOptions(0))
	if err != nil {
		t.Fatal(err)
	}
}

func TestFakeClient(t *testing.T) {
	svcStr := `{"kind":"Service","apiVersion":"v1","metadata":{"name":"wjf-test","namespace":"default","selfLink":"/api/v1/namespaces/default/services/wjf-test","uid":"70643289-3871-48c6-aa8c-fe5535dbebb3","resourceVersion":"8011981","creationTimestamp":"2021-12-29T08:34:31Z"},"spec":{"ports":[{"name":"port-1","protocol":"TCP","port":9090,"targetPort":9090}],"selector":{"app":"wjf-test"},"clusterIP":"20.111.42.59","type":"ClusterIP","sessionAffinity":"None"},"status":{"loadBalancer":{}}}`

	client, err := NewFakeClient(func(request *http.Request) (response *http.Response, e error) {
		switch path, method := request.URL.Path, request.Method; {
		case path == "/api/v1/namespaces/default/services/wjf-test" && method == "GET":
			return &http.Response{StatusCode: http.StatusOK, Header: rest.DefaultHeader(), Body: rest.StringBody(svcStr)}, nil
		default:
			panic("unexpected request")
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	resourceInterface := client.Resource(svcGroupVersionResource).Namespace("default")
	svc := &v1.Service{}
	err = resourceInterface.Get("wjf-test", svc, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if svc.Name != testName {
		t.Fatal("expect svc name = wjf-test,but got " + svc.Name)
	}
}

func TestTableClient(t *testing.T) {
	rest.SetLogLevel(0, 0)
	clientConfig, err := clientcmd.NewClientConfigFromBytes([]byte(kubeConfig))
	if err != nil {
		t.Fatal(err)
	}
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		t.Fatal(err)
	}

	restConfig.Timeout = 10 * time.Second

	client, err := NewClient(restConfig)
	if err != nil {
		t.Fatal(err)
	}

	table := &metav1beta1.Table{}
	err = client.Resource(svcGroupVersionResource).Transform(AsTable).Namespace("").List(table, metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	//此处定义表头对应的字段
	type test struct {
		Name  string
		Type  string
		Age   string
		Ports string `json:"Port(s)"`
	}

	var intoData []test
	err = TableHandle(table, &intoData)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(intoData)

	err = client.Resource(svcGroupVersionResource).Get("sdsd", nil, metav1.GetOptions{})
	t.Log(err) //code: 404,reason: NotFound,message: the server could not find the requested resource

	t.Log(errors.IsNotFound(err)) //true
}
