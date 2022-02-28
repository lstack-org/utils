
使用
---
```go
import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lstack-org/utils/pkg/catalog/component"
)

func CatalogComponentEdit(ctx *gin.Context) {
	component.NewCatalogHandler(ctx, &C{})
}


type C struct {
	
}

func (c *C) Edit(arg component.EditArg) error {
	fmt.Println(arg)
	return nil
}

func (c *C) Del(arg component.DeleteArg) error {
	fmt.Println(arg)
	return nil
}
```


说明：
1. 实现CatalogInterface接口即可
2. 返回错误（代表失败）时，catalog将会在一定时间后重复调用，直到返回nil（代表成功）
3. 在catalog规定的超时次数后，尽管仍未成功，catalog也将不再通知