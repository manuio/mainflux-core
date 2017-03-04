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
)

type (
	NatsMsg struct {
		Channel   string `json:"channel"`
		Publisher string `json:"publisher"`
		Protocol  string `json:"protocol"`
		Payload   []byte `json:"payload"`
	}
)

var (
	NatsConn *nats.Conn
)

func msgHandler(nm *nats.Msg) {
	fmt.Printf("Received a message: %s\n", string(nm.Data))

	// Re-publish it
	NatsConn.Publish("mainflux/core/out", nm.Data)

	// And write it into the database
	m := NatsMsg{}
	if len(nm.Data) > 0 {
		if err := json.Unmarshal(nm.Data, &m); err != nil {
			println("Can not decode NATS msg")
			return
		}
	}

	println("Calling writeMessage()")
	fmt.Println(m.Publisher, m.Protocol, m.Channel, m.Payload)
	writeMessage(m)
}

func NatsInit(host string, port int) error {
	/** Connect to NATS broker */
	var err error
	NatsConn, err = nats.Connect("nats://" + host + ":" + strconv.Itoa(port))
	if err != nil {
		log.Fatalf("NATS: Can't connect: %v\n", err)
	}

	// Create MQTT bridge
	NatsConn.Subscribe("mainflux/core/in", msgHandler)

	return err
}
