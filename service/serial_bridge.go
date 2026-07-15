package service

import (
	alog "agent/logger"
	"runtime"
	"time"

	"github.com/tarm/serial"
)

const (
	serial1Linux = "/dev/ttyS1"
	serial2Linux = "/dev/ttyS2"
	serial1Win   = "COM3"
	serial2Win   = "COM4"
)

func SerialBridgeStart() {
	sourceName, targetName := serialBridgePorts()
	source, err := openSerialPort(sourceName)
	if err != nil {
		alog.Log.Println("SerialBridge open source error:", sourceName, err)
		return
	}
	target, err := openSerialPort(targetName)
	if err != nil {
		alog.Log.Println("SerialBridge open target error:", targetName, err)
		return
	}

	alog.Log.Println("SerialBridge started:", sourceName, "->", targetName)
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

func openSerialPort(name string) (*serial.Port, error) {
	return serial.OpenPort(&serial.Config{
		Name:        name,
		Baud:        115200,
		ReadTimeout: time.Second,
	})
}

func serialBridgePorts() (string, string) {
	if runtime.GOOS == "windows" {
		return serial1Win, serial2Win
	}
	return serial1Linux, serial2Linux
}
