package public

import (
	alog "agent/logger"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

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

	if strings.Contains(model, "IGO-668") {
		IsPubDualDoorMod = true
	} else {
		IsPubDualDoorMod = false
	}
	alog.Log.Println(model)

	if IsMountedSdCard && !IsSdCardNotFind {

		//check if /config/config.ini exist or not, if not, cp from /etc/smartshop
		configFileNameSdCardPath := GetSdcardPath() + "/config/config.ini"
		configFileNameEtcPath := GetCurPath() + "/config/config.ini"
		if !IsExist(configFileNameSdCardPath) {
			if IsExist(configFileNameEtcPath) {
				err := CopyFile(configFileNameEtcPath, configFileNameSdCardPath)
				if err != nil {
					alog.Log.Println("Error copying file:", err)
					IsSdCardNotFind = true
					return
				} else {
					alog.Log.Println("Copying file:", configFileNameEtcPath, configFileNameSdCardPath)
				}
				time.Sleep(4 * time.Second)
			}
		}
		//如果已经加载了SD卡，尝试从SD卡里读取配置文件
		cfgSdCard, errSdCard := config.ReadDefault(GetSdcardPath() + "/config/config.ini")
		//log.Println(GetSdcardPath() + "/config/config.ini")
		checkErr(errSdCard)
		cameraTypeSdCard, err := cfgSdCard.String("common", "camera_type")
		checkErr(err)
		cameraCountSdCard, err := cfgSdCard.String("common", "camera_count")
		checkErr(err)
		lockTypeSdCard, err := cfgSdCard.String("common", "lock_type")
		//alog.Log.Println("InitConfig, lockType:", lockTypeSdCard)
		checkErr(err)
		modelSdCard, err := cfgSdCard.String("common", "model")
		checkErr(err)
		snSdCard, err := cfgSdCard.String("common", "sn")
		checkErr(err)

		if (cameraType != cameraTypeSdCard) || (cameraCount != cameraCountSdCard) || (lockType != lockTypeSdCard) || (model != modelSdCard) || (sn != snSdCard) {
			// 配置文件有不同
			data := "config common 'common'" + "\n"
			data = data + "  option model '" + modelSdCard + "'" + "\n"
			data = data + "  option camera_type '" + cameraTypeSdCard + "'" + "\n"
			data = data + "  option camera_count '" + cameraCountSdCard + "'" + "\n"
			data = data + "  option lock_type '" + lockTypeSdCard + "'" + "\n"
			data = data + "  option sn '" + snSdCard + "'" + "\n"
			CreateSnFile(true, snSdCard)
			WriteFile("/etc/config/ss_agent", data)
			time.Sleep(5 * time.Second)
			ExecShell("reload_config ss_agent")
			os.Exit(-1)
		}
	}
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
			//panic(err)
			alog.Log.Println("CreateSnFile:json", err)
			return
		}

		err = os.WriteFile("/www/sn.json", content, 0644)
		if err != nil {
			//panic(err)
			alog.Log.Println("CreateSnFile:WriteFile", err)
			return
		}
		alog.Log.Println("CreateSnFile:", strsn)
	}

}
