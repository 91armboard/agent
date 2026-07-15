package main

import (
	"fmt"

	"agent/public"
	"agent/service"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("main recover:", err)
		}
	}()

	public.CheckSdCardMounted()
	public.InitConfig()

	server := fmt.Sprintf("tcp://%s:%s", public.AppConfig.MQTT.Host, public.AppConfig.MQTT.Port)
	topic1 := fmt.Sprintf("/IGO/DEVICE/%s", public.AppConfig.Common.SN)
	topic2 := fmt.Sprintf("/IGO/MODELS/%s", public.AppConfig.Common.Model)

	go service.SerialBridgeStart()
	go service.CmdStart()

	service.MqttStart(server, public.AppConfig.Common.SN, public.AppConfig.MQTT.Username, public.AppConfig.MQTT.Password, topic1, topic2, 2)
}
