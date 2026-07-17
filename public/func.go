package public

import (
	alog "agent/logger"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/nfnt/resize"
)

func SendMqttStatus(cType, cAction, cData, cId string) {
	if cType == TYPE_DEVICE {
		ChMqtt <- fmt.Sprintf("%s:%s:%s", CHANNEL_TYPE_MQTT, TOPIC_STATUS_DEVICE+Config["SN"], TYPE_DEVICE+cAction+cData)
	} else if cType == TYPE_CMD {
		ChMqtt <- fmt.Sprintf("%s:%s:%s", CHANNEL_TYPE_MQTT, TOPIC_STATUS_CMD+Config["SN"], TYPE_CMD+cAction+cData)
	} else {
		ChMqtt <- fmt.Sprintf("%s:%s:%s", CHANNEL_TYPE_MQTT, TOPIC_STATUS_OTHER, TYPE_OTHER+cAction+cData)
	}
}

func HttpRequest(method, url string, data map[string]string) (error, int, []byte) {
	// return nil, 200, nil
	/*ctx, cancel := context.WithCancel(context.TODO())
	timer := time.AfterFunc(5*time.Second, func() {
		cancel()
	})
	defer timer.Stop()*/
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var r http.Request
	r.ParseForm()
	for k, v := range data {
		r.Form.Add(k, v)
	}
	bodyStr := strings.TrimSpace(r.Form.Encode())
	request, err := http.NewRequest(method, url, strings.NewReader(bodyStr))
	if err != nil {
		return err, 404, nil
	}
	request = request.WithContext(ctx)
	request.Header.Set("User-Agent", "smartshop-agent/"+VERSION)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Connection", "Keep-Alive")
	request.Header.Set("Auth", API_AUTH)

	// Create an HTTP client with default settings
	//client := &http.Client{}

	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		return err, 404, nil
	}
	defer resp.Body.Close()
	status := resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err, status, nil
	}
	return nil, status, body
}

func HttpGet(url string) (error, int, []byte) {
	// return nil, 200, nil
	ctx, cancel := context.WithCancel(context.TODO())
	timer := time.AfterFunc(3*time.Second, func() {
		cancel()
	})
	defer timer.Stop()
	request, err := http.NewRequest("GET", url, strings.NewReader(""))
	if err != nil {
		return err, 404, nil
	}
	request = request.WithContext(ctx)

	request.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:admin")))
	request.Header.Set("Connection", "Keep-Alive")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		return err, 404, nil
	}
	defer resp.Body.Close()
	status := resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err, status, nil
	}
	return nil, status, body
}

func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

func IsFile(f string) bool {
	fi, e := os.Stat(f)
	if e != nil {
		return false
	}
	return !fi.IsDir()
}

func ExecShell(s string) (error, string) {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	alog.Log.Println("ExecShell:", s)
	cmds := strings.Split(s, " ")
	var cmd *exec.Cmd
	if len(cmds) == 1 {
		cmd = exec.Command(cmds[0])
	} else {
		cmd = exec.Command(cmds[0], cmds[1:]...)
	}

	// Set process group ID for the command to start in a separate process group
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	//读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	var out bytes.Buffer
	cmd.Stdout = &out

	//Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	//err := cmd.Run()
	err := cmd.Start()
	if err != nil {
		//log.Fatal(err)
		return err, out.String()
	}
	timeout := time.After(30 * time.Second) // set a 30-second timeout
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-timeout:
		err = fmt.Errorf("command timed out\r\n")
	case err = <-done:
		if err != nil {
			//log.Fatal(err)
			return err, out.String()
		}
	}
	//state := cmd.ProcessState
	//log.Println("exit code:", state.ExitCode(), err)
	return err, out.String()
}

func ExecWget(url, outputPath string) (error, string) {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	alog.Log.Println("ExecWget:", url, outputPath)
	cmd := exec.Command("wget", url, "-O", outputPath)
	// Set process group ID for the command to start in a separate process group
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	//读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	var out bytes.Buffer
	cmd.Stdout = &out

	//Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	//err := cmd.Run()
	err := cmd.Start()
	if err != nil {
		//log.Fatal(err)
		return err, out.String()
	}
	timeout := time.After(60 * time.Second) // set a 60-second timeout
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-timeout:
		err = fmt.Errorf("command timed out\r\n")
	case err = <-done:
		if err != nil {
			//log.Fatal(err)
			return err, out.String()
		}
	}
	//state := cmd.ProcessState
	//log.Println("exit code:", state.ExitCode(), err)
	return err, out.String()
}

func ExecWgetEn(url string) (error, string) {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	alog.Log.Println("ExecWgetEn:", url)
	//0404wget -P -N /tmp http://upload.shop.ijooz.sg/agent/md5.upgrade
	cmd := exec.Command("wget", "-P", "/tmp", url)
	// Set process group ID for the command to start in a separate process group
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	//读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	var out bytes.Buffer
	cmd.Stdout = &out

	//Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	//err := cmd.Run()
	err := cmd.Start()
	if err != nil {
		//log.Fatal(err)
		return err, out.String()
	}
	timeout := time.After(60 * time.Second) // set a 60-second timeout
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-timeout:
		err = fmt.Errorf("command timed out\r\n")
	case err = <-done:
		if err != nil {
			//log.Fatal(err)
			return err, out.String()
		}
	}
	//state := cmd.ProcessState
	//log.Println("exit code:", state.ExitCode(), err)
	return err, out.String()
}

func GetCurPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	rst := filepath.Dir(path)
	if len(os.Args) > 1 {
		if os.Args[1] == "debug" {
			file, _ := os.Getwd()
			rst = file
		}
	}
	return rst
}

func GetSdcardPath() string {
	if len(os.Args) > 1 {
		if os.Args[1] == "debug" {
			return GetCurPath()
		}
	}
	//return "/mnt/mmcblk0p1"
	if IsMountedSdCard && !IsSdCardNotFind {
		return "/mnt/mmcblk0p1"
	} else {
		return GetTmpPath()
	}
}

func DetectTmpSmartShopFolder() bool {
	_, err := os.Stat("/tmp/smartshop")
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir("/tmp/smartshop", os.ModePerm)
			if err != nil {
				//log.Fatal(err)
				alog.Log.Println("Create /tmp/smartshop directory fail")
				return false
			}
			alog.Log.Println("Create /tmp/smartshop directory if it doesn't exist")
			return true
		} else {
			//log.Fatal(err)
			alog.Log.Println(err)
			return false
		}
	}
	alog.Log.Println("/tmp/smartshop directory exists")
	return true
}

func GetTmpVideoMegaSize() (totalSize int64, err error) {
	Tmpsize_mutex.Lock()
	defer Tmpsize_mutex.Unlock()
	// fi, err := os.Stat("/tmp/smartshop/video")
	// if err != nil {
	// 	//log.Fatal(err)
	// 	log.Println(err)
	// 	return 256
	// }
	// log.Printf("/tmp/smartshop/video size: %d bytes %dMbytes\n", fi.Size(), fi.Size()/1024/1024)
	// return fi.Size() / 1024 / 1024

	dirPath := "/tmp/smartshop/video"
	if _, err = os.Stat(dirPath); os.IsNotExist(err) {
		// create directory if not exists
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			alog.Log.Println(err)
			return 256, err
		}
		alog.Log.Println("MkdirAll:", dirPath)
	}

	err = filepath.Walk("/tmp/smartshop/video", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		//log.Fatal(err)
		alog.Log.Println(err)
		return 256, err
	}
	alog.Log.Printf("/video size: %d bytes %d MB\n", totalSize, totalSize/1024/1024)
	return totalSize / 1024 / 1024, nil
}

func GetTmpPath() string {
	if len(os.Args) > 1 {
		if os.Args[1] == "debug" {
			return GetCurPath()
		}
	}
	bTmpHas := DetectTmpSmartShopFolder()
	var err error
	if bTmpHas {
		_, err = os.Stat("/tmp/smartshop/cache")
		if os.IsNotExist(err) {
			err = os.Mkdir("/tmp/smartshop/cache", os.ModePerm)
		}
		_, err = os.Stat("/tmp/smartshop/video")
		if os.IsNotExist(err) {
			err = os.Mkdir("/tmp/smartshop/video", os.ModePerm)
		}
		_, err = os.Stat("/tmp/smartshop/image")
		if os.IsNotExist(err) {
			err = os.Mkdir("/tmp/smartshop/image", os.ModePerm)
		}
	}
	var siz int64
	siz, err = GetTmpVideoMegaSize()
	//logger.Logger.Println()
	alog.Log.Printf("GetTmpPath:tmp-video size is %d M", siz)
	return "/tmp/smartshop"
}

func GetImgPath(activityId string, i int, bfType string) string {
	year := activityId[0:4]
	month := activityId[4:6]
	day := activityId[6:8]
	return fmt.Sprintf("image/%s/%s/%s/%s_%d_%s.jpg", year, month, day, activityId, i, bfType)
}

func GetVideoPath(activityId string, i int) string {
	//log.Println("GetVideoPath:", activityId, i)
	year := activityId[0:4]
	month := activityId[4:6]
	day := activityId[6:8]
	return fmt.Sprintf("video/%s/%s/%s/%s_%d.mp4", year, month, day, activityId, i)
}

func DeleteFileOnDisk(localPath string) error {
	if err := os.Remove(localPath); err != nil {
		return err
	}
	return nil
}

func DeleteEpyFileFolder(localPath string) error {
	if err := os.RemoveAll(localPath); err != nil {
		return err
	}
	return nil
}

func ResizeJpg(imgPath string, newWidth uint) error {
	// open "test.jpg"
	file, err := os.Open(imgPath)
	if err != nil {
		return err
	}

	// decode jpeg into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		return err
	}
	file.Close()

	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	m := resize.Resize(newWidth, 0, img, resize.Lanczos3)

	out, err := os.Create(imgPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// write new image to file
	jpeg.Encode(out, m, nil)
	return nil
}

func GetCameraHost() string {
	re := "192.168.199.10%d"
	cameraType := Config["CAMERA_TYPE"]
	if cameraType == "haha" {
		re = "192.168.199.3%d"
	}
	if cameraType == "ijooz" {
		re = "192.168.199.10%d"
	}
	if cameraType == "test" {
		re = "192.168.99.10%d"
	}
	return re
}

func testWrite(path string) error {
	// Remove the file if it exists.
	os.Remove(path + "/sd.txt")
	// Create an empty file.
	file, err := os.Create(path + "/sd.txt")
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

func testRead(filePath string) error {
	// Read the content of the file.
	_, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return nil
}

func CheckSdCardMounted() {
	IsMountedSdCard = false
	sdCardPath := "/mnt/mmcblk0p1"
	filePath := sdCardPath + "/sd.txt"
	if !IsExist(filePath) {
		alog.Log.Println("SD card init done: fail not mounted")
		return
	}

	if ICheckSdCardCnt < 3 {
		if err := testWrite(sdCardPath); err != nil {
			alog.Log.Println("SD card init done: fail write", err)
			return
		}
		if err := testRead(filePath); err != nil {
			alog.Log.Println("SD card init done: fail read", err)
			return
		}
	}

	ICheckSdCardCnt++
	if ICheckSdCardCnt > 1000 {
		ICheckSdCardCnt = 0
	}
	IsMountedSdCard = true
	alog.Log.Println("SD card init done: ok", sdCardPath)
}

func OpenFile(filename string) (*os.File, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("文件不存在")
		return os.Create(filename) //创建文件
	}
	fmt.Println("文件存在")
	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755) //打开文件
}
func WriteFile(filename string, data string) {
	f1, err1 := OpenFile(filename)
	if err1 != nil {
		alog.Log.Fatal(err1.Error())
	}
	defer f1.Close()
	n, err1 := io.WriteString(f1, data) //写入文件(字符串)
	if err1 != nil {
		alog.Log.Fatal(err1.Error())
	}
	fmt.Printf("写入 %d 个字节\n", n)
}

func CopyFile(src, dst string) error {
	// Read the content of the source file
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Write the content to the destination file
	err = os.WriteFile(dst, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
