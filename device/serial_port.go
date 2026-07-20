package device

import (
	"agent/public"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tarm/serial"
)

func openSerialPort(name string, baudRate int) (*serial.Port, error) {
	return serial.OpenPort(&serial.Config{
		Name:        name,
		Baud:        baudRate,
		ReadTimeout: time.Second,
		Size:        8,
		Parity:      serial.ParityNone,
		StopBits:    serial.Stop1,
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

func serialBufferSize() int {
	size, err := strconv.Atoi(public.AppConfig.Serial.BufferSize)
	if err != nil || size <= 0 {
		return 1024
	}
	return size
}

func serialFrameIdle() time.Duration {
	idleMs, err := strconv.Atoi(public.AppConfig.Serial.FrameIdleMs)
	if err != nil || idleMs <= 0 {
		idleMs = 50
	}
	return time.Duration(idleMs) * time.Millisecond
}

func formatSerialHex(data []byte) string {
	items := make([]string, len(data))
	for i, b := range data {
		items[i] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(items, " ")
}
