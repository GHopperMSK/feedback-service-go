package repository

import (
	"database/sql"
	"encoding/json"
	"net/url"
	"reflect"
	"time"
)

type CreateRequest struct {
	ParentId                      int    `json:"parent_id"`
	SenderUuid                    string `json:"sender_uuid"`
	SenderName                    string `json:"sender_name"`
	SenderAvatar                  string `json:"sender_avatar"`
	ReceiverUuid                  string `json:"receiver_uuid"`
	ReceiverName                  string `json:"receiver_name"`
	ReceiverAvatar                string `json:"receiver_avatar"`
	OfferHash                     string `json:"offer_hash"`
	OfferAthorized                bool   `json:"offer_authorized"`
	OfferOwnerUuid                string `json:"offer_owner_uuid"`
	OfferType                     string `json:"offer_type"`
	OfferPaymentMethod            string `json:"offer_payment_method"`
	OfferPaymentMethodSlug        string `json:"offer_payment_method_slug"`
	OfferCurrencyCode             string `json:"offer_currency_code"`
	TradeHash                     string `json:"trade_hash"`
	TradeFiatAmountRequestedInUsd string `json:"trade_fiat_amount_requested_in_usd"`
	TradeStatus                   string `json:"trade_status"`
	Message                       string `json:"message"`
	FeedbackType                  string `json:"feedback_type"`
	CreatedAt                     string `json:"created_at"`
}

type UpdateRequest struct {
	FeedbackId   int    `json:"feedback_id"`
	Message      string `json:"message"`
	FeedbackType string `json:"feedback_type"`
}

type DeleteRequest struct {
	FeedbackId int `json:"feedback_id"`
}

type DeleteOfferRequest struct {
	OfferHash string `json:"offer_hash"`
	DeletedAt string `json:"deleted_at"`
}

type ChangeTradeStatusRequest struct {
	TradeHash   string `json:"trade_hash"`
	TradeStatus string `json:"trade_status"`
}

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
	Find(filter *RequestFilter) (*FeedbackResponse, error)
	Create(request *CreateRequest) (int, error)
	Update(request *UpdateRequest) error
	Delete(request *DeleteRequest) error
	DeleteOffer(request *DeleteOfferRequest) error
	ChangeTradeStatus(request *ChangeTradeStatusRequest) error
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
	ID                            int        `json:"id"`
	ParentId                      NullInt64  `json:"parent_id"`
	SenderUuid                    string     `json:"sender_uuid"`
	SenderName                    string     `json:"sender_name"`
	SenderAvatar                  string     `json:"sender_avatar"`
	ReceiverUuid                  string     `json:"receiver_uuid"`
	ReceiverName                  string     `json:"receiver_name"`
	ReceiverAvatar                string     `json:"receiver_avatar"`
	OfferHash                     string     `json:"offer_hash"`
	OfferAthorized                bool       `json:"offer_authorized"`
	OfferOwnerUuid                string     `json:"offer_owner_uuid"`
	OfferType                     string     `json:"offer_type"`
	OfferPaymentMethod            string     `json:"offer_payment_method"`
	OfferPaymentMethodSlug        string     `json:"offer_payment_method_slug"`
	OfferCurrencyCode             string     `json:"offer_currency_code"`
	OfferDeletedAt                NullString `json:"offer_deleted_at"`
	TradeHash                     string     `json:"trade_hash"`
	TradeFiatAmountRequestedInUsd string     `json:"trade_fiat_amount_requested_in_usd"`
	TradeStatus                   string     `json:"trade_status"`
	Message                       string     `json:"message"`
	FeedbackType                  string     `json:"feedback_type"`
	CreatedAt                     string     `json:"created_at"`
	UpdatedAt                     string     `json:"updated_at"`
	DeletedAt                     NullString `json:"deleted_at"`
}

type FeedbackResponse struct {
	Total  int         `json:"total"`
	Items  []*Feedback `json:"items"`
	Offser int         `json:"offset"`
	Limit  int         `json:"limit"`
}

type RequestFilter struct {
	SenderUuid   string `json:"sender_uuid"`
	ReceiverUuid string `json:"receiver_uuid"`
	OfferHash    string `json:"offer_hash"`
	TradeHash    string `json:"trade_hash"`
	WithTrashed  bool   `json:"with_trashed"`
	Offset       int
	Limit        int
}
