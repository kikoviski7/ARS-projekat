package handlers

import (
	"encoding/json"
	"net/http"
	"projekat/model"
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

// POST /groups/{name}/{version}
func (c ConfigGroupHandler) PostGroup(w http.ResponseWriter, r *http.Request) {

	name := mux.Vars(r)["name"]

	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var configs []model.Config

	err = json.NewDecoder(r.Body).Decode(&configs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = c.service.PostGroup(name, versionInt, configs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GET /groups
func (c ConfigGroupHandler) GetAllGroups(w http.ResponseWriter, r *http.Request) {

	groups, err := c.service.GetAllGroups()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(groups)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// GET /groups/{name}/{version}
func (c ConfigGroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {

	name := mux.Vars(r)["name"]

	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	group, err := c.service.GetGroup(name, versionInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(group)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// DELETE /groups/{name}/{version}
func (c ConfigGroupHandler) DeleteGroupByVersion(w http.ResponseWriter, r *http.Request) {

	name := mux.Vars(r)["name"]

	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.service.DeleteGroupByVersion(name, versionInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /groups/{groupName}/{groupVersion}/configs/{configName}/{configVersion}
func (c ConfigGroupHandler) DeleteConfigByVersion(w http.ResponseWriter, r *http.Request) {

	groupName := mux.Vars(r)["groupName"]

	groupVersion := mux.Vars(r)["groupVersion"]
	groupVersionInt, err := strconv.Atoi(groupVersion)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	configName := mux.Vars(r)["configName"]

	configVersion := mux.Vars(r)["configVersion"]
	configVersionInt, err := strconv.Atoi(configVersion)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.service.DeleteConfigByVersion(
		groupName,
		groupVersionInt,
		configName,
		configVersionInt,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PUT /groups/{name}/{version}
func (c ConfigGroupHandler) PutGroup(w http.ResponseWriter, r *http.Request) {

	groupName := mux.Vars(r)["name"]

	version := mux.Vars(r)["version"]
	groupVersion, err := strconv.Atoi(version)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var config model.Config

	err = json.NewDecoder(r.Body).Decode(&config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = c.service.PutGroup(config, groupName, groupVersion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GET /configsGroup/{name}/{version}/search
func (c ConfigGroupHandler) GetConfigsByLabels(w http.ResponseWriter, r *http.Request) {

	name := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse labels from query parameters (e.g., ?environment=production&team=backend)
	labels := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			labels[key] = values[0]
		}
	}

	if len(labels) == 0 {
		http.Error(w, "no labels provided in query parameters", http.StatusBadRequest)
		return
	}

	configs, err := c.service.GetConfigsByLabels(name, versionInt, labels)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(configs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}
