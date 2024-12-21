package application

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/neptship/calc-yandex-go/pkg/calculation"
)

type Config struct {
	Addr string
}

type Application struct {
	config *Config
}

type Request struct {
	Expression string `json:"expression"`
}

type Response struct {
	Result float64 `json:"result,omitempty"`
	Error  string  `json:"error,omitempty"`
}

func New() *Application {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return &Application{
		config: &Config{
			Addr: port,
		},
	}
}

func CalcHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendError(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}
	request := new(Request)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		SendError(w, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	if request.Expression == "" {
		SendError(w, http.StatusBadRequest, ErrEmptyExpression)
		return
	}

	result, err := calculation.Calc(request.Expression)

	if err != nil {
		if err == calculation.ErrInvalidCharacter || err == calculation.ErrInvalidNumber || err == calculation.ErrConsecutiveOperators || err == calculation.ErrMismatchedBrackets || err == calculation.ErrDivisionByZero {
			SendError(w, http.StatusUnprocessableEntity, err)
		} else {
			SendError(w, http.StatusInternalServerError, ErrServerError)
		}
		return
	}
	SendJSON(w, http.StatusOK, Response{
		Result: result,
	})
}

func SendError(w http.ResponseWriter, code int, err error) {
	SendJSON(w, code, Response{
		Error: err.Error(),
	})
}

func SendJSON(w http.ResponseWriter, code int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(response)
}

func (a *Application) RunServer() error {
	http.HandleFunc("/api/v1/calculate", CalcHandler)
	return http.ListenAndServe(":"+a.config.Addr, nil)
}
