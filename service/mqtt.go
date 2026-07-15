package service

import (
	"agent/device"
	alog "agent/logger"
	"agent/public"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func init() {

}

func onMqttChannel(ch chan string, client MQTT.Client) {
	var input string
	for {
		input = <-ch
		inputs := strings.Split(input, ":")
		if len(inputs) == 3 && inputs[0] == public.CHANNEL_TYPE_MQTT {
			client.Publish(inputs[1], 1, false, inputs[2])
		}
	}
}

func onMqttMessage(client MQTT.Client, message MQTT.Message) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("onMqttMessage 捕获异常:", err)
		}
	}()
	alog.Log.Printf("MQTT MSG: %s: %s\n", message.Topic(), message.Payload())
	//0101:1629113854954090496:1
	//topic := message.Topic()
	msg := message.Payload()
	if len(msg) < 4 {
		alog.Log.Println("MQTT MSG: msg too short:", len(msg))
		//delete it
		return
	}
	sType := string(msg[0:2])
	sAction := string(msg[2:4])
	sData := string(msg[4:])
	bDualMod := false
	if sType == public.TYPE_CMD {
		//执行一个指令
		sData = strings.Replace(string(sData), ":", "=>", -1)
		public.ChCmd <- fmt.Sprintf("%s:%s", sAction, sData)
		alog.Log.Println("TYPE_CMD:", sAction, sData)
		return
	}
	if sType == public.TYPE_ACTIVITY && sAction == public.ACTION_BEGIN {
		//开始一个新的activity
		if len(sData) < 10 {
			alog.Log.Println("MQTT MSG: activityId too short:", len(sData))
			//delete it
			return
		}
		activityId := sData
		needUploadVideo := "1"
		if strings.Index(sData, ":") > -1 {
			inputs := strings.Split(sData, ":")
			if len(inputs) == 2 {
				activityId = inputs[0]
				needUploadVideo = inputs[1]
			} else if len(inputs) == 3 {
				activityId = inputs[0]
				needUploadVideo = inputs[1]
				if inputs[2] == "B" && public.IsPubDualDoorMod {
					bDualMod = true
					alog.Log.Println("MQTT MSG: B Door Trans")
				} else {
					alog.Log.Println("MQTT MSG: A Door Trans")
				}
			}
		}
		match, _ := regexp.MatchString("([A-Z0-9]+)", activityId)
		if match {
			if bDualMod {
				public.ChActivity <- fmt.Sprintf("%s:%s:%s:%s", public.CHANNEL_TYPE_ACTIVITYEN, public.ACTION_BEGIN, activityId, needUploadVideo)
			} else {
				public.ChActivity <- fmt.Sprintf("%s:%s:%s:%s", public.CHANNEL_TYPE_ACTIVITY, public.ACTION_BEGIN, activityId, needUploadVideo)
			}
			client.Publish(public.TOPIC_STATUS_ACTIVITY+activityId, 1, false, public.TYPE_ACTIVITY+public.ACTION_BEGINED)
			device.SetIMqttTransFailCnt(1)
			return
		}
	}
	client.Publish(public.TOPIC_STATUS_DEVICE+public.Config["SN"], 1, false, public.TYPE_DEVICE+public.ACTION_ERROR+public.ERROR_PARAM)
}

func MqttStart(server string, clientid string, username string, password string, topic1 string, topic2 string, qos int) {
	alog.Log.Println("MQTT START")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	connOpts := MQTT.NewClientOptions().AddBroker(server).SetClientID(clientid).SetCleanSession(true)
	if username != "" {
		connOpts.SetUsername(username)
		if password != "" {
			connOpts.SetPassword(password)
		}
	}
	// tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	// connOpts.SetTLSConfig(tlsConfig)
	connOpts.MaxReconnectInterval = 60 * time.Second
	connOpts.AutoReconnect = true

	connOpts.OnConnect = func(c MQTT.Client) {
		device.IsOnline = true
		alog.Log.Println("MQTT ONCONNECT！")
		faltalcnt := 0
		for {
			time.Sleep(1 * time.Second)
			if token1 := c.Subscribe(topic1, byte(qos), onMqttMessage); token1.Wait() && token1.Error() != nil {
				faltalcnt++
				if faltalcnt > 3 {
					//panic(token1.Error())
					alog.Log.Fatalln("MQTT SUBSCRIBE FAIL！ REBOOT!", token1)
				}
				continue
			}
			break
		}
		time.Sleep(1 * time.Second)
		if token2 := c.Subscribe(topic2, byte(qos), onMqttMessage); token2.Wait() && token2.Error() != nil {
			// panic(token2.Error())
			alog.Log.Println("MQTT SUBSCRIBE FAIL！", topic2)
		}
		alog.Log.Println("MQTT SUBSCRIBE SUCCESS！", topic1, topic2)

		public.ExecShell("ifconfig wwan0 mtu 1300")
		public.ExecShell("ifconfig usb0 mtu 1300")

		defer time.AfterFunc(2*time.Second, func() {
			//send MQTT 7621-start run to Server
			data := make(map[string]string)
			data["VER1"] = public.VERSION
			data["VER2"] = "1.02"
			data["LONG"] = "na"
			data["LALT"] = "na"
			dataStr, err := json.Marshal(&data)
			if err == nil {
				dataStr2 := strings.Replace(string(dataStr), ":", "=>", -1)
				public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_STARTRUN, dataStr2, "")
			}
			alog.Log.Println("Inform 7621 is running by MQTT！")
		})
	}

	connOpts.OnConnectionLost = func(c MQTT.Client, e error) {
		device.IsOnline = false
		alog.Log.Println("MQTT ONCONNECT_LOST！")
	}

	client := MQTT.NewClient(connOpts)

	go onMqttChannel(public.ChMqtt, client)

CONNECT:
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		time.Sleep(10 * time.Second)
		alog.Log.Printf("MQTT CONNECT ERROR, RETRY!\n")
		goto CONNECT
	} else {
		alog.Log.Printf("MQTT CONNECTED TO %s\n", server)
	}

	<-c
}

func InitFFmpeg() {
	public.ChActivity <- fmt.Sprintf("%s:%s:%s:%s", public.CHANNEL_TYPE_ACTIVITY, public.ACTION_BEGIN, "1111000011110000000", "3")
}
