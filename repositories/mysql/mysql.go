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

	tx, err := r.GetDB().Begin()
	if err != nil {
		return 0, err
	}

	res, err := r.db.Exec(sql)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = createStats(r, request.ReceiverId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = updateStats(r, request.ReceiverId, request.Type, true)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	lastInsertedId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(lastInsertedId), nil
}

func (r *mysqlRepository) Update(id int, request *repository.FeedbackRequest) error {
	const queryTemplate string = "UPDATE feedbacks SET parent_id=%d, sender_id=%d, receiver_id=%d, trade_id=%d, message=\"%s\", type=\"%s\", created_at=\"%s\", updated_at=NOW() WHERE id=%d"

	feedback, err := r.FindByID(id)
	if err != nil {
		return err
	}

	if request.ParentId != 0 {
		feedback.ParentId.Int64 = int64(request.ParentId)
	}

	if request.SenderId != 0 {
		feedback.SenderId = request.SenderId
	}

	if request.ReceiverId != 0 {
		feedback.ReceiverId = request.ReceiverId
	}

	if request.TradeId != 0 {
		feedback.TradeId = request.TradeId
	}

	if request.Message != "" {
		feedback.Message = request.Message
	}

	tx, err := r.GetDB().Begin()
	if err != nil {
		return err
	}

	if request.Type != "" && request.Type != feedback.Type {
		err = updateStats(r, id, feedback.Type, false)
		if err != nil {
			tx.Rollback()
			return err
		}

		err = updateStats(r, id, request.Type, true)
		if err != nil {
			tx.Rollback()
			return err
		}

		feedback.Type = request.Type
	}

	if request.CreatedAt != "" {
		feedback.CreatedAt = request.CreatedAt
	}

	sql := fmt.Sprintf(
		queryTemplate,
		feedback.ParentId.Int64,
		feedback.SenderId,
		feedback.ReceiverId,
		feedback.TradeId,
		feedback.Message,
		feedback.Type,
		feedback.CreatedAt,
		id,
	)
	log.Println(sql)

	_, err = r.db.Exec(sql)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (r *mysqlRepository) Delete(id int) error {
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

	tx, err := r.GetDB().Begin()
	if err != nil {
		return err
	}

	err = updateStats(r, feedback.ReceiverId, feedback.Type, false)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = r.GetDB().Exec(sql)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func createStats(r *mysqlRepository, userId int) error {
	// TODO: increase appropriate field
	log.Println("checking for stats")

	row := r.db.QueryRow("SELECT user_id FROM feedback_stats WHERE user_id=?", userId)
	var dbData int
	row.Scan(&dbData)
	if dbData == userId {
		log.Println("have found")
		return nil
	}

	log.Println("didn't find")

	const queryTemplate string = "INSERT INTO feedback_stats (user_id) VALUES(%d)"

	sql := fmt.Sprintf(
		queryTemplate,
		userId,
	)
	log.Println(sql)

	_, err := r.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func updateStats(r *mysqlRepository, userId int, feedbackType string, isIncrease bool) error {
	var queryTemplate string
	if isIncrease {
		queryTemplate = "UPDATE feedback_stats SET %s = %s + 1 WHERE user_id=%d"
	} else {
		queryTemplate = "UPDATE feedback_stats SET %s = %s - 1 WHERE user_id=%d"
	}

	sql := fmt.Sprintf(
		queryTemplate,
		feedbackType,
		feedbackType,
		userId,
	)
	log.Println(sql)

	_, err := r.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}
