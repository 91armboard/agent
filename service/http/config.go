package http

import (
	"github.com/kataras/iris/v12"
	"github.com/larspensjo/config"
)

func ConfigGet(ctx iris.Context) {
	if !CheckAuth(ctx) {
		NotAuth(ctx)
		return
	}
	section := ctx.FormValueDefault("section", "common")
	sn := ctx.FormValueDefault("option", "sn")
	fileName := "./config/config.ini"
	cfg, err := config.ReadDefault(fileName)
	if err != nil {
		Error(ctx, err)
		return
	}

	v, err := cfg.String(section, sn)
	if err != nil {
		Error(ctx, err)
		return
	}
	ctx.JSON(iris.Map{
		"code": "1",
		"info": "SUCCESS",
		"data": v,
	})
}
