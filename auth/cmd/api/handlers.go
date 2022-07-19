package main

import "net/http"

func (app *Config) Auth(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the authentication service!!",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}
