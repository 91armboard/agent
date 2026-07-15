package public

import (
	alog "agent/logger"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/larspensjo/config"
)

func init() {
	ChMqtt = make(chan string)
	ChActivity = make(chan string)
	ChUpload = make(chan string)
	ChCmd = make(chan string)
}

func InitConfig() {
	cfg, err := config.ReadDefault(GetCurPath() + "/config/config.ini")
	checkErr(err)

	cameraType, err := cfg.String("common", "camera_type")
	checkErr(err)
	cameraCount, err := cfg.String("common", "camera_count")
	checkErr(err)
	lockType, err := cfg.String("common", "lock_type")
	checkErr(err)
	model, err := cfg.String("common", "model")
	checkErr(err)
	sn, err := cfg.String("common", "sn")
	checkErr(err)

	Config = map[string]string{}
	Config["CAMERA_TYPE"] = cameraType
	Config["CAMERA_COUNT"] = cameraCount
	Config["LOCK_TYPE"] = lockType
	Config["MODEL"] = model
	Config["SN"] = sn

	IsPubDualDoorMod = strings.Contains(model, "IGO-668")
	alog.Log.Println(model)
	CreateSnFile(false, sn)
}

func checkErr(err error) {
	if err != nil {
		alog.Log.Println(err)
		os.Exit(-1)
	}
}

func setFrpIni(fileName, sn string) {
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		alog.Log.Println("READ FRP INI ERROR!")
	}
	fileContent, err := io.ReadAll(f)
	temp := strings.Replace(string(fileContent), "{SN}", sn, -1)
	if os.WriteFile(fileName, []byte(temp), 0766) != nil {
		alog.Log.Println("WRITE FRP INI ERROR!")
	}
}

func CreateSnFile(again bool, strsn string) {
	if !IsExist("/www/sn.json") {
		again = true
		alog.Log.Println("CreateSnFile:exist")
	}

	if again {
		data := map[string]string{"sn": strsn}
		content, err := json.Marshal(data)
		if err != nil {
			alog.Log.Println("CreateSnFile:json", err)
			return
		}

		err = os.WriteFile("/www/sn.json", content, 0644)
		if err != nil {
			alog.Log.Println("CreateSnFile:WriteFile", err)
			return
		}
		alog.Log.Println("CreateSnFile:", strsn)
	}
}
