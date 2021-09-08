package repository

import (
	"database/sql"
	"encoding/json"
	"net/url"
	"reflect"
	"time"
)

type FeedbackRequest struct {
	ParentId   int `json:"parent_id"`
	SenderId   int `json:"sender_id"`
	ReceiverId int `json:"receiver_id"`
	TradeId    int `json:"trade_id"`
	Message    string
	Type       string
	CreatedAt  string `json:"created_at"`
}

func (request *FeedbackRequest) Validate() url.Values {
	errs := url.Values{}

	if request.ParentId < 0 {
		errs.Add("parent_id", "The parent_id field must be a positive number!")
	}

	if request.SenderId < 1 {
		errs.Add("sender_id", "The sender_id field is required and must be more then 0!")
	}

	if request.ReceiverId < 1 {
		errs.Add("receiver_id", "The receiver_id field is required and must be more then 0!")
	}

	if request.TradeId < 1 {
		errs.Add("trade_id", "The trade_id field is required and must be more then 0!")
	}

	if len(request.Message) < 5 {
		errs.Add("message", "The message field must be longer than 5 chars!")
	}

	if request.Type != "positive" && request.Type != "negative" {
		errs.Add("type", "The type field must be either 'positive' or 'negative'!")
	}

	if request.CreatedAt != "" {
		_, err := time.Parse("2006-01-02 15:04:05", request.CreatedAt)
		if err != nil {
			errs.Add("created_at", err.Error())
		}
	}

	return errs
}

type Repository interface {
	GetDB() *sql.DB
	Close()
	FindByID(id int) (*Feedback, error)
	Find(filter *FeedbackFilter) ([]*Feedback, error)
	Create(request *FeedbackRequest) (int, error)
	Update(id int, request *FeedbackRequest) error
	Delete(id int) error
}

type NullInt64 sql.NullInt64
type NullString sql.NullString

func (ni *NullInt64) Scan(value interface{}) error {
	var i sql.NullInt64
	if err := i.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*ni = NullInt64{i.Int64, false}
	} else {
		*ni = NullInt64{i.Int64, true}
	}
	return nil
}

func (ni *NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ni.Int64)
}

func (ni *NullString) Scan(value interface{}) error {
	var i sql.NullString
	if err := i.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*ni = NullString{i.String, false}
	} else {
		*ni = NullString{i.String, true}
	}
	return nil
}

func (ni *NullString) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ni.String)
}

type Feedback struct {
	ID         int       `json:"id"`
	ParentId   NullInt64 `json:"parent_id"`
	SenderId   int       `json:"sender_id"`
	ReceiverId int       `json:"receiver_id"`
	TradeId    int       `json:"trade_id"`
	Message    string
	Type       string
	CreatedAt  string     `json:"created_at"`
	UpdatedAt  string     `json:"updated_at"`
	DeletedAt  NullString `json:"deleted_at"`
}

type FeedbackFilter struct {
	SenderId    int  `json:"sender_id"`
	ReceiverId  int  `json:"receiver_id"`
	TradeId     int  `json:"trade_id"`
	WithTrashed bool `json:"with_trashed"`
}
