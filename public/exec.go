package public

import (
	alog "agent/logger"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/larspensjo/config"
)

func init() {
	ChMqtt = make(chan string)
	ChActivity = make(chan string)
	ChUpload = make(chan string)
	ChCmd = make(chan string)
}

func InitConfig() {
	Config = defaultConfig()

	cfg, err := config.ReadDefault(CONFIG_FILE_PATH)
	if err != nil {
		alog.Log.Println("InitConfig: using defaults, config not loaded:", CONFIG_FILE_PATH, err)
	} else {
		readConfigValue(cfg, "common", "model", "MODEL")
		readConfigValue(cfg, "common", "sn", "SN")
		readConfigValue(cfg, "network", "a_ip", "A_IP")
		readConfigValue(cfg, "network", "a_port", "A_PORT")
		readConfigValue(cfg, "network", "b_ip", "B_IP")
		readConfigValue(cfg, "network", "b_port", "B_PORT")
		readConfigValue(cfg, "serial", "serial1", "SERIAL1")
		readConfigValue(cfg, "serial", "serial2", "SERIAL2")
		readConfigValue(cfg, "serial", "baudrate", "BAUDRATE")
		readConfigValue(cfg, "mqtt", "host", "MQTT_HOST")
		readConfigValue(cfg, "mqtt", "port", "MQTT_PORT")
		readConfigValue(cfg, "mqtt", "username", "MQTT_USERNAME")
		readConfigValue(cfg, "mqtt", "password", "MQTT_PASSWORD")
	}

	IsPubDualDoorMod = false
	alog.Log.Println("InitConfig loaded:", Config["MODEL"], Config["SN"])
	CreateSnFile(false, Config["SN"])
}

func defaultConfig() map[string]string {
	return map[string]string{
		"SN":            DEFAULT_SN,
		"MODEL":         DEFAULT_MODEL,
		"A_IP":          DEFAULT_A_IP,
		"A_PORT":        DEFAULT_A_PORT,
		"B_IP":          DEFAULT_B_IP,
		"B_PORT":        DEFAULT_B_PORT,
		"SERIAL1":       DEFAULT_SERIAL1,
		"SERIAL2":       DEFAULT_SERIAL2,
		"BAUDRATE":      DEFAULT_BAUDRATE,
		"MQTT_HOST":     DEFAULT_MQTT_HOST,
		"MQTT_PORT":     DEFAULT_MQTT_PORT,
		"MQTT_USERNAME": DEFAULT_MQTT_USERNAME,
		"MQTT_PASSWORD": DEFAULT_MQTT_PASSWORD,
	}
}

func readConfigValue(cfg *config.Config, section string, option string, key string) {
	value, err := cfg.String(section, option)
	if err == nil && strings.TrimSpace(value) != "" {
		Config[key] = strings.TrimSpace(value)
	}
}

func checkErr(err error) {
	if err != nil {
		alog.Log.Println(err)
		os.Exit(-1)
	}
}

func setFrpIni(fileName, sn string) {
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		alog.Log.Println("READ FRP INI ERROR!")
	}
	fileContent, err := io.ReadAll(f)
	temp := strings.Replace(string(fileContent), "{SN}", sn, -1)
	if os.WriteFile(fileName, []byte(temp), 0766) != nil {
		alog.Log.Println("WRITE FRP INI ERROR!")
	}
}

func CreateSnFile(again bool, strsn string) {
	if !IsExist("/www/sn.json") {
		again = true
		alog.Log.Println("CreateSnFile:exist")
	}

	if again {
		data := map[string]string{"sn": strsn}
		content, err := json.Marshal(data)
		if err != nil {
			alog.Log.Println("CreateSnFile:json", err)
			return
		}

		err = os.WriteFile("/www/sn.json", content, 0644)
		if err != nil {
			alog.Log.Println("CreateSnFile:WriteFile", err)
			return
		}
		alog.Log.Println("CreateSnFile:", strsn)
	}
}
