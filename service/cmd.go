package service

import (
	"agent/device"
	alog "agent/logger"
	"agent/public"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/larspensjo/config"
)

func init() {
	go onCmdChannel(public.ChCmd)
}

func onCmdChannel(ch chan string) {
	var input string
	for {
		input = <-ch
		inputs := strings.Split(input, ":")
		if len(inputs) == 2 {
			doCmd(inputs[0], inputs[1])
		}
	}
}

func CmdStart() {
	// for {
	// 	time.Sleep(1 * time.Second)
	// }
}

func doCmd(action string, sdata string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("onCmdChannel 捕获异常:", err)
		}
	}()
	alog.Log.Println("DO_CMD", action, sdata)
	if action == public.CMD_GET_CONFIG {
		// 获取系统配置
		data := make(map[string]string)
		data["SN"] = public.Config["SN"]
		data["MODEL"] = public.Config["MODEL"]
		data["CAMERA_COUNT"] = public.Config["CAMERA_COUNT"]
		data["CAMERA_TYPE"] = public.Config["CAMERA_TYPE"]
		data["LOCK_TYPE"] = public.Config["LOCK_TYPE"]
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_GET_CONFIG, dataStr2, "")
		}
		return
	}

	if action == public.CMD_SET_CONFIG {
		// TODO 设置配置
		//如果已经加载了SD卡，尝试从SD卡里读取配置文件
		cfgSdCard, errSdCard := config.ReadDefault(public.GetSdcardPath() + "/config/config.ini")
		//log.Println(GetSdcardPath() + "/config/config.ini")
		fmt.Println(errSdCard)
		cameraTypeSdCard, err := cfgSdCard.String("common", "camera_type")
		fmt.Println(err)
		cameraCountSdCard, err := cfgSdCard.String("common", "camera_count")
		fmt.Println(err)
		lockTypeSdCard, err := cfgSdCard.String("common", "lock_type")
		//alog.Log.Println("InitConfig, lockType:", lockTypeSdCard)
		fmt.Println(err)
		modelSdCard, err := cfgSdCard.String("common", "model")
		fmt.Println(err)
		snSdCard, err := cfgSdCard.String("common", "sn")
		fmt.Println(err)

		if sdata == "cm=ijooz" {
			// 配置文件有不同
			alog.Log.Println("CMD_RUN_SHELL", ":change cm type to ijooz")
			cameraTypeSdCard = "ijooz"
			cdata := "config common 'common'" + "\n"
			cdata = cdata + "  option model '" + modelSdCard + "'" + "\n"
			cdata = cdata + "  option camera_type '" + cameraTypeSdCard + "'" + "\n"
			cdata = cdata + "  option camera_count '" + cameraCountSdCard + "'" + "\n"
			cdata = cdata + "  option lock_type '" + lockTypeSdCard + "'" + "\n"
			cdata = cdata + "  option sn '" + snSdCard + "'" + "\n"
			os.Remove("/mnt/mmcblk0p1//config/config.ini")
			public.CreateSnFile(true, snSdCard)
			public.WriteFile("/etc/config/ss_agent", cdata)
			time.Sleep(3 * time.Second)
			public.ExecShell("reload_config ss_agent")
			//os.Exit(-1)
		} else if sdata == "cm=haha" {
			// 配置文件有不同
			alog.Log.Println("CMD_RUN_SHELL", ":change cm type to haha")
			cameraTypeSdCard = "haha"
			cdata := "config common 'common'" + "\n"
			cdata = cdata + "  option model '" + modelSdCard + "'" + "\n"
			cdata = cdata + "  option camera_type '" + cameraTypeSdCard + "'" + "\n"
			cdata = cdata + "  option camera_count '" + cameraCountSdCard + "'" + "\n"
			cdata = cdata + "  option lock_type '" + lockTypeSdCard + "'" + "\n"
			cdata = cdata + "  option sn '" + snSdCard + "'" + "\n"
			os.Remove("/mnt/mmcblk0p1//config/config.ini")
			public.CreateSnFile(true, snSdCard)
			public.WriteFile("/etc/config/ss_agent", cdata)
			time.Sleep(3 * time.Second)
			public.ExecShell("reload_config ss_agent")
			//os.Exit(-1)
		} else if sdata == "lk=ijooz" {
			// 配置文件有不同
			alog.Log.Println("CMD_RUN_SHELL", ":change lk type to ijooz")
			lockTypeSdCard = "ijooz"
			cdata := "config common 'common'" + "\n"
			cdata = cdata + "  option model '" + modelSdCard + "'" + "\n"
			cdata = cdata + "  option camera_type '" + cameraTypeSdCard + "'" + "\n"
			cdata = cdata + "  option camera_count '" + cameraCountSdCard + "'" + "\n"
			cdata = cdata + "  option lock_type '" + lockTypeSdCard + "'" + "\n"
			cdata = cdata + "  option sn '" + snSdCard + "'" + "\n"
			os.Remove("/mnt/mmcblk0p1//config/config.ini")
			public.CreateSnFile(true, snSdCard)
			public.WriteFile("/etc/config/ss_agent", cdata)
			time.Sleep(3 * time.Second)
			public.ExecShell("reload_config ss_agent")
			//os.Exit(-1)
		} else if sdata == "lk=haha" {
			// 配置文件有不同
			alog.Log.Println("CMD_RUN_SHELL", ":change lk type to haha")
			lockTypeSdCard = "haha"
			cdata := "config common 'common'" + "\n"
			cdata = cdata + "  option model '" + modelSdCard + "'" + "\n"
			cdata = cdata + "  option camera_type '" + cameraTypeSdCard + "'" + "\n"
			cdata = cdata + "  option camera_count '" + cameraCountSdCard + "'" + "\n"
			cdata = cdata + "  option lock_type '" + lockTypeSdCard + "'" + "\n"
			cdata = cdata + "  option sn '" + snSdCard + "'" + "\n"
			os.Remove("/mnt/mmcblk0p1//config/config.ini")
			public.CreateSnFile(true, snSdCard)
			public.WriteFile("/etc/config/ss_agent", cdata)
			time.Sleep(3 * time.Second)
			public.ExecShell("reload_config ss_agent")
			//os.Exit(-1)
		} else if sdata == "cnt=2" {
			// 配置文件有不同
			alog.Log.Println("CMD_RUN_SHELL", ":change lk type to haha")
			cameraCountSdCard = "2"
			cdata := "config common 'common'" + "\n"
			cdata = cdata + "  option model '" + modelSdCard + "'" + "\n"
			cdata = cdata + "  option camera_type '" + cameraTypeSdCard + "'" + "\n"
			cdata = cdata + "  option camera_count '" + cameraCountSdCard + "'" + "\n"
			cdata = cdata + "  option lock_type '" + lockTypeSdCard + "'" + "\n"
			cdata = cdata + "  option sn '" + snSdCard + "'" + "\n"
			os.Remove("/mnt/mmcblk0p1//config/config.ini")
			public.CreateSnFile(true, snSdCard)
			public.WriteFile("/etc/config/ss_agent", cdata)
			time.Sleep(3 * time.Second)
			public.ExecShell("reload_config ss_agent")
			//os.Exit(-1)
		} else if sdata == "cnt=1" {
			// 配置文件有不同
			alog.Log.Println("CMD_RUN_SHELL", ":change lk type to haha")
			cameraCountSdCard = "1"
			cdata := "config common 'common'" + "\n"
			cdata = cdata + "  option model '" + modelSdCard + "'" + "\n"
			cdata = cdata + "  option camera_type '" + cameraTypeSdCard + "'" + "\n"
			cdata = cdata + "  option camera_count '" + cameraCountSdCard + "'" + "\n"
			cdata = cdata + "  option lock_type '" + lockTypeSdCard + "'" + "\n"
			cdata = cdata + "  option sn '" + snSdCard + "'" + "\n"
			os.Remove("/mnt/mmcblk0p1//config/config.ini")
			public.CreateSnFile(true, snSdCard)
			public.WriteFile("/etc/config/ss_agent", cdata)
			time.Sleep(3 * time.Second)
			public.ExecShell("reload_config ss_agent")
			//os.Exit(-1)
		}
		data := make(map[string]string)
		data["CAMERA_TYPE"] = cameraTypeSdCard
		data["CAMERA_COUNT"] = cameraCountSdCard
		data["LOCK_TYPE"] = lockTypeSdCard
		data["MODEL"] = modelSdCard
		data["SN"] = snSdCard
		dataStr, err := json.Marshal(&data)

		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_GET_CONFIG, dataStr2, "")
		}
		return
	}

	if action == public.CMD_GET_STATUS {
		// 获取状态
		data := make(map[string]string)
		data["CurActivityId"] = device.CurActivityId
		data["IsOnline"] = strconv.FormatBool(device.IsOnline)
		data["Is4GOn"] = strconv.FormatBool(device.Is4GPowOn)
		data["IsLog"] = strconv.FormatBool(device.IsLog)
		data["IsActivityRuning"] = strconv.FormatBool(device.GetIsActivityRunning())
		data["IsDownByLock"] = strconv.FormatBool(device.IsOperationDownByLock)
		data["IsDownByDoor"] = strconv.FormatBool(device.IsOperationDownByDoor)
		data["IsDownByLock2"] = strconv.FormatBool(device.IsOperationDownByLock2)
		data["IsDownByDoor2"] = strconv.FormatBool(device.IsOperationDownByDoor2)
		data["IsBuzzerOn"] = strconv.FormatBool(device.IsBuzzerOn)
		data["IsCamra1On"] = strconv.FormatBool(device.IsCamraOn[0])
		data["IsCamra2On"] = strconv.FormatBool(device.IsCamraOn[1])
		data["IsCamra3On"] = strconv.FormatBool(device.IsCamraOn[2])
		data["IsCamra4On"] = strconv.FormatBool(device.IsCamraOn[3])
		data["IsDoorClosed"] = strconv.FormatBool(device.IsDoorClosed)
		data["IsLockLocked"] = strconv.FormatBool(device.IsLockLocked)
		data["IsDoor2Closed"] = strconv.FormatBool(device.IsDoor2Closed)
		data["IsLock2Locked"] = strconv.FormatBool(device.IsLock2Locked)
		data["IsMPowerOn"] = strconv.FormatBool(device.IsMPowerOn)
		data["IsSerialing"] = strconv.FormatBool(device.IsSerialing)
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_GET_STATUS, dataStr2, "")
		}
		return
	}

	if action == public.CMD_RUN_SHELL {
		// 运行Shell
		sdata = strings.Replace(string(sdata), "=>", ":", -1)
		//0404wget http://upload.shop.ijooz.sg/agent/md5.upgrade -P /tmp
		err, re := public.ExecShell(sdata)
		alog.Log.Println("CMD_RUN_SHELL", err, re)
		data := make(map[string]string)
		if err == nil {
			data["result"] = re
		} else {
			data["result"] = "false"
		}
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_RUN_SHELL, dataStr2, "")
		}
		time.Sleep(1 * time.Second)
		return
	}

	if action == public.CMD_GET_VERSION {
		// 获取版本号
		v := public.ACTION_BEGIN
		alog.Log.Println(v)
		data := make(map[string]string)
		data["VERSION"] = public.VERSION
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_GET_VERSION, dataStr2, "")
		}
		return
	}
	if action == public.CMD_OPEN_LOCK {
		// 开锁
		data := make(map[string]string)
		var itype byte
		if !device.GetIsActivityRunning() {
			if sdata == "8302" || sdata == "5089" {
				device.SetIsActivityRunning(true)
				if sdata == "8302" {
					itype = 0x01
					data["door"] = "A"
				} else {
					itype = 0x02
					data["door"] = "B"
				}
				err := device.OpenAllLock(itype)
				if err == nil {
					data["result"] = "true"
				} else {
					data["result"] = "false"
					data["door"] = "na"
				}
				device.SetIsActivityRunning(false)
			} else {
				data["result"] = "auth fail"
			}
		} else {
			data["result"] = "busy"
		}
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_OPEN_LOCK, dataStr2, "")
		}
		time.Sleep(1 * time.Second)
		return
	}
	if action == public.CMD_CLOSE_LOCK {
		// 关锁
		device.CloseAllLock(0x03)
		data := make(map[string]string)
		data["result"] = "done"
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_CLOSE_LOCK, dataStr2, "")
		}
		return
	}
	if action == public.CMD_SET_LIGHT {
		// 设置亮度
		var err error = errors.New("")
		if sdata == "0" {
			err = device.SetLight0()
		} else if sdata == "1" {
			err = device.SetLight1()
		} else if sdata == "2" {
			err = device.SetLight2()
		} else if sdata == "3" {
			err = device.SetLight3()
		} else if sdata == "4" {
			err = device.SetLight4()
		}
		data := make(map[string]string)
		if err == nil {
			data["result"] = "true"
		} else {
			data["result"] = "false"
		}
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_SET_LIGHT, dataStr2, "")
		}
		return
	}
	if action == public.CMD_RESET_CAMERA {
		// 重启摄像头
		i, err := strconv.Atoi(sdata)
		data := make(map[string]string)
		if err == nil {
			err := device.ResetCamra(i)
			if err == nil {
				data["result"] = "true"
			} else {
				data["result"] = "false"
			}
		} else {
			data["result"] = "false"
		}
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_RESET_CAMERA, dataStr2, "")
		}
		time.Sleep(1 * time.Second)
		return
	}
	if action == public.CMD_CLEAR_ACTIVITY {
		// 清除活动
		err := device.ClearActivity()
		data := make(map[string]string)
		if err == nil {
			data["result"] = "true"
		} else {
			data["result"] = "false"
		}
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_CLEAR_ACTIVITY, dataStr2, "")
		}
		return
	}
	if action == public.CMD_RESTART {
		// 重启
		var err error
		data := make(map[string]string)
		if err == nil {
			data["result"] = "reset cpu"
		} else {
			data["result"] = "false"
		}
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_RESTART, dataStr2, "")
		}
		err = device.ResetCPU()
		return
	}
	if action == public.CMD_RESET_ALL {
		// 重启
		err := device.ResetAll()
		data := make(map[string]string)
		if err == nil {
			data["result"] = "rest all"
		} else {
			data["result"] = "rest done"
		}
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_RESET_ALL, dataStr2, "")
		}
		return
	}
	if action == public.CMD_DOWNLOAD {
		// 下载更新包
		data := make(map[string]string)
		if device.GetIsActivityRunning() {
			data["result"] = "busy"
		} else {
			if public.IsExist("/tmp/ss_main.upgrade") {
				os.Remove("/tmp/ss_main.upgrade")
			}
			if public.IsExist("/tmp/md5.upgrade") {
				os.Remove("/tmp/md5.upgrade")
			}
			err, re := public.ExecShell("/etc/smartshop_go/download.sh") //0404/etc/smartshop_go/download.sh
			alog.Log.Println("CMD_RUN_SHELL", err, re)
			if err == nil {
				data["result"] = "ok" //re
			} else {
				data["result"] = "false"
			}
		}
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_DOWNLOAD, dataStr2, "")
		}
		time.Sleep(1 * time.Second)
		return
	}
	if action == public.CMD_UPGRADE {
		// 更新主控
		data := make(map[string]string)

		if device.GetIsActivityRunning() && (!device.IsOperationDownByLock && !device.IsOperationDownByLock2) {
			data["result"] = "busy1"
			dataStr, err := json.Marshal(&data)
			if err == nil {
				dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
				public.SendMqttStatus(public.TYPE_CMD, public.CMD_DOWNLOAD, dataStr2, "")
			}
			return
		}
		//delete all empty folders here
		vDelEmptyFolders()
		if !public.IsExist("/tmp/md5.upgrade") {
			urlmd5 := "http://upload.shop.ijooz.sg/agent/md5.upgrade"
			md5Path := "/var/md5.upgrade"

			err, _ := public.ExecWget(urlmd5, md5Path)
			if err == nil {
				data["md5.upgrade"] = "ok"
			} else {
				data["md5.upgrade"] = "na"
			}
			dataStr, err := json.Marshal(&data)
			if err == nil {
				dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
				public.SendMqttStatus(public.TYPE_CMD, public.CMD_UPGRADE, dataStr2, "")
			}
		}

		bMD5FileCompare := false
		if public.IsExist("/tmp/md5.upgrade") {
			//compare old md5, if same then return
			if public.IsExist("/tmp/md5.bak") {
				bMD5FileCompare, _ = CompareFiles("/tmp/md5.upgrade", "/tmp/md5.bak")
			}
		}

		if !bMD5FileCompare {
			data["md5_cmp"] = "diff"
			dataStr, err := json.Marshal(&data)
			if err == nil {
				dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
				public.SendMqttStatus(public.TYPE_CMD, public.CMD_UPGRADE, dataStr2, "")
			}
		} else {
			data["md5_cmp"] = "same"
			dataStr, err := json.Marshal(&data)
			if err == nil {
				dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
				public.SendMqttStatus(public.TYPE_CMD, public.CMD_UPGRADE, dataStr2, "")
			}
			return
		}

		if !public.IsExist("/tmp/ss_main.upgrade") {
			urlmain := "http://upload.shop.ijooz.sg/agent/main"
			mainPath := "/var/ss_main.upgrade"

			err, _ := public.ExecWget(urlmain, mainPath)
			if err == nil {
				data["main"] = "ok"
			} else {
				data["main"] = "na"
			}
			dataStr, err := json.Marshal(&data)
			if err == nil {
				dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
				public.SendMqttStatus(public.TYPE_CMD, public.CMD_UPGRADE, dataStr2, "")
			}

			time.Sleep(2 * time.Second)
		}

		if device.GetIsActivityRunning() {
			data["result"] = "busy2"
			dataStr, err := json.Marshal(&data)
			if err == nil {
				dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
				public.SendMqttStatus(public.TYPE_CMD, public.CMD_UPGRADE, dataStr2, "")
			}
			return
		}

		if true {
			isMD5Ok, err := VerifyMD5Hash("/var/ss_main.upgrade", "/var/md5.upgrade")
			if !isMD5Ok {
				data["md5_chk"] = "fail"
			} else {
				data["md5_chk"] = "pass"
			}
			dataStr, err := json.Marshal(&data)
			if err == nil {
				dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
				public.SendMqttStatus(public.TYPE_CMD, public.CMD_UPGRADE, dataStr2, "")
			}
			if !isMD5Ok {
				return
			}
		}

		//rename md5 before upgrade
		os.Rename("/var/md5.upgrade", "/var/md5.bak")

		err, re := public.ExecShell("/etc/smartshop_go/upgrade.sh") //0404/etc/smartshop_go/upgrade.sh
		alog.Log.Println("CMD_RUN_SHELL", err, re)
		if err == nil {
			data["result"] = re
		} else {
			data["result"] = "upgrade..."
		}
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_UPGRADE, dataStr2, "")
		}
		return
	}
	if action == public.CMD_GET_HBT {
		// get heartbeat msg
		data := make(map[string]string)
		data["Ht"] = strconv.Itoa(device.IHeartBeatCnt)
		data["4G"] = device.StrTelProvider //"F" //"G" "L"
		data["MN"] = device.StrSimCardNumb
		data["TX"] = device.StrTxByteNum
		data["RX"] = device.StrRxByteNum
		cameraCount, _ := strconv.Atoi(public.Config["CAMERA_COUNT"])
		if cameraCount > 4 {
			cameraCount = 4
		}
		data["Cm1"] = "X"
		data["Cm2"] = "X"
		data["Cm3"] = "X"
		data["Cm4"] = "X"
		if public.IsPubDualDoorMod {
			if cameraCount > 2 {
				cameraCount = 4
			} else {
				cameraCount *= 2
			}
		}
		for i := 0; i < cameraCount; i++ {
			strCamra := fmt.Sprintf("Cm%d", i+1)
			if device.IsCamraOn[i] {
				data[strCamra] = "Y"
			} else {
				data[strCamra] = "N"
			}
		}
		if device.IsADLive {
			data["Ad"] = "Y"
		} else {
			data["Ad"] = "N"
		}
		data["Tp"] = strconv.FormatFloat(device.Temperature, 'f', 2, 64)
		if device.IsDoorClosed && device.IsDoor2Closed {
			data["Dr"] = "C"
		} else if device.IsDoorClosed {
			data["Dr"] = "B"
		} else if device.IsDoor2Closed {
			data["Dr"] = "A"
		} else {
			data["Dr"] = "O"
		}
		if device.IsLockLocked && device.IsLock2Locked {
			data["Lk"] = "L"
		} else if device.IsLockLocked {
			data["Lk"] = "B"
		} else if device.IsLock2Locked {
			data["Lk"] = "A"
		} else {
			data["Lk"] = "U"
		}
		if public.IsMountedSdCard && !public.IsSdCardNotFind {
			data["Sd"] = "Y"
		} else {
			siz, _ := public.GetTmpVideoMegaSize()
			if siz > 60 {
				data["Sd"] = "O"
			} else {
				data["Sd"] = "N"
			}
		}
		if device.IsOperationDownByDoor || device.IsOperationDownByLock {
			data["St"] = "N"
		} else {
			data["St"] = "Y"
		}
		data["Nt"] = strconv.Itoa(device.INoTransTim)
		data["Ts"] = strconv.Itoa(device.ITotalTransCnt)
		device.I24HTransCnt = 0
		for i := 0; i < 24; i++ {
			device.I24HTransCnt += device.SalesOf24H[i]
			if i == device.Iindexof24H {
				fmt.Println("SALESOF24H", i, device.SalesOf24H[i])
			} else {
				fmt.Println("SalesOf24H", i, device.SalesOf24H[i])
			}
		}
		data["24H"] = strconv.Itoa(device.I24HTransCnt)
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_GET_HBT, dataStr2, "")
		}
		return
	}
	if action == public.CMD_GET_UPLOAD {
		// send MTU and upload num
		alog.Log.Println("CMD_GET_UPLOAD")
		data := make(map[string]string)
		data["MTU"] = strconv.Itoa(public.IUploadMTUnum)
		data["UploadErrCnt"] = strconv.Itoa(public.IUploadErrCnt)
		data["BigFile"] = strconv.FormatBool(public.IsUploadBigfile)
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_GET_UPLOAD, dataStr2, "")
		}
		return
	}
	if action == public.CMD_WGET_DNLOAD {
		// send MTU and upload num
		alog.Log.Println("CMD_WGET_DNLOAD")
		data := make(map[string]string)
		sdata = strings.Replace(string(sdata), "=>", ":", -1)
		err, re := public.ExecWgetEn(sdata)
		alog.Log.Println("CMD_RUN_WGET", err, re)
		if err == nil {
			data["wget"] = "ok"
		} else {
			data["wget"] = "fail"
		}
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_WGET_DNLOAD, dataStr2, "")
		}
		return
	}
	if action == public.CMD_GET_SIMINFO {
		// send MTU and upload num
		alog.Log.Println("CMD_GET_SIMINFO")
		device.Get4GProvider()
		data := make(map[string]string)
		data["MN"] = device.StrSimCardNumb
		data["4G"] = device.StrTelProvider
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_GET_SIMINFO, dataStr2, "")
		}
		return
	}
	if action == public.CMD_GET_DATUSAG {
		// send MTU and upload num
		alog.Log.Println("CMD_GET_DATUSAG")
		device.Get4GDataUsage()
		data := make(map[string]string)
		data["TX"] = device.StrTxByteNum
		data["RX"] = device.StrRxByteNum
		dataStr, err := json.Marshal(&data)
		if err == nil {
			dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
			public.SendMqttStatus(public.TYPE_CMD, public.CMD_GET_DATUSAG, dataStr2, "")
		}
		return
	}

}

// VerifyMD5Hash checks if the MD5 hash of the specified file matches the expected hash value
func VerifyMD5Hash(filePath string, expectedHashFilePath string) (bool, error) {
	// Read the contents of the expected hash file
	expectedHashBytes, err := os.ReadFile(expectedHashFilePath)
	if err != nil {
		return false, err
	}
	if len(expectedHashBytes) == 34 {
		// Replace all occurrences of "\r\n" with an empty byte slice.
		expectedHashBytes = expectedHashBytes[:len(expectedHashBytes)-2]
	}
	//log.Println(expectedHashBytes)
	// Convert the expected hash value to a byte slice
	expectedHash := make([]byte, hex.DecodedLen(len(expectedHashBytes)))
	_, err = hex.Decode(expectedHash, expectedHashBytes)
	if err != nil {
		return false, err
	}
	//log.Println(expectedHash)
	// Open the file that we want to check
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Compute the MD5 hash of the file
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false, err
	}

	// Compare the computed hash value to the expected hash value
	computedHash := hash.Sum(nil)
	return hex.EncodeToString(computedHash) == hex.EncodeToString(expectedHash), nil
}

func CompareFiles(file1, file2 string) (bool, error) {
	info1, err := os.Stat(file1)
	if err != nil {
		return false, fmt.Errorf("error getting file info for %s: %v", file1, err)
	}
	info2, err := os.Stat(file2)
	if err != nil {
		return false, fmt.Errorf("error getting file info for %s: %v", file2, err)
	}

	if info1.Size() == 0 || info2.Size() == 0 {
		return false, fmt.Errorf("empty file(s)")
	}

	content1, err := os.ReadFile(file1)
	if err != nil {
		return false, fmt.Errorf("error reading %s: %v", file1, err)
	}
	content2, err := os.ReadFile(file2)
	if err != nil {
		return false, fmt.Errorf("error reading %s: %v", file2, err)
	}

	return string(content1) == string(content2), nil
}

func vDelEmptyFolders() {
	device.SetIsActivityRunning(true)
	defer func() {
		device.SetIsActivityRunning(false)
	}()
	videofolder := public.GetSdcardPath() + "/video"
	err := deleteEmptyFolders(videofolder)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func deleteEmptyFolders(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	// Recursively delete empty folders
	for _, entry := range entries {
		if entry.IsDir() {
			subDirPath := filepath.Join(dirPath, entry.Name())
			err = deleteEmptyFolders(subDirPath)
			if err != nil {
				return err
			}
		}
	}

	// Check if directory is empty and delete it if it is
	entries, err = os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	isEmpty := true
	for _, entry := range entries {
		if entry.Name() != "." && entry.Name() != ".." {
			isEmpty = false
			break
		}
	}
	if isEmpty {
		err = os.Remove(dirPath)
		if err != nil {
			return err
		}
		fmt.Printf("Deleted directory %s\n", dirPath)
	}

	return nil
}
