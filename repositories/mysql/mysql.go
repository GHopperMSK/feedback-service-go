package mysqlrepository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
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
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		"db",
		os.Getenv("MYSQL_DATABASE"),
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
	const queryTemplate string = `SELECT * FROM feedbacks WHERE id = ? AND deleted_at IS NULL`

	var feedback repository.Feedback
	result := r.db.QueryRow(queryTemplate, id)
	err := result.Scan(
		&feedback.ID,
		&feedback.ParentId,
		&feedback.SenderId,
		&feedback.ReceiverId,
		&feedback.TradeId,
		&feedback.Message,
		&feedback.Type,
		&feedback.CreatedAt,
		&feedback.UpdatedAt,
		&feedback.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &feedback, nil
}

func (r *mysqlRepository) Find(filter *repository.FeedbackFilter) (*repository.FeedbackResponse, error) {
	feedbacks := make([]*repository.Feedback, 0)

	sql := "SELECT %s FROM feedbacks WHERE 1=1"
	if !filter.WithTrashed {
		sql += " AND deleted_at IS NULL"
	}
	if filter.SenderId != 0 {
		sql += fmt.Sprintf(" AND sender_id = %d", filter.SenderId)
	}
	if filter.ReceiverId != 0 {
		sql += fmt.Sprintf(" AND receiver_id = %d", filter.ReceiverId)
	}
	if filter.TradeId != 0 {
		sql += fmt.Sprintf(" AND trade_id = %d", filter.TradeId)
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
			&feedback.SenderId,
			&feedback.ReceiverId,
			&feedback.TradeId,
			&feedback.Message,
			&feedback.Type,
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

func (r *mysqlRepository) Create(request *repository.FeedbackRequest) (int, error) {
	const queryTemplate string = "INSERT INTO feedbacks(parent_id, sender_id, receiver_id, trade_id, message, type, created_at) VALUES(%s, %d, %d, %d, '%s', '%s', %s)"

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
		request.SenderId,
		request.ReceiverId,
		request.TradeId,
		request.Message,
		request.Type,
		createdAt,
	)
	log.Println(sql)

	res, err := r.db.Exec(sql)
	if err != nil {
		return 0, err
	}

	lastInsertedId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(lastInsertedId), nil
}

func (r *mysqlRepository) Update(id int, request *repository.FeedbackRequest) error {
	const queryTemplate string = "UPDATE feedbacks SET %supdated_at=NOW() WHERE id=%d"

	updatedColumns := ""

	if request.ParentId > 0 {
		updatedColumns += fmt.Sprintf("parent_id=%s, ", strconv.Itoa(request.ParentId))
	}

	if request.SenderId > 0 {
		updatedColumns += fmt.Sprintf("sender_id=%s, ", strconv.Itoa(request.SenderId))
	}

	if request.ReceiverId > 0 {
		updatedColumns += fmt.Sprintf("receiver_id=%s, ", strconv.Itoa(request.ReceiverId))
	}

	if request.TradeId > 0 {
		updatedColumns += fmt.Sprintf("trade_id=%s, ", strconv.Itoa(request.TradeId))
	}

	if request.Message != "" {
		updatedColumns += fmt.Sprintf("message=\"%s\", ", request.Message)
	}

	if request.Type != "" {
		updatedColumns += fmt.Sprintf("type=\"%s\", ", request.Type)
	}

	if request.CreatedAt != "" {
		updatedColumns += fmt.Sprintf("created_at=\"%s\", ", request.CreatedAt)
	}

	sql := fmt.Sprintf(
		queryTemplate,
		updatedColumns,
		id,
	)
	log.Println(sql)

	_, err := r.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func (r *mysqlRepository) Delete(id int) error {
	const queryTemplate string = "DELETE FROM feedbacks WHERE id=%d"

	sql := fmt.Sprintf(
		queryTemplate,
		id,
	)
	log.Println(sql)

	_, err := r.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}
