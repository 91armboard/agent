package http

import (
	"agent/device"

	"github.com/kataras/iris/v12"
)

func DoorGetCurStatus(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	isClose := device.IsDoorClosed
	ctx.JSON(iris.Map{
		"code": "1",
		"info": "SUCCESS",
		"data": map[string]bool{
			"is_close": isClose,
		},
	})
}
