package orchestrator

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/neptship/calc-yandex-go/internal/config"
	"github.com/neptship/calc-yandex-go/internal/database"
	"github.com/neptship/calc-yandex-go/internal/models"
	"github.com/neptship/calc-yandex-go/pkg/calculation"
)

var (
	ErrExpressionNotFound = errors.New("expression not found")
	ErrTaskNotFound       = errors.New("task not found")
	ErrInvalidExpression  = errors.New("invalid expression")
	ErrInvalidTaskResult  = errors.New("invalid task result")
	ErrUnauthorized       = errors.New("unauthorized")
)

type ExpressionResult struct {
	Value     float64
	Completed bool
}

type Service struct {
	db     *database.Database
	config *config.Config
	mu     sync.Mutex

	tasks        map[int]*models.Task
	pendingTasks []*models.Task
	results      map[string]*ExpressionResult
	expressions  map[int]*models.Expression
	nextTaskID   int
}

func NewService(cfg *config.Config, db *database.Database) *Service {
	return &Service{
		db:           db,
		config:       cfg,
		tasks:        make(map[int]*models.Task),
		pendingTasks: make([]*models.Task, 0),
		results:      make(map[string]*ExpressionResult),
		expressions:  make(map[int]*models.Expression),
		nextTaskID:   1,
	}
}

func (s *Service) AddExpression(userID int, expressionStr string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ops, err := calculation.ParseExpression(expressionStr)
	if err != nil {
		log.Printf("Error parsing expression: %v", err)
		return 0, ErrInvalidExpression
	}

	expressionID, err := s.db.SaveExpression(userID, expressionStr, models.StatusProcessing)
	if err != nil {
		log.Printf("Error saving expression: %v", err)
		return 0, fmt.Errorf("failed to save expression: %w", err)
	}

	log.Printf("Added expression ID=%d for user ID=%d: %s", expressionID, userID, expressionStr)

	err = s.createTasksFromOperations(userID, expressionID, ops)
	if err != nil {
		log.Printf("Error creating tasks for expression ID=%d: %v", expressionID, err)
		return 0, fmt.Errorf("failed to create tasks: %w", err)
	}

	return expressionID, nil
}

func (s *Service) AddSimpleExpression(userID int, expressionStr string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, err := strconv.ParseFloat(expressionStr, 64)
	if err != nil {
		return 0, ErrInvalidExpression
	}

	expressionID, err := s.db.SaveExpression(userID, expressionStr, models.StatusCompleted)
	if err != nil {
		return 0, fmt.Errorf("failed to save expression: %w", err)
	}

	err = s.db.SetExpressionResult(expressionID, value)
	if err != nil {
		return 0, fmt.Errorf("failed to set expression result: %w", err)
	}

	log.Printf("Added simple expression ID=%d for user ID=%d: %s = %f", expressionID, userID, expressionStr, value)
	return expressionID, nil
}

func (s *Service) GetExpressionByID(userID, expressionID int) (*models.Expression, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expr, err := s.db.GetExpression(expressionID)
	if err != nil {
		return nil, ErrExpressionNotFound
	}

	var exprUserID int
	err = s.db.GetDB().QueryRow("SELECT user_id FROM expressions WHERE id = ?", expressionID).Scan(&exprUserID)
	if err != nil {
		return nil, ErrExpressionNotFound
	}

	if exprUserID != userID {
		return nil, ErrUnauthorized
	}

	return expr, nil
}

func (s *Service) GetAllExpressions(userID int) ([]*models.Expression, error) {
	return s.db.GetUserExpressions(userID)
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

			log.Printf("Task ID=%d, operation=%s for expression ID=%d assigned",
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

	err := s.db.SetTaskResult(id, result)
	if err != nil {
		return fmt.Errorf("failed to save task result: %w", err)
	}

	resultID := getResultID(task.ExpressionID, id)
	s.results[resultID] = &ExpressionResult{
		Value:     result,
		Completed: true,
	}

	log.Printf("Received result for task ID=%d: %f", id, result)

	s.checkExpressionCompletion(task.ExpressionID)

	return nil
}

func (s *Service) createTasksFromOperations(userID, expressionID int, ops []calculation.Operation) error {
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

		dbTaskID, err := s.db.SaveTask(task)
		if err != nil {
			return fmt.Errorf("failed to save task to database: %w", err)
		}

		if dbTaskID != taskID {
			log.Printf("Warning: memory taskID %d differs from database taskID %d", taskID, dbTaskID)
		}

		log.Printf("Created task ID=%d (%s) for expression ID=%d",
			taskID, task.Operation, expressionID)

		if i == len(ops)-1 {
			rootID := getRootResultID(expressionID)
			s.results[rootID] = &ExpressionResult{
				Value:     0,
				Completed: false,
			}

			err := s.db.SaveResult(rootID, expressionID, &taskID, 0, false)
			if err != nil {
				return fmt.Errorf("failed to save root result: %w", err)
			}
		}
	}

	return nil
}

func (s *Service) checkExpressionCompletion(expressionID int) {
	expr, exists := s.expressions[expressionID]
	if !exists {
		dbExpr, err := s.db.GetExpression(expressionID)
		if err != nil {
			log.Printf("Error loading expression ID=%d: %v", expressionID, err)
			return
		}
		expr = dbExpr
		s.expressions[expressionID] = expr
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

			err := s.db.SetExpressionResult(expressionID, result.Value)
			if err != nil {
				log.Printf("Error updating expression result in database: %v", err)
			}

			log.Printf("Expression ID=%d completed successfully, result=%f",
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

		err := s.db.UpdateExpressionStatus(expressionID, models.StatusFailed)
		if err != nil {
			log.Printf("Error updating expression status in database: %v", err)
		}

		log.Printf("Expression ID=%d marked as FAILED: all tasks completed but no final result",
			expressionID)
	} else if completedTasks < totalTasks {
		expr.Status = models.StatusProcessing

		err := s.db.UpdateExpressionStatus(expressionID, models.StatusProcessing)
		if err != nil {
			log.Printf("Error updating expression status in database: %v", err)
		}

		log.Printf("Expression ID=%d in progress: %d/%d tasks completed",
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

	err := s.db.SaveResult(resultID, task.ExpressionID, &id, 0, true)
	if err != nil {
		log.Printf("Error saving result to database: %v", err)
	}

	log.Printf("Task ID=%d failed with error: %s", id, errorMsg)

	expr, exists := s.expressions[task.ExpressionID]
	if exists {
		expr.Status = models.StatusFailed

		err := s.db.UpdateExpressionStatus(task.ExpressionID, models.StatusFailed)
		if err != nil {
			log.Printf("Error updating expression status in database: %v", err)
		}
	}

	return nil
}
