package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"projekat/model"
	"projekat/services"
	"strconv"

	"github.com/gorilla/mux"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ConfigGroupHandler struct {
	service services.ConfigGroupService
	tracer  trace.Tracer
}

func NewConfigGroupHandler(service services.ConfigGroupService) ConfigGroupHandler {
	return ConfigGroupHandler{
		service: service,
		tracer:  otel.Tracer("config-handler"),
	}
}

// POST /groups/{name}/{version}
func (c ConfigGroupHandler) PostGroup(w http.ResponseWriter, r *http.Request) {

	ctx, span := c.tracer.Start(r.Context(), "ConfigGroupHandler.PostGroup")
	defer span.End()

	name := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.String("group.version", version),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to convert version ascii to int")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		span.RecordError(fmt.Errorf("missing Idempotency-Key header"))
		span.SetStatus(codes.Error, "missing Idempotency-Key header")
		http.Error(w, "Idempotency-Key header is required", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("idempotency.key", idempotencyKey),
	)

	var configs []model.Config

	err = json.NewDecoder(r.Body).Decode(&configs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decode request body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = c.service.PostGroup(ctx, name, versionInt, configs, idempotencyKey)
	if err != nil {
		if errors.Is(err, model.ErrGroupAlreadyExists) {
			http.Error(w, "group already exists", http.StatusConflict)
			return
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to post group")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "group posted successfully")

	w.WriteHeader(http.StatusCreated)
}

// GET /groups
func (c ConfigGroupHandler) GetAllGroups(w http.ResponseWriter, r *http.Request) {

	ctx, span := c.tracer.Start(r.Context(), "ConfigGroupHandler.GetAllGroups")
	defer span.End()

	groups, err := c.service.GetAllGroups(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get all groups")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(groups)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal config")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "all groups configs retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// GET /groups/{name}/{version}
func (c ConfigGroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {

	ctx, span := c.tracer.Start(r.Context(), "ConfigGroupHandler.GetGroup")
	defer span.End()

	name := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.String("group.version", version),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to convert version ascii to int")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	group, err := c.service.GetGroup(ctx, name, versionInt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get group")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(group)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal config")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "group retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// DELETE /groups/{name}/{version}
func (c ConfigGroupHandler) DeleteGroupByVersion(w http.ResponseWriter, r *http.Request) {

	ctx, span := c.tracer.Start(r.Context(), "ConfigGroupHandler.DeleteGroupByVersion")
	defer span.End()

	name := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.String("group.version", version),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to convert version ascii to int")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.service.DeleteGroupByVersion(ctx, name, versionInt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete group by version")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	span.SetStatus(codes.Ok, "config deleted via version successfully")

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /groups/{groupName}/{groupVersion}/configs/{configName}/{configVersion}
func (c ConfigGroupHandler) DeleteConfigByVersion(w http.ResponseWriter, r *http.Request) {

	ctx, span := c.tracer.Start(r.Context(), "ConfigGroupHandler")
	defer span.End()

	groupName := mux.Vars(r)["groupName"]
	groupVersion := mux.Vars(r)["groupVersion"]
	groupVersionInt, err := strconv.Atoi(groupVersion)

	span.SetAttributes(
		attribute.String("group.name", groupName),
		attribute.String("group.version", groupVersion),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to convert version ascii to int")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	configName := mux.Vars(r)["configName"]
	configVersion := mux.Vars(r)["configVersion"]
	configVersionInt, err := strconv.Atoi(configVersion)

	span.SetAttributes(
		attribute.String("config.name", configName),
		attribute.String("config.version", configVersion),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to convert version ascii to int")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.service.DeleteConfigByVersion(
		ctx,
		groupName,
		groupVersionInt,
		configName,
		configVersionInt,
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete config by version")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	span.SetStatus(codes.Ok, "config by version deleted successfully")

	w.WriteHeader(http.StatusNoContent)
}

// PUT /groups/{name}/{version}
func (c ConfigGroupHandler) PutGroup(w http.ResponseWriter, r *http.Request) {

	ctx, span := c.tracer.Start(r.Context(), "ConfigGroupHandler.PutGroup")
	defer span.End()

	groupName := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]
	groupVersion, err := strconv.Atoi(version)

	span.SetAttributes(
		attribute.String("group.name", groupName),
		attribute.Int("group.version", groupVersion),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to convert version ascii to int")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var config model.Config

	err = json.NewDecoder(r.Body).Decode(&config)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decode request body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = c.service.PutGroup(ctx, config, groupName, groupVersion)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to put config into group")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "config successfully put into group")

	w.WriteHeader(http.StatusOK)
}

// GET /configsGroup/{name}/{version}/search
func (c ConfigGroupHandler) GetConfigsByLabels(w http.ResponseWriter, r *http.Request) {

	ctx, span := c.tracer.Start(r.Context(), "ConfigGroupHandler.GetConfigsByLabels")
	defer span.End()

	name := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.String("group.version", version),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to convert version ascii to int")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	labels := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			labels[key] = values[0]
		}
	}

	if len(labels) == 0 {
		span.SetStatus(codes.Error, "no labels provided in query parameters")
		http.Error(w, "no labels provided in query parameters", http.StatusBadRequest)
		return
	}

	configs, err := c.service.GetConfigsByLabels(ctx, name, versionInt, labels)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get configs by labels")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(configs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal config")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "config retrieved from group by label successfully")

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// DELETE /configsGroup/{name}/{version}/search
func (c ConfigGroupHandler) DeleteConfigsByLabels(w http.ResponseWriter, r *http.Request) {

	ctx, span := c.tracer.Start(r.Context(), "ConfigGroupHandler.DeleteConfigsByLabels")
	defer span.End()

	name := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.String("group.version", version),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to convert version ascii to int")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	labels := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			labels[key] = values[0]
		}
	}

	if len(labels) == 0 {
		span.SetStatus(codes.Error, "no labels provided in query parameters")
		http.Error(w, "no labels provided in query parameters", http.StatusBadRequest)
		return
	}

	err = c.service.DeleteConfigsByLabels(ctx, name, versionInt, labels)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete configs by labels")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	span.SetStatus(codes.Ok, "configs inside the group deleted successfully")

	w.WriteHeader(http.StatusNoContent)
}
