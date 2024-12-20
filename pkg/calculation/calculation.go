package calculation

import (
	"fmt"
	"strconv"
)

func Calc(expression string) (float64, error) {
	if len(expression) == 0 {
		return 0, fmt.Errorf("error")
	}

	for i := 0; i < len(expression); i++ {
		if expression[i] == '(' {
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
				return 0, fmt.Errorf("error")
			}
			innerResult, err := Calc(expression[i+1 : j-1])
			if err != nil {
				return 0, err
			}
			expression = expression[:i] + strconv.FormatFloat(innerResult, 'f', -1, 64) + expression[j:]
			i--
		}
	}

	numbers := make([]float64, 0)
	operations := make([]string, 0)
	currentNum := ""
	lastWasOp := true

	for i := 0; i < len(expression); i++ {
		char := string(expression[i])

		if expression[i] == ' ' {
			continue
		}

		if (expression[i] >= '0' && expression[i] <= '9') || expression[i] == '.' ||
			(expression[i] == '-' && lastWasOp && i+1 < len(expression) &&
				(expression[i+1] >= '0' && expression[i+1] <= '9')) {
			currentNum += char
			lastWasOp = false
			continue
		}

		if currentNum != "" {
			num, err := strconv.ParseFloat(currentNum, 64)
			if err != nil {
				return 0, fmt.Errorf("error: %s", currentNum)
			}
			numbers = append(numbers, num)
			currentNum = ""
		}

		if char == "+" || char == "-" || char == "*" || char == "/" {
			if lastWasOp && char != "-" {
				return 0, fmt.Errorf("error")
			}
			operations = append(operations, char)
			lastWasOp = true
		} else if char != " " && char != "(" && char != ")" {
			return 0, fmt.Errorf("error: %s", char)
		}
	}

	if currentNum != "" {
		num, err := strconv.ParseFloat(currentNum, 64)
		if err != nil {
			return 0, fmt.Errorf("error: %s", currentNum)
		}
		numbers = append(numbers, num)
	}

	if len(numbers) == 0 {
		return 0, fmt.Errorf("error")
	}
	if len(numbers) != len(operations)+1 {
		return 0, fmt.Errorf("error")
	}

	for i := 0; i < len(operations); i++ {
		if operations[i] == "*" || operations[i] == "/" {
			if operations[i] == "*" {
				numbers[i] = numbers[i] * numbers[i+1]
			} else {
				if numbers[i+1] == 0 {
					return 0, fmt.Errorf("division by zero")
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
		} else if operations[i] == "-" {
			result -= numbers[i+1]
		}
	}

	return result, nil
}
