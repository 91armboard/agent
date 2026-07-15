package public

const (
	TYPE_DEVICE = "02"
	TYPE_CMD    = "04"
	TYPE_OTHER  = "05"

	ACTION_STARTRUN = "90"
	ACTION_ERROR    = "99"

	ERROR_PARAM = "800"

	CMD_GET_CONFIG  = "01"
	CMD_RUN_SHELL   = "04"
	CMD_GET_VERSION = "05"
	CMD_DOWNLOAD    = "12"
	CMD_UPGRADE     = "13"
	CMD_WGET_DNLOAD = "18"

	TOPIC_STATUS_DEVICE = "/IGO/STATUS/DEVICE/"
	TOPIC_STATUS_CMD    = "/IGO/STATUS/CMD/"
	TOPIC_STATUS_OTHER  = "/IGO/STATUS/OTHER/"

	CHANNEL_TYPE_MQTT = "MQTT"

	API_HOST      = "device.shop.ijooz.sg"
	API_AUTH      = "BgdLPcVjGtJccE8u6DrGF9ZQiSuMFmzX"
	MQTT_HOST     = "mqtt.shop.ijooz.sg"
	MQTT_PORT     = "1883"
	MQTT_USERNAME = "agent"
	MQTT_PASSWORD = "ZdqRI5U9ZqPmNguG"

	VERSION = "1.20260715.C"
)

type ApiReturn struct {
	Code int    `json:"code"`
	Info string `json:"info"`
	Data string `json:"data,omitempty"`
}
