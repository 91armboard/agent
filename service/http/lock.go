package http

import (
	"agent/device"

	"github.com/kataras/iris/v12"
)

func LockGetCurStatus(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	// 获取当前锁的状态
	isClose := device.IsLockLocked
	ctx.JSON(iris.Map{
		"code": "1",
		"info": "SUCCESS",
		"data": map[string]bool{
			"is_close": isClose,
		},
	})
}

func LockPowerOpen(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	// 开锁电源
	err := device.OpenLockPower()
	if err != nil {
		Error(ctx, err)
	} else {
		ctx.JSON(iris.Map{
			"code": "1",
			"info": "SUCCESS",
		})
	}
}

func LockPowerClose(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	// 关锁电源
	err := device.CloseLockPower()
	if err != nil {
		Error(ctx, err)
	} else {
		ctx.JSON(iris.Map{
			"code": "1",
			"info": "SUCCESS",
		})
	}
}

func LockOpen(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	// 开锁
	err := device.OpenLock()
	if err != nil {
		Error(ctx, err)
	} else {
		ctx.JSON(iris.Map{
			"code": "1",
			"info": "SUCCESS",
		})
	}
}

func LockClose(ctx iris.Context) {
	// 关锁
	err := device.CloseLockEn()
	if err != nil {
		Error(ctx, err)
	} else {
		ctx.JSON(iris.Map{
			"code": "1",
			"info": "SUCCESS",
		})
	}
}
