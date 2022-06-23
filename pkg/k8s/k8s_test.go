package k8s

import (
	"context"
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

func TestSkipLog(t *testing.T) {
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

	err = client.Resource(svcGroupVersionResource).Transform(SkipLog).Get("sdsd", nil, metav1.GetOptions{})
	t.Log(err)
}

const manifest = `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wjf
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      run: wjf
  template:
    metadata:
      labels:
        run: wjf
    spec:
      containers:
        - name: main
          image: nginx:latest
          ports:
            - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: wjf
  namespace: default
spec:
  selector:
    run: wjf
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: wjf
  namespace: default
spec:
  rules:
  - http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: wjf
            port:
              number: 80
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: wjf
  namespace: default
spec:
  storageClassName: wjf
  accessModes:
    - RWO
  resources:
    requests:
      storage: 1Gi
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: wjf
  namespace: default
data:
  key: value
---
apiVersion: v1
kind: Secret
metadata:
  name: wjf
  namespace: default
# 请正确填写所需类型，如：
# Opaque 用户定义的任意数据
# kubernetes.io/service-account-token 服务账号令牌
# kubernetes.io/dockercfg ~/.dockercfg 文件的序列化形式
# kubernetes.io/dockerconfigjson ~/.docker/config.json 文件的序列化形式
# kubernetes.io/basic-auth 用于基本身份认证的凭据
# kubernetes.io/ssh-auth 用于 SSH 身份认证的凭据
# kubernetes.io/tls 用于 TLS 客户端或者服务器端的数据
# bootstrap.kubernetes.io/token 启动引导令牌数据
# 更多信息可见文档：https://kubernetes.io/zh/docs/concepts/configuration/secret/
type: kubernetes.io/dockercfg
data:
  .dockercfg: |
        "<base64 encoded ~/.dockercfg file>"
`

func TestYamlsApply(t *testing.T) {
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

	err = client.YamlsApply(context.TODO(), manifest)
	if err != nil {
		t.Fatal(err)
	}

}
