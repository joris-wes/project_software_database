package main

import (
	"encoding/json"
	"fmt"
	"net/url"

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
	u, err := url.Parse("http://admin:weatherdata@couchdb:5984/")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	couch, err := couchdb.NewClient(u)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	info, err := couch.Info()
	if err != nil {
		panic(err)
	}
	fmt.Println(info)

	db := couch.Use("weatherdata")

	opts := mqtt.NewClientOptions()
	opts.AddBroker("mqtt://eu1.cloud.thethings.network:1883").SetClientID("project-software-engineering")
	opts.SetUsername("project-software-engineering@ttn")
	opts.SetPassword("NNSXS.DTT4HTNBXEQDZ4QYU6SG73Q2OXCERCZ6574RVXI.CQE6IG6FYNJOO2MOFMXZVWZE4GXTCC2YXNQNFDLQL4APZMWU6ZGA")

	opts.SetOnConnectHandler(getConnectHandler())
	opts.SetDefaultPublishHandler(getMessageHandler(db))

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Block forever
	select {}

}
