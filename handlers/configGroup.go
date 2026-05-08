package handlers

import (
	"encoding/json"
	"net/http"
	"projekat/services"
	"strconv"

	"github.com/gorilla/mux"
)

type ConfigGroupHandler struct {
	service services.ConfigGroupService
}

func NewConfigGroupHandler(service services.ConfigGroupService) ConfigGroupHandler {
	return ConfigGroupHandler{
		service: service,
	}
}

/*
grupa{

	configs: [

		]

	}

}
*/
// POST /configsGroup/{name}/{version}
func (c ConfigGroupHandler) PostGroup(w http.ResponseWriter, r *http.Request) {

	name := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)

	var params []map[string]any

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

	err = c.service.PostGroup(name, versionInt, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (c ConfigGroupHandler) GetAllGroups(w http.ResponseWriter, r *http.Request) {
	configs, err := c.service.GetAllGroups()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(configs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content−Type", "application/json")
	w.Write(resp)

}

// GET /configs/{name}/{version}
func (c ConfigGroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	// dobavi naziv i verziju
	name := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// pozovi servis metodu
	config, err := c.service.GetGroup(name, versionInt)
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

	w.Header().Set("Content−Type", "application/json")
	w.Write(resp)
}

func (c ConfigGroupHandler) DeleteGroupByVersion(w http.ResponseWriter, r *http.Request) {
	// dobavi naziv i verziju
	name := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// pozovi servis metodu
	err = c.service.DeleteGroupByVersion(name, versionInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c ConfigGroupHandler) DeleteConfigByVersion(w http.ResponseWriter, r *http.Request) {

}

func (c ConfigGroupHandler) PutGroup(w http.ResponseWriter, r *http.Request) {

}
