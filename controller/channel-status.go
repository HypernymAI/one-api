package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/channeltype"
)

type ModelStatus struct {
	Model     string `json:"model"`
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}

type ChannelStatusResponse struct {
	Success bool          `json:"success"`
	Channel string        `json:"channel"`
	Type    string        `json:"type"`
	Models  []ModelStatus `json:"models"`
	Time    float64       `json:"time"`
}

// CheckChannelModelsStatus checks if models are available without invoking them
func CheckChannelModelsStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	channel, err := model.GetChannelById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	tik := time.Now()
	var modelStatuses []ModelStatus

	// Get list of models for this channel
	models := strings.Split(channel.Models, ",")
	for i := range models {
		models[i] = strings.TrimSpace(models[i])
	}

	switch channel.Type {
	case channeltype.AwsClaude:
		modelStatuses = checkAWSBedrockModels(channel, models)
	case channeltype.Gemini:
		modelStatuses = checkGeminiModels(channel, models)
	case channeltype.Azure:
		modelStatuses = checkAzureModels(channel, models)
	default:
		// For unsupported types, return all models as "unknown"
		for _, model := range models {
			modelStatuses = append(modelStatuses, ModelStatus{
				Model:     model,
				Available: false,
				Error:     "Status check not supported for this channel type",
			})
		}
	}

	tok := time.Now()
	consumedTime := float64(tok.Sub(tik).Milliseconds()) / 1000.0

	response := ChannelStatusResponse{
		Success: true,
		Channel: channel.Name,
		Type:    getChannelTypeName(channel.Type),
		Models:  modelStatuses,
		Time:    consumedTime,
	}

	c.JSON(http.StatusOK, response)
}

func checkAWSBedrockModels(channel *model.Channel, models []string) []ModelStatus {
	var statuses []ModelStatus
	
	// Parse AWS credentials from channel key
	// Expected format: "ACCESS_KEY|SECRET_KEY|REGION"
	keyParts := strings.Split(channel.Key, "|")
	if len(keyParts) != 3 {
		for _, model := range models {
			statuses = append(statuses, ModelStatus{
				Model:     model,
				Available: false,
				Error:     "Invalid AWS credentials format",
			})
		}
		return statuses
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(keyParts[2]),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			keyParts[0], keyParts[1], "",
		)),
	)
	if err != nil {
		for _, model := range models {
			statuses = append(statuses, ModelStatus{
				Model:     model,
				Available: false,
				Error:     fmt.Sprintf("Failed to configure AWS: %v", err),
			})
		}
		return statuses
	}

	client := bedrock.NewFromConfig(cfg)

	// Get model mappings from AWS adapter
	modelMap := getAWSModelMap()

	for _, modelName := range models {
		// Convert friendly name to AWS model ID
		awsModelID, exists := modelMap[modelName]
		if !exists {
			statuses = append(statuses, ModelStatus{
				Model:     modelName,
				Available: false,
				Error:     "Model mapping not found",
			})
			continue
		}

		// For inference profiles (starting with "us." or "global."), we can't use GetFoundationModel
		// So we'll mark them as available if they're in our known list
		if strings.HasPrefix(awsModelID, "us.") || strings.HasPrefix(awsModelID, "global.") {
			// These are inference profiles, assume available if configured
			statuses = append(statuses, ModelStatus{
				Model:     modelName,
				Available: true,
			})
			continue
		}

		// For base model IDs, check with GetFoundationModel
		input := &bedrock.GetFoundationModelInput{
			ModelIdentifier: &awsModelID,
		}

		result, err := client.GetFoundationModel(ctx, input)
		if err != nil {
			statuses = append(statuses, ModelStatus{
				Model:     modelName,
				Available: false,
				Error:     fmt.Sprintf("Model not accessible: %v", err),
			})
			continue
		}

		// Check if model is active
		if result.ModelDetails != nil && 
		   result.ModelDetails.ModelLifecycle != nil &&
		   string(result.ModelDetails.ModelLifecycle.Status) == "ACTIVE" {
			statuses = append(statuses, ModelStatus{
				Model:     modelName,
				Available: true,
			})
		} else {
			statuses = append(statuses, ModelStatus{
				Model:     modelName,
				Available: false,
				Error:     "Model not active",
			})
		}
	}

	return statuses
}

func checkGeminiModels(channel *model.Channel, models []string) []ModelStatus {
	var statuses []ModelStatus
	
	apiKey := channel.Key
	
	for _, modelName := range models {
		// Check model availability via Gemini API
		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s?key=%s", 
			modelName, apiKey)
		
		resp, err := http.Get(url)
		if err != nil {
			statuses = append(statuses, ModelStatus{
				Model:     modelName,
				Available: false,
				Error:     fmt.Sprintf("Failed to check model: %v", err),
			})
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// Model exists and is accessible
			body, _ := io.ReadAll(resp.Body)
			var modelInfo map[string]interface{}
			if err := json.Unmarshal(body, &modelInfo); err == nil {
				// Check if model supports generateContent
				if methods, ok := modelInfo["supportedGenerationMethods"].([]interface{}); ok {
					hasGenerateContent := false
					for _, method := range methods {
						if method == "generateContent" {
							hasGenerateContent = true
							break
						}
					}
					statuses = append(statuses, ModelStatus{
						Model:     modelName,
						Available: hasGenerateContent,
						Error:     func() string {
							if !hasGenerateContent {
								return "Model doesn't support generateContent"
							}
							return ""
						}(),
					})
				} else {
					statuses = append(statuses, ModelStatus{
						Model:     modelName,
						Available: true,
					})
				}
			} else {
				statuses = append(statuses, ModelStatus{
					Model:     modelName,
					Available: true, // Model exists but couldn't parse details
				})
			}
		} else if resp.StatusCode == http.StatusNotFound {
			statuses = append(statuses, ModelStatus{
				Model:     modelName,
				Available: false,
				Error:     "Model not found",
			})
		} else {
			body, _ := io.ReadAll(resp.Body)
			statuses = append(statuses, ModelStatus{
				Model:     modelName,
				Available: false,
				Error:     fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)),
			})
		}
	}

	return statuses
}

func checkAzureModels(channel *model.Channel, models []string) []ModelStatus {
	var statuses []ModelStatus
	
	// Parse Azure config
	cfg, err := channel.LoadConfig()
	if err != nil {
		for _, model := range models {
			statuses = append(statuses, ModelStatus{
				Model:     model,
				Available: false,
				Error:     fmt.Sprintf("Failed to load config: %v", err),
			})
		}
		return statuses
	}

	baseURL := channel.GetBaseURL()
	
	// Use default API version if not configured
	apiVersion := cfg.APIVersion
	if apiVersion == "" {
		apiVersion = "2024-02-15-preview" // Default Azure OpenAI API version
	}
	
	// Use Azure's /openai/models endpoint to list available models
	url := fmt.Sprintf("%s/openai/models?api-version=%s", baseURL, apiVersion)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		for _, model := range models {
			statuses = append(statuses, ModelStatus{
				Model:     model,
				Available: false,
				Error:     fmt.Sprintf("Failed to create request: %v", err),
			})
		}
		return statuses
	}
	
	req.Header.Set("api-key", channel.Key)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		for _, model := range models {
			statuses = append(statuses, ModelStatus{
				Model:     model,
				Available: false,
				Error:     fmt.Sprintf("Failed to list models: %v", err),
			})
		}
		return statuses
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		for _, model := range models {
			statuses = append(statuses, ModelStatus{
				Model:     model,
				Available: false,
				Error:     fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)),
			})
		}
		return statuses
	}

	// Parse models response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		for _, model := range models {
			statuses = append(statuses, ModelStatus{
				Model:     model,
				Available: false,
				Error:     fmt.Sprintf("Failed to read response: %v", err),
			})
		}
		return statuses
	}

	var modelsResp struct {
		Data []struct {
			ID           string `json:"id"`
			Status       string `json:"status"`
			Capabilities struct {
				ChatCompletion bool `json:"chat_completion"`
			} `json:"capabilities"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &modelsResp); err != nil {
		for _, model := range models {
			statuses = append(statuses, ModelStatus{
				Model:     model,
				Available: false,
				Error:     fmt.Sprintf("Failed to parse models: %v", err),
			})
		}
		return statuses
	}

	// Create a map of available models
	availableModels := make(map[string]bool)
	for _, modelInfo := range modelsResp.Data {
		// Only include models that support chat completion and are succeeded
		if modelInfo.Capabilities.ChatCompletion && modelInfo.Status == "succeeded" {
			availableModels[modelInfo.ID] = true
			// Also check model name without dots for compatibility
			modelNoDots := strings.Replace(modelInfo.ID, ".", "", -1)
			availableModels[modelNoDots] = true
			
			// For models with dates (like gpt-5-mini-2025-08-07), also add base name
			if strings.Contains(modelInfo.ID, "-20") {
				// Extract base model name by removing date suffix (e.g., -2025-08-07)
				parts := strings.Split(modelInfo.ID, "-20")
				if len(parts) > 1 {
					baseModel := parts[0]
					// Remove the last part if it's a date component
					for i := len(parts) - 1; i > 0; i-- {
						baseModel = strings.TrimSuffix(baseModel, "-"+parts[i])
					}
					availableModels[baseModel] = true
				}
			}
		}
	}

	// Check each requested model
	for _, modelName := range models {
		// Check exact match first
		if availableModels[modelName] {
			statuses = append(statuses, ModelStatus{
				Model:     modelName,
				Available: true,
			})
		} else {
			// Check without dots for older naming conventions
			modelNameNoDots := strings.Replace(modelName, ".", "", -1)
			if availableModels[modelNameNoDots] {
				statuses = append(statuses, ModelStatus{
					Model:     modelName,
					Available: true,
				})
			} else {
				// Also check with Azure's naming convention (gpt-35 instead of gpt-3.5, etc)
				azureModelName := strings.Replace(modelName, "gpt-3.5", "gpt-35", 1)
				azureModelName = strings.Replace(azureModelName, "gpt-4.5", "gpt-45", 1)
				azureModelName = strings.Replace(azureModelName, "gpt-4.1", "gpt-41", 1)
				if availableModels[azureModelName] {
					statuses = append(statuses, ModelStatus{
						Model:     modelName,
						Available: true,
					})
				} else {
					statuses = append(statuses, ModelStatus{
						Model:     modelName,
						Available: false,
						Error:     "Model not found in available models list",
					})
				}
			}
		}
	}
	
	return statuses
}

func getAWSModelMap() map[string]string {
	// Copy from relay/adaptor/aws/main.go
	return map[string]string{
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
		"llama-3-8b":                 "meta.llama3-8b-instruct-v1:0",
		"llama-3-70b":                "meta.llama3-70b-instruct-v1:0",
		"llama-3.1-8b":               "meta.llama3-1-8b-instruct-v1:0",
		"llama-3.1-70b":              "meta.llama3-1-70b-instruct-v1:0",
		"llama-3.2-1b":               "meta.llama3-2-1b-instruct-v1:0",
		"llama-3.2-3b":               "meta.llama3-2-3b-instruct-v1:0",
		"llama-3.2-11b":              "meta.llama3-2-11b-instruct-v1:0",
		"llama-3.2-90b":              "meta.llama3-2-90b-instruct-v1:0",
		"llama-3.3-70b":              "meta.llama3-3-70b-instruct-v1:0",
		"llama-4-scout-17b":          "us.meta.llama4-scout-17b-instruct-v1:0",
		"llama-4-maverick-17b":       "us.meta.llama4-maverick-17b-instruct-v1:0",
	}
}

func getChannelTypeName(channelType int) string {
	switch channelType {
	case channeltype.AwsClaude:
		return "AWS Bedrock"
	case channeltype.Gemini:
		return "Google Gemini"
	case channeltype.Azure:
		return "Azure OpenAI"
	default:
		return fmt.Sprintf("Type %d", channelType)
	}
}