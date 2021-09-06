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

		var inputFeedback KafkaFeedback
		err = json.Unmarshal(rawMsg.Value, &inputFeedback)
		if err != nil {
			panic(err.Error())
		}

		// TODO: check if inputFeedback.Version valid

		request := repository.FeedbackRequest{
			ParentId:   inputFeedback.ParentId,
			SenderId:   inputFeedback.SenderId,
			ReceiverId: inputFeedback.ReceiverId,
			TradeId:    inputFeedback.TradeId,
			Message:    inputFeedback.Message,
			Type:       inputFeedback.Type,
			CreatedAt:  inputFeedback.CreatedAt,
		}

		if len(request.Validate()) == 0 {
			repo.Create(&request)
		} else {
			log.Println("request validation error")
		}

		// after receiving the message, log its value
		fmt.Println("sucessfully got:", string(rawMsg.Value))
	}
}

type KafkaFeedback struct {
	Version    string
	ParentId   int `json:"parent_id"`
	SenderId   int `json:"sender_id"`
	ReceiverId int `json:"receiver_id"`
	TradeId    int `json:"trade_id"`
	Message    string
	Type       string
	CreatedAt  string `json:"created_at"`
}
