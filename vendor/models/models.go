package models

import (
	"net/http"
	"encoding/json"
	"strings"
)

const(
	sep = ","
)

type Response struct {
	StatusCode    uint            `json:"status_code"`
	StatusMessage string        `json:"status_message"`
	Data          interface{}    `json:"data"`
}

// Load forces the program to call all the init() funcs in each models file
func Load() {
	//router.Get("/", Welcome)
}

func Welcome(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{2000, "Welcome to xShowroom", nil})
}

func isValueInList(value string, list []string) bool {
	for _, v := range list {
		if strings.ToLower(v) == strings.ToLower(value) {
			return true
		}
	}
	return false
}