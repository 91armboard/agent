package http

import (
	"encoding/base64"
	"io/ioutil"

	"agent/device"
	"agent/public"

	"github.com/kataras/iris/v12"
)

func CameraGetCurImg(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	// 获取当前照片
	req := ctx.FormValueDefault("req", "HTTP")
	err, fileNames := device.SnapShot(req, public.BF_AFTER)
	if err == nil {
		for k, item := range fileNames {
			image, _ := ioutil.ReadFile(public.GetSdcardPath() + "/" + item)
			imageBase64 := base64.StdEncoding.EncodeToString(image)
			fileNames[k] = imageBase64
			public.DeleteFileOnDisk(item)
		}
		ctx.JSON(iris.Map{
			"code": "1",
			"info": "SUCCESS",
			"data": fileNames,
		})
	} else {
		Error(ctx, err)
	}
}

func ResetCPU(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	// 重置主板
	err := device.ResetCPU()
	if err != nil {
		Error(ctx, err)
	} else {
		ctx.JSON(iris.Map{
			"code": "1",
			"info": "SUCCESS",
		})
	}
}
