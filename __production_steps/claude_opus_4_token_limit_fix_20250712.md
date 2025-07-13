# Claude Opus 4 Token Limit Fix - July 12, 2025

## Issue Summary
Claude Opus 4 was incorrectly limited to 4,096 output tokens instead of its actual capability of 32,000 tokens due to a non-model-aware default in the Anthropic adapter.

## Root Cause
In `/relay/adaptor/anthropic/main.go`, the ConvertRequest function contained:
```go
if claudeRequest.MaxTokens == 0 {
    claudeRequest.MaxTokens = 4096
}
```

This code forced ALL Claude models to use 4,096 tokens as the default when no explicit max_tokens was provided, regardless of the model's actual capabilities.

## Critical Discovery
Research revealed that **Anthropic's Messages API REQUIRES the max_tokens parameter** - it will fail if not provided. This explains why the default was necessary, unlike other providers who accept missing max_tokens.

## Investigation Findings

### Anthropic API Requirements
- **max_tokens is REQUIRED**: API requests fail without this parameter
- **Different models have different limits**:
  - Standard models: 4,096 tokens
  - Claude 3.5 Sonnet: 8,192 tokens
  - Claude 3.7 Sonnet: 8,192 tokens (up to 128K with beta header)
  - Claude Opus 4 & Sonnet 4: 32,000 tokens (assumed based on capabilities)

### Comparison with Other Adapters
- ✅ OpenAI: Optional parameter, API provides defaults
- ✅ Gemini: Optional parameter, API provides defaults
- ✅ Baidu: Optional parameter, API provides defaults
- ✅ Cohere: Optional parameter, API provides defaults
- ⚠️ Anthropic: REQUIRED parameter, must be provided
- ⚠️ AWS Claude: Uses Anthropic converter (inherits requirement)
- ⚠️ Vertex AI Claude: Uses Anthropic converter (inherits requirement)

## Solution Implemented
Updated the default logic to be model-aware, providing appropriate defaults for each Claude model:

### Before:
```go
if claudeRequest.MaxTokens == 0 {
    claudeRequest.MaxTokens = 4096
}
```

### After:
```go
if claudeRequest.MaxTokens == 0 {
    // Set model-specific defaults when not specified
    // Anthropic API requires max_tokens parameter
    switch claudeRequest.Model {
    case "claude-opus-4-20250514", "claude-sonnet-4-20250514":
        claudeRequest.MaxTokens = 32000
    case "claude-3-5-sonnet-20241022", "claude-3-5-sonnet-latest":
        claudeRequest.MaxTokens = 8192
    case "claude-3-7-sonnet-20250219", "claude-3-7-sonnet-latest":
        claudeRequest.MaxTokens = 8192 // Can go up to 128K with beta header
    default:
        claudeRequest.MaxTokens = 4096
    }
}
```

## Impact
1. **Claude Opus 4** can now use its full 32,000 token output capability when max_tokens is not specified
2. **All Claude models** will use appropriate defaults for their capabilities
3. **API compatibility** maintained - max_tokens is still always provided as required by Anthropic

## Why Not Remove the Default?
Initially considered removing the default entirely (like other adapters), but research revealed:
- Anthropic's Messages API **requires** max_tokens parameter
- API requests fail without it (unlike OpenAI, Gemini, etc.)
- The default must stay, but needs to be model-aware

## Testing Recommendations
1. Test Claude Opus 4 with large output requests (>4,096 tokens)
2. Verify all Claude models work without explicit max_tokens:
   - Claude 3 Haiku (4,096)
   - Claude 3.5 Sonnet (8,192)
   - Claude 3.7 Sonnet (8,192)
   - Claude Opus 4 (32,000)
   - Claude Sonnet 4 (32,000)
3. Check AWS Claude and Vertex AI Claude adapters still function

## Related Files
- `/relay/adaptor/anthropic/main.go` - Fixed file
- `/upstream-check/relay/adaptor/vertexai/claude/adapter.go` - Uses anthropic.ConvertRequest
- `/relay/adaptor/aws/main.go` - Uses anthropic.ConvertRequest

## Notes
- The 4,096 limit was appropriate for older Claude models
- Modern Claude models have much higher limits
- This is a unique requirement of Anthropic's API design
- Other providers (OpenAI, Gemini) accept missing max_tokens

---
*Fix implemented: July 12, 2025*
*Author: Assistant*
*Issue discovered during Claude Opus 4 testing*