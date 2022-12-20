package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
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

		db.Post(&decodedMessage)
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
	u, err := url.Parse("http://admin:weatherdata@raspberryjoris.tplinkdns.com:5984/")
	if err != nil {
		panic(err)
	}

	couch, err := couchdb.NewClient(u)
	if err != nil {
		panic(err)
	}

	// use your new "dummy" database and create a document
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

	router := gin.Default()
	router.GET("/", getMainPage)
	router.GET("/messages", getMessages)
	router.GET("/messages/:sensorId", getMessagesBySensorId)

	router.Run("localhost:8080")
}

func getMainPage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"title": "Main Page", "message": "Hello World"})
}

func getMessages(c *gin.Context) {
	// c.IndentedJSON(http.StatusOK, messages)
}

func getMessagesBySensorId(c *gin.Context) {
	// sensorId := c.Param("sensorId")
	// filteredMessages := []map[string]any{}
	// for _, message := range messages {
	// 	if message["end_device_ids"].(map[string]interface{})["device_id"] == sensorId {
	// 		filteredMessages = append(filteredMessages, message)
	// 	}
	// }
	// c.IndentedJSON(http.StatusOK, filteredMessages)
}
