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

	fmt.Println("User:", os.Getenv("COUCHDB_USER"))
	fmt.Println("Password:", os.Getenv("COUCHDB_PASSWORD"))
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
