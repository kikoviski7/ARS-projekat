package handlers

import (
	"encoding/json"
	"net/http"
	"projekat/services"
	"strconv"

	"github.com/gorilla/mux"
)

type ConfigHandler struct {
	service services.ConfigService
}

func NewConfigHandler(service services.ConfigService) ConfigHandler {
	return ConfigHandler{
		service: service,
	}
}

// POST /configs/{name}/{version}
func (c ConfigHandler) Post(w http.ResponseWriter, r *http.Request) {

	name := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)

	var params map[string]string

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// vrati odgovor
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = c.service.Post(name, versionInt, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (c ConfigHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	configs, err := c.service.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(configs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)

}

func (c ConfigHandler) GetByName(w http.ResponseWriter, r *http.Request) {
	// dobavi naziv
	name := mux.Vars(r)["name"]

	// pozovi servis metodu
	configs, err := c.service.GetByName(name)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// vrati odgovor
	resp, err := json.Marshal(configs)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// GET /configs/{name}/{version}
func (c ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	// dobavi naziv i verziju
	name := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]

	versionInt, err := strconv.Atoi(version)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// pozovi servis metodu
	config, err := c.service.Get(name, versionInt)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// vrati odgovor
	resp, err := json.Marshal(config)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (c ConfigHandler) DeleteByVersion(w http.ResponseWriter, r *http.Request) {
	// dobavi naziv i verziju
	name := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// pozovi servis metodu
	err = c.service.DeleteByVerison(name, versionInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
