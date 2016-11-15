/**
 * Copyright (c) Mainflux
 *
 * Mainflux server is licensed under an Apache license, version 2.0.
 * All rights not explicitly granted in the Apache license, version 2.0 are reserved.
 * See the included LICENSE file for more details.
 */

package models

import (
	"github.com/cisco/senml"
)

type (
	// SenMLRecord struct as implemented by Cisco (https://github.com/cisco/senml)
	// Copied here for quick documentatnion
	/**
	https://tools.ietf.org/html/draft-ietf-core-senml-04#section-11.2

	IANA will create a new registry for SenML labels.  The initial
	content of the registry is:

	+---------------+-------+------+----------+----+---------+
	|          name | label | cbor | xml type | id | note    |
	+---------------+-------+------+----------+----+---------+
	|     base name | bn    | -2   | string   | a  | rfcxxxx |
	|      base sum | bs    | -6   | double   | a  | rfcxxxx |
	|     base time | bt    | -3   | double   | a  | rfcxxxx |
	|     base unit | bu    | -4   | string   | a  | rfcxxxx |
	|    base value | bv    | -5   | double   | a  | rfcxxxx |
	|  base version | bver  | -1   | int      | a  | rfcxxxx |
	| boolean value | vb    | 4    | boolean  | a  | rfcxxxx |
	|    data value | vd    | 8    | string   | a  | rfcxxxx |
	|          name | n     | 0    | string   | a  | rfcxxxx |
	|  string value | vs    | 3    | string   | a  | rfcxxxx |
	|          time | t     | 6    | double   | a  | rfcxxxx |
	|          unit | u     | 1    | string   | a  | rfcxxxx |
	|   update time | ut    | 7    | double   | a  | rfcxxxx |
	|         value | v     | 2    | double   | a  | rfcxxxx |
	|     value sum | s     | 5    | double   | a  | rfcxxxx |
	|          link | l     | 9    | string   | a  | rfcxxxx |
	+---------------+-------+------+----------+----+---------+

	SenMLRecord struct {
		XMLName *bool `json:"_,omitempty" xml:"senml"`

		BaseName    string  `json:"bn,omitempty"  xml:"bn,attr,omitempty"`
		BaseTime    float64 `json:"bt,omitempty"  xml:"bt,attr,omitempty"`
		BaseUnit    string  `json:"bu,omitempty"  xml:"bu,attr,omitempty"`
		BaseVersion int     `json:"bver,omitempty"  xml:"bver,attr,omitempty"`

		Link string `json:"l,omitempty"  xml:"l,attr,omitempty"`

		Name       string  `json:"n,omitempty"  xml:"n,attr,omitempty"`
		Unit       string  `json:"u,omitempty"  xml:"u,attr,omitempty"`
		Time       float64 `json:"t,omitempty"  xml:"t,attr,omitempty"`
		UpdateTime float64 `json:"ut,omitempty"  xml:"ut,attr,omitempty"`

		Value       *float64 `json:"v,omitempty"  xml:"v,attr,omitempty"`
		StringValue string   `json:"vs,omitempty"  xml:"vs,attr,omitempty"`
		DataValue   string   `json:"vd,omitempty"  xml:"vd,attr,omitempty"`
		BoolValue   *bool    `json:"vb,omitempty"  xml:"vb,attr,omitempty"`

		Sum *float64 `json:"s,omitempty"  xml:"sum,,attr,omitempty"`
	}
	*/

	// ChannelEntry
	// "Raw" - i.e. unresolved - SenML message that comes from the sensor
	// These messages are recorded in this array "as-is".
	// Resolved version is kept in Ts array in Channel struct
	ChannelEntry struct {
		SenML     []senml.SenMLRecord `json:"senml"`
		Timestamp string              `json:"timestamp"`
	}

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

		Entries []ChannelEntry `json:"entries"`

		// Time Series
		// Concatenated array of resolved SenML entries,
		// separated by measurement (i.e. each measurement is one
		// member of the array)
		Ts []senml.SenMLRecord `json:"ts"`

		Created string `json:"created"`
		Updated string `json:"updated"`

		Metadata map[string]interface{} `json:"metadata"`
	}
)
