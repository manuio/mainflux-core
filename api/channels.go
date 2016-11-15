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

	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2/bson"

	"github.com/cisco/senml"

	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-zoo/bone"
)

type (
	// ChannelWriteStatus is a type of Go chan
	// that is used to communicate request status
	ChannelWriteStatus struct {
		Nb  int
		Str string
	}
)

/** == Functions == */

// createChannel function
func createChannel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	if len(data) > 0 {
		var body map[string]interface{}
		if err := json.Unmarshal(data, &body); err != nil {
			panic(err)
		}
	}

	/**
	if validateJsonSchema("channel", body) != true {
		println("Invalid schema")
		w.WriteHeader(http.StatusBadRequest)
		str := `{"response": "invalid json schema in request"}`
		io.WriteString(w, str)
		return
	}
	**/

	// Init new Mongo session
	// and get the "channels" collection
	// from this new session
	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	c := models.Channel{}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &c); err != nil {
			panic(err)
		}
	}

	// Creating UUID Version 4
	uuid := uuid.NewV4()
	fmt.Println(uuid.String())

	c.ID = uuid.String()

	// Insert reference to DeviceID
	did := bone.GetValue(r, "device_id")
	if len(did) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		str := `{"response": "no device ID provided in request"}`
		io.WriteString(w, str)
		return
	}

	c.Device = did

	// Timestamp
	t := time.Now().UTC().Format(time.RFC3339)

	// Append channelID to the Device's `Channels` registry
	if err := Db.C("devices").Update(bson.M{"id": did},
		bson.M{"$addToSet": bson.M{"channels": c.ID},
			"$set": bson.M{"updated": t}}); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "cannot create channel for device ` + did + `"}`
		io.WriteString(w, str)
		return
	}

	// Insert Channel
	c.Created, c.Updated = t, t
	if err := Db.C("channels").Insert(c); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		str := `{"response": "cannot create channel"}`
		io.WriteString(w, str)
		return
	}

	w.WriteHeader(http.StatusOK)
	str := `{"response": "created", "id": "` + c.ID + `"}`
	io.WriteString(w, str)
}

// getChannels function
func getChannels(w http.ResponseWriter, r *http.Request) {
	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	// Get fileter values from parameters:
	// - climit = count limit, limits number of returned `channel` elements
	// - vlimit = value limit, limits number of values within the channel
	var climit, vlimit int
	var err error
	s := r.URL.Query().Get("climit")
	if len(s) == 0 {
		// Set default limit to -100
		climit = -100
	} else {
		climit, err = strconv.Atoi(s)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			str := `{"response": "wrong count limit"}`
			io.WriteString(w, str)
			return
		}
	}

	s = r.URL.Query().Get("vlimit")
	if len(s) == 0 {
		// Set default limit to -100
		vlimit = -100
	} else {
		vlimit, err = strconv.Atoi(s)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			str := `{"response": "wrong value limit"}`
			io.WriteString(w, str)
			return
		}
	}

	// Query DB
	results := []models.Channel{}
	if err := Db.C("channels").Find(nil).
		Select(bson.M{"values": bson.M{"$slice": vlimit}}).
		Sort("-_id").Limit(climit).All(&results); err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	res, err := json.Marshal(results)
	if err != nil {
		log.Print(err)
	}
	io.WriteString(w, string(res))
}

// getChannel function
func getChannel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	cid := bone.GetValue(r, "channel_id")

	var vlimit int
	var err error
	s := r.URL.Query().Get("vlimit")
	if len(s) == 0 {
		// Set default limit to -5
		vlimit = -100
	} else {
		vlimit, err = strconv.Atoi(s)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			str := `{"response": "wrong limit"}`
			io.WriteString(w, str)
			return
		}
	}

	result := models.Channel{}
	if err := Db.C("channels").Find(bson.M{"id": cid}).
		Select(bson.M{"values": bson.M{"$slice": vlimit}}).
		One(&result); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "not found", "id": "` + cid + `"}`
		io.WriteString(w, str)
		return
	}

	w.WriteHeader(http.StatusOK)
	res, err := json.Marshal(result)
	if err != nil {
		log.Print(err)
	}
	io.WriteString(w, string(res))
}

// writeChannel function
// Generic function that updates the channel value.
// Can be called via various protocols.
func writeChannel(cid string, data []byte) {

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	s := ChannelWriteStatus{}

	var m senml.SenML
	var err error
	if m, err = senml.Decode(data, senml.JSON); err != nil {
		s.Nb = http.StatusBadRequest
		s.Str = "Invalid request: SenML can not be decoded"
		fmt.Println(s.Nb, s.Str)
		return
	}

	// Add the "raw" SenML message to channel Entries
	e := models.ChannelEntry{}
	e.SenML = make([]senml.SenMLRecord, len(m.Records))
	copy(e.SenML, m.Records)
	// Timestamp
	t := time.Now().UTC().Format(time.RFC3339)
	e.Timestamp = t

	if err := Db.C("channels").Update(bson.M{"id": cid},
		bson.M{"$push": bson.M{"entries": e}, "$set": bson.M{"updated": t}}); err != nil {
		log.Print(err)
		s.Nb = http.StatusNotFound
		s.Str = "Not inserted"
		fmt.Println(s.Nb, s.Str)
		return
	}

	// Normalize (i.e. resolve) SenML
	n := senml.Normalize(m)

	/** Insert entry in DB */
	colQuerier := bson.M{"id": cid}
	// Timestamp
	t = time.Now().UTC().Format(time.RFC3339)
	// Append entry to exiting array
	change := bson.M{"$push": bson.M{"ts": bson.M{"$each": n.Records}}, "$set": bson.M{"updated": t}}
	if err := Db.C("channels").Update(colQuerier, change); err != nil {
		log.Print(err)
		s.Nb = http.StatusNotFound
		s.Str = "Not inserted"
		fmt.Println(s.Nb, s.Str)
		return
	}

	s.Nb = http.StatusOK
	s.Str = "Updated"

	println(s.Str)
}

// updateChannel function
func updateChannel(w http.ResponseWriter, r *http.Request) {
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

	did := bone.GetValue(r, "device_id")
	cid := bone.GetValue(r, "channel_id")

	// Publish the channel update.
	// This will be catched by the MQTT main client (subscribed to all channel topics)
	// and then written in the DB in the MQTT handler
	token := mqttClient.Publish("mainflux/devices/"+did+"/channels/"+cid, 0, false, string(data))
	token.Wait()

	// Send back response to HTTP client
	// We have accepted the request and published it over MQTT,
	// but we do not know if it will be executed or not (MQTT is not req-reply protocol)
	w.WriteHeader(http.StatusAccepted)
	str := `{"response": "channel update published"}`
	io.WriteString(w, str)
}

// deleteChannel function
func deleteChannel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	did := bone.GetValue(r, "device_id")
	cid := bone.GetValue(r, "channel_id")

	// Remove channelID from the Device's `Channels` registry
	t := time.Now().UTC().Format(time.RFC3339)
	err := Db.C("devices").Update(bson.M{"id": did},
		bson.M{"$pull": bson.M{"channels": cid}, "$set": bson.M{"updated": t}})
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "cannot remove channel for device ` + did + `"}`
		io.WriteString(w, str)
		return
	}

	// Delete channel
	if err := Db.C("channels").Remove(bson.M{"id": cid}); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "not deleted", "id": "` + cid + `"}`
		io.WriteString(w, str)
		return
	}

	w.WriteHeader(http.StatusOK)
	str := `{"response": "deleted", "id": "` + cid + `"}`
	io.WriteString(w, str)
}
