package http

import (
	"agent/public"

	"github.com/kataras/iris/v12"
)

func NotFound(ctx iris.Context) {
	ctx.JSON(iris.Map{
		"code": 404,
		"info": "NOT_FOUND",
	})
}

func Error(ctx iris.Context, err error) {
	ctx.JSON(iris.Map{
		"code": 505,
		"info": "ERROR",
		"data": err.Error(),
	})
}

func NotAuth(ctx iris.Context) {
	ctx.JSON(iris.Map{
		"code": 909,
		"info": "NOT_AUTH",
	})
}

func CheckAuth(ctx iris.Context) bool {
	auth := ctx.GetHeader("Auth")
	if auth != string(public.HTTP_AUTH) {
		return false
	} else {
		return true
	}
}
