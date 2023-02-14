package file

import (
	"fmt"
	"testing"
)

func TestInitClient(t *testing.T) {
	client, err := InitCloudClient(&CloudVendors{ServerType: ServerTypeAliyun, BucketName: "", Endpoint: "", Ak: "", Sk: ""})
	if nil != err {
		panic(err)
	}
	fmt.Println("已初始化client---", client)
}
