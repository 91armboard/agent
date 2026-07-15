package service

import (
	alog "agent/logger"
	"agent/public"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func init() {
	go onCmdChannel(public.ChCmd)
}

func onCmdChannel(ch chan string) {
	for {
		input := <-ch
		inputs := strings.Split(input, ":")
		if len(inputs) == 2 {
			doCmd(inputs[0], inputs[1])
		}
	}
}

func CmdStart() {
	alog.Log.Println("Command service started")
}

func doCmd(action string, sdata string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("onCmdChannel recover:", err)
		}
	}()

	alog.Log.Println("DO_CMD", action, sdata)
	switch action {
	case public.CMD_GET_CONFIG:
		sendCmdResult(action, map[string]string{
			"SN":    public.Config["SN"],
			"MODEL": public.Config["MODEL"],
		})
	case public.CMD_RUN_SHELL:
		sdata = strings.Replace(sdata, "=>", ":", -1)
		err, output := public.ExecShell(sdata)
		result := "false"
		if err == nil {
			result = output
		}
		sendCmdResult(action, map[string]string{"result": result})
	case public.CMD_GET_VERSION:
		sendCmdResult(action, map[string]string{"VERSION": public.VERSION})
	case public.CMD_DOWNLOAD:
		runScriptCommand(action, "/etc/smartshop_go/download.sh")
	case public.CMD_UPGRADE:
		runScriptCommand(action, "/etc/smartshop_go/upgrade.sh")
	case public.CMD_WGET_DNLOAD:
		sdata = strings.Replace(sdata, "=>", ":", -1)
		err, output := public.ExecWgetEn(sdata)
		result := "false"
		if err == nil {
			result = output
		}
		sendCmdResult(action, map[string]string{"result": result})
	default:
		alog.Log.Println("Unsupported command:", action)
		public.SendMqttError(public.TYPE_CMD, public.ERROR_PARAM, "")
	}
	time.Sleep(10 * time.Millisecond)
}

func runScriptCommand(action string, script string) {
	err, output := public.ExecShell(script)
	result := "false"
	if err == nil {
		result = output
	}
	sendCmdResult(action, map[string]string{"result": result})
}

func sendCmdResult(action string, data map[string]string) {
	dataStr, err := json.Marshal(&data)
	if err != nil {
		alog.Log.Println("sendCmdResult marshal error:", err)
		return
	}
	dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
	public.SendMqttStatus(public.TYPE_CMD, action, dataStr2, "")
}
