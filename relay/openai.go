package relay

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	openai "github.com/sashabaranov/go-openai"
)

// HandleOpenAIRequest handles OpenAI requests
func HandleOpenAIRequest(c echo.Context, modelInfo map[string]string, promptsJSON string) error {
	config := openai.DefaultConfig(modelInfo["apiKey"])
	config.BaseURL = modelInfo["modelEndpoint"]
	var prompts []openai.ChatCompletionMessage
	if err := json.Unmarshal([]byte(promptsJSON), &prompts); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid prompts format"})
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
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create chat stream"})
	}
	defer stream.Close()

	// 设置响应头以支持流式输出
	res := c.Response()
	res.Header().Set("Content-Type", "text/event-stream")
	res.Header().Set("Cache-Control", "no-cache")
	res.Header().Set("Connection", "keep-alive")
	res.Header().Set("Transfer-Encoding", "chunked")

	// 创建一个 Flusher
	flusher, ok := res.Writer.(http.Flusher)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Streaming not supported"})
	}

	for {
		streamResponse, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to receive chat stream"})
		}

		// 写入数据并立即刷新
		if _, err = res.Writer.Write([]byte(streamResponse.Choices[0].Delta.Content)); err != nil {
			return err
		}
		flusher.Flush()
	}
	return nil
}
