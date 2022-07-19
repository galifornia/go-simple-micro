package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/galifornia/go-micro-auth/data"
)

const PORT = "80"

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Starting authentication service")

	// TODO: connect to DB

	app := Config{}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", PORT),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()

	if err != nil {
		log.Panic(err)
	}
}
