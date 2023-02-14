package file

import (
	"errors"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"k8s.io/klog/v2"
)

var _ Client = &obsClientImpl{}

// obsClientImpl obs客户端
type obsClientImpl struct {
	ObsClient  *obs.ObsClient
	BucketName string
}

// NewObsClient
/**
 * 功能描述：初始化obs客户端
 * @author KangXu
 * @param cloudVendors 云服务商信息
 * @return *obsClientImpl, error
 */
func NewObsClient(cloudVendors *CloudVendors) (*obsClientImpl, error) {
	// 创建ObsClient结构体
	client, err := obs.New(cloudVendors.Ak, cloudVendors.Sk, cloudVendors.Endpoint)
	if err != nil {
		klog.Error(err)
		return &obsClientImpl{}, err
	}
	return &obsClientImpl{ObsClient: client, BucketName: cloudVendors.BucketName}, nil
}

// Close 关闭obs的连接
func (obsClient *obsClientImpl) Close() {
	if obsClient.ObsClient != nil {
		obsClient.ObsClient.Close()
	}
}

// DownloadFile https://support.huaweicloud.com/sdk-go-devg-obs/obs_23_0509.html#section4
/**
 * 功能描述：obs下载到本地文件
 * @author KangXu
 * @param fileName 文件名称
 * @param localFile 本地存储路径
 * @return string, error
 */
func (obsClient *obsClientImpl) DownloadFile(fileName, localFile string) (string, error) {
	// 使用访问OBS
	input := &obs.DownloadFileInput{}
	input.Bucket = obsClient.BucketName
	input.Key = fileName
	// 下载对象的本地文件全路径
	input.DownloadFile = localFile
	// 开启断点续传模式
	input.EnableCheckpoint = true
	// 指定分段大小为9MB
	input.PartSize = 9 * 1024 * 1024
	// 指定分段下载时的最大并发数
	input.TaskNum = 5
	_, downErr := obsClient.ObsClient.DownloadFile(input)
	if obsError, ok := downErr.(obs.ObsError); ok {
		klog.Error("Code:%s\n", obsError.Code)
		klog.Error("Message:%s\n", obsError.Message)
		if obsError.Code == "404" {
			return "", errors.New("404")
		}
		return "", obsError
	}
	return localFile, nil
}

// CreateSignedUrl https://support.huaweicloud.com/sdk-go-devg-obs/obs_33_0601.html#section4
/**
 * 功能描述：创建obs的带授权的url
 * @author KangXu
 * @param fileName 文件名称
 * @param expires 过期时间
 * @return string, error
 */
func (obsClient *obsClientImpl) CreateSignedUrl(fileName string, expires int) (string, error) {
	// 生成下载对象的带授权信息的URL
	getObjectInput := &obs.CreateSignedUrlInput{Bucket: obsClient.BucketName, Key: fileName,
		Method: obs.HttpMethodGet, Expires: expires}
	getObjectOutput, err := obsClient.ObsClient.CreateSignedUrl(getObjectInput)
	if err != nil {
		return "", err
	}
	return getObjectOutput.SignedUrl, nil
}

// DeleteFiles https://support.huaweicloud.com/sdk-go-devg-obs/obs_33_0507.html#section5
/**
 * 功能描述：批量删除obs指定桶中的多个文件
 * @author KangXu
 * @param fileNames 要删除的文件名称的数组
 * @return error
 */
func (obsClient *obsClientImpl) DeleteFiles(fileNames []string) error {
	var objects []obs.ObjectToDelete
	for _, fileName := range fileNames {
		objects = append(objects, obs.ObjectToDelete{Key: fileName})
	}
	input := &obs.DeleteObjectsInput{Bucket: obsClient.BucketName, Objects: objects}
	output, err := obsClient.ObsClient.DeleteObjects(input)
	if err != nil {
		if obsError, ok := err.(obs.ObsError); ok {
			klog.Error(obsError.Code)
			klog.Error(obsError.Message)
		}
	} else {
		deletes := make([]string, 0)
		for _, deleted := range output.Deleteds {
			deletes = append(deletes, deleted.Key)
		}
		if len(deletes) > 0 {
			klog.Info("已成功删除的文件：", deletes)
		}
		errors := make([]string, 0)
		for _, unDeleted := range output.Errors {
			errors = append(errors, unDeleted.Key)
		}
		if len(errors) > 0 {
			klog.Info("删除失败的文件：", errors)
		}
	}
	return err
}

// UploadFile https://support.huaweicloud.com/sdk-go-devg-obs/obs_23_0402.html#section5
/**
 * 功能描述：将指定的文件上传至obs指定桶中
 * @author KangXu
 * @param fileName 对应obs的文件名
 * @param content 文件内容
 * @return error
 */
func (obsClient *obsClientImpl) UploadFile(fileName string, content []byte) error {
	input := &obs.PutObjectInput{}
	input.Bucket = obsClient.BucketName
	input.Key = fileName
	input.Body = strings.NewReader(string(content))
	_, err := obsClient.ObsClient.PutObject(input)
	if err != nil {
		if obsError, ok := err.(obs.ObsError); ok {
			klog.Error(obsError.Code)
			klog.Error(obsError.Message)
		}
	}
	return err
}
