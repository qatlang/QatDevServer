package main

import (
	"log"
	"net/http"
)

func compileHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received at: ", r.URL.Path)
	switch r.Method {
	case "POST":
		{
			// TODO
		}
	default:
		{
			w.WriteHeader(http.StatusNotFound)
		}
	}
}
