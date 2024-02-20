package relay

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// ModelInput 定义了模型输入的结构体
type ModelInput struct {
	Messages []Message `json:"messages"`
}

// Message 定义了消息的结构体
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DashscopeRequest 请求结构体
type DashscopeRequest struct {
	Model string     `json:"model"`
	Input ModelInput `json:"input"`
}

// 定义全局 http.Client
var httpClient = &http.Client{}

// HandleAliDashscopeRequest 通用文本生成api处理请求
func HandleAliDashscopeRequest(c *gin.Context, modelInfo map[string]string, promptsJSON string) {
	const url = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"

	var prompts []Message
	if err := json.Unmarshal([]byte(promptsJSON), &prompts); err != nil {
		http.Error(c.Writer, "Invalid prompts format", http.StatusBadRequest)
		return
	}

	dashscopeReq := DashscopeRequest{
		Model: modelInfo["modelId"],
		Input: ModelInput{Messages: prompts},
	}

	requestBody, err := json.Marshal(dashscopeReq)
	if err != nil {
		http.Error(c.Writer, "Failed to marshal request body", http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		http.Error(c.Writer, "Failed to create request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+modelInfo["apiKey"])
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-DashScope-SSE", "enable")

	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(c.Writer, "Failed to send request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			http.Error(c.Writer, "Failed to decode error response", http.StatusInternalServerError)
			return
		}
		c.JSON(resp.StatusCode, errorResponse)
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	reader := bufio.NewReader(resp.Body)
	previousText := ""
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				http.Error(c.Writer, "Error reading stream", http.StatusInternalServerError)
			}
			break
		}

		if !bytes.HasPrefix(line, []byte("data:")) {
			continue
		}

		text := gjson.Get(string(line[5:]), "output.text").String()
		newText := strings.Replace(text, previousText, "", 1)
		previousText = text
		if _, writeErr := c.Writer.WriteString(newText); writeErr != nil {
			http.Error(c.Writer, "Error writing to client", http.StatusInternalServerError)
			break
		}
		c.Writer.Flush()
	}
}
