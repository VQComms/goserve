package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	jsonFileName := os.Getenv("JSON_FILENAME")
	if jsonFileName == "" {
		jsonFileName = "config.json"
	}
	log.Printf("Serving configmap from /" + jsonFileName)

	r := mux.NewRouter()
	r.HandleFunc("/"+jsonFileName, serveConfig).Methods("GET")
	r.HandleFunc("/healthz", healthz)
	r.HandleFunc("/readyz", readyz(isReady))
	// Serve index page on all unhandled routes
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})
	r.PathPrefix("/").HandlerFunc(serveConfigFile).Methods("GET")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))
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
	cm := GetConfigMap()

	if cm == nil {
		http.NotFound(w, req)
		return
	}

	jsonFileName := os.Getenv("JSON_FILENAME")
	if jsonFileName == "" {
		jsonFileName = "config.json"
	}

	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, cm.Data[jsonFileName])
}

func serveConfigFile(w http.ResponseWriter, r *http.Request) {
	cm := GetConfigMap()

	if cm == nil {
		http.ServeFile(w, r, "./static"+r.URL.Path)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	underscoredURLPath := strings.ReplaceAll(path, "/", "__")

	binaryData, exists := cm.BinaryData[underscoredURLPath]

	if !exists {
		http.ServeFile(w, r, filepath.Join("./static", r.URL.Path))
		return
	}

	w.Write(binaryData)
}
