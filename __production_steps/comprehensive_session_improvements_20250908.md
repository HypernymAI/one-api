# Comprehensive Session Improvements - Production Log
## Date: 2025-09-08

### Summary
Major session implementing multiple critical improvements: Anthropic Opus 4.1 support, API key channel filtering, AWS Bedrock Llama 4 inference profiles, Azure GPT-4.1 dot handling, multi-resource Azure support, and reasoning model parameter translation for v1 API compatibility.

## 1. Claude Opus 4.1 Support Verification

### Issue
Reports that Claude Opus 4.1 wasn't working on channel 2 (Anthropic direct).

### Investigation
- Verified Opus 4.1 is available in Anthropic's API: `claude-opus-4-1-20250805`
- Tested successfully through direct API and localhost:3000
- Found missing billing ratio

### Solution
Added billing ratio to `/relay/billing/ratio/model.go`:
```go
"claude-opus-4-1-20250805": 15.0 / 1000 * USD,
```

### Result
✅ Claude Opus 4.1 fully functional on channel 2

## 2. API Key Channel-Specific Model Filtering

### Issue
Admin API keys with channel suffixes (e.g., `-10`) were returning all models instead of channel-specific models.

### Implementation
Modified `/controller/model.go` ListModels function:
```go
// Check if admin is using a specific channel suffix
specificChannelId := c.GetString(ctxkey.SpecificChannelId)
if specificChannelId != "" {
    channel, err := model.GetChannelById(channelIdInt, false)
    if err == nil && channel.Models != "" {
        availableModels = strings.Split(channel.Models, ",")
    }
}
```

### Result
- ✅ `-10` suffix returns only AWS Bedrock models
- ✅ `-2` suffix returns only Anthropic models
- ✅ No suffix returns all available models

## 3. AWS Bedrock Llama 4 Inference Profiles

### Issue
Llama 4 models (Scout, Maverick) failing with "on-demand throughput not supported" errors.

### Discovery
AWS now requires inference profiles for newer models:
- Inference Profile IDs: `us.meta.llama4-scout-17b-instruct-v1:0`
- Regular Model IDs: `meta.llama4-scout-17b-instruct-v1:0`

### Solution
Updated `/relay/adaptor/aws/main.go` model mappings:
```go
"llama-4-scout-17b":    "us.meta.llama4-scout-17b-instruct-v1:0",
"llama-4-maverick-17b": "us.meta.llama4-maverick-17b-instruct-v1:0",
```

Also added Claude 4 and 3.7 models:
```go
"claude-opus-4-20250514":     "anthropic.claude-opus-4-20250514-v1:0",
"claude-opus-4-1-20250805":   "anthropic.claude-opus-4-1-20250805-v1:0",
"claude-3-7-sonnet-20250219": "anthropic.claude-3-7-sonnet-20250219-v1:0",
"claude-sonnet-4-20250514":   "anthropic.claude-sonnet-4-20250514-v1:0",
```

### Result
- ✅ Llama 4 Scout working
- ✅ Llama 4 Maverick working
- ❌ Claude 4/3.7 require inference profiles (billing issue)

## 4. Azure GPT-4.1 Dot Handling Fix

### Issue
GPT-4.1 models failing because code was removing dots from model names (`gpt-4.1-mini` → `gpt-41-mini`).

### Root Cause
Legacy code from upstream (commit 1aa374cc, Feb 2024) removed dots because old Azure didn't support them.

### Solution
Implemented conditional dot removal in `/relay/adaptor/openai/adaptor.go`:
```go
// Only remove dots for API versions before 2025
// Newer Azure deployments (2025+) support and require dots for models like gpt-4.1
if len(meta.Config.APIVersion) >= 4 && meta.Config.APIVersion[:4] < "2025" {
    model_ = strings.Replace(model_, ".", "", -1)
}
```

### Result
- ✅ GPT-4.1 models work with 2025-01-01-preview API version
- ✅ Legacy models still work with older API versions

## 5. Multi-Resource Azure Support (Channel 11)

### Issue
GPT-4.1-nano deployed on different Azure resource: `https://chris-mfagl5hd-westus.cognitiveservices.azure.com/`

### Solution
Created Channel 11:
- Name: `azure_westus_gpt41nano`
- Base URL: `https://chris-mfagl5hd-westus.cognitiveservices.azure.com`
- API Key: Separate westus resource key
- API Version: `2025-01-01-preview`
- Models: `gpt-4.1-nano`

### Result
✅ GPT-4.1-nano accessible via `-11` suffix

## 6. Reasoning Models Parameter Translation

### Issue
GPT-5-mini and GPT-5-nano require `max_completion_tokens` instead of `max_tokens`, breaking v1 API compatibility.

### Root Cause Analysis
1. Azure channels have APIType = OpenAI (not in ToAPIType switch)
2. OpenAI type channels skip ConvertRequest function
3. No translation occurring for reasoning models

### Solution (Two Parts)

#### Part 1: Force conversion for Azure reasoning models
Modified `/relay/controller/text.go`:
```go
// Azure channels need special handling for reasoning models
needsConversion := meta.ChannelType == channeltype.Azure && 
    (textRequest.Model == "gpt-5-mini" || textRequest.Model == "gpt-5-nano")

if meta.APIType == apitype.OpenAI && meta.ChannelType != channeltype.GoogleOpenAI && !needsConversion {
    // Skip conversion except for Azure reasoning models
```

#### Part 2: Custom request struct for reasoning models
Modified `/relay/adaptor/openai/adaptor.go` ConvertRequest:
```go
if modelName == "gpt-5-mini" || modelName == "gpt-5-nano" {
    type AzureReasoningRequest struct {
        // ... fields without max_tokens
        MaxCompletionTokens int `json:"max_completion_tokens,omitempty"`
        // ... other fields
    }
    
    azureReq := AzureReasoningRequest{
        MaxCompletionTokens: request.MaxTokens, // Translate
        // ... copy other fields
    }
    return azureReq, nil
}
```

### Result
- ✅ GPT-5-mini accepts standard `max_tokens`
- ✅ GPT-5-nano accepts standard `max_tokens`
- ✅ Full v1 API compatibility maintained
- ✅ Other models unaffected

## Working Models Summary

### Channel 8 (Azure Main Resource)
- ✅ gpt-5-chat (standard model)
- ✅ gpt-5-mini (reasoning model, translated)
- ✅ gpt-5-nano (reasoning model, translated)
- ✅ gpt-4.1 (with dots preserved)
- ✅ gpt-4.1-mini (with dots preserved)

### Channel 10 (AWS Bedrock)
- ✅ All Llama 3.x models
- ✅ llama-4-scout-17b (inference profile)
- ✅ llama-4-maverick-17b (inference profile)
- ⚠️ Claude models (limited by billing)

### Channel 11 (Azure West US)
- ✅ gpt-4.1-nano

### Channel 2 (Anthropic Direct)
- ✅ claude-opus-4-20250514
- ✅ claude-opus-4-1-20250805
- ✅ All other Claude models

## Technical Insights

### Inference Profiles vs On-Demand
- On-demand: Direct model access (older models)
- Inference profiles: Required for newer models (Llama 4, Claude 4)
- Identified by `us.` prefix in model ID

### Azure API Version Significance
- Pre-2025: Remove dots from model names
- 2025+: Preserve dots (required for GPT-4.1)

### Reasoning Models Architecture
- Require `max_completion_tokens` parameter
- Use reasoning tokens in usage tracking
- Need sufficient tokens (200+) to generate responses

## Files Modified
1. `/relay/billing/ratio/model.go` - Added Opus 4.1 billing
2. `/controller/model.go` - Channel-specific model filtering
3. `/relay/adaptor/aws/main.go` - Llama 4 and Claude 4 model mappings
4. `/relay/adaptor/openai/adaptor.go` - Conditional dot removal and reasoning model translation
5. `/relay/controller/text.go` - Force conversion for Azure reasoning models

## Database Changes
- Channel 8: Updated API version to 2025-01-01-preview
- Channel 10: Model list updated via UI
- Channel 11: New channel created for westus resource

## Lessons Learned
1. **Never modify database directly** - Always use UI
2. **API version impacts behavior** - Azure API versions affect feature support
3. **Upstream code may have legacy constraints** - Review git history for context
4. **Channel type vs API type** - Important distinction affecting request flow
5. **Test with actual API calls** - AWS CLI behavior differs from SDK

---
Generated: 2025-09-08 02:00 UTC