package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	couchdb "github.com/zemirco/couchdb"
)

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

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		fmt.Println("Connected")

		client.Subscribe("$SYS/#", 0, nil)
		client.Subscribe("#", 0, nil)
	})

	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		fmt.Println("Topic:", msg.Topic())
		var decodedMessage CouchSensorMessage
		json.Unmarshal(msg.Payload(), &decodedMessage)
		transformSensor(&decodedMessage)

		_, err := db.Post(&decodedMessage)
		if err != nil {
			panic(err)
		}
	})

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Block forever
	select {}

}
