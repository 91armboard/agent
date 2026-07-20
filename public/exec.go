package public

import (
	alog "agent/logger"
	"encoding/json"
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
	Model string `json:"model"`
	SN    string `json:"sn"`
}

type NetworkConfig struct {
	AIP             string `json:"a_ip"`
	APort           string `json:"a_port"`
	BIP             string `json:"b_ip"`
	BPort           string `json:"b_port"`
	TCPServerIP     string `json:"tcp_server_ip"`
	TCPServerPort   string `json:"tcp_server_port"`
	TCPReconnectSec string `json:"tcp_reconnect_sec"`
	UDPMode         string `json:"udp_mode"`
	UDPTargetIP     string `json:"udp_target_ip"`
	UDPTargetPort   string `json:"udp_target_port"`
	UDPLocalPort    string `json:"udp_local_port"`
	UDPSendInterval string `json:"udp_send_interval_sec"`
}

type SerialConfig struct {
	Serial1     string `json:"serial1"`
	Serial2     string `json:"serial2"`
	BaudRate    string `json:"baudrate"`
	BufferSize  string `json:"buffer_size"`
	FrameIdleMs string `json:"frame_idle_ms"`
}

type MQTTConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"-"`
}

var AppConfig AgentConfig

func InitConfig() {
	AppConfig = defaultAgentConfig()
	if !loadINIConfig(CONFIG_FILE_PATH, &AppConfig) {
		loadINIConfig(CONFIG_LOCAL_FILE_PATH, &AppConfig)
	}

	alog.Log.Println("Config init done: ok", AppConfig.Common.Model, AppConfig.Common.SN)
	CreateSnFile(false, AppConfig.Common.SN)
}

func defaultAgentConfig() AgentConfig {
	return AgentConfig{
		Common: CommonConfig{
			Model: DEFAULT_MODEL,
			SN:    DEFAULT_SN,
		},
		Network: NetworkConfig{
			AIP:             DEFAULT_A_IP,
			APort:           DEFAULT_A_PORT,
			BIP:             DEFAULT_B_IP,
			BPort:           DEFAULT_B_PORT,
			TCPServerIP:     DEFAULT_TCP_SERVER_IP,
			TCPServerPort:   DEFAULT_TCP_SERVER_PORT,
			TCPReconnectSec: DEFAULT_TCP_RECONNECT_SEC,
			UDPMode:         DEFAULT_UDP_MODE,
			UDPTargetIP:     DEFAULT_UDP_TARGET_IP,
			UDPTargetPort:   DEFAULT_UDP_TARGET_PORT,
			UDPLocalPort:    DEFAULT_UDP_LOCAL_PORT,
			UDPSendInterval: DEFAULT_UDP_SEND_INTERVAL_SEC,
		},
		Serial: SerialConfig{
			Serial1:     DEFAULT_SERIAL1,
			Serial2:     DEFAULT_SERIAL2,
			BaudRate:    DEFAULT_BAUDRATE,
			BufferSize:  DEFAULT_SERIAL_BUFFER_SIZE,
			FrameIdleMs: DEFAULT_SERIAL_FRAME_IDLE_MS,
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

	setINIString(cfg, "network", "a_ip", &cfgOut.Network.AIP)
	setINIString(cfg, "network", "a_port", &cfgOut.Network.APort)
	setINIString(cfg, "network", "b_ip", &cfgOut.Network.BIP)
	setINIString(cfg, "network", "b_port", &cfgOut.Network.BPort)
	setINIString(cfg, "network", "tcp_server_ip", &cfgOut.Network.TCPServerIP)
	setINIString(cfg, "network", "tcp_server_port", &cfgOut.Network.TCPServerPort)
	setINIString(cfg, "network", "tcp_reconnect_sec", &cfgOut.Network.TCPReconnectSec)
	setINIString(cfg, "network", "udp_mode", &cfgOut.Network.UDPMode)
	setINIString(cfg, "network", "udp_target_ip", &cfgOut.Network.UDPTargetIP)
	setINIString(cfg, "network", "udp_target_port", &cfgOut.Network.UDPTargetPort)
	setINIString(cfg, "network", "udp_local_port", &cfgOut.Network.UDPLocalPort)
	setINIString(cfg, "network", "udp_send_interval_sec", &cfgOut.Network.UDPSendInterval)

	setINIString(cfg, "serial", "serial1", &cfgOut.Serial.Serial1)
	setINIString(cfg, "serial", "serial2", &cfgOut.Serial.Serial2)
	setINIString(cfg, "serial", "baudrate", &cfgOut.Serial.BaudRate)
	setINIString(cfg, "serial", "buffer_size", &cfgOut.Serial.BufferSize)
	setINIString(cfg, "serial", "frame_idle_ms", &cfgOut.Serial.FrameIdleMs)

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

func CreateSnFile(again bool, strsn string) {
	if !fileExists("/www/sn.json") {
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
