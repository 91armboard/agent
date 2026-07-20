package device

import "fmt"

func buildVMHeartbeatFrame() []byte {
	return []byte{0x02, 0x47, 0x33, 0x55, 0x55, 0x55, 0x55, 0x37, 0x34, 0x03}
}

func isVMHeartbeatAckFrame(frame []byte) bool {
	if len(frame) != 10 || frame[0] != 0x02 || frame[9] != 0x03 {
		return false
	}

	crc := frame[1]
	for i := 2; i <= 6; i++ {
		crc ^= frame[i]
	}
	crcText := fmt.Sprintf("%X", crc)
	if len(crcText) == 1 {
		return frame[7] == crcText[0] && frame[8] == 0x00
	}
	return frame[7] == crcText[0] && frame[8] == crcText[1]
}
