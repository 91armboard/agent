package public

import (
	alog "agent/logger"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/larspensjo/config"
)

type AgentConfig struct {
	Common  CommonConfig  `json:"common"`
	Network NetworkConfig `json:"network"`
	Serial  SerialConfig  `json:"serial"`
	MQTT    MQTTConfig    `json:"mqtt"`
}

type CommonConfig struct {
	Model       string `json:"model"`
	SN          string `json:"sn"`
	CameraType  string `json:"camera_type,omitempty"`
	CameraCount string `json:"camera_count,omitempty"`
	LockType    string `json:"lock_type,omitempty"`
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
	Password string `json:"-"`
}

func init() {
	ChMqtt = make(chan string)
	ChActivity = make(chan string)
	ChUpload = make(chan string)
	ChCmd = make(chan string)
}

func InitConfig() {
	AppConfig = defaultAgentConfig()
	if !loadINIConfig(CONFIG_FILE_PATH, &AppConfig) {
		loadINIConfig(CONFIG_LOCAL_FILE_PATH, &AppConfig)
	}
	Config = AppConfig.ToMap()

	IsPubDualDoorMod = false
	alog.Log.Println("Config init done: ok", AppConfig.Common.Model, AppConfig.Common.SN)
	CreateSnFile(false, AppConfig.Common.SN)
}

func defaultAgentConfig() AgentConfig {
	return AgentConfig{
		Common: CommonConfig{
			Model:       DEFAULT_MODEL,
			SN:          DEFAULT_SN,
			CameraType:  "haha",
			CameraCount: "2",
			LockType:    "haha",
		},
		Network: NetworkConfig{
			AIP:   DEFAULT_A_IP,
			APort: DEFAULT_A_PORT,
			BIP:   DEFAULT_B_IP,
			BPort: DEFAULT_B_PORT,
		},
		Serial: SerialConfig{
			Serial1:  DEFAULT_SERIAL1,
			Serial2:  DEFAULT_SERIAL2,
			BaudRate: DEFAULT_BAUDRATE,
		},
		MQTT: MQTTConfig{
			Host:     DEFAULT_MQTT_HOST,
			Port:     DEFAULT_MQTT_PORT,
			Username: DEFAULT_MQTT_USERNAME,
			Password: DEFAULT_MQTT_PASSWORD,
		},
	}
}

func loadINIConfig(fileName string, cfgOut *AgentConfig) bool {
	cfg, err := config.ReadDefault(fileName)
	if err != nil {
		alog.Log.Println("Config load done: fail", fileName, err)
		return false
	}

	setINIString(cfg, "common", "model", &cfgOut.Common.Model)
	setINIString(cfg, "common", "sn", &cfgOut.Common.SN)
	setINIString(cfg, "common", "camera_type", &cfgOut.Common.CameraType)
	setINIString(cfg, "common", "camera_count", &cfgOut.Common.CameraCount)
	setINIString(cfg, "common", "lock_type", &cfgOut.Common.LockType)

	setINIString(cfg, "network", "a_ip", &cfgOut.Network.AIP)
	setINIString(cfg, "network", "a_port", &cfgOut.Network.APort)
	setINIString(cfg, "network", "b_ip", &cfgOut.Network.BIP)
	setINIString(cfg, "network", "b_port", &cfgOut.Network.BPort)

	setINIString(cfg, "serial", "serial1", &cfgOut.Serial.Serial1)
	setINIString(cfg, "serial", "serial2", &cfgOut.Serial.Serial2)
	setINIString(cfg, "serial", "baudrate", &cfgOut.Serial.BaudRate)

	setINIString(cfg, "mqtt", "host", &cfgOut.MQTT.Host)
	setINIString(cfg, "mqtt", "port", &cfgOut.MQTT.Port)
	setINIString(cfg, "mqtt", "username", &cfgOut.MQTT.Username)
	setINIString(cfg, "mqtt", "password", &cfgOut.MQTT.Password)
	alog.Log.Println("Config load done: ok", fileName)
	return true
}

func setINIString(cfg *config.Config, section string, option string, target *string) {
	value, err := cfg.String(section, option)
	if err == nil {
		value = strings.TrimSpace(value)
		if value != "" {
			*target = value
		}
	}
}

func (cfg AgentConfig) ToMap() map[string]string {
	return map[string]string{
		"SN":            cfg.Common.SN,
		"MODEL":         cfg.Common.Model,
		"CAMERA_TYPE":   cfg.Common.CameraType,
		"CAMERA_COUNT":  cfg.Common.CameraCount,
		"LOCK_TYPE":     cfg.Common.LockType,
		"A_IP":          cfg.Network.AIP,
		"A_PORT":        cfg.Network.APort,
		"B_IP":          cfg.Network.BIP,
		"B_PORT":        cfg.Network.BPort,
		"SERIAL1":       cfg.Serial.Serial1,
		"SERIAL2":       cfg.Serial.Serial2,
		"BAUDRATE":      cfg.Serial.BaudRate,
		"MQTT_HOST":     cfg.MQTT.Host,
		"MQTT_PORT":     cfg.MQTT.Port,
		"MQTT_USERNAME": cfg.MQTT.Username,
		"MQTT_PASSWORD": cfg.MQTT.Password,
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
		alog.Log.Println("SN file init pending: missing /www/sn.json")
	}

	if again {
		data := map[string]string{"sn": strsn}
		content, err := json.Marshal(data)
		if err != nil {
			alog.Log.Println("SN file init done: fail json", err)
			return
		}

		err = os.WriteFile("/www/sn.json", content, 0644)
		if err != nil {
			alog.Log.Println("SN file init done: fail write", err)
			return
		}
		alog.Log.Println("SN file init done: ok", strsn)
		return
	}
	alog.Log.Println("SN file init done: ok existing")
}
