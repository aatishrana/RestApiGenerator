package controllers

import (
	"log"
	"strconv"
	"router"
	"net/http"
	"strings"
	"encoding/json"
)

// Load forces the program to call all the init() funcs in each models file
func Load() {
	router.Get("/", Welcome)
}

func Welcome(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Welcome")
}

func isValueInList(value string, list []string) bool {
	for _, v := range list {
		if strings.ToLower(v) == strings.ToLower(value) {
			return true
		}
	}
	return false
}

func StringToUInt(ID string) uint {
	u64, err := strconv.ParseUint(ID, 10, 32)
	if err != nil {
		log.Println(err)
		return 0
	}
	wd := uint(u64)
	return wd
}
