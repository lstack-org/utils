package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/restmapper"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lstack-org/utils/pkg/gorun"
	local "github.com/lstack-org/utils/pkg/rest"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

var (
	deleteScheme          = runtime.NewScheme()
	parameterScheme       = runtime.NewScheme()
	deleteOptionsCodec    = serializer.NewCodecFactory(deleteScheme)
	restParameterCodec    = runtime.NewParameterCodec(parameterScheme)
	dynamicParameterCodec = runtime.NewParameterCodec(parameterScheme)
	versionV1             = schema.GroupVersion{Version: "v1"}
)

func init() {
	metav1.AddToGroupVersion(parameterScheme, versionV1)
	metav1.AddToGroupVersion(deleteScheme, versionV1)
}

func NewClient(restConfig *rest.Config, fns ...ReqTransformFn) (Interface, error) {
	return newClient(restConfig, nil, fns...)
}

func NewClientInCluster(fns ...ReqTransformFn) (Interface, error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return newClient(restConfig, nil, fns...)
}

func newClient(restConfig *rest.Config, customize *http.Client, fns ...ReqTransformFn) (Interface, error) {
	config := dynamic.ConfigFor(restConfig)
	// for serializing the options
	config.GroupVersion = &schema.GroupVersion{}
	config.APIPath = "/if-you-see-this-search-for-the-break"
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}
	if customize != nil {
		restClient.Client = customize
	} else {
		tripper := restClient.Client.Transport
		restClient.Client.Transport = local.NewLogTrace("k8s", tripper)
	}
	return &dynamicInterface{
		restConfig:          restConfig,
		restClient:          restClient,
		transformRequestFns: fns,
	}, nil
}

type ReqTransformFn func(req *rest.Request)

type dynamicInterface struct {
	restConfig          *rest.Config
	restClient          *rest.RESTClient
	transformRequestFns []ReqTransformFn
	once                sync.Once
	mapper              meta.RESTMapper
}

func (d *dynamicInterface) yamlsDo(ctx context.Context, reader io.Reader, do func(mapping *meta.RESTMapping, obj unstructured.Unstructured) error) error {
	var (
		sb       strings.Builder
		buf      = make([]byte, 1024)
		errCache error
	)

	//从reader中读取数据
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				return err
			}
			//gc
			buf = nil
			break
		}
		sb.Write(buf[:n])
	}

	d.once.Do(func() {
		discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(d.restConfig)
		apiGroups, err := restmapper.GetAPIGroupResources(discoveryClient)
		if err != nil {
			errCache = err
			return
		}
		d.mapper = restmapper.NewDiscoveryRESTMapper(apiGroups)
	})
	if errCache != nil {
		return errCache
	}

	var (
		resouces = ManifestToResouces(sb.String())
		actions  []gorun.BatchTaskAction
	)

	for index := range resouces {
		var (
			resource         = resouces[index]
			groupVersionKind = schema.FromAPIVersionAndKind(resource.GetAPIVersion(), resource.GetKind())
		)

		actions = append(actions, func(ctx gorun.BatchContext) {
			mapping, err := d.mapper.RESTMapping(schema.GroupKind{
				Group: groupVersionKind.Group,
				Kind:  groupVersionKind.Kind,
			}, groupVersionKind.Version)
			if err != nil {
				ctx.AddError(err)
				return
			}

			if mapping == nil {
				ctx.AddError(fmt.Errorf("resource type %v not found", groupVersionKind))
			} else {
				ctx.AddError(do(mapping, resource))
			}
		})
	}

	_, err := gorun.Tasks(actions...).Await(ctx)
	return err
}

func (d *dynamicInterface) YamlsDelete(ctx context.Context, reader io.Reader, dryrun ...string) error {
	return d.yamlsDo(ctx, reader, func(mapping *meta.RESTMapping, obj unstructured.Unstructured) error {
		var err error
		if mapping.Scope.Name() == meta.RESTScopeNameRoot {
			err = d.Resource(mapping.Resource).Delete(obj.GetName(), &metav1.DeleteOptions{
				DryRun: dryrun,
			})
		} else {
			err = d.Resource(mapping.Resource).Namespace(obj.GetNamespace()).Delete(obj.GetName(), &metav1.DeleteOptions{
				DryRun: dryrun,
			})
		}

		if err != nil {
			//忽略资源不存在
			if !errors.IsNotFound(err) {
				return err
			}
		}
		return nil
	})
}

func (d *dynamicInterface) YamlsApply(ctx context.Context, reader io.Reader, dryrun ...string) error {
	return d.yamlsDo(ctx, reader, func(mapping *meta.RESTMapping, obj unstructured.Unstructured) error {
		if mapping.Scope.Name() == meta.RESTScopeNameRoot {
			return d.Resource(mapping.Resource).PatchApply(&obj, nil, dryrun...)
		} else {
			return d.Resource(mapping.Resource).Namespace(obj.GetNamespace()).PatchApply(&obj, nil, dryrun...)
		}
	})
}

func (d *dynamicInterface) Resource(resource schema.GroupVersionResource) NamespaceableResourceInterface {
	return &dynamicClient{
		resource:            resource,
		restC:               d.restClient,
		transformRequestFns: d.transformRequestFns,
	}
}

var _ NamespaceableResourceInterface = &dynamicClient{}

type dynamicClient struct {
	namespace           string
	resource            schema.GroupVersionResource
	restC               *rest.RESTClient
	transformRequestFns []ReqTransformFn
}

func (d *dynamicClient) Transform(reqFns ...ReqTransformFn) NamespaceableResourceInterface {
	d.transformRequestFns = append(d.transformRequestFns, reqFns...)
	return d
}

func (d *dynamicClient) Namespace(namespace string) ResourceInterface {
	if len(namespace) == 0 {
		namespace = "default"
	}
	d.namespace = namespace
	return d
}

func (d *dynamicClient) Create(body, rcv interface{}, options v1.CreateOptions) error {
	return d.request(d.tryTransformRequest(d.restC.
		Post().
		AbsPath(d.makeURLSegments("")...).
		Body(body).
		SpecificallyVersionedParams(&options, dynamicParameterCodec, versionV1)), rcv)
}

func (d *dynamicClient) Update(body, rcv interface{}, options v1.UpdateOptions) error {
	accessor, err := meta.Accessor(body)
	if err != nil {
		return err
	}
	name := accessor.GetName()
	if len(name) == 0 {
		return fmt.Errorf("name is required")
	}
	return d.request(d.tryTransformRequest(d.restC.
		Put().
		AbsPath(d.makeURLSegments(name)...).
		Body(body).
		SpecificallyVersionedParams(&options, restParameterCodec, versionV1)), rcv)
}

func (d *dynamicClient) Delete(name string, options *v1.DeleteOptions) error {
	if len(name) == 0 {
		return fmt.Errorf("name is required")
	}
	if options == nil {
		options = &v1.DeleteOptions{}
	}
	deleteOptionsByte, err := runtime.Encode(deleteOptionsCodec.LegacyCodec(schema.GroupVersion{Version: "v1"}), options)
	if err != nil {
		return err
	}
	return d.request(d.tryTransformRequest(d.restC.
		Delete().
		AbsPath(d.makeURLSegments(name)...).
		Body(deleteOptionsByte)), nil)
}

func (d *dynamicClient) DeleteCollection(listOptions v1.ListOptions, options *v1.DeleteOptions) error {
	if options == nil {
		options = &v1.DeleteOptions{}
	}
	deleteOptionsByte, err := runtime.Encode(deleteOptionsCodec.LegacyCodec(schema.GroupVersion{Version: "v1"}), options)
	if err != nil {
		return err
	}
	return d.request(d.tryTransformRequest(d.restC.
		Delete().
		AbsPath(d.makeURLSegments("")...).
		Body(deleteOptionsByte).
		SpecificallyVersionedParams(&listOptions, dynamicParameterCodec, versionV1)), nil)
}

func (d *dynamicClient) Get(name string, rcv interface{}, options v1.GetOptions) error {
	if len(name) == 0 {
		return fmt.Errorf("name is required")
	}
	return d.request(d.tryTransformRequest(d.restC.
		Get().
		AbsPath(d.makeURLSegments(name)...).
		SpecificallyVersionedParams(&options, restParameterCodec, versionV1)), rcv)
}

func (d *dynamicClient) List(rcv interface{}, opts v1.ListOptions) error {
	return d.request(d.tryTransformRequest(d.restC.
		Get().
		AbsPath(d.makeURLSegments("")...).
		SpecificallyVersionedParams(&opts, restParameterCodec, versionV1)), rcv)
}

func (d *dynamicClient) Patch(name string, pt types.PatchType, body, rcv interface{}, options v1.PatchOptions) error {
	if len(name) == 0 {
		return fmt.Errorf("name is required")
	}
	var (
		bodyBytes []byte
		err       error
	)

	switch body.(type) {
	case []byte:
		bodyBytes = body.([]byte)
	default:
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	return d.request(d.tryTransformRequest(d.restC.
		Patch(pt).
		AbsPath(d.makeURLSegments(name)...).
		Body(bodyBytes).
		SpecificallyVersionedParams(&options, dynamicParameterCodec, versionV1)), rcv)
}

func (d *dynamicClient) PatchApply(body, rcv interface{}, dryrun ...string) error {
	return patchApply(d, body, rcv, dryrun...)
}

func (d *dynamicClient) Apply(body, rcv interface{}, dryrun ...string) error {
	bodyObj, err := meta.Accessor(body)
	if err != nil {
		return err
	}

	item, err := gorun.UntilWithTimeout(func(until gorun.Until) {
		objectMeta := core.Pod{}
		err = d.Get(bodyObj.GetName(), &objectMeta, v1.GetOptions{})
		if err != nil {
			//获取对应的资源失败
			//若不是NotFound错误，则break
			if !errors.IsNotFound(err) {
				until.ErrorBreak(err)
			} else {
				//资源不存在，退出
				until.Cancel()
			}
		} else {
			bodyObj.SetResourceVersion(objectMeta.ResourceVersion)
			//对应的资源存在，保存，用于后续判断是否要创建
			until.ItemSave(objectMeta)
			//执行update
			err = d.Update(bodyObj, rcv, v1.UpdateOptions{
				DryRun: dryrun,
			})
			if err != nil {
				//若update返回冲突，则执行重试
				if errors.IsConflict(err) {
					klog.V(5).Infof("Error : %v, Apply retry ....", err)
					until.ErrorSave(err)
				} else {
					//不是冲突错误，则break
					until.ErrorBreak(err)
				}
			} else {
				//更新成功，退出
				until.Cancel()
			}
		}
	}, 100*time.Millisecond, 30*time.Second)
	if err != nil {
		return err
	}

	//资源不存在
	if item == nil {
		return d.Create(bodyObj, rcv, v1.CreateOptions{
			DryRun: dryrun,
		})
	}

	return nil
}

func (d *dynamicClient) CreateIfNotExist(body, rcv interface{}) error {
	bodyObj, err := meta.Accessor(body)
	if err != nil {
		return err
	}
	err = d.Get(bodyObj.GetName(), rcv, v1.GetOptions{})
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}
	return d.Create(body, rcv, v1.CreateOptions{})
}

func (d *dynamicClient) tryTransformRequest(req *rest.Request) *rest.Request {
	if len(d.transformRequestFns) > 0 {
		for _, fn := range d.transformRequestFns {
			fn(req)
		}
	}
	return req
}

func (d *dynamicClient) request(request *rest.Request, rcv interface{}) error {
	result, err := request.DoRaw(context.TODO())
	if err != nil {
		return local.ErrorConvert(result, err)
	}

	if rcv != nil {
		return json.Unmarshal(result, rcv)
	}
	return nil
}

func (d *dynamicClient) makeURLSegments(name string) []string {
	var url []string
	if len(d.resource.Group) == 0 {
		url = append(url, "api")
	} else {
		url = append(url, "apis", d.resource.Group)
	}
	url = append(url, d.resource.Version)

	if len(d.namespace) > 0 {
		url = append(url, "namespaces", d.namespace)
	}
	url = append(url, d.resource.Resource)

	if len(name) > 0 {
		url = append(url, name)
	}

	return url
}
