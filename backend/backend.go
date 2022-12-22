package main

import (
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
	couchdb "github.com/zemirco/couchdb"
)

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

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.GET("/messages", func(c *gin.Context) {
		params := couchdb.QueryParameters{}
		view, err := db.AllDocs(&params)
		if err != nil {
			panic(err)
		}
		c.JSON(200, view)
	})

	r.Run(":9000")
}
