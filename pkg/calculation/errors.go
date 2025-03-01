package calculation

import "errors"

var (
	ErrInvalidExpression    = errors.New("invalid expression")
	ErrDivisionByZero       = errors.New("division by zero")
	ErrInvalidNumber        = errors.New("invalid number")
	ErrConsecutiveOperators = errors.New("consecutive operators")
	ErrMismatchedBrackets   = errors.New("mismatched parentheses")
	ErrInvalidCharacter     = errors.New("invalid character")
	ErrUnsupportedExpr      = errors.New("unsupported expression type")
)
