package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"projekat/handlers"
	"projekat/model"
	"projekat/repositories"
	"projekat/services"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	repo := repositories.NewConfigInMemRepository()
	service := services.NewConfigService(repo)
	config := model.Config{
		Name:    "db_config",
		Id:      "1",
		Version: 1,
		Params: map[string]string{
			"name":     "mare",
			"password": "marejetata123",
		},
	}

	// context.WithDeadline(parent Context, d time.Time) (ctx Context, cancel CancelFunc)
	// context.WithTimeout(parent Context , timeout time.Duration) (ctx Context , cancel CancelFunc)
	// context.WithCancel(parent Context) ( ctx Context , cancel CancelFunc)
	// context.WithValue(parent Context , key , val interface , key string) ctx

	service.Add(config)
	handler := handlers.NewConfigHandler(service)

	router := mux.NewRouter()

	router.HandleFunc("/configs/{name}/{version}", handler.Get).Methods("GET")
	router.HandleFunc("/configs", handler.GetAll).Methods("GET")
	router.HandleFunc("/configs/{name}/{version}", handler.Post).Methods("POST")
	router.HandleFunc("/configs/{name}/{version}", handler.DeleteByVersion).Methods("DELETE")

	server := &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: router,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(ctx)

	fmt.Println("Server stopped")
}
