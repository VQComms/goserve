package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	cm, exists := os.LookupEnv("CONFIGMAP_NAME")
	if !exists {
		panic(errors.New("Please provide the CONFIGMAP_NAME environment variable"))
	}

	go InitializeInformer(cm)

	log.Print("Starting the service...")
	router := Router()
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, loggedRouter))
}
