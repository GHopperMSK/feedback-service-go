package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	khandler "feedback-service-go/handlers/kafka"
	rhandler "feedback-service-go/handlers/rest"
	mysql "feedback-service-go/repositories/mysql"
)

func main() {
	log.Println("Start server")

	repository, err := mysql.New()
	if err != nil {
		panic(err.Error())
	}
	log.Println("Successfully connected to the storage")

	ctx := context.Background()
	go khandler.Consume(ctx, repository)

	restHandler := rhandler.New(repository)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/feedback/{id}", restHandler.GetFeedback).Methods("GET")
	router.HandleFunc("/feedbacks", restHandler.GetFeedbacksByFilter).Methods("GET")
	router.HandleFunc("/feedback", restHandler.CreateFeedback).Methods("POST")
	router.HandleFunc("/feedback/{id}", restHandler.UpdateFeedback).Methods("PATCH")
	router.HandleFunc("/feedback/{id}", restHandler.DeleteFeedback).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))
}
