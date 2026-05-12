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
	groupService := services.NewConfigGroupService(repo)
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

	http.ListenAndServe("0.0.0.0:8000", router)
}
