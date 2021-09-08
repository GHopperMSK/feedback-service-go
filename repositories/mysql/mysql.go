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

type MysqlRepository struct {
	db *sql.DB
}

func (r *MysqlRepository) GetDB() *sql.DB {
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

	return &MysqlRepository{db: dbConnection}, nil
}

func (r *MysqlRepository) Close() {
	r.db.Close()
}

func (r *MysqlRepository) FindByID(id int) (*repository.Feedback, error) {
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

func (r *MysqlRepository) Find(filter *repository.FeedbackFilter) ([]*repository.Feedback, error) {
	feedbacks := make([]*repository.Feedback, 0)

	sql := "SELECT * FROM feedbacks WHERE 1=1"
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

	results, err := r.db.Query(sql)
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

	return feedbacks, nil
}

func (r *MysqlRepository) Create(request *repository.FeedbackRequest) (int, error) {
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

func (r *MysqlRepository) Update(id int, request *repository.FeedbackRequest) error {
	const queryTemplate string = "UPDATE feedbacks SET parent_id=%s, sender_id=%d, receiver_id=%d, trade_id=%d, message='%s', type='%s', created_at=%s, updated_at=NOW() WHERE id=%d"

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
		id,
	)
	log.Println(sql)

	_, err := r.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func (r *MysqlRepository) Delete(id int) error {
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
