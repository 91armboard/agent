package service

import (
	alog "agent/logger"
	"agent/public"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

func TCPClientStart() {
	address := net.JoinHostPort(public.AppConfig.Network.TCPServerIP, public.AppConfig.Network.TCPServerPort)
	reconnectDelay := tcpReconnectDelay()

	alog.Log.Println("TCP client init done: ok server:", address, "reconnect:", reconnectDelay)
	for {
		conn, err := dialTCPServer(address)
		if err != nil {
			alog.Log.Println("TCP connect done: fail", address, err, "next_retry:", reconnectDelay)
			time.Sleep(reconnectDelay)
			continue
		}

		alog.Log.Println("TCP connect done: ok", address)
		handleTCPServer(conn)
		alog.Log.Println("TCP connection closed:", address, "next_retry:", reconnectDelay)
		time.Sleep(reconnectDelay)
	}
}

func dialTCPServer(address string) (net.Conn, error) {
	dialer := net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	return dialer.Dial("tcp", address)
}

func handleTCPServer(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			data := append([]byte(nil), buf[:n]...)
			alog.Log.Println("TCP RX:", formatNetData(data), "len:", n)
			if err := writeTCPResponse(conn); err != nil {
				alog.Log.Println("TCP TX fail:", err)
				return
			}
		}
		if err != nil {
			if err == io.EOF {
				alog.Log.Println("TCP server disconnected")
			} else {
				alog.Log.Println("TCP read fail:", err)
			}
			return
		}
	}
}

func writeTCPResponse(conn net.Conn) error {
	response := []byte("ok" + public.AppConfig.Common.SN)
	if _, err := conn.Write(response); err != nil {
		return err
	}
	alog.Log.Println("TCP TX:", string(response))
	return nil
}

func tcpReconnectDelay() time.Duration {
	seconds, err := strconv.Atoi(public.AppConfig.Network.TCPReconnectSec)
	if err != nil || seconds <= 0 {
		seconds = 5
	}
	return time.Duration(seconds) * time.Second
}

func formatNetData(data []byte) string {
	return fmt.Sprintf("%q", string(data))
}
