package service

import (
	alog "agent/logger"
	"agent/public"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func onMqttChannel(ch chan string, client MQTT.Client) {
	for {
		input := <-ch
		inputs := strings.SplitN(input, ":", 3)
		if len(inputs) == 3 && inputs[0] == public.CHANNEL_TYPE_MQTT {
			client.Publish(inputs[1], 1, false, inputs[2])
		}
	}
}

func onMqttMessage(client MQTT.Client, message MQTT.Message) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("onMqttMessage recover:", err)
		}
	}()

	alog.Log.Printf("MQTT MSG: %s: %s\n", message.Topic(), message.Payload())
	msg := message.Payload()
	if len(msg) < 4 {
		alog.Log.Println("MQTT MSG ignored: message too short:", len(msg))
		return
	}

	sType := string(msg[0:2])
	sAction := string(msg[2:4])
	sData := string(msg[4:])
	if sType == public.TYPE_CMD {
		public.ChCmd <- fmt.Sprintf("%s:%s", sAction, sData)
		alog.Log.Println("TYPE_CMD:", sAction, sData)
		return
	}

	alog.Log.Println("MQTT MSG ignored:", sType, sAction, sData)
	public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_ERROR, public.ERROR_PARAM, "")
}

func MqttStart(server string, clientid string, username string, password string, topic1 string, topic2 string, qos int) {
	alog.Log.Println("MQTT init start")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	lostCh := make(chan error, 1)

	connOpts := MQTT.NewClientOptions().AddBroker(server).SetClientID(clientid).SetCleanSession(true)
	if username != "" {
		connOpts.SetUsername(username)
		if password != "" {
			connOpts.SetPassword(password)
		}
	}
	connOpts.AutoReconnect = false

	connOpts.OnConnect = func(c MQTT.Client) {
		alog.Log.Println("MQTT connect done: ok")
		fatalCnt := 0
		for {
			time.Sleep(1 * time.Second)
			if token := c.Subscribe(topic1, byte(qos), onMqttMessage); token.Wait() && token.Error() != nil {
				fatalCnt++
				if fatalCnt > 3 {
					alog.Log.Fatalln("MQTT SUBSCRIBE FAIL:", token.Error())
				}
				continue
			}
			break
		}
		time.Sleep(1 * time.Second)
		if token := c.Subscribe(topic2, byte(qos), onMqttMessage); token.Wait() && token.Error() != nil {
			alog.Log.Println("MQTT SUBSCRIBE FAIL:", topic2, token.Error())
		}
		alog.Log.Println("MQTT SUBSCRIBE SUCCESS:", topic1, topic2)

		defer time.AfterFunc(2*time.Second, func() {
			data := map[string]string{
				"VER1": public.VERSION,
				"VER2": "1.02",
				"LONG": "na",
				"LALT": "na",
			}
			dataStr, err := json.Marshal(&data)
			if err == nil {
				public.SendMqttStatus(public.TYPE_DEVICE, public.ACTION_STARTRUN, string(dataStr), "")
			}
			alog.Log.Println("Inform agent is running by MQTT")
		})
	}

	connOpts.OnConnectionLost = func(c MQTT.Client, e error) {
		alog.Log.Println("MQTT connection lost:", e)
		select {
		case lostCh <- e:
		default:
		}
	}

	client := MQTT.NewClient(connOpts)
	go onMqttChannel(public.ChMqtt, client)

	failures := 0
	for {
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			failures++
			delay := mqttRetryDelay(failures)
			alog.Log.Printf("MQTT connect done: fail attempts=%d next_retry=%s error=%v\n", failures, delay, token.Error())
			if waitForMqttRetry(c, delay) {
				return
			}
			continue
		}

		failures = 0
		alog.Log.Printf("MQTT connected to %s\n", server)
		select {
		case <-c:
			client.Disconnect(250)
			return
		case err := <-lostCh:
			failures++
			delay := mqttRetryDelay(failures)
			alog.Log.Printf("MQTT reconnect pending: attempts=%d next_retry=%s reason=%v\n", failures, delay, err)
			if waitForMqttRetry(c, delay) {
				return
			}
		}
	}
}

func mqttRetryDelay(failures int) time.Duration {
	if failures > 5 {
		return 2 * time.Hour
	}
	return 30 * time.Minute
}

func waitForMqttRetry(c chan os.Signal, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-c:
		return true
	case <-timer.C:
		return false
	}
}

func InitFFmpeg() {
	alog.Log.Println("InitFFmpeg ignored: camera/FFmpeg flow is disabled")
}
