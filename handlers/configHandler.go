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

type ConfigHandler struct {
	service services.ConfigService
	tracer  trace.Tracer
}

func NewConfigHandler(service services.ConfigService) ConfigHandler {
	return ConfigHandler{
		service: service,
		tracer:  otel.Tracer("config-handler"),
	}
}

// POST /configs/{name}/{version}
func (c ConfigHandler) Post(w http.ResponseWriter, r *http.Request) {

	ctx, span := c.tracer.Start(r.Context(), "ConfigHandler.Post")
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

	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decode request body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = c.service.Post(ctx, name, versionInt, params, idempotencyKey)
	if err != nil {
		if errors.Is(err, model.ErrConfigAlreadyExists) {
			http.Error(w, "config already exists", http.StatusConflict)
			return
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to post config")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "config posted successfully")

	w.WriteHeader(http.StatusCreated)

}

func (c ConfigHandler) GetAll(w http.ResponseWriter, r *http.Request) {

	ctx, span := c.tracer.Start(r.Context(), "ConfigHandler.GetAll")
	defer span.End()

	configs, err := c.service.GetAll(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get all configurations")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(configs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal config")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	span.SetStatus(codes.Ok, "all configs retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)

}

func (c ConfigHandler) GetByName(w http.ResponseWriter, r *http.Request) {

	ctx, span := c.tracer.Start(r.Context(), "ConfigHandler.GetByName")
	defer span.End()

	name := mux.Vars(r)["name"]

	configs, err := c.service.GetByName(ctx, name)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve config by name")
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

	span.SetStatus(codes.Ok, "config retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// GET /configs/{name}/{version}
func (c ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {

	ctx, span := c.tracer.Start(r.Context(), "ConfigHandler.Get")
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
		span.SetStatus(codes.Error, "failed to convert version ascii to int")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	config, err := c.service.Get(ctx, name, versionInt)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get config")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(config)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal config")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "config retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (c ConfigHandler) DeleteByVersion(w http.ResponseWriter, r *http.Request) {

	ctx, span := c.tracer.Start(r.Context(), "ConfigHandler.DeleteByVersion")
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
		span.SetStatus(codes.Error, "failed to convert version ascii to int")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.service.DeleteByVersion(ctx, name, versionInt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete config by version")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	span.SetStatus(codes.Ok, "config deleted by version successfully")

	w.WriteHeader(http.StatusNoContent)
}

func (c ConfigHandler) Put(w http.ResponseWriter, r *http.Request) {

	ctx, span := c.tracer.Start(r.Context(), "ConfigHandler.DeleteByVersion")
	defer span.End()

	oldName := mux.Vars(r)["name"]
	oldVersion := mux.Vars(r)["version"]

	oldVersionInt, err := strconv.Atoi(oldVersion)

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

	err = c.service.Put(ctx, config, oldName, oldVersionInt, config.IdempotencyKey)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed put config")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(config)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "updated config successfully")

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}
