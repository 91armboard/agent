package device

import (
	alog "agent/logger"
	"os"
)

var (
	isMountedSdCard bool
	checkSdCardCnt  int
)

func CheckSdCardMounted() {
	isMountedSdCard = false
	sdCardPath := "/mnt/mmcblk0p1"
	filePath := sdCardPath + "/sd.txt"
	if !isExist(filePath) {
		alog.Log.Println("SD card init done: fail not mounted")
		return
	}

	if checkSdCardCnt < 3 {
		if err := testWrite(sdCardPath); err != nil {
			alog.Log.Println("SD card init done: fail write", err)
			return
		}
		if err := testRead(filePath); err != nil {
			alog.Log.Println("SD card init done: fail read", err)
			return
		}
	}

	checkSdCardCnt++
	if checkSdCardCnt > 1000 {
		checkSdCardCnt = 0
	}
	isMountedSdCard = true
	alog.Log.Println("SD card init done: ok", sdCardPath)
}

func IsSdCardMounted() bool {
	return isMountedSdCard
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func testWrite(path string) error {
	_ = os.Remove(path + "/sd.txt")
	file, err := os.Create(path + "/sd.txt")
	if err != nil {
		return err
	}
	return file.Close()
}

func testRead(filePath string) error {
	_, err := os.ReadFile(filePath)
	return err
}
