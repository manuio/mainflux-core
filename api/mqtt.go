package api

import (
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
)

var (
	// MqttClient is used in HTTP server to communicate HTTP value updates/requests
	mqttClient mqtt.Client
)

//define a function for the default message handler
var msgHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())

	s := strings.Split(msg.Topic(), "/")
	chanID := s[len(s)-1]
	writeChannel(chanID, msg.Payload())
}

// MqttSub function - we subscribe to topic `mainflux/#` (no trailing `/`)
func (mqc *MqttConn) MqttSub(cfg config.Config) {
	// Create a ClientOptions struct setting the broker address, clientid, turn
	// off trace output and set the default message handler
	mqc.Opts = mqtt.NewClientOptions().AddBroker("tcp://" + cfg.MQTTHost + ":" + strconv.Itoa(cfg.MQTTPort))
	mqc.Opts.SetClientID("mainflux")
	mqc.Opts.SetDefaultPublishHandler(msgHandler)

	//create and start a client using the above ClientOptions
	mqc.Client = mqtt.NewClient(mqc.Opts)
	if token := mqc.Client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Subscribe to all channels of all the devices and request messages to be delivered
	// at a maximum qos of zero, wait for the receipt to confirm the subscription
	// Topic is in the form:
	// mainflux/<channel_id>
	if token := mqc.Client.Subscribe("mainflux/#", 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}

	mqttClient = mqc.Client
}
