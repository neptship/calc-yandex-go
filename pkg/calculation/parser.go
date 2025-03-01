package calculation

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

type Operation struct {
	Left     interface{}
	Right    interface{}
	Operator string
}

func ParseExpression(expr string) ([]Operation, error) {
	exprAST, err := parser.ParseExpr(expr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidExpression, err)
	}

	var operations []Operation
	nextResultID := 1

	resultID, err := buildOperations(exprAST, &operations, &nextResultID)
	if err != nil {
		return nil, err
	}

	if resultID == 0 {
		return nil, ErrInvalidExpression
	}

	return operations, nil
}

func buildOperations(node ast.Expr, operations *[]Operation, nextResultID *int) (interface{}, error) {
	switch n := node.(type) {
	case *ast.BinaryExpr:
		left, err := buildOperations(n.X, operations, nextResultID)
		if err != nil {
			return 0, err
		}

		right, err := buildOperations(n.Y, operations, nextResultID)
		if err != nil {
			return 0, err
		}

		op := ""
		switch n.Op {
		case token.ADD:
			op = "+"
		case token.SUB:
			op = "-"
		case token.MUL:
			op = "*"
		case token.QUO:
			op = "/"
		default:
			return 0, fmt.Errorf("unsupported operator: %v", n.Op)
		}

		operation := Operation{
			Left:     left,
			Right:    right,
			Operator: op,
		}

		*operations = append(*operations, operation)

		resultID := *nextResultID
		*nextResultID++

		return resultID, nil

	case *ast.BasicLit:
		if n.Kind != token.INT && n.Kind != token.FLOAT {
			return 0, fmt.Errorf("unsupported literal type: %v", n.Kind)
		}

		value, err := strconv.ParseFloat(n.Value, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number: %v", n.Value)
		}

		return value, nil

	case *ast.ParenExpr:
		return buildOperations(n.X, operations, nextResultID)

	case *ast.UnaryExpr:
		if n.Op != token.SUB {
			return 0, fmt.Errorf("unsupported unary operator: %v", n.Op)
		}

		operand, err := buildOperations(n.X, operations, nextResultID)
		if err != nil {
			return 0, err
		}

		if value, ok := operand.(float64); ok {
			return -value, nil
		}

		operation := Operation{
			Left:     -1.0,
			Right:    operand,
			Operator: "*",
		}

		*operations = append(*operations, operation)

		resultID := *nextResultID
		*nextResultID++

		return resultID, nil

	default:
		return 0, ErrUnsupportedExpr
	}
}

func EvaluateOperation(left, right float64, op string) (float64, error) {
	switch op {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	case "*":
		return left * right, nil
	case "/":
		if right == 0 {
			return 0, ErrDivisionByZero
		}
		return left / right, nil
	default:
		return 0, fmt.Errorf("unknown operator: %s", op)
	}
}
