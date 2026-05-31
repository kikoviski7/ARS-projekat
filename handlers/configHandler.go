package handlers

import (
	"encoding/json"
	"net/http"
	"projekat/services"
	"strconv"

	"github.com/gorilla/mux"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type ConfigHandler struct {
	service services.ConfigService
	tracer trace.Tracer
}

func NewConfigHandler(service services.ConfigService) ConfigHandler {
	return ConfigHandler{
		service: service,
		tracer: otel.Tracer("config-handler"),
	}
}

// POST /configs/{name}/{version}
func (c ConfigHandler) Post(w http.ResponseWriter, r *http.Request) {

	ctx, span := h.tracer.Start(r.Context(), "ConfigHandler.Post")
	defer span.End()

	name := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)

	span.SetAttributes(
		attribute.String("config.name", name),
		attribute.String("config.version", version),
	)


	var params map[string]string

	if err != nil {
		span.RecordError(err)
		span.SetStatus(code.Error, "failed to convert version ascii to int")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(code.Error, "failed to decode request body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = c.service.Post(ctx, name, versionInt, params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(code.Error, "failed to post config")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(code.Ok, "config posted successfully")

	w.WriteHeader(http.StatusCreated)

}

func (c ConfigHandler) GetAll(w http.ResponseWriter, r *http.Request) {

	ctx, span := h.tracer.Start(r.Context(), "ConfigHandler.GetAll")
	defer span.End()

	configs, err := c.service.GetAll(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(code.Error, "failed to get all configurations")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(configs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(code.Error, "failed to marshal config")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	span.SetStatus(code.Ok, "all configs retrieved successfully")

	w.Header().Set("Content−Type", "application/json")
	w.Write(resp)

}

// GET /configs/{name}/{version}
func (c ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {

	ctx, span := h.tracer.Start(r.Context(), "ConfigHandler.Get")
	defer span.End()

	name := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)

	span.SetAttributes(
		attribute.String("config.name", name),
		attribute.String("config.version", version),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(code.Error, "failed to convert version ascii to int")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	config, err := c.service.Get(ctx, name, versionInt)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(code.Error, "failed to get config")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(config)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(code.Error, "failed to marshal config")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(code.Ok, "config retrieved successfully")

	w.Header().Set("Content−Type", "application/json")
	w.Write(resp)
}

func (c ConfigHandler) DeleteByVersion(w http.ResponseWriter, r *http.Request) {

	ctx, span := h.tracer.Start(r.Context(), "ConfigHandler.DeleteByVersion")
	defer span.End()

	name := mux.Vars(r)["name"]
	version := mux.Vars(r)["version"]
	versionInt, err := strconv.Atoi(version)

	span.SetAttributes(
		attributes.String("config.name", name),
		attributes.String("config.version", version),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(code.Error, "failed to convert version ascii to int")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.service.DeleteByVerison(ctx, name, versionInt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(code.Error, "failed to delete config by version")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	span.SetStatus(code.Ok, "config deleted by version successfully")

	w.WriteHeader(http.StatusNoContent)
}
