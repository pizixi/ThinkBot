package main

import (
	"ThinkBot/relay"
	"embed"
	"flag"
	"io"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/labstack/echo/v4"
	"github.com/tidwall/gjson"
)

//go:embed templates/*
var ViewsFS embed.FS

//go:embed static
var StaticFS embed.FS

// ChatRequest 定义请求和响应结构体
type ChatRequest struct {
	Prompts           string `form:"prompts"`
	APIKey            string `form:"apiKey"`
	SelectedModelInfo string `form:"selectedModelInfo"`
	Model             string `form:"model"`
}

func main() {
	var port int
	flag.IntVar(&port, "port", 20093, "The port to run the server on.")
	flag.Parse()

	// 创建 Echo 实例
	e := echo.New()

	// 静态文件不打包
	e.Static("/static", "static")
	// 设置模板渲染器
	e.Renderer = newTemplateRenderer("templates/*")

	// 路由设置
	e.GET("/", indexHandler)
	e.POST("/chat", chatHandler)
	e.GET("/api/proxy/models", handleModelProxy)

	// 启动服务器
	e.Logger.Fatal(e.Start(":" + strconv.Itoa(port)))
}

func indexHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "chat.html", nil)
}

func chatHandler(c echo.Context) error {
	var chatReq ChatRequest
	if err := c.Bind(&chatReq); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request parameters"})
	}

	modelInfo := parseModelInfo(chatReq.SelectedModelInfo)
	switch strings.ToLower(modelInfo["provider"]) {
	case "onehub":
		return relay.HandleOpenAIRequest(c, modelInfo, chatReq.Prompts)
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Unknown provider"})
	}
}

func parseModelInfo(selectedModelInfo string) map[string]string {
	return map[string]string{
		"modelId":       gjson.Get(selectedModelInfo, "modelId").String(),
		"provider":      gjson.Get(selectedModelInfo, "provider").String(),
		"modelEndpoint": gjson.Get(selectedModelInfo, "modelEndpoint").String(),
		"apiKey":        gjson.Get(selectedModelInfo, "apiKey").String(),
		"oneHubToken":   gjson.Get(selectedModelInfo, "oneHubToken").String(),
	}
}

func handleModelProxy(c echo.Context) error {
	endpoint := c.QueryParam("endpoint")
	oneHubToken := c.QueryParam("oneHubToken")

	if endpoint == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "endpoint is required"})
	}

	if oneHubToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "oneHubToken is required"})
	}

	// 构建目标URL
	targetURL := strings.TrimSuffix(endpoint, "/v1") + "/api/user/models"

	// 创建新的请求
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create request"})
	}

	// 添加认证头
	req.Header.Set("Authorization", "Bearer "+oneHubToken)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch models"})
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read response"})
	}

	// 设置响应头
	c.Response().Header().Set("Content-Type", "application/json")

	// 返回响应
	return c.String(resp.StatusCode, string(body))
}

// 模板渲染器
type templateRenderer struct {
	templates *template.Template
}

func newTemplateRenderer(pattern string) *templateRenderer {
	return &templateRenderer{
		templates: template.Must(template.ParseGlob(pattern)),
	}
}

func (t *templateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
