package handlers

import (
	"database/sql"
	"encoding/json"
	repository "feedback-service-go/repositories"
	"io/ioutil"
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

func (h *restHandler) CreateFeedback(w http.ResponseWriter, r *http.Request) {
	log.Println("CreateFeedback")
	reqBody, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		panic("Kindly enter data with the event title and description only in order to update")
	}

	var feedbackRequest repository.FeedbackRequest
	err = json.Unmarshal(reqBody, &feedbackRequest)
	if err != nil {
		panic(err.Error())
	}

	validErrs := feedbackRequest.Validate()
	if len(validErrs) > 0 {
		err := map[string]interface{}{"validationError": validErrs}
		w.Header().Set("Content-type", "applciation/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	id, err := h.repo.Create(&feedbackRequest)
	if err != nil {
		panic(err.Error())
	}

	feedback, err := h.GetById(int(id))
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(feedback)
}

func (h *restHandler) UpdateFeedback(w http.ResponseWriter, r *http.Request) {
	log.Println("UpdateFeedback")
	inputFeedbackID := mux.Vars(r)["id"]
	var feedbackRequest repository.FeedbackRequest

	reqBody, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		panic("Kindly enter data with the event title and description only in order to update")
	}

	err = json.Unmarshal(reqBody, &feedbackRequest)
	if err != nil {
		panic(err.Error())
	}

	feedbackID, err := strconv.Atoi(inputFeedbackID)
	if err != nil {
		panic("can't convert string to int")
	}

	if validErrs := feedbackRequest.Validate(); len(validErrs) > 0 {
		err := map[string]interface{}{"validationError": validErrs}
		w.Header().Set("Content-type", "applciation/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	err = h.repo.Update(feedbackID, &feedbackRequest)
	if err != nil {
		panic(err.Error())
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *restHandler) DeleteFeedback(w http.ResponseWriter, r *http.Request) {
	log.Println("DeleteFeedback")
	inputFeedbackID := mux.Vars(r)["id"]
	feedbackID, err := strconv.Atoi(inputFeedbackID)
	if err != nil {
		panic("can't convert string to int")
	}

	err = h.repo.Delete(feedbackID)
	if err != nil {
		panic(err.Error())
	}

	w.WriteHeader(http.StatusNoContent)
}

func getFilter(query url.Values) *repository.FeedbackFilter {
	var err error

	filter := repository.FeedbackFilter{}

	filter.SenderId, err = getIntParam(query, "sender_id", 0)
	if err != nil {
		panic(err.Error())
	}

	filter.ReceiverId, err = getIntParam(query, "receiver_id", 0)
	if err != nil {
		panic(err.Error())
	}

	filter.TradeId, err = getIntParam(query, "trade_id", 0)
	if err != nil {
		panic(err.Error())
	}

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
