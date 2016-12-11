/**
 * Copyright (c) Mainflux
 *
 * Mainflux server is licensed under an Apache license, version 2.0.
 * All rights not explicitly granted in the Apache license, version 2.0 are reserved.
 * See the included LICENSE file for more details.
 */

package api_test

import (
	"log"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/mainflux/mainflux-core/api"
	mfdb "github.com/mainflux/mainflux-core/db"

	"github.com/ory-am/dockertest"
	"gopkg.in/mgo.v2"
)

var ts *httptest.Server

func TestMain(m *testing.M) {
	var (
		db  *mgo.Session
		err error
	)

	c, err := dockertest.ConnectToMongoDB(15, time.Millisecond*500, func(url string) bool {
		db, err = mgo.Dial(url)
		if err != nil {
			return false
		}

		if err = db.Ping(); err != nil {
			return false
		}

		mfdb.SetMainSession(db)
		mfdb.SetMainDb("mainflux_test")
		return true
	})

	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	// Start the HTTP server
	ts = httptest.NewServer(api.HTTPServer())
	defer ts.Close()

	// Run tests
	result := m.Run()

	// Close database connection.
	db.Close()

	// Clean up image.
	c.KillRemove()

	// Exit tests.
	os.Exit(result)
}
