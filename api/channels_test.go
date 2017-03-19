/**
 * Copyright (c) Mainflux
 *
 * Mainflux server is licensed under an Apache license, version 2.0.
 * All rights not explicitly granted in the Apache license, version 2.0 are reserved.
 * See the included LICENSE file for more details.
 */

package api_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/mainflux/mainflux-core/db"
	"github.com/mainflux/mainflux-core/models"
)

func TestCreateChannel(t *testing.T) {
	cases := []struct {
		body   string
		header string
		code   int
	}{
		{"", "api-key", http.StatusCreated},
		{"invalid", "api-key", http.StatusBadRequest},
		{`{"id":"0000"}`, "api-key", http.StatusBadRequest},
		{`{"created":"0000"}`, "api-key", http.StatusBadRequest},
		{`{"devices":"0000"}`, "api-key", http.StatusBadRequest},
	}

	url := fmt.Sprintf("%s/channels", ts.URL)

	for i, c := range cases {
		b := strings.NewReader(c.body)

		req, _ := http.NewRequest("POST", url, b)
		req.Header.Set("Authorization", c.header)
		req.Header.Set("Content-Type", "application/json")

		cli := &http.Client{}
		res, err := cli.Do(req)
		defer res.Body.Close()

		if err != nil {
			t.Errorf("case %d: %s", i+1, err.Error())
		}

		if res.StatusCode != c.code {
			t.Errorf("case %d: expected status %d, got %d", i+1, c.code, res.StatusCode)
		}

		if res.StatusCode == http.StatusCreated {
			location := res.Header.Get("Location")

			if len(location) == 0 {
				t.Errorf("case %d: expected 'Location' to be set", i+1)
			}

			if !strings.HasPrefix(location, "/channels/") {
				t.Errorf("case %d: invalid 'Location' %s", i+1, location)
			}
		}
	}
}

func TestUpdateChannel(t *testing.T) {
	cases := []struct {
		body   string
		header string
		code   int
	}{
		{`{"name":"test"}`, "api-key", http.StatusOK},
		{`{"description":"test"}`, "api-key", http.StatusOK},
	}

	// Init MongoDB
	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	// Insert device in DB
	d := models.Channel{}
	d.ID = "IDtestChannel"
	Db.C("channels").Insert(d)

	url := fmt.Sprintf("%s/channels/%s", ts.URL, "IDtestChannel")

	for i, c := range cases {
		b := strings.NewReader(c.body)

		req, _ := http.NewRequest("PUT", url, b)
		req.Header.Set("Authorization", c.header)
		req.Header.Add("Content-Type", "application/json")

		cli := &http.Client{}
		res, err := cli.Do(req)
		defer res.Body.Close()

		if err != nil {
			t.Errorf("case %d: %s", i+1, err.Error())
		}

		if res.StatusCode != c.code {
			t.Errorf("case %d: expected status %d, got %d", i+1, c.code, res.StatusCode)
		}

		if res.StatusCode == http.StatusCreated {
			location := res.Header.Get("Location")

			if len(location) == 0 {
				t.Errorf("case %d: expected 'Location' to be set", i+1)
			}

			if !strings.HasPrefix(location, "/channels/") {
				t.Errorf("case %d: invalid 'Location' %s", i+1, location)
			}
		}
	}
}

func TestDeleteChannel(t *testing.T) {
	cases := []struct {
		ID   string
		header string
		code   int
	}{
		{"existentTestID",  "api-key", http.StatusOK},
		{"invalid",         "api-key", http.StatusNotFound},
	}

	// Init MongoDB
	Db := db.MgoDb{}
	Db.Init()
	defer Db.Close()

	// Insert device in DB
	d := models.Channel{}
	d.ID = cases[0].ID
	Db.C("channels").Insert(d)

	for i, c := range cases {
		url := fmt.Sprintf("%s/channels/%s", ts.URL, c.ID)

		req, err := http.NewRequest("DELETE", url, nil)
		req.Header.Set("Authorization", c.header)

		cli := &http.Client{}
		res, err := cli.Do(req)
		defer res.Body.Close()

		if err != nil {
			t.Errorf("case %d: %s", i+1, err.Error())
		}

		if res.StatusCode != c.code {
			t.Errorf("case %d: expected status %d, got %d", i+1, c.code, res.StatusCode)
		}
	}
}
