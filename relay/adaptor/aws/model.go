package aws

import "github.com/songquanpeng/one-api/relay/adaptor/anthropic"

// Request is the request to AWS Claude
//
// https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-anthropic-claude-messages.html
type Request struct {
	// AnthropicVersion should be "bedrock-2023-05-31"
	AnthropicVersion string              `json:"anthropic_version"`
	Messages         []anthropic.Message `json:"messages"`
	MaxTokens        int                 `json:"max_tokens,omitempty"`
	Temperature      float64             `json:"temperature,omitempty"`
	TopP             float64             `json:"top_p,omitempty"`
	TopK             int                 `json:"top_k,omitempty"`
	StopSequences    []string            `json:"stop_sequences,omitempty"`
}

// LlamaRequest is the request to AWS Llama models
//
// https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-meta.html
type LlamaRequest struct {
	Prompt     string   `json:"prompt"`
	MaxGenLen  int      `json:"max_gen_len,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	TopP       float64  `json:"top_p,omitempty"`
}

// LlamaResponse is the response from AWS Llama models
type LlamaResponse struct {
	Generation           string `json:"generation"`
	PromptTokenCount     int    `json:"prompt_token_count"`
	GenerationTokenCount int    `json:"generation_token_count"`
	StopReason           string `json:"stop_reason"`
}

// LlamaStreamResponse is the streaming response from AWS Llama models
type LlamaStreamResponse struct {
	Generation           string `json:"generation"`
	PromptTokenCount     int    `json:"prompt_token_count"`
	GenerationTokenCount int    `json:"generation_token_count"`
	StopReason           string `json:"stop_reason,omitempty"`
}
