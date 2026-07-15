package service

import (
	"agent/aliyun/oos"
	"agent/device"
	alog "agent/logger"
	"agent/public"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/asdine/storm/v3"
)

const (
	VIDEO_DELAY_SEC = 12
	FIRST_DELAY_SEC = 16
)

var (
	uploadNow           bool
	activityMutex       sync.Mutex
	wg                  sync.WaitGroup
	iUploadOpenDbErrCnt int
	transStartMutex     sync.Mutex
	//iUploadNotifyErrCnt int
)

func init() {
	//iUploadNotifyErrCnt = 0
	go onUploadChannel(public.ChUpload)
}

func onUploadChannel(ch chan string) {
	var input string
	for {
		input = <-ch
		if input == "N" {
			// 接受上传通知, 立刻上传
			uploadNow = true
		}
	}
}

func UploadStart() {
	alog.Log.Println("UPLOAD START")

	expiredCache := public.NewCache(1)
	for {
		time.Sleep(500 * time.Microsecond)
		if uploadNow {
			// 立刻上传
			uploadNow = false
			expiredCache.SetWithExpiredTime("UPLOADING_IMG", []byte("1"), 10)
			time.Sleep(900 * time.Microsecond)
			upload()
		} else {
			// 每10s 检查一次上传
			if !expiredCache.Exist("UPLOADING_IMG") {
				expiredCache.SetWithExpiredTime("UPLOADING_IMG", []byte("1"), 10)
				upload()
			}
		}
	}
}

func upload() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("UploadStart 捕获异常:", err)
		}
	}()
	alog.Log.Println("upload:cnt", public.IUploadErrCnt)

	if !device.IsMPowerOn {
		// 断电, 退出
		return
	}

	if !public.CheckSdCardAndDbStatus() {
		alog.Log.Println("upload:SD card mount error, pls replace SD card")

		iUploadOpenDbErrCnt++
		if iUploadOpenDbErrCnt > 100 {
			alog.Log.Println("upload:DB File system open err")
			public.IsSdCardNotFind = true
			if public.DB != nil {
				err := public.DB.Close()
				if err != nil {
					// handle error
					alog.Log.Println(err)
				}
				public.DB = nil
			}
		}
		return
	}
	iUploadOpenDbErrCnt = 0

	var uploadInfoList []public.UploadInfo
	//err := public.DB.All(&uploadInfoList, storm.Limit(2))
	err := getUploadInfobyDBAll(&uploadInfoList)
	if err != nil {
		alog.Log.Println("upload:NO MORE UPLOAD INFO NEED UPLOAD。。。", err)
		return
	}
	recordcount := len(uploadInfoList)
	// log.Println(uploadInfoList)
	if recordcount > 0 {
		var uploadInfo public.UploadInfo
		if public.IUploadErrCnt > 10 && public.IsUploadBigfile ||
			public.IUploadErrCnt > 100 { //try to send next file if Uploading took too long a time
			if recordcount > 1 {
				public.IUploadErrCnt = 0
				uploadInfo = uploadInfoList[1]
				alog.Log.Println("upload:switch to 2nd info", uploadInfo.ActivityId)
			} else {
				uploadInfo = uploadInfoList[0]
				alog.Log.Println("upload:switch to 1st info", uploadInfo.ActivityId)
			}
		} else {
			uploadInfo = uploadInfoList[0]
			alog.Log.Println("upload:choose 1st info", uploadInfo.ActivityId)
		}

		iUploadIdLen := len(uploadInfo.ActivityId)
		if iUploadIdLen == 0 {
			alog.Log.Println("upload:find empty ActivityId--delete it")
			//
			err = public.DB.DeleteStruct(&uploadInfo)
			if err != nil {
				alog.Log.Println("数据库错误6", err)
			}
			return
		}

		// 有待上传的文件
		var uploadFileList []public.UploadFile
		//err = public.DB.Find("ActivityId", uploadInfo.ActivityId, &uploadFileList, storm.Limit(10))
		err = getFileListbyDBFind(uploadInfo.ActivityId, &uploadFileList)
		if err != nil {
			public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_UPLOAD7, uploadInfo.ActivityId)
			alog.Log.Println("upload:can not find files to upload。。。", err)
		}

		if iUploadIdLen < 10 || uploadInfo.ActivityId == "1111000011110000000" {
			alog.Log.Println("upload:uploadInfo.ActivityId len is wrong", iUploadIdLen)
			//delete this info or choose 2nd info
			if len(uploadFileList) > 0 {
				for i := 0; i < len(uploadFileList); i++ {
					filePath := public.GetSdcardPath() + "/" + uploadFileList[i].ImgPath

					err = public.DB.DeleteStruct(&uploadFileList[i])
					if err != nil {
						alog.Log.Println("DELETE UPLOAD FILE ERROR3", err)
					}

					alog.Log.Println("Delete rubish file:" + filePath)
					public.DeleteFileOnDisk(filePath) //上传成功后删除本地文件
				}
			}
			err = public.DB.DeleteStruct(&uploadInfo)
			if err != nil {
				alog.Log.Println("数据库错误5", err)
			}
			return
		} else {
			alog.Log.Println("upload:uploadInfo.ActivityId len is ok", iUploadIdLen)
		}

		// log.Println(uploadFileList)
		public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_UPLOADING, "", uploadInfo.ActivityId)

		if len(uploadFileList) > 0 {
			for i := 0; i < len(uploadFileList); i++ {
				path := uploadFileList[i].ImgPath
				realPath := public.GetSdcardPath() + "/" + uploadFileList[i].ImgPath

				// 检查文件有没有问题
				fileError := false
				if !fileError {
					if !public.IsExist(realPath) {
						public.IUploadErrCnt = 0
						public.IsUploadBigfile = false
						public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_UPLOAD4, uploadInfo.ActivityId)
						alog.Log.Println("FILE IS NOT EXIST", realPath)
						fileError = fileError || true
					}
				}

				if !fileError {
					realFileinfo, err := os.Stat(realPath)
					if err != nil {
						public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_UPLOAD4, uploadInfo.ActivityId)
						alog.Log.Println("FILE STATUS ERROR", err)
						fileError = fileError || true
					}
					if !fileError {
						if realFileinfo == nil {
							fileError = fileError || true
						}
					}
					if !fileError {
						videosize := realFileinfo.Size()
						if videosize == 0 {
							public.IUploadErrCnt = 0
							public.IsUploadBigfile = false
							public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_UPLOAD5, uploadInfo.ActivityId)
							alog.Log.Println("FILE SIZE IS 0-deleted:", i)
							public.DeleteFileOnDisk(realPath) //删除本地文件
							fileError = fileError || true
						} else if videosize > 1024*1024*1024 {
							public.IUploadErrCnt = 0
							public.IsUploadBigfile = false
							public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_UPLOAD6, uploadInfo.ActivityId)
							alog.Log.Println("FILE SIZE IS More than 1G-deleted:", i)
							public.DeleteFileOnDisk(realPath) //果断删除本地文件
							fileError = fileError || true
						} else if videosize > 100*1024*1024 {
							public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_UPLOAD6, uploadInfo.ActivityId)
							public.IsUploadBigfile = true
							alog.Log.Println("FILE SIZE IS too big:", videosize, i)
						} else {
							public.IsUploadBigfile = false
							alog.Log.Println("FILE SIZE IS normal:", videosize, i)
						}
					}
				}

				if fileError {
					// 如果图片有问题，删除掉，继续
					err = public.DB.DeleteStruct(&uploadFileList[i])
					if err != nil {
						alog.Log.Println("DELETE UPLOAD FILE ERROR2", err)
					}
					// 如果文件有问题， 尝试重启摄像头
					string_slice := strings.Split(uploadFileList[i].Id, "_")
					if len(string_slice) == 3 {
						n, _ := strconv.Atoi(string_slice[1])
						if n > 3 {
							n = 3
						}
						device.CamraOffTimes[n] = -1
						if n != 3 {
							device.ResetCamra(n + 1)
							alog.Log.Println("VIDEO FILE NOT FOUND, RESET_CAMRA:", n+1)
						}
					}
					continue
				}
				// 检查文件有没有问题

				//	log.Println("SIMPLE UPLOAD")
				//	// 上传到阿里云 简单上传
				//	log.Println("UPLOAD TO OSS:", path, realPath)
				//	err = bucket.PutObjectFromFile(path, realPath)
				//	if err != nil {
				//		public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_UPLOAD1, uploadInfo.ActivityId)
				//		log.Println("UPLOAD OSS ERROR", err)
				//	} else {
				//		// 上传成功！
				//		err = public.DB.DeleteStruct(&uploadFileList[i])
				//		if err != nil {
				//			log.Println("DELETE UPLOAD FILE ERROR", err)
				//		}
				//		public.DeleteFileOnDisk(realPath) //上传成功后删除本地文件
				//	}
				//	// 上传到阿里云 简单上传

				// 上传到阿里云 断点续传
				var bucket *oss.Bucket
				var err3 error
				bucket, err3 = oos.GetSimpleBucket(oos.BucketName)
				if err3 != nil {
					alog.Log.Println("GET BUCKET ERROR", err3)
					continue
				}
				time.Sleep(1 * time.Second)
				var lastFour string
				if len(realPath) > 5 {
					lastFour = realPath[len(realPath)-5 : len(realPath)-4]
				} else {
					lastFour = ""
				}
				err = bucket.UploadFile(path, realPath, 100*1024, oss.Routines(1), oss.Checkpoint(true, fmt.Sprintf("/tmp/%s_%s.cp", uploadInfo.ActivityId, lastFour)))
				if err != nil {
					public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_UPLOAD1, uploadInfo.ActivityId)
					alog.Log.Println("UPLOAD OSS ERROR", err)

					//increase the err counter for 1008 and decrease MTU down to 500
					public.IUploadErrCnt++
					if (public.IUploadErrCnt%10 == 0 && public.IUploadMTUnum > 900) ||
						(public.IUploadErrCnt%30 == 0 && public.IUploadMTUnum > 500) {
						public.IUploadMTUnum -= 100
						strmtu1 := fmt.Sprintf("ifconfig usb0 mtu %d", public.IUploadMTUnum)
						strmtu2 := fmt.Sprintf("ifconfig wwan0 mtu %d", public.IUploadMTUnum)
						public.ExecShell(strmtu1)
						public.ExecShell(strmtu2)
					}
					return
				} else {
					// 上传成功！
					if (public.IUploadErrCnt == 0) && (public.IUploadMTUnum < 1000) {
						public.IUploadMTUnum += 100
						strmtu1 := fmt.Sprintf("ifconfig usb0 mtu %d", public.IUploadMTUnum)
						strmtu2 := fmt.Sprintf("ifconfig wwan0 mtu %d", public.IUploadMTUnum)
						public.ExecShell(strmtu1)
						public.ExecShell(strmtu2)
					}
					public.IUploadErrCnt = 0
					public.IsUploadBigfile = false
					//if upload time is very fast, maybe increase MTU

					err = public.DB.DeleteStruct(&uploadFileList[i])
					if err != nil {
						alog.Log.Println("DELETE UPLOAD FILE ERROR1", err)
					}
					alog.Log.Println("Delete file:" + realPath)
					public.DeleteFileOnDisk(realPath) //上传成功后删除本地文件
					// if i+1 == len(uploadFileList) {
					// 	dirPath := filepath.Dir(realPath)
					// 	log.Println("Delete Parent folder:", dirPath)
					// 	public.DeleteEpyFileFolder(dirPath)
					// }
				}
				// 上传到阿里云 断点续传
			}
		} else {
			public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_UPLOAD8, uploadInfo.ActivityId)
			alog.Log.Println("upload:can not find files to upload2")
		}

		// 检查是否上传完成
		var uploadFileList2 []public.UploadFile
		//err = public.DB.Find("ActivityId", uploadInfo.ActivityId, &uploadFileList2, storm.Limit(10))
		err = getFileListbyDBFind(uploadInfo.ActivityId, &uploadFileList2)
		if err != nil {
			alog.Log.Println("NO MORE FILE to UPLOAD:", err, uploadInfo.ActivityId)
		}

		if len(uploadFileList2) < 1 {
			// 图片已上传完成，通知服务器
			alog.Log.Println("ACTIVITY NOTIFY SEND:", uploadInfo.ImgType)
			if uploadInfo.ImgType == "B" {
				uploadInfo.ImgType = "A"
			}
			source := rand.NewSource(time.Now().UnixNano())
			rng := rand.New(source)
			randomNumber := rng.Intn(100000)
			err, status, res := public.HttpRequest("POST", fmt.Sprintf("http://%s/activity/done", public.API_HOST), map[string]string{"activity_id": uploadInfo.ActivityId,
				"type": uploadInfo.ImgType, "img_count": strconv.Itoa(uploadInfo.ImgAllCount), "rdm": strconv.Itoa(randomNumber)})
			alog.Log.Println("ACTIVITY NOTIFY RETURN:", string(res))
			if err != nil {
				alog.Log.Println("ACTIVITY NOTIFY ERROR", err)
				//do something, add a counter, del uploadInfo
				sdata := fmt.Sprintf("curl -H \"Content-Type:application/json\" -X POST http://%s/activity/done?activity_id=%s&type=%s&img_count=%d",
					public.API_HOST, uploadInfo.ActivityId, uploadInfo.ImgType, uploadInfo.ImgAllCount)
				err, re := public.ExecShell(sdata)
				alog.Log.Println(err)
				alog.Log.Println(re)
				if strings.Contains(re, "SUCCESS") {
					err = public.DB.DeleteStruct(&uploadInfo)
					if err != nil {
						alog.Log.Println("数据库错误5", err)
					}
					alog.Log.Println("Curl OK")
				} else {
					alog.Log.Println("Crul Err")
				}
				public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_UPLOAD9, uploadInfo.ActivityId)
			} else {
				if status == 200 {
					var re public.ApiReturn
					json.Unmarshal(res, &re)
					if re.Code == 1 {
						//已经通知完。删除掉
						err = public.DB.DeleteStruct(&uploadInfo)
						if err != nil {
							alog.Log.Println("数据库错误4", err)
						}
						//iUploadNotifyErrCnt = 0
						//public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_UPLOADED, "", item["activity_id"])
					} else {
						public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_UPLOAD2, uploadInfo.ActivityId)
						alog.Log.Println("ACTIVITY RETURN CODE IS NOT 1")
					}
				} else {
					public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_UPLOAD3, uploadInfo.ActivityId)
					alog.Log.Println("ACTIVITY RETURN STATUS IS NOT 200")
				}
			}
		}
	} else {
		if !device.GetIsActivityRunning() {
			if device.IRAMNoSpaceCnt > 2 {
				device.IRAMNoSpaceCnt = 0
				os.RemoveAll("/tmp/smartshop/video")
				alog.Log.Println("upload:clean RAM video!!!")
			}
		}
	}
}

func getUploadInfobyDBAll(puploadInfoList *[]public.UploadInfo) error {
	// Acquire the mutex lock before accessing the database
	device.DBMutex.Lock()
	defer device.DBMutex.Unlock()
	//var uploadInfoList []public.UploadInfo
	if err := public.DB.All(puploadInfoList, storm.Limit(2)); err != nil {
		//alog.Log.Println("upload:NO MORE UPLOAD INFO NEED UPLOAD。。。", err)
		return err
	}
	alog.Log.Println("upload:InfoList", len(*puploadInfoList))
	return nil
}

func getFileListbyDBFind(activityid string, puploadFileList *[]public.UploadFile) error {
	device.DBMutex.Lock()
	defer device.DBMutex.Unlock()
	//var uploadFileList []public.UploadFile
	if err := public.DB.Find("ActivityId", activityid, puploadFileList, storm.Limit(10)); err != nil {
		//alog.Log.Println("NO MORE FILE to UPLOAD:", err, activityid)
		return err
	}
	alog.Log.Println("upload:FileList", len(*puploadFileList))
	return nil
}
