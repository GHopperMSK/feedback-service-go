package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	fbhandler "feedback-service-go/handlers"
	mysql "feedback-service-go/repositories/mysql"

	kafka "github.com/segmentio/kafka-go"
)

const (
	topic         = "test"
	brokerAddress = "192.168.0.40:29092"
)

func main() {
	addr, err := net.LookupIP("kafka:29092")
	if err != nil {
		fmt.Println("Unknown host")
	} else {
		fmt.Println("IP address: ", addr)
	}
	time.Sleep(5 * time.Second)
	log.Println("Start server")

	ctx := context.Background()
	// go produce(ctx)
	// go consume(ctx)

	////////////////////////
	i := 0

	l := log.New(os.Stdout, "kafka writer: ", 0)
	// intialize the writer with the broker addresses, and the topic
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"192.168.0.40:29092"},
		Topic:   "test",
		// assign the logger to the writer
		Logger: l,
	})

	for {
		// each kafka message has a key and value. The key is used
		// to decide which partition (and consequently, which broker)
		// the message gets published on
		err := w.WriteMessages(ctx, kafka.Message{
			Key: []byte(strconv.Itoa(i)),
			// create an arbitrary message payload for the value
			Value: []byte("this is message" + strconv.Itoa(i)),
		})
		if err != nil {
			panic("could not write message " + err.Error())
		}

		// log a confirmation once the message is written
		fmt.Println("writes:", i)
		i++
		// sleep for a second
		time.Sleep(time.Second)
	}
	////////////////////////

	repository, err := mysql.New()
	if err != nil {
		panic(err.Error())
	}
	log.Println("Successfully connected to the storage")

	feedbackHandler := fbhandler.New(repository)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/feedback/{id}", feedbackHandler.GetFeedback).Methods("GET")
	router.HandleFunc("/feedbacks", feedbackHandler.GetFeedbacksByFilter).Methods("GET")
	router.HandleFunc("/feedback", feedbackHandler.CreateFeedback).Methods("POST")
	router.HandleFunc("/feedback/{id}", feedbackHandler.UpdateFeedback).Methods("PATCH")
	router.HandleFunc("/feedback/{id}", feedbackHandler.DeleteFeedback).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func produce(ctx context.Context) {
	// initialize a counter
	i := 0

	l := log.New(os.Stdout, "kafka writer: ", 0)
	// intialize the writer with the broker addresses, and the topic
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerAddress},
		Topic:   topic,
		// assign the logger to the writer
		Logger: l,
	})

	for {
		// each kafka message has a key and value. The key is used
		// to decide which partition (and consequently, which broker)
		// the message gets published on
		err := w.WriteMessages(ctx, kafka.Message{
			Key: []byte(strconv.Itoa(i)),
			// create an arbitrary message payload for the value
			Value: []byte("this is message" + strconv.Itoa(i)),
		})
		if err != nil {
			panic("could not write message " + err.Error())
		}

		// log a confirmation once the message is written
		fmt.Println("writes:", i)
		i++
		// sleep for a second
		time.Sleep(time.Second)
	}
}

func consume(ctx context.Context) {
	// create a new logger that outputs to stdout
	// and has the `kafka reader` prefix
	l := log.New(os.Stdout, "kafka reader: ", 0)
	// initialize a new reader with the brokers and topic
	// the groupID identifies the consumer and prevents
	// it from receiving duplicate messages
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddress},
		Topic:   topic,
		GroupID: "my-group",
		// assign the logger to the reader
		Logger: l,
	})
	for {
		// the `ReadMessage` method blocks until we receive the next event
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}
		// after receiving the message, log its value
		fmt.Println("received: ", string(msg.Value))
	}
}
