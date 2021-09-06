package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	fbhandler "feedback-service-go/handlers/feedback"
	khandler "feedback-service-go/handlers/kafka"
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

	feedbackHandler := fbhandler.New(repository)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/feedback/{id}", feedbackHandler.GetFeedback).Methods("GET")
	router.HandleFunc("/feedbacks", feedbackHandler.GetFeedbacksByFilter).Methods("GET")
	router.HandleFunc("/feedback", feedbackHandler.CreateFeedback).Methods("POST")
	router.HandleFunc("/feedback/{id}", feedbackHandler.UpdateFeedback).Methods("PATCH")
	router.HandleFunc("/feedback/{id}", feedbackHandler.DeleteFeedback).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))
}
