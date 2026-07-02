package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"projekat/handlers"
	"projekat/middleware"
	"projekat/repositories"
	"projekat/services"
	"time"

	"github.com/gorilla/mux"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// @title Config Service API
// @version 1.0
// @host localhost:8000
// @description This is a Config Service API.
func main() {

	tp, err := initTracer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize tracer: %v\n", err)
		os.Exit(1)
	}

	otel.SetTracerProvider(tp)

	defer func() {
		if err := tp.ForceFlush(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to flush tracer: %v\n", err)
		}
		if err := tp.Shutdown(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to shutdown tracer: %v\n", err)
		}
	}()

	consulRepo := repositories.NewConfigConsulRepository()

	service := services.NewConfigService(consulRepo)
	groupService := services.NewConfigGroupService(consulRepo)
	handler := handlers.NewConfigHandler(service)
	groupHandler := handlers.NewConfigGroupHandler(groupService)
	metricsCollector := handlers.NewMetrics()

	router := mux.NewRouter()
	router.Use(middleware.TracingMiddleware)

	router.HandleFunc("/configs/{name}/{version}", handler.Get).Methods("GET")
	router.HandleFunc("/configs/{name}", handler.GetByName).Methods("GET")
	router.HandleFunc("/configs", handler.GetAll).Methods("GET")
	router.HandleFunc("/configs/{name}/{version}", handler.Post).Methods("POST")
	router.HandleFunc("/configs/{name}/{version}", handler.Put).Methods("PUT")
	router.HandleFunc("/configs/{name}/{version}", handler.DeleteByVersion).Methods("DELETE")

	router.HandleFunc("/configsGroup/{name}/{version}", groupHandler.GetGroup).Methods("GET")

	router.HandleFunc("/configsGroup", groupHandler.GetAllGroups).Methods("GET")
	router.HandleFunc("/configsGroup/{name}/{version}", groupHandler.PostGroup).Methods("POST")
	router.HandleFunc("/configsGroup/{name}/{version}", groupHandler.DeleteGroupByVersion).Methods("DELETE")
	router.HandleFunc("/configsGroup/{name}/{version}", groupHandler.PutGroup).Methods("PUT")

	router.HandleFunc("/configsGroup/{name}/{version}/search", groupHandler.GetConfigsByLabels).Methods("GET")
	router.HandleFunc("/configsGroup/{name}/{version}/search", groupHandler.DeleteConfigsByLabels).Methods("DELETE")
	router.Handle("/metrics", promhttp.Handler())

	router.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	rateLimiter := middleware.NewRateLimiter(100, 10)
	router.Use(rateLimiter.Middleware)
	router.Use(corsMiddleware)

	metricsHandler := middleware.MetricsMiddleware(router, metricsCollector)

	server := &http.Server{
		Addr:    ":8000",
		Handler: metricsHandler,
	}

	go func() {
		server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(ctx)

	fmt.Println("Server stopped")
}

func initTracer() (*sdktrace.TracerProvider, error) {

	ctx := context.Background()

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("jaeger:4317"),
		otlptracegrpc.WithInsecure(),
	)

	if err != nil {
		return nil, err
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", "config-service"),
			attribute.String("service.version", "1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	return tp, nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
