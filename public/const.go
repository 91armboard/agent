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

	CONFIG_FILE_PATH      = "/etc/smartshop_go/config/config.ini"
	CONFIG_JSON_FILE_PATH = "/etc/smartshop_go/config/config.json"

	DEFAULT_SN       = "SN99999"
	DEFAULT_MODEL    = "AGENT-MT7621"
	DEFAULT_A_IP     = "192.168.16.8"
	DEFAULT_A_PORT   = "5000"
	DEFAULT_B_IP     = "192.168.188.123"
	DEFAULT_B_PORT   = "5000"
	DEFAULT_SERIAL1  = "/dev/ttyS1"
	DEFAULT_SERIAL2  = "/dev/ttyS2"
	DEFAULT_BAUDRATE = "115200"

	DEFAULT_MQTT_HOST     = "mqtt.shop.ijooz.sg"
	DEFAULT_MQTT_PORT     = "1883"
	DEFAULT_MQTT_USERNAME = "agent"
	DEFAULT_MQTT_PASSWORD = "ZdqRI5U9ZqPmNguG"

	API_HOST = "device.shop.ijooz.sg"
	API_AUTH = "BgdLPcVjGtJccE8u6DrGF9ZQiSuMFmzX"

	VERSION = "1.20260715.E"
)

type ApiReturn struct {
	Code int    `json:"code"`
	Info string `json:"info"`
	Data string `json:"data,omitempty"`
}
