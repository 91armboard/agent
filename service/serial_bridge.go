package service

import (
	alog "agent/logger"
	"agent/public"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tarm/serial"
)

var (
	serialCOM1 *serial.Port
	serialCOM2 *serial.Port
)

const (
	serialLogReset = "\033[0m"
	serialLogTx    = "\033[36m"
	serialLogRx    = "\033[32m"
)

func SerialBridgeStart() {
	com1Name, com2Name := serialBridgePorts()
	baudRate := serialBaudRate()
	com1Buffer := newSerialRingBuffer(serialBufferSize())
	com1, err := openSerialPort(com1Name, baudRate)
	if err != nil {
		alog.Log.Println("Serial COM1 init done: fail", com1Name, err)
		return
	}
	com2, err := openSerialPort(com2Name, baudRate)
	if err != nil {
		com1.Close()
		alog.Log.Println("Serial COM2 init done: fail", com2Name, err)
		return
	}
	serialCOM1 = com1
	serialCOM2 = com2

	alog.Log.Println("Serial COM1 init done: ok", com1Name, "baudrate:", baudRate, "N,8,1")
	alog.Log.Println("Serial COM2 init done: ok", com2Name, "baudrate:", baudRate, "N,8,1")
	alog.Log.Println("Serial COM1 buffer init done: ok capacity:", com1Buffer.Capacity())
	alog.Log.Println("Serial COM1 frame idle:", serialFrameIdle())
	go serialHeartbeat(com1)
	go serialTestData(com1)
	go readSerialCOM1(com1, com1Buffer)
	processSerialCOM1(com1Buffer)
}

func readSerialCOM1(com1 *serial.Port, rxBuffer *serialRingBuffer) {
	buf := make([]byte, 256)
	for {
		n, _ := com1.Read(buf)
		if n > 0 {
			if ok := rxBuffer.Push(buf[:n]); !ok {
				alog.Log.Println("Serial COM1 RX buffer full: discard", n, "bytes")
			}
		}
	}
}

func processSerialCOM1(rxBuffer *serialRingBuffer) {
	frameIdle := serialFrameIdle()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	var frame []byte
	lastDataAt := time.Time{}
	for range ticker.C {
		data := rxBuffer.PopAll()
		if len(data) > 0 {
			frame = append(frame, data...)
			lastDataAt = time.Now()
			continue
		}
		if len(frame) > 0 && time.Since(lastDataAt) >= frameIdle {
			logSerialRX(frame, rxBuffer.Len())
			frame = frame[:0]
		}
	}
}

func serialHeartbeat(com1 *serial.Port) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		heartbeat := []byte{0x02, 0x47, 0x33, 0x55, 0x55, 0x55, 0x55, 0x37, 0x34, 0x03}
		writeSerialCOM1(com1, "HB", heartbeat)
	}
}

func serialTestData(com1 *serial.Port) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	seq := 1
	for range ticker.C {
		payload := []byte(fmt.Sprintf("test %02d", seq))
		writeSerialCOM1(com1, "TEST", payload)
		seq++
		if seq > 99 {
			seq = 1
		}
	}
}

func writeSerialCOM1(com1 *serial.Port, label string, data []byte) {
	if _, err := com1.Write(data); err != nil {
		alog.Log.Println("Serial COM1", label, "write error:", err)
		return
	}
	logSerialTX(label, data)
}

func logSerialTX(label string, data []byte) {
	alog.Log.Printf("%s%-10s%s %-32s len=%d\n", serialLogTx, "TX["+label+"]", serialLogReset, formatSerialHex(data), len(data))
}

func logSerialRX(data []byte, queueLen int) {
	alog.Log.Printf("%s%-10s%s %-32s len=%d q=%d\n", serialLogRx, "RX", serialLogReset, formatSerialHex(data), len(data), queueLen)
}

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

func formatSerialHex(data []byte) string {
	items := make([]string, len(data))
	for i, b := range data {
		items[i] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(items, " ")
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

type serialRingBuffer struct {
	mu       sync.Mutex
	data     []byte
	head     int
	tail     int
	length   int
	capacity int
}

func newSerialRingBuffer(capacity int) *serialRingBuffer {
	if capacity <= 0 {
		capacity = 1024
	}
	return &serialRingBuffer{
		data:     make([]byte, capacity),
		capacity: capacity,
	}
}

func (r *serialRingBuffer) Push(data []byte) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(data) == 0 {
		return true
	}
	if len(data) > r.capacity-r.length {
		return false
	}
	for _, b := range data {
		r.data[r.tail] = b
		r.tail = (r.tail + 1) % r.capacity
		r.length++
	}
	return true
}

func (r *serialRingBuffer) PopAll() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.length == 0 {
		return nil
	}
	out := make([]byte, r.length)
	for i := range out {
		out[i] = r.data[r.head]
		r.head = (r.head + 1) % r.capacity
	}
	r.length = 0
	r.tail = r.head
	return out
}

func (r *serialRingBuffer) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.length
}

func (r *serialRingBuffer) Capacity() int {
	return r.capacity
}
