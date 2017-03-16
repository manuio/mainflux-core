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
	"reflect"

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

func validateGeneralSchema(body map[string]interface{}) (bool, string) {
	for k := range body {
		switch k {
			case "id", "updated", "created":
				str := `{"response": "invalid request: ` + k + ` is read-only"}`
				return true, str
			case "name":
				if (len(body[k].(string)) > 32) {
					str := `{"response": "max name size 32"}`
					return true, str
				}
				break
			case "description":
				if reflect.ValueOf(body[k]).Kind() != reflect.String  {
					str := `{"response": "` + k +
					       ` parameter is of type string"}`
					return true, str
				}
				if (len(body[k].(string)) > 256) {
					str := `{"response": "max description size 256"}`
					return true, str
				}
				break
			case "metadata":
				if reflect.ValueOf(body[k]).Kind() != reflect.Map  {
					str := `{"response": "parameter metadata is of type object"}`
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

func validateDeviceSchema(data []byte) (bool, string) {
	var body map[string]interface{}

	if err := json.Unmarshal(data, &body); err != nil {
		str := `{"response": "cannot decode body"}`
		return true, str
	}

	for k := range body {
		switch k {
			case "channels", "connected_at", "disconnected_at", "online":
				str := `{"response": "invalid request: ` + k + `is read-only"}`
				return true, str
			default :
				return validateGeneralSchema(body)
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
			case "devices", "visibility", "owner":
				str := `{"response": "invalid request: ` + k + ` is read-only"}`
				return true, str

			default :
				return validateGeneralSchema(body)
		}
	}

	return false, ""
}
