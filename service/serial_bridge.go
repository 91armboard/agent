package service

import (
	alog "agent/logger"
	"agent/public"
	"io"
	"strconv"
	"time"

	"github.com/tarm/serial"
)

func SerialBridgeStart() {
	sourceName, targetName := serialBridgePorts()
	baudRate := serialBaudRate()
	source, err := openSerialPort(sourceName, baudRate)
	if err != nil {
		alog.Log.Println("Serial bridge init done: fail source", sourceName, err)
		return
	}
	target, err := openSerialPort(targetName, baudRate)
	if err != nil {
		alog.Log.Println("Serial bridge init done: fail target", targetName, err)
		return
	}

	alog.Log.Println("Serial bridge init done: ok", sourceName, "->", targetName, "baudrate:", baudRate)
	buf := make([]byte, 256)
	for {
		n, err := source.Read(buf)
		if n > 0 {
			if _, err := target.Write(buf[:n]); err != nil {
				alog.Log.Println("SerialBridge write error:", err)
			}
		}
		if err == nil || err == io.EOF {
			continue
		}
		alog.Log.Println("SerialBridge read error:", err)
		time.Sleep(time.Second)
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
	return public.AppConfig.Serial.Serial1, public.AppConfig.Serial.Serial2
}

func serialBaudRate() int {
	baudRate, err := strconv.Atoi(public.AppConfig.Serial.BaudRate)
	if err != nil || baudRate <= 0 {
		return 115200
	}
	return baudRate
}
