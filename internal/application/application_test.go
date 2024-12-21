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
			name:               "simple addition",
			expression:         "2+2",
			expectedResult:     4,
			expectedStatusCode: 200,
			wantError:          false,
		},
		{
			name:               "complex expression",
			expression:         "3+5*2-8/4",
			expectedResult:     11,
			expectedStatusCode: 200,
			wantError:          false,
		},
		{
			name:               "negative multiplication",
			expression:         "-10*(-5+2)",
			expectedResult:     30,
			expectedStatusCode: 200,
			wantError:          false,
		},
		{
			name:               "invalid character",
			expression:         "2+a",
			expectedResult:     0,
			expectedStatusCode: 422,
			wantError:          true,
		},
		{
			name:               "division by zero",
			expression:         "10/0",
			expectedResult:     0,
			expectedStatusCode: 422,
			wantError:          true,
		},
		{
			name:               "consecutive operators",
			expression:         "5++3",
			expectedResult:     0,
			expectedStatusCode: 422,
			wantError:          true,
		},
		{
			name:               "mismatched parentheses",
			expression:         "2*(3+2",
			expectedResult:     0,
			expectedStatusCode: 422,
			wantError:          true,
		},
		{
			name:               "empty expression",
			expression:         "",
			expectedResult:     0,
			expectedStatusCode: 400,
			wantError:          true,
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
