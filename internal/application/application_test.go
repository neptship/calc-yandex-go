package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApplication(t *testing.T) {
	TestApplicationCases := []struct {
		name               string
		expression         string
		expectedResult     float64
		expectedStatusCode int
		wantError          bool
	}{
		{
			name:               "negative",
			expression:         "-5*(3-6)",
			expectedResult:     15,
			expectedStatusCode: 200,
			wantError:          false,
		},
	}
	for _, testCase := range TestApplicationCases {
		t.Run(testCase.name, func(t *testing.T) {
			jsonExpression := fmt.Sprintf("{\"expression\": \"%s\"}", testCase.expression)
			request, _ := http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte(jsonExpression)))
			response := httptest.NewRecorder()
			CalcHandler(response, request)
			if testCase.wantError {
				if testCase.expectedStatusCode != response.Code {
					t.Errorf("expected status code %d, but got %d", testCase.expectedStatusCode, response.Code)
				}
			} else {
				var result Response
				if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if result.Result != testCase.expectedResult {
					t.Errorf("expected result %f, but got %f", testCase.expectedResult, result.Result)
				}
			}
		})
	}

}
