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

	"github.com/codegangsta/negroni"
	"github.com/go-zoo/bone"
)

// HTTPServer function
func HTTPServer() http.Handler {
	mux := bone.New()

	// Status
	mux.Get("/status", http.HandlerFunc(getStatus))

	// Devices
	mux.Post("/devices", http.HandlerFunc(createDevice))
	mux.Get("/devices", http.HandlerFunc(getDevices))

	mux.Get("/devices/:device_id", http.HandlerFunc(getDevice))
	mux.Put("/devices/:device_id", http.HandlerFunc(updateDevice))

	mux.Delete("/devices/:device_id", http.HandlerFunc(deleteDevice))

	mux.Post("/devices/:device_id/plug", http.HandlerFunc(plugDevice))
	mux.Post("/devices/:device_id/unplug", http.HandlerFunc(unplugDevice))

	// Channels
	mux.Post("/channels", http.HandlerFunc(createChannel))
	mux.Get("/channels", http.HandlerFunc(getChannels))

	mux.Get("/channels/:channel_id", http.HandlerFunc(getChannel))
	mux.Put("/channels/:channel_id", http.HandlerFunc(updateChannel))

	mux.Delete("/channels/:channel_id", http.HandlerFunc(deleteChannel))

	mux.Post("/channels/:channel_id/plug", http.HandlerFunc(plugChannel))
	mux.Post("/channels/:channel_id/unplug", http.HandlerFunc(unplugChannel))

	mux.Post("/channels/:channel_id/msg", http.HandlerFunc(sendMessage))
	mux.Get("/channels/:channel_id/msg", http.HandlerFunc(getMessage))

	n := negroni.Classic()
	n.UseHandler(mux)
	return n
}
