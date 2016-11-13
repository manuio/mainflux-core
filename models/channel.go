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
	// Channel Entry - a record of recieved data in the channel
	Entry struct {
		// Message sent by the device or app on the channel
		Msg Message `json:"msg"`
		// Parsed SenML from messages, i.e. derived values
		Values []gosenml.Entry `json:"values"`
		// Timestamp of Entry
		Timestamp string `json:"timestamp"`
	}

	// Channel struct
	Channel struct {
		ID     string `json:"id"`
		Device string `json:"device"`

		// Name is optional. If present, it is pre-pended to `bn` member of SenML.
		Name string `json:"name"`
		// Unit is optional. If present, it is pre-pended to `bu` member of SenML.
		Unit string `json:"unit"`

		// Entries in the DB
		Entries []Entry `json:"entries"`

		Created string `json:"created"`
		Updated string `json:"updated"`

		Metadata map[string]interface{} `json:"metadata"`
	}
)
