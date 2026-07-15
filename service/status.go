package service

import (
	"agent/device"
	alog "agent/logger"
	"agent/public"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	iChkDoorCnt  int
	iChkLock1Cnt int
	iChkLock2Cnt int

	isCamraOn [4]bool // 摄像头1状态
	isADLive  bool
	fTemp     float64

	isDoorClosed  bool
	isDoor2Closed bool
	isLockLocked  bool
	isLock2Locked bool

	isMountedSdCard bool
	isSdCardNotFind bool
	isSdSizeOK      bool

	isOperationDownByLock  bool
	isOperationDownByDoor  bool
	isOperationDownByLock2 bool
	isOperationDownByDoor2 bool

	strTxByteNum string
	strRxByteNum string
)

func init() {
	device.SetLight2()
	go onStatusChannel(public.ChStatus)
}
func onStatusChannel(ch chan string) {
	var input string
	for {
		input = <-ch
		alog.Log.Printf("%s \n", input)
	}
}

func StatusStart() {
	alog.Log.Println("STATUS START")
	expiredCache := public.NewCache(6)

	for {
		time.Sleep(1 * time.Second)
		if !expiredCache.Exist("CHECKING_PING_PONG") {
			expiredCache.SetWithExpiredTime("CHECKING_PING_PONG", []byte("1"), 21)
			CheckPingpong()
		}
		if !expiredCache.Exist("CHECKING_CAMRA") {
			expiredCache.SetWithExpiredTime("CHECKING_CAMRA", []byte("1"), 17)
			CheckCamra()
		}

		if !expiredCache.Exist("CHECKING_DOOR") {
			expiredCache.SetWithExpiredTime("CHECKING_DOOR", []byte("1"), 6)
			CheckDoor() //检查门的状态...
		}

		//ping 192.168.199.88--if offline, reset LED4
		if !expiredCache.Exist("CHECKING_4G") {
			expiredCache.SetWithExpiredTime("CHECKING_4G", []byte("1"), 31)
			Check4G()
			CheckADLive()
		}

		// if !expiredCache.Exist("CHECKING_POWER") {
		// 	expiredCache.SetWithExpiredTime("CHECKING_POWER", []byte("1"), 30)
		// 	CheckPower()
		// }

		if !expiredCache.Exist("CHECKING_CONFIG") {
			expiredCache.SetWithExpiredTime("CHECKING_CONFIG", []byte("1"), 60)
			CheckConfig()
		}

		//every 5min //get temprature also
		if !expiredCache.Exist("MQTT HEARTBEAT") {
			expiredCache.SetWithExpiredTime("MQTT HEARTBEAT", []byte("1"), 353)
			MqttHeartbeat()
			//GetTemp()
		}

	}
}

// 获取摄像头状态，ping 命令
func getCamraStatus(i int) bool {
	snapShotHost := public.GetCameraHost()
	j := 0
	cameraType := public.Config["CAMERA_TYPE"]
	if cameraType == "haha" && public.IsPubDualDoorMod {
		j = i + 0
	} else {
		j = i + 1
	}
	// log.Println("PING: ", fmt.Sprintf(snapShotHost, i+1))
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", fmt.Sprintf(snapShotHost, j), "-n", "1")
	} else {
		cmd = exec.Command("ping", fmt.Sprintf(snapShotHost, j), "-c", "1", "-W", "5")
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
	if strings.Count(out.String(), "100% packet loss") > 0 {
		return false
	} else {
		return true
	}
}

// 检测网络
func Check4G() {
	alog.Log.Println("CHECK 4G...")
	if device.IsOnline {
		device.OfflineTimes = 0
	} else {
		device.OfflineTimes = device.OfflineTimes + 1
	}
	if device.OfflineTimes > 20 {
		device.OfflineTimes = 0
		//if err := device.Reset4G(); err == nil {
		if err := device.ResetAll(); err == nil {
			alog.Log.Println("4G RESTART ResetAll")
		} else {
			alog.Log.Println("4G RESTART ResetAll ERROR")
		}
	} else {
		alog.Log.Println("4G OFFTIMES:", device.OfflineTimes)
	}
}

// 检查摄像头并处理
func CheckCamra() {
	if device.GetIsActivityRunning() {
		alog.Log.Println("CHECK CAMERA ACTIVITY_IS_RUNNING")
		return
	}

	alog.Log.Println("CHK CAM...")
	var tempstr string
	cameraCount, _ := strconv.Atoi(public.Config["CAMERA_COUNT"])
	if cameraCount > 4 {
		return
	}
	if public.IsPubDualDoorMod {
		if cameraCount > 2 {
			cameraCount = 4
		} else {
			cameraCount *= 2
		}
	}
	for i := 0; i < cameraCount; i++ {
		re := getCamraStatus(i)
		if re {
			device.IsCamraOn[i] = true
			device.CamraOffTimes[i] = 0
			tempstr += fmt.Sprintf("CAM%d_IS_OK; ", i+1)
			//alog.Log.Printf("CAM %d IS OK\r\n", i+1)
		} else {
			camraNeedRestart := false
			device.IsCamraOn[i] = false
			device.CamraOffTimes[i] = device.CamraOffTimes[i] + 1
			if ((device.CamraOffTimes[i]%10 == 0) && (device.CamraOffTimes[i] < 1000)) ||
				((device.CamraOffTimes[i]%100 == 0) && (device.CamraOffTimes[i] > 1000)) {
				camraNeedRestart = true
				//device.CamraOffTimes[i] = 0
			}

			tempstr += fmt.Sprintf("CAM%d_IS_ER:%d; ", i+1, device.CamraOffTimes[i])
			//alog.Log.Printf("CAM %d IS ERROR:%d\r\n", i+1, device.CamraOffTimes[i])

			if device.GetIsActivityRunning() {
				alog.Log.Println("CheckCamra:GetActivityIsRunning")
				return
			}

			if camraNeedRestart && i+1 != 4 {
				public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_RESETCAMRA, "", "")
				alog.Log.Println("DO RESTART CAMRA", i+1)
				if err := device.ResetCamra(i + 1); err == nil {
					alog.Log.Println("CAMRA_RESTARTED", i+1)
				} else {
					alog.Log.Println("CAMRA_RESTART_ERROR", i+1)
				}
			}
		}
	}
	alog.Log.Println(tempstr)
}

// 检查电源状态
func CheckPower() {
	if device.GetIsActivityRunning() {
		alog.Log.Println("CHECK POWER ACTIVITY_IS_RUNNING")
		return
	}
	alog.Log.Println("CHECK POWER...")
	if !device.IsMPowerOn {
		alog.Log.Println("POWER IS OFF")
		public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_POWEROFF, "", "")
		time.Sleep(3 * time.Second)
		alog.Log.Println("CLOSE POWER")
		if device.GetIsActivityRunning() {
			alog.Log.Println("CheckPower:GetActivityIsRunning")
			return
		}
		if err := device.CloseCPU(); err == nil {
			alog.Log.Println("CLOSE_CPU SUCCESS")
		} else {
			alog.Log.Println("CLOSE_CPU ERROR")
		}
	} else {
		alog.Log.Println("POWER IS ON")
		public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_POWERON, "", "")
	}
}

// 心跳包
func CheckPingpong() {
	//log.Println("CHECK PINGPONG...")
	if err := device.Pingpong(true); err == nil {
		alog.Log.Println("CHK IO SUCCESS")
	} else {
		alog.Log.Println("CHK IO ERROR")
	}
}

// 检查门的状态, 如果门开了锁下来了,拉起锁关门
func CheckDoor() {
	iChkDoorCnt++
	iChkLock1Cnt++
	iChkLock2Cnt++
	if device.GetIsActivityRunning() {
		//ERROR_ACTION_HBT
		if device.IChkTransCnt < 999 {
			device.IChkTransCnt++
		}
		if device.IChkTransCnt%4 == 0 {
			//device.ITransLastTim
			n, _ := strconv.Atoi(public.ERROR_ACTION_HBT)
			newStr := strconv.Itoa(n + device.IChkTransCnt/4)
			public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_REPORT, newStr, device.CurActivityId)
			alog.Log.Println("CheckDoor:", device.CurActivityId, newStr)
		}
		if iChkDoorCnt%256 == 1 {
			alog.Log.Println("CHECK DOOR AND LOCK ACTIVITY_IS_RUNNING")
		}
		return
	}
	if (!device.IsDoorClosed) || (!device.IsLockLocked) {
		if iChkLock1Cnt%256 == 1 {
			if !device.IsDoorClosed {
				public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_ERROR, public.ERROR_NO_ACTIVITY_DOOROPEN, "")
			} else {
				public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_ERROR, public.ERROR_NO_ACTIVITY_LOCKOPEN, "")
			}
		}
		if iChkLock1Cnt%32 == 1 {
			alog.Log.Println("CheckDoor1:", device.IsDoorClosed, device.IsLockLocked)
		}
		if device.GetIsActivityRunning() {
			alog.Log.Println("CheckDoor1:GetActivityIsRunning")
			return
		}
		if (device.IsLockLocked) && (!device.IsDoorClosed) {
			//device.OpenLock()
		}
		if (!device.IsLockLocked) && (device.IsDoorClosed) {
			alog.Log.Println("CheckDoor:close door1")
			device.CloseAllLock(0x01)
		}
	} else {
		iChkLock1Cnt = 0
		if iChkDoorCnt%256 == 1 {
			device.CloseAllLock(0x01)
			alog.Log.Println("CHK DR1 AND LK1 OK")
		}
	}

	if public.IsPubDualDoorMod {
		if (!device.IsDoor2Closed) || (!device.IsLock2Locked) {
			if iChkLock2Cnt%256 == 1 {
				if !device.IsDoor2Closed {
					public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_ERROR, public.ERROR_NO_ACTIVITY_DOOROPEN, "")
				} else {
					public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_ERROR, public.ERROR_NO_ACTIVITY_LOCKOPEN, "")
				}
			}
			if iChkLock2Cnt%32 == 1 {
				alog.Log.Println("CheckDoor2:", device.IsDoor2Closed, device.IsLock2Locked)
			}
			if device.GetIsActivityRunning() {
				alog.Log.Println("CheckDoor2:GetActivityIsRunning")
				return
			}
			if (device.IsLock2Locked) && (!device.IsDoor2Closed) {
				//device.OpenLock()
			}
			if (!device.IsLock2Locked) && (device.IsDoor2Closed) {
				alog.Log.Println("CheckDoor:close door2")
				device.CloseAllLock(0x02)
			}
		} else {
			iChkLock2Cnt = 0
			device.CloseAllLock(0x02)
			if iChkDoorCnt%256 == 1 {
				alog.Log.Println("CHK DR2 AND LK2 OK")
			}
		}
	}
}

// 检查config，如果config文件发生变化，同步到系统
func CheckConfig() {
	device.ITransLastTim++
	device.ICurHourTim++
	device.INoTransTim++
	if device.GetIsActivityRunning() {
		alog.Log.Println("CHECK CONFIG ACTIVITY_IS_RUNNING:", device.IActivityRunningTim)
		device.INoTransTim = 0
		// if device.ITransLastTim > 60*2 {
		// 	device.ITransLastTim = 0
		// 	//report an translasttoolong error
		// 	public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_ERROR, public.ERROR_MACHINE_DOWN4, "")
		// }
		device.IActivityRunningTim++
		if device.IActivityRunningTim > 20 { //force to stop the trans...
			if device.IsDoorClosed && device.IsLockLocked {
				device.IActivityRunningTim = 0
				//stop video and stop trans
				device.StopVideo()
				//find it and delete it
				device.ClearActivity()
				time.Sleep(1 * time.Second)
				device.SetIsActivityRunning(false)
				device.CurActivityId = ""
				public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_ERROR, public.ERROR_MACHINE_DOWN4, "")
				return
			}
		}
		if public.IsSdCardNotFind && public.IsUseTmpFolder {
			siz, _ := public.GetTmpVideoMegaSize()
			if siz > 100 {
				public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_ERROR, public.ERROR_MACHINE_DOWN6, "")
				alog.Log.Println("Stop Video due to /tmp no space!")
				device.StopVideo()
			}
		}
		if device.IActivityRunningTim%60 == 0 {
			alog.Log.Println("Stop Video due to trans time > 1 hour!")
			public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_ERROR, public.ERROR_MACHINE_DOWN4, "")
			device.StopVideo()
		}
		return
	}

	device.CheckIMqttTransFailCnt()

	if device.ICurHourTim >= 60 {
		device.ICurHourTim -= 60
		device.SalesOf24H[device.Iindexof24H] = device.ICurHourTransCnt
		device.ICurHourTransCnt = 0
		device.Iindexof24H++
		if device.Iindexof24H >= 24 {
			device.Iindexof24H = 0
		}

		if public.IsSdCardNotFind && public.IsUseTmpFolder {
			siz, _ := public.GetTmpVideoMegaSize()
			if siz > 80 {
				//send RAM not enough warning msg out
				public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_ERROR, public.ERROR_MACHINE_DOWN7, "")
			}
		}
		//alog.Log.Println("CheckConfig:", device.Iindexof24H, device.ICurHourTransCnt)
	}

	if device.IsOperationDownByLock || device.IsOperationDownByDoor {
		if device.IsDoorClosed && device.IsLockLocked {
			device.IsOperationDownByLock = false
			device.IsOperationDownByDoor = false
			alog.Log.Println("Lock1 and Door1 back to normal")
		}
	}
	if device.IsOperationDownByLock2 || device.IsOperationDownByDoor2 {
		if device.IsDoor2Closed && device.IsLock2Locked {
			device.IsOperationDownByLock2 = false
			device.IsOperationDownByDoor2 = false
			alog.Log.Println("Lock2 and Door2 back to normal")
		}
	}
	public.CheckSdCardMounted()
	public.InitConfig()
	// if !public.IsMountedSdCard {
	// 	log.Println("SD card mount error, pls replace SD card")
	// 	public.SendMqttError(public.TYPE_DEVICE, public.ERROR_MOUNT_SD, "")
	// }
	device.ITransLastTim = 0
	device.IActivityRunningTim = 0
	alog.Log.Println("CHK CFG IDLE:", device.INoTransTim)
	if device.INoTransTim%(60*24) == 0 { //24 hour no trans, reboot
		public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_ERROR, public.ERROR_MACHINE_DOWN5, "")
		alog.Log.Println("12 hours no trans, will reset cpu!")
		time.Sleep(3 * time.Second)
		if device.GetIsActivityRunning() { //check again if any trans going on
			//device.INoTransTim = 0
		} else {
			//device.INoTransTim = 0
			//device.ResetCPU()
		}
	}
}

func MqttHeartbeat() {
	isneedupload := false
	isinfochanged := false
	if device.GetIsActivityRunning() {
		alog.Log.Println("CHECK HEARTBEAT ACTIVITY_IS_RUNNING")
		return
	}
	device.IHeartBeatCnt++
	if device.IHeartBeatCnt <= 1 {
		alog.Log.Println("Ignore first time Heartbeat...")
		return
	} else if device.IHeartBeatCnt == 2 {
		//createSnFile()
		if device.StrSimCardNumb == "" {
			go device.Get4GProvider()
		} else {
			go device.Get4GDataUsage()
		}
	}
	if (device.IHeartBeatCnt < 16) || (device.IHeartBeatCnt%64 == 0) {
		isneedupload = true
		if device.StrSimCardNumb == "" {
			alog.Log.Println("StrSimCardNumb is empty...")
			if device.IHeartBeatCnt%4 == 0 {
				go device.Get4GProvider()
			}
		} else if device.IHeartBeatCnt > 16 {
			go device.Get4GDataUsage()
		}
	}
	alog.Log.Printf("...MQTT HEARTBEAT %d...\r\n", device.IHeartBeatCnt-1)
	var err error
	// send Heartbeat data
	//i, err := strconv.Atoi(data)
	data := make(map[string]string)
	data["Ht"] = strconv.Itoa(device.IHeartBeatCnt - 1)

	if isneedupload {
		data["4G"] = device.StrTelProvider //"G" "L"
		if device.StrSimCardNumb != "" {
			data["MN"] = device.StrSimCardNumb
		}
		if device.StrTxByteNum != "" && device.StrRxByteNum != "" {
			data["TX"] = device.StrTxByteNum
			data["RX"] = device.StrRxByteNum
		}
	}

	cameraCount, _ := strconv.Atoi(public.Config["CAMERA_COUNT"])
	if cameraCount > 4 {
		cameraCount = 4
	}

	i := 0
	for i = 0; i < cameraCount; i++ {
		if isCamraOn[i] != device.IsCamraOn[i] {
			isCamraOn[i] = device.IsCamraOn[i]
			isinfochanged = true
		}
	}
	if isneedupload || isinfochanged {
		isinfochanged = false
		data["Cm1"] = "X"
		data["Cm2"] = "X"
		data["Cm3"] = "X"
		data["Cm4"] = "X"
		for i := 0; i < cameraCount; i++ {
			strCamra := fmt.Sprintf("Cm%d", i+1)
			if device.IsCamraOn[i] {
				data[strCamra] = "Y"
			} else {
				data[strCamra] = "N"
			}
		}
	}

	if isADLive != device.IsADLive {
		isADLive = device.IsADLive
		isinfochanged = true
	}
	if isneedupload || isinfochanged {
		isinfochanged = false
		if device.IsADLive {
			data["Ad"] = "Y"
		} else {
			data["Ad"] = "N"
		}
	}

	if fTemp != device.Temperature {
		fTemp = device.Temperature
		isinfochanged = true
	}
	if isneedupload || isinfochanged {
		isinfochanged = false
		data["Tp"] = strconv.FormatFloat(fTemp, 'f', 2, 64)
	}

	if isDoorClosed != device.IsDoorClosed {
		isinfochanged = true
		isDoorClosed = device.IsDoorClosed
	}
	if isDoor2Closed != device.IsDoor2Closed {
		isinfochanged = true
		isDoor2Closed = device.IsDoor2Closed
	}
	if isneedupload || isinfochanged {
		isinfochanged = false
		if device.IsDoorClosed && device.IsDoor2Closed {
			data["Dr"] = "C"
		} else if device.IsDoorClosed {
			data["Dr"] = "B"
		} else if device.IsDoor2Closed {
			data["Dr"] = "A"
		} else {
			data["Dr"] = "O"
		}
	}

	if isLockLocked != device.IsLockLocked {
		isinfochanged = true
		isLockLocked = device.IsLockLocked
	}
	if isLock2Locked != device.IsLock2Locked {
		isinfochanged = true
		isLock2Locked = device.IsLock2Locked
	}
	if isneedupload || isinfochanged {
		isinfochanged = false
		if device.IsLockLocked && device.IsLock2Locked {
			data["Lk"] = "L"
		} else if device.IsLockLocked {
			data["Lk"] = "B"
		} else if device.IsLock2Locked {
			data["Lk"] = "A"
		} else {
			data["Lk"] = "U"
		}
	}

	if isMountedSdCard != public.IsMountedSdCard {
		isinfochanged = true
		isMountedSdCard = public.IsMountedSdCard
	}
	if isSdCardNotFind != public.IsSdCardNotFind {
		isinfochanged = true
		isSdCardNotFind = public.IsSdCardNotFind
	}
	if isneedupload || isinfochanged {
		isinfochanged = false
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
	}

	if isOperationDownByLock != device.IsOperationDownByLock {
		isinfochanged = true
		isOperationDownByLock = device.IsOperationDownByLock
	}
	if isOperationDownByLock2 != device.IsOperationDownByLock2 {
		isinfochanged = true
		isOperationDownByLock2 = device.IsOperationDownByLock2
	}
	if isOperationDownByDoor != device.IsOperationDownByDoor {
		isinfochanged = true
		isOperationDownByDoor = device.IsOperationDownByDoor
	}
	if isOperationDownByDoor2 != device.IsOperationDownByDoor2 {
		isinfochanged = true
		isOperationDownByDoor2 = device.IsOperationDownByDoor2
	}
	if isneedupload || isinfochanged {
		isinfochanged = false
		if device.IsOperationDownByLock || device.IsOperationDownByDoor ||
			device.IsOperationDownByLock2 || device.IsOperationDownByDoor2 {
			data["St"] = "N"
		} else {
			data["St"] = "Y"
		}
	}
	dataStr, err := json.Marshal(&data)
	if err == nil {
		dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
		public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_STATUS, dataStr2, "")
	}
}

func CheckADLive() {
	if device.GetIsActivityRunning() {
		alog.Log.Println("CHECK ADV ACTIVITY_IS_RUNNING")
		return
	}
	//in /mnt/mmcblk0p1/config/config.ini---check AD IP
	//in /etc/smartshop_go/config/config.ini...
	re := "192.168.199.88"
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", re, "-n", "1")
	} else {
		cmd = exec.Command("ping", re, "-c", "1", "-W", "5")
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
	if strings.Count(out.String(), "100% packet loss") > 0 {
		device.ADOffTimes++
		alog.Log.Println("CHK AD but AD disconnected:", device.ADOffTimes)
		if ((device.ADOffTimes%10 == 0) && (device.ADOffTimes < 1000)) ||
			((device.ADOffTimes%100 == 0) && (device.ADOffTimes > 1000)) {
			public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_RESETAD, "", "")
			device.IsADLive = false
			//device.ADOffTimes = 0
			device.ResetCamra(4)
			alog.Log.Println("AD RESTARTED")
		}
	} else {
		alog.Log.Println("CHK AD OK")
		device.ADOffTimes = 0
		device.IsADLive = true
	}
}
