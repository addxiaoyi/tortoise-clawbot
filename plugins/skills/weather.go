// Weather 技能实现
package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// WeatherSkill 天气技能
type WeatherSkill struct {
	apiKey string
}

// NewWeatherSkill 创建天气技能
func NewWeatherSkill(apiKey string) *WeatherSkill {
	return &WeatherSkill{apiKey: apiKey}
}

// Execute 执行技能
func (s *WeatherSkill) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Action string `json:"action"`
		City   string `json:"city,omitempty"`
		Lat    float64 `json:"lat,omitempty"`
		Lon    float64 `json:"lon,omitempty"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	switch params.Action {
	case "current":
		if params.City != "" {
			return s.getCurrentByCity(params.City)
		}
		return s.getCurrentByCoords(params.Lat, params.Lon)
	case "forecast":
		if params.City != "" {
			return s.getForecastByCity(params.City)
		}
		return s.getForecastByCoords(params.Lat, params.Lon)
	default:
		return json.Marshal(map[string]interface{}{
			"error": "Unknown action: " + params.Action,
		})
	}
}

func (s *WeatherSkill) getCurrentByCity(city string) (json.RawMessage, error) {
	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric",
		city, s.apiKey,
	)
	return s.fetchWeather(url)
}

func (s *WeatherSkill) getCurrentByCoords(lat, lon float64) (json.RawMessage, error) {
	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric",
		lat, lon, s.apiKey,
	)
	return s.fetchWeather(url)
}

func (s *WeatherSkill) getForecastByCity(city string) (json.RawMessage, error) {
	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/forecast?q=%s&appid=%s&units=metric",
		city, s.apiKey,
	)
	return s.fetchWeather(url)
}

func (s *WeatherSkill) getForecastByCoords(lat, lon float64) (json.RawMessage, error) {
	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/forecast?lat=%f&lon=%f&appid=%s&units=metric",
		lat, lon, s.apiKey,
	)
	return s.fetchWeather(url)
}

func (s *WeatherSkill) fetchWeather(url string) (json.RawMessage, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return json.Marshal(map[string]interface{}{"error": err.Error()})
	}

	resp, err := client.Do(req)
	if err != nil {
		return json.Marshal(map[string]interface{}{"error": err.Error()})
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return json.Marshal(map[string]interface{}{"error": err.Error()})
	}

	// 解析并格式化响应
	var weather map[string]interface{}
	if err := json.Unmarshal(data, &weather); err != nil {
		return json.Marshal(map[string]interface{}{"error": err.Error()})
	}

	// 格式化输出
	result := s.formatWeather(weather)
	return json.Marshal(result)
}

func (s *WeatherSkill) formatWeather(data map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"success": true,
	}

	if name, ok := data["name"].(string); ok {
		result["city"] = name
	}

	if main, ok := data["main"].(map[string]interface{}); ok {
		result["temperature"] = main["temp"]
		result["feels_like"] = main["feels_like"]
		result["humidity"] = main["humidity"]
		result["pressure"] = main["pressure"]
	}

	if weather, ok := data["weather"].([]interface{}); ok && len(weather) > 0 {
		if w, ok := weather[0].(map[string]interface{}); ok {
			result["description"] = w["description"]
			result["icon"] = w["icon"]
		}
	}

	if wind, ok := data["wind"].(map[string]interface{}); ok {
		result["wind_speed"] = wind["speed"]
	}

	if clouds, ok := data["clouds"].(map[string]interface{}); ok {
		result["clouds"] = clouds["all"]
	}

	return result
}

// Metadata 返回技能元数据
func (s *WeatherSkill) Metadata() *SkillMetadata {
	return &SkillMetadata{
		Name:        "weather",
		Description: "Get current weather and forecasts",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"description": "Action: current or forecast",
					"enum":        []string{"current", "forecast"},
				},
				"city": map[string]interface{}{
					"type":        "string",
					"description": "City name",
				},
				"lat": map[string]interface{}{
					"type":        "number",
					"description": "Latitude",
				},
				"lon": map[string]interface{}{
					"type":        "number",
					"description": "Longitude",
				},
			},
			"anyOf": []map[string]interface{}{
				{"required": []string{"city"}},
				{"required": []string{"lat", "lon"}},
			},
		},
	}
}
