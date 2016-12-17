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
	"log"
	"time"

	"github.com/mainflux/mainflux-core/db"
	"github.com/mainflux/mainflux-core/models"

	"gopkg.in/mgo.v2/bson"

	"github.com/cisco/senml"

	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-zoo/bone"
)

// writeMessage function
// Writtes message into DB.
// Can be called via various protocols.
func writeMessage(publisher string, channel_id string, data []byte) {

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	var s senml.SenML
	var err error
	if s, err = senml.Decode(data, senml.JSON); err != nil {
		return
	}

	// Normalize (i.e. resolve) SenMLRecord
	sn := senml.Normalize(s)

	// Timestamp
	t := time.Now().UTC().Format(time.RFC3339)
	for _, r := range sn.Records {

		m := models.Message{}

		// Copy SenMLRecord struct to Message struct
		b, err := json.Marshal(r)
		if err != nil {
			log.Print(err)
			return
		}
		if err := json.Unmarshal(b, &m); err != nil {
			log.Print(err)
			return
		}

		// Fill-in Mainflux stuff
		m.Channel = channel_id
		m.Publisher = publisher
		m.Timestamp = t

		// Insert message in DB
		if err := Db.C("messages").Insert(m); err != nil {
			log.Print(err)
			return
		}
	}

	fmt.Println("Msg written")
}

// sendMessage function
func sendMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	if len(data) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		str := `{"response": "no data provided"}`
		io.WriteString(w, str)
		return
	}

	cid := bone.GetValue(r, "channel_id")

	// Publish the channel update.
	// This will be catched by the MQTT main client (subscribed to all channel topics)
	// and then written in the DB in the MQTT handler
	m := MqttMsg{}
	m.Topic = "mainflux/channels/" + cid
	m.Publisher = mainfluxCoreUUID
	m.Payload = data

	b, err := json.Marshal(m)
	if err != nil {
		log.Print(err)
	}
	token := mqttClient.Publish("mainflux/core/pub", 0, false, b)
	token.Wait()

	// Send back response to HTTP client
	// We have accepted the request and published it over MQTT,
	// but we do not know if it will be executed or not (MQTT is not req-reply protocol)
	w.WriteHeader(http.StatusAccepted)
	str := `{"response": "channel update published"}`
	io.WriteString(w, str)
}

// getMessage function
func getMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	cid := bone.GetValue(r, "channel_id")

	results := []models.Message{}
	if err := Db.C("messages").Find(bson.M{"channel": cid}).
		All(&results); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "not found", "id": "` + cid + `"}`
		io.WriteString(w, str)
		return
	}

	w.WriteHeader(http.StatusOK)
	res, err := json.Marshal(results)
	if err != nil {
		log.Print(err)
	}
	io.WriteString(w, string(res))
}
