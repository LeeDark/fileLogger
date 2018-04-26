package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type appError struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	HttpStatus int    `json:"status"`
	ExitCode   int    `json:"exitcode"`
}

type errorResource struct {
	Data appError `json:"data"`
}

func Home(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, "HOME")
}

func displayAppError(w http.ResponseWriter, handlerError error, message string, code int, exitCode int) {
	errObj := appError{
		Error:      "nil",
		Message:    message,
		HttpStatus: code,
		ExitCode:   exitCode,
	}

	if handlerError != nil {
		errObj.Error = handlerError.Error()
	}

	fmt.Printf("[app error]: %s\n", handlerError)

	respondWithJSON(w, code, errorResource{Data: errObj})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
