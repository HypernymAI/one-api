package cohere

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
)

var (
	WebSearchConnector = Connector{ID: "web-search"}
)

func stopReasonCohere2OpenAI(reason *string) string {
	if reason == nil {
		return ""
	}
	switch *reason {
	case "COMPLETE":
		return "stop"
	default:
		return *reason
	}
}

func ConvertRequest(textRequest model.GeneralOpenAIRequest) *Request {
	cohereRequest := Request{
		Model:            textRequest.Model,
		Message:          "",
		MaxTokens:        textRequest.MaxTokens,
		Temperature:      textRequest.Temperature,
		P:                textRequest.TopP,
		K:                textRequest.TopK,
		Stream:           textRequest.Stream,
		FrequencyPenalty: textRequest.FrequencyPenalty,
		PresencePenalty:  textRequest.FrequencyPenalty,
		Seed:             int(textRequest.Seed),
	}
	if cohereRequest.Model == "" {
		cohereRequest.Model = "command-r"
	}
	if strings.HasSuffix(cohereRequest.Model, "-internet") {
		cohereRequest.Model = strings.TrimSuffix(cohereRequest.Model, "-internet")
		cohereRequest.Connectors = append(cohereRequest.Connectors, WebSearchConnector)
	}
	for _, message := range textRequest.Messages {
		if message.Role == "user" {
			cohereRequest.Message = message.Content.(string)
		} else {
			var role string
			if message.Role == "assistant" {
				role = "CHATBOT"
			} else if message.Role == "system" {
				role = "SYSTEM"
			} else {
				role = "USER"
			}
			cohereRequest.ChatHistory = append(cohereRequest.ChatHistory, ChatMessage{
				Role:    role,
				Message: message.Content.(string),
			})
		}
	}
	return &cohereRequest
}

func ConvertRequestV2(textRequest model.GeneralOpenAIRequest) *RequestV2 {
	cohereRequest := RequestV2{
		Model:       textRequest.Model,
		Stream:      textRequest.Stream,
		Temperature: textRequest.Temperature,
		MaxTokens:   textRequest.MaxTokens,
		P:           textRequest.TopP,
		K:           textRequest.TopK,
		Seed:        int(textRequest.Seed),
		Messages:    []MessageV2Request{},
	}

	// Convert OpenAI messages to Cohere v2 format
	for _, message := range textRequest.Messages {
		cohereRequest.Messages = append(cohereRequest.Messages, MessageV2Request{
			Role:    message.Role,
			Content: message.Content.(string),
		})
	}

	return &cohereRequest
}

func StreamResponseCohere2OpenAI(cohereResponse *StreamResponse) (*openai.ChatCompletionsStreamResponse, *Response) {
	var response *Response
	var responseText string
	var finishReason string

	switch cohereResponse.EventType {
	case "stream-start":
		return nil, nil
	case "text-generation":
		responseText += cohereResponse.Text
	case "stream-end":
		usage := cohereResponse.Response.Meta.Tokens
		response = &Response{
			Meta: Meta{
				Tokens: Usage{
					InputTokens:  usage.InputTokens,
					OutputTokens: usage.OutputTokens,
				},
			},
		}
		finishReason = *cohereResponse.Response.FinishReason
	default:
		return nil, nil
	}

	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = responseText
	choice.Delta.Role = "assistant"
	if finishReason != "" {
		choice.FinishReason = &finishReason
	}
	var openaiResponse openai.ChatCompletionsStreamResponse
	openaiResponse.Object = "chat.completion.chunk"
	openaiResponse.Choices = []openai.ChatCompletionsStreamResponseChoice{choice}
	return &openaiResponse, response
}

func ResponseCohere2OpenAI(cohereResponse *Response) *openai.TextResponse {
	choice := openai.TextResponseChoice{
		Index: 0,
		Message: model.Message{
			Role:    "assistant",
			Content: cohereResponse.Text,
			Name:    nil,
		},
		FinishReason: stopReasonCohere2OpenAI(cohereResponse.FinishReason),
	}
	fullTextResponse := openai.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", cohereResponse.ResponseID),
		Model:   "model",
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: []openai.TextResponseChoice{choice},
	}
	return &fullTextResponse
}

func StreamHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	createdTime := helper.GetTimestamp()
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})

	dataChan := make(chan string)
	stopChan := make(chan bool)
	go func() {
		for scanner.Scan() {
			data := scanner.Text()
			dataChan <- data
		}
		stopChan <- true
	}()
	common.SetEventStreamHeaders(c)
	var usage model.Usage
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			// some implementations may add \r at the end of data
			data = strings.TrimSuffix(data, "\r")
			var cohereResponse StreamResponse
			err := json.Unmarshal([]byte(data), &cohereResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			response, meta := StreamResponseCohere2OpenAI(&cohereResponse)
			if meta != nil {
				usage.PromptTokens += meta.Meta.Tokens.InputTokens
				usage.CompletionTokens += meta.Meta.Tokens.OutputTokens
				return true
			}
			if response == nil {
				return true
			}
			response.Id = fmt.Sprintf("chatcmpl-%d", createdTime)
			response.Model = c.GetString("original_model")
			response.Created = createdTime
			jsonStr, err := json.Marshal(response)
			if err != nil {
				logger.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonStr)})
			return true
		case <-stopChan:
			c.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}
	})
	_ = resp.Body.Close()
	return nil, &usage
}

func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	// Check if it's a v2 response by looking for the "id" field
	var checkResponse map[string]interface{}
	err = json.Unmarshal(responseBody, &checkResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_check_response_failed", http.StatusInternalServerError), nil
	}

	// If it has "id" field and "message" with "content" array, it's v2
	if id, hasID := checkResponse["id"]; hasID && id != nil {
		if msg, hasMsg := checkResponse["message"]; hasMsg && msg != nil {
			return HandlerV2(c, responseBody, modelName)
		}
	}

	// Otherwise, handle as v1
	var cohereResponse Response
	err = json.Unmarshal(responseBody, &cohereResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if cohereResponse.ResponseID == "" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: cohereResponse.Message,
				Type:    cohereResponse.Message,
				Param:   "",
				Code:    resp.StatusCode,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := ResponseCohere2OpenAI(&cohereResponse)
	fullTextResponse.Model = modelName
	usage := model.Usage{
		PromptTokens:     cohereResponse.Meta.Tokens.InputTokens,
		CompletionTokens: cohereResponse.Meta.Tokens.OutputTokens,
		TotalTokens:      cohereResponse.Meta.Tokens.InputTokens + cohereResponse.Meta.Tokens.OutputTokens,
	}
	fullTextResponse.Usage = usage
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &usage
}

func HandlerV2(c *gin.Context, responseBody []byte, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	var cohereResponse ResponseV2
	err := json.Unmarshal(responseBody, &cohereResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_v2_response_failed", http.StatusInternalServerError), nil
	}

	// Extract text content from the response (ignore thinking for now)
	var textContent string
	for _, content := range cohereResponse.Message.Content {
		if content.Type == "text" {
			textContent = content.Text
			break
		}
	}

	// Convert to OpenAI format
	choice := openai.TextResponseChoice{
		Index: 0,
		Message: model.Message{
			Role:    "assistant",
			Content: textContent,
		},
		FinishReason: strings.ToLower(cohereResponse.FinishReason),
	}

	fullTextResponse := openai.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", cohereResponse.ID),
		Model:   modelName,
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: []openai.TextResponseChoice{choice},
	}

	// Use billed units for usage tracking
	usage := model.Usage{
		PromptTokens:     cohereResponse.Usage.BilledUnits.InputTokens,
		CompletionTokens: cohereResponse.Usage.BilledUnits.OutputTokens,
		TotalTokens:      cohereResponse.Usage.BilledUnits.InputTokens + cohereResponse.Usage.BilledUnits.OutputTokens,
	}
	fullTextResponse.Usage = usage

	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_v2_response_failed", http.StatusInternalServerError), nil
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(http.StatusOK)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &usage
}
