/**
 * Copyright (c) Mainflux
 *
 * Mainflux server is licensed under an Apache license, version 2.0.
 * All rights not explicitly granted in the Apache license, version 2.0 are reserved.
 * See the included LICENSE file for more details.
 */

package api

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/go-nats"
	"log"
	"strconv"
	"strings"
)

type (
	MqttMsg struct {
		Topic     string `json:"topic"`
		Publisher string `json:"publisher"`
		Payload   []byte `json:"payload"`
	}
)

var (
	NatsConn *nats.Conn
)

func mqttHandler(nm *nats.Msg) {
	fmt.Printf("Received a message: %s\n", string(nm.Data))

	m := MqttMsg{}
	if len(nm.Data) > 0 {
		if err := json.Unmarshal(nm.Data, &m); err != nil {
			println("Can not decode MQTT msg")
			return
		}
	}

	s := strings.Split(m.Topic, "/")
	channelID := s[len(s)-1]

	println("Calling writeMessage()")
	fmt.Println(m.Publisher, channelID, m.Payload)
	writeMessage(m.Publisher, channelID, m.Payload)
}

func NatsInit(host string, port int) error {
	/** Connect to NATS broker */
	var err error
	NatsConn, err = nats.Connect("nats://" + host + ":" + strconv.Itoa(port))
	if err != nil {
		log.Fatalf("NATS: Can't connect: %v\n", err)
	}

	// Create MQTT bridge
	NatsConn.Subscribe("mainflux/mqtt/core", mqttHandler)

	return err
}
