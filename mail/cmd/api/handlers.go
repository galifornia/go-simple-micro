package main

import "net/http"

type JSONPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func (app *Config) sendMail(w http.ResponseWriter, r *http.Request) {

}
