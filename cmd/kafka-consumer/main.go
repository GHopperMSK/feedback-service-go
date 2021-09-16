package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	khandler "feedback-service-go/handlers/kafka"
	mysql "feedback-service-go/repositories/mysql"

	kafka "github.com/segmentio/kafka-go"
)

func main() {
	log.Println("Start kafka consumer server")

	repository, err := mysql.New()
	if err != nil {
		panic(err.Error())
	}
	log.Println("Kafka consumer server successfully connected to the storage")

	ctx := context.Background()

	// create a new logger that outputs to stdout
	// and has the `kafka reader` prefix
	l := log.New(os.Stdout, "kafka reader: ", 0)
	// initialize a new reader with the brokers and topic
	// the groupID identifies the consumer and prevents
	// it from receiving duplicate messages
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKER_ADDRESS")},
		Topic:   os.Getenv("KAFKA_TOPOI_NAME"),
		// GroupID: "feedback-group",
		Logger:      l,
		MaxWait:     time.Duration(10000000000),
		MaxAttempts: 10,
	})
	for {
		// the `ReadMessage` method blocks until we receive the next event
		rawMsg, err := r.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}

		var inputRequest khandler.KafkaRequest
		err = json.Unmarshal(rawMsg.Value, &inputRequest)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("sucessfully got from Kafka:", string(rawMsg.Value))

		switch inputRequest.Action {
		case "create-action":
			// TODO: check for inputRequest.Version
			go khandler.CreateFeedback(inputRequest.Payload, repository)
		case "update-action":
			// TODO: check for inputRequest.Version
			go khandler.UpdateFeedback(inputRequest.Payload, repository)
		case "delete-action":
			// TODO: check for inputRequest.Version
			go khandler.DeleteFeedback(inputRequest.Payload, repository)
		default:
			fmt.Println("got unknown action:", inputRequest.Action)
		}
	}
}
