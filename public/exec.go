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

	if err := loadJSONConfig(CONFIG_JSON_FILE_PATH); err == nil {
		alog.Log.Println("InitConfig: loaded JSON config:", CONFIG_JSON_FILE_PATH)
	} else {
		alog.Log.Println("InitConfig: JSON config not loaded:", CONFIG_JSON_FILE_PATH, err)
		loadINIConfig(CONFIG_FILE_PATH)
	}

	IsPubDualDoorMod = false
	alog.Log.Println("InitConfig loaded:", Config["MODEL"], Config["SN"])
	CreateSnFile(false, Config["SN"])
}

type AgentConfigFile struct {
	Common  CommonConfig  `json:"common"`
	Network NetworkConfig `json:"network"`
	Serial  SerialConfig  `json:"serial"`
	MQTT    MQTTConfig    `json:"mqtt"`
}

type CommonConfig struct {
	Model string `json:"model"`
	SN    string `json:"sn"`
}

type NetworkConfig struct {
	AIP   string `json:"a_ip"`
	APort string `json:"a_port"`
	BIP   string `json:"b_ip"`
	BPort string `json:"b_port"`
}

type SerialConfig struct {
	Serial1  string `json:"serial1"`
	Serial2  string `json:"serial2"`
	BaudRate string `json:"baudrate"`
}

type MQTTConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func loadJSONConfig(fileName string) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	var cfg AgentConfigFile
	if err := json.Unmarshal(content, &cfg); err != nil {
		return err
	}

	setConfigValue("MODEL", cfg.Common.Model)
	setConfigValue("SN", cfg.Common.SN)
	setConfigValue("A_IP", cfg.Network.AIP)
	setConfigValue("A_PORT", cfg.Network.APort)
	setConfigValue("B_IP", cfg.Network.BIP)
	setConfigValue("B_PORT", cfg.Network.BPort)
	setConfigValue("SERIAL1", cfg.Serial.Serial1)
	setConfigValue("SERIAL2", cfg.Serial.Serial2)
	setConfigValue("BAUDRATE", cfg.Serial.BaudRate)
	setConfigValue("MQTT_HOST", cfg.MQTT.Host)
	setConfigValue("MQTT_PORT", cfg.MQTT.Port)
	setConfigValue("MQTT_USERNAME", cfg.MQTT.Username)
	setConfigValue("MQTT_PASSWORD", cfg.MQTT.Password)
	return nil
}

func loadINIConfig(fileName string) {
	cfg, err := config.ReadDefault(fileName)
	if err != nil {
		alog.Log.Println("InitConfig: using defaults, INI config not loaded:", fileName, err)
		return
	}

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
	alog.Log.Println("InitConfig: loaded INI config:", fileName)
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
	if err == nil {
		setConfigValue(key, value)
	}
}

func setConfigValue(key string, value string) {
	value = strings.TrimSpace(value)
	if value != "" {
		Config[key] = value
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
