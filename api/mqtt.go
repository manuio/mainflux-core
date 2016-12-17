package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mainflux/mainflux-core/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type (
	// MqttConn struct
	MqttConn struct {
		Opts   *mqtt.ClientOptions
		Client mqtt.Client
	}

	MqttMsg struct {
		Topic     string `json:"topic"`
		Publisher string `json:"publisher"`
		Payload   []byte `json:"payload"`
	}
)

var (
	// MqttClient is used in HTTP server to communicate HTTP value updates/requests
	mqttClient       mqtt.Client
	mainfluxCoreUUID string
)

//define a function for the default message handler
var msgHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())

	m := MqttMsg{}
	if len(msg.Payload()) > 0 {
		if err := json.Unmarshal(msg.Payload(), &m); err != nil {
			println("Can not decode MQTT msg")
			return
		}
	}

	s := strings.Split(m.Topic, "/")
	channelID := s[len(s)-1]

	writeMessage(m.Publisher, channelID, m.Payload)
}

// MqttSub function - we subscribe to topic `mainflux/channels/#` (no trailing `/`)
func (mqc *MqttConn) MqttSub(cfg config.Config) {
	// Create a ClientOptions struct setting the broker address, clientid, turn
	// off trace output and set the default message handler
	mqc.Opts = mqtt.NewClientOptions().AddBroker("tcp://" + cfg.MQTTHost + ":" + strconv.Itoa(cfg.MQTTPort))

	// A UUID is a 16-octet (128-bit) number.
	// In its canonical form, a UUID is represented by 32 lowercase hexadecimal digits,
	// displayed in five groups separated by hyphens, in the form 8-4-4-4-12 for a
	// total of 36 characters (32 alphanumeric characters and four hyphens).
	//
	// For example:
	// 123e4567-e89b-12d3-a456-426655440000
	mainfluxCoreUUID = "12345678-1234-1234-1234-123456789012"
	mqc.Opts.SetClientID(mainfluxCoreUUID)
	mqc.Opts.SetDefaultPublishHandler(msgHandler)

	//create and start a client using the above ClientOptions
	mqc.Client = mqtt.NewClient(mqc.Opts)
	if token := mqc.Client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Subscribe to all channels of all the devices and request messages to be delivered
	// at a maximum qos of zero, wait for the receipt to confirm the subscription
	// Topic is in the form:
	// mainflux/channels/#
	if token := mqc.Client.Subscribe("mainflux/system/messages", 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}

	mqttClient = mqc.Client
}
