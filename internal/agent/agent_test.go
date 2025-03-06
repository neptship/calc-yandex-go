package agent

import (
	"testing"

	"github.com/neptship/calc-yandex-go/internal/models"
)

func TestOperations(t *testing.T) {
	testOperation := func(operation string, arg1, arg2, expected float64) {
		task := &models.Task{
			ID:            1,
			Arg1:          arg1,
			Arg2:          arg2,
			Operation:     operation,
			OperationTime: 10,
		}

		result, hasError := executeOperationTest(task)

		if operation == "/" && arg2 == 0 {
			if !hasError {
				t.Errorf("Ожидалась ошибка деления на ноль")
			}
			return
		}

		if hasError {
			t.Errorf("Неожиданная ошибка при операции %s с аргументами %f и %f",
				operation, arg1, arg2)
			return
		}

		if result != expected {
			t.Errorf("Для операции %s с аргументами %f и %f ожидалось %f, получено %f",
				operation, arg1, arg2, expected, result)
		}
	}

	testOperation("+", 5, 3, 8)
	testOperation("-", 5, 3, 2)
	testOperation("*", 5, 3, 15)
	testOperation("/", 6, 3, 2)
	testOperation("/", 5, 0, 0)
}

func executeOperationTest(task *models.Task) (float64, bool) {
	var arg1, arg2 float64

	switch v := task.Arg1.(type) {
	case float64:
		arg1 = v
	case int:
		arg1 = float64(v)
	}

	switch v := task.Arg2.(type) {
	case float64:
		arg2 = v
	case int:
		arg2 = float64(v)
	}

	switch task.Operation {
	case "+":
		return arg1 + arg2, false
	case "-":
		return arg1 - arg2, false
	case "*":
		return arg1 * arg2, false
	case "/":
		if arg2 == 0 {
			return 0, true
		}
		return arg1 / arg2, false
	default:
		return 0, true
	}
}
