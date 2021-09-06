package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	mysql "feedback-service-go/repositories/mysql"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestGetEmptyFeedbackList(t *testing.T) {
	resetDatabase()

	jsonResponse := getBody("GET", "http://app:8080/feedbacks", nil, http.StatusOK)

	expected := `[]`
	if expected != string(jsonResponse) {
		t.Errorf("Bad response! Expected: %v, extual: %v", expected, string(jsonResponse))
	}
}

func TestCreateFeedback(t *testing.T) {
	resetDatabase()

	var requestJson = []byte(`{"sender_id":1,"parent_id":1,"receiver_id":2,"trade_id":1,"message":"text message","type":"positive","created_at":"2021-09-06 05:01:43"}`)
	jsonResponse := getBody("POST", "http://app:8080/feedback", bytes.NewBuffer(requestJson), http.StatusCreated)

	re, err := regexp.Compile(`{"Message":"text message","Type":"positive","created_at":"2021-09-06 05:01:43","deleted_at":null,"id":\d+,"parent_id":1,"receiver_id":2,"sender_id":1,"trade_id":1,"updated_at":"\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}"}`)
	if err != nil {
		t.Errorf(err.Error())
	}

	found := re.MatchString(string(jsonResponse))
	if !found {
		t.Errorf(err.Error())
	}
}

func TestUpdateFeedback(t *testing.T) {
	resetDatabase()

	var requestJson = []byte(`{"sender_id":1,"receiver_id":2,"trade_id":1,"message":"text message","type":"positive"}`)
	jsonResponse := getBody("POST", "http://app:8080/feedback", bytes.NewBuffer(requestJson), http.StatusCreated)
	re, err := regexp.Compile(`{"Message":"text message","Type":"positive","created_at":"\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}","deleted_at":null,"id":(\d+),"parent_id":null,"receiver_id":2,"sender_id":1,"trade_id":1,"updated_at":"\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}"}`)
	if err != nil {
		t.Errorf(err.Error())
	}

	match := re.FindStringSubmatch(string(jsonResponse))
	id := match[1]
	requestJson = []byte(`{"sender_id":111,"receiver_id":222,"trade_id":333,"message":"text message new","type":"negative"}`)
	getBody("PATCH", "http://app:8080/feedback/"+id, bytes.NewBuffer(requestJson), http.StatusNoContent)

	jsonResponse = getBody("GET", "http://app:8080/feedbacks", nil, http.StatusOK)

	re, err = regexp.Compile(`\[{"Message":"text message new","Type":"negative","created_at":"\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}","deleted_at":null,"id":1,"parent_id":null,"receiver_id":222,"sender_id":111,"trade_id":333,"updated_at":"\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}"}\]`)
	if err != nil {
		t.Errorf(err.Error())
	}

	found := re.MatchString(string(jsonResponse))
	if !found {
		t.Errorf(err.Error())
	}
}

func TestDeleteFeedback(t *testing.T) {
	resetDatabase()

	// create the first feedback
	var requestJson = []byte(`{"sender_id":1,"receiver_id":2,"trade_id":1,"message":"text message","type":"positive"}`)
	jsonResponse := getBody("POST", "http://app:8080/feedback", bytes.NewBuffer(requestJson), http.StatusCreated)
	re, err := regexp.Compile(`{"Message":"text message","Type":"positive","created_at":"\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}","deleted_at":null,"id":(\d+),"parent_id":null,"receiver_id":2,"sender_id":1,"trade_id":1,"updated_at":"\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}"}`)
	if err != nil {
		t.Errorf(err.Error())
	}
	match := re.FindStringSubmatch(string(jsonResponse))
	id := match[1]

	// create the second one
	requestJson = []byte(`{"sender_id":1,"receiver_id":3,"trade_id":2,"message":"text message2","type":"negative"}`)
	getBody("POST", "http://app:8080/feedback", bytes.NewBuffer(requestJson), http.StatusCreated)

	// delete the first one
	getBody("DELETE", fmt.Sprintf("http://app:8080/feedback/%s", id), nil, http.StatusNoContent)

	// check the result
	jsonResponse = getBody("GET", "http://app:8080/feedbacks", nil, http.StatusOK)
	re, err = regexp.Compile(`\[{"Message":"text message2","Type":"negative","created_at":"\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}","deleted_at":null,"id":\d+,"parent_id":null,"receiver_id":3,"sender_id":1,"trade_id":2,"updated_at":"\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}"}\]`)
	if err != nil {
		t.Errorf(err.Error())
	}
	found := re.MatchString(string(jsonResponse))
	if !found {
		t.Errorf(err.Error())
	}
}

func resetDatabase() {
	repo, err := mysql.New()
	if err != nil {
		panic(err.Error())
	}

	loadSQLFile(repo.GetDB(), "init_test.sql")
}

func loadSQLFile(db *sql.DB, sqlFile string) error {
	file, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, q := range strings.Split(string(file), ";") {
		q := strings.TrimSpace(q)
		if q == "" {
			continue
		}
		if _, err := tx.Exec(q); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func sendRequest(method, url string, body io.Reader) *http.Response {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	return resp
}

func getBody(method, url string, body io.Reader, expectedStatusCode int) []byte {
	resp := sendRequest(method, url, body)

	if resp.StatusCode != expectedStatusCode {
		str, _ := fmt.Printf("invalid status code! %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
		panic(str)
	}

	responseRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var responsePayload interface{}
	err = json.Unmarshal(responseRaw, &responsePayload)
	if err != nil {
		return []byte{}
	}

	jsonResponse, err := json.Marshal(responsePayload)
	if err != nil {
		panic(err)
	}

	resp.Body.Close()

	return jsonResponse

}
