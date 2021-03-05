package main

import (
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
	if exists {
		go InitializeInformer(cm)
	} else {
		log.Print("No CONFIGMAP_NAME environment variable, so skipping k8s functionality!")
	}

	log.Print("Starting the service...")
	router := Router()
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, loggedRouter))
}
