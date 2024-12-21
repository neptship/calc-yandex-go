package calculation

import (
	"strconv"
)

func Calc(expression string) (float64, error) {
	var numbers []float64
	var operations []string
	var currentNum string
	var lastWasOp bool

	for i := 0; i < len(expression); i++ {
		char := expression[i]

		if char == '(' {
			brackets := 1
			j := i + 1
			for j < len(expression) && brackets > 0 {
				if expression[j] == '(' {
					brackets++
				}
				if expression[j] == ')' {
					brackets--
				}
				j++
			}
			if brackets > 0 {
				return 0, ErrMismatchedBrackets
			}
			innerResult, err := Calc(expression[i+1 : j-1])
			if err != nil {
				return 0, err
			}
			expression = expression[:i] + strconv.FormatFloat(innerResult, 'f', -1, 64) + expression[j:]
			i--
			continue
		}

		if (char >= '0' && char <= '9') || char == '.' || (char == '-' && (i == 0 || lastWasOp)) {
			currentNum += string(char)
			lastWasOp = false
			continue
		}

		if currentNum != "" {
			num, err := strconv.ParseFloat(currentNum, 64)
			if err != nil {
				return 0, ErrInvalidNumber
			}
			numbers = append(numbers, num)
			currentNum = ""
		}

		if char == '+' || char == '-' || char == '*' || char == '/' {
			if lastWasOp && char != '-' {
				return 0, ErrConsecutiveOperators
			}
			operations = append(operations, string(char))
			lastWasOp = true
		} else {
			return 0, ErrInvalidCharacter
		}
	}

	if currentNum != "" {
		num, err := strconv.ParseFloat(currentNum, 64)
		if err != nil {
			return 0, ErrInvalidNumber
		}
		numbers = append(numbers, num)
	}

	if len(numbers) == 0 {
		return 0, ErrInvalidExpression
	}
	if len(numbers) != len(operations)+1 {
		return 0, ErrInvalidExpression
	}

	for i := 0; i < len(operations); i++ {
		if operations[i] == "*" || operations[i] == "/" {
			if operations[i] == "*" {
				numbers[i] = numbers[i] * numbers[i+1]
			} else {
				if numbers[i+1] == 0 {
					return 0, ErrDivisionByZero
				}
				numbers[i] = numbers[i] / numbers[i+1]
			}
			numbers = append(numbers[:i+1], numbers[i+2:]...)
			operations = append(operations[:i], operations[i+1:]...)
			i--
		}
	}

	result := numbers[0]
	for i := 0; i < len(operations); i++ {
		if operations[i] == "+" {
			result += numbers[i+1]
		} else {
			result -= numbers[i+1]
		}
	}

	return result, nil
}
