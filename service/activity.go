package service

import (
	alog "agent/logger"
	"agent/public"
)

func HandleActivityCommand(action string, data string) {
	alog.Log.Println("Activity command ignored:", action, data)
	public.SendMqttError(public.TYPE_CMD, public.ERROR_PARAM, "")
}

func ActivityStartNow(activityID string) {
	alog.Log.Println("ActivityStartNow ignored:", activityID)
}

func ActivityStart(mode string) {
	alog.Log.Println("ActivityStart ignored:", mode)
}
