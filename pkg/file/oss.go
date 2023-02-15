package file

import (
	"bytes"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"k8s.io/klog/v2"
)

var _ Client = &ossClientImpl{}

// ossClientImpl oss客户端
type ossClientImpl struct {
	OssClient  *oss.Client
	BucketName string
}

// newOssClient
/**
 * 功能描述：初始化oss客户端
 * @author KangXu
 * @param cloudVendors 云服务商信息
 * @return *ossClientImpl, error
 */
func newOssClient(cloudVendors *CloudVendors) (*ossClientImpl, error) {
	// 创建OSSClient实例
	client, err := oss.New("https://"+cloudVendors.Endpoint, cloudVendors.Ak, cloudVendors.Sk)
	if err != nil {
		klog.Error(err)
		return &ossClientImpl{}, err
	}
	return &ossClientImpl{OssClient: client, BucketName: cloudVendors.BucketName}, nil
}

// Close 关闭oss的连接
func (ossClient *ossClientImpl) Close() {
	if ossClient.OssClient != nil {
		ossClient.OssClient.HTTPClient = nil
		ossClient.OssClient.Conn = nil
		ossClient.OssClient.Config = nil
	}
}

// DownloadFile https://help.aliyun.com/document_detail/88620.html
/**
 * 功能描述：oss下载到本地文件
 * @author KangXu
 * @param fileName 文件名称
 * @param localFile 本地存储路径
 * @param fileDir 本地存储初始化文件夹
 * @return string, error
 */
func (ossClient *ossClientImpl) DownloadFile(fileName, localFile string) (string, error) {
	// 获取存储空间
	bucket, err := ossClient.OssClient.Bucket(ossClient.BucketName)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	// 下载文件到本地文件，并保存到指定的本地路径中。如果指定的本地文件存在会覆盖，不存在则新建。
	err = bucket.GetObjectToFile(fileName, localFile)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	return localFile, nil
}

// CreateSignedUrl https://help.aliyun.com/document_detail/59670.html#section-ygd-qxw-kfb
/**
 * 功能描述：创建oss的带授权的url
 * @author KangXu
 * @param fileName 文件名称
 * @param expires 过期时间
 * @return string, error
 */
func (ossClient *ossClientImpl) CreateSignedUrl(fileName string, expires int) (string, error) {
	// 生成下载对象的带授权信息的URL
	bucket, err := ossClient.OssClient.Bucket(ossClient.BucketName)
	if err != nil {
		return "", err
	}
	return bucket.SignURL(fileName, oss.HTTPGet, int64(expires))
}

// DeleteFiles https://help.aliyun.com/document_detail/88644.htm?spm=a2c4g.11186623.0.0.7168282d7ypFDo#h3-url-1
/**
 * 功能描述：批量删除oss指定桶中的多个文件
 * @author KangXu
 * @param fileNames 要删除的文件名称的数组
 * @return error
 */
func (ossClient *ossClientImpl) DeleteFiles(fileNames []string) error {
	bucket, err := ossClient.OssClient.Bucket(ossClient.BucketName)
	if err != nil {
		return err
	}
	// 填写需要删除的多个文件完整路径，文件完整路径中不能包含Bucket名称。
	delRes, err := bucket.DeleteObjects(fileNames)
	if err != nil {
		return err
	}
	if len(delRes.DeletedObjects) > 0 {
		klog.Info("已成功删除的文件：", delRes.DeletedObjects)
	}
	return nil
}

// UploadFile https://help.aliyun.com/document_detail/88601.html#section-dv9-wut-ect
/**
 * 功能描述：将指定的文件上传至oss指定桶中
 * @author KangXu
 * @param fileName 对应obs的文件名
 * @param content 文件内容
 * @return error
 */
func (ossClient *ossClientImpl) UploadFile(fileName string, content []byte) error {
	bucket, err := ossClient.OssClient.Bucket(ossClient.BucketName)
	if err != nil {
		klog.Error(err)
		return err
	}
	err = bucket.PutObject(fileName, bytes.NewReader(content))
	if err != nil {
		klog.Error(err)
	}
	return err
}
