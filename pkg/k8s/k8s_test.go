package k8s

import (
	"github.com/lstack-org/utils/pkg/rest"
	v1 "k8s.io/api/core/v1"
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
			Name: "wjf-test",
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
				"app": "wjf-test",
			},
			Type: v1.ServiceTypeClusterIP,
		},
	}, nil, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	svc := &v1.Service{}
	err = resourceInterface.Get("wjf-test", svc, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if svc.Name != "wjf-test" {
		t.Fatal("expect svc name = wjf-test,but got " + svc.Name)
	}

	svc.Spec.Type = v1.ServiceTypeNodePort
	err = resourceInterface.Apply(svc, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = resourceInterface.Delete("wjf-test", metav1.NewDeleteOptions(0))
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

	if svc.Name != "wjf-test" {
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
		Age string
		Ports string `json:"Port(s)"`
	}

	var intoData []test
	err = TableHandle(table, &intoData)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(intoData)

	// output
	// k8s_test.go:146: [{canary NodePort 40d 443:30290/TCP} {clickhouse-service NodePort 53d 9000:31201/TCP,8123:32513/TCP} {forward-xlauncher1-lstack-internal ClusterIP 48d 443/TCP} {gitlab-provider NodePort 21d 8080:31009/TCP} {hello-service NodePort 34d 9999:31074/TCP} {idp-base NodePort 13d 9090:30573/TCP} {idp-catalog NodePort 20d 9090:30555/TCP} {idp-ci ClusterIP 8h 9090/TCP} {idp-dataservice NodePort 27d 9090:32047/TCP} {idp-infrastructure ClusterIP 15d 9090/TCP} {idp-logserver NodePort 15d 9090:30985/TCP,9091:30350/TCP} {idp-server NodePort 15d 8080:30637/TCP} {keystone ClusterIP 52d 80/TCP,443/TCP} {keystone-api ClusterIP 52d 5000/TCP} {kubernetes ClusterIP 53d 443/TCP} {log-server-lsh-mcp-log-service NodePort 36d 8080:32745/TCP} {logs-agent-lyf NodePort 49d 8080:31079/TCP} {logs-agent-server NodePort 48d 8080:31718/TCP} {lsh-lpm-cluster-maintenance-service ClusterIP 52d 9090/TCP,8081/TCP} {lsh-mcp-app-manage NodePort 52d 9090:30512/TCP,9091:31968/TCP} {lsh-mcp-cc-alert-service ClusterIP 52d 9090/TCP,8081/TCP} {lsh-mcp-cc-sms-service ClusterIP 52d 9090/TCP} {lsh-mcp-chartmuseum NodePort 52d 8080:30386/TCP} {lsh-mcp-cloudmarket-apigateway-service ClusterIP 52d 9090/TCP} {lsh-mcp-cloudmarket-service ClusterIP 52d 9090/TCP} {lsh-mcp-cnops NodePort 52d 9090:32017/TCP} {lsh-mcp-common-jobserver NodePort 53d 8081:30098/TCP} {lsh-mcp-cpm-log-service ClusterIP 52d 9090/TCP} {lsh-mcp-cpm-om-service ClusterIP 52d 9090/TCP,8081/TCP} {lsh-mcp-csm-event-service ClusterIP 52d 9090/TCP,8081/TCP} {lsh-mcp-csm-log-service NodePort 52d 9090:30097/TCP} {lsh-mcp-csm-om-service ClusterIP 52d 9090/TCP,8081/TCP} {lsh-mcp-elasticsearch-lsh-mcp-elasticsearch NodePort 49d 9200:30176/TCP,9300:31351/TCP} {lsh-mcp-elasticsearch-lsh-mcp-elasticsearch-headless ClusterIP 49d 9200/TCP,9300/TCP} {lsh-mcp-iam-apigateway-service ClusterIP 52d 9090/TCP} {lsh-mcp-iam-auth-service NodePort 52d 9090:30194/TCP} {lsh-mcp-iam-notification-service ClusterIP 52d 9090/TCP} {lsh-mcp-iam-tenantm-service NodePort 52d 9090:32477/TCP} {lsh-mcp-idp-cd ClusterIP 14d 9090/TCP} {lsh-mcp-lcr-cops NodePort 52d 9090:31532/TCP} {lsh-mcp-lcr-service ClusterIP 52d 8080/TCP} {lsh-mcp-lcs-appmarket ClusterIP 52d 8080/TCP} {lsh-mcp-lcs-console-api-go NodePort 52d 8080:32418/TCP} {lsh-mcp-lcs-consoleapi-service NodePort 52d 9090:30802/TCP} {lsh-mcp-lcs-om-service NodePort 52d 9090:30963/TCP} {lsh-mcp-lcs-out-gateway NodePort 52d 80:30261/TCP} {lsh-mcp-lcs-reso-service NodePort 52d 9090:30109/TCP} {lsh-mcp-lcs-timer ClusterIP 52d 80/TCP} {lsh-mcp-luc-operation-center NodePort 52d 9090:30885/TCP,5005:30888/TCP} {lsh-mcp-luc-operative-service NodePort 52d 9090:30126/TCP} {lsh-mcp-luc-pay-service ClusterIP 52d 9090/TCP} {lsh-mcp-mongo NodePort 53d 27017:30289/TCP} {lsh-mcp-wo-service ClusterIP 52d 9090/TCP} {lsh-operation-management-service NodePort 42d 9090:31892/TCP} {lsh-static-server NodePort 21d 80:30080/TCP,443:30081/TCP,8669:30082/TCP,5000:30083/TCP} {mysql-service NodePort 53d 3306:32644/TCP} {mysql-test-mysql NodePort 27d 3306:30950/TCP} {nginx-lyf NodePort 47d 80:31190/TCP} {ouyang-mysql-svc NodePort 8d 3306:32306/TCP} {redis NodePort 53d 6379:30433/TCP} {server-forward NodePort 52d 9999:31985/TCP,443:31400/TCP,80:31480/TCP} {syjtest-xlauncher1-lstack-internal ClusterIP 14d 443/TCP}]
}
