package main

import (
	"errors"
	"io"
	"log"
	"mime"
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

	r := mux.NewRouter()
	r.HandleFunc("/healthz", healthz).Methods("GET")
	r.HandleFunc("/readyz", readyz(isReady)).Methods("GET")
	// Serve index page on all unhandled routes
	r.HandleFunc("/", handleIndexRedirect).Methods("GET")
	r.PathPrefix("/").HandlerFunc(serveFiles).Methods("GET")
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

func handleIndexRedirect(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

func serveFiles(w http.ResponseWriter, r *http.Request) {
	cm := GetConfigMap()

	if cm == nil {
		http.ServeFile(w, r, filepath.Join("./static", r.URL.Path))
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	underscoredURLPath := strings.ReplaceAll(path, "/", "__")

	data, exists := cm.Data[underscoredURLPath]

	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(underscoredURLPath)))

	if exists {
		io.WriteString(w, data)
		return
	}

	binaryData, exists := cm.BinaryData[underscoredURLPath]

	if exists {
		w.Write(binaryData)
		return
	}

	http.ServeFile(w, r, filepath.Join("./static", r.URL.Path))
}
