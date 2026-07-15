package device

import (
	alog "agent/logger"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"agent/public"

	"github.com/asdine/storm/v3"
	"github.com/tarm/serial"
)

var (
	CurActivityId  string
	MqttActivityId string

	IsOnline         bool // 是否在线
	IsActivityRuning bool // activity是否正在运行
	IsActivityQueing bool
	IsSerialing      bool // 是否正在串口通讯中
	IsLog            bool // 是否记录log

	IsMPowerOn    bool    // 是否有电（交流电）
	IsCamraOn     [4]bool // 摄像头1状态
	IsBuzzerOn    bool    // 蜂鸣器状态
	Is4GPowOn     bool    // 4G POWER状态
	IsLockLocked  bool    // 锁状态
	IsDoorClosed  bool    // 门状态
	IsLock2Locked bool    // 锁2状态
	IsDoor2Closed bool    // 门2状态

	IsADLive   bool //AD board has network response
	ADOffTimes int

	CamraOffTimes [4]int
	OfflineTimes  int // 设备离线次数

	Temperature float64 // 温度
	Version     string  //版本号

	lastData         []byte          // 上次数据
	isCmdReturn      map[string]bool // 命令是否已经返回
	sendSleepTime    time.Duration
	recvSleepTime    time.Duration
	recvSleepTimeOut time.Duration
	p                *serial.Port
	dtel             *serial.Port
	etel             *serial.Port
	fintel           *serial.Port
	isportusb0       bool
	ctxVideo         context.Context
	videoCmds        map[int]*exec.Cmd
	ipingpongcnt     int

	IsVideoStopped   bool
	ICurStep         int
	BLastDoorStatus  bool //last time the door1 status
	BLastDoor2Status bool //last time the door2 status
	IHeartBeatCnt    int
	IRAMNoSpaceCnt   int

	INoTransTim       int //No trans counter, if more than 12 hours, will reset ALL
	ITransLastTim     int //Transaction last time by mins
	IOpenDoorFailCnt  int //Door not open failed times
	ITakeVideoFailCnt int //Door not open failed times
	ITotalTransCnt    int
	I24HTransCnt      int
	SalesOf24H        [30]int
	Iindexof24H       int
	ICurHourTim       int
	ICurHourTransCnt  int
	IChkTransCnt      int

	IActivityRunningTim int
	IsFirstTimeRun      bool
	IsFirstTimeDualRun  bool
	pingpong_mutex      sync.Mutex
	stopvideo_mutex     sync.Mutex
	cmdreturn_mutex     sync.Mutex
	DBMutex             sync.Mutex
	atcmd_mutex         sync.Mutex
	queMutex            sync.Mutex

	IsOperationDownByLock  bool //stop Fridge1 operation
	IsOperationDownByDoor  bool //stop Fridge1 operation
	IsOperationDownByLock2 bool //stop Fridge2 operation
	IsOperationDownByDoor2 bool //stop Fridge2 operation

	IsDualDoorEnabled bool
	CameraFileSeq     [4]int
	StrTelProvider    string
	StrSimCardNumb    string
	StrTxByteNum      string
	StrRxByteNum      string

	IMqttTransFailCnt int
	IMkdirErrCnt      int
)

var cb_recbuf = NewCircularBuffer(256)

type CircularBuffer struct {
	buffer []byte
	head   int
	tail   int
	size   int
}

func NewCircularBuffer(size int) *CircularBuffer {
	return &CircularBuffer{
		buffer: make([]byte, size),
		head:   0,
		tail:   0,
		size:   size,
	}
}

func (cb *CircularBuffer) Put(data byte) {
	cb.buffer[cb.head] = data
	cb.head = (cb.head + 1) % cb.size
	if cb.head == cb.tail {
		cb.tail = (cb.tail + 1) % cb.size
	}
}

func (cb *CircularBuffer) Get() (byte, error) {
	if cb.head == cb.tail {
		//fmt.Println("CircularBuffer is empty!")
		return 0, errors.New("empty")
	}
	data := cb.buffer[cb.tail]
	cb.tail = (cb.tail + 1) % cb.size
	return data, nil
}

func init() {
	alog.Log.Println("device.go" + "init()")
	serialInit()
	go vserialRec()
	defer time.AfterFunc(1*time.Second, func() { go serialListen() })
	//go serialListen()
}

func serialInit() {
	IsLog = true
	IsMPowerOn = true
	IsSerialing = false
	isCmdReturn = make(map[string]bool)
	isCmdReturn["G0"] = false
	isCmdReturn["G1"] = false
	isCmdReturn["G2"] = false
	isCmdReturn["G3"] = false
	isCmdReturn["L0"] = false
	isCmdReturn["L1"] = false
	isCmdReturn["L2"] = false
	isCmdReturn["L4"] = false //add 2nd lock
	isCmdReturn["L5"] = false //add 2nd lock
	isCmdReturn["L6"] = false //add 2nd lock
	isCmdReturn["V1"] = false
	isCmdReturn["V2"] = false
	isCmdReturn["V3"] = false
	isCmdReturn["V4"] = false
	isCmdReturn["V7"] = false
	isCmdReturn["V8"] = false
	isCmdReturn["P0"] = false
	isCmdReturn["P1"] = false
	isCmdReturn["M0"] = false
	isCmdReturn["M1"] = false
	isCmdReturn["Z0"] = false
	isCmdReturn["Z1"] = false
	isCmdReturn["Z2"] = false
	isCmdReturn["Z3"] = false
	isCmdReturn["S0"] = false
	isCmdReturn["S1"] = false
	isCmdReturn["S2"] = false
	isCmdReturn["S3"] = false
	IsCamraOn = [4]bool{false, false, false, false}
	IsDoorClosed = true
	IsLockLocked = true

	CamraOffTimes = [4]int{0, 0, 0, 0}
	CameraFileSeq = [4]int{0, 0, 0, 0}
	OfflineTimes = 0
	IsActivityRuning = true
	IsActivityQueing = false
	IsVideoStopped = false
	ICurStep = 0
	BLastDoorStatus = false
	BLastDoor2Status = false
	IHeartBeatCnt = 0
	INoTransTim = 0
	ITransLastTim = 0
	IOpenDoorFailCnt = 0
	ITakeVideoFailCnt = 0
	IActivityRunningTim = 0
	IsFirstTimeRun = true
	IsFirstTimeDualRun = true
	IsOperationDownByLock = false
	IsOperationDownByDoor = false
	IsOperationDownByLock2 = false
	IsOperationDownByDoor2 = false
	IsDualDoorEnabled = false
	ITotalTransCnt = 0
	I24HTransCnt = 0
	//SalesOf24H = [30]int{0}
	Iindexof24H = 0
	ICurHourTim = 0
	ICurHourTransCnt = 0
	IChkTransCnt = 0
	IRAMNoSpaceCnt = 0
	Temperature = 0

	sendSleepTime = 10 * time.Millisecond
	recvSleepTime = 20 * time.Millisecond
	recvSleepTimeOut = 100 * time.Millisecond

	StrTelProvider = "F"
	IMqttTransFailCnt = 0
	IMkdirErrCnt = 0

	var c, d, e *serial.Config
	if runtime.GOOS == "windows" {
		c = &serial.Config{Name: "COM3", Baud: 115200, ReadTimeout: time.Second * 1}
		d = &serial.Config{Name: "COM4", Baud: 115200, ReadTimeout: time.Second * 1}
		e = &serial.Config{Name: "COM5", Baud: 115200, ReadTimeout: time.Second * 1}
	} else {
		c = &serial.Config{Name: "/dev/ttyS1", Baud: 115200, ReadTimeout: time.Second * 1}
		d = &serial.Config{Name: "/dev/ttyUSB0", Baud: 115200, ReadTimeout: time.Second * 1}
		e = &serial.Config{Name: "/dev/ttyUSB2", Baud: 115200, ReadTimeout: time.Second * 1}
	}
	fmt.Println(c)
	fmt.Println(d)
	fmt.Println(e)

	ctxVideo = context.Background()
	videoCmds = make(map[int]*exec.Cmd)

	//打开串口
	var err error
	p, err = serial.OpenPort(c)
	if err != nil {
		alog.Log.Println(err)
	}
	p.Flush()

	dtel, err = serial.OpenPort(d)
	if err != nil {
		alog.Log.Println(err)
		alog.Log.Println("/dev/ttyUSB0 not in use")
	} else {
		alog.Log.Println("/dev/ttyUSB0 in use")
		dtel.Flush()
	}

	etel, err = serial.OpenPort(e)
	if err != nil {
		alog.Log.Println(err)
		alog.Log.Println("/dev/ttyUSB2 not in use")
	} else {
		alog.Log.Println("/dev/ttyUSB2 in use")
		etel.Flush()
	}
	isportusb0 = true
}

func Fulsh() {
	p.Flush()
}

func vserialRec() {
	for {
		time.Sleep(recvSleepTime)
		if p == nil {
			serialInit()
			continue
		}
		buf := make([]byte, 100)
		n, err := p.Read(buf)
		if n < 1 {
			continue
		}

		//fmt.Println(n, buf[0:n])
		if err != nil {
			alog.Log.Println(err)
		}

		for i := 0; i < n; i++ {
			//fmt.Println(buf[i])
			cb_recbuf.Put(buf[i])
		}
	}
}

// 监听串口消息
func serialListen() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("serialListen 捕获异常:", err)
		}
	}()
	//return
	iLen := 0
	buf := make([]byte, 100)
	for {
		var tempstr string
		lockType := public.Config["LOCK_TYPE"]
		//log.Println("serialListen", lockType)
		if lockType == "" {
			lockType = "ijooz"
			alog.Log.Println("locktype change2:", lockType)
		}
		recbyte, err := cb_recbuf.Get()
		if err != nil {
			if iLen == 0 {
				time.Sleep(recvSleepTime)
			}
			continue
		}
		if recbyte == 0x02 {
			iLen = 0
		}
		buf[iLen] = recbyte
		iLen++
		//fmt.Println(recbyte, " ", iLen)
		if iLen == 10 {
			iLen = 0
			if buf[0] != 0x02 || buf[9] != 0x03 {
				alog.Log.Printf("start or end error, %d, %d", buf[0], buf[9])
				continue
			}
		} else {
			if iLen > 80 {
				iLen = 0
			}
			continue
		}

		// 暂时不处理长度不等于10的数据包
		if IsLog {
			alog.Log.Println("RECV ", hex.EncodeToString(buf[0:10]))
		}
		copy(lastData, buf)

		if buf[1] == 0x47 {
			// G 全局
			// 获取IO状态
			if buf[2] == 0x30 {
				cmdreturn_mutex.Lock()
				isCmdReturn[string([]byte{buf[1], buf[2]})] = true
				cmdreturn_mutex.Unlock()

				if IsLog {
					tempstr += ("lockType:" + lockType + " ")
					//alog.Log.Println("lockType:", lockType)
					tempstr += ("curStep:" + strconv.Itoa(ICurStep) + " ")
					//alog.Log.Println("curStep:", ICurStep)
				}

				//io, err := strconv.Atoi(string([]byte{buf[3], buf[4]}))
				io, err := strconv.Atoi(string([]byte{buf[4]}))
				if err != nil {
					alog.Log.Println("read IO data error")
					continue
				}
				if IsLog {
					tempstr += ("io:" + strconv.Itoa(io) + " ")
					//alog.Log.Println("io: ", io)
				}
				isDoorClosedTmp := true
				if io|0b00000001 == io {
					isDoorClosedTmp = false
				}
				isLockLockedTmp := true
				if io|0b00000010 == io {
					isLockLockedTmp = false
				}
				is4GPowOnTmp := true
				if io|0b00000100 == io {
					is4GPowOnTmp = false
				}
				isBuzzerOnTmp := false
				if io|0b00001000 == io {
					isBuzzerOnTmp = true
				}
				IsDoorClosed = isDoorClosedTmp
				IsLockLocked = isLockLockedTmp
				if lockType != "haha" {
					IsLockLocked = !IsLockLocked
				}

				var IsCamraOnTmp [4]bool // 摄像头1状态
				IsCamraOnTmp[0] = true
				if io|0b00010000 == io {
					IsCamraOnTmp[0] = false
				}
				IsCamraOnTmp[1] = true
				if io|0b00100000 == io {
					IsCamraOnTmp[1] = false
				}
				IsCamraOnTmp[2] = true
				if io|0b01000000 == io {
					IsCamraOnTmp[2] = false
				}
				IsCamraOnTmp[3] = true
				if io|0b10000000 == io {
					IsCamraOnTmp[3] = false
				}

				var io2 int
				if buf[6] != 0xAA && public.IsPubDualDoorMod {
					io2, err = strconv.Atoi(string([]byte{buf[6]}))
					if err != nil {
						alog.Log.Println("read IO2 data error")
						continue
					}
					if IsLog {
						tempstr += ("io2:" + strconv.Itoa(io2) + " ")
						//alog.Log.Println("io2:", io2)
					}
				} else {
					if lockType != "haha" {
						io2 = 2
					} else {
						io2 = 0
					}
				}

				isDoor2ClosedTmp := true
				if io2|0b00000001 == io2 {
					isDoor2ClosedTmp = false
				}
				isLock2LockedTmp := true
				if io2|0b00000010 == io2 {
					isLock2LockedTmp = false
				}
				IsDoor2Closed = isDoor2ClosedTmp
				IsLock2Locked = isLock2LockedTmp
				if lockType != "haha" {
					IsLock2Locked = !IsLock2Locked
				}

				Is4GPowOn = is4GPowOnTmp
				IsBuzzerOn = isBuzzerOnTmp
				strDoorStatus := "open"
				if IsDoorClosed {
					strDoorStatus = "close"
				}
				strLockStatus := "unlock"
				if IsLockLocked {
					strLockStatus = "locked"
				}

				strDoor2Status := ""
				strLock2Status := ""
				if public.IsPubDualDoorMod {
					strDoor2Status = "open2"
					if IsDoor2Closed {
						strDoor2Status = "close2"
					}
					strLock2Status = "unlock2"
					if IsLock2Locked {
						strLock2Status = "locked2"
					}
				}

				if IsLog {
					alog.Log.Println(tempstr)
					alog.Log.Println("IO_STATUS:", strDoorStatus, strLockStatus, Is4GPowOn, IsBuzzerOn, IsCamraOnTmp[0], IsCamraOnTmp[1], IsCamraOnTmp[2], IsCamraOnTmp[3], strDoor2Status, strLock2Status)
				}
			} else if buf[2] == 0x31 { // 获取RTC
				cmdreturn_mutex.Lock()
				isCmdReturn[string([]byte{buf[1], buf[2]})] = true
				cmdreturn_mutex.Unlock()
			} else if buf[2] == 0x32 { // 获取温度
				cmdreturn_mutex.Lock()
				isCmdReturn[string([]byte{buf[1], buf[2]})] = true
				cmdreturn_mutex.Unlock()
				temp, err := strconv.Atoi(string([]byte{buf[3], buf[4], buf[6], buf[5]}))
				if err != nil {
					alog.Log.Println("read temp data error")
					continue
				}
				Temperature = float64(temp) / 100
				alog.Log.Println("Temperature: ", Temperature)
			} else if buf[2] == 0x33 { // 获取版本号
				cmdreturn_mutex.Lock()
				isCmdReturn[string([]byte{buf[1], buf[2]})] = true
				cmdreturn_mutex.Unlock()
				Version = string([]byte{buf[3], buf[4], buf[5], buf[6]})
				alog.Log.Println("Version: ", Version)
			}
		} else if buf[1] == 0x4C {
			// L 锁开关
			cmdreturn_mutex.Lock()
			if buf[2] == 0x30 {
				// 关锁，完成关锁
				isCmdReturn[string([]byte{buf[1], buf[2]})] = true
			} else if buf[2] == 0x31 {
				// 开锁
				isCmdReturn[string([]byte{buf[1], buf[2]})] = true
			} else if buf[2] == 0x32 {
				// 带条件开锁
				isCmdReturn[string([]byte{buf[1], buf[2]})] = true
			} else if buf[2] == 0x34 {
				// 关锁，完成关锁
				isCmdReturn[string([]byte{buf[1], buf[2]})] = true
			} else if buf[2] == 0x35 {
				// 开锁
				isCmdReturn[string([]byte{buf[1], buf[2]})] = true
			} else if buf[2] == 0x36 {
				// 带条件开锁
				isCmdReturn[string([]byte{buf[1], buf[2]})] = true
			}
			cmdreturn_mutex.Unlock()
		} else if buf[1] == 0x56 {
			// V 控制摄像头电源 或者 LED亮度
			cmdreturn_mutex.Lock()
			isCmdReturn[string([]byte{buf[1], buf[2]})] = true
			cmdreturn_mutex.Unlock()
		} else if buf[1] == 0x50 {
			// P 4G模块电源
			cmdreturn_mutex.Lock()
			isCmdReturn[string([]byte{buf[1], buf[2]})] = true
			cmdreturn_mutex.Unlock()
		} else if buf[1] == 0x4D {
			// P 主控电源
			cmdreturn_mutex.Lock()
			isCmdReturn[string([]byte{buf[1], buf[2]})] = true
			cmdreturn_mutex.Unlock()
		} else if buf[1] == 0x5A {
			// P 蜂鸣器
			cmdreturn_mutex.Lock()
			isCmdReturn[string([]byte{buf[1], buf[2]})] = true
			cmdreturn_mutex.Unlock()
		} else if buf[1] == 0x53 {
			// S 设置
			cmdreturn_mutex.Lock()
			isCmdReturn[string([]byte{buf[1], buf[2]})] = true
			cmdreturn_mutex.Unlock()
		}
	}
}

func serialSendData(buf []byte) error {
	pingpong_mutex.Lock()
	defer pingpong_mutex.Unlock()
	if p == nil {
		IsSerialing = false
		serialInit()
		return errors.New("serial Init Error")
	}
	iTry := 0
	for {
		if IsSerialing {
			iTry++
			if iTry > 60 {
				return errors.New("serial in use")
			} else {
				time.Sleep(sendSleepTime)
				continue
			}
		}
		break
	}
	IsSerialing = true
	defer func() {
		IsSerialing = false
	}()
	data := makePack(buf)
	k := string([]byte{data[1], data[2]})

	cmdreturn_mutex.Lock()
	isCmdReturn[k] = false //before recv, should be false
	cmdreturn_mutex.Unlock()

	//send 3 times to make sure the data be accepted by 51
	for j := 0; j < 3; j++ {
		_, err := p.Write(data)
		if IsLog {
			alog.Log.Println("WRITE", hex.EncodeToString(data))
		}
		if err != nil {
			alog.Log.Println(err)
			//IsSerialing = false
			return err
		}
		for i := 0; i < 3; i++ {
			time.Sleep(recvSleepTimeOut)
			cmdreturn_mutex.Lock()
			if isCmdReturn[k] {
				cmdreturn_mutex.Unlock()
				// 返回了数据。。。
				//IsSerialing = false
				return nil
			}
			cmdreturn_mutex.Unlock()
		}
	}

	//IsSerialing = false
	return errors.New("unknown error")
}

func GetUnfinishActivity() (error, interface{}) {
	// 获取未完成的活动信息
	var activityInfoList []public.ActivityInfo
	err := public.DB.All(&activityInfoList, storm.Limit(1))
	if err != nil {
		err := errors.New("HAVE NO UNFINISH ACTIVITY")
		return err, nil
	}
	if len(activityInfoList) > 0 {
		return nil, activityInfoList[0]
	} else {
		err := errors.New("HAVE NO UNFINISH ACTIVITY")
		return err, nil
	}
}

func SetLight4() error {
	return serialSendData([]byte{0x56, 0x37, 0x30, 0x30})
}

func SetLight3() error {
	return serialSendData([]byte{0x56, 0x37, 0x30, 0x30})
}

func SetLight2() error {
	return serialSendData([]byte{0x56, 0x37, 0x30, 0x30})
}

func SetLight1() error {
	return serialSendData([]byte{0x56, 0x37, 0x30, 0x30})
}

func SetLight0() error {
	return serialSendData([]byte{0x56, 0x37, 0x30, 0x30})
}

func OpenLock() error {
	CloseLock()
	// 开锁
	time.Sleep(1500 * time.Millisecond)
	alog.Log.Println("OpenLock:", "sendcmd", IsDualDoorEnabled)
	var err error
	if IsDualDoorEnabled {
		err = serialSendData([]byte{0x4C, 0x36, 0x30, 0x31, 0x30, 0x30})
	} else {
		err = serialSendData([]byte{0x4C, 0x32, 0x30, 0x31, 0x30, 0x30})
	}
	if err != nil {
		CloseLock()
		return err
	}
	i := 0
	j := 0
	for {
		time.Sleep(sendSleepTime)
		i++

		if i%5 == 0 {
			//check status
			go Pingpong(false)
		}

		if (!IsDoorClosed && !IsDualDoorEnabled) || (!IsDoor2Closed && IsDualDoorEnabled) {
			alog.Log.Println("OpenLock:", "door is open", IsDualDoorEnabled)
			return nil
		}

		if (!IsLockLocked && !IsDualDoorEnabled) || (!IsLock2Locked && IsDualDoorEnabled) {
			j++
			if j > 200 {
				alog.Log.Println("OpenLock:", "unlocked", IsDualDoorEnabled)
				return nil
			}
		} else {
			j = 0
		}
		// if i == 200 {
		// 	log.Println("OpenLock:", "2seconds")
		// }
		if i > 802 {
			CloseLock()
			alog.Log.Println("OpenLock:", "sensorerr", IsDualDoorEnabled)
			return errors.New("openlock timeout")
		}
	}
}

func OpenAllLock(itype byte) error {
	CloseAllLock(itype)
	// 开锁
	//time.Sleep(1 * time.Second)
	alog.Log.Println("OpenAllLock:", "sendcmd", itype)
	var err error
	if itype&0x02 == 0x02 {
		err = serialSendData([]byte{0x4C, 0x36, 0x30, 0x31, 0x30, 0x30})
	}
	if itype&0x01 == 0x01 {
		err = serialSendData([]byte{0x4C, 0x32, 0x30, 0x31, 0x30, 0x30})
	}
	if err != nil {
		CloseAllLock(itype)
		return err
	}
	i := 0
	j := 0
	for {
		time.Sleep(sendSleepTime)
		i++

		if i%5 == 0 {
			//check status
			go Pingpong(false)
		}

		if (!IsDoorClosed && itype&0x01 == 0x01) || (!IsDoor2Closed && itype&0x02 == 0x02) {
			alog.Log.Println("OpenAllLock:", "door is open", itype)
			return nil
		}

		if (!IsLockLocked && itype&0x01 == 0x01) || (!IsLock2Locked && itype&0x02 == 0x02) {
			j++
			if j > 200 {
				alog.Log.Println("OpenAllLock:", "unlocked", itype)
				return nil
			}
		} else {
			j = 0
		}
		// if i == 200 {
		// 	log.Println("OpenLock:", "2seconds")
		// }
		if i > 1502 {
			CloseAllLock(itype)
			alog.Log.Println("OpenAllLock:", "sensor err", itype)
			return errors.New("openalllock timeout")
		}
	}
}

func CloseLockEn() error {
	// 关锁
	//alog.Log.Println("CloseLock:", "sendcmd", IsDualDoorEnabled)
	err := CloseLock()
	if err != nil {
		alog.Log.Println("CloseLockEn:", err, IsDualDoorEnabled)
		return err
	}
	i := 0
	for {
		time.Sleep(sendSleepTime)
		i++

		if i%5 == 0 {
			//check status
			go Pingpong(false)
		}

		if (IsLockLocked && !IsDualDoorEnabled) || (IsLock2Locked && IsDualDoorEnabled) {
			alog.Log.Println("CloseLockEn:done", IsDualDoorEnabled)
			return nil
		}
		if i > 802 {
			CloseLock()
			alog.Log.Println("CloseLockEn:sensor err", IsDualDoorEnabled)
			return errors.New("closelock timeout")
		}
	}
}

func OpenLockOnly() error {
	// 开锁
	var err error
	if IsDualDoorEnabled {
		err = serialSendData([]byte{0x4C, 0x35})
	} else {
		err = serialSendData([]byte{0x4C, 0x31})
	}
	if err != nil {
		return err
	}
	i := 0
	for {
		time.Sleep(sendSleepTime)
		i++
		if (!IsLockLocked && !IsDualDoorEnabled) || (!IsLock2Locked && IsDualDoorEnabled) {
			return nil
		}
		if i > 100 {
			return errors.New("open lock timeout")
		}
	}
}

func Pingpong(islog bool) error {
	// Pingpong
	IsLog = islog
	ipingpongcnt++
	//alog.Log.Println("ipingpongcnt:", ipingpongcnt)
	if ipingpongcnt < 20 || islog == false {
		return serialSendData([]byte{0x47, 0x30})
	} else {
		ipingpongcnt = 0
		return serialSendData([]byte{0x47, 0x32})
	}
}

func CloseLock() error {
	// 关锁
	IsLog = true
	if IsDualDoorEnabled {
		return serialSendData([]byte{0x4C, 0x34})
	} else {
		return serialSendData([]byte{0x4C, 0x30})
	}
}

func CloseAllLock(itype byte) {
	// 关锁
	if itype&0x01 == 0x01 {
		serialSendData([]byte{0x4C, 0x30})
	}
	if itype&0x02 == 0x02 {
		serialSendData([]byte{0x4C, 0x34})
	}
}

func OpenLockPower() error {
	// 锁上电
	return nil
}

func CloseLockPower() error {
	// 锁停电
	return nil
}

func CloseCPU() error {
	// 主控停电
	return serialSendData([]byte{0x4D, 0x30, 0x30, 0x30})
}

func ResetCPU() error {
	// 重启主控模块
	return serialSendData([]byte{0x4D, 0x30, 0x31, 0x34})
}

func Reset4G() error {
	// 重启4G模块
	return serialSendData([]byte{0x50, 0x30, 0x31, 0x45})
}

func ResetAll() error {
	// 重启所有模块
	return serialSendData([]byte{0x52, 0x30, 0x55, 0x55})
}

func ResetCamra(i int) error {
	if i == 1 {
		// 重启摄像头1
		return serialSendData([]byte{0x56, 0x31, 0x33, 0x32})
	}
	if i == 2 {
		// 重启摄像头2
		return serialSendData([]byte{0x56, 0x32, 0x33, 0x32})
	}
	if i == 3 {
		// 重启摄像头3
		return serialSendData([]byte{0x56, 0x33, 0x33, 0x32})
	}
	if i == 4 {
		// 重启摄像头4
		return serialSendData([]byte{0x56, 0x34, 0x33, 0x32})
	}
	return errors.New("i is error")
}

func WaitCloseDoor(activityid string) error {
	// 等待用户开门及关门
	startTime := time.Now().Unix()
	thisTime := time.Now().Unix()
	tempTime := time.Now().Unix()
	for {
		thisTime = time.Now().Unix()
		tempTime = thisTime
		if (!IsDoorClosed && !IsDualDoorEnabled) || (!IsDoor2Closed && IsDualDoorEnabled) {
			//门开
			IOpenDoorFailCnt = 0
			if (!BLastDoorStatus && !IsDualDoorEnabled) || (!BLastDoor2Status && IsDualDoorEnabled) {
				if IsDualDoorEnabled {
					BLastDoor2Status = true
				} else {
					BLastDoorStatus = true
				}
				public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_DOOROPENED, "", activityid)
				alog.Log.Println("ACTIVITY OPENDOOR2", IsDualDoorEnabled)
			} else {
				alog.Log.Println("ACTIVITY OPENDOOR1", IsDualDoorEnabled)
			}
			break
		} else {
			//门关
			if thisTime-startTime > 10 {
				IOpenDoorFailCnt++
				if IOpenDoorFailCnt > 3 {
					IOpenDoorFailCnt = 0
					public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_MACHINE_DOWN8, activityid)
				}
				//超过10秒，门没开，直接关上锁--force to close door
				alog.Log.Println("WaitCloseDoor:door never open", IsDualDoorEnabled, IOpenDoorFailCnt)
				err := CloseLockEn()
				return err
			}
		}
		// time.Sleep(0.5e9)
	}
	alog.Log.Println("WaitCloseDoor:WAITING door close", IsDualDoorEnabled)
WAITING:
	isAllOk := true
	isDoorOk := true
	for i := 0; i < 3; i++ {
		if i == 0 {
			time.Sleep(0.3e9)
		} else if i == 1 {
			time.Sleep(0.1e9)
		} else {
			time.Sleep(0.01e9)
		}
		//最后检查一次门与锁状态。
		if IsDualDoorEnabled {
			isAllOk = isAllOk && IsDoor2Closed && IsLock2Locked
			isDoorOk = isDoorOk && IsDoor2Closed
		} else {
			isAllOk = isAllOk && IsDoorClosed && IsLockLocked
			isDoorOk = isDoorOk && IsDoorClosed
		}
	}

	if isDoorOk {
		if (BLastDoorStatus && !IsDualDoorEnabled) || (BLastDoor2Status && IsDualDoorEnabled) {
			if IsDualDoorEnabled {
				BLastDoor2Status = false
			} else {
				BLastDoorStatus = false
			}
			public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_DOORCLODED, "", activityid)
			alog.Log.Println("ACTIVITY CLOSE-DOOR", IsDualDoorEnabled)
		}
	} else {
		if (!BLastDoorStatus && !IsDualDoorEnabled) || (!BLastDoor2Status && IsDualDoorEnabled) {
			if IsDualDoorEnabled {
				BLastDoor2Status = true
			} else {
				BLastDoorStatus = true
			}
			IOpenDoorFailCnt = 0
			public.SendMqttStatus(public.TYPE_ACTIVITY, public.ACTION_DOOROPENED, "", activityid)
			alog.Log.Println("ACTIVITY OPENDOOR3", IsDualDoorEnabled)
		}
	}

	if isAllOk {
		return nil
	} else {
		thisTime = time.Now().Unix()
		if thisTime-startTime > 300 {
			// video超过5分钟已关门的话。
			if isDoorOk {
				alog.Log.Println("DoorOk but Lock is not", IsDualDoorEnabled)
				err := CloseLockEn()
				return err
			}
		}
		if thisTime-startTime > 3600 {
			// video超过一个小时的话。
			alog.Log.Println("Door & Lock is not")
			err := CloseLockEn()
			return err
		}
		if thisTime-tempTime > 10 {
			alog.Log.Println("wait 10s,door:", isDoorOk, IsDualDoorEnabled)
			tempTime = thisTime
			if isDoorOk {
				alog.Log.Println("DoorOk but Lock unlocked", IsDualDoorEnabled)
				CloseLockEn()
			}
		}
		if ITransLastTim%30 == 0 {
			ITransLastTim++
			//report an translasttoolong error every 30 mins
			public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_MACHINE_DOWN2, activityid)
		}
		goto WAITING
	}
}

func makePack(bufs []byte) []byte {
	lenc := len(bufs)
	if lenc > 6 {
		return make([]byte, 10)
	}
	data := make([]byte, 10)
	data[0] = 0x02
	var re byte
	for i := 0; i < lenc; i++ {
		data[i+1] = bufs[i]
		if i == 0 {
			re = bufs[i]
		} else {
			re = bufs[i] ^ re
		}
	}
	hexStr := fmt.Sprintf("%X", re)
	// log.Println("hexStr", hexStr)

	crc1 := byte(hexStr[0])
	crc2 := byte(0x00)
	if len(hexStr) > 1 {
		crc2 = byte(hexStr[1])
	}

	if lenc < 6 {
		for i := 0; i < 6-lenc; i++ {
			data[lenc+i+1] = 0x55
		}
	}
	data[7] = crc1
	data[8] = crc2
	data[9] = 0x03
	return data
}

func snapShotOne(ch chan string, activityId string, i int, bfType string) {
	snapShotHost := public.GetCameraHost()
	cacheFileName := public.GetImgPath(activityId, i, bfType)
	if public.IsFile(cacheFileName) {
		// 照片已经拍过，略过
		ch <- cacheFileName
		return
	} else {
		// url := "http://upload.fuhuibaotech.com/"
		// url := "http://" + fmt.Sprintf(snapShotHost, i) + ":8080/?action=snapshot"
		url := "http://" + fmt.Sprintf(snapShotHost, i) + "/web/tmpfs/snap.jpg"
		err, status, body := public.HttpGet(url)
		if err != nil {
			alog.Log.Println("SNAP FAIL:", err, status)
			ch <- "error:http get"
			return
		} else {
			alog.Log.Println("SNAP SUCCESS:", status)
			cacheFileName2 := public.GetSdcardPath() + "/" + cacheFileName
			cacheFileName2Path, _ := filepath.Split(cacheFileName2)
			if !public.IsExist(cacheFileName2Path) {
				err := os.MkdirAll(cacheFileName2Path, 0777)
				if err != nil {
					ch <- "error:make dir"
					return
				}
			}
			err = os.WriteFile(cacheFileName2, body, 0766)
			if err != nil {
				ch <- "error:write file"
				return
			} else {
				//拍完后，检查文件大小，如果=0，返回失败
				fileinfo, err := os.Stat(cacheFileName2)
				if err != nil {
					ch <- "error:check file size read error"
					return
				}
				if fileinfo.Size() == 0 {
					ch <- "error:file size is zero"
					return
				}
				//拍完后，缩放
				//err = public.ResizeJpg(cacheFileName2, 720)
				//if err != nil {
				//	ch <- "error:resize img"
				//	return
				//}
			}
			ch <- cacheFileName
			return
		}
	}
}

func SnapShot(activityId string, bfType string) (error, map[int]string) {
	if len(activityId) < 8 {
		err := errors.New("activity_id too short")
		return err, nil
	}
	// 拍照
	var fileNames map[int]string
	fileNames = make(map[int]string)

	cameraCount, _ := strconv.Atoi(public.Config["CAMERA_COUNT"])
	chs := make([]chan string, cameraCount)

	for i := 1; i < cameraCount+1; i++ {
		chs[i-1] = make(chan string)
		go snapShotOne(chs[i-1], activityId, i, bfType)
	}

	var input string
	for k, ch := range chs {
		input = <-ch
		if input[0:5] != "error" {
			fileNames[k+1] = input
		} else {
			err := errors.New(input[6:])
			return err, nil
		}
	}

	alog.Log.Println("ACTIVITY SNAPSHOT", bfType)
	return nil, fileNames
}

func SnapShotUpload(activityId string, bfType string) (error, map[int]string) {
	tmp := make(map[int]string)
	return nil, tmp
	err, fileNames := SnapShot(activityId, bfType)
	if err == nil {
		//拍照成功，加入到上传队列
		uploadInfo := public.UploadInfo{
			Id:             fmt.Sprintf("%s_%s", activityId, bfType),
			ActivityId:     activityId,
			ImgAllCount:    len(fileNames),
			ImgUploadCount: 0,
			ImgType:        bfType,
			IsNotified:     false,
		}
		public.DB.Save(&uploadInfo)

		for k, item := range fileNames {
			uploadFile := public.UploadFile{
				Id:          fmt.Sprintf("%s_%d_%s", activityId, k, bfType),
				InfoId:      fmt.Sprintf("%s_%s", activityId, bfType),
				ActivityId:  activityId,
				ImgPath:     item,
				ImgUploaded: false,
			}
			public.DB.Save(&uploadFile)
		}
		public.ChUpload <- "N"
	}
	return err, fileNames
}

func StartVideo() error {
	stopvideo_mutex.Lock()
	defer stopvideo_mutex.Unlock()
	noError := false
	var errorMsg []string
	IsVideoStopped = false
	cameraCount, _ := strconv.Atoi(public.Config["CAMERA_COUNT"])
	if cameraCount > 4 {
		return errors.New("StartVideo:camera count can not gt 4")
	}
	cameraType := public.Config["CAMERA_TYPE"]
	camStartArray := [4]int{0, 1, 2, 3}
	camStartIpadd := [4]int{1, 2, 3, 4}
	if public.IsPubDualDoorMod {
		if cameraCount > 2 {
			cameraCount = 2
		}
		if IsDualDoorEnabled {
			if cameraType == "haha" {
				camStartArray = [4]int{0, 3, 0, 0}
				camStartIpadd = [4]int{0, 3, 0, 0}
			} else {
				camStartArray = [4]int{2, 3, 0, 0}
				camStartIpadd = [4]int{3, 4, 0, 0}
			}
		} else {
			if cameraType == "haha" {
				camStartArray = [4]int{1, 2, 0, 0}
				camStartIpadd = [4]int{1, 2, 0, 0}
			} else {
				camStartArray = [4]int{0, 1, 0, 0}
				camStartIpadd = [4]int{1, 2, 0, 0}
			}
		}
	}
	CameraFileSeq = [4]int{0, 0, 0, 0}
	j := 0
	for i := 0; i < cameraCount; i++ {
		if !IsCamraOn[camStartArray[i]] {
			tempmsg := fmt.Sprintf("Camra %d Is Not On!", camStartArray[i]+1)
			errorMsg = append(errorMsg, tempmsg)
			continue
		}
		CameraFileSeq[j] = camStartArray[i]
		snapShotHost := public.GetCameraHost()
		var snapShotHost2 string
		snapShotHost2 = fmt.Sprintf(snapShotHost, camStartIpadd[i])
		cacheFileName := public.GetVideoPath(CurActivityId, j)
		cacheFileNameFull := public.GetSdcardPath() + "/" + cacheFileName
		cacheFileNameFull = strings.ReplaceAll(cacheFileNameFull, "\\", "/")
		cacheFileNameFullPath, _ := filepath.Split(cacheFileNameFull)
		if public.IsExist(cacheFileNameFull) {
			alog.Log.Println("StartVideo:cFile exist:", cacheFileNameFull)
			continue
		} else {
			alog.Log.Println("StartVideo:cFile:", cacheFileNameFull)
		}
		if !public.IsExist(cacheFileNameFullPath) {
			err := os.MkdirAll(cacheFileNameFullPath, 0777)
			if err != nil {
				j++
				IMkdirErrCnt++
				errorMsg = append(errorMsg, "mkdir error!")
				continue
			}
			alog.Log.Println("new cFilePath: ", cacheFileNameFullPath)
		} else {
			alog.Log.Println("old cFilePath: ", cacheFileNameFullPath)
		}
		IMkdirErrCnt = 0
		c := ""
		//cameraType := public.Config["CAMERA_TYPE"]
		if cameraType == "dev" {
			// dev摄像头
			c = fmt.Sprintf("ffmpeg -nostdin -y -use_wallclock_as_timestamps 1 -rtsp_transport tcp -i rtsp://%s:554/user=admin_password=123456_channel=1_stream=0 -bufsize 1MB -timeout 3000000 -an -vcodec copy %s &", snapShotHost2, cacheFileNameFull)
		} else if cameraType == "ijooz" {
			// ijooz摄像头
			c = fmt.Sprintf("ffmpeg -nostdin -y -use_wallclock_as_timestamps 1 -rtsp_transport tcp -i rtsp://%s:8554/live0 -bufsize 1MB -timeout 3000000 -an -vcodec copy %s &", snapShotHost2, cacheFileNameFull)
		} else if cameraType == "haha" {
			// 哈哈摄像头
			c = fmt.Sprintf("ffmpeg -nostdin -y -use_wallclock_as_timestamps 1 -rtsp_transport tcp -i rtsp://%s:554/0 -bufsize 1MB -timeout 3000000 -an -vcodec copy %s &", snapShotHost2, cacheFileNameFull)
		} else {
			// 默认摄像头
			c = fmt.Sprintf("ffmpeg -nostdin -y -use_wallclock_as_timestamps 1 -rtsp_transport tcp -i rtsp://%s:8554/live0 -bufsize 1MB -timeout 3000000 -vcodec copy %s &", snapShotHost2, cacheFileNameFull)
		}
		//note, use -t 3600 to make fixed duration video.
		alog.Log.Println("GET VIDEO CMD: ", c)

		if runtime.GOOS == "windows" {
			videoCmds[i] = exec.CommandContext(ctxVideo, "sh", "-c", c)
		} else {
			videoCmds[i] = exec.CommandContext(ctxVideo, "sh", "-c", c)
			videoCmds[i].SysProcAttr = &syscall.SysProcAttr{
				Setpgid: true,
			}
		}

		videoCmds[i].Stdout = os.Stdout
		videoCmds[i].Stderr = os.Stderr

		if err := videoCmds[i].Start(); err != nil {
			errorMsg = append(errorMsg, fmt.Sprintf("start video error: %d", i+1))
			continue
		}

		// time.Sleep(4 * time.Second)
		// if !public.IsExist(cacheFileNameFull) {
		// 	return errors.New("start video error!")
		// }

		// if runtime.GOOS != "windows" {
		// 	errCmdCh := make(chan error, 1)
		// 	go func() {
		// 		errr := videoCmds[i].Wait()
		// 		fmt.Println("errr: ", errr)
		// 		errCmdCh <- errr
		// 	}()
		// }
		j++
		noError = noError || true
	}
	if !noError {
		//StopVideo() //will lock the program here!!!
		return errors.New(strings.Join(errorMsg, ","))
	}
	return nil
}

func StopVideo() error {
	stopvideo_mutex.Lock()
	defer stopvideo_mutex.Unlock()
	alog.Log.Println("StopVideo")
	IsVideoStopped = true
	cameraCount, _ := strconv.Atoi(public.Config["CAMERA_COUNT"])
	if public.IsPubDualDoorMod {
		if cameraCount > 2 {
			cameraCount = 4
		} else {
			cameraCount *= 2
		}
	}
	for i := 0; i < cameraCount; i++ {
		if runtime.GOOS == "windows" {
			if videoCmds[i] != nil {
				time.Sleep(100 * time.Millisecond)
				c := exec.Command("taskkill.exe", "/f", "/im", "ffmpeg.exe")
				c.Start()
			}
		} else {
			if videoCmds[i] != nil {
				if false {
					time.Sleep(100 * time.Millisecond)
					if err := syscall.Kill(-1*videoCmds[i].Process.Pid, 2); err != nil {
						alog.Log.Println("failed to kill process: ", err)
						// return err
					}
					time.Sleep(100 * time.Millisecond)
					if err := syscall.Kill(-1*videoCmds[i].Process.Pid, 2); err != nil {
						alog.Log.Println("failed to kill process: ", err)
						// return err
					}
					videoCmds[i] = nil
				} else {
					proc, err := os.FindProcess(videoCmds[i].Process.Pid)
					if err != nil {
						//return err
						continue
					}

					// Send a SIGTERM signal to the process
					err = proc.Signal(syscall.SIGTERM)
					if err != nil {
						//return err
						continue
					}

					// Wait for the process to terminate gracefully
					time.Sleep(100 * time.Millisecond)

					// Wait for the process to exit, but with a timeout of 5 seconds
					done := make(chan error, 1)
					go func() {
						_, err := proc.Wait()
						done <- err
					}()

					select {
					case <-time.After(5 * time.Second):
						// Timed out, the process is still running
						alog.Log.Println("FFmpeg Process is still running after 5 seconds, sending SIGKILL...", i)
						err = proc.Kill()
						if err != nil {
							// handle error
							alog.Log.Println("FFmpeg Process kill failed:", i)
						} else {
							videoCmds[i] = nil
						}
					case err := <-done:
						// The process has exited
						if err != nil {
							// handle error
						} else {
							videoCmds[i] = nil
						}
						alog.Log.Println("FFmpeg Process has exited:", i)
					}
				}
			} else {
				alog.Log.Println("FFmpeg Process is nil:", i)
			}
		}
	}
	public.ExecShell("killall ffmpeg")
	return nil
}

func AddUpladVideo(needUploadVideo string) {
	DBMutex.Lock()
	defer DBMutex.Unlock()

	cameraCount, _ := strconv.Atoi(public.Config["CAMERA_COUNT"])
	if cameraCount > 4 {
		cameraCount = 4
	}
	//拍照成功，加入到上传队列
	sType := "A"
	if IsDualDoorEnabled {
		sType = "B"
		if cameraCount > 2 {
			cameraCount = 2
		}
	}
	uploadInfo := public.UploadInfo{
		Id:             fmt.Sprintf("%s_%s", CurActivityId, sType),
		ActivityId:     CurActivityId,
		ImgAllCount:    cameraCount,
		ImgUploadCount: 0,
		ImgType:        sType,
		IsNotified:     false,
	}
	public.DB.Save(&uploadInfo)
	//alog.Log.Println("AddUpladVideo: UploadInfo ", CurActivityId)
	for i := 0; i < cameraCount; i++ {
		cacheFileName := public.GetVideoPath(CurActivityId, i)
		cacheFileNameFull := public.GetSdcardPath() + "/" + cacheFileName
		cacheFileNameFull = strings.ReplaceAll(cacheFileNameFull, "\\", "/")
		if !public.IsExist(cacheFileNameFull) {
			if needUploadVideo == "1" || needUploadVideo == "5" {
				public.SendMqttError(public.TYPE_ACTIVITY, public.ERROR_UPLOAD5, CurActivityId)
			}
			alog.Log.Println("AddUpladVideo: file not exist: ", needUploadVideo, cacheFileNameFull)
			continue
		}
		if CameraFileSeq[i] == 0 {
			CameraFileSeq[i] = i
		}
		uploadFile := public.UploadFile{
			Id:          fmt.Sprintf("%s_%d_%s", CurActivityId, CameraFileSeq[i], sType),
			InfoId:      fmt.Sprintf("%s_%s", CurActivityId, sType),
			ActivityId:  CurActivityId,
			ImgPath:     cacheFileName,
			ImgUploaded: false,
		}
		public.DB.Save(&uploadFile)
		//alog.Log.Println("AddUpladVideo: uploadFile ", cacheFileNameFull)
	}
	public.ChUpload <- "N"
}

func DelUnusedVideo() {
	cameraCount, _ := strconv.Atoi(public.Config["CAMERA_COUNT"])
	if cameraCount > 4 {
		cameraCount = 4
	}
	//开门不成功，删除视频
	if IsDualDoorEnabled {
		if cameraCount > 2 {
			cameraCount = 2
		}
	}
	for i := 0; i < cameraCount; i++ {
		cacheFileName := public.GetVideoPath(CurActivityId, i)
		cacheFileNameFull := public.GetSdcardPath() + "/" + cacheFileName
		cacheFileNameFull = strings.ReplaceAll(cacheFileNameFull, "\\", "/")
		if !public.IsExist(cacheFileNameFull) {
			alog.Log.Println("DelUnusedVideo: file not exist: ", cacheFileNameFull)
			continue
		}
		public.DeleteFileOnDisk(cacheFileNameFull)
	}
}

func ClearActivity() error {
	// 慎用，此功能是在设备被很多卡卡住的时候，清除activity用的。
	var activityInfoList []public.ActivityInfo
	err := public.DB.All(&activityInfoList, storm.Limit(1))
	if err != nil {
		return err
	} else {
		if len(activityInfoList) > 0 {
			activityInfo := activityInfoList[0]
			if err := public.DB.DeleteStruct(&activityInfo); err != nil {
				alog.Log.Println("ACTIVITY DELETE ERROR:", err)
				return err
			} else {
				return nil
			}
		} else {
			return errors.New("have no activity")
		}
	}
}

// GetIsActivityRunning returns the value of IsActivityRunning.
func GetIsActivityRunning() bool {
	queMutex.Lock()
	defer queMutex.Unlock()
	return IsActivityRuning
}

// SetIsActivityRunning sets the value of IsActivityRunning.
func SetIsActivityRunning(value bool) {
	queMutex.Lock()
	defer queMutex.Unlock()
	IsActivityRuning = value
}

func GetIsActivityQueing() bool {
	queMutex.Lock()
	defer queMutex.Unlock()
	return IsActivityQueing
}

// SetIsActivityRunning sets the value of IsActivityRunning.
func SetIsActivityQueing(value bool) {
	queMutex.Lock()
	defer queMutex.Unlock()
	IsActivityQueing = value
}

func GetIMqttTransFailCnt() int {
	queMutex.Lock()
	defer queMutex.Unlock()
	return IMqttTransFailCnt
}

// SetIsActivityRunning sets the value of IsActivityRunning.
func SetIMqttTransFailCnt(value int) {
	queMutex.Lock()
	defer queMutex.Unlock()
	IMqttTransFailCnt = value
}

func CheckIMqttTransFailCnt() {
	queMutex.Lock()
	defer queMutex.Unlock()

	if IMqttTransFailCnt > 0 {
		IMqttTransFailCnt++
		if IMqttTransFailCnt > 100 {
			IMqttTransFailCnt = 0
			public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_ERROR, public.ERROR_MACHINE_DOWN9, "")
			//reset cpu here.
			time.Sleep(3 * time.Second)
			ResetCPU()
		}
	}
}

func isFFMPEGRunning() bool {
	cmd := exec.Command("ps")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		return false
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "ffmpeg") {
			return true
		}
	}
	return false
}

func TestAllCameras() {
	defer func() {
		go Get4GProvider()
	}()
	if isFFMPEGRunning() {
		alog.Log.Println("TestAllCameras:ffmpeg is running, bypass")
		return
	}
	alog.Log.Println("TestAllCameras:killall ffmpeg")
	public.ExecShell("killall ffmpeg")
	time.Sleep(1000 * time.Millisecond)
	cameraCount, _ := strconv.Atoi(public.Config["CAMERA_COUNT"])
	if cameraCount > 4 {
		cameraCount = 4
	}
	if public.IsPubDualDoorMod {
		if cameraCount > 2 {
			cameraCount = 2
		}
		cameraCount *= 2
	}
	alog.Log.Println("TestAllCameras:cameraCount", cameraCount)
	for i := 0; i < cameraCount; i++ {
		snapShotHost := public.GetCameraHost()
		var snapShotHost2 string
		cameraType := public.Config["CAMERA_TYPE"]
		if cameraType == "haha" && public.IsPubDualDoorMod {
			snapShotHost2 = fmt.Sprintf(snapShotHost, i+0)
		} else {
			snapShotHost2 = fmt.Sprintf(snapShotHost, i+1)
		}
		cacheFileName := fmt.Sprintf("video/test/ijooz%d.mp4", i)
		cacheFileNameFull := public.GetTmpPath() + "/" + cacheFileName
		cacheFileNameFull = strings.ReplaceAll(cacheFileNameFull, "\\", "/")
		cacheFileNameFullPath, _ := filepath.Split(cacheFileNameFull)
		if !public.IsExist(cacheFileNameFullPath) {
			err := os.MkdirAll(cacheFileNameFullPath, 0777)
			if err != nil {
				continue
			}
			alog.Log.Println("new cacheFileNameFullPath: ", cacheFileNameFullPath)
		} else {
			alog.Log.Println("old cacheFileNameFullPath: ", cacheFileNameFullPath)
		}

		c := ""
		//cameraType := public.Config["CAMERA_TYPE"]
		if cameraType == "dev" {
			// dev摄像头
			c = fmt.Sprintf("ffmpeg -nostdin -y -use_wallclock_as_timestamps 1 -rtsp_transport tcp -i rtsp://%s:554/user=admin_password=123456_channel=1_stream=0 -timeout 3000000 -an -vcodec copy %s &", snapShotHost2, cacheFileNameFull)
		} else if cameraType == "ijooz" {
			// ijooz摄像头
			c = fmt.Sprintf("ffmpeg -nostdin -y -use_wallclock_as_timestamps 1 -rtsp_transport tcp -i rtsp://%s:8554/live0 -timeout 3000000 -an -vcodec copy %s &", snapShotHost2, cacheFileNameFull)
		} else if cameraType == "haha" {
			// 哈哈摄像头
			c = fmt.Sprintf("ffmpeg -nostdin -y -use_wallclock_as_timestamps 1 -rtsp_transport tcp -i rtsp://%s:554/0 -timeout 3000000 -an -vcodec copy %s &", snapShotHost2, cacheFileNameFull)
		} else {
			// 默认摄像头
			c = fmt.Sprintf("ffmpeg -nostdin -y -use_wallclock_as_timestamps 1 -rtsp_transport tcp -i rtsp://%s:8554/live0 -timeout 3000000 -vcodec copy %s &", snapShotHost2, cacheFileNameFull)
		}
		//note, use -t 3600 to make fixed duration video.
		alog.Log.Println("GET VIDEO CMD: ", c)

		if runtime.GOOS == "windows" {
			videoCmds[i] = exec.CommandContext(ctxVideo, "sh", "-c", c)
		} else {
			videoCmds[i] = exec.CommandContext(ctxVideo, "sh", "-c", c)
			videoCmds[i].SysProcAttr = &syscall.SysProcAttr{
				Setpgid: true,
			}
		}

		videoCmds[i].Stdout = os.Stdout
		videoCmds[i].Stderr = os.Stderr

		if err := videoCmds[i].Start(); err != nil {
			continue
		}

		// time.Sleep(4 * time.Second)
		// if !public.IsExist(cacheFileNameFull) {
		// 	return errors.New("start video error!")
		// }

		// if runtime.GOOS != "windows" {
		// 	errCmdCh := make(chan error, 1)
		// 	go func() {
		// 		errr := videoCmds[i].Wait()
		// 		fmt.Println("errr: ", errr)
		// 		errCmdCh <- errr
		// 	}()
	}

	time.Sleep(10 * time.Second)
	//stop all videos here
	alog.Log.Println("TestAllCameras:StopVideo")
	if err := StopVideo(); err != nil {
		alog.Log.Println("TestAllCameras:StopVideo fail")
	} else {
		alog.Log.Println("TestAllCameras:StopVideo done")
	}
	os.RemoveAll("/tmp/smartshop/video/test")
}

func Get4GProvider() {
	atcmd_mutex.Lock()
	defer atcmd_mutex.Unlock()
	var err error
	var n int
	if dtel != nil {
		alog.Log.Println("Get4GProvider:/dev/ttyUSB0 is open")
		defer func() {
			dtel.Close()
			dtel = nil
			fintel = nil
		}()
	} else {
		d := &serial.Config{Name: "/dev/ttyUSB0", Baud: 115200, ReadTimeout: time.Second * 1}
		dtel, err = serial.OpenPort(d)
		if err != nil {
			alog.Log.Println("Get4GProvider:/dev/ttyUSB0 busy")
			dtel = nil
			//return
		} else {
			alog.Log.Println("Get4GProvider:/dev/ttyUSB0 reopen")
		}
	}
	if etel != nil {
		alog.Log.Println("Get4GProvider:/dev/ttyUSB2 is open")
		defer func() {
			etel.Close()
			etel = nil
			fintel = nil
		}()
	} else {
		e := &serial.Config{Name: "/dev/ttyUSB2", Baud: 115200, ReadTimeout: time.Second * 1}
		etel, err = serial.OpenPort(e)
		if err != nil {
			alog.Log.Println("Get4GProvider:/dev/ttyUSB2 busy")
			etel = nil
			//return
		} else {
			alog.Log.Println("Get4GProvider:/dev/ttyUSB2 reopen")
		}
	}
	defer alog.Log.Println("Get4GProvider:done")
	time.Sleep(1 * time.Second)
	alog.Log.Println("Get4GProvider:begin")
	if dtel != nil {
		_, err = dtel.Write([]byte("AT\r\n"))
		if err != nil {
			alog.Log.Println("USB0 Write AT:", err)
		} else {
			time.Sleep(2 * time.Second)
			buf := make([]byte, 256)
			n, err = dtel.Read(buf)
			if err != nil {
				alog.Log.Println("USB0 Read AT:", err)
			} else if n > 0 {
				resp := string(buf[:n])
				if strings.Contains(resp, "OK") {
					isportusb0 = true
					fintel = dtel
					alog.Log.Println("Get4GProvider:USB0")
				}
			}
		}
	}

	if fintel == nil && etel != nil {
		_, err = etel.Write([]byte("AT\r\n"))
		if err != nil {
			alog.Log.Println("USB2 Write AT:", err)
		} else {
			time.Sleep(2 * time.Second)
			buf := make([]byte, 256)
			n, err = etel.Read(buf)
			if err != nil {
				alog.Log.Println("USB2 Read AT:", err)
			} else if n > 0 {
				resp := string(buf[:n])
				if strings.Contains(resp, "OK") {
					isportusb0 = false
					fintel = etel
					alog.Log.Println("get4GProvider:USB2")
				}
			}
		}
	}

	if fintel == nil {
		alog.Log.Println("Get4GProvider:fail")
		return
	}

	for x := 0; x < 3; x++ {
		StrTelProvider = "F"
		buf := make([]byte, 256)
		_, err = fintel.Write([]byte("AT+COPS?\r\n"))
		if err != nil {
			alog.Log.Println("Write AT+COPS:", err, x)
			continue
		}
		time.Sleep(2 * time.Second)
		n, err = fintel.Read(buf)
		if err != nil {
			alog.Log.Println("Read AT+COPS:", err, x)
			StrTelProvider += "01"
			time.Sleep(10 * time.Second)
			continue
		}
		//alog.Log.Println(buf)
		response := string(buf[:n])
		fields := strings.Split(response, ",")
		if len(fields) > 2 {
			provider := fields[2]
			fmt.Println("4G telecom provider:", provider)
			//str:="STAR"
			if strings.Contains(strings.ToLower(provider), "star") {
				fmt.Println("star")
				StrTelProvider = "H"
			} else if strings.Contains(strings.ToLower(provider), "sing") {
				fmt.Println("sing")
				StrTelProvider = "S"
			} else {
				fmt.Println("NA")
				StrTelProvider = "O"
			}
		} else {
			StrTelProvider = "W"
			alog.Log.Println("Wrong Pr:", response)
		}
		// Send the AT command to get the signal quality
		_, err = fintel.Write([]byte("AT+CSQ\r\n"))
		if err != nil {
			alog.Log.Println("Write AT+CSQ:", err, x)
			StrTelProvider += "00"
			continue
		}
		time.Sleep(2 * time.Second)
		// Read the response
		n, err = fintel.Read(buf)
		if err != nil {
			alog.Log.Println("Read AT+CSQ:", err, x)
			StrTelProvider += "02"
			time.Sleep(10 * time.Second)
			continue
		}
		// Parse the response to get the signal quality
		response = string(buf[:n])
		if strings.Contains(response, "+CSQ:") {
			parts := strings.Split(response, " ")
			if len(parts) > 1 {
				quality := strings.TrimSpace(parts[1])
				qs := strings.Split(quality, ",")
				alog.Log.Println("Signal quality:", qs[0])
				qs2 := fmt.Sprintf("%02s", qs[0])
				StrTelProvider += qs2
			}
		} else {
			alog.Log.Println("Error getting signal quality")
			StrTelProvider += "03"
		}

		StrSimCardNumb = ""
		_, err = fintel.Write([]byte("AT+CCID\r\n"))
		if err != nil {
			alog.Log.Println("Write AT+CCID:", err, x)
			StrTelProvider += "00"
			continue
		}
		time.Sleep(2 * time.Second)
		// Read the response
		n, err = fintel.Read(buf)
		if err != nil {
			alog.Log.Println("Read AT+CCID:", err, x)
			time.Sleep(10 * time.Second)
			continue
		}
		// Parse the response to get the signal quality
		response = string(buf[:n])
		if strings.Contains(response, "+CCID:") {
			parts := strings.Split(response, " ")
			if len(parts) > 1 {
				//fmt.Printf("%q", parts[1])
				//alog.Log.Println("Get CCID1:", StrSimCardNumb)
				StrSimCardNumb = strings.TrimSpace(parts[1])
				//alog.Log.Println("Get CCID2:", StrSimCardNumb)
				StrSimCardNumb = strings.Trim(StrSimCardNumb, "\r\n")
				StrSimCardNumb = strings.Trim(StrSimCardNumb, "\r\n")
				//alog.Log.Println("Get CCID3:", StrSimCardNumb)
				StrSimCardNumb = strings.TrimSuffix(StrSimCardNumb, "OK")
				StrSimCardNumb = strings.TrimSuffix(StrSimCardNumb, "\r\n")
				StrSimCardNumb = strings.TrimSuffix(StrSimCardNumb, "\r\n")
				alog.Log.Println("Get CCID:", StrSimCardNumb)
			} else {
				alog.Log.Println("Error SIMcardnumber1", response)
			}
		} else {
			alog.Log.Println("Error SIMcardnumber2", response)
		}
		return
	}
}

// AT+QGDCNT?
func Get4GDataUsage() {
	atcmd_mutex.Lock()
	defer atcmd_mutex.Unlock()
	var f *serial.Config
	var err error
	if isportusb0 {
		f = &serial.Config{Name: "/dev/ttyUSB0", Baud: 115200, ReadTimeout: time.Second * 1}
		alog.Log.Println("Get4GDataUsage:/dev/ttyUSB0")
	} else {
		f = &serial.Config{Name: "/dev/ttyUSB2", Baud: 115200, ReadTimeout: time.Second * 1}
		alog.Log.Println("Get4GDataUsage:/dev/ttyUSB2")
	}
	fintel, err = serial.OpenPort(f)
	if err != nil {
		alog.Log.Println("Get4GDataUsage:port busy")
		fintel = nil
	} else {
		alog.Log.Println("Get4GDataUsage:port reopen")
	}
	if fintel == nil {
		return
	}
	defer func() {
		fintel.Close()
		fintel = nil
	}()
	buf := make([]byte, 256)
	_, err = fintel.Write([]byte("AT+QGDCNT?\r\n"))
	if err != nil {
		alog.Log.Println("Write AT+QGDCNT:", err)
		return
	}
	time.Sleep(2 * time.Second)
	n := 0
	n, err = fintel.Read(buf)
	if err != nil {
		alog.Log.Println("Read AT+QGDCNT:", err)
		return
	}
	//alog.Log.Println(buf)
	response := string(buf[:n])
	if strings.Contains(response, "+QGDCNT:") {
		fields := strings.Split(response, " ")
		if len(fields) > 1 {
			strusage := fields[1]
			parts := strings.Split(strusage, "\r\n")
			rxtxs := strings.Split(parts[0], ",")
			if len(rxtxs) > 1 {
				StrTxByteNum = rxtxs[0]
				StrRxByteNum = rxtxs[1]
				alog.Log.Println(StrTxByteNum, StrRxByteNum)
			} else {
				alog.Log.Println("Err DataUsage1:", response)
			}
		} else {
			alog.Log.Println("Err DataUsage2:", response)
		}
	}
	//AT+QGDCNT?
}
