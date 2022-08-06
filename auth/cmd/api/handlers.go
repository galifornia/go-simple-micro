package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (app *Config) Auth(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validate the user
	user, err := app.Models.User.GetByEmail(payload.Email)

	if err != nil {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	valid, err := user.PasswordMatches(payload.Password)
	if err != nil || !valid {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	err = app.logRequest("authentication", fmt.Sprintf("User %s has been logged in", user.FirstName))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	response := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("User %s has been logged in", user.FirstName),
		Data:    user,
	}

	app.writeJSON(w, http.StatusAccepted, response)
}

func (app *Config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := &http.Client{}

	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}
