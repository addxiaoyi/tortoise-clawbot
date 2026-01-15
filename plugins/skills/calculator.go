// Calculator 技能实现
package skills

import (
	"context"
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// CalculatorSkill 计算器技能
type CalculatorSkill struct{}

// NewCalculatorSkill 创建计算器技能
func NewCalculatorSkill() *CalculatorSkill {
	return &CalculatorSkill{}
}

// Execute 执行技能
func (s *CalculatorSkill) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Expression string `json:"expression"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	result, err := s.calculate(params.Expression)
	if err != nil {
		return json.Marshal(map[string]interface{}{
			"error": err.Error(),
		})
	}

	return json.Marshal(map[string]interface{}{
		"expression": params.Expression,
		"result":      result,
	})
}

// calculate 计算表达式
func (s *CalculatorSkill) calculate(expr string) (float64, error) {
	// 移除空格
	expr = strings.ReplaceAll(expr, " ", "")

	// 基本验证
	if len(expr) == 0 {
		return 0, nil
	}

	// 使用简单评估器
	return s.evaluate(expr)
}

// evaluate 表达式求值
func (s *CalculatorSkill) evaluate(expr string) (float64, error) {
	// 处理括号
	for {
		start := strings.LastIndex(expr, "(")
		if start == -1 {
			break
		}
		end := strings.Index(expr[start:], ")")
		if end == -1 {
			return 0, &CalcError{"Mismatched parentheses"}
		}
		result, err := s.evaluate(expr[start+1 : start+end])
		if err != nil {
			return 0, err
		}
		expr = expr[:start] + strconv.FormatFloat(result, 'f', 10, 64) + expr[start+end+1:]
	}

	// 处理加减乘除
	tokens := s.tokenize(expr)
	if len(tokens) == 0 {
		return 0, &CalcError{"Empty expression"}
	}

	return s.evalTokens(tokens)
}

// tokenize 分词
func (s *CalculatorSkill) tokenize(expr string) []string {
	var tokens []string
	var num strings.Builder

	for i := 0; i < len(expr); i++ {
		c := expr[i]
		if (c >= '0' && c <= '9') || c == '.' {
			num.WriteByte(c)
		} else if c == '-' && (len(tokens) == 0 || tokens[len(tokens)-1] == "(" || s.isOperator(tokens[len(tokens)-1])) {
			// 负数
			num.WriteByte(c)
		} else if s.isOperator(string(c)) {
			if num.Len() > 0 {
				tokens = append(tokens, num.String())
				num.Reset()
			}
			tokens = append(tokens, string(c))
		}
	}

	if num.Len() > 0 {
		tokens = append(tokens, num.String())
	}

	return tokens
}

// isOperator 检查是否为运算符
func (s *CalculatorSkill) isOperator(token string) bool {
	return token == "+" || token == "-" || token == "*" || token == "/" || token == "^"
}

// evalTokens 求值
func (s *CalculatorSkill) evalTokens(tokens []string) (float64, error) {
	if len(tokens) == 0 {
		return 0, nil
	}

	// 处理加减
	var result float64
	var pendingOp string

	for i, token := range tokens {
		if s.isOperator(token) {
			pendingOp = token
			continue
		}

		num, err := strconv.ParseFloat(token, 64)
		if err != nil {
			return 0, &CalcError{"Invalid number: " + token}
		}

		if i == 0 {
			result = num
		} else {
			switch pendingOp {
			case "+":
				result += num
			case "-":
				result -= num
			case "*":
				result *= num
			case "/":
				if num == 0 {
					return 0, &CalcError{"Division by zero"}
				}
				result /= num
			case "^":
				result = math.Pow(result, num)
			}
		}
	}

	return result, nil
}

// Metadata 返回技能元数据
func (s *CalculatorSkill) Metadata() *SkillMetadata {
	return &SkillMetadata{
		Name:        "calculator",
		Description: "Perform mathematical calculations",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"expression": map[string]interface{}{
					"type":        "string",
					"description": "The mathematical expression to evaluate",
				},
			},
			"required": []string{"expression"},
		},
	}
}

// CalcError 计算错误
type CalcError struct {
	Msg string
}

func (e *CalcError) Error() string {
	return e.Msg
}
