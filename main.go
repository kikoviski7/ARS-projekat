package main

import (
	"net/http"
	"projekat/handlers"
	"projekat/model"
	"projekat/repositories"
	"projekat/services"

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
	service.Add(config)
	handler := handlers.NewConfigHandler(service)

	router := mux.NewRouter()

	router.HandleFunc("/configs/{name}/{version}", handler.Get).Methods("GET")
	router.HandleFunc("/configs", handler.GetAll).Methods("GET")
	router.HandleFunc("/configs/{name}/{version}", handler.Post).Methods("POST")
	router.HandleFunc("/configs/{name}/{version}", handler.DeleteByVersion).Methods("DELETE")

	http.ListenAndServe("0.0.0.0:8000", router)
}
