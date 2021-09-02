package main

import (
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

// Router sets up the routes
func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/healthz", healthz).Methods("GET")
	r.HandleFunc("/readyz", readyz).Methods("GET")
	// Serve index page on all unhandled routes
	r.HandleFunc("/", handleIndexRedirect).Methods("GET")
	r.PathPrefix("/").HandlerFunc(serveFiles).Methods("GET")

	return r
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// readyz is a readiness probe.
func readyz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func handleIndexRedirect(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

func serveFiles(w http.ResponseWriter, r *http.Request) {
	cm := GetConfigMap()

	if cm == nil {
		// Disable directory listing
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
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
