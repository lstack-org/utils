package file

import (
	"fmt"
	"log"
	"testing"
)

const (
	ossType       = "aliyun"
	ossBucketName = ""
	ossEndpoint   = ""
	ossAk         = ""
	ossSk         = ""
)

// CreateOssClient 创建oss的客户端
func CreateOssClient() *ossClientImpl {
	cloudVendors := CloudVendors{ServerType: ossType, BucketName: ossBucketName,
		Endpoint: ossEndpoint, Ak: ossAk, Sk: ossSk}
	client, err := newOssClient(&cloudVendors)
	if nil != err {
		panic(err)
	}
	return client
}

func TestCreateOssClient(t *testing.T) {
	fmt.Println(CreateOssClient())
}

// TestOssDeleteFiles 批量删除文件
func TestOssDeleteFiles(t *testing.T) {
	client := CreateOssClient()
	defer client.Close()
	err := client.DeleteFiles([]string{"123.xml"})
	if err != nil {
		panic(err)
	}
	log.Println("已成功删除")
}

func TestOssDownloadFile(t *testing.T) {
	client := CreateOssClient()
	defer client.Close()
	fileName := "postman/light.json"
	localFile := "/postmannewman.json"
	_, err := client.DownloadFile(fileName, localFile)
	if nil != err {
		panic(err)
	}
	log.Println("已成功下载")
}

func TestOssUploadFile(t *testing.T) {
	client := CreateOssClient()
	defer client.Close()
	var data []byte
	err := client.UploadFile("gitlog/123.json", data)
	if nil != err {
		panic(err)
	}
	log.Println("已成功上传")
}
