package public

import (
	alog "agent/logger"
	"os"
)

func CheckSdCardMounted() {
	IsMountedSdCard = false
	sdCardPath := "/mnt/mmcblk0p1"
	filePath := sdCardPath + "/sd.txt"
	if !IsExist(filePath) {
		alog.Log.Println("SD card init done: fail not mounted")
		return
	}

	if ICheckSdCardCnt < 3 {
		if err := testWrite(sdCardPath); err != nil {
			alog.Log.Println("SD card init done: fail write", err)
			return
		}
		if err := testRead(filePath); err != nil {
			alog.Log.Println("SD card init done: fail read", err)
			return
		}
	}

	ICheckSdCardCnt++
	if ICheckSdCardCnt > 1000 {
		ICheckSdCardCnt = 0
	}
	IsMountedSdCard = true
	alog.Log.Println("SD card init done: ok", sdCardPath)
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
