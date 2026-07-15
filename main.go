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

	server := fmt.Sprintf("tcp://%s:%s", public.Config["MQTT_HOST"], public.Config["MQTT_PORT"])
	topic1 := fmt.Sprintf("/IGO/DEVICE/%s", public.Config["SN"])
	topic2 := fmt.Sprintf("/IGO/MODELS/%s", public.Config["MODEL"])

	go service.SerialBridgeStart()
	go service.CmdStart()

	service.MqttStart(server, public.Config["SN"], public.Config["MQTT_USERNAME"], public.Config["MQTT_PASSWORD"], topic1, topic2, 2)
}
