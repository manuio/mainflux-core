/**
 * Copyright (c) Mainflux
 *
 * Mainflux server is licensed under an Apache license, version 2.0.
 * All rights not explicitly granted in the Apache license, version 2.0 are reserved.
 * See the included LICENSE file for more details.
 */

package api

import (
	"fmt"
	"log"
	"os"

	"encoding/json"
	"github.com/xeipuuv/gojsonschema"
)

/**
 * Function validates JSON schema for `device` od `channel` models
 * By convention, Schema files must be kept as:
 * - ./models/deviceSchema.json
 * - ./models/channelSchema.json
 */
func validateJSONSchema(model string, body map[string]interface{}) bool {
	pwd, _ := os.Getwd()
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + pwd +
		"/schema/" + model + "Schema.json")
	bodyLoader := gojsonschema.NewGoLoader(body)
	result, err := gojsonschema.Validate(schemaLoader, bodyLoader)
	if err != nil {
		log.Print(err.Error())
	}

	if !result.Valid() {
		fmt.Printf("The document is not valid. See errors :\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
		return false
	}

	fmt.Printf("The document is valid\n")
	return true
}

func validateDeviceSchema(data []byte) (bool, string) {
	var body map[string]interface{}

	if err := json.Unmarshal(data, &body); err != nil {
		str := `{"response": "cannot decode body"}`
		return true, str
	}

	for k := range body {
		switch k {
			case "id":
				str := `{"response": "invalid request: ` +
					   `device id is read-only"}`
				return true, str
			case "created":
				str := `{"response": "invalid request: ` +
					   `created is read-only"}`
				return true, str
			case "channels":
				str := `{"response": "invalid request: ` +
					   `channels is read-only"}`
				return true, str
			case "name":
				if (len(body[k].(string)) > 20) {
					str := `{"response": "max name size: 20"}`
					return true, str
				}
				break
			default :
				str := `{"response": "invalid request: ` + k +
					   ` is not a device parameter"}`
				return true, str
		}
	}

	return false, ""
}

func validateChannelSchema(data []byte) (bool, string) {
	var body map[string]interface{}

	if err := json.Unmarshal(data, &body); err != nil {
		str := `{"response": "cannot decode body"}`
		return true, str
	}

	for k := range body {
		switch k {
			case "id":
				str := `{"response": "invalid request: ` +
					   `device id is read-only"}`
				return true, str
			case "created":
				str := `{"response": "invalid request: ` +
					   `created is read-only"}`
				return true, str
			case "devices":
				str := `{"response": "invalid request: ` +
					   `channels is read-only"}`
				return true, str
			case "name":
				if (len(body[k].(string)) > 20) {
					str := `{"response": "max name size: 20"}`
					return true, str
				}
				break
			default :
				str := `{"response": "invalid request: ` + k +
					   ` is not a device parameter"}`
				return true, str
		}
	}

	return false, ""
}
