# Understanding max_tokens vs max_completion_tokens

## Overview
OpenAI and other providers have introduced a new parameter `max_completion_tokens` for reasoning models, replacing the traditional `max_tokens` parameter. This document explains the differences and which models require which parameter.

## The Fundamental Difference

### max_tokens (Traditional Parameter)
- **Purpose**: Limits the total number of tokens in the model's response
- **Behavior**: Direct limit on output length
- **Models**: Used by standard chat models (GPT-3.5, GPT-4, etc.)
- **Token Usage**: All tokens go directly to the visible response

### max_completion_tokens (Reasoning Model Parameter)
- **Purpose**: Limits the total tokens available for BOTH reasoning and response
- **Behavior**: Must account for internal reasoning tokens + output tokens
- **Models**: Required by reasoning models (o1, o3, o4, GPT-5 mini/nano)
- **Token Usage**: Split between invisible reasoning process and visible response

## Why the Change?

Reasoning models perform internal "thinking" before generating responses:
1. They analyze the problem internally
2. Generate reasoning tokens (invisible to users)
3. Then produce the final response

Example allocation for a reasoning model with `max_completion_tokens: 100`:
- 60 tokens used for internal reasoning (invisible)
- 40 tokens available for the actual response (visible)
- Total: 100 tokens

## Model Requirements

### Models Using `max_tokens`
Standard models that generate responses directly:
- ✅ gpt-3.5-turbo (all variants)
- ✅ gpt-4 (all variants)
- ✅ gpt-4-turbo
- ✅ gpt-4o
- ✅ gpt-5-chat
- ✅ claude-* (all models)
- ✅ gemini-* (all models)
- ✅ Most other traditional chat models

### Models Requiring `max_completion_tokens`
Reasoning models that perform internal analysis:
- ✅ o1 (all variants)
- ✅ o1-preview
- ✅ o1-mini
- ✅ o3
- ✅ o3-mini
- ✅ o4-mini
- ✅ gpt-5-mini
- ✅ gpt-5-nano

## API Error Examples

### Wrong Parameter Error
```json
// Request with max_tokens to reasoning model
{
  "model": "o3-mini",
  "messages": [{"role": "user", "content": "Hello"}],
  "max_tokens": 100  // ❌ WRONG
}

// Error Response
{
  "error": {
    "message": "Unsupported parameter: 'max_tokens' is not supported with this model. Use 'max_completion_tokens' instead.",
    "type": "invalid_request_error",
    "param": "max_tokens",
    "code": "unsupported_parameter"
  }
}
```

### Correct Usage
```json
// Correct request to reasoning model
{
  "model": "o3-mini",
  "messages": [{"role": "user", "content": "Hello"}],
  "max_completion_tokens": 100  // ✅ CORRECT
}
```

## Important Considerations

### 1. Allocate Sufficient Tokens for Reasoning Models
Reasoning models need tokens for both thinking and responding:
- ❌ `max_completion_tokens: 10` - Too small, may result in empty response
- ✅ `max_completion_tokens: 100` - Minimum recommended
- ✅ `max_completion_tokens: 1000` - Good for most use cases

### 2. Token Usage Breakdown
When a reasoning model uses `max_completion_tokens: 1000`:
```json
{
  "usage": {
    "completion_tokens": 1000,
    "completion_tokens_details": {
      "reasoning_tokens": 850,    // Internal reasoning (invisible)
      "audio_tokens": 0,
      "accepted_prediction_tokens": 0,
      "rejected_prediction_tokens": 0
    }
  }
}
```
Only 150 tokens were used for the visible response!

### 3. Cost Implications
- Reasoning tokens are billed the same as output tokens
- A response using 1000 completion tokens costs the same whether:
  - 1000 tokens of visible text (standard model)
  - 900 reasoning + 100 visible tokens (reasoning model)

## Migration Guide

### For Application Developers
```javascript
// Old code for all models
const params = {
  model: modelName,
  messages: messages,
  max_tokens: 100
};

// New code with model detection
const params = {
  model: modelName,
  messages: messages
};

// List of reasoning models
const reasoningModels = ['o1', 'o1-mini', 'o3', 'o3-mini', 'o4-mini', 'gpt-5-mini', 'gpt-5-nano'];

if (reasoningModels.some(m => modelName.includes(m))) {
  params.max_completion_tokens = 1000;  // Reasoning models
} else {
  params.max_tokens = 100;              // Standard models
}
```

### For one-api Users
No changes needed! one-api passes parameters through to the underlying API, which will:
- Accept `max_tokens` for standard models
- Reject `max_tokens` and require `max_completion_tokens` for reasoning models

## Common Issues and Solutions

### Issue: Empty Responses from Reasoning Models
**Cause**: Insufficient tokens allocated
```json
{
  "max_completion_tokens": 50  // Too small!
}
```
**Solution**: Increase token allocation
```json
{
  "max_completion_tokens": 500  // Better
}
```

### Issue: API Rejects max_tokens
**Cause**: Using `max_tokens` with reasoning model
**Solution**: Switch to `max_completion_tokens`

### Issue: Unexpected Token Usage
**Cause**: Not accounting for reasoning tokens
**Solution**: Monitor `reasoning_tokens` in usage response and adjust accordingly

## Best Practices

1. **Default Token Allocations**
   - Standard models: `max_tokens: 500`
   - Reasoning models: `max_completion_tokens: 1000`

2. **Model Detection**
   - Implement model-aware parameter selection
   - Maintain a list of reasoning models
   - Handle API errors gracefully

3. **Cost Management**
   - Monitor reasoning token usage
   - Set appropriate limits based on use case
   - Consider using standard models when reasoning isn't needed

## Summary

The shift from `max_tokens` to `max_completion_tokens` reflects the fundamental difference in how reasoning models operate. While standard models generate responses directly, reasoning models need token allocation for both thinking and responding. Understanding this difference is crucial for effectively using modern AI models.

### Quick Reference
- **Standard models**: Use `max_tokens`
- **Reasoning models**: Use `max_completion_tokens` with higher values
- **Minimum recommendation**: 1000 tokens for reasoning models
- **Error handling**: Catch and handle parameter rejection errors

---
*Last updated: August 26, 2025*
*Applies to: OpenAI API, Azure OpenAI, and compatible services*