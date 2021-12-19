package main

import (
	"github.com/khan1507017/s3clientApp/config"
	"github.com/khan1507017/s3clientApp/router"
	"github.com/khan1507017/s3clientApp/server"
)

func main() {

	srv := server.New()
	router.Routes(srv)
	srv.Logger.Fatal(srv.Start(":" + config.ServerPort))
}
