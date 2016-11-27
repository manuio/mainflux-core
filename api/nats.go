/**
 * Copyright (c) Mainflux
 *
 * Mainflux server is licensed under an Apache license, version 2.0.
 * All rights not explicitly granted in the Apache license, version 2.0 are reserved.
 * See the included LICENSE file for more details.
 */

package api

import (
	"github.com/nats-io/go-nats"
	"log"
	"strconv"
)

var (
	NatsConn *nats.Conn
)

func InitNats(host string, port int) error {
	/** Connect to NATS broker */
	println(host, port)
	var err error
	NatsConn, err = nats.Connect("nats://" + host + ":" + strconv.Itoa(port))
	if err != nil {
		log.Fatalf("NATS: Can't connect: %v\n", err)
	}

	return err
}
