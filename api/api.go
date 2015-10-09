package api

import (
	"github.com/hectane/hectane/queue"

	"encoding/json"
	"net/http"
)

// Request methods.
const (
	get  = "GET"
	post = "POST"
)

// HTTP API for managing a mail queue.
type API struct {
	server  *http.Server
	config  *Config
	queue   *queue.Queue
	storage *queue.Storage
}

// Ensure the request is valid.
func (a *API) validRequest(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return false
	}
	if a.config.Username != "" && a.config.Password != "" {
		if username, password, ok := r.BasicAuth(); ok {
			if username != a.config.Username || password != a.config.Password {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return false
			}
		} else {
			w.Header().Set("WWW-Authenticate", "Basic realm=Hectane")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return false
		}
	}
	return true
}

// Respond with the specified error message. No error checking is done when
// writing the data since nothing could really be done about it.
func (a *API) respondWithJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(v)
}

// Create a new API instance for the specified queue.
func New(config *Config, queue *queue.Queue, storage *queue.Storage) *API {
	var (
		s = http.NewServeMux()
		a = &API{
			server: &http.Server{
				Addr:    config.Addr,
				Handler: s,
			},
			config:  config,
			queue:   queue,
			storage: storage,
		}
	)
	s.HandleFunc("/v1/send", a.send)
	s.HandleFunc("/v1/status", a.status)
	s.HandleFunc("/v1/version", a.version)
	return a
}

// Listen for new requests.
func (a *API) Listen() error {
	if a.config.TLSCert != "" && a.config.TLSKey != "" {
		return a.server.ListenAndServeTLS(a.config.TLSKey, a.config.TLSKey)
	} else {
		return a.server.ListenAndServe()
	}
}
