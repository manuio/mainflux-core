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
)

func TestCreateDevice(t *testing.T) {
	var n [32 + 1]byte
	var d [256 + 1]byte

	cases := []struct {
		body   string
		header string
		code   int
	}{
		{"",                                 "api-key", http.StatusCreated},
		{`{"name": "test"}`,                 "api-key", http.StatusCreated},
		{`{"description": "test"}`,          "api-key", http.StatusCreated},
		{`{"metadata": {"test": "mTest"}}`,  "api-key", http.StatusCreated},
		{`{"name": "test",` +
		 `"metadata": {"m1": "test",` +
		 `"m2": "test" }}`,                  "api-key", http.StatusCreated},

		{"invalid",                          "api-key", http.StatusBadRequest},
		{`{"id": "test"}`,                   "api-key", http.StatusBadRequest},
		{`{"created": "test"}`,              "api-key", http.StatusBadRequest},
		{`{"channels": "test"}`,             "api-key", http.StatusBadRequest},
		{`{"metadata": "string"`,            "api-key", http.StatusBadRequest},
		{`{"connected_at": "0"`,             "api-key", http.StatusBadRequest},
		{`{"disconnected_at": "0"`,          "api-key", http.StatusBadRequest},
		{`{"name": "` + string(n[:]) + `"}`, "api-key", http.StatusBadRequest},
		{`{"description": "` + string(d[:]) + `"}`, "api-key", http.StatusBadRequest},
	}

	url := fmt.Sprintf("%s/devices", ts.URL)

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

			if !strings.HasPrefix(location, "/devices/") {
				t.Errorf("case %d: invalid 'Location' %s", i+1, location)
			}
		}
	}
}
