package public

const (
	TYPE_ACTIVITY  = "01"
	TYPE_DEVICE    = "02"
	TYPE_MODELS    = "03"
	TYPE_CMD       = "04"
	TYPE_OTHER     = "05"
	TYPE_ACTIVITY2 = "06" //Double-door:will choose another door in this case.

	ACTION_BEGIN   = "01"
	ACTION_BEGINED = "02"
	ACTION_OPENING = "03"
	ACTION_OPENED  = "04"
	ACTION_CLOSED  = "05"
	ACTION_SNAPING = "06"
	ACTION_SNAPED  = "07"

	ACTION_DOOROPENED = "08" //the door is opened by a user
	ACTION_DOORCLODED = "09" //the door is closed after a trans

	//ACTION_RFIDING   = "08"
	//ACTION_RFIDED    = "09"
	ACTION_ORCING    = "10"
	ACTION_ORCED     = "11"
	ACTION_DONE      = "12"
	ACTION_UPLOADING = "13"
	ACTION_UPLOADED  = "14"
	ACTION_NOCHANGE  = "15"
	ACTION_PAYING    = "16"
	ACTION_PAYED     = "17"
	ACTION_STORING   = "18"
	ACTION_STORED    = "19"
	ACTION_END       = "20"
	ACTION_STATUS    = "21" //report heartbeat every 5 mins
	//ACTION_HEARTBEAT  = "22"
	ACTION_STARTRUN   = "90" //7621-ss_main start running, program restarted
	ACTION_REPORT     = "92"
	ACTION_RESETAD    = "95"
	ACTION_RESETCAMRA = "96"
	ACTION_POWEROFF   = "97"
	ACTION_POWERON    = "98"
	ACTION_ERROR      = "99"

	ERROR_PARAM       = "800"
	ERROR_SERVER      = "900"
	ERROR_CLICK       = "901"
	ERROR_NETWORK     = "404"
	ERROR_TIMEOUT     = "405"
	ERROR_INUSE       = "1001"
	ERROR_SNAP_BEFORE = "1002"
	ERROR_DUPL_RECORD = "1003"
	//ERROR_RFID_BEFORE   = "1003"
	ERROR_OPEN       = "1004"
	ERROR_CLOSE      = "1005"
	ERROR_SNAP_AFTER = "1006"
	ERROR_DOOR_OPEN  = "1007"
	//ERROR_RFID_AFTER    = "1007"
	ERROR_UPLOAD1       = "1008"
	ERROR_UPLOAD2       = "1009"
	ERROR_UPLOAD3       = "1010"
	ERROR_UPLOAD4       = "1011"
	ERROR_UPLOAD5       = "1038"
	ERROR_UPLOAD6       = "1039"
	ERROR_UPLOAD7       = "1040"
	ERROR_UPLOAD8       = "1041"
	ERROR_UPLOAD9       = "1042"
	ERROR_RETRY_TIMEOUT = "1020"
	ERROR_INUSE_RETRY   = "1021"
	ERROR_IDLEN_SHORT   = "1031"
	ERROR_ID_NOTMATCH   = "1032"

	ERROR_OPEN_DB  = "1022"
	ERROR_MOUNT_SD = "1023"

	ERROR_MACHINE_DOWN1 = "1024" //continously 3 times door not opened after unlock
	ERROR_MACHINE_DOWN2 = "1025" //current lock and door status wrong
	ERROR_MACHINE_DOWN3 = "1026" //the machine encounter a low quality network connection
	ERROR_MACHINE_DOWN4 = "1027" //trans last time too long(2 hours), it maybe a faulty door or lock issue
	ERROR_MACHINE_DOWN5 = "1028" //12 hours no trans
	ERROR_MACHINE_DOWN6 = "1029" //use RAM and ram space not enough before open door
	ERROR_MACHINE_DOWN7 = "1030" //use RAM and ram space not enough during idle
	ERROR_MACHINE_DOWN8 = "1033"
	ERROR_MACHINE_DOWN9 = "1034"

	ERROR_NO_ACTIVITY_DOOROPEN = "2000"
	ERROR_NO_ACTIVITY_LOCKOPEN = "2001"

	ERROR_ACTION_HBT = "4000"

	BF_BEFORE = "B"
	BF_AFTER  = "A"

	CMD_GET_CONFIG     = "01"
	CMD_SET_CONFIG     = "02"
	CMD_GET_STATUS     = "03"
	CMD_RUN_SHELL      = "04"
	CMD_GET_VERSION    = "05"
	CMD_OPEN_LOCK      = "06" //improved: must add magic code"8302" and will lock after 10s
	CMD_CLOSE_LOCK     = "07"
	CMD_SET_LIGHT      = "08"
	CMD_RESET_CAMERA   = "09"
	CMD_CLEAR_ACTIVITY = "10"
	CMD_RESTART        = "11"
	CMD_DOWNLOAD       = "12"
	CMD_UPGRADE        = "13"
	CMD_RESET_ALL      = "14"
	CMD_GET_TEMP       = "15"
	CMD_GET_HBT        = "16"
	CMD_GET_UPLOAD     = "17"
	CMD_WGET_DNLOAD    = "18"
	CMD_GET_SIMINFO    = "19"
	CMD_GET_DATUSAG    = "20"

	TOPIC_STATUS_DEVICE   = "/IGO/STATUS/DEVICE/"
	TOPIC_STATUS_ACTIVITY = "/IGO/STATUS/ACTIVITY/"
	TOPIC_STATUS_MODELS   = "/IGO/STATUS/MODELS/"
	TOPIC_STATUS_CMD      = "/IGO/STATUS/CMD/"
	TOPIC_STATUS_OTHER    = "/IGO/STATUS/OTHER/"

	CHANNEL_TYPE_MQTT       = "MQTT"
	CHANNEL_TYPE_ACTIVITY   = "ACTIVITY"
	CHANNEL_TYPE_UPLOAD     = "UPLOAD"
	CHANNEL_TYPE_CMD        = "CMD"
	CHANNEL_TYPE_ACTIVITYEN = "ACTIVITYEN"

	API_HOST      = "device.shop.ijooz.sg"
	API_AUTH      = "BgdLPcVjGtJccE8u6DrGF9ZQiSuMFmzX"
	HTTP_AUTH     = "BgdLPcVjGtJccE8u6DrGF9ZQiSuMFmzX"
	HTTP_PORT     = "8585"
	UPGRADE_HOST  = "upgrade.shop.ijooz.sg"
	MQTT_HOST     = "mqtt.shop.ijooz.sg"
	MQTT_PORT     = "1883"
	MQTT_USERNAME = "agent"
	MQTT_PASSWORD = "ZdqRI5U9ZqPmNguG"

	/*
		API_HOST      = "device-shop.ijooz.cn"
		API_AUTH      = "BgdLPcVjGtJccE8u6DrGF9ZQiSuMFmzX"
		HTTP_AUTH     = "BgdLPcVjGtJccE8u6DrGF9ZQiSuMFmzX"
		HTTP_PORT     = "8585"
		UPGRADE_HOST  = "upgrade-shop.ijooz.cn"
		MQTT_HOST     = "mqtt-shop.ijooz.cn"
		MQTT_PORT     = "1883"
		MQTT_USERNAME = "agent"
		MQTT_PASSWORD = "ZdqRI5U9ZqPmNguG"
	*/
	//GPIO_DOOR_STATUS  = 203
	//GPIO_LOCK_STATUS  = 199
	//GPIO_LOCK_POWER   = 6
	//GPIO_LOCK_CONTROL = 2

	//GPIO_LOCK_IS_CLOSED = 1
	//GPIO_LOCK_IS_OPENED = 0
	//GPIO_DOOR_IS_CLOSED = 0
	//GPIO_DOOR_IS_OPENED = 1
	//GPIO_LOCK_POWER_ON  = 1
	//GPIO_LOCK_POWER_OFF = 0
	//GPIO_LOCK_OPEN      = 1
	//GPIO_LOCK_CLOSE     = 0

	VERSION = "1.20260715.A"
)

type ApiReturn struct {
	Code int    `json:"code"`
	Info string `json:"info"`
	Data string `json:"data,omitempty"`
}
