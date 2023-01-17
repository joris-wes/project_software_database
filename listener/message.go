package main

import (
	"math"
	"strings"

	couchdb "github.com/zemirco/couchdb"
)

type CouchSensorMessage struct {
	couchdb.Document
	End_device_ids struct {
		Device_id string `json:"device_id"`
	} `json:"end_device_ids"`
	Uplink_message struct {
		Decoded_payload map[string]any `json:"decoded_payload"`
		Settings        struct {
			Data_rate struct {
				Bandwidth        int
				Spreading_factor int
			}
			Frequency string
			Timestamp int
			Time      string
		}
		Rx_metadata      []map[string]any `json:"rx_metadata"`
		Received_at      string
		Consumed_airtime string
	} `json:"uplink_message"`
}

func transformPySensor(message *CouchSensorMessage) {
	decodedPayload := message.Uplink_message.Decoded_payload

	if _, ok := decodedPayload["light"]; ok {
		decodedPayload["light"] = decodedPayload["light"].(float64) / 2.55
	}
}

func transformLhtSensor(message *CouchSensorMessage) {
	decodedPayload := message.Uplink_message.Decoded_payload

	if _, ok := decodedPayload["ILL_lx"]; ok {
		lux := decodedPayload["ILL_lx"].(float64)
		if lux >= 123 {
			// Calculate logarithm with base 1.04
			lux = math.Round(math.Log(lux) / math.Log(1.04))
			if lux > 255 {
				lux = 255
			}
		}
		decodedPayload["light"] = lux / 2.55
		delete(decodedPayload, "ILL_lx")
	}

	if _, ok := decodedPayload["Hum_SHT"]; ok {
		decodedPayload["humidity"] = decodedPayload["Hum_SHT"].(float64)
		delete(decodedPayload, "Hum_SHT")
	}

	if _, ok := decodedPayload["TempC_SHT"]; ok {
		decodedPayload["temperature_out"] = decodedPayload["TempC_SHT"].(float64)
		delete(decodedPayload, "TempC_SHT")
	}

	delete(decodedPayload, "Work_Mode")
}

func transformLhtSaxionSensor(message *CouchSensorMessage) {
	decodedPayload := message.Uplink_message.Decoded_payload

	decodedPayload["temperature"] = decodedPayload["temperature_out"].(float64)
	decodedPayload["temperature_out"] = decodedPayload["TempC_DS"].(float64)
	delete(decodedPayload, "TempC_DS")
}

func transformSensor(message *CouchSensorMessage) {
	device_id := message.End_device_ids.Device_id
	switch {
	case strings.HasPrefix(device_id, "py-") || strings.HasPrefix(device_id, "eui-"):
		transformPySensor(message)
	case device_id == "lht-saxion":
		transformLhtSensor(message)
		transformLhtSaxionSensor(message)
	case strings.HasPrefix(device_id, "lht-"):
		transformLhtSensor(message)
	}
}
