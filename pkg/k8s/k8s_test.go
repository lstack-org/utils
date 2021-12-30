package k8s

import (
	"github.com/lstack-org/utils/pkg/rest"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"testing"
)

const kubeConfig = `
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeE1URXhNakE0TWpneE0xb1hEVE14TVRFeE1EQTRNamd4TTFvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTzl2Ck1wNXFnb1hWaDJTbXBTcXRVVEJvWVpKVW4yajAxTmowTzYxdWZKN2pIZUhxait1T0ZyaTFRRXN5dGE5SzVqR0QKdlFkVG5KazZ4aGN4ZGtSVm1zS1g1TldWQlh6cCs2QmxXdnFGdGNFeXhaeVhwTkZyU2NadStVQ21iVVMyb0V6cwpJMmlvdUN6TzRrRXJDRnc1cFNFSUtwQTdTQ1hodkd2L3E0MnZSbDR4YWN6bWNHd1p0THljdkNidDdKZ3BXNTh4Cnp6bHRSNlZDME1RZzlnOXRoQ3N0ZVgvSlhUTmlpS3I3bnlDVnBzcTAvcFRUSitZeWNXRDdTMW5GZng1Sm83ZzMKQmVwZEh1MkI3ZWpHOFhMS2d0QXo1Q3NpUUs0eWpVZ09MT0VCcHFUdWtXMVZoSE1rbDNTSThvYjgySWdhNW03MAowVnpVM0NoYTh5a3p1TzBlSzlNQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFIajltUWJWYXQ0SzhEa1ZhRWJlTXdmU01aWHoKNGoyRGVWZDYxWHl2STk5NXd6aGhnWmdMa3NzT0QxUEM5UGowTW1PTkd0bVhPMWl2R044K1RuTld1azUvRnlNRgpIZjRCUWlSYUFaSUpXTnAzRGZEM3FEdExwaEFxUG5mRjZmYzA2QWNqaW5JZUtrMmlLZXhqRE00VUszc2ZObnM1CjhhMmo2VGVNT3dyd2k1cXZ4dngvbGlhTjZqbFpBelBmbXBHTzlXZ0xDaHk1TElicHREZUpMdHRMYkU1OFFnNmEKczVRSElDcU9uQmJhSVJEbG9oeHFvNWpPc2FsREExM0J3em9BRklLT1FjWVhvWUpDQVQ4T2ROOVZMUityRDlHbAozaVVTK2xTNkZyVnNDNkovc1F4WS9QTjRwNGgwaHR6bnRJWVpVSm1wUWhza0g0WHhXcGthVGcyTFl2OD0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    server: https://8.16.0.211:6443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: kubernetes-admin
  name: kubernetes-admin@kubernetes
current-context: kubernetes-admin@kubernetes
kind: Config
preferences: {}
users:
- name: kubernetes-admin
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM4akNDQWRxZ0F3SUJBZ0lJQ1NKTHlxUDZoVjB3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TVRFeE1USXdPREk0TVROYUZ3MHlNakV4TVRJd09ESTRNVFZhTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXhINzhUVCsyTU1IUmlLamQKaWxnendvYWRQcURDSktleURGQkZxTDFIWEg4YWpBMnBIRlhCMW5zcnh6SHhJRTg2Nng0ekhmc2tTV3JPMjZ4YgpmeXNlTEhxNXBPa1hSNHdEb3k4a1hMNkdPdVZSTWs3dUN4YjhoV2lzcW5jdmZNQ3RjQUROQURHVkJpd01laVZJCmljK3lYTzB4OUdBRVZZU0xlU1VoYVBLNm5IR1FxTGFPQkMzelF1dEdPOU1IN0x2ZzFnbU5aT3ltWmRid0VzcmYKQURlU2szTzQyanNQZjNTZ2JLYkFubjlFMDBSWUJOb2tKdUdza0pqeDN6ZGxQeWF6RzRFYkk0ay9OeU53NGpRRAp4M0pPWmtSTnVDTTcxL2J3TzcyQlNVUncvdFh4QkpRZzl5clZKMy9uVGlnWlUzZWx5V3dTNzA5dmRRZ0l2T1pNCnloVCs0d0lEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFFVmtUeVN1ckFNaXRnTEdET29mbTFVdk5Fc2ticXFzVXB0VwpENTBBazgvMEh5RzFWR2NJSnN3R3R4UEp2Yjh3MXVVNEhhbENyQnlhck9PM0FRNmdiejljc3BvTTR2VkdDVDI2ClgrazREeWVwdG1SRmZRd01ISTBoNXBobGptVDlLRjFqVFJIbUdybG8xSHh6ekNxVllHSWZ4K28rZ05UUkJwL0oKalJGRW1jMW9GelRIUkRvbEpUU1VTdzVmb2l6SmFSbEVmZ1gvNjJ5bHRkSjRkSlJadUtRbkQvNE91V2JPOERTSQpGbkdvVkNHSENBLzgzUUlUOGE1eUd3YUpPZGdTV1NoZm9JRFZRakhMUms5b3JvQVlvMVUwNkhwVHp2SEpjM2RECkZvVW9nNG1LaWczS0pVam1ucE9odjdOaVZDTnlDcDRFRVkvd3N1VlFDOXhlTStxUVBqRT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb2dJQkFBS0NBUUVBeEg3OFRUKzJNTUhSaUtqZGlsZ3p3b2FkUHFEQ0pLZXlERkJGcUwxSFhIOGFqQTJwCkhGWEIxbnNyeHpIeElFODY2eDR6SGZza1NXck8yNnhiZnlzZUxIcTVwT2tYUjR3RG95OGtYTDZHT3VWUk1rN3UKQ3hiOGhXaXNxbmN2Zk1DdGNBRE5BREdWQml3TWVpVklpYyt5WE8weDlHQUVWWVNMZVNVaGFQSzZuSEdRcUxhTwpCQzN6UXV0R085TUg3THZnMWdtTlpPeW1aZGJ3RXNyZkFEZVNrM080MmpzUGYzU2diS2JBbm45RTAwUllCTm9rCkp1R3NrSmp4M3pkbFB5YXpHNEViSTRrL055Tnc0alFEeDNKT1prUk51Q003MS9id083MkJTVVJ3L3RYeEJKUWcKOXlyVkozL25UaWdaVTNlbHlXd1M3MDl2ZFFnSXZPWk15aFQrNHdJREFRQUJBb0lCQUJweWtSa0FxMUFTdGxZegpqR1lUaXh2eXJIV0NnNzhWUnpTN0ZUVXFETkhaVmNSbURrMy9DUEVLY1JFRm10UGpkaVd4VWVZR0tKTXRLaHlOCkxWK0hlUzg1Y1lWTnpsRlYraU5idEFRN3JLdCt0QmdXWVpuaWhTaWJ0eW5Xa3ZDeXFtVjU1aDNSanFKZkNXcmoKVzhrWXlJUVRkUGJVZWFEZER6ekdENklsa1pKK3hqUGptUzR3RHExS1lDRzBrTE9LRzVPdVArY0hrYTVVeTZVLwpKeHZMazRtTy9xOGxFUWVGZzRYdWp6T1lUK3IrZEttNjAvYjZFeExab0RGOU1jVkFTc0pLb0hCb0xKTEZCV2pDCi9kOTZCRWhnUVp0TkpFYmF4Z1A1MHhSNmZaR0E3d3kzejMxaklRTERqbjVTM3VHRjd0YVZqWlJsTW5iSXBteTQKK2ZGa3hQRUNnWUVBNkwxRVZ0ZEl1SFBXcmhKQU16TVJRVmpURWtLVGFnWVJ2dG9kWnV3Q2ZyUkNsVW1ibVlJVgp2QjdjMW11SFIxZDk2OE00R3E0M2NLL1luNkJyR3FlNG9PcHFUTTA4YzFUTDhnZ2dkQldDazdaOHUxejlhazlwCncxRVB3N3pnVktCZkVoUCtrd25Henk1NS9UQnN6a3F4VVRUMGwrdHdBT09FSGdZbFpLRkllSGtDZ1lFQTJDSnIKUDVHZmFGVnkvVlJjT3J1N2lneGFSdHVxeTRweWdxQTdrK21HajQ3bkVaRkV3MStYM0NvMHJuOUFjVDN5dUVtbwpYMkdRQnFBWmlGYWRFTXZtT0d6a09xVWE0VGQ1Vzh6TlNBNVlOdFJ5Z241NWNTNSs3Zm5jbEdwNktubEVDYkxqClltc0pXUjkxMS9PNkNnc2FFNDV5cUg1Wnd5TlloZjk0WE1saVV6c0NnWUJ2RTVXQUZMTkNSUmJhY1I5dTBCcVcKSTN4cEpKa1NhdDhoUlJ2dk9RaGZ6RXhTejVTUmlRSXlqRkE5allnOHhrYjB0SEVjV3JWZTlLM2dVVUdNc1N0dQpzVElXZ1lVdVRmUWdDVHpqNmpndG8xU1lYMk1hejlmY1BkM1dQMWlaU3dqVXFmSS8zdFNob0w3YjFiYTRKZkhHCm5nMTJUQWxpZ3pOVTJQNFRydDNWa1FLQmdFTGVRemdqb2FIeDdlV2FsLzVEM3IzVEhJc1hvenZkMVplOFl6SmIKNlptNHFKeXl5UWQ1Sjg2aDhES2NoQitFL3ZjdE1yNXZ2Tk9QN05aVmxicUFtdldTR3ZwWjRuc1RZcVNZTkZxNgp0V2doU2x3OUxPMXJhVEhQUUFOYS9manVFN0s4ZWNVVlFJc21SSnRQZUp0cTIrSjVDOWc5WHlBVWEycnBveDl4CjNzM0pBb0dBSk1NKzkwVU9ubXdEQis0TVQ5WGw1dlZFZ1cyNVMzM2paZ2d2UnlHQm54NmhvNzhOMDRYV0EyNUIKZkVaVllUbTNDSWo2UGZERHQ3L09QOG9KaTJGbVBvRTBoZ0lybW1OdWJuTjkwRzl4ZmNBTnJ5bVlJSVp3MkNWVQpSQUxQSzdzdXhEdE1CaHhZZXlJMlFmSVpUNkN2S3k1ZW9UMGNQRzFqd2ZOK2RjVXVsNUk9Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==`

var (
	svcGroupVersionResource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
)

func TestClient(t *testing.T) {
	rest.SetLogLevel(0,0)
	clientConfig, err := clientcmd.NewClientConfigFromBytes([]byte(kubeConfig))
	if err != nil {
		t.Fatal(err)
	}
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		t.Fatal(err)
	}

	dynamicInterface, err := NewClient(restConfig)
	if err != nil {
		t.Fatal(err)
	}

	resourceInterface := dynamicInterface.Resource(svcGroupVersionResource).Namespace("default")

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
	clientConfig, err := clientcmd.NewClientConfigFromBytes([]byte(kubeConfig))
	if err != nil {
		t.Fatal(err)
	}
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		t.Fatal(err)
	}

	svcStr := `{"kind":"Service","apiVersion":"v1","metadata":{"name":"wjf-test","namespace":"default","selfLink":"/api/v1/namespaces/default/services/wjf-test","uid":"70643289-3871-48c6-aa8c-fe5535dbebb3","resourceVersion":"8011981","creationTimestamp":"2021-12-29T08:34:31Z"},"spec":{"ports":[{"name":"port-1","protocol":"TCP","port":9090,"targetPort":9090}],"selector":{"app":"wjf-test"},"clusterIP":"20.111.42.59","type":"ClusterIP","sessionAffinity":"None"},"status":{"loadBalancer":{}}}`

	client, err := NewFakeClient(restConfig, fake.CreateHTTPClient(func(request *http.Request) (response *http.Response, e error) {
		switch path, method := request.URL.Path, request.Method; {
		case path == "/api/v1/namespaces/default/services/wjf-test" && method == "GET":
			return &http.Response{StatusCode: http.StatusOK, Header: rest.DefaultHeader(), Body: rest.StringBody(svcStr)}, nil
		default:
			panic("unexpected request")
		}
	}))
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
