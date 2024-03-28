package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bincooo/edge-api"
	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
)

const (
	cookie = "xxx"

	KievAuth = "xxx"
	RwBf     = "xxx"
)

var pMessages = []edge.ChatMessage{
	// {
	// 	"author": "user",
	// 	"text":   "XXXX",
	// },
	// {
	// 	"author": "bot",
	// 	"text":   "XXXX",
	// },
}

// HandleCopilotRequest 获取Copilot的回复
func HandleCopilotRequest(c *gin.Context, modelInfo map[string]string, promptsJSON string) {
	// modelInfo["modelEndpoint"] 、modelInfo["apiKey"]、modelInfo["modelId"]
	options, err := edge.NewDefaultOptions(cookie, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create edge options"})
		return
	}
	options.KievAuth(KievAuth, RwBf)
	var prompts []openai.ChatCompletionMessage
	if err := json.Unmarshal([]byte(promptsJSON), &prompts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prompts format"})
		return
	}
	var prompt string
	for i, p := range prompts {
		if i == len(prompts)-1 && p.Role == "user" {
			prompt = p.Content
			break
		}
		if p.Role == "assistant" {
			pMessages = append(pMessages, edge.ChatMessage{
				"author": "bot",
				"text":   p.Content,
			})
		} else {
			pMessages = append(pMessages, edge.ChatMessage{
				"author": "user",
				"text":   p.Content,
			})
		}
	}
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prompts format"})
		return
	}
	// Sydney 模式需要自行维护历史对话
	chat := edge.New(options.
		Proxies("socks5://127.0.0.1:7788").
		Model(modelInfo["modelId"]).
		Temperature(1.0).
		TopicToE(true))

	fmt.Println("You: ", prompt)
	fmt.Print("Copilot: ")

	partialResponse, err := chat.Reply(context.Background(), prompt, nil, pMessages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to receive chat partialResponse"})
		return
	}
	msg := ""
	lastMsg := ""
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	for {
		message, ok := <-partialResponse
		if !ok {
			fmt.Println("Copilot完整回复: ", msg)
			break
		}

		if message.Error != nil {
			log.Fatal(message.Error)
		}

		if len(message.Text) > 0 {
			msg = message.Text
		}
		streamMsg := ""
		if lastMsg != "" {
			streamMsg = strings.ReplaceAll(strings.TrimSpace(message.Text), lastMsg, "")
		} else {
			streamMsg = strings.TrimSpace(message.Text)
		}
		fmt.Print(streamMsg)
		_, err = c.Writer.WriteString(streamMsg)
		if err != nil {
			c.Error(err) // 发送错误并中断请求
			return
		}
		c.Writer.Flush()

		lastMsg = strings.TrimSpace(message.Text)

		// fmt.Println("===============")
		if message.T != nil {
			fmt.Printf("\n%d / %d\n", message.T.Max, message.T.Used)
		}
	}

	if err = chat.Delete(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to Delete chat stream"})
		return
	}

}
