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

// createChannel function
func createChannel(w http.ResponseWriter, r *http.Request) {
	c := models.Channel{}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(data) > 0 {
		if err, str := validateChannelSchema(data); err {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, str)
			return
		}

		if err := json.Unmarshal(data, &c); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			str := `{"response": "cannot decode body"}`
			io.WriteString(w, str)
			return
		}
	}

	// Creating UUID Version 4
	c.ID = uuid.NewV4().String()

	// Timestamp
	ts := time.Now().UTC().Format(time.RFC3339)
	c.Created, c.Updated = ts, ts

	c.Owner = ""
	c.Visibility = "private"

	// Init MongoDB
	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	// Insert Channel
	if err := Db.C("channels").Insert(c); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Publish on NATS
	go func() {
		hdr := r.Header.Get("Authorization")
		msg := `{"type": "channel", "id":"` + c.ID + `", "owner": "` + hdr + `"}`
		NatsConn.Publish("mainflux/core/auth", []byte(msg))
	}()

	// Send RSP
	w.Header().Set("Location", fmt.Sprintf("/channels/%s", c.ID))
	w.WriteHeader(http.StatusCreated)
}

// getChannels function
func getChannels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	// Get fileter values from parameters:
	// - climit = count limit, limits number of returned `channel` elements
	// - vlimit = value limit, limits number of values within the channel
	var climit int
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

	// Query DB
	results := []models.Channel{}
	if err := Db.C("channels").Find(nil).
		Sort("-_id").Limit(climit).All(&results); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		str := `{"response": "` + err.Error() + `"}`
		io.WriteString(w, str)
		return
	}
	if len(results) == 0 {
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "no channel found"}`
		io.WriteString(w, str)
		return
	}

	res, err := json.Marshal(results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		str := `{"response": "` + err.Error() + `"}`
		io.WriteString(w, str)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(res))
}

// getChannel function
func getChannel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	id := bone.GetValue(r, "channel_id")

	result := models.Channel{}
	if err := Db.C("channels").Find(bson.M{"id": id}).
		One(&result); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "not found", "id": "` + id + `"}`
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

// updateChannel function
func updateChannel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate JSON schema
	if len(data) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		str := `{"response": "no data provided"}`
		io.WriteString(w, str)
		return
	}

	if err, str := validateChannelSchema(data); err {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, str)
		return
	}

	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		panic(err)
	}

	// Timestamp
	body["updated"] = time.Now().UTC().Format(time.RFC3339)

	// Channel id
	id := bone.GetValue(r, "channel_id")

	colQuerier := bson.M{"id": id}
	change := bson.M{"$set": body}
	if err := Db.C("channels").Update(colQuerier, change); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "not updated", "id": "` + id + `"}`
		io.WriteString(w, str)
		return
	}

	w.WriteHeader(http.StatusOK)
	str := `{"response": "updated", "id": "` + id + `"}`
	io.WriteString(w, str)
}

// deleteChannel function
func deleteChannel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	cid := bone.GetValue(r, "channel_id")

	// Get channel
	c := models.Channel{}
	if err := Db.C("channels").Find(bson.M{"id": cid}).
		One(&c); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "not found", "id": "` + cid + `"}`
		io.WriteString(w, str)
		return
	}

	// Loop to all devices plugged into this channel
	for _, did := range c.Devices {
		// Remove channelID from the Device's `Channels` registry
		t := time.Now().UTC().Format(time.RFC3339)
		err := Db.C("devices").Update(bson.M{"id": did},
			bson.M{"$pull": bson.M{"channels": cid}, "$set": bson.M{"updated": t}})
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusNotFound)
			str := `{"response": "cannot unplug channel for device ` + did + `"}`
			io.WriteString(w, str)
			return
		}
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

// plugChannel function
// Plugs given channel into devices - i.e. creates a
// connection between channel and list of devices provided
func plugChannel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	/**
	if len(data) > 0 {
		var body map[string]interface{}
		if err := json.Unmarshal(data, &body); err != nil {
			panic(err)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		str := `{"response": "no data provided"}`
		io.WriteString(w, str)
		return
	}
	*/

	/**
	if validateJsonSchema("channel", body) != true {
		println("Invalid schema")
		w.WriteHeader(http.StatusBadRequest)
		str := `{"response": "invalid json schema in request"}`
		io.WriteString(w, str)
		return
	}
	**/

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	cid := bone.GetValue(r, "channel_id")

	var devices []string
	if err := json.Unmarshal(data, &devices); err != nil {
		panic(err)
	}

	for _, did := range devices {
		// Timestamp
		t := time.Now().UTC().Format(time.RFC3339)
		// Append channelID to the Device's `Channels` registry
		if err := Db.C("devices").Update(bson.M{"id": did},
			bson.M{"$addToSet": bson.M{"channels": cid},
				"$set": bson.M{"updated": t}}); err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusNotFound)
			str := `{"response": "cannot plug channel into device ` + did + `"}`
			io.WriteString(w, str)
			return
		}

	}

	/** Append device list to channel's Devices[] */
	colQuerier := bson.M{"id": cid}
	// Timestamp
	t := time.Now().UTC().Format(time.RFC3339)
	// Append entry to exiting array
	change := bson.M{"$addToSet": bson.M{"devices": bson.M{"$each": devices}}, "$set": bson.M{"updated": t}}
	if err := Db.C("channels").Update(colQuerier, change); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "not inserted", "id": "` + cid + `"}`
		io.WriteString(w, str)
		return
	}

	w.WriteHeader(http.StatusOK)
	str := `{"response": "plugged", "id": "` + cid + `"}`
	io.WriteString(w, str)
}

// unplugChannel function
// Unlugs given list of devices from given channel - i.e. removes
// connection between channel and list of devices provided
func unplugChannel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	/**
	if len(data) > 0 {
		var body map[string]interface{}
		if err := json.Unmarshal(data, &body); err != nil {
			panic(err)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		str := `{"response": "no data provided"}`
		io.WriteString(w, str)
		return
	}
	*/

	/**
	if validateJsonSchema("channel", body) != true {
		println("Invalid schema")
		w.WriteHeader(http.StatusBadRequest)
		str := `{"response": "invalid json schema in request"}`
		io.WriteString(w, str)
		return
	}
	**/

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	cid := bone.GetValue(r, "channel_id")

	var devices []string
	if err := json.Unmarshal(data, &devices); err != nil {
		panic(err)
	}

	for _, did := range devices {
		// Remove cid from the Device's `Channels` registry
		t := time.Now().UTC().Format(time.RFC3339)
		err := Db.C("devices").Update(bson.M{"id": did},
			bson.M{"$pull": bson.M{"channels": cid}, "$set": bson.M{"updated": t}})
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusNotFound)
			str := `{"response": "cannot unplug channel ` + cid + ` for device ` + did + `"}`
			io.WriteString(w, str)
			return
		}

	}

	/** Remove device list from channel's Devices[] */
	colQuerier := bson.M{"id": cid}
	// Timestamp
	t := time.Now().UTC().Format(time.RFC3339)
	// Remove entry from exiting array
	change := bson.M{"$pull": bson.M{"devices": bson.M{"$in": devices}}, "$set": bson.M{"updated": t}}
	if err := Db.C("channels").Update(colQuerier, change); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "not inserted", "id": "` + cid + `"}`
		io.WriteString(w, str)
		return
	}

	w.WriteHeader(http.StatusOK)
	str := `{"response": "uplugged", "id": "` + cid + `"}`
	io.WriteString(w, str)
}
