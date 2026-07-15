package service

import (
	"agent/device"
	alog "agent/logger"
	"agent/public"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/asdine/storm/v3"
)

func init() {
	go onActivityChannel(public.ChActivity)
}

func onActivityChannel(ch chan string) {
	var input string
	for {
		input = <-ch
		inputs := strings.Split(input, ":")
		if len(inputs) == 3 && inputs[0] == public.CHANNEL_TYPE_ACTIVITY {
			if inputs[1] == public.ACTION_BEGIN {
				activityId := inputs[2]
				go ActivityStartNow(
					public.ActivityInfo{
						ActivityId: activityId,
						Step:       0,
						TryTimes:   0,
					}, "1")
			}
		}
		if len(inputs) == 4 && inputs[0] == public.CHANNEL_TYPE_ACTIVITY {
			if inputs[1] == public.ACTION_BEGIN {
				activityId := inputs[2]
				needUploadVideo := inputs[3]
				if needUploadVideo != "0" && needUploadVideo != "1" && needUploadVideo != "3" {
					needUploadVideo = "1"
				}
				go ActivityStartNow(
					public.ActivityInfo{
						ActivityId: activityId,
						Step:       0,
						TryTimes:   0,
					}, needUploadVideo)
			}
		}
		if len(inputs) == 4 && inputs[0] == public.CHANNEL_TYPE_ACTIVITYEN {
			if inputs[1] == public.ACTION_BEGIN {
				activityId := inputs[2]
				needUploadVideo := inputs[3]
				if needUploadVideo == "0" {
					needUploadVideo = "4" //no need upload video and use 2nd door
				} else {
					needUploadVideo = "5" //need upload video and use 2nd door
				}
				go ActivityStartNow(
					public.ActivityInfo{
						ActivityId: activityId,
						Step:       0,
						TryTimes:   0,
					}, needUploadVideo)
			}
		}
		time.Sleep(10 * time.Microsecond)
	}
}

func ActivityStartNow(activityInfo public.ActivityInfo, needUploadVideo string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("ActivityStartNow 捕获异常:", err)
		} else {
			device.SetIsActivityQueing(false)
		}
	}()

	if device.GetIsActivityQueing() {
		device.SetIMqttTransFailCnt(0)
		// 如果当前有活动在QUE，返回错误
		alog.Log.Println("ActivityStartNow:ACITIVITY IN_QUEUE")
		public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_INUSE, activityInfo.ActivityId)
		return
	}
	device.SetIMqttTransFailCnt(0)
	device.SetIsActivityQueing(true)

	alog.Log.Println("ACITIVITY START NOW")
	if activityInfo.ActivityId == "" {
		alog.Log.Println("ACITIVITY ID IS ''")
		return
	}
	//make sure the activityid is correct length first
	iActLen := len(activityInfo.ActivityId)
	if iActLen < 10 {
		alog.Log.Println("ActivityStartNow:length err ", iActLen)
		return
	}
	if device.GetIsActivityRunning() {
		// 如果当前有活动在进行，返回错误
		alog.Log.Println("ActivityStartNow:ACITIVITY ERROR_INUSE")
		public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_INUSE, activityInfo.ActivityId)
		return
	}
	if needUploadVideo == "4" || needUploadVideo == "5" {
		if device.IsOperationDownByLock2 {
			alog.Log.Println("ActivityStartNow:ACITIVITY ERROR_CLOSE2")
			public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_CLOSE, activityInfo.ActivityId)
			//return
		}
		if device.IsOperationDownByDoor2 {
			alog.Log.Println("ActivityStartNow:ACITIVITY ERROR_DOOR_OPEN2")
			public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_DOOR_OPEN, activityInfo.ActivityId)
			//return
		}
	} else {
		if device.IsOperationDownByLock {
			alog.Log.Println("ActivityStartNow:ACITIVITY ERROR_CLOSE1")
			public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_CLOSE, activityInfo.ActivityId)
			//return
		}
		if device.IsOperationDownByDoor {
			alog.Log.Println("ActivityStartNow:ACITIVITY ERROR_DOOR_OPEN1")
			public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_DOOR_OPEN, activityInfo.ActivityId)
			//return
		}
	}

	// 判断数据库是否打开，没有打开的话，打开数据库
	if !public.CheckSdCardAndDbStatus() {
		alog.Log.Println("ActivityStartNow:OPEN DB ERROR 2")
		public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_MOUNT_SD, activityInfo.ActivityId)
		return
	}
	if public.DB == nil {
		alog.Log.Println("ActivityStartNow:DB not exist!")
		public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_OPEN_DB, activityInfo.ActivityId)
		return
	}
	if public.IsUseTmpFolder {
		siz, _ := public.GetTmpVideoMegaSize()
		//check remaing space inside tmp folder, then give a max time for video capture.
		if siz > 88 {
			alog.Log.Println("ActivityStartNow:tmp alomst full:", siz)
			public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_MACHINE_DOWN6, activityInfo.ActivityId)
			device.IRAMNoSpaceCnt++
			return
		} else {
			device.IRAMNoSpaceCnt = 0
		}
	}

	//check if such activityInfo.ActivityId already in my.db or not
	alog.Log.Println("ActivityStartNow:check listinfo")
	var activeInfo []public.ActivityInfo
	err := public.DB.Find("ActivityId", activityInfo.ActivityId, &activeInfo, storm.Limit(1))
	if err == nil {
		alog.Log.Println("ActivityStartNow:find duplicated activeInfo", activityInfo.ActivityId)
		public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_DUPL_RECORD, activityInfo.ActivityId)
		return
	}

	alog.Log.Println("ActivityStartNow:check uploadinfo")
	var uploadInfo []public.UploadInfo
	err = public.DB.Find("ActivityId", activityInfo.ActivityId, &uploadInfo, storm.Limit(1))
	if err == nil {
		alog.Log.Println("ActivityStartNow:find duplicated uploadInfo", activityInfo.ActivityId)
		public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_DUPL_RECORD, activityInfo.ActivityId)
		return
	} //If Find more than 1, this err return will not be nil! use[]

	alog.Log.Println("ActivityStartNow:check filelist")
	var uploadFileList []public.UploadFile
	err = public.DB.Find("ActivityId", activityInfo.ActivityId, &uploadFileList, storm.Limit(1))
	if err == nil {
		alog.Log.Println("ActivityStartNow:find duplicated files to upload", activityInfo.ActivityId)
		public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_DUPL_RECORD, activityInfo.ActivityId)
		return
	}

	//in case 2 trans come in the same time, let 1 trans done first.
	transStartMutex.Lock()
	defer transStartMutex.Unlock()

	device.IActivityRunningTim = 0
	device.IChkTransCnt = 0
	device.ITransLastTim = 1
	if needUploadVideo == "4" || needUploadVideo == "5" {
		activityInfo.TryTimes = 10000
	}
	device.MqttActivityId = activityInfo.ActivityId
	alog.Log.Println("ActivityStartNow:save activityInfo:", device.MqttActivityId, needUploadVideo)
	public.DB.Save(&activityInfo)
	// 将活动插入到数据库
	ActivityStart(needUploadVideo)
}

func ActivityStart(needUploadVideo string) {
	for {
		// 判断数据库是否打开，没有打开的话，打开数据库
		if !public.CheckSdCardAndDbStatus() {
			alog.Log.Println("ActivityStart:OPEN DB ERROR 1")
			setEndActivity()
			return
		}
		if public.DB == nil {
			alog.Log.Println("ActivityStart:DB not exist!")
			setEndActivity()
			return
		}
		// if public.IsUseTmpFolder {
		// 	//check remaing space inside tmp folder, then give a max time for video capture.
		// 	if public.GetTmpVideoMegaSize() > 80 {
		// 		log.Println("ActivityStart:tmp alomst full:", public.GetTmpVideoMegaSize())
		//		setEndActivity()
		// 		return
		// 	}
		// }
		var videostartTime, videostopTime int64
		// 执行任务
		var activityInfoList []public.ActivityInfo
		var activityInfo public.ActivityInfo
		err := public.DB.All(&activityInfoList, storm.Limit(1))
		if err != nil {
			device.ICurStep = 0
			alog.Log.Println("HAVE NO ACTIVITY NEED TO DEAL...", err)
			setEndActivity()
			return
		} else {
			if len(activityInfoList) > 0 {
				activityInfo = activityInfoList[0]
				if device.GetIsActivityRunning() && activityInfo.Step == 0 {
					// 如果当前有活动在进行，返回错误
					if device.IsVideoStopped == false {
						device.IsVideoStopped = true
						defer time.AfterFunc(10*time.Second, func() {
							if !device.GetIsActivityRunning() {
								device.StopVideo()
							}
						})
					}
					if err := public.DB.DeleteStruct(&activityInfo); err != nil {
						alog.Log.Println("ACTIVITY DELETE ERROR3:", err)
					}
					alog.Log.Println("ACITIVITY ERROR_INUSE AND END")
					public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_INUSE, activityInfo.ActivityId)
					setEndActivity()
					return
				}

				//need check activityInfoList[0].ActivityId
				activityInfo = activityInfoList[0]
				if needUploadVideo == "2" {
					device.MqttActivityId = activityInfo.ActivityId
				}
				iLen := len(activityInfo.ActivityId)
				if iLen < 10 {
					public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_IDLEN_SHORT, activityInfo.ActivityId)
					alog.Log.Println("ActivityStart: Id len err", activityInfo.ActivityId, iLen)
					if err := public.DB.DeleteStruct(&activityInfo); err != nil {
						alog.Log.Println("ACTIVITY DELETE ERROR2:", err)
					}
					//setEndActivity()
					continue
				}

				if device.MqttActivityId != activityInfo.ActivityId {
					public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_ID_NOTMATCH, activityInfo.ActivityId)
					alog.Log.Println("ActivityStart: abnormal Id", activityInfo.ActivityId, device.MqttActivityId)
					if err := public.DB.DeleteStruct(&activityInfo); err != nil {
						alog.Log.Println("ACTIVITY DELETE ERROR3:", err)
					}
					//setEndActivity()
					continue
				}

				device.IsDualDoorEnabled = false
				if activityInfo.TryTimes >= 10000 || needUploadVideo == "4" || needUploadVideo == "5" {
					device.IsDualDoorEnabled = true
				}
				setStartActivity(activityInfo.ActivityId)
				//log.Println("ACTIVITY start:" + device.CurActivityId)
			} else {
				device.ICurStep = 0
				alog.Log.Println("HAVE NO ACTIVITY NEED TO DEAL。。。")
				setEndActivity()
				return
			}
		}

		//activityInfo = activityInfoList[0]
		device.Fulsh()

		// 锁上电
		if err := device.OpenLockPower(); err != nil {
			alog.Log.Println("ACTIVITY ERROR POWER ON:", err)
		}

		device.INoTransTim = 0
		curStep := activityInfo.Step

		// 未完成的任务重试 超过10 次，发送提醒消息给服务器
		if (activityInfo.TryTimes > 10 && activityInfo.TryTimes < 10000) || activityInfo.TryTimes > 10010 {
			public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_RETRY_TIMEOUT, activityInfo.ActivityId)
			if device.IsVideoStopped == false {
				device.IsVideoStopped = true
				device.StopVideo()
				public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_SNAPED, "", activityInfo.ActivityId)
			}
			if activityInfo.TryTimes < 10000 {
				activityInfo.TryTimes = 0
			} else {
				activityInfo.TryTimes = 10000
			}

			setActivityStep(activityInfo.ActivityId, 2)
			curStep = 2
			alog.Log.Println("步骤2+:", device.ICurStep)
		}

		device.ICurStep = curStep
		alog.Log.Println("ActivityStart:step =", device.ICurStep, activityInfo.TryTimes)
		//device.INoTransTim = 0
		// 步骤1，拍照，开门
		if curStep < 1 {
			public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_SNAPING, "", activityInfo.ActivityId)
			if needUploadVideo == "1" || needUploadVideo == "2" || needUploadVideo == "3" || needUploadVideo == "5" {
				if err := device.StartVideo(); err != nil {
					public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_SNAP_BEFORE, activityInfo.ActivityId)
					alog.Log.Println("ACTIVITY ERROR SNAPSHOT BEFORE:", err, device.ITakeVideoFailCnt)
					device.StopVideo()
					device.ITakeVideoFailCnt++
					if device.ITakeVideoFailCnt > 3 {
						//device.ITakeVideoFailCnt = 0
						public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_MACHINE_DOWN1, activityInfo.ActivityId)
						//reset cpu here.
						time.Sleep(3 * time.Second)
						if err := public.DB.DeleteStruct(&activityInfo); err != nil {
							alog.Log.Println("ACTIVITY DELETE ERROR:", err)
						}

						if device.IMkdirErrCnt > 3 {
							oldFilePath := "/mnt/mmcblk0p1/sd.txt"
							newFilePath := "/mnt/mmcblk0p1/sd1.txt"
							err := os.Rename(oldFilePath, newFilePath)
							if err != nil {
								alog.Log.Println("Error renaming file sd.txt:", err)
							} else {
								alog.Log.Println("Renaming file sd.txt to sd1.txt!")
							}
							time.Sleep(2 * time.Second)
						}
						device.ResetCPU()
					}
					goto OVER
				}
				device.ITakeVideoFailCnt = 0
				videostartTime = time.Now().Unix()
				public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_SNAPED, "", activityInfo.ActivityId)
			} else {
				setActivityStep(activityInfo.ActivityId, 1)
			}
			// if err, _ := device.SnapShotUpload(activityInfo.ActivityId, public.BF_BEFORE); err != nil {
			// 	public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_SNAP_BEFORE, activityInfo.ActivityId)
			// 	log.Println("ACTIVITY ERROR SNAPSHOT BEFORE:", err)
			// 	goto OVER
			// }
			device.BLastDoorStatus = !device.IsDoorClosed
			device.BLastDoor2Status = !device.IsDoor2Closed
			if (device.BLastDoorStatus && !device.IsDualDoorEnabled) || (device.BLastDoor2Status && device.IsDualDoorEnabled) {
				public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_DOOROPENED, "", activityInfo.ActivityId)
			}

			public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_OPENING, "", activityInfo.ActivityId)
			//add 1 sec delay to make sure FFmpeg is starting..
			if (device.IsFirstTimeRun && !device.IsDualDoorEnabled) ||
				(device.IsFirstTimeDualRun && device.IsDualDoorEnabled) {
				time.Sleep(2 * time.Second)
			}
			if needUploadVideo != "3" {
				if err := device.OpenLock(); err != nil {
					public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_OPEN, activityInfo.ActivityId)
					alog.Log.Println("ACTIVITY ERROR OPENLOCK:", err)
					goto OVER
				}
			}

			if needUploadVideo == "1" || needUploadVideo == "2" || needUploadVideo == "3" || needUploadVideo == "5" {
				setActivityStep(activityInfo.ActivityId, 1)
			}
			device.IChkTransCnt = 0
			curStep = 1
			device.ICurStep = curStep
			alog.Log.Println("步骤1:", device.ICurStep)
			public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_OPENED, "", activityInfo.ActivityId)
		}

		// 步骤2，等待关门
		if curStep < 2 {
			if err := device.WaitCloseDoor(activityInfo.ActivityId); err != nil {
				if activityInfo.TryTimes == 0 {
					public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_CLOSE, activityInfo.ActivityId)
				}
				alog.Log.Println("ACTIVITY ERROR WAITCLOSEDOOR:", err, activityInfo.TryTimes)
				//give another 10min to check if door really closed, if not, report error
				//and kill ffmpeg by force. //will kill a new trans's video, but whatever!
				if device.IsVideoStopped == false {
					device.IsVideoStopped = true
					time.AfterFunc(10*time.Minute, func() {
						if activityInfo.ActivityId == getCurrentActivityId() {
							device.StopVideo()
							public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_SNAPED, "", activityInfo.ActivityId)
						}
					})
				}
				if device.IsDualDoorEnabled {
					device.IsOperationDownByLock2 = true
				} else {
					device.IsOperationDownByLock = true
				}
				goto OVER
			}
			setActivityStep(activityInfo.ActivityId, 2)
			curStep = 2
			device.ICurStep = curStep
			alog.Log.Println("步骤2:", device.ICurStep)
			public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_CLOSED, "", activityInfo.ActivityId)
		}

		// 步骤3，拍照
		if curStep < 3 {
			public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_SNAPING, "", activityInfo.ActivityId)
			time.Sleep(1 * time.Second)
			if needUploadVideo == "1" || needUploadVideo == "2" || needUploadVideo == "3" || needUploadVideo == "5" {
				videostopTime = time.Now().Unix()
				waitCnt := 0
				var difsec int64
				if (device.IsFirstTimeRun && !device.IsDualDoorEnabled) ||
					(device.IsFirstTimeDualRun && device.IsDualDoorEnabled) {
					difsec = FIRST_DELAY_SEC
				} else {
					difsec = VIDEO_DELAY_SEC
				}
				if device.IsDualDoorEnabled {
					device.IsFirstTimeDualRun = false
				} else {
					device.IsFirstTimeRun = false
				}

				if (videostopTime-videostartTime < difsec) && (curStep == 2) {
					alog.Log.Println("Video less than regular:", videostopTime-videostartTime, difsec)
					for i := videostopTime; i < videostartTime+difsec; i++ {
						time.Sleep(1 * time.Second)
						waitCnt++
						if waitCnt > FIRST_DELAY_SEC {
							break
						}
					}
					alog.Log.Println("Video less than regular:Done")
				} //less than 10s, just wait.
				if err := device.StopVideo(); err != nil {
					public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_SNAP_AFTER, activityInfo.ActivityId)
					alog.Log.Println("ACTIVITY ERROR SNAPSHOT AFTER:", err)
					goto OVER
				}
			}

			// if err, _ := device.SnapShotUpload(activityInfo.ActivityId, public.BF_AFTER); err != nil {
			// 	public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_SNAP_AFTER, activityInfo.ActivityId)
			// 	log.Println("ACTIVITY ERROR SNAPSHOT AFTER:", err)
			// 	goto OVER
			// }

			setActivityStep(activityInfo.ActivityId, 3)
			curStep = 3
			device.ICurStep = curStep
			alog.Log.Println("步骤3:", device.ICurStep)
			public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_SNAPED, "", activityInfo.ActivityId)
		}
		goto OVER
	OVER:
		// 保险起见, 再次关锁
		time.Sleep(100 * time.Millisecond)
		device.CloseLock()
		// 锁断电
		if err := device.CloseLockPower(); err != nil {
			alog.Log.Println("ACTIVITY ERROR POWER OFF:", err)
		}
		// 如果当前步骤等于0 或者 等于3，删除掉当前任务 || activityInfo.TryTimes > 10
		if curStep == 0 || curStep == 3 {
			if curStep == 3 {
				// 在补充一次停止录像，如果（0：开门出错，未开始开门，3：成功开门）
				if err := device.StopVideo(); err != nil {
					public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_SNAP_AFTER, activityInfo.ActivityId)
				}
				if err := public.DB.DeleteStruct(&activityInfo); err != nil {
					alog.Log.Println("ACTIVITY DELETE ERROR1:", err)
				}
				if device.IsDualDoorEnabled {
					device.IsFirstTimeDualRun = false
				} else {
					device.IsFirstTimeRun = false
				}
				device.ITotalTransCnt++
				if needUploadVideo == "1" || needUploadVideo == "5" {
					device.ICurHourTransCnt++
					device.SalesOf24H[device.Iindexof24H] = device.ICurHourTransCnt
				}
				//log.Println("ACTIVITY Upld:", activityInfo.ActivityId, device.CurActivityId)
				device.AddUpladVideo(needUploadVideo)
				//log.Println("ACTIVITY DONE:", activityInfo.ActivityId, device.CurActivityId)
				public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_DONE, "", activityInfo.ActivityId)
			} else {
				isLockeUnlocked := false
				for ii := 0; ii < 20; ii++ {
					time.Sleep(50 * time.Millisecond)
					device.CloseLock()
					time.Sleep(50 * time.Millisecond)
					device.Pingpong(true)
					//just wait about 5s to make sure the door is Locked
					if (device.IsDualDoorEnabled && device.IsDoor2Closed && device.IsLock2Locked) ||
						(!device.IsDualDoorEnabled && device.IsDoorClosed && device.IsLockLocked) {
						isLockeUnlocked = false
						alog.Log.Println("ACTIVITY Fail:", "locked the door", ii)
					} else {
						isLockeUnlocked = true
						alog.Log.Println("ACTIVITY Resm:", "locked the door", ii)
						break
					}
				}
				if isLockeUnlocked == false {
					// 在补充一次停止录像，如果（0：开门出错，未开始开门，3：成功开门）
					if err := device.StopVideo(); err != nil {
						public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_SNAP_AFTER, activityInfo.ActivityId)
					}
					if err := public.DB.DeleteStruct(&activityInfo); err != nil {
						alog.Log.Println("ACTIVITY DELETE ERROR1:", err)
					}
					device.DelUnusedVideo()
					alog.Log.Println("ACTIVITY Fail:", activityInfo.ActivityId)
				} else {
					if needUploadVideo == "1" || needUploadVideo == "2" || needUploadVideo == "3" || needUploadVideo == "5" {
						setActivityStep(activityInfo.ActivityId, 1)
					}
					curStep = 1
					device.ICurStep = curStep
					alog.Log.Println("步骤1+:", device.ICurStep)
					public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_OPENED, "", activityInfo.ActivityId)

					alog.Log.Println("ACTIVITY Resm:", activityInfo.ActivityId, activityInfo.TryTimes)
					activityInfo.TryTimes++
					setActivityTrytimes(activityInfo.ActivityId, activityInfo.TryTimes)
				}
			}
			device.INoTransTim = 0
		} else {
			activityInfo.TryTimes++
			setActivityTrytimes(activityInfo.ActivityId, activityInfo.TryTimes)
		}
		time.Sleep(1 * time.Second)
	}
}

func setStartActivity(acitivityId string) {
	activityMutex.Lock()
	defer activityMutex.Unlock()
	//============================================================
	device.SetIsActivityRunning(true)
	device.SetIsActivityQueing(false)
	device.SetIMqttTransFailCnt(0)
	device.CurActivityId = acitivityId
	//log.Println("ACTIVITY RUNNING:", acitivityId)
	alog.Log.Println("ACTIVITY START:"+acitivityId, device.IsDualDoorEnabled)
	//============================================================
}

func setEndActivity() {
	activityMutex.Lock()
	defer activityMutex.Unlock()
	//============================================================
	alog.Log.Println("ACTIVITY END1")
	wg.Add(1)
	go setEndActivityAndSleep()
	wg.Wait()
	//============================================================
}
func setEndActivityAndSleep() {
	defer wg.Done()
	time.Sleep(1 * time.Second)
	device.SetIsActivityRunning(false)
	device.CurActivityId = ""
	device.MqttActivityId = ""
	alog.Log.Println("ACTIVITY END2")
}
func setActivityTrytimes(acitivityId string, tryTimes int) error {
	err := public.DB.UpdateField(&public.ActivityInfo{ActivityId: acitivityId}, "TryTimes", tryTimes)
	return err
}
func setActivityStep(acitivityId string, step int) error {
	err := public.DB.UpdateField(&public.ActivityInfo{ActivityId: acitivityId}, "Step", step)
	return err
}

// func StartActivityAftReboot() {
// 	ActivityStart("2")
// 	if device.IsFirstTimeRun {
// 		InitFFmpeg()
// 	}
// }

func getCurrentActivityId() string {
	return device.CurActivityId
}
