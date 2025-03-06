package orchestrator_test

import (
	"testing"

	"github.com/neptship/calc-yandex-go/internal/config"
	"github.com/neptship/calc-yandex-go/internal/orchestrator"
)

func TestAddValidExpression(t *testing.T) {
	service := orchestrator.NewService(&config.Config{})
	id, err := service.AddExpression("2+2")

	if err != nil {
		t.Fatalf("Не удалось добавить выражение: %v", err)
	}

	if id <= 0 {
		t.Fatalf("Ожидался положительный ID, получено: %d", id)
	}

	expr, err := service.GetExpressionByID(id)
	if err != nil {
		t.Fatalf("Не удалось получить выражение: %v", err)
	}

	if expr.Expression != "2+2" {
		t.Fatalf("Сохранённое выражение %q не совпадает с ожидаемым %q",
			expr.Expression, "2+2")
	}
}

func TestAddInvalidExpression(t *testing.T) {
	service := orchestrator.NewService(&config.Config{})
	_, err := service.AddExpression("2++2")

	if err == nil {
		t.Fatal("Ожидалась ошибка при некорректном выражении, но её нет")
	}

	if err != orchestrator.ErrInvalidExpression {
		t.Fatalf("Ожидалась ошибка %v, получена %v",
			orchestrator.ErrInvalidExpression, err)
	}
}

func TestGetExistingExpression(t *testing.T) {
	service := orchestrator.NewService(&config.Config{})

	id, err := service.AddExpression("3*4")
	if err != nil {
		t.Fatalf("Не удалось добавить выражение: %v", err)
	}

	expr, err := service.GetExpressionByID(id)
	if err != nil {
		t.Fatalf("Не удалось получить выражение: %v", err)
	}

	if expr.ID != id {
		t.Fatalf("ID выражения %d не совпадает с ожидаемым %d", expr.ID, id)
	}
}

func TestGetNonExistingExpression(t *testing.T) {
	service := orchestrator.NewService(&config.Config{})

	_, err := service.GetExpressionByID(9999)

	if err == nil {
		t.Fatal("Ожидалась ошибка при запросе несуществующего выражения")
	}

	if err != orchestrator.ErrExpressionNotFound {
		t.Fatalf("Ожидалась ошибка %v, получена %v",
			orchestrator.ErrExpressionNotFound, err)
	}
}

func TestTaskResult(t *testing.T) {
	service := orchestrator.NewService(&config.Config{})

	exprID, err := service.AddExpression("5+7")
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

	expr, err := service.GetExpressionByID(exprID)
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
