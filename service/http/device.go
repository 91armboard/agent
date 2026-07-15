package http

import (
	"agent/device"
	"agent/public"

	"github.com/kataras/iris/v12"
)

func DeviceGetCurStatus(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	// 获取当前设备状态
	isRunning := device.GetIsActivityRunning()
	isOnline := device.IsOnline
	activityId := device.CurActivityId
	isRunningStr := "false"
	if isRunning {
		isRunningStr = "true"
	}
	isOnlineStr := "false"
	if isOnline {
		isOnlineStr = "true"
	}
	ctx.JSON(iris.Map{
		"code": "1",
		"info": "SUCCESS",
		"data": map[string]string{
			"is_running":  isRunningStr,
			"is_online":   isOnlineStr,
			"activity_id": activityId,
		},
	})
}

func DeviceRunShell(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	cmd := ctx.FormValueDefault("cmd", "ls")
	// 获取当前设备状态
	err, output := public.ExecShell(cmd)
	if err != nil {
		Error(ctx, err)
	} else {
		ctx.JSON(iris.Map{
			"code": "1",
			"info": "SUCCESS",
			"data": output,
		})
	}
}

func DeviceGetVersion(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	// 获取当前设备状态
	ctx.JSON(iris.Map{
		"code": "1",
		"info": "SUCCESS",
		"data": public.VERSION,
	})
}

func PowerStatus(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	// 获取电源状态
	isOn := device.IsMPowerOn
	ctx.JSON(iris.Map{
		"code": "1",
		"info": "SUCCESS",
		"data": map[string]bool{
			"is_close": isOn,
		},
	})
}

func ClosePower(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	// 重启电源
	err := device.CloseCPU()
	if err != nil {
		Error(ctx, err)
	} else {
		ctx.JSON(iris.Map{
			"code": "1",
			"info": "SUCCESS",
		})
	}
}

func SetLight0(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	// 灯光调暗
	err := device.SetLight0()
	if err != nil {
		Error(ctx, err)
	} else {
		ctx.JSON(iris.Map{
			"code": "1",
			"info": "SUCCESS",
		})
	}
}

func SetLight1(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	// 灯光调暗
	err := device.SetLight1()
	if err != nil {
		Error(ctx, err)
	} else {
		ctx.JSON(iris.Map{
			"code": "1",
			"info": "SUCCESS",
		})
	}
}

func SetLight2(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	// 灯光调暗
	err := device.SetLight2()
	if err != nil {
		Error(ctx, err)
	} else {
		ctx.JSON(iris.Map{
			"code": "1",
			"info": "SUCCESS",
		})
	}
}

func SetLight3(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	// 灯光调暗
	err := device.SetLight3()
	if err != nil {
		Error(ctx, err)
	} else {
		ctx.JSON(iris.Map{
			"code": "1",
			"info": "SUCCESS",
		})
	}
}

func SetLight4(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	// 灯光调亮
	err := device.SetLight4()
	if err != nil {
		Error(ctx, err)
	} else {
		ctx.JSON(iris.Map{
			"code": "1",
			"info": "SUCCESS",
		})
	}
}
