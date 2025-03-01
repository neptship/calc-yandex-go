package models

type ExpressionStatus string

const (
	StatusPending    ExpressionStatus = "pending"
	StatusProcessing ExpressionStatus = "processing"
	StatusCompleted  ExpressionStatus = "completed"
	StatusFailed     ExpressionStatus = "failed"
)

type Expression struct {
	ID         int              `json:"id"`
	Expression string           `json:"expression"`
	Status     ExpressionStatus `json:"status"`
	Result     *float64         `json:"result,omitempty"`
}

type Task struct {
	ID            int         `json:"id"`
	Arg1          interface{} `json:"arg1"`
	Arg2          interface{} `json:"arg2"`
	Operation     string      `json:"operation"`
	OperationTime int         `json:"operation_time"`
	ExpressionID  int         `json:"-"`
}

type TaskResult struct {
	ID     int     `json:"id"`
	Result float64 `json:"result"`
}

type Response struct {
	Task Task `json:"task"`
}
