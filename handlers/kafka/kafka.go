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
		default:
			fmt.Println("got unknown action:", inputRequest.Action)
		}
	}
}

func CreateFeedback(payload json.RawMessage, repo repository.Repository) {
	var inputData CreateRequest
	err := json.Unmarshal(payload, &inputData)
	if err != nil {
		panic(err.Error())
	}

	request := repository.FeedbackRequest{
		ParentId:   inputData.ParentId,
		SenderId:   inputData.SenderId,
		ReceiverId: inputData.ReceiverId,
		TradeId:    inputData.TradeId,
		Message:    inputData.Message,
		Type:       inputData.Type,
		CreatedAt:  inputData.CreatedAt,
	}

	if len(request.Validate()) == 0 {
		repo.Create(&request)
	} else {
		log.Println("request validation error")
	}
}

func UpdateFeedback(payload json.RawMessage, repo repository.Repository) {
	var inputData UpdateRequest
	err := json.Unmarshal(payload, &inputData)
	if err != nil {
		panic(err.Error())
	}

	request := repository.FeedbackRequest{
		ParentId:   inputData.ParentId,
		SenderId:   inputData.SenderId,
		ReceiverId: inputData.ReceiverId,
		TradeId:    inputData.TradeId,
		Message:    inputData.Message,
		Type:       inputData.Type,
		CreatedAt:  inputData.CreatedAt,
	}

	if len(request.Validate()) == 0 {
		repo.Update(inputData.Id, &request)
	} else {
		log.Println("request validation error")
	}
}

func DeleteFeedback(payload json.RawMessage, repo repository.Repository) {
	var inputData DeleteRequest
	err := json.Unmarshal(payload, &inputData)
	if err != nil {
		panic(err.Error())
	}

	repo.Delete(inputData.Id)
}

type KafkaRequest struct {
	Action  string
	Version string
	Payload json.RawMessage
}

type CreateRequest struct {
	ParentId   int `json:"parent_id"`
	SenderId   int `json:"sender_id"`
	ReceiverId int `json:"receiver_id"`
	TradeId    int `json:"trade_id"`
	Message    string
	Type       string
	CreatedAt  string `json:"created_at"`
}

type UpdateRequest struct {
	Id         int
	ParentId   int `json:"parent_id"`
	SenderId   int `json:"sender_id"`
	ReceiverId int `json:"receiver_id"`
	TradeId    int `json:"trade_id"`
	Message    string
	Type       string
	CreatedAt  string `json:"created_at"`
}

type DeleteRequest struct {
	Id int
}
