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

	c := models.Channel{Visibility: "private", Owner: ""}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &c); err != nil {
			panic(err)
		}
	}

	// Creating UUID Version 4
	uuid := uuid.NewV4()
	fmt.Println(uuid.String())

	c.ID = uuid.String()

	// Timestamp
	t := time.Now().UTC().Format(time.RFC3339)

	// Insert Channel
	c.Created, c.Updated = t, t
	if err := Db.C("channels").Insert(c); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		str := `{"response": "cannot create channel"}`
		io.WriteString(w, str)
		return
	}

	// Publish on NATS
	hdr := r.Header.Get("Authorization")
	msg := `{"type": "channel", "id":"` + c.ID + `", "owner": "` + hdr + `"}`
	NatsConn.Publish("core-auth", []byte(msg))

	// Send RSP
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

	id := bone.GetValue(r, "channel_id")

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
	if err := Db.C("channels").Find(bson.M{"id": id}).
		Select(bson.M{"values": bson.M{"$slice": vlimit}}).
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

// Removes all the base items and expands records to have items that include
// what previosly in base iterms. Convets relative times to absoltue times.
// Only Base Name is omitted and added to every record alonside with the "n"
// So we can have distinct Base Name and Name for querying database.
func normalizeEntries(senml senml.SenML) []models.ChannelEntry {
	var bname string = ""
	var btime float64 = 0
	var bunit string = ""
	//var ver = 5
	var ret []models.ChannelEntry

	var totalRecords int = 0
	for _, r := range senml.Records {
		if (r.Value != nil) || (len(r.StringValue) > 0) || (r.BoolValue != nil) {
			totalRecords += 1
		}
	}

	ret = make([]models.ChannelEntry, totalRecords)
	var numRecords = 0

	for _, r := range senml.Records {
		if r.BaseTime != 0 {
			btime = r.BaseTime
		}
		if len(r.BaseUnit) > 0 {
			bunit = r.BaseUnit
		}
		if len(r.BaseName) > 0 {
			bname = r.BaseName
		} else {
			r.BaseName = bname
		}
		r.BaseTime = 0
		r.BaseUnit = ""
		r.Time = btime + r.Time
		if len(r.Unit) == 0 {
			r.Unit = bunit
		}
		//r.BaseVersion = ver

		if r.Time <= 0 {
			// convert to absolute time
			var now int64 = time.Now().UnixNano()
			var t int64 = now / 1000000000.0
			r.Time = float64(t) + r.Time
		}

		if (r.Value != nil) || (len(r.StringValue) > 0) || (r.BoolValue != nil) {
			// Copy SenMLRecord struct to ChannelEntry
			b, err := json.Marshal(r)
			if err != nil {
				log.Print(err)
			}
			if err := json.Unmarshal(b, &ret[numRecords]); err != nil {
			}

			////
			// Mainflux Stuff
			////
			ret[numRecords].Publisher = bname
			// Timestamp
			t := time.Now().UTC().Format(time.RFC3339)
			ret[numRecords].Timestamp = t

			// Go to next record
			numRecords += 1
		}
	}

	return ret
}

// writeChannel function
// Generic function that updates the channel value.
// Can be called via various protocols.
func writeChannel(id string, data []byte) {

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

	// Normalize (i.e. resolve) SenMLRecord
	e := normalizeEntries(m)

	/** Insert entry in DB */
	colQuerier := bson.M{"id": id}
	// Timestamp
	t := time.Now().UTC().Format(time.RFC3339)
	// Append entry to exiting array
	change := bson.M{"$push": bson.M{"entries": bson.M{"$each": e}}, "$set": bson.M{"updated": t}}
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

	id := bone.GetValue(r, "channel_id")

	// Publish the channel update.
	// This will be catched by the MQTT main client (subscribed to all channel topics)
	// and then written in the DB in the MQTT handler
	token := mqttClient.Publish("mainflux/channels/"+id, 0, false, string(data))
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

	cid := bone.GetValue(r, "channel_id")

	// Get channel
	c := models.Channel{}
	if err := Db.C("channels").Find(bson.M{"id": cid}).
		Select(bson.M{"entries": bson.M{"$slice": 1}}).
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
	change := bson.M{"$push": bson.M{"devices": bson.M{"$each": devices}}, "$set": bson.M{"updated": t}}
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
	change := bson.M{"$pull": bson.M{"devices": bson.M{"$each": devices}}, "$set": bson.M{"updated": t}}
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
