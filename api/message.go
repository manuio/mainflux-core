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
	"strconv"
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
func writeMessage(nm NatsMsg) error {

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	var s senml.SenML
	var err error
	if s, err = senml.Decode(nm.Payload, senml.JSON); err != nil {
		return err
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
			return err
		}
		if err := json.Unmarshal(b, &m); err != nil {
			log.Print(err)
			return err
		}

		// Fill-in Mainflux stuff
		m.Channel = nm.Channel
		m.Publisher = nm.Publisher
		m.Protocol = nm.Protocol
		m.Timestamp = t

		// Insert message in DB
		if err := Db.C("messages").Insert(m); err != nil {
			log.Print(err)
			return err
		}
	}

	fmt.Println("Msg written")
	return nil
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

	// check if channel exist
	if err = Db.C("channels").Find(bson.M{"id": cid}).One(nil); err != nil {
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "not found", "id": "` + cid + `"}`
		io.WriteString(w, str)
		return
	}

	// Publisher ID header
	hdr := r.Header.Get("Client-ID")

	// Publish message on MQTT via NATS
	m := NatsMsg{}
	m.Channel = cid
	m.Publisher = hdr
	m.Protocol = "http"
	m.Payload = data

	b, err := json.Marshal(m)
	if err != nil {
		log.Print(err)
	}
	NatsConn.Publish("mainflux/core/out", b)

	// Write the message in DB
	if err := writeMessage(m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		str := `{"response": "` + err.Error() + `"}`
		io.WriteString(w, str)
		return
	}

	// Send back response to HTTP client
	// We have accepted the request and published it over MQTT,
	// but we do not know if it will be executed or not (MQTT is not req-reply protocol)
	w.WriteHeader(http.StatusAccepted)
	str := `{"response": "message sent"}`
	io.WriteString(w, str)
}

// getMessage function
func getMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	cid := bone.GetValue(r, "channel_id")

	if err := Db.C("channels").Find(bson.M{"id": cid}).One(nil); err != nil {
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "not found", "id": "` + cid + `"}`
		io.WriteString(w, str)
		return
	}

	// Get fileter values from parameters:
	// - start_time = messages from this moment. UNIX time format.
	// - end_time = messages to this moment. UNIX time format.
	var st float64
	var et float64
	var err error
	var s string
	s = r.URL.Query().Get("start_time")
	if len(s) == 0 {
		st = 0
	} else {
		st, err = strconv.ParseFloat(s, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			str := `{"response": "wrong start_time format"}`
			io.WriteString(w, str)
			return
		}
	}
	s = r.URL.Query().Get("end_time")
	if len(s) == 0 {
		et = float64(time.Now().Unix())
	} else {
		et, err = strconv.ParseFloat(s, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			str := `{"response": "wrong end_time format"}`
			io.WriteString(w, str)
			return
		}
	}

	results := []models.Message{}
	if err := Db.C("messages").Find(bson.M{"channel": cid, "time": bson.M{"$gt": st, "$lt": et}}).
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
