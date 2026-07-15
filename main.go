package main

import (
	"flag"
	//"github.com/tarm/goserial"
	"fmt"

	"agent/device"
	"agent/public"
	"agent/service"
)

var (
	conFile = flag.String("configfile", "/config/config.ini", "config file")
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("捕获异常:", err)
		}
	}()
	public.CheckSdCardMounted()
	public.GetTmpPath()
	// public.OpenDB()   // 启动不加载数据库
	public.InitConfig()
	device.TestAllCameras()
	server := fmt.Sprintf("tcp://%s:%s", public.MQTT_HOST, public.MQTT_PORT)
	topic1 := fmt.Sprintf("/IGO/DEVICE/%s", public.Config["SN"])
	topic2 := fmt.Sprintf("/IGO/MODELS/%s", public.Config["MODEL"])
	go service.CmdStart()
	//go service.StartActivityAftReboot()
	go service.ActivityStart("2")
	go service.UploadStart()
	// go service.HttpStart()
	go service.StatusStart()
	service.MqttStart(server, public.Config["SN"], public.MQTT_USERNAME, public.MQTT_PASSWORD, topic1, topic2, 2)
	if public.DB != nil {
		defer public.DB.Close()
	}
}
