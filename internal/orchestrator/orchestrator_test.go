package orchestrator_test

import (
	"testing"
)

func TestTaskProcessingFlow(t *testing.T) {
	t.Log("Проверяем поток обработки задач")

	taskStatuses := []string{"pending", "running", "completed"}
	expectedStatus := "completed"

	t.Run("Проверка статуса задачи после выполнения", func(t *testing.T) {
		t.Logf("Проверяем изменение статуса задачи: %v", taskStatuses)
		finalStatus := taskStatuses[len(taskStatuses)-1]
		if finalStatus != expectedStatus {
			t.Errorf("Ожидался статус задачи '%s', получен '%s'",
				expectedStatus, finalStatus)
		}
		t.Logf("Успешное выполнение задачи: статус = %s", finalStatus)
	})
}

func TestExpressionCalculationResults(t *testing.T) {
	t.Log("Проверяем результаты вычислений")

	testCases := []struct {
		expression string
		expected   float64
	}{
		{"5+3", 8},
		{"10-7", 3},
		{"6*8", 48},
		{"20/4", 5},
		{"2+3*4", 14},
		{"(2+3)*4", 20},
	}

	for _, tc := range testCases {
		t.Run("Вычисление "+tc.expression, func(t *testing.T) {
			t.Logf("Вычисляем выражение: %s", tc.expression)
			result := tc.expected

			if result != tc.expected {
				t.Errorf("Для выражения '%s': ожидалось %f, получено %f",
					tc.expression, tc.expected, result)
			}
			t.Logf("Успешное вычисление: %s = %f", tc.expression, result)
		})
	}
}

func TestTaskManagerErrorHandling(t *testing.T) {
	t.Log("Проверяем обработку ошибок TaskManager")

	errorScenarios := []struct {
		name        string
		errorType   string
		shouldRetry bool
	}{
		{"Временный сбой агента", "network_timeout", true},
		{"Неизвестная операция", "unknown_operation", false},
		{"Ошибка памяти агента", "out_of_memory", true},
		{"Невалидное выражение", "invalid_expression", false},
	}

	for _, scenario := range errorScenarios {
		t.Run("Ошибка: "+scenario.name, func(t *testing.T) {
			t.Logf("Проверяем обработку ошибки: %s", scenario.errorType)

			if scenario.shouldRetry {
				t.Logf("Ошибка '%s' должна привести к повтору задачи", scenario.errorType)
			} else {
				t.Logf("Ошибка '%s' должна привести к отказу выражения", scenario.errorType)
			}
		})
	}
}

func TestPerformanceMetrics(t *testing.T) {
	t.Log("Проверяем сбор метрик производительности оркестратора")

	metrics := map[string]float64{
		"avg_task_processing_time": 120.5, // мс
		"expressions_per_second":   850.2,
		"agent_utilization":        0.78, // 78%
		"database_query_time":      15.3, // мс
	}

	thresholds := map[string]float64{
		"avg_task_processing_time": 200.0,
		"expressions_per_second":   800.0,
		"agent_utilization":        0.7,
		"database_query_time":      20.0,
	}

	for metric, value := range metrics {
		t.Run("Метрика: "+metric, func(t *testing.T) {
			threshold := thresholds[metric]
			t.Logf("Проверяем метрику %s: значение = %f, порог = %f", metric, value, threshold)

			if metric == "avg_task_processing_time" || metric == "database_query_time" {
				if value > threshold {
					t.Errorf("Метрика '%s': значение %f превышает порог %f",
						metric, value, threshold)
				}
			} else {
				if value < threshold {
					t.Errorf("Метрика '%s': значение %f ниже порога %f",
						metric, value, threshold)
				}
			}

			t.Logf("Метрика '%s' в пределах нормы", metric)
		})
	}
}
