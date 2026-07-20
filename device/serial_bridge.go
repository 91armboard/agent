package device

import (
	alog "agent/logger"
	"fmt"
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
	rxFrameCh := make(chan []byte, 4)
	go serialTxScheduler(com1, rxFrameCh)
	go readSerialCOM1(com1, com1Buffer)
	processSerialCOM1(com1Buffer, rxFrameCh)
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

func processSerialCOM1(rxBuffer *serialRingBuffer, rxFrameCh chan<- []byte) {
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
			notifySerialFrame(rxFrameCh, frame)
			frame = frame[:0]
		}
	}
}

func serialTxScheduler(com1 *serial.Port, rxFrameCh <-chan []byte) {
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	nextHeartbeatAt := time.Now().Add(5 * time.Second)
	nextTestAt := time.Now().Add(3 * time.Second)
	lastTestAt := time.Time{}
	testBlockedUntil := time.Time{}
	heartbeatPending := false
	heartbeatDeadline := time.Time{}
	seq := 1

	for {
		select {
		case frame := <-rxFrameCh:
			if heartbeatPending && isVMHeartbeatAckFrame(frame) {
				heartbeatPending = false
				testBlockedUntil = time.Now().Add(50 * time.Millisecond)
				alog.Log.Println("HB ACK len:", len(frame), "test_after:", testBlockedUntil.Format("15:04:05.000"))
			}
		case now := <-ticker.C:
			if heartbeatPending && now.After(heartbeatDeadline) {
				heartbeatPending = false
				testBlockedUntil = now
				alog.Log.Println("HB ACK timeout: test resumed")
			}

			if !heartbeatPending && !now.Before(nextHeartbeatAt) && canSendHeartbeat(now, lastTestAt) {
				writeSerialCOM1(com1, "HB", buildVMHeartbeatFrame())
				heartbeatPending = true
				heartbeatDeadline = now.Add(1 * time.Second)
				nextHeartbeatAt = now.Add(5 * time.Second)
				continue
			}

			if !heartbeatPending && !now.Before(nextTestAt) && !now.Before(testBlockedUntil) {
				payload := []byte(fmt.Sprintf("test %02d", seq))
				writeSerialCOM1(com1, "TEST", payload)
				lastTestAt = now
				nextTestAt = now.Add(3 * time.Second)
				seq++
				if seq > 99 {
					seq = 1
				}
			}
		}
	}
}

func notifySerialFrame(rxFrameCh chan<- []byte, frame []byte) {
	copied := append([]byte(nil), frame...)
	select {
	case rxFrameCh <- copied:
	default:
	}
}

func canSendHeartbeat(now time.Time, lastTestAt time.Time) bool {
	return lastTestAt.IsZero() || now.Sub(lastTestAt) >= 200*time.Millisecond
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
