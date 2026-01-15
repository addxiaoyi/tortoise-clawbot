package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WeatherPlugin 天气插件
type WeatherPlugin struct {
	config PluginConfig
	client *http.Client
}

// PluginConfig 插件配置
type PluginConfig struct {
	APIKey string `json:"apiKey"`
	Units  string `json:"units"`
}

// WeatherResponse 天气响应
type WeatherResponse struct {
	Main    MainData    `json:"main"`
	Weather []Weather   `json:"weather"`
	Wind    WindData    `json:"wind"`
	Name    string      `json:"name"`
}

// MainData 主要数据
type MainData struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	Humidity  int     `json:"humidity"`
}

// Weather 天气状况
type Weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// WindData 风数据
type WindData struct {
	Speed float64 `json:"speed"`
	Deg   float64 `json:"deg"`
}

// FunctionCall 函数调用请求
type FunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// Plugin 插件接口
var Plugin *WeatherPlugin

// Init 初始化插件
func Init(config map[string]interface{}) error {
	apiKey, ok := config["apiKey"].(string)
	if !ok || apiKey == "" {
		return fmt.Errorf("apiKey is required")
	}

	units, _ := config["units"].(string)
	if units == "" {
		units = "metric"
	}

	Plugin = &WeatherPlugin{
		config: PluginConfig{
			APIKey: apiKey,
			Units:  units,
		},
		client: &http.Client{Timeout: 10 * time.Second},
	}

	return nil
}

// Start 启动插件
func Start() error {
	return nil
}

// Stop 停止插件
func Stop() error {
	return nil
}

// Name 获取插件名称
func Name() string {
	return "weather"
}

// Execute 执行函数
func Execute(call FunctionCall) (string, error) {
	switch call.Name {
	case "get_weather":
		return Plugin.getWeather(call.Arguments)
	case "get_forecast":
		return Plugin.getForecast(call.Arguments)
	default:
		return "", fmt.Errorf("unknown function: %s", call.Name)
	}
}

// getWeather 获取天气
func (p *WeatherPlugin) getWeather(args json.RawMessage) (string, error) {
	var params struct {
		Location string `json:"location"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", err
	}

	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=%s",
		params.Location, p.config.APIKey, p.config.Units,
	)

	resp, err := p.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("weather API error: %s", string(body))
	}

	var weather WeatherResponse
	if err := json.Unmarshal(body, &weather); err != nil {
		return "", err
	}

	return formatWeather(weather), nil
}

// getForecast 获取预报
func (p *WeatherPlugin) getForecast(args json.RawMessage) (string, error) {
	var params struct {
		Location string `json:"location"`
		Days     int    `json:"days"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", err
	}

	if params.Days == 0 {
		params.Days = 5
	}

	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/forecast?q=%s&appid=%s&units=%s&cnt=%d",
		params.Location, p.config.APIKey, p.config.Units, params.Days*8,
	)

	resp, err := p.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// formatWeather 格式化天气信息
func formatWeather(w WeatherResponse) string {
	desc := ""
	if len(w.Weather) > 0 {
		desc = w.Weather[0].Description
	}

	return fmt.Sprintf(
		"📍 %s\n\n🌡️ 温度: %.1f°C\n"+"💧 体感: %.1f°C\n"+"💨 湿度: %d%%\n"+"🌬️ 风速: %.1f m/s\n\n%s",
		w.Name, w.Main.Temp, w.Main.FeelsLike,
		w.Main.Humidity, w.Wind.Speed, desc,
	)
}
