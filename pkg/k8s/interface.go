package k8s

import (
	"context"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type Interface interface {
	//YamlsApply 实现: kubectl apply -f
	YamlsApply(ctx context.Context, reader io.Reader) error
	//YamlsDelete 实现: kubectl delete -f
	YamlsDelete(ctx context.Context, reader io.Reader) error
	Resource(resource schema.GroupVersionResource) NamespaceableResourceInterface
}

//ResourceInterface k8s资源操作接口
//以下rcv参数都用于接收k8s返回的数据
type ResourceInterface interface {
	//Create 用于创建k8s资源，body 不能为nil，rcv可以为nil
	Create(body, rcv interface{}, options metav1.CreateOptions) error
	//Update 用于更新k8s资源，body 不能为nil，rcv可以为nil
	Update(body, rcv interface{}, options metav1.UpdateOptions) error
	//Delete 用于删除指定k8s资源，name 不能为空，options可以为nil
	Delete(name string, options *metav1.DeleteOptions) error
	//DeleteCollection 用于批量删除k8s资源，options可以为nil
	DeleteCollection(listOptions metav1.ListOptions, options *metav1.DeleteOptions) error
	//Get 用于获取指定k8s资源，rvc可以为nil
	Get(name string, rcv interface{}, options metav1.GetOptions) error
	//List 用于获取k8s资源列表，rvc可以为nil
	List(rcv interface{}, opts metav1.ListOptions) error
	//Patch 用于patch指定k8s资源,name，pt，body不能为空，rvc可以为nik
	Patch(name string, pt types.PatchType, body, rcv interface{}, options metav1.PatchOptions) error
	//Apply 用于创建或更新资源，当资源存在时，执行更新，不存在时，执行创建,body不能为空，rvc可以为空
	Apply(body, rcv interface{}, applyCheckFncs ...ApplyCheckFnc) error
	//CreateIfNotExist 用于创建指定资源，当资源已存在时，直接返回该资源，存到rcv中，不存在时，会执行创建，创建结果也会存到rcv中
	CreateIfNotExist(body, rcv interface{}) error
}

type NamespaceableResourceInterface interface {
	Transform(...ReqTransformFn) NamespaceableResourceInterface
	Namespace(string) ResourceInterface
	ResourceInterface
}

type ApplyCheckFnc func(obj metav1.Object) error
