package handlers

import (
	"database/sql"
	"encoding/json"
	repository "feedback-service-go/repositories"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
)

const (
	defaultLimit = 10
	maxLimit     = 1000
)

type restHandler struct {
	repo repository.Repository
}

func New(repo repository.Repository) *restHandler {
	return &restHandler{
		repo: repo,
	}
}

func (h *restHandler) GetFeedback(w http.ResponseWriter, r *http.Request) {
	log.Println("GetFeedback")

	inputFeedbackID := mux.Vars(r)["id"]
	feedbackID, err := strconv.Atoi(inputFeedbackID)
	if err != nil {
		panic("can't convert string to int")
	}

	feedback, err := h.GetById(feedbackID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			w.WriteHeader(http.StatusNotFound)
			return
		default:
			panic(err.Error())
		}

	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(feedback)
}

func (h *restHandler) GetFeedbacksByFilter(w http.ResponseWriter, r *http.Request) {
	log.Println("GetFeedbacksByFilter")

	filter := getFilter(r.URL.Query())

	response, err := h.repo.Find(filter)
	if err != nil {
		panic(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *restHandler) GetById(id int) (*repository.Feedback, error) {
	return h.repo.FindByID(id)
}

func getFilter(query url.Values) *repository.RequestFilter {
	var err error

	filter := repository.RequestFilter{}

	filter.SenderUuid = query.Get("sender_uuid")
	filter.ReceiverUuid = query.Get("receiver_uuid")
	filter.TradeHash = query.Get("trade_hase")
	inputWithTrashed := query.Get("with_trashed")
	if inputWithTrashed == "1" {
		filter.WithTrashed = true
	} else {
		filter.WithTrashed = false
	}

	filter.Offset, err = getIntParam(query, "offset", 0)
	if err != nil {
		panic(err.Error())
	}

	intVal := 0
	inputLimit := query.Get("limit")
	if inputLimit != "" {
		var err error
		intVal, err = strconv.Atoi(inputLimit)
		if err != nil {
			panic(err.Error())
		}
	} else {
		intVal = defaultLimit
	}
	filter.Limit = min(intVal, maxLimit)

	return &filter
}

func getIntParam(query url.Values, paramName string, defaultValue int) (int, error) {
	inputSenderId := query.Get(paramName)
	if inputSenderId != "" {
		intVal, err := strconv.Atoi(inputSenderId)
		if err != nil {
			return 0, err
		}

		return intVal, nil
	}
	return defaultValue, nil
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
