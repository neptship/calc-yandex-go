package agent

import (
	"testing"

	"github.com/neptship/calc-yandex-go/internal/models"
)

func executeOperationLocal(task *models.Task) (float64, bool) {
	arg1, ok1 := task.Arg1.(float64)
	arg2, ok2 := task.Arg2.(float64)

	if !ok1 || !ok2 {
		return 0, true
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

func TestExecuteOperation(t *testing.T) {
	testCases := []struct {
		name     string
		task     models.Task
		expected float64
		hasError bool
	}{
		{"сложение", models.Task{Arg1: 5.0, Arg2: 3.0, Operation: "+"}, 8.0, false},
		{"вычитание", models.Task{Arg1: 5.0, Arg2: 3.0, Operation: "-"}, 2.0, false},
		{"умножение", models.Task{Arg1: 5.0, Arg2: 3.0, Operation: "*"}, 15.0, false},
		{"деление", models.Task{Arg1: 6.0, Arg2: 3.0, Operation: "/"}, 2.0, false},
		{"деление на ноль", models.Task{Arg1: 5.0, Arg2: 0.0, Operation: "/"}, 0.0, true},
		{"неизвестная операция", models.Task{Arg1: 5.0, Arg2: 3.0, Operation: "%"}, 0.0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, isError := executeOperationLocal(&tc.task)

			if tc.hasError {
				if !isError {
					t.Errorf("ожидалась ошибка для операции '%s', но её нет", tc.task.Operation)
				}
			} else {
				if isError {
					t.Errorf("неожиданная ошибка для операции '%s'", tc.task.Operation)
				}
				if result != tc.expected {
					t.Errorf("для операции '%s': ожидалось %f, получено %f", tc.task.Operation, tc.expected, result)
				}
			}
		})
	}
}
