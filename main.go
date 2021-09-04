package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	fbhandler "feedback-service-go/handlers"
	repository "feedback-service-go/repositories"
	mysql "feedback-service-go/repositories/mysql"

	kafka "github.com/segmentio/kafka-go"
)

const (
	topic         = "test"
	brokerAddress = "kafka:9092"
)

func main() {
	log.Println("Start server")

	repository, err := mysql.New()
	if err != nil {
		panic(err.Error())
	}
	log.Println("Successfully connected to the storage")

	ctx := context.Background()
	go consume(ctx, repository)

	feedbackHandler := fbhandler.New(repository)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/feedback/{id}", feedbackHandler.GetFeedback).Methods("GET")
	router.HandleFunc("/feedbacks", feedbackHandler.GetFeedbacksByFilter).Methods("GET")
	router.HandleFunc("/feedback", feedbackHandler.CreateFeedback).Methods("POST")
	router.HandleFunc("/feedback/{id}", feedbackHandler.UpdateFeedback).Methods("PATCH")
	router.HandleFunc("/feedback/{id}", feedbackHandler.DeleteFeedback).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func consume(ctx context.Context, repo repository.Repository) {
	// create a new logger that outputs to stdout
	// and has the `kafka reader` prefix
	l := log.New(os.Stdout, "kafka reader: ", 0)
	// initialize a new reader with the brokers and topic
	// the groupID identifies the consumer and prevents
	// it from receiving duplicate messages
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddress},
		Topic:   topic,
		// GroupID: "feedback-group",
		// assign the logger to the reader
		Logger: l,
	})
	for {
		// the `ReadMessage` method blocks until we receive the next event
		rawMsg, err := r.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}

		var inputFeedback repository.KafkaFeedback
		err = json.Unmarshal(rawMsg.Value, &inputFeedback)
		if err != nil {
			panic(err.Error())
		}

		// TODO: check if inputFeedback.Version valid

		parentId := 0
		if inputFeedback.ParentId.Valid {
			parentId = int(inputFeedback.ParentId.Int64)
		}
		request := repository.FeedbackRequest{
			ParentId:   parentId,
			SenderId:   inputFeedback.SenderId,
			ReceiverId: inputFeedback.ReceiverId,
			TradeId:    inputFeedback.TradeId,
			Message:    inputFeedback.Message,
			Type:       inputFeedback.Type,
		}

		repo.Create(&request)

		// after receiving the message, log its value
		fmt.Println("sucessfully got:", string(rawMsg.Value))
	}
}
