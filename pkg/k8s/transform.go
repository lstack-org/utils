package k8s

import (
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"reflect"
	"strings"

	"helm.sh/helm/v3/pkg/releaseutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"

	local "github.com/lstack-org/utils/pkg/rest"
)

//SkipLog 用于跳过打印本次范围k8s的日志
func SkipLog(req *rest.Request) {
	req.SetHeader(local.SkipLogHeaderKey, local.SkipLogHeaderValue)
}

//AsTable 用于指定k8s返回的数据格式为table格式
func AsTable(req *rest.Request) {
	req.SetHeader("Accept", strings.Join([]string{
		fmt.Sprintf("application/json;as=Table;v=%s;g=%s", metav1.SchemeGroupVersion.Version, metav1.GroupName),
		fmt.Sprintf("application/json;as=Table;v=%s;g=%s", metav1beta1.SchemeGroupVersion.Version, metav1beta1.GroupName),
		"application/json",
	}, ","))
}

//TableHandle 将table中的数据存入into，into必须是一个数组指针
func TableHandle(table *metav1beta1.Table, into interface{}) error {
	v := reflect.ValueOf(into)
	if v.Kind() != reflect.Ptr {
		panic("type must be ptr")
	}
	v = v.Elem()

	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		panic("type must be Array or Slice")
	}

	m := make([]map[string]interface{}, 0)
	for _, row := range table.Rows {
		r := make(map[string]interface{})
		for index, cell := range row.Cells {
			name := table.ColumnDefinitions[index].Name
			r[name] = cell
		}
		m = append(m, r)
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, into)
}

func ManifestToResouces(manifest string) []unstructured.Unstructured {
	manifests := releaseutil.SplitManifests(manifest)
	var objs []unstructured.Unstructured
	for _, manifest := range manifests {
		var u unstructured.Unstructured

		if err := yaml.Unmarshal([]byte(manifest), &u); err != nil {
			continue
		}

		if u.IsList() {
			l, err := u.ToList()
			if err != nil {
				continue
			}
			objs = append(objs, l.Items...)
			continue
		}
		objs = append(objs, u)
	}
	return objs
}
