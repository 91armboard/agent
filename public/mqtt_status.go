package public

import "fmt"

func SendMqttStatus(cType, cAction, cData, cId string) {
	if cType == TYPE_DEVICE {
		ChMqtt <- fmt.Sprintf("%s:%s:%s", CHANNEL_TYPE_MQTT, TOPIC_STATUS_DEVICE+Config["SN"], TYPE_DEVICE+cAction+cData)
		return
	}
	if cType == TYPE_CMD {
		ChMqtt <- fmt.Sprintf("%s:%s:%s", CHANNEL_TYPE_MQTT, TOPIC_STATUS_CMD+Config["SN"], TYPE_CMD+cAction+cData)
		return
	}
	ChMqtt <- fmt.Sprintf("%s:%s:%s", CHANNEL_TYPE_MQTT, TOPIC_STATUS_OTHER, TYPE_OTHER+cAction+cData)
}
