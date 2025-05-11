package database

import (
	"database/sql"
	"log"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/neptship/calc-yandex-go/internal/models"
	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	dir := dbPath[:len(dbPath)-len("/calculator.db")]
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(Schema); err != nil {
		db.Close()
		return nil, err
	}

	log.Println("Database initialized successfully")
	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) GetDB() *sql.DB {
	return d.db
}

func (d *Database) GetUserByLogin(login string) (int, string, error) {
	var id int
	var passwordHash string
	err := d.db.QueryRow("SELECT id, password_hash FROM users WHERE login = ?", login).Scan(&id, &passwordHash)
	return id, passwordHash, err
}

func (d *Database) CreateUser(login, passwordHash string) error {
	_, err := d.db.Exec("INSERT INTO users (login, password_hash) VALUES (?, ?)",
		login, passwordHash)
	return err
}

func (d *Database) CheckUserExists(login string) (bool, error) {
	var exists bool
	err := d.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)", login).Scan(&exists)
	return exists, err
}

func (d *Database) SaveExpression(userID int, expression string, status models.ExpressionStatus) (int, error) {
	result, err := d.db.Exec(
		"INSERT INTO expressions (user_id, expression, status) VALUES (?, ?, ?)",
		userID, expression, status)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (d *Database) UpdateExpressionStatus(id int, status models.ExpressionStatus) error {
	_, err := d.db.Exec("UPDATE expressions SET status = ? WHERE id = ?", status, id)
	return err
}

func (d *Database) SetExpressionResult(id int, result float64) error {
	_, err := d.db.Exec(
		"UPDATE expressions SET status = ?, result = ? WHERE id = ?",
		models.StatusCompleted, result, id)
	return err
}

func (d *Database) GetExpression(id int) (*models.Expression, error) {
	expr := &models.Expression{ID: id}
	var resultValue sql.NullFloat64
	var status string

	err := d.db.QueryRow(
		"SELECT expression, status, result FROM expressions WHERE id = ?",
		id).Scan(&expr.Expression, &status, &resultValue)

	if err != nil {
		return nil, err
	}

	expr.Status = models.ExpressionStatus(status)

	if resultValue.Valid {
		result := resultValue.Float64
		expr.Result = &result
	}

	return expr, nil
}

func (d *Database) GetUserExpressions(userID int) ([]*models.Expression, error) {
	rows, err := d.db.Query(
		"SELECT id, expression, status, result FROM expressions WHERE user_id = ? ORDER BY id DESC",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expressions := []*models.Expression{}
	for rows.Next() {
		expr := &models.Expression{}
		var status string
		var resultValue sql.NullFloat64

		if err := rows.Scan(&expr.ID, &expr.Expression, &status, &resultValue); err != nil {
			return nil, err
		}

		expr.Status = models.ExpressionStatus(status)

		if resultValue.Valid {
			result := resultValue.Float64
			expr.Result = &result
		}

		expressions = append(expressions, expr)
	}

	return expressions, nil
}

func (d *Database) SaveTask(task *models.Task) (int, error) {
	arg1Str := convertArgToString(task.Arg1)
	arg2Str := convertArgToString(task.Arg2)

	result, err := d.db.Exec(
		"INSERT INTO tasks (expression_id, arg1, arg2, operation, operation_time, completed) VALUES (?, ?, ?, ?, ?, ?)",
		task.ExpressionID, arg1Str, arg2Str, task.Operation, task.OperationTime, false)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (d *Database) GetUncompletedTasks(expressionID int) ([]*models.Task, error) {
	rows, err := d.db.Query(
		"SELECT id, expression_id, arg1, arg2, operation, operation_time FROM tasks WHERE expression_id = ? AND completed = 0",
		expressionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []*models.Task{}
	for rows.Next() {
		task := &models.Task{}
		var arg1Str, arg2Str string

		if err := rows.Scan(&task.ID, &task.ExpressionID, &arg1Str, &arg2Str, &task.Operation, &task.OperationTime); err != nil {
			return nil, err
		}

		task.Arg1 = parseArgument(arg1Str)
		task.Arg2 = parseArgument(arg2Str)

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (d *Database) SetTaskResult(taskID int, result float64) error {
	_, err := d.db.Exec(
		"UPDATE tasks SET completed = 1, result = ? WHERE id = ?",
		result, taskID)
	return err
}

func (d *Database) SaveResult(resultID string, expressionID int, taskID *int, value float64, completed bool) error {
	var taskIDValue interface{}
	if taskID != nil {
		taskIDValue = *taskID
	} else {
		taskIDValue = nil
	}

	_, err := d.db.Exec(
		"INSERT INTO results (id, expression_id, task_id, value, completed) VALUES (?, ?, ?, ?, ?)",
		resultID, expressionID, taskIDValue, value, completed)
	return err
}

func (d *Database) GetResult(resultID string) (float64, bool, error) {
	var value float64
	var completed bool

	err := d.db.QueryRow(
		"SELECT value, completed FROM results WHERE id = ?",
		resultID).Scan(&value, &completed)

	return value, completed, err
}

func (d *Database) UpdateResult(resultID string, value float64, completed bool) error {
	_, err := d.db.Exec(
		"UPDATE results SET value = ?, completed = ? WHERE id = ?",
		value, completed, resultID)
	return err
}

func convertArgToString(arg interface{}) string {
	switch v := arg.(type) {
	case float64:
		return "n:" + strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return "s:" + v
	default:
		return ""
	}
}

func parseArgument(argStr string) interface{} {
	if len(argStr) < 2 {
		return nil
	}

	prefix := argStr[0:2]
	value := argStr[2:]

	if prefix == "n:" {
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return value
		}
		return val
	} else if prefix == "s:" {
		return value
	}

	return argStr
}
