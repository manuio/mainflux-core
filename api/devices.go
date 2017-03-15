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
	"log"
	"time"
	"fmt"

	"github.com/mainflux/mainflux-core/db"
	"github.com/mainflux/mainflux-core/models"

	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2/bson"

	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-zoo/bone"

)

/** == Functions == */

// createDevice function
func createDevice(w http.ResponseWriter, r *http.Request) {
	// Set up defaults and pick up new values from user-provided JSON
	d := models.Device{}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	if len(data) > 0 {
		if err, str := validateDeviceSchema(data); err {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, str)
			return
		}

		if err := json.Unmarshal(data, &d); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			str := `{"response": "cannot decode body"}`
			io.WriteString(w, str)
			return
		}
	}

	// Creating UUID Version 4
	uuid := uuid.NewV4()
	d.ID = uuid.String()

	// Timestamp
	t := time.Now().UTC().Format(time.RFC3339)
	d.Created, d.Updated = t, t

	// Init MongoDB
	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	// Insert Device
	if err := Db.C("devices").Insert(d); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		str := `{"response": "cannot create device"}`
		io.WriteString(w, str)
		return
	}

	// Publish on NATS
	hdr := r.Header.Get("Authorization")
	msg := `{"type": "device", "id":"` + d.ID + `", "owner": "` + hdr + `"}`
	NatsConn.Publish("core-auth", []byte(msg))

	// Send RSP
	w.Header().Set("Location", fmt.Sprintf("/devices/%s", d.ID))
	w.WriteHeader(http.StatusCreated)
}

// getDevices function
func getDevices(w http.ResponseWriter, r *http.Request) {
	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	results := []models.Device{}
	if err := Db.C("devices").Find(nil).All(&results); err != nil {
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

// getDevice function
func getDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	id := bone.GetValue(r, "device_id")

	result := models.Device{}
	err := Db.C("devices").Find(bson.M{"id": id}).One(&result)
	if err != nil {
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

// updateDevice function
func updateDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

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

	if err, str := validateDeviceSchema(data); err {
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

	// Init MongoDB
	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	// Device id
	id := bone.GetValue(r, "device_id")

	colQuerier := bson.M{"id": id}
	change := bson.M{"$set": body}
	if err := Db.C("devices").Update(colQuerier, change); err != nil {
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

// deleteDevice function
func deleteDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	did := bone.GetValue(r, "device_id")

	// Get Device
	d := models.Device{}
	err := Db.C("devices").Find(bson.M{"id": did}).One(&d)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "not found", "id": "` + did + `"}`
		if err != nil {
			log.Print(err)
		}
		io.WriteString(w, str)
		return
	}

	// Remove this device from all the channels it was plugged into
	for _, cid := range d.Channels {
		// Remove did from the Channels's `Devices` registry
		t := time.Now().UTC().Format(time.RFC3339)
		err := Db.C("channels").Update(bson.M{"id": cid},
			bson.M{"$pull": bson.M{"devices": did}, "$set": bson.M{"updated": t}})
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusNotFound)
			str := `{"response": "cannot unplug channel ` + cid + ` for device ` + did + `"}`
			io.WriteString(w, str)
			return
		}
	}

	// Delete device
	if err := Db.C("devices").Remove(bson.M{"id": did}); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "not deleted", "id": "` + did + `"}`
		io.WriteString(w, str)
		return
	}

	w.WriteHeader(http.StatusOK)
	str := `{"response": "deleted", "id": "` + did + `"}`
	io.WriteString(w, str)
}

// plugDevice function
// Plugs given list of channle into device - i.e. creates a
// connection between device and list of devices provided
func plugDevice(w http.ResponseWriter, r *http.Request) {
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

	did := bone.GetValue(r, "device_id")

	var channels []string
	if err := json.Unmarshal(data, &channels); err != nil {
		panic(err)
	}

	for _, cid := range channels {
		// Timestamp
		t := time.Now().UTC().Format(time.RFC3339)
		// Append channelID to the Device's `Channels` registry
		if err := Db.C("channels").Update(bson.M{"id": cid},
			bson.M{"$addToSet": bson.M{"devices": did},
				"$set": bson.M{"updated": t}}); err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusNotFound)
			str := `{"response": "cannot plug channel ` + cid + ` into device ` + did + `"}`
			io.WriteString(w, str)
			return
		}

	}

	/** Add channel list to device's Channels[] */
	colQuerier := bson.M{"id": did}
	// Timestamp
	t := time.Now().UTC().Format(time.RFC3339)
	// Append entry to exiting array
	change := bson.M{"$addToSet": bson.M{"channels": bson.M{"$each": channels}}, "$set": bson.M{"updated": t}}
	if err := Db.C("devices").Update(colQuerier, change); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "cannot plug channels into device ` + did + `"}`
		io.WriteString(w, str)
		return
	}

	w.WriteHeader(http.StatusOK)
	str := `{"response": "plugged", "id": "` + did + `"}`
	io.WriteString(w, str)
}

// unplugDevice function
// Unlugs given device from a list of channels - i.e. removes
// connection between device and list of channels provided
func unplugDevice(w http.ResponseWriter, r *http.Request) {
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

	did := bone.GetValue(r, "device_id")

	var channels []string
	if err := json.Unmarshal(data, &channels); err != nil {
		panic(err)
	}

	for _, cid := range channels {
		// Remove cid from the Device's `Channels` registry
		t := time.Now().UTC().Format(time.RFC3339)
		err := Db.C("channels").Update(bson.M{"id": cid},
			bson.M{"$pull": bson.M{"devices": did}, "$set": bson.M{"updated": t}})
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusNotFound)
			str := `{"response": "cannot unplug device ` + did + ` from channel ` + cid + `"}`
			io.WriteString(w, str)
			return
		}

	}

	/** Remove channel list from devices's Channels[] */
	colQuerier := bson.M{"id": did}
	// Timestamp
	t := time.Now().UTC().Format(time.RFC3339)
	// Remove entry from exiting array
	change := bson.M{"$pull": bson.M{"channels": bson.M{"$in": channels}}, "$set": bson.M{"updated": t}}
	if err := Db.C("devices").Update(colQuerier, change); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		str := `{"response": "not unplugged", "id": "` + did + `"}`
		io.WriteString(w, str)
		return
	}

	w.WriteHeader(http.StatusOK)
	str := `{"response": "uplugged", "id": "` + did + `"}`
	io.WriteString(w, str)
}
