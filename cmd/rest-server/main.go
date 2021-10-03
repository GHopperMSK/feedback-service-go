package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	rhandler "feedback-service-go/handlers/rest"
	mysql "feedback-service-go/repositories/mysql"
)

func main() {
	log.Println("Start rest server")

	repository, err := mysql.New()
	if err != nil {
		panic(err.Error())
	}
	log.Println("REST successfully connected to the storage")

	restHandler := rhandler.New(repository)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/feedback/{id}", restHandler.GetFeedback).Methods("GET")
	router.HandleFunc("/feedbacks", restHandler.GetFeedbacksByFilter).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}
