package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"projekat/handlers"
	"projekat/middleware"
	"projekat/model"
	"projekat/repositories"
	"projekat/services"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// repo := repositories.NewConfigInMemRepository()
	consulRepo := repositories.NewConfigConsulRepository()
	service := services.NewConfigService(consulRepo)
	groupService := services.NewConfigGroupService(consulRepo)
	config := model.Config{
		Name:    "db_config",
		Version: 1,
		Params: map[string]string{
			"name":     "mare",
			"password": "marejetata123",
		},
	}

	service.Add(config)
	handler := handlers.NewConfigHandler(service)
	groupHandler := handlers.NewConfigGroupHandler(groupService)

	router := mux.NewRouter()

	router.HandleFunc("/configs/{name}/{version}", handler.Get).Methods("GET")
	router.HandleFunc("/configs", handler.GetAll).Methods("GET")
	router.HandleFunc("/configs/{name}/{version}", handler.Post).Methods("POST")
	router.HandleFunc("/configs/{name}/{version}", handler.DeleteByVersion).Methods("DELETE")

	router.HandleFunc("/configsGroup/{name}/{version}", groupHandler.GetGroup).Methods("GET")
	router.HandleFunc("/configsGroup", groupHandler.GetAllGroups).Methods("GET")
	router.HandleFunc("/configsGroup/{name}/{version}", groupHandler.PostGroup).Methods("POST")
	router.HandleFunc("/configsGroup/{name}/{version}", groupHandler.DeleteGroupByVersion).Methods("DELETE")
	router.HandleFunc("/configsGroup/{name}/{version}", groupHandler.PutGroup).Methods("PUT")

	router.HandleFunc("/configsGroup/{name}/{version}/search", groupHandler.GetConfigsByLabels).Methods("GET")
	router.HandleFunc("/configsGroup/{name}/{version}/search", groupHandler.DeleteConfigsByLabels).Methods("DELETE")

	rateLimiter := middleware.NewRateLimiter(100, 10)
	router.Use(rateLimiter.Middleware)

	server := &http.Server{
		Addr:    ":8000",
		Handler: router,
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
