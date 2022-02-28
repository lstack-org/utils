package component

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type CatalogInterface interface {
	// Edit 修改组件
	Edit(arg EditArg) error
	// Del 删除组件
	Del(arg DeleteArg) error
}

type Notify struct {
	*gin.Context
	EditHandler
	DeleteHandler
	EditArg
	DeleteArg
}

type EditHandler func(EditArg) error
type DeleteHandler func(DeleteArg) error

func NewCatalogHandler(ctx *gin.Context, catalogInterface CatalogInterface) {
	c := &Notify{
		Context:       ctx,
		EditHandler:   catalogInterface.Edit,
		DeleteHandler: catalogInterface.Del,
	}
	err := ctx.ShouldBindJSON(c)
	if err != nil {
		c.JSON(400, Fail(err))
		return
	}
	switch ctx.Request.Method {
	case http.MethodPut:
		err = c.EditHandler(c.EditArg)
	case http.MethodDelete:
		err = c.DeleteHandler(c.DeleteArg)
	}
	if err != nil {
		c.JSON(400, Fail(err))
		return
	}
	c.JSON(200, Success())
	return
}

type Res struct {
	Status int    `json:"status"`
	ResMsg string `json:"resMsg"`
}

func Success() Res {
	return Res{Status: 200}
}

func Fail(err error) Res {
	return Res{Status: 200, ResMsg: err.Error()}
}
