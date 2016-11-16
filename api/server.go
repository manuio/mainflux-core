/**
 * Copyright (c) Mainflux
 *
 * Mainflux server is licensed under an Apache license, version 2.0.
 * All rights not explicitly granted in the Apache license, version 2.0 are reserved.
 * See the included LICENSE file for more details.
 */

package api

import (
	"net/http"
	"strconv"

	"github.com/mainflux/mainflux-core/config"

	"github.com/go-zoo/bone"
)

// HTTPServer function
func HTTPServer(cfg config.Config) {

	mux := bone.New()

	/**
	 * Routes
	 */
	// Status
	mux.Get("/status", http.HandlerFunc(getStatus))

	// Devices
	mux.Post("/devices", http.HandlerFunc(createDevice))
	mux.Get("/devices", http.HandlerFunc(getDevices))

	mux.Get("/devices/:device_id", http.HandlerFunc(getDevice))
	mux.Put("/devices/:device_id", http.HandlerFunc(updateDevice))

	mux.Delete("/devices/:device_id", http.HandlerFunc(deleteDevice))

	// Channels
	mux.Post("/channels", http.HandlerFunc(createChannel))
	mux.Get("/channels", http.HandlerFunc(getChannels))

	mux.Get("/channels/:channel_id", http.HandlerFunc(getChannel))
	mux.Put("/channels/:channel_id", http.HandlerFunc(updateChannel))

	mux.Delete("/channels/:channel_id", http.HandlerFunc(deleteChannel))

	/**
	 * Server
	 */
	http.ListenAndServe(cfg.HTTPHost+":"+strconv.Itoa(cfg.HTTPPort), mux)
}
