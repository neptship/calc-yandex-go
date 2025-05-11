package orchestrator_test

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/neptship/calc-yandex-go/internal/config"
	"github.com/neptship/calc-yandex-go/internal/database"
	"github.com/neptship/calc-yandex-go/internal/orchestrator"
	_ "modernc.org/sqlite"
)

const testDBPath = "test_calculator.db"

func setupTestDB(t *testing.T) *database.Database {
	os.Remove(testDBPath)

	db, err := database.NewDatabase(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	return db
}

func teardownTestDB(db *database.Database) {
	if db != nil {
		db.Close()
	}
	os.Remove(testDBPath)
}

func createTestUser(t *testing.T, db *database.Database) int {
	result, err := db.GetDB().Exec("INSERT INTO users (login, password_hash) VALUES (?, ?)",
		"testuser", "hashedpassword")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get user ID: %v", err)
	}

	return int(userID)
}

func TestAddValidExpression(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	userID := createTestUser(t, db)
	service := orchestrator.NewService(&config.Config{}, db)

	id, err := service.AddExpression(userID, "2+2")

	if err != nil {
		t.Fatalf("Не удалось добавить выражение: %v", err)
	}

	if id <= 0 {
		t.Fatalf("Ожидался положительный ID, получено: %d", id)
	}

	expr, err := service.GetExpressionByID(userID, id)
	if err != nil {
		t.Fatalf("Не удалось получить выражение: %v", err)
	}

	if expr.Expression != "2+2" {
		t.Fatalf("Сохранённое выражение %q не совпадает с ожидаемым %q",
			expr.Expression, "2+2")
	}
}

func TestAddInvalidExpression(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	userID := createTestUser(t, db)
	service := orchestrator.NewService(&config.Config{}, db)

	_, err := service.AddExpression(userID, "2++2")

	if err == nil {
		t.Fatal("Ожидалась ошибка при некорректном выражении, но её нет")
	}

	if err != orchestrator.ErrInvalidExpression {
		t.Fatalf("Ожидалась ошибка %v, получена %v",
			orchestrator.ErrInvalidExpression, err)
	}
}

func TestGetExistingExpression(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	userID := createTestUser(t, db)
	service := orchestrator.NewService(&config.Config{}, db)

	id, err := service.AddExpression(userID, "3*4")
	if err != nil {
		t.Fatalf("Не удалось добавить выражение: %v", err)
	}

	expr, err := service.GetExpressionByID(userID, id)
	if err != nil {
		t.Fatalf("Не удалось получить выражение: %v", err)
	}

	if expr.ID != id {
		t.Fatalf("ID выражения %d не совпадает с ожидаемым %d", expr.ID, id)
	}
}

func TestGetNonExistingExpression(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	userID := createTestUser(t, db)
	service := orchestrator.NewService(&config.Config{}, db)

	_, err := service.GetExpressionByID(userID, 9999)

	if err == nil {
		t.Fatal("Ожидалась ошибка при запросе несуществующего выражения")
	}

	if err != orchestrator.ErrExpressionNotFound {
		t.Fatalf("Ожидалась ошибка %v, получена %v",
			orchestrator.ErrExpressionNotFound, err)
	}
}

func TestTaskResult(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	userID := createTestUser(t, db)
	service := orchestrator.NewService(&config.Config{}, db)

	exprID, err := service.AddExpression(userID, "5+7")
	if err != nil {
		t.Fatalf("Не удалось добавить выражение: %v", err)
	}

	task, err := service.GetNextTask()
	if err != nil {
		t.Fatalf("Не удалось получить задачу: %v", err)
	}

	err = service.SetTaskResult(task.ID, 12.0)
	if err != nil {
		t.Fatalf("Не удалось установить результат: %v", err)
	}

	expr, err := service.GetExpressionByID(userID, exprID)
	if err != nil {
		t.Fatalf("Не удалось получить выражение: %v", err)
	}

	if expr.Result == nil {
		t.Fatal("Результат выражения равен nil")
	}

	if *expr.Result != 12.0 {
		t.Fatalf("Ожидался результат %f, получен %f", 12.0, *expr.Result)
	}
}

func TestUnauthorizedAccess(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	user1ID := createTestUser(t, db)

	result, err := db.GetDB().Exec("INSERT INTO users (login, password_hash) VALUES (?, ?)",
		"testuser2", "hashedpassword2")
	if err != nil {
		t.Fatalf("Failed to create second test user: %v", err)
	}

	user2ID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get second user ID: %v", err)
	}

	service := orchestrator.NewService(&config.Config{}, db)

	exprID, err := service.AddExpression(user1ID, "10-5")
	if err != nil {
		t.Fatalf("Не удалось добавить выражение: %v", err)
	}

	_, err = service.GetExpressionByID(int(user2ID), exprID)

	if err == nil {
		t.Fatal("Ожидалась ошибка при доступе к чужому выражению")
	}

	if err != orchestrator.ErrUnauthorized {
		t.Fatalf("Ожидалась ошибка %v, получена %v",
			orchestrator.ErrUnauthorized, err)
	}
}

func TestAddSimpleExpression(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	userID := createTestUser(t, db)
	service := orchestrator.NewService(&config.Config{}, db)

	id, err := service.AddSimpleExpression(userID, "42")

	if err != nil {
		t.Fatalf("Не удалось добавить простое выражение: %v", err)
	}

	expr, err := service.GetExpressionByID(userID, id)
	if err != nil {
		t.Fatalf("Не удалось получить выражение: %v", err)
	}

	if expr.Result == nil {
		t.Fatal("Результат выражения равен nil")
	}

	if *expr.Result != 42.0 {
		t.Fatalf("Ожидался результат %f, получен %f", 42.0, *expr.Result)
	}
}
