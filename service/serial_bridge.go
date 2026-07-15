package service

import (
	alog "agent/logger"
	"agent/public"
	"runtime"
	"strconv"
	"time"

	"github.com/tarm/serial"
)

func SerialBridgeStart() {
	sourceName, targetName := serialBridgePorts()
	baudRate := serialBaudRate()
	source, err := openSerialPort(sourceName, baudRate)
	if err != nil {
		alog.Log.Println("SerialBridge open source error:", sourceName, err)
		return
	}
	target, err := openSerialPort(targetName, baudRate)
	if err != nil {
		alog.Log.Println("SerialBridge open target error:", targetName, err)
		return
	}

	alog.Log.Println("SerialBridge started:", sourceName, "->", targetName, "baudrate:", baudRate)
	buf := make([]byte, 256)
	for {
		n, err := source.Read(buf)
		if err != nil {
			alog.Log.Println("SerialBridge read error:", err)
			time.Sleep(time.Second)
			continue
		}
		if n == 0 {
			continue
		}
		if _, err := target.Write(buf[:n]); err != nil {
			alog.Log.Println("SerialBridge write error:", err)
		}
	}
}

func openSerialPort(name string, baudRate int) (*serial.Port, error) {
	return serial.OpenPort(&serial.Config{
		Name:        name,
		Baud:        baudRate,
		ReadTimeout: time.Second,
	})
}

func serialBridgePorts() (string, string) {
	if runtime.GOOS == "windows" {
		return public.AppConfig.Serial.Serial1Win, public.AppConfig.Serial.Serial2Win
	}
	return public.AppConfig.Serial.Serial1, public.AppConfig.Serial.Serial2
}

func serialBaudRate() int {
	baudRate, err := strconv.Atoi(public.AppConfig.Serial.BaudRate)
	if err != nil || baudRate <= 0 {
		return 115200
	}
	return baudRate
}
