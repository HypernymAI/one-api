// Package aws provides the AWS adaptor for the relay service.
package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"io"
	"net/http"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/aws/smithy-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/relay/adaptor/anthropic"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

func wrapErr(err error) *relaymodel.ErrorWithStatusCode {
	statusCode := http.StatusInternalServerError
	
	// Check for AWS API errors using smithy error interface
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		// Map AWS error codes to HTTP status codes
		switch apiErr.ErrorCode() {
		case "ThrottlingException", "TooManyRequestsException", "RequestLimitExceeded", "ServiceQuotaExceededException":
			statusCode = http.StatusTooManyRequests
		case "AccessDeniedException":
			statusCode = http.StatusForbidden
		case "ValidationException":
			statusCode = http.StatusBadRequest
		case "ResourceNotFoundException":
			statusCode = http.StatusNotFound
		}
	}
	
	return &relaymodel.ErrorWithStatusCode{
		StatusCode: statusCode,
		Error: relaymodel.Error{
			Message: fmt.Sprintf("%s", err.Error()),
		},
	}
}

// https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html
var awsModelIDMap = map[string]string{
	// Claude models
	"claude-instant-1.2":         "anthropic.claude-instant-v1",
	"claude-2.0":                 "anthropic.claude-v2",
	"claude-2.1":                 "anthropic.claude-v2:1",
	"claude-3-sonnet-20240229":   "us.anthropic.claude-3-sonnet-20240229-v1:0",
	"claude-3-opus-20240229":     "us.anthropic.claude-3-opus-20240229-v1:0",
	"claude-3-haiku-20240307":    "us.anthropic.claude-3-haiku-20240307-v1:0",
	"claude-3-5-sonnet-20240620": "us.anthropic.claude-3-5-sonnet-20240620-v1:0",
	"claude-3-5-sonnet-20241022": "us.anthropic.claude-3-5-sonnet-20241022-v2:0",
	"claude-3-5-haiku-20241022":  "us.anthropic.claude-3-5-haiku-20241022-v1:0",
	"claude-opus-4-20250514":     "us.anthropic.claude-opus-4-20250514-v1:0",
	"claude-opus-4-1-20250805":   "us.anthropic.claude-opus-4-1-20250805-v1:0",
	"claude-3-7-sonnet-20250219": "us.anthropic.claude-3-7-sonnet-20250219-v1:0",
	"claude-sonnet-4-20250514":   "us.anthropic.claude-sonnet-4-20250514-v1:0",
	"claude-sonnet-4-5-20250929": "us.anthropic.claude-sonnet-4-5-20250929-v1:0",

	// Llama models
	"llama-3-8b":           "meta.llama3-8b-instruct-v1:0",
	"llama-3-70b":          "meta.llama3-70b-instruct-v1:0",
	"llama-3.1-8b":         "meta.llama3-1-8b-instruct-v1:0",
	"llama-3.1-70b":        "meta.llama3-1-70b-instruct-v1:0",
	"llama-3.2-1b":         "meta.llama3-2-1b-instruct-v1:0",
	"llama-3.2-3b":         "meta.llama3-2-3b-instruct-v1:0",
	"llama-3.2-11b":        "meta.llama3-2-11b-instruct-v1:0",
	"llama-3.2-90b":        "meta.llama3-2-90b-instruct-v1:0",
	"llama-3.3-70b":        "meta.llama3-3-70b-instruct-v1:0",
	"llama-4-scout-17b":    "us.meta.llama4-scout-17b-instruct-v1:0",
	"llama-4-maverick-17b": "us.meta.llama4-maverick-17b-instruct-v1:0",
}

func awsModelID(requestModel string) (string, error) {
	if awsModelID, ok := awsModelIDMap[requestModel]; ok {
		return awsModelID, nil
	}

	return "", errors.Errorf("model %s not found", requestModel)
}

// Check if the model is a Llama model
func isLlamaModel(model string) bool {
	return strings.HasPrefix(model, "llama-")
}

// Llama prompt template
const llamaPromptTemplate = `<|begin_of_text|>{{range .Messages}}<|start_header_id|>{{.Role}}<|end_header_id|>{{.StringContent}}<|eot_id|>{{end}}<|start_header_id|>assistant<|end_header_id|>
`

var llamaPromptTpl = template.Must(template.New("llama-chat").Parse(llamaPromptTemplate))

// RenderLlamaPrompt renders messages into Llama format
func RenderLlamaPrompt(messages []relaymodel.Message) (string, error) {
	var buf bytes.Buffer
	err := llamaPromptTpl.Execute(&buf, struct{ Messages []relaymodel.Message }{messages})
	if err != nil {
		return "", errors.Wrap(err, "render llama prompt")
	}
	return buf.String(), nil
}

// ConvertLlamaRequest converts OpenAI request to Llama request
func ConvertLlamaRequest(request relaymodel.GeneralOpenAIRequest) (*LlamaRequest, error) {
	prompt, err := RenderLlamaPrompt(request.Messages)
	if err != nil {
		return nil, err
	}
	
	maxGenLen := request.MaxTokens
	if maxGenLen == 0 {
		maxGenLen = 2048
	}
	if maxGenLen > 8192 {
		maxGenLen = 8192
	}
	
	return &LlamaRequest{
		Prompt:      prompt,
		MaxGenLen:   maxGenLen,
		Temperature: request.Temperature,
		TopP:        request.TopP,
	}, nil
}

// ResponseLlama2OpenAI converts Llama response to OpenAI format
func ResponseLlama2OpenAI(llamaResponse *LlamaResponse, model string) *openai.TextResponse {
	choice := openai.TextResponseChoice{
		Index: 0,
		Message: relaymodel.Message{
			Role:    "assistant",
			Content: llamaResponse.Generation,
		},
		FinishReason: llamaResponse.StopReason,
	}
	return &openai.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", random.GetUUID()),
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Model:   model,
		Choices: []openai.TextResponseChoice{choice},
		Usage: relaymodel.Usage{
			PromptTokens:     llamaResponse.PromptTokenCount,
			CompletionTokens: llamaResponse.GenerationTokenCount,
			TotalTokens:      llamaResponse.PromptTokenCount + llamaResponse.GenerationTokenCount,
		},
	}
}

// StreamResponseLlama2OpenAI converts Llama stream response to OpenAI format
func StreamResponseLlama2OpenAI(llamaResponse *LlamaStreamResponse) *openai.ChatCompletionsStreamResponse {
	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = llamaResponse.Generation
	choice.Delta.Role = "assistant"
	finishReason := llamaResponse.StopReason
	if finishReason != "" && finishReason != "null" {
		choice.FinishReason = &finishReason
	}
	return &openai.ChatCompletionsStreamResponse{
		Object:  "chat.completion.chunk",
		Choices: []openai.ChatCompletionsStreamResponseChoice{choice},
	}
}

func Handler(c *gin.Context, awsCli *bedrockruntime.Client, modelName string) (*relaymodel.ErrorWithStatusCode, *relaymodel.Usage) {
	requestModel := c.GetString(ctxkey.RequestModel)
	awsModelId, err := awsModelID(requestModel)
	if err != nil {
		return wrapErr(errors.Wrap(err, "awsModelID")), nil
	}

	awsReq := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(awsModelId),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	}

	// Check if it's a Llama model
	if isLlamaModel(requestModel) {
		// Handle Llama model
		generalReq, ok := c.Get(ctxkey.ConvertedRequest)
		if !ok {
			return wrapErr(errors.New("request not found")), nil
		}
		
		// Convert to Llama request
		openAIReq := generalReq.(*relaymodel.GeneralOpenAIRequest)
		llamaReq, err := ConvertLlamaRequest(*openAIReq)
		if err != nil {
			return wrapErr(errors.Wrap(err, "convert llama request")), nil
		}
		
		awsReq.Body, err = json.Marshal(llamaReq)
		if err != nil {
			return wrapErr(errors.Wrap(err, "marshal llama request")), nil
		}
		
		awsResp, err := awsCli.InvokeModel(c.Request.Context(), awsReq)
		if err != nil {
			return wrapErr(errors.Wrap(err, "InvokeModel llama")), nil
		}
		
		var llamaResponse LlamaResponse
		err = json.Unmarshal(awsResp.Body, &llamaResponse)
		if err != nil {
			return wrapErr(errors.Wrap(err, "unmarshal llama response")), nil
		}
		
		openaiResp := ResponseLlama2OpenAI(&llamaResponse, modelName)
		c.JSON(http.StatusOK, openaiResp)
		return nil, &openaiResp.Usage
	} else {
		// Handle Claude model (existing code)
		claudeReq_, ok := c.Get(ctxkey.ConvertedRequest)
		if !ok {
			return wrapErr(errors.New("request not found")), nil
		}
		claudeReq := claudeReq_.(*anthropic.Request)
		awsClaudeReq := &Request{
			AnthropicVersion: "bedrock-2023-05-31",
		}
		if err = copier.Copy(awsClaudeReq, claudeReq); err != nil {
			return wrapErr(errors.Wrap(err, "copy request")), nil
		}

		awsReq.Body, err = json.Marshal(awsClaudeReq)
		if err != nil {
			return wrapErr(errors.Wrap(err, "marshal request")), nil
		}

		awsResp, err := awsCli.InvokeModel(c.Request.Context(), awsReq)
		if err != nil {
			return wrapErr(errors.Wrap(err, "InvokeModel")), nil
		}

		claudeResponse := new(anthropic.Response)
		err = json.Unmarshal(awsResp.Body, claudeResponse)
		if err != nil {
			return wrapErr(errors.Wrap(err, "unmarshal response")), nil
		}

		openaiResp := anthropic.ResponseClaude2OpenAI(claudeResponse)
		openaiResp.Model = modelName
		usage := relaymodel.Usage{
			PromptTokens:     claudeResponse.Usage.InputTokens,
			CompletionTokens: claudeResponse.Usage.OutputTokens,
			TotalTokens:      claudeResponse.Usage.InputTokens + claudeResponse.Usage.OutputTokens,
		}
		openaiResp.Usage = usage

		c.JSON(http.StatusOK, openaiResp)
		return nil, &usage
	}
}

func StreamHandler(c *gin.Context, awsCli *bedrockruntime.Client) (*relaymodel.ErrorWithStatusCode, *relaymodel.Usage) {
	createdTime := helper.GetTimestamp()
	requestModel := c.GetString(ctxkey.RequestModel)
	awsModelId, err := awsModelID(requestModel)
	if err != nil {
		return wrapErr(errors.Wrap(err, "awsModelID")), nil
	}

	awsReq := &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(awsModelId),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	}

	// Check if it's a Llama model
	if isLlamaModel(requestModel) {
		// Handle Llama streaming
		generalReq, ok := c.Get(ctxkey.ConvertedRequest)
		if !ok {
			return wrapErr(errors.New("request not found")), nil
		}
		
		openAIReq := generalReq.(*relaymodel.GeneralOpenAIRequest)
		llamaReq, err := ConvertLlamaRequest(*openAIReq)
		if err != nil {
			return wrapErr(errors.Wrap(err, "convert llama request")), nil
		}
		
		awsReq.Body, err = json.Marshal(llamaReq)
		if err != nil {
			return wrapErr(errors.Wrap(err, "marshal llama request")), nil
		}
	} else {
		// Handle Claude streaming (existing code)
		claudeReq_, ok := c.Get(ctxkey.ConvertedRequest)
		if !ok {
			return wrapErr(errors.New("request not found")), nil
		}
		claudeReq := claudeReq_.(*anthropic.Request)

		awsClaudeReq := &Request{
			AnthropicVersion: "bedrock-2023-05-31",
		}
		if err = copier.Copy(awsClaudeReq, claudeReq); err != nil {
			return wrapErr(errors.Wrap(err, "copy request")), nil
		}
		awsReq.Body, err = json.Marshal(awsClaudeReq)
		if err != nil {
			return wrapErr(errors.Wrap(err, "marshal request")), nil
		}
	}

	awsResp, err := awsCli.InvokeModelWithResponseStream(c.Request.Context(), awsReq)
	if err != nil {
		return wrapErr(errors.Wrap(err, "InvokeModelWithResponseStream")), nil
	}
	stream := awsResp.GetStream()
	defer stream.Close()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	var usage relaymodel.Usage
	var id string = fmt.Sprintf("chatcmpl-%s", random.GetUUID())
	
	c.Stream(func(w io.Writer) bool {
		event, ok := <-stream.Events()
		if !ok {
			c.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}

		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			if isLlamaModel(requestModel) {
				// Handle Llama streaming response
				var llamaResp LlamaStreamResponse
				err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&llamaResp)
				if err != nil {
					logger.SysError("error unmarshalling llama stream response: " + err.Error())
					return false
				}

				// Update usage if provided
				if llamaResp.PromptTokenCount > 0 {
					usage.PromptTokens = llamaResp.PromptTokenCount
				}
				if llamaResp.StopReason == "stop" || llamaResp.StopReason == "length" {
					usage.CompletionTokens = llamaResp.GenerationTokenCount
					usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
				}

				response := StreamResponseLlama2OpenAI(&llamaResp)
				response.Id = id
				response.Model = c.GetString(ctxkey.OriginalModel)
				response.Created = createdTime
				
				jsonStr, err := json.Marshal(response)
				if err != nil {
					logger.SysError("error marshalling llama stream response: " + err.Error())
					return true
				}
				c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonStr)})
				return true
			} else {
				// Handle Claude streaming response (existing code)
				claudeResp := new(anthropic.StreamResponse)
				err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(claudeResp)
				if err != nil {
					logger.SysError("error unmarshalling stream response: " + err.Error())
					return false
				}

				response, meta := anthropic.StreamResponseClaude2OpenAI(claudeResp)
				if meta != nil {
					usage.PromptTokens += meta.Usage.InputTokens
					usage.CompletionTokens += meta.Usage.OutputTokens
					id = fmt.Sprintf("chatcmpl-%s", meta.Id)
					return true
				}
				if response == nil {
					return true
				}
				response.Id = id
				response.Model = c.GetString(ctxkey.OriginalModel)
				response.Created = createdTime
				jsonStr, err := json.Marshal(response)
				if err != nil {
					logger.SysError("error marshalling stream response: " + err.Error())
					return true
				}
				c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonStr)})
				return true
			}
		case *types.UnknownUnionMember:
			fmt.Println("unknown tag:", v.Tag)
			return false
		default:
			fmt.Println("union is nil or unknown type")
			return false
		}
	})

	return nil, &usage
}
