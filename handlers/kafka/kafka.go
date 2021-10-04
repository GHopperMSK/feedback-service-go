package handlers

import (
	"context"
	"encoding/json"
	repository "feedback-service-go/repositories"
	"fmt"
	"log"
	"os"

	kafka "github.com/segmentio/kafka-go"
)

const (
	topic         = "test"
	brokerAddress = "kafka:9092"
)

type KafkaRequest struct {
	Action  string          `json:"action"`
	Version string          `json:"version"`
	Payload json.RawMessage `json:"payload"`
}

func Consume(ctx context.Context, repo repository.Repository) {
	// create a new logger that outputs to stdout
	// and has the `kafka reader` prefix
	l := log.New(os.Stdout, "kafka reader: ", 0)
	// initialize a new reader with the brokers and topic
	// the groupID identifies the consumer and prevents
	// it from receiving duplicate messages
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddress},
		Topic:   topic,
		GroupID: "feedback-service-group",
		// assign the logger to the reader
		Logger: l,
	})
	for {
		// the `ReadMessage` method blocks until we receive the next event
		rawMsg, err := r.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}

		var inputRequest KafkaRequest
		err = json.Unmarshal(rawMsg.Value, &inputRequest)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("sucessfully got from Kafka:", string(rawMsg.Value))

		switch inputRequest.Action {
		case "create-action":
			// TODO: check for inputRequest.Version
			go CreateFeedback(inputRequest.Payload, repo)
		case "update-action":
			// TODO: check for inputRequest.Version
			go UpdateFeedback(inputRequest.Payload, repo)
		case "delete-action":
			// TODO: check for inputRequest.Version
			go DeleteFeedback(inputRequest.Payload, repo)
		case "delete-offer-action":
			// TODO: check for inputRequest.Version
			go DeleteOffer(inputRequest.Payload, repo)
		default:
			fmt.Println("got unknown action:", inputRequest.Action)
		}
	}
}

func CreateFeedback(payload json.RawMessage, repo repository.Repository) {
	var request repository.CreateRequest
	err := json.Unmarshal(payload, &request)
	if err != nil {
		panic(err.Error())
	}

	repo.Create(&request)
}

func UpdateFeedback(payload json.RawMessage, repo repository.Repository) {
	var request repository.UpdateRequest
	err := json.Unmarshal(payload, &request)
	if err != nil {
		panic(err.Error())
	}

	repo.Update(&request)
}

func DeleteFeedback(payload json.RawMessage, repo repository.Repository) {
	var request repository.DeleteRequest
	err := json.Unmarshal(payload, &request)
	if err != nil {
		panic(err.Error())
	}

	repo.Delete(&request)
}

func DeleteOffer(payload json.RawMessage, repo repository.Repository) {
	var request repository.DeleteOfferRequest
	err := json.Unmarshal(payload, &request)
	if err != nil {
		panic(err.Error())
	}

	repo.DeleteOffer(&request)
}

func ChangeTradeStatus(payload json.RawMessage, repo repository.Repository) {
	var request repository.ChangeTradeStatusRequest
	err := json.Unmarshal(payload, &request)
	if err != nil {
		panic(err.Error())
	}

	repo.ChangeTradeStatus(&request)
}
