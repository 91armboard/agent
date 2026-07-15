package public

import alog "agent/logger"

type ActivityInfo struct {
	ActivityId string
	Step       int
	TryTimes   int
}

type UploadInfo struct {
	Id             string
	ActivityId     string
	ImgAllCount    int
	ImgUploadCount int
	ImgType        string
	IsNotified     bool
}

type UploadFile struct {
	Id          string
	InfoId      string
	ActivityId  string
	ImgPath     string
	ImgUploaded bool
}

func OpenDB() bool {
	alog.Log.Println("OpenDB ignored: legacy database is disabled")
	return false
}

func CheckSdCardAndDbStatus() bool {
	CheckSdCardMounted()
	return true
}
