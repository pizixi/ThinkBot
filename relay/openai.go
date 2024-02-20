package relay

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
)

// HandleOpenAIRequest handles OpenAI requests
func HandleOpenAIRequest(c *gin.Context, modelInfo map[string]string, promptsJSON string) {
	config := openai.DefaultConfig(modelInfo["apiKey"])
	config.BaseURL = modelInfo["modelEndpoint"]
	var prompts []openai.ChatCompletionMessage
	if err := json.Unmarshal([]byte(promptsJSON), &prompts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prompts format"})
		return
	}
	ctx := context.Background()
	client := openai.NewClientWithConfig(config)

	req := openai.ChatCompletionRequest{
		Model: modelInfo["modelId"],
		// MaxTokens: 20,
		Messages: prompts,
		Stream:   true,
	}
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat stream"})
		return
	}
	defer stream.Close()

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to receive chat stream"})
			return
		}
		_, err = c.Writer.WriteString(response.Choices[0].Delta.Content)
		if err != nil {
			c.Error(err) // 发送错误并中断请求
			return
		}
		c.Writer.Flush()
	}
}
