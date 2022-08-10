package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/rpc"

	"github.com/galifornia/go-micro-broker/event"
)

type RequestPayload struct {
	Action string        `json:"action"`
	Auth   AuthPayload   `json:"auth,omitempty"`
	Logger LoggerPayload `json:"logger,omitempty"`
	Mailer MailPayload   `json:"mail,omitempty"`
}

type MailPayload struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Message string `json:"message"`
	Subject string `json:"subject"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoggerPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker!!",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) handleSubmission(w http.ResponseWriter, r *http.Request) {
	var payload RequestPayload

	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errorJSON(w, err)
	}

	switch payload.Action {
	case "auth":
		app.authenticate(w, payload.Auth)
	case "logger":
		app.logItemViaRPC(w, payload.Logger)
		// app.logEventViaRabbit(w, payload.Logger)
		// app.logEntry(w, payload.Logger)
	case "mail":
		app.sendMail(w, payload.Mailer)
	default:
		app.errorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) sendMail(w http.ResponseWriter, entry MailPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	mailServiceURL := "http://mail-service/send"

	request, err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("something went trying to send the email through the Mail service"), response.StatusCode)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "mail sent succesfully to " + entry.To

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logEntry(w http.ResponseWriter, entry LoggerPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t") // In production shoulld be Marshal wo Indent
	logServiceURL := "http://logger-service/log"       // !FIXME: read config from env

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, err)
		return
	}
	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	jsonData, _ := json.MarshalIndent(a, "", "\t")
	req, err := http.NewRequest("POST", "http://auth-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"))
		return
	} else if res.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("something went wrong with Auth service"))
		return
	}

	var response jsonResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if response.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = response.Data

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logEventViaRabbit(w http.ResponseWriter, l LoggerPayload) {
	err := app.pushToQueue(l)
	if err != nil {
		log.Println(err)
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged via RabbitMQ"

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) pushToQueue(payload LoggerPayload) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}

	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logItemViaRPC(w http.ResponseWriter, p LoggerPayload) error {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.errorJSON(w, err)
		return err
	}

	rpcPayload := RPCPayload{
		Name: p.Name,
		Data: p.Data,
	}

	result := ""
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.errorJSON(w, err)
		return err
	}

	payload := jsonResponse{
		Error:   false,
		Message: result,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
	return nil
}
