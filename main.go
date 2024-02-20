package main

import (
	"ThinkBot/relay"
	"embed"
	"flag"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
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

	router := gin.Default()

	// 静态文件不打包
	// router.Static("/static", "static")
	// router.LoadHTMLGlob("templates/*")
	// router.GET("/", indexHandler)

	// 将嵌入的文件系统转换为fs.FS
	viewsFS := embedFolder(ViewsFS, "templates")
	// 解析模板
	tmpl := template.Must(template.New("").ParseFS(viewsFS, "*.html"))
	// 使用解析后的模板
	router.SetHTMLTemplate(tmpl)
	// 使用embed打包静态资源
	staticRootFS, _ := fs.Sub(StaticFS, "static")
	// 设置静态资源路由
	router.StaticFS("/static", http.FS(staticRootFS))
	// 主页路由
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "chat.html", nil)
	})

	// 聊天路由
	router.POST("/chat", chatHandler)
	router.Run(":" + strconv.Itoa(port))
}

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "chat.html", nil)
}

// 嵌入的文件系统需要转换为fs.FS（即io/fs中的FS接口）
func embedFolder(efs embed.FS, path string) fs.FS {
	fsys, err := fs.Sub(efs, path)
	if err != nil {
		panic(err) // 如果有错误，抛出异常
	}
	return fsys
}
func chatHandler(c *gin.Context) {
	var chatReq ChatRequest
	if err := c.ShouldBind(&chatReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	modelInfo := parseModelInfo(chatReq.SelectedModelInfo)
	switch strings.ToLower(modelInfo["provider"]) {
	case "fucker":
		// 内置
	case "openai":
		relay.HandleOpenAIRequest(c, modelInfo, chatReq.Prompts)
	case "google":
		c.JSON(http.StatusBadRequest, gin.H{"error": "Google provider is not supported"})
	case "ali":
		relay.HandleAliDashscopeRequest(c, modelInfo, chatReq.Prompts)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown provider"})
	}
}

func parseModelInfo(selectedModelInfo string) map[string]string {
	return map[string]string{
		"modelId":       gjson.Get(selectedModelInfo, "modelId").String(),
		"provider":      gjson.Get(selectedModelInfo, "provider").String(),
		"modelEndpoint": gjson.Get(selectedModelInfo, "modelEndpoint").String(),
		"apiKey":        gjson.Get(selectedModelInfo, "apiKey").String(),
	}
}
