package file

import (
	"fmt"
	"log"
	"testing"
)

const (
	obsType       = "huaweiyun"
	obsBucketName = ""
	obsEndpoint   = ""
	obsAk         = ""
	obsSk         = ""
)

// CreateObsClient 创建obs的客户端
func CreateObsClient() *obsClientImpl {
	cloudVendors := CloudVendors{ServerType: obsType, BucketName: obsBucketName,
		Endpoint: obsEndpoint, Ak: obsAk, Sk: obsSk}
	client, err := newObsClient(&cloudVendors)
	if nil != err {
		panic(err)
	}
	return client
}

func TestCreateObsClient(t *testing.T) {
	fmt.Println(CreateObsClient())
}

// TestObsDeleteFiles 批量删除文件
func TestObsDeleteFiles(t *testing.T) {
	client := CreateObsClient()
	defer client.Close()
	err := client.DeleteFiles([]string{"tmp/123.xml"})
	if nil != err {
		panic(err)
	}
	log.Println("已成功删除")
}

func TestObsDownloadFile(t *testing.T) {
	client := CreateObsClient()
	defer client.Close()
	fileName := "002-eslint-12-06-08-39-08.xml"
	localFile := "/eslint.xml"
	_, err := client.DownloadFile(fileName, localFile)
	if nil != err {
		panic(err)
	}
	log.Println("已成功下载")
}

func TestObsUploadFile(t *testing.T) {
	client := CreateObsClient()
	defer client.Close()
	var data []byte
	err := client.UploadFile("dev/idp-dataservice-go/gitlog/123.json", data)
	if nil != err {
		panic(err)
	}
	log.Println("已成功上传")
}
