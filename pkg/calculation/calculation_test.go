package calculation_test

import (
	"testing"

	"github.com/neptship/calc-yandex-go/pkg/calculation"
)

func TestCalculate(t *testing.T) {
	testCases := []struct {
		name     string
		expr     string
		expected float64
		hasError bool
	}{
		{"простое сложение", "2+3", 5, false},
		{"сложение и умножение", "2+3*4", 14, false},
		{"скобки", "(2+3)*4", 20, false},
		{"деление", "10/2", 5, false},
		{"деление на ноль", "1/0", 0, true},
		{"некорректное выражение", "2++3", 0, true},
		{"пустое выражение", "", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := calculation.Calc(tc.expr)

			if tc.hasError {
				if err == nil {
					t.Errorf("ожидалась ошибка для выражения '%s', но её нет", tc.expr)
				}
			} else {
				if err != nil {
					t.Errorf("неожиданная ошибка для выражения '%s': %v", tc.expr, err)
				}
				if result != tc.expected {
					t.Errorf("для выражения '%s': ожидалось %f, получено %f", tc.expr, tc.expected, result)
				}
			}
		})
	}
}
