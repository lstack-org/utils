package file

import (
	"k8s.io/apimachinery/pkg/api/errors"
)

// CloudVendors 云厂商基本参数
type CloudVendors struct {
	// ServerType 参数描述：服务商类型
	ServerType string
	// BucketName 参数描述：桶名
	BucketName string
	// Endpoint 参数描述：节点
	Endpoint string
	// Ak 参数描述：云服务器的ak
	Ak string
	// Sk 参数描述：云服务器的sk
	Sk string
}

// Client 初始化客户端接口
type Client interface {
	DownloadFile(fileName, localFile string) (string, error)
	CreateSignedUrl(fileName string, expires int) (string, error)
	DeleteFiles(fileNames []string) error
	Close()
	UploadFile(fileName string, content []byte) error
}

const (
	ServerTypeAliyun    = "aliyun"
	ServerTypeHuaweiyun = "huaweiyun"
)

// InitCloudClient 初始化服务商客户端
func InitCloudClient(cloudVendors *CloudVendors) (client Client, err error) {
	switch cloudVendors.ServerType {
	case ServerTypeAliyun:
		client, err = NewOssClient(cloudVendors)
	case ServerTypeHuaweiyun:
		client, err = NewObsClient(cloudVendors)
	default:
		return nil, errors.NewBadRequest("unsupported client")
	}
	return
}
