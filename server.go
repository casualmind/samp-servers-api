package main

import (
	"net/http"

	"encoding/json"

	"fmt"

	"net/url"

	"strings"

	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Server stores the standard SA:MP query fields as well as an additional details type that stores
// additional details implemented by this API and modern server browsers.
// The json keys are short to cut down on network traffic.
type Server struct {
	Address    string            `json:"ip"`
	Hostname   string            `json:"hn"`
	Players    int               `json:"pc"`
	MaxPlayers int               `json:"pm"`
	Gamemode   string            `json:"gm"`
	Language   string            `json:"la"`
	Password   bool              `json:"pa"`
	Rules      map[string]string `json:"ru"`
	PlayerList []string          `json:"pl"`
}

// Validate checks the contents of a Server object to ensure all the required fields are valid.
func (server *Server) Validate() (errs []error) {
	errs = append(errs, ValidateAddress(server.Address)...)

	if len(server.Hostname) < 1 {
		errs = append(errs, fmt.Errorf("hostname is empty"))
	}

	if server.MaxPlayers == 0 {
		errs = append(errs, fmt.Errorf("maxplayers is empty"))
	}

	if len(server.Gamemode) < 1 {
		errs = append(errs, fmt.Errorf("gamemode is empty"))
	}

	return
}

// ValidateAddress validates an address field for a server and ensures it contains the correct
// combination of host:port with either "samp://" or an empty scheme.
func ValidateAddress(address string) (errs []error) {
	if len(address) < 1 {
		errs = append(errs, fmt.Errorf("address is empty"))
	}

	if !strings.Contains(address, "://") {
		address = fmt.Sprintf("samp://%s", address)
	}

	u, err := url.Parse(address)
	if err != nil {
		errs = append(errs, err)
		return
	}

	if u.User != nil {
		errs = append(errs, fmt.Errorf("address contains a user:password component"))
	}

	if u.Scheme != "samp" && u.Scheme != "" {
		errs = append(errs, fmt.Errorf("address contains invalid scheme '%s', must be either empty or 'samp://'", u.Scheme))
	}

	portStr := u.Port()

	if portStr != "" {
		port, err := strconv.Atoi(u.Port())
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid port '%s' specified", u.Port()))
			return
		}

		if port < 1024 || port > 49152 {
			errs = append(errs, fmt.Errorf("port %d falls within reserved or ephemeral range", port))
		}
	}

	return
}

// Server handles either posting a server object or requesting a server object
func (app *App) Server(w http.ResponseWriter, r *http.Request) {
	address, ok := mux.Vars(r)["address"]
	if !ok {
		logger.Fatal("no address specified in request",
			zap.String("request", r.URL.String()))
	}

	switch r.Method {
	case "GET":
		logger.Debug("getting server",
			zap.String("address", address))

		var (
			err error
		)

		errs := ValidateAddress(address)
		if errs != nil {
			WriteErrors(w, http.StatusBadRequest, errs)
			return
		}

		server, err := app.GetServer(address)
		if err != nil {
			WriteError(w, http.StatusNotFound, err)
			return
		}

		err = json.NewEncoder(w).Encode(&server)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

	case "POST":
		logger.Debug("posting server",
			zap.String("address", address))

		server := Server{}
		err := json.NewDecoder(r.Body).Decode(&server)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		errs := server.Validate()
		if errs != nil {
			WriteErrors(w, http.StatusUnprocessableEntity, errs)
		}

		err = app.UpsertServer(server)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err)
		}
	}
}

// GetServer looks up a server via the address
func (app *App) GetServer(address string) (server Server, err error) {
	return
}

// UpsertServer creates or updates a server object in the database.
func (app *App) UpsertServer(server Server) (err error) {
	return
}
