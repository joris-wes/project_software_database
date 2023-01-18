package main

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"time"

	"github.com/gin-contrib/cors"
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
	r.Use(cors.Default())

	// Ping for health checks
	r.GET("/api/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// List all devices
	r.GET("/api/list", func(c *gin.Context) {
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
	r.GET("/api/:id/fields", func(c *gin.Context) {
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

	// Raw data
	r.GET("/api/:id/data/raw", func(c *gin.Context) {
		id := "\"" + c.Param("id") + "\""
		t := true
		view, err := device_data.Get(
			"docs",
			couchdb.QueryParameters{Key: &id, IncludeDocs: &t},
		)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		docs := make([]map[string]any, len(view.Rows))
		for i, row := range view.Rows {
			docs[i] = row.Doc
		}

		c.JSON(200, docs)
	})

	// Get all data for a device and field
	r.GET("/api/:id/data/:kind/:time", func(c *gin.Context) {
		id := c.Param("id")       // Id of the device
		kind := c.Param("kind")   // Kind of data requested
		period := c.Param("time") // Period of time requested

		// Check for frontend errors
		if kind == "undefined" {
			c.JSON(200, make([]any, 0))
			return
		}
		if period == "undefined" {
			c.JSON(200, make([]any, 0))
			return
		}

		const layout = "2006-01-02T15:04:05.000000000Z" // Layout for the time in the database
		var start_date time.Time                        // Edge for the recent data
		var grouper func(time time.Time) time.Time      // Function to group data
		switch period {
		case "hour":
			// By hour shows data from the last 24 hours without grouping
			start_date = time.Now().Add(-25 * time.Hour)
			grouper = func(t time.Time) time.Time {
				return t.Truncate(10 * time.Minute)
			}
		case "day":
			// By day shows data from the last 7 days grouped by hour
			start_date = time.Now().Add(-7 * 24 * time.Hour)
			grouper = func(t time.Time) time.Time {
				return t.Truncate(3 * time.Hour)
			}
		case "month":
			// By month shows data from the last year grouped by week
			start_date = time.Now().Add(-12 * 30 * 24 * time.Hour)
			grouper = func(t time.Time) time.Time {
				return t.Truncate(3 * 24 * time.Hour)
			}
		}

		fmt.Println("START", start_date.Format(layout))
		// CouchDB requires a string, so generating one
		start_key := fmt.Sprintf("[\"%s\", \"%s\", \"%s\"]", id, kind, start_date.Format(layout))
		end_key := fmt.Sprintf("[\"%s\", \"%s\", {}]", id, kind)
		view, err := device_data.Get(
			"data",
			couchdb.QueryParameters{
				StartKey: &start_key,
				EndKey:   &end_key,
			},
		)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		if len(view.Rows) == 0 {
			c.JSON(204, gin.H{"error": "No data found"})
			return
		}

		// Filter and transform recieved data
		type Row struct {
			Value float64   `json:"value"`
			Time  time.Time `json:"time"`
		}

		values := make([]Row, 0)

		// Get initial state for grouping
		first := view.Rows[0].Value.(map[string]any)
		first_time, err := time.Parse(layout, first["time"].(string))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		runner := Row{
			Value: first["value"].(float64),
			Time:  grouper(first_time),
		}
		count := 1

		// Filter only recent data and group it by time
		for _, row := range view.Rows[1:] {
			value := row.Value.(map[string]any)
			time, err := time.Parse(layout, value["time"].(string))

			// Skip if time is invalid or data is old
			if err != nil || time.Before(start_date) {
				continue
			}

			if grouper(time) != grouper(runner.Time) {
				// Round to 2 decimal places. Yes, there is no round function in Go
				runner.Value = math.Floor(runner.Value/float64(count)*100) / 100
				values = append(values, runner)
				runner = Row{
					Value: value["value"].(float64),
					Time:  grouper(time),
				}
				count = 1

			} else {
				runner.Value += value["value"].(float64)
				count++
			}
		}

		c.JSON(200, values[1:]) // First value is noise, so it is skipped
	})

	r.Run(":9000")
}
