package main

import (
	"fmt"

	"agent/device"
	"agent/public"
	"agent/service"
)

func main() {
	fmt.Println("ss_main version:", public.VERSION)

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("main recover:", err)
		}
	}()

	device.CheckSdCardMounted()
	public.InitConfig()

	server := fmt.Sprintf("tcp://%s:%s", public.AppConfig.MQTT.Host, public.AppConfig.MQTT.Port)
	topic1 := fmt.Sprintf("/IGO/DEVICE/%s", public.AppConfig.Common.SN)
	topic2 := fmt.Sprintf("/IGO/MODELS/%s", public.AppConfig.Common.Model)

	go device.SerialBridgeStart()
	go service.CmdStart()
	go service.TCPClientStart()
	go service.UDPModeStart()

	service.MqttStart(server, public.AppConfig.Common.SN, public.AppConfig.MQTT.Username, public.AppConfig.MQTT.Password, topic1, topic2, 2)
}
