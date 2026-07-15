package public

import (
	alog "agent/logger"
	"os"
	"sync"
	"time"

	"github.com/asdine/storm/v3"
	bolt "go.etcd.io/bbolt"
)

type ActivityInfo struct {
	ActivityId string `storm:"id"`
	Step       int
	TryTimes   int
}

type UploadInfo struct {
	Id             string `storm:"id"`
	ActivityId     string `storm:"index"`
	ImgAllCount    int    `storm:"index"`
	ImgUploadCount int    `storm:"index"`
	ImgType        string
	IsNotified     bool `storm:"index"`
}

type UploadFile struct {
	Id          string `storm:"id"`
	InfoId      string `storm:"index"`
	ActivityId  string `storm:"index"`
	ImgPath     string
	ImgUploaded bool `storm:"index"`
}

var (
	DB            *storm.DB
	iDBopenerrCnt int
	dbmutex       sync.Mutex
)

func OpenDB() bool {
	defer func() {
		if err := recover(); err != nil {
			IsSdCardNotFind = true
			alog.Log.Println("OpenDB Error:", err)
		}
	}()
	var err error = nil
	//, storm.BoltOptions(0755, &bolt.Options{Timeout: 1 * time.Second, NoFreelistSync: true})
	//https://dev.to/go/using-boltdb-as-internal-database-39bd
	// Set the DB.OpenTimeout option to 5 seconds.
	// Set a timeout of 5 seconds
	opts := storm.BoltOptions(0666, &bolt.Options{Timeout: 5 * time.Second}) //, NoFreelistSync: true})
	alog.Log.Println("Open:", GetSdcardPath()+"/cache/my.db")
	DB, err = storm.Open(GetSdcardPath()+"/cache/my.db", opts)
	if err != nil {
		iDBopenerrCnt++
		//log.Println(GetSdcardPath() + "/cache/my.db")
		alog.Log.Println("Open my.db Error:", iDBopenerrCnt, err)
		if iDBopenerrCnt > 5 {
			if iDBopenerrCnt > 10 {
				if os.IsPermission(err) {
					alog.Log.Println("OpenDB:File system is read-only")
					IsSdCardNotFind = true
					if DB != nil {
						err := DB.Close()
						if err != nil {
							// handle error
							alog.Log.Println(err)
						}
						DB = nil
					}
				} else {
					alog.Log.Println("OpenDB:Exit ss_main!!!")
					os.Exit(0)
				}
			}
			return false
		}
		//DB.Close() //if add this, will not return!!
	} else {
		alog.Log.Println("my.db Opened!")
		iDBopenerrCnt = 0
	}

	//defer DB.Close() //can't use this!!
	return true
}

func CheckSdCardAndDbStatus() bool {
	dbmutex.Lock()
	defer dbmutex.Unlock()
	CheckSdCardMounted()
	if IsMountedSdCard && !IsSdCardNotFind {
		INoSdCardCnt = 0
		IsUseTmpFolder = false
		alog.Log.Println("CkSdAndDb:SD ok")
	} else {
		if (INoSdCardCnt%10 == 0 && INoSdCardCnt < 1000) ||
			(INoSdCardCnt%1000 == 0 && INoSdCardCnt > 1000) {
			SendMqttError(TYPE_DEVICE, ERROR_MOUNT_SD, "")
			alog.Log.Println("CkSdAndDb:SD err")
		}
		INoSdCardCnt++
		if IsSdCardNotFind == false {
			IsSdCardNotFind = true
			if DB != nil {
				err := DB.Close()
				if err != nil {
					// handle error
					alog.Log.Println(err)
				}
				DB = nil
			}
		} else {
			alog.Log.Println("CkSdAndDb:use tmp")
		}
	}

	if DB == nil {
		dbIsOpen := OpenDB()
		if dbIsOpen {
			// 数据库打开成功
			if IsMountedSdCard && !IsSdCardNotFind {
				IsUseTmpFolder = false
			} else {
				IsUseTmpFolder = true
			}
			IUploadErrCnt = 0
			IsUploadBigfile = false
			IUploadMTUnum = 1300
			return true
		} else {
			// 数据库打开失败
			SendMqttError(TYPE_DEVICE, ERROR_OPEN_DB, "")
			//return false
			alog.Log.Println("CheckSdCardAndDbStatus:DB not open")
			time.Sleep(2 * time.Second)
			//os.Exit(0)
			return false
		}
	} else {
		return true
	}
}
