package main

import (
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	var err error
	if len(os.Args) < 2 {
		err = godotenv.Load(".env")
	} else {
		err = godotenv.Load(path.Join(os.Args[1], ".env"))
	}
	if err != nil {
		log.Fatalf("Error occured loading environment variables: %s", err)
	}
	r := mux.NewRouter()
	var collections Collections
	ConnectDB(&collections)
	if len(os.Args) != 2 {
		os.RemoveAll(os.Getenv("COMPILE_DIR"))
	} else {
		os.RemoveAll(path.Join(os.Args[1], os.Getenv("COMPILE_DIR")))
	}
	r.HandleFunc("/compile", compileHandler)
	r.HandleFunc("/releases", releaseListHandler(&collections))
	r.HandleFunc("/downloadedRelease", downloadedReleaseHandler(&collections))
	err = http.ListenAndServe(os.Getenv("HOST")+":"+os.Getenv("PORT"), r)
	log.Println("Server connection failed")
	panic(err)
}
