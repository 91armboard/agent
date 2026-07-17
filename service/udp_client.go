package service

import (
	alog "agent/logger"
	"agent/public"
	"fmt"
	"net"
	"strconv"
	"time"
)

func UDPModeStart() {
	if public.AppConfig.Network.UDPMode != "enabled" {
		alog.Log.Println("UDP mode init done: disabled")
		return
	}

	targetAddress := net.JoinHostPort(public.AppConfig.Network.UDPTargetIP, public.AppConfig.Network.UDPTargetPort)
	targetUDPAddr, err := net.ResolveUDPAddr("udp", targetAddress)
	if err != nil {
		alog.Log.Println("UDP target init done: fail", targetAddress, err)
		return
	}

	localPort := parsePositiveInt(public.AppConfig.Network.UDPLocalPort, 12345)
	sendInterval := parseDurationSeconds(public.AppConfig.Network.UDPSendInterval, 3)

	for {
		conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: localPort})
		if err != nil {
			alog.Log.Println("UDP listen done: fail port:", localPort, err, "next_retry:", 5*time.Second)
			time.Sleep(5 * time.Second)
			continue
		}

		alog.Log.Println("UDP mode init done: ok local_port:", localPort, "target:", targetAddress, "interval:", sendInterval)
		runUDPTest(conn, targetUDPAddr, sendInterval)
		alog.Log.Println("UDP mode closed, next_retry:", 5*time.Second)
		time.Sleep(5 * time.Second)
	}
}

func runUDPTest(conn *net.UDPConn, target *net.UDPAddr, sendInterval time.Duration) {
	defer conn.Close()

	done := make(chan struct{})
	go readUDP(conn, done)

	seq := 0
	for {
		select {
		case <-done:
			return
		default:
		}

		payload := []byte(fmt.Sprintf("test%02d", seq))
		if _, err := conn.WriteToUDP(payload, target); err != nil {
			alog.Log.Println("UDP TX fail:", err)
			return
		}

		alog.Log.Println("UDP TX:", string(payload), "to:", target.String())
		seq = (seq + 1) % 100
		time.Sleep(sendInterval)
	}
}

func readUDP(conn *net.UDPConn, done chan<- struct{}) {
	defer close(done)

	buf := make([]byte, 1024)
	for {
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			alog.Log.Println("UDP RX fail:", err)
			return
		}
		if n > 0 {
			data := append([]byte(nil), buf[:n]...)
			alog.Log.Println("UDP RX:", formatNetData(data), "from:", remote.String(), "len:", n)
		}
	}
}

func parseDurationSeconds(value string, defaultSeconds int) time.Duration {
	seconds := parsePositiveInt(value, defaultSeconds)
	return time.Duration(seconds) * time.Second
}

func parsePositiveInt(value string, defaultValue int) int {
	number, err := strconv.Atoi(value)
	if err != nil || number <= 0 {
		return defaultValue
	}
	return number
}
