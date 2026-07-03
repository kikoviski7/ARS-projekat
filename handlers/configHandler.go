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

// @Summary POST add config
// @Description Kreira konfiguraciju sa zadatim nazivom i verzijom.
// @Tags configs
// @Accept json
// @Param configParams body map[string]string false "Config parameters"
// @Param name path string true "Config name"
// @Param version path int true "Config version"
// @Param Idempotency-Key header string true "Idempotency key for idempotent requests"
// @Success 201 "Konfiguracija je kreirana"
// @Failure 409 "Konfiguracija sa tim nazivom i verzijom već postoji"
// @Failure 429 "Previše zahteva, pokušajte kasnije"
// @Failure 500 "Interna greška servera"
// @Router /configs/{name}/{version} [post]
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

// @Summary GET all configs
// @Description Dobavlja sve konfiguracije u sistemu.
// @Tags configs
// @Produce json
// @Success 200 {object} []model.Config "Sve konfiguracije"
// @Failure 429 "Previše zahteva, pokušajte kasnije"
// @Failure 500 "Interna greška servera"
// @Router /configs [get]
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

// @Summary GET config by name
// @Description Vraća sve verzije jedne konfiguracije.
// @Tags configs
// @Produce json
// @Param name path string true "Config name"
// @Success 200 {object} []model.Config "Sve verzije jedne konfiguracije"
// @Failure 404 "Config not found"
// @Failure 429 "Previše zahteva, pokušajte kasnije"
// @Failure 500 "Interna greška servera"
// @Router /configs/{name} [get]
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

// @Summary GET config by name and version
// @Description Vraća konfiguraciju sa tim {name} i tim {version}.
// @Tags configs
// @Produce json
// @Param name path string true "Config name"
// @Param version path int true "Config version"
// @Success 200 {object} model.Config "Konfiguracija"
// @Failure 404 "Config not found"
// @Failure 429 "Previše zahteva, pokušajte kasnije"
// @Failure 500 "Interna greška servera"
// @Router /configs/{name}/{version} [get]
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

// @Summary DELETE config by name and version
// @Description Briše konfiguraciju sa tim {name} i tim {version}.
// @Tags configs
// @Param name path string true "Config name"
// @Param version path int true "Config version"
// @Success 204 "Konfiguracija je obrisana"
// @Failure 404 "Config not found"
// @Failure 429 "Previše zahteva, pokušajte kasnije"
// @Failure 500 "Interna greška servera"
// @Router /configs/{name}/{version} [delete]
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

// @Summary PUT edit config
// @Description Menja konfiguraciju pod tim nazivom i verzijom
// @Tags configs
// @Accept  json
// @Param name path string true "Config name"
// @Param version path int true "Config version"
// @Param configParams body map[string]string false "Config parameters"
// @Success 200 "Konfiguracija je izmenjena"
// @Failure 404 "Config not found"
// @Failure 429 "Previše zahteva, pokušajte kasnije"
// @Failure 500 "Interna greška servera"
// @Router /configs/{name}/{version} [put]
func (c ConfigHandler) Put(w http.ResponseWriter, r *http.Request) {

	ctx, span := c.tracer.Start(r.Context(), "ConfigHandler.Put")
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

	err = c.service.Put(ctx, config, oldName, oldVersionInt, idempotencyKey)

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
