package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/gorilla/mux"
)

// Router sets up the routes
func Router() *mux.Router {
	isReady := &atomic.Value{}
	isReady.Store(false)
	go func() {
		if _, exists := os.LookupEnv("CONFIGMAP_NAME"); !exists {
			panic(errors.New("Please provide the CONFIGMAP_NAME environment variable"))
		}
		isReady.Store(true)
		log.Printf("Application is ready")
	}()

	r := mux.NewRouter()
	r.HandleFunc("/config.json", serveConfig).Methods("GET")
	r.HandleFunc("/healthz", healthz)
	r.HandleFunc("/readyz", readyz(isReady))
	return r
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// readyz is a readiness probe.
func readyz(isReady *atomic.Value) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if isReady == nil || !isReady.Load().(bool) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func serveConfig(w http.ResponseWriter, req *http.Request) {
	cm, err := FetchConfigMap(os.Getenv("CONFIGMAP_NAME"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	json, err := json.Marshal(cm.Data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
