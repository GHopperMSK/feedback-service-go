package mysqlrepository

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	repository "feedback-service-go/repositories"
)

type mysqlRepository struct {
	db *sql.DB
}

func (r *mysqlRepository) GetDB() *sql.DB {
	return r.db
}

func New() (repository.Repository, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s",
		"db_user",
		"secret",
		"localhost",
		"feedback_service",
	)

	dbConnection, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = dbConnection.Ping()
	if err != nil {
		return nil, err
	}

	// dbConnection.SetMaxIdleConns(idleConn)
	// dbConnection.SetMaxOpenConns(maxConn)

	return &mysqlRepository{db: dbConnection}, nil
}

func (r *mysqlRepository) Close() {
	r.db.Close()
}

func (r *mysqlRepository) FindByID(id int) (*repository.Feedback, error) {
	const queryTemplate string = `SELECT id, parent_id, BIN_TO_UUID(sender_uuid), sender_name, sender_avater, BIN_TO_UUID(receiver_uuid), receiver_name, receiver_avater, offer_hash, offer_authorized, BIN_TO_UUID(offer_owner_uuid), offer_type, offer_payment_method, offer_payment_method_slug, offer_currency_code, offer_deleted_at, trade_hash, trade_fiat_amount_requested_in_usd, trade_status, message, feedback_type, created_at, updated_at, deleted_at FROM feedbacks WHERE id = ? AND deleted_at IS NULL`

	var feedback repository.Feedback
	result := r.db.QueryRow(queryTemplate, id)
	err := result.Scan(
		&feedback.ID,
		&feedback.ParentId,
		&feedback.SenderUuid,
		&feedback.SenderName,
		&feedback.SenderAvatar,
		&feedback.ReceiverUuid,
		&feedback.ReceiverName,
		&feedback.ReceiverAvatar,
		&feedback.OfferHash,
		&feedback.OfferAthorized,
		&feedback.OfferOwnerUuid,
		&feedback.OfferType,
		&feedback.OfferPaymentMethod,
		&feedback.OfferPaymentMethodSlug,
		&feedback.OfferCurrencyCode,
		&feedback.DeletedAt,
		&feedback.TradeHash,
		&feedback.TradeFiatAmountRequestedInUsd,
		&feedback.TradeStatus,
		&feedback.Message,
		&feedback.FeedbackType,
		&feedback.CreatedAt,
		&feedback.UpdatedAt,
		&feedback.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &feedback, nil
}

func (r *mysqlRepository) Find(filter *repository.RequestFilter) (*repository.FeedbackResponse, error) {
	feedbacks := make([]*repository.Feedback, 0)

	sql := "SELECT %s FROM feedbacks WHERE 1=1"
	if !filter.WithTrashed {
		sql += " AND deleted_at IS NULL"
	}
	if filter.SenderUuid != "" {
		sql += fmt.Sprintf(" AND sender_uuid = UUID_TO_BIN('%s')", filter.SenderUuid)
	}
	if filter.ReceiverUuid != "" {
		sql += fmt.Sprintf(" AND receiver_uuid = UUID_TO_BIN('%s')", filter.ReceiverUuid)
	}
	if filter.TradeHash != "" {
		sql += fmt.Sprintf(" AND trade_hash = UUID_TO_BIN('%s')", filter.TradeHash)
	}

	var cnt int
	countSql := fmt.Sprintf(sql, "COUNT(*)")
	result := r.db.QueryRow(countSql)
	err := result.Scan(&cnt)
	if err != nil {
		return nil, err
	}

	sql += fmt.Sprintf(" LIMIT %d, %d", filter.Offset, filter.Limit)

	results, err := r.db.Query(fmt.Sprintf(sql, "*"))
	if err != nil {
		return nil, err
	}

	for results.Next() {
		var feedback repository.Feedback

		err = results.Scan(
			&feedback.ID,
			&feedback.ParentId,
			&feedback.SenderUuid,
			&feedback.SenderName,
			&feedback.SenderAvatar,
			&feedback.ReceiverUuid,
			&feedback.ReceiverName,
			&feedback.ReceiverAvatar,
			&feedback.OfferHash,
			&feedback.OfferAthorized,
			&feedback.OfferOwnerUuid,
			&feedback.OfferType,
			&feedback.OfferPaymentMethod,
			&feedback.OfferPaymentMethodSlug,
			&feedback.OfferCurrencyCode,
			&feedback.DeletedAt,
			&feedback.TradeHash,
			&feedback.TradeFiatAmountRequestedInUsd,
			&feedback.TradeStatus,
			&feedback.Message,
			&feedback.FeedbackType,
			&feedback.CreatedAt,
			&feedback.UpdatedAt,
			&feedback.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		feedbacks = append(feedbacks, &feedback)
	}

	response := repository.FeedbackResponse{
		Total:  cnt,
		Items:  feedbacks,
		Offser: filter.Offset,
		Limit:  filter.Limit,
	}

	return &response, nil
}

func (r *mysqlRepository) Create(request *repository.CreateRequest) (int, error) {
	tx, err := r.db.Begin()
	log.Println("transaction start")
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			log.Println(err.Error())
			log.Println("rollback")
			tx.Rollback()
			return
		}
		log.Println("commit")
		err = tx.Commit()
	}()

	const queryTemplate string = "INSERT INTO feedbacks(parent_id, sender_uuid, sender_name, sender_avater, receiver_uuid, receiver_name, receiver_avater, offer_hash, offer_authorized, offer_owner_uuid, offer_type, offer_payment_method, offer_payment_method_slug, offer_currency_code, trade_hash, trade_fiat_amount_requested_in_usd, trade_status, message, feedback_type, created_at) VALUES(%s, UUID_TO_BIN('%s'), '%s', '%s', UUID_TO_BIN('%s'), '%s', '%s', '%s', %t, UUID_TO_BIN('%s'), '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', %s)"

	parentId := "NULL"
	if request.ParentId > 0 {
		parentId = strconv.Itoa(request.ParentId)
	}

	createdAt := "NOW()"
	if request.CreatedAt != "" {
		createdAt = "'" + request.CreatedAt + "'"
	}

	sql := fmt.Sprintf(
		queryTemplate,
		parentId,
		request.SenderUuid,
		request.SenderName,
		request.SenderAvatar,
		request.ReceiverUuid,
		request.ReceiverName,
		request.ReceiverAvatar,
		request.OfferHash,
		request.OfferAthorized,
		request.OfferOwnerUuid,
		request.OfferType,
		request.OfferPaymentMethod,
		request.OfferPaymentMethodSlug,
		request.OfferCurrencyCode,
		request.TradeHash,
		request.TradeFiatAmountRequestedInUsd,
		request.TradeStatus,
		request.Message,
		request.FeedbackType,
		createdAt,
	)
	log.Println(sql)

	res, err := tx.Exec(sql)
	if err != nil {
		return 0, err
	}

	err = createStats(tx, request.ReceiverUuid)
	if err != nil {
		return 0, err
	}

	err = updateStats(tx, request.ReceiverUuid, request.FeedbackType, true)
	if err != nil {
		return 0, err
	}

	lastInsertedId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(lastInsertedId), nil
}

func (r *mysqlRepository) Update(id int, request *repository.UpdateRequest) error {
	tx, err := r.db.Begin()
	log.Println("transaction start")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			log.Println(err.Error())
			log.Println("rollback")
			tx.Rollback()
			return
		}
		log.Println("commit")
		err = tx.Commit()
	}()

	const queryTemplate string = "UPDATE feedbacks SET message=\"%s\", feedback_type=\"%s\", updated_at=NOW() WHERE id=%d"

	feedback, err := r.FindByID(id)
	if err != nil {
		return err
	}

	if request.Message != "" {
		feedback.Message = request.Message
	}

	if request.FeedbackType != "" && request.FeedbackType != feedback.FeedbackType {
		err = updateStats(tx, feedback.ReceiverUuid, feedback.FeedbackType, false)
		if err != nil {
			return err
		}

		err = updateStats(tx, feedback.ReceiverUuid, request.FeedbackType, true)
		if err != nil {
			return err
		}

		feedback.FeedbackType = request.FeedbackType
	}

	sql := fmt.Sprintf(
		queryTemplate,
		feedback.Message,
		feedback.FeedbackType,
		id,
	)
	log.Println(sql)

	_, err = r.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func (r *mysqlRepository) Delete(id int) error {
	tx, err := r.db.Begin()
	log.Println("transaction start")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			log.Println(err.Error())
			log.Println("rollback")
			tx.Rollback()
			return
		}
		log.Println("commit")
		err = tx.Commit()
	}()

	const queryTemplate string = "DELETE FROM feedbacks WHERE id=%d"

	feedback, err := r.FindByID(id)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(
		queryTemplate,
		id,
	)
	log.Println(sql)

	err = updateStats(tx, feedback.ReceiverUuid, feedback.FeedbackType, false)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func createStats(tx *sql.Tx, userUuid string) error {
	// TODO: increase appropriate field
	log.Println("checking for stats")

	row := tx.QueryRow("SELECT user_uuid FROM feedback_stats WHERE user_id=?", userUuid)
	var dbData string
	row.Scan(&dbData)
	if dbData == userUuid {
		log.Println("have found")
		return nil
	}

	log.Println("didn't find")

	const queryTemplate string = "INSERT INTO feedback_stats (user_uuid) VALUES(UUID_TO_BIN('%s'))"

	sql := fmt.Sprintf(
		queryTemplate,
		userUuid,
	)
	log.Println(sql)

	_, err := tx.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func updateStats(tx *sql.Tx, userUuid string, feedbackType string, isIncrease bool) error {
	var queryTemplate string
	if isIncrease {
		queryTemplate = "UPDATE feedback_stats SET %s = %s + 1 WHERE user_uuid=UUID_TO_BIN('%s')"
	} else {
		queryTemplate = "UPDATE feedback_stats SET %s = %s - 1 WHERE user_uuid=UUID_TO_BIN('%s')"
	}

	sql := fmt.Sprintf(
		queryTemplate,
		feedbackType,
		feedbackType,
		userUuid,
	)
	log.Println(sql)

	_, err := tx.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}
