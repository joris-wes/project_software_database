package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	couchdb "github.com/zemirco/couchdb"
)

type CouchSensorMessage struct {
	couchdb.Document
	Recieved_at    string `json:"recieved_at"`
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

func getMessageHandler(db couchdb.DatabaseService) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Topic: %s\n", msg.Topic())
		var decodedMessage CouchSensorMessage
		json.Unmarshal(msg.Payload(), &decodedMessage)

		_, err := db.Post(&decodedMessage)
		if err != nil {
			panic(err)
		}
	}
}

func getConnectHandler() mqtt.OnConnectHandler {
	return func(client mqtt.Client) {
		fmt.Println("Connected")

		client.Subscribe("$SYS/#", 0, nil)
		client.Subscribe("#", 0, nil)
	}
}

func main() {
	u, err := url.Parse("http://couchdb:5984/")
	if err != nil {
		panic(err)
	}

	u.User = url.UserPassword(os.Getenv("COUCHDB_USER"), os.Getenv("COUCHDB_PASSWORD"))

	couch, err := couchdb.NewClient(u)
	if err != nil {
		panic(err)
	}

	info, err := couch.Info()
	if err != nil {
		panic(err)
	}
	fmt.Println(info)

	db := couch.Use("weatherdata")

	opts := mqtt.NewClientOptions()
	opts.AddBroker(os.Getenv("MQTT_URL")).SetClientID("project-software-engineering")
	opts.SetUsername(os.Getenv("MQTT_USER"))
	opts.SetPassword(os.Getenv("MQTT_PASSWORD"))

	opts.SetOnConnectHandler(getConnectHandler())
	opts.SetDefaultPublishHandler(getMessageHandler(db))

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Block forever
	select {}

}
