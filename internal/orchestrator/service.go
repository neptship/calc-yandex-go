package orchestrator

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/neptship/calc-yandex-go/internal/config"
	"github.com/neptship/calc-yandex-go/internal/models"
	"github.com/neptship/calc-yandex-go/pkg/calculation"
)

var (
	ErrExpressionNotFound = errors.New("expression not found")
	ErrTaskNotFound       = errors.New("task not found")
	ErrInvalidExpression  = errors.New("invalid expression")
	ErrInvalidTaskResult  = errors.New("invalid task result")
)

type ExpressionResult struct {
	Value     float64
	Completed bool
}

type Service struct {
	expressions      map[int]*models.Expression
	tasks            map[int]*models.Task
	results          map[string]*ExpressionResult
	pendingTasks     []*models.Task
	nextExpressionID int
	nextTaskID       int
	mu               sync.Mutex
	config           *config.Config
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		expressions:      make(map[int]*models.Expression),
		tasks:            make(map[int]*models.Task),
		results:          make(map[string]*ExpressionResult),
		pendingTasks:     []*models.Task{},
		nextExpressionID: 1,
		nextTaskID:       1,
		config:           cfg,
	}
}

func (s *Service) AddExpression(expressionStr string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ops, err := calculation.ParseExpression(expressionStr)
	if err != nil {
		log.Printf("Ошибка парсинга выражения: %v", err)
		return 0, ErrInvalidExpression
	}

	expressionID := s.nextExpressionID
	s.nextExpressionID++

	expression := &models.Expression{
		ID:         expressionID,
		Expression: expressionStr,
		Status:     models.StatusPending,
		Result:     nil,
	}
	s.expressions[expressionID] = expression
	log.Printf("Добавлено выражение ID=%d: %s", expressionID, expressionStr)

	expression.Status = models.StatusProcessing
	s.createTasksFromOperations(expressionID, ops)

	return expressionID, nil
}

func (s *Service) GetExpressionByID(id int) (*models.Expression, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expr, exists := s.expressions[id]
	if !exists {
		return nil, ErrExpressionNotFound
	}
	return expr, nil
}

func (s *Service) GetAllExpressions() []*models.Expression {
	s.mu.Lock()
	defer s.mu.Unlock()

	expressions := make([]*models.Expression, 0, len(s.expressions))
	for _, expr := range s.expressions {
		expressions = append(expressions, expr)
	}
	return expressions
}

func (s *Service) GetNextTask() (*models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, task := range s.pendingTasks {
		canExecute := true

		var arg1 interface{} = task.Arg1
		if arg1ID, isString := task.Arg1.(string); isString {
			if result, exists := s.results[arg1ID]; !exists || !result.Completed {
				canExecute = false
			} else {
				arg1 = result.Value
			}
		}

		var arg2 interface{} = task.Arg2
		if arg2ID, isString := task.Arg2.(string); isString {
			if result, exists := s.results[arg2ID]; !exists || !result.Completed {
				canExecute = false
			} else {
				arg2 = result.Value
			}
		}

		if canExecute {
			s.pendingTasks = append(s.pendingTasks[:i], s.pendingTasks[i+1:]...)

			taskToExecute := &models.Task{
				ID:           task.ID,
				Operation:    task.Operation,
				Arg1:         arg1,
				Arg2:         arg2,
				ExpressionID: task.ExpressionID,
			}

			switch task.Operation {
			case "+":
				taskToExecute.OperationTime = s.config.AdditionMs
			case "-":
				taskToExecute.OperationTime = s.config.SubtractionMs
			case "*":
				taskToExecute.OperationTime = s.config.MultiplicationMs
			case "/":
				taskToExecute.OperationTime = s.config.DivisionMs
			}

			log.Printf("Выдана задача ID=%d, операция=%s для выражения ID=%d",
				taskToExecute.ID, taskToExecute.Operation, taskToExecute.ExpressionID)
			return taskToExecute, nil
		}
	}
	return nil, ErrTaskNotFound
}

func (s *Service) SetTaskResult(id int, result float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return ErrTaskNotFound
	}

	resultID := getResultID(task.ExpressionID, id)
	s.results[resultID] = &ExpressionResult{
		Value:     result,
		Completed: true,
	}
	log.Printf("Получен результат для задачи ID=%d: %f", id, result)

	s.checkExpressionCompletion(task.ExpressionID)

	return nil
}

func (s *Service) AddSimpleExpression(expressionStr string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, err := strconv.ParseFloat(expressionStr, 64)
	if err != nil {
		return 0, ErrInvalidExpression
	}

	expressionID := s.nextExpressionID
	s.nextExpressionID++

	expression := &models.Expression{
		ID:         expressionID,
		Expression: expressionStr,
		Status:     models.StatusCompleted,
		Result:     &value,
	}
	s.expressions[expressionID] = expression
	log.Printf("Добавлено простое выражение ID=%d: %s, результат=%f", expressionID, expressionStr, value)

	return expressionID, nil
}

func (s *Service) createTasksFromOperations(expressionID int, ops []calculation.Operation) {
	opToTaskMap := make(map[int]int)

	for i, op := range ops {
		taskID := s.nextTaskID
		s.nextTaskID++

		task := &models.Task{
			ID:           taskID,
			Operation:    op.Operator,
			ExpressionID: expressionID,
		}

		opToTaskMap[i+1] = taskID

		if leftID, isLeftTaskID := op.Left.(int); isLeftTaskID {
			actualTaskID := opToTaskMap[leftID]
			leftTaskResultID := getResultID(expressionID, actualTaskID)
			task.Arg1 = leftTaskResultID
		} else {
			task.Arg1 = op.Left
		}

		if rightID, isRightTaskID := op.Right.(int); isRightTaskID {
			actualTaskID := opToTaskMap[rightID]
			rightTaskResultID := getResultID(expressionID, actualTaskID)
			task.Arg2 = rightTaskResultID
		} else {
			task.Arg2 = op.Right
		}

		s.tasks[taskID] = task
		s.pendingTasks = append(s.pendingTasks, task)

		log.Printf("Создана задача ID=%d (%s) для выражения ID=%d",
			taskID, task.Operation, expressionID)

		if i == len(ops)-1 {
			rootID := getRootResultID(expressionID)
			s.results[rootID] = &ExpressionResult{
				Value:     0,
				Completed: false,
			}
		}
	}
}

func (s *Service) checkExpressionCompletion(expressionID int) {
	expr, exists := s.expressions[expressionID]
	if !exists {
		return
	}

	var lastTaskID int
	for taskID, task := range s.tasks {
		if task.ExpressionID == expressionID && taskID > lastTaskID {
			lastTaskID = taskID
		}
	}

	if lastTaskID > 0 {
		resultID := getResultID(expressionID, lastTaskID)
		result, exists := s.results[resultID]

		if exists && result.Completed {
			expr.Status = models.StatusCompleted
			expr.Result = &result.Value
			log.Printf("Выражение ID=%d завершено успешно, результат=%f",
				expressionID, result.Value)
			return
		}
	}

	totalTasks := 0
	completedTasks := 0

	for _, task := range s.tasks {
		if task.ExpressionID == expressionID {
			totalTasks++
			resultID := getResultID(expressionID, task.ID)
			result, exists := s.results[resultID]
			if exists && result.Completed {
				completedTasks++
			}
		}
	}

	if totalTasks > 0 && completedTasks == totalTasks && expr.Status != models.StatusCompleted {
		expr.Status = models.StatusFailed
		log.Printf("Выражение ID=%d помечено как FAILED: все задачи выполнены, но нет финального результата",
			expressionID)
	} else if completedTasks < totalTasks {
		expr.Status = models.StatusProcessing
		log.Printf("Выражение ID=%d в процессе: выполнено %d/%d задач",
			expressionID, completedTasks, totalTasks)
	}
}

func getResultID(expressionID, taskID int) string {
	return fmt.Sprintf("expr_%d_task_%d", expressionID, taskID)
}

func getRootResultID(expressionID int) string {
	return fmt.Sprintf("expr_%d_root", expressionID)
}

func (s *Service) SetTaskError(id int, errorMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return ErrTaskNotFound
	}

	resultID := getResultID(task.ExpressionID, id)
	s.results[resultID] = &ExpressionResult{
		Value:     0,
		Completed: true,
	}

	log.Printf("Задача ID=%d завершилась с ошибкой: %s", id, errorMsg)

	expr, exists := s.expressions[task.ExpressionID]
	if exists {
		expr.Status = models.StatusFailed
	}

	return nil
}
