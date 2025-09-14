package openai

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/minimax"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
	"golang.org/x/oauth2/google"
	"io"
	"net/http"
	"strings"
)

type Adaptor struct {
	ChannelType int
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.ChannelType = meta.ChannelType
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	switch meta.ChannelType {
	case channeltype.Azure:
		if meta.Mode == relaymode.ImagesGenerations {
			// https://learn.microsoft.com/en-us/azure/ai-services/openai/dall-e-quickstart?tabs=dalle3%2Ccommand-line&pivots=rest-api
			// https://{resource_name}.openai.azure.com/openai/deployments/dall-e-3/images/generations?api-version=2024-03-01-preview
			fullRequestURL := fmt.Sprintf("%s/openai/deployments/%s/images/generations?api-version=%s", meta.BaseURL, meta.ActualModelName, meta.Config.APIVersion)
			return fullRequestURL, nil
		}

		// https://learn.microsoft.com/en-us/azure/cognitive-services/openai/chatgpt-quickstart?pivots=rest-api&tabs=command-line#rest-api
		requestURL := strings.Split(meta.RequestURLPath, "?")[0]
		requestURL = fmt.Sprintf("%s?api-version=%s", requestURL, meta.Config.APIVersion)
		task := strings.TrimPrefix(requestURL, "/v1/")
		model_ := meta.ActualModelName
		// Only remove dots for API versions before 2025
		// Newer Azure deployments (2025+) support and require dots for models like gpt-4.1
		// Check if API version starts with 2025 or later
		if len(meta.Config.APIVersion) >= 4 && meta.Config.APIVersion[:4] < "2025" {
			model_ = strings.Replace(model_, ".", "", -1)
		}
		//https://github.com/songquanpeng/one-api/issues/1191
		// {your endpoint}/openai/deployments/{your azure_model}/chat/completions?api-version={api_version}
		requestURL = fmt.Sprintf("/openai/deployments/%s/%s", model_, task)
		return GetFullRequestURL(meta.BaseURL, requestURL, meta.ChannelType), nil
	case channeltype.GoogleOpenAI:
		// https://${ENDPOINT}/v1/projects/${PROJECT_ID}/locations/${REGION}/endpoints/openapi/chat/completions
		task := strings.TrimPrefix(meta.RequestURLPath, "/v1/")
		requestURL := fmt.Sprintf("/v1/projects/%s/locations/%s/endpoints/openapi/%s", meta.Config.ProjectID, meta.Config.Region, task)
		return GetFullRequestURL(meta.BaseURL, requestURL, meta.ChannelType), nil
	case channeltype.Minimax:
		return minimax.GetRequestURL(meta)
	default:
		return GetFullRequestURL(meta.BaseURL, meta.RequestURLPath, meta.ChannelType), nil
	}
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	if meta.ChannelType == channeltype.Azure {
		req.Header.Set("api-key", meta.APIKey)
		return nil
	}
	if meta.ChannelType == channeltype.GoogleOpenAI {
		// Check if APIKey is service account JSON
		if strings.HasPrefix(meta.APIKey, "{") && strings.Contains(meta.APIKey, "private_key") {
			// Generate OAuth token from service account JSON
			token, err := getGoogleCloudToken(meta.APIKey)
			if err != nil {
				return fmt.Errorf("failed to get Google Cloud token: %v", err)
			}
			req.Header.Set("Authorization", "Bearer "+token)
		} else {
			// Treat as access token
			req.Header.Set("Authorization", "Bearer "+meta.APIKey)
		}
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	if meta.ChannelType == channeltype.OpenRouter {
		req.Header.Set("HTTP-Referer", "https://github.com/songquanpeng/one-api")
		req.Header.Set("X-Title", "One API")
	}
	return nil
}

func getGoogleCloudToken(serviceAccountJSON string) (string, error) {
	ctx := context.Background()
	
	// Parse the service account JSON
	creds, err := google.CredentialsFromJSON(ctx, []byte(serviceAccountJSON), 
		"https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", fmt.Errorf("failed to parse service account JSON: %v", err)
	}
	
	// Get the token
	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get token: %v", err)
	}
	
	return token.AccessToken, nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	
	// Azure-specific handling for reasoning models (gpt-5-mini, gpt-5-nano)
	// These models require:
	// 1. max_completion_tokens instead of max_tokens
	// 2. temperature must be exactly 1.0 (reasoning models don't accept other values)
	// TODO: Remove this translation when industry standardizes v2 API for reasoning models
	// This maintains v1 API compatibility while internally using Azure's required parameters
	if a.ChannelType == channeltype.Azure {
		modelName := c.GetString(ctxkey.RequestModel)
		// Only gpt-5-mini and gpt-5-nano require special handling
		if modelName == "gpt-5-mini" || modelName == "gpt-5-nano" {
			// Create a custom struct that excludes max_tokens and includes max_completion_tokens
			type AzureReasoningRequest struct {
				Messages            []model.Message       `json:"messages,omitempty"`
				Model               string                `json:"model,omitempty"`
				FrequencyPenalty    float64               `json:"frequency_penalty,omitempty"`
				MaxCompletionTokens int                   `json:"max_completion_tokens,omitempty"`
				N                   int                   `json:"n,omitempty"`
				PresencePenalty     float64               `json:"presence_penalty,omitempty"`
				ResponseFormat      *model.ResponseFormat `json:"response_format,omitempty"`
				Seed                float64               `json:"seed,omitempty"`
				Stream              bool                  `json:"stream,omitempty"`
				Temperature         float64               `json:"temperature,omitempty"`
				TopP                float64               `json:"top_p,omitempty"`
				TopK                int                   `json:"top_k,omitempty"`
				Tools               []model.Tool          `json:"tools,omitempty"`
				ToolChoice          any                   `json:"tool_choice,omitempty"`
				FunctionCall        any                   `json:"function_call,omitempty"`
				Functions           any                   `json:"functions,omitempty"`
				User                string                `json:"user,omitempty"`
			}
			
			azureReq := AzureReasoningRequest{
				Messages:            request.Messages,
				Model:               request.Model,
				FrequencyPenalty:    request.FrequencyPenalty,
				MaxCompletionTokens: request.MaxTokens, // Translate max_tokens to max_completion_tokens
				N:                   request.N,
				PresencePenalty:     request.PresencePenalty,
				ResponseFormat:      request.ResponseFormat,
				Seed:                request.Seed,
				Stream:              request.Stream,
				Temperature:         1.0, // REQUIRED: GPT-5 reasoning models only accept temperature=1.0
				TopP:                request.TopP,
				TopK:                request.TopK,
				Tools:               request.Tools,
				ToolChoice:          request.ToolChoice,
				FunctionCall:        request.FunctionCall,
				Functions:           request.Functions,
				User:                request.User,
			}
			
			return azureReq, nil
		}
	}
	
	return request, nil
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		var responseText string
		err, responseText, usage = StreamHandler(c, resp, meta.Mode)
		if usage == nil || usage.TotalTokens == 0 {
			usage = ResponseText2Usage(responseText, meta.ActualModelName, meta.PromptTokens)
		}
		if usage.TotalTokens != 0 && usage.PromptTokens == 0 { // some channels don't return prompt tokens & completion tokens
			usage.PromptTokens = meta.PromptTokens
			usage.CompletionTokens = usage.TotalTokens - meta.PromptTokens
		}
	} else {
		switch meta.Mode {
		case relaymode.ImagesGenerations:
			err, _ = ImageHandler(c, resp)
		default:
			err, usage = Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
		}
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	_, modelList := GetCompatibleChannelMeta(a.ChannelType)
	return modelList
}

func (a *Adaptor) GetChannelName() string {
	channelName, _ := GetCompatibleChannelMeta(a.ChannelType)
	return channelName
}
