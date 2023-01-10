package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
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
	device_data := db.View("device-data")

	r := gin.Default()

	// Ping for health checks
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// List all devices
	r.GET("/list", func(c *gin.Context) {
		t := true // CouchDB library requires a pointer to a bool, so additional variable is needed
		view, err := device_data.Get(
			"ids",
			couchdb.QueryParameters{Group: &t},
		)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		keys := make([]string, len(view.Rows))
		for i, row := range view.Rows {
			keys[i] = row.Key.(string)
		}
		c.JSON(200, keys)
	})

	// Get all fields for a device
	r.GET("/:id/fields", func(c *gin.Context) {
		// CouchDB library does not put quotes around the key when querying
		// so we have to do it ourselves
		id := "\"" + c.Param("id") + "\"" // Id of the device
		limit := 1                        // Limit to only one result, since they are all the same
		view, err := device_data.Get(
			"fields",
			couchdb.QueryParameters{Key: &id, Limit: &limit},
		)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, view.Rows[0].Value)
	})

	// Get all data for a device and field
	r.GET("/:id/data/:kind", func(c *gin.Context) {
		id := c.Param("id")     // Id of the device
		kind := c.Param("kind") // Kind of data requested
		key := fmt.Sprintf("[\"%s\", \"%s\"]", id, kind)
		view, err := device_data.Get(
			"data",
			couchdb.QueryParameters{Key: &key},
		)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, view.Rows)
	})

	r.Run(":9000")
}
