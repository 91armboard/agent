package service

import alog "agent/logger"

func StatusStart() {
	alog.Log.Println("StatusStart ignored: legacy status checks are disabled")
}
