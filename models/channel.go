/**
 * Copyright (c) Mainflux
 *
 * Mainflux server is licensed under an Apache license, version 2.0.
 * All rights not explicitly granted in the Apache license, version 2.0 are reserved.
 * See the included LICENSE file for more details.
 */

package models

import (
	"github.com/krylovsk/gosenml"
)

type (
	// Channel is a bidirectional pipe of communication
	// between devices and applications.
	//
	// SENSOR: If a channel is used for sending the measirements from
	// the sensor of the device, the usual case it that device
	// writes into th channel and applications listen.
	//
	// ACTUATOR: If a channel is used for triggering action (switches, buttons, relays)
	// and similar then application must publish the message into the channel, and
	// device must be subscribed to the channel.
	//
	// Channels are tightly connected to MQTT topics - one channel ID corresponds to one topic.
	Channel struct {
		ID string `json:"id"`

		// ID of device to which this channel belongs to.
		// Channels always belong to one device which uses them to
		// publish the info of it's properties, or to listen on them
		// messages that applications send to this device.
		Device string `json:"device"`

		Values []gosenml.Entry `json:"values"`

		Created string `json:"created"`
		Updated string `json:"updated"`

		Metadata map[string]interface{} `json:"metadata"`
	}
)
