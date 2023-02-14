package file

import (
	"encoding/json"
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
	client, err := NewObsClient(&cloudVendors)
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
	tmp := "[{\"id\":\"b903769f7ccd786a725891f8836e9d2dce2454f1\",\"branch\":\"\",\"short_id\":\"b903769f\",\"committer_name\":\"lstack\",\"committer_email\":\"lstack@xlauncher.io\",\"created_at\":\"2023-02-10T10:56:09Z\",\"message\":\"refactor(dev-V1.3.0): 修改 组件自动同步commit相关接口\\n\",\"committed_date\":\"2023-02-10T10:56:09Z\",\"parent_ids\":[\"4ff233589069ea9a003dc278ec2323497404e7b4\"],\"stats\":{\"additions\":130,\"deletions\":3,\"total\":133}},{\"id\":\"4c3e0ea2137fa71581b2619efc947483d6a1925f\",\"branch\":\"\",\"short_id\":\"4c3e0ea2\",\"committer_name\":\"wangdaoke\",\"committer_email\":\"dkwangk@isoftstone.com\",\"created_at\":\"2023-02-10T08:29:38Z\",\"message\":\"refactor(root) 提交至DealWith2分支\\n\",\"committed_date\":\"2023-02-10T08:29:38Z\",\"parent_ids\":[\"2b598f27392c36be43fd913d7d80cdcd70605033\"],\"stats\":{\"additions\":6,\"deletions\":6,\"total\":12}}]"
	data, _ := json.Marshal(tmp)
	err := client.UploadFile("dev/idp-dataservice-go/gitlog/123.json", data)
	if nil != err {
		panic(err)
	}
	log.Println("已成功上传")
}
