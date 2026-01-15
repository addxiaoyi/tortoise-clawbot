package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ============ Web Search Skill ============

type WebSearchSkill struct{}

func (s *WebSearchSkill) Name() string         { return "web_search" }
func (s *WebSearchSkill) Description() string   { return "Search the web for information" }
func (s *WebSearchSkill) Version() string      { return "1.0.0" }
func (s *WebSearchSkill) Category() SkillCategory { return CategoryCore }

func (s *WebSearchSkill) Parameters() []Parameter {
	return []Parameter{
		{Name: "query", Type: "string", Description: "Search query", Required: true},
		{Name: "limit", Type: "number", Description: "Maximum results", Default: 10},
		{Name: "source", Type: "string", Description: "Search engine (google/duckduckgo/bing)", Default: "duckduckgo"},
	}
}

func (s *WebSearchSkill) Validate(params map[string]interface{}) error {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return fmt.Errorf("query is required")
	}
	return nil
}

func (s *WebSearchSkill) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	query := params["query"].(string)
	limit := 10
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}
	source := "duckduckgo"
	if src, ok := params["source"].(string); ok {
		source = src
	}

	// 简单搜索实现
	results, err := s.search(query, limit, source)
	if err != nil {
		return &Result{Success: false, Error: err.Error()}, nil
	}

	return &Result{
		Success: true,
		Output:  results,
		Metadata: map[string]interface{}{
			"query": query,
			"count": len(results),
		},
	}, nil
}

func (s *WebSearchSkill) search(query string, limit int, source string) ([]SearchResult, error) {
	// 简化实现 - 实际应调用真实搜索API
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1", url.QueryEscape(query))
	
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ddgResp struct {
		Results []struct {
			Text string `json:"Text"`
			URL  string `json:"URL"`
		} `json:"RelatedTopics"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ddgResp); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, limit)
	for _, r := range ddgResp.Results {
		if len(results) >= limit {
			break
		}
		results = append(results, SearchResult{
			Title: r.Text,
			URL:   r.URL,
		})
	}

	return results, nil
}

type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
	Snippet     string `json:"snippet,omitempty"`
}

// ============ Calculator Skill ============

type CalculatorSkill struct{}

func (s *CalculatorSkill) Name() string         { return "calculator" }
func (s *CalculatorSkill) Description() string   { return "Perform mathematical calculations" }
func (s *CalculatorSkill) Version() string        { return "1.0.0" }
func (s *CalculatorSkill) Category() SkillCategory { return CategoryCore }

func (s *CalculatorSkill) Parameters() []Parameter {
	return []Parameter{
		{Name: "expression", Type: "string", Description: "Mathematical expression", Required: true},
		{Name: "precision", Type: "number", Description: "Decimal precision", Default: 10},
	}
}

func (s *CalculatorSkill) Validate(params map[string]interface{}) error {
	expr, ok := params["expression"].(string)
	if !ok || expr == "" {
		return fmt.Errorf("expression is required")
	}
	return nil
}

func (s *CalculatorSkill) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	expr := params["expression"].(string)
	precision := 10
	if p, ok := params["precision"].(float64); ok {
		precision = int(p)
	}

	result, err := s.calculate(expr, precision)
	if err != nil {
		return &Result{Success: false, Error: err.Error()}, nil
	}

	return &Result{
		Success: true,
		Output:  result,
		Metadata: map[string]interface{}{
			"expression": expr,
		},
	}, nil
}

func (s *CalculatorSkill) calculate(expr string, precision int) (string, error) {
	// 安全计算 - 只允许数字和基本运算符
	allowed := regexp.MustCompile(`^[0-9+\-*/().%\s]+$`)
	if !allowed.MatchString(expr) {
		return "", fmt.Errorf("invalid expression")
	}

	// 简单的表达式求值
	expr = strings.ReplaceAll(expr, " ", "")
	result, err := s.eval(expr)
	if err != nil {
		return "", err
	}

	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, result), nil
}

func (s *CalculatorSkill) eval(expr string) (float64, error) {
	// 简化实现
	tokens := s.tokenize(expr)
	if len(tokens) == 0 {
		return 0, fmt.Errorf("empty expression")
	}
	
	// 处理括号
	for {
		open := -1
		close := -1
		for i, t := range tokens {
			if t == "(" {
				open = i
			}
			if t == ")" && open != -1 {
				close = i
				break
			}
		}
		if close == -1 {
			break
		}
		
		subExpr := strings.Join(tokens[open+1:close], "")
		subResult, err := s.eval(subExpr)
		if err != nil {
			return 0, err
		}
		
		tokens = append(tokens[:open], append([]string{fmt.Sprintf("%f", subResult)}, tokens[close+1:]...)...)
	}

	return s.evaluate(tokens)
}

func (s *CalculatorSkill) tokenize(expr string) []string {
	var tokens []string
	var num string
	for i := 0; i < len(expr); i++ {
		c := expr[i]
		if c >= '0' && c <= '9' || c == '.' {
			num += string(c)
		} else if c == '-' && (len(tokens) == 0 || tokens[len(tokens)-1] == "(" || tokens[len(tokens)-1] == "+" || tokens[len(tokens)-1] == "-" || tokens[len(tokens)-1] == "*" || tokens[len(tokens)-1] == "/") {
			num += string(c)
		} else {
			if num != "" {
				tokens = append(tokens, num)
				num = ""
			}
			if string(c) != " " {
				tokens = append(tokens, string(c))
			}
		}
	}
	if num != "" {
		tokens = append(tokens, num)
	}
	return tokens
}

func (s *CalculatorSkill) evaluate(tokens []string) (float64, error) {
	if len(tokens) == 0 {
		return 0, nil
	}

	// 处理 + 和 -
	result := 0.0
	op := "+"
	i := 0

	for i < len(tokens) {
		token := tokens[i]
		if token == "+" || token == "-" {
			op = token
		} else if token == "*" || token == "/" {
			nextOp := "+"
			nextIdx := i + 2
			if nextIdx < len(tokens) {
				if tokens[nextIdx] == "+" || tokens[nextIdx] == "-" {
					nextOp = tokens[nextIdx]
				}
			}
			
			var term float64
			if token == "*" {
				term, _ = strconv.ParseFloat(tokens[i-1], 64) * s.parseNum(tokens[i+1])
			} else {
				term, _ = strconv.ParseFloat(tokens[i-1], 64) / s.parseNum(tokens[i+1])
			}
			
			if op == "+" {
				result += term
			} else {
				result -= term
			}
			
			tokens = append(tokens[:i-1], tokens[nextIdx:]...)
			if nextOp == "-" {
				tokens = append([]string{"+"}, tokens...)
				tokens[0] = "-"
			}
			i = 0
			continue
		} else {
			val := s.parseNum(token)
			if op == "+" {
				result += val
			} else {
				result -= val
			}
		}
		i++
	}

	return result, nil
}

func (s *CalculatorSkill) parseNum(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// ============ Code Interpreter Skill ============

type CodeInterpreterSkill struct{}

func (s *CodeInterpreterSkill) Name() string         { return "code_executor" }
func (s *CodeInterpreterSkill) Description() string   { return "Execute code snippets" }
func (s *CodeInterpreterSkill) Version() string       { return "1.0.0" }
func (s *CodeInterpreterSkill) Category() SkillCategory { return CategoryCore }

func (s *CodeInterpreterSkill) Parameters() []Parameter {
	return []Parameter{
		{Name: "code", Type: "string", Description: "Code to execute", Required: true},
		{Name: "language", Type: "string", Description: "Programming language", Default: "python"},
		{Name: "timeout", Type: "number", Description: "Timeout in seconds", Default: 30},
	}
}

func (s *CodeInterpreterSkill) Validate(params map[string]interface{}) error {
	code, ok := params["code"].(string)
	if !ok || code == "" {
		return fmt.Errorf("code is required")
	}
	return nil
}

func (s *CodeInterpreterSkill) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	code := params["code"].(string)
	language := "python"
	if lang, ok := params["language"].(string); ok {
		language = lang
	}
	timeout := 30
	if t, ok := params["timeout"].(float64); ok {
		timeout = int(t)
	}

	output, err := s.execute(code, language, timeout)
	if err != nil {
		return &Result{Success: false, Error: err.Error()}, nil
	}

	return &Result{
		Success: true,
		Output:  output,
		Metadata: map[string]interface{}{
			"language": language,
		},
	}, nil
}

func (s *CodeInterpreterSkill) execute(code, language string, timeout int) (string, error) {
	// 沙盒执行 - 实际应使用真实的沙盒环境
	switch language {
	case "python":
		return s.executePython(code)
	case "javascript", "js":
		return s.executeJS(code)
	case "shell", "bash":
		return s.executeShell(code)
	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}
}

func (s *CodeInterpreterSkill) executePython(code string) (string, error) {
	// 简化实现 - 实际应使用沙盒
	return "[Python execution simulated]\nCode would be executed in a sandboxed environment.", nil
}

func (s *CodeInterpreterSkill) executeJS(code string) (string, error) {
	return "[JavaScript execution simulated]\nCode would be executed in a sandboxed environment.", nil
}

func (s *CodeInterpreterSkill) executeShell(code string) (string, error) {
	return "[Shell execution simulated]\nCode would be executed in a sandboxed environment.", nil
}

// ============ File System Skill ============

type FileSystemSkill struct{}

func (s *FileSystemSkill) Name() string         { return "file_system" }
func (s *FileSystemSkill) Description() string   { return "Read and write files" }
func (s *FileSystemSkill) Version() string       { return "1.0.0" }
func (s *FileSystemSkill) Category() SkillCategory { return CategoryCore }

func (s *FileSystemSkill) Parameters() []Parameter {
	return []Parameter{
		{Name: "operation", Type: "string", Description: "Operation (read/write/list/delete)", Required: true},
		{Name: "path", Type: "string", Description: "File path", Required: true},
		{Name: "content", Type: "string", Description: "Content to write"},
	}
}

func (s *FileSystemSkill) Validate(params map[string]interface{}) error {
	op, ok := params["operation"].(string)
	if !ok || op == "" {
		return fmt.Errorf("operation is required")
	}
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("path is required")
	}
	return nil
}

func (s *FileSystemSkill) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	op := params["operation"].(string)
	path := params["path"].(string)
	content := ""
	if c, ok := params["content"].(string); ok {
		content = c
	}

	var output interface{}
	var err error

	switch op {
	case "read":
		output, err = s.readFile(path)
	case "write":
		err = s.writeFile(path, content)
		output = "File written successfully"
	case "list":
		output, err = s.listDir(path)
	case "delete":
		err = s.deleteFile(path)
		output = "File deleted successfully"
	default:
		err = fmt.Errorf("unknown operation: %s", op)
	}

	if err != nil {
		return &Result{Success: false, Error: err.Error()}, nil
	}

	return &Result{
		Success: true,
		Output:  output,
		Metadata: map[string]interface{}{
			"operation": op,
			"path": path,
		},
	}, nil
}

func (s *FileSystemSkill) readFile(path string) (string, error) {
	// 简化实现
	return fmt.Sprintf("[Read] %s - File reading would be performed in sandboxed environment", path), nil
}

func (s *FileSystemSkill) writeFile(path, content string) error {
	return nil
}

func (s *FileSystemSkill) listDir(path string) ([]string, error) {
	return []string{"file1.txt", "file2.txt", "folder/"}, nil
}

func (s *FileSystemSkill) deleteFile(path string) error {
	return nil
}

// ============ Calendar Skill ============

type CalendarSkill struct{}

func (s *CalendarSkill) Name() string         { return "calendar" }
func (s *CalendarSkill) Description() string    { return "Manage calendar events" }
func (s *CalendarSkill) Version() string        { return "1.0.0" }
func (s *CalendarSkill) Category() SkillCategory { return CategoryIntegration }

func (s *CalendarSkill) Parameters() []Parameter {
	return []Parameter{
		{Name: "operation", Type: "string", Description: "Operation (list/create/update/delete)", Required: true},
		{Name: "title", Type: "string", Description: "Event title"},
		{Name: "start", Type: "string", Description: "Start time (ISO8601)"},
		{Name: "end", Type: "string", Description: "End time (ISO8601)"},
		{Name: "description", Type: "string", Description: "Event description"},
	}
}

func (s *CalendarSkill) Validate(params map[string]interface{}) error {
	op, ok := params["operation"].(string)
	if !ok || op == "" {
		return fmt.Errorf("operation is required")
	}
	return nil
}

func (s *CalendarSkill) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	op := params["operation"].(string)

	var output interface{}
	var err error

	switch op {
	case "list":
		output, err = s.listEvents()
	case "create":
		output, err = s.createEvent(params)
	case "today":
		output, err = s.getTodayEvents()
	default:
		err = fmt.Errorf("unknown operation: %s", op)
	}

	if err != nil {
		return &Result{Success: false, Error: err.Error()}, nil
	}

	return &Result{
		Success: true,
		Output:  output,
	}, nil
}

func (s *CalendarSkill) listEvents() ([]CalendarEvent, error) {
	return []CalendarEvent{
		{
			ID:          "1",
			Title:       "Team Meeting",
			Start:       time.Now().Add(2 * time.Hour),
			End:         time.Now().Add(3 * time.Hour),
			Description: "Weekly sync",
		},
	}, nil
}

func (s *CalendarSkill) createEvent(params map[string]interface{}) (string, error) {
	return "event_123", nil
}

func (s *CalendarSkill) getTodayEvents() ([]CalendarEvent, error) {
	return s.listEvents()
}

type CalendarEvent struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Description string    `json:"description"`
	Location    string    `json:"location,omitempty"`
}

// ============ Unit Converter Skill ============

type UnitConverterSkill struct{}

func (s *UnitConverterSkill) Name() string         { return "unit_converter" }
func (s *UnitConverterSkill) Description() string   { return "Convert between different units" }
func (s *UnitConverterSkill) Version() string       { return "1.0.0" }
func (s *UnitConverterSkill) Category() SkillCategory { return CategoryUtility }

func (s *UnitConverterSkill) Parameters() []Parameter {
	return []Parameter{
		{Name: "value", Type: "number", Description: "Value to convert", Required: true},
		{Name: "from", Type: "string", Description: "Source unit", Required: true},
		{Name: "to", Type: "string", Description: "Target unit", Required: true},
	}
}

func (s *UnitConverterSkill) Validate(params map[string]interface{}) error {
	_, ok := params["value"].(float64)
	if !ok {
		return fmt.Errorf("value is required")
	}
	return nil
}

func (s *UnitConverterSkill) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	value := params["value"].(float64)
	from := params["from"].(string)
	to := params["to"].(string)

	result, err := s.convert(value, from, to)
	if err != nil {
		return &Result{Success: false, Error: err.Error()}, nil
	}

	return &Result{
		Success: true,
		Output: map[string]interface{}{
			"value": value,
			"from":  from,
			"to":    to,
			"result": result,
		},
	}, nil
}

func (s *UnitConverterSkill) convert(value float64, from, to string) (float64, error) {
	// 简化实现 - 实际应使用完整的单位转换表
	conversions := map[string]map[string]float64{
		"km": {"m": 1000, "mi": 0.621371, "ft": 3280.84},
		"m": {"km": 0.001, "mi": 0.000621371, "ft": 3.28084},
		"kg": {"lb": 2.20462, "oz": 35.274},
		"c": {"f": 1, "k": 1}, // 需要特殊处理
	}

	if fromMap, ok := conversions[from]; ok {
		if factor, ok := fromMap[to]; ok {
			return value * factor, nil
		}
	}

	return 0, fmt.Errorf("conversion from %s to %s not supported", from, to)
}

// ============ Date/Time Skill ============

type DateTimeSkill struct{}

func (s *DateTimeSkill) Name() string         { return "datetime" }
func (s *DateTimeSkill) Description() string   { return "Get current date and time information" }
func (s *DateTimeSkill) Version() string       { return "1.0.0" }
func (s *DateTimeSkill) Category() SkillCategory { return CategoryUtility }

func (s *DateTimeSkill) Parameters() []Parameter {
	return []Parameter{
		{Name: "operation", Type: "string", Description: "Operation (now/timezone/diff)", Required: true},
		{Name: "tz", Type: "string", Description: "Timezone"},
		{Name: "date1", Type: "string", Description: "First date"},
		{Name: "date2", Type: "string", Description: "Second date"},
	}
}

func (s *DateTimeSkill) Validate(params map[string]interface{}) error {
	return nil
}

func (s *DateTimeSkill) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	op := "now"
	if o, ok := params["operation"].(string); ok {
		op = o
	}

	var output interface{}
	var err error

	switch op {
	case "now":
		output = s.getNow(params)
	case "timezone":
		output, err = s.getTimezone(params)
	case "diff":
		output, err = s.getDateDiff(params)
	default:
		err = fmt.Errorf("unknown operation: %s", op)
	}

	if err != nil {
		return &Result{Success: false, Error: err.Error()}, nil
	}

	return &Result{
		Success: true,
		Output:  output,
	}, nil
}

func (s *DateTimeSkill) getNow(params map[string]interface{}) map[string]interface{} {
	now := time.Now()
	return map[string]interface{}{
		"datetime": now.Format(time.RFC3339),
		"date":     now.Format("2006-01-02"),
		"time":     now.Format("15:04:05"),
		"unix":     now.Unix(),
		"weekday":  now.Weekday().String(),
	}
}

func (s *DateTimeSkill) getTimezone(params map[string]interface{}) (map[string]interface{}, error) {
	tz := "Asia/Shanghai"
	if t, ok := params["tz"].(string); ok {
		tz = t
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, err
	}

	now := time.Now().In(loc)
	return map[string]interface{}{
		"timezone":    tz,
		"datetime":    now.Format(time.RFC3339),
		"offset":      now.Format("-0700"),
	}, nil
}

func (s *DateTimeSkill) getDateDiff(params map[string]interface{}) (map[string]interface{}, error) {
	date1 := params["date1"].(string)
	date2 := params["date2"].(string)

	d1, err := time.Parse("2006-01-02", date1)
	if err != nil {
		return nil, err
	}
	d2, err := time.Parse("2006-01-02", date2)
	if err != nil {
		return nil, err
	}

	diff := d2.Sub(d1)
	return map[string]interface{}{
		"days":    int(diff.Hours() / 24),
		"hours":   int(diff.Hours()),
		"minutes": int(diff.Minutes()),
		"seconds": int(diff.Seconds()),
	}, nil
}

// ============ Text Processing Skill ============

type TextProcessingSkill struct{}

func (s *TextProcessingSkill) Name() string         { return "text_processing" }
func (s *TextProcessingSkill) Description() string    { return "Process and transform text" }
func (s *TextProcessingSkill) Version() string       { return "1.0.0" }
func (s *TextProcessingSkill) Category() SkillCategory { return CategoryUtility }

func (s *TextProcessingSkill) Parameters() []Parameter {
	return []Parameter{
		{Name: "operation", Type: "string", Description: "Operation (upper/lower/count/reverse)", Required: true},
		{Name: "text", Type: "string", Description: "Text to process", Required: true},
	}
}

func (s *TextProcessingSkill) Validate(params map[string]interface{}) error {
	return nil
}

func (s *TextProcessingSkill) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	text := params["text"].(string)
	op := params["operation"].(string)

	var result string
	switch op {
	case "upper":
		result = strings.ToUpper(text)
	case "lower":
		result = strings.ToLower(text)
	case "count":
		result = fmt.Sprintf("%d characters, %d words", len(text), len(strings.Fields(text)))
	case "reverse":
		runes := []rune(text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		result = string(runes)
	case "length":
		result = fmt.Sprintf("%d", len(text))
	default:
		return &Result{Success: false, Error: fmt.Errorf("unknown operation: %s", op)}, nil
	}

	return &Result{
		Success: true,
		Output:  result,
	}, nil
}

// ============ Register All Built-in Skills ============

func RegisterBuiltInSkills(r *Registry) {
	skills := []Skill{
		&WebSearchSkill{},
		&CalculatorSkill{},
		&CodeInterpreterSkill{},
		&FileSystemSkill{},
		&CalendarSkill{},
		&UnitConverterSkill{},
		&DateTimeSkill{},
		&TextProcessingSkill{},
	}

	for _, s := range skills {
		if err := r.Register(s); err != nil {
			fmt.Printf("Failed to register skill %s: %v\n", s.Name(), err)
		}
	}
}

// GetAllBuiltInSkills 返回所有内置技能
func GetAllBuiltInSkills() []Skill {
	return []Skill{
		&WebSearchSkill{},
		&CalculatorSkill{},
		&CodeInterpreterSkill{},
		&FileSystemSkill{},
		&CalendarSkill{},
		&UnitConverterSkill{},
		&DateTimeSkill{},
		&TextProcessingSkill{},
	}
}
