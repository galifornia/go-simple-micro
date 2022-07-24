package main

import (
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
		app.errorJSON(w, errors.New("Invalid credentials"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(payload.Password)
	if err != nil || !valid {
		app.errorJSON(w, errors.New("Invalid credentials"), http.StatusBadRequest)
		return
	}

	response := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("User %s has been logged in", user.FirstName),
		Data:    user,
	}

	app.writeJSON(w, http.StatusAccepted, response)
}
