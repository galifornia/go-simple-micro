package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Config struct {
	Mailer Mail
}

const PORT = "80"

func main() {
	app := Config{
		Mailer: createMailer(),
	}

	log.Println("Starting Mail service on port", PORT)

	srv := &http.Server{
		Addr:    fmt.Sprintf("localhost:%s", PORT),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic("Unable to start web server")
	}
}

func createMailer() Mail {
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	m := Mail{
		Domain:     os.Getenv("MAIL_DOMAIN"),
		Host:       os.Getenv("MAIL_HOST"),
		Port:       port,
		Username:   os.Getenv("MAIL_USERNAME"),
		Password:   os.Getenv("MAIL_PASSWORD"),
		Encryption: os.Getenv("MAIL_ENCRYPTION"),
		FromName:   os.Getenv("MAIL_FROM_NAME"),
		From:       os.Getenv("MAIL_FROM"),
	}

	return m
}
