package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	var err error
	if len(os.Args) < 2 {
		err = godotenv.Load(".env")
	} else {
		err = godotenv.Load(os.Args[1])
	}
	if err != nil {
		log.Fatalf("Error occured loading environment variables: %s", err)
	}
	r := mux.NewRouter()
	var collections Collections
	ConnectDB(&collections)
	os.RemoveAll(os.Getenv("COMPILE_DIR"))
	r.HandleFunc("/compile", compileHandler)
	r.HandleFunc("/releases", releaseListHandler(&collections))
	err = http.ListenAndServe(os.Getenv("HOST")+":"+os.Getenv("PORT"), r)
	log.Println("Server connection failed")
	panic(err)
}
