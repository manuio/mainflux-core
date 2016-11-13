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
	// Message struct
	Message struct {
		// User generated ID - keep it unique
		ID string `json:"id"`

		// True if we need to reply
		// Reply will be sent over `mainflux/reply/<chan_ID>/<msg_ID>`
		Reply bool `json:"reply"`

		// Is message realyed (HTTP-to-MQTT or LWM2M-to-MQTT),
		// i.e. if it was published by Mainflux Core
		// (after is has been recieved from non-MQTT protocol).
		// This is important to know, because realyed messages are
		// already persisted in the database by Mainflux Core
		// (it saves them in the database before publishing them on MQTT),
		// so when recieved as a loopback over MQTT they can be ignored.
		//
		// By default this should be `false`, Mainflux Core will put it to
		// `true` if it relays the message.
		Relayed bool `json:"relayed"`

		// Sender ID - MQTT does not have notion of a sender,
		// so we keep sender ID inside the message
		Sender string `json:"sender"`

		// Mesage content is SenML
		SenML gosenml.Message `json:"senml"`

		// Message creation timestamp
		Created string `json:"created"`
	}
)
