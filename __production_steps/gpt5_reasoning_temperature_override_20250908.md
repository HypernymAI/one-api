# GPT-5 Reasoning Models Temperature Override
**Date**: 2025-09-08  
**Author**: Claude Code  
**Impact**: Critical - Enables GPT-5 reasoning models to work with any temperature value

## Problem Statement

GPT-5 reasoning models (gpt-5-mini and gpt-5-nano) only accept `temperature=1.0`. Any other temperature value causes the request to fail. This breaks v1 API compatibility where clients expect to control temperature.

## Solution

Override the temperature to 1.0 for GPT-5 reasoning models only, while preserving the user's temperature for all other models.

## Implementation

**Modified**: `/relay/adaptor/openai/adaptor.go`

### Code Changes

```go
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
        // ... existing struct definition ...
        
        azureReq := AzureReasoningRequest{
            Messages:            request.Messages,
            Model:               request.Model,
            FrequencyPenalty:    request.FrequencyPenalty,
            MaxCompletionTokens: request.MaxTokens,
            N:                   request.N,
            PresencePenalty:     request.PresencePenalty,
            ResponseFormat:      request.ResponseFormat,
            Seed:                request.Seed,
            Stream:              request.Stream,
            Temperature:         1.0, // REQUIRED: GPT-5 reasoning models only accept temperature=1.0
            TopP:                request.TopP,
            // ... rest of fields ...
        }
        
        return azureReq, nil
    }
}

return request, nil  // All other models use original request unchanged
```

## Scope and Safety

The temperature override is **ONLY** applied when ALL conditions are met:
1. Channel type is Azure (`a.ChannelType == channeltype.Azure`)
2. Model is exactly `gpt-5-mini` OR `gpt-5-nano`

All other models (including GPT-4.1, Claude, Llama, etc.) continue to use their requested temperature values unchanged.

## Testing Verification

Tested models with various temperatures:
- **GPT-4.1-mini with temperature=0.3**: ✅ Uses 0.3 as requested
- **Claude-3-haiku with temperature=0.7**: ✅ Uses 0.7 as requested  
- **GPT-5-mini with temperature=0.3**: ✅ Works (internally uses 1.0)
- **GPT-5-nano with temperature=0.7**: ✅ Works (internally uses 1.0)

## Deployment

1. Pull latest changes
2. Run build: `./build.sh`
3. Restart service:
   ```bash
   pkill -f one-api
   screen -dmS oneapi bash -c './one-api-en 2>&1 | tee -a oneapi.log'
   ```

## Notes

- This is a temporary compatibility fix until the industry standardizes v2 API for reasoning models
- The override is transparent to clients - they can send any temperature value
- Works in conjunction with the `max_completion_tokens` translation for full v1 API compatibility