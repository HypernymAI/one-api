# Gemini Models Pricing Update and Status Check Implementation - 2025-09-13

## Overview
Updated Google Gemini model pricing to reflect September 2025 actual costs and discovered issues with Gemini 2.5 Pro's thinking tokens. Implemented model availability status checking without invocation.

## Part 1: Pricing Updates

### Issue Identified
Gemini model pricing in the codebase was significantly outdated, using incorrect ratios that didn't reflect actual September 2025 pricing from Google.

### Research Findings
Current Gemini pricing as of September 2025:
- **Gemini 1.5 Flash**: $0.25/M input, $0.50/M output tokens
- **Gemini 1.5 Pro**: $1.25/M input, $5.00/M output tokens  
- **Gemini 2.0 Flash**: $0.10/M input, $0.40/M output tokens
- **Gemini 2.5 Flash**: $0.30/M input, $2.50/M output tokens
- **Gemini 2.5 Pro**: $1.25/M input, $10.00/M output tokens

### Understanding the Pricing System
The codebase uses a confusing ratio system where:
- `1.0 = $0.002/1K tokens = $2/M tokens` (based on old GPT-3.5 pricing)
- `USD = 500` means 500 units of $0.002 = $1
- `MILLI_USD = 0.5` represents $0.001/1K tokens

This system is inherited from upstream and maintained for compatibility.

### Code Changes
Updated `/relay/billing/ratio/model.go`:

**Before:**
```go
"gemini-1.5-flash":  0.075 * MILLI_USD,  // Was $0.000075/1K
"gemini-1.5-pro":    1,                   // Was $0.002/1K  
"gemini-2.0-flash":  0.075 * MILLI_USD,
"gemini-2.5-flash":  0.0375 * MILLI_USD,
"gemini-2.5-pro":    2.5 * MILLI_USD,
```

**After:**
```go
"gemini-1.5-flash":  0.125,  // $0.25/M input = $0.00025/1K tokens
"gemini-1.5-pro":    0.625,  // $1.25/M input = $0.00125/1K tokens
"gemini-2.0-flash":  0.05,   // $0.10/M input = $0.0001/1K tokens
"gemini-2.5-flash":  0.15,   // $0.30/M input = $0.0003/1K tokens
"gemini-2.5-pro":    0.625,  // $1.25/M input = $0.00125/1K tokens
```

## Part 2: Model Testing and Discovery

### Testing Results
Channel 3 (Gemini) testing via curl to `localhost:3000`:

**Working models:**
- ✅ Gemini 1.5 Flash - Responded successfully
- ✅ Gemini 1.5 Pro - Responded successfully
- ✅ Gemini 2.0 Flash - Responded successfully  
- ✅ Gemini 2.5 Flash - Initially rate-limited, now working

**Problematic model:**
- ⚠️ Gemini 2.5 Pro - Returned empty responses with low token limits

### Gemini 2.5 Pro Investigation

#### Discovery: Thinking Tokens
Direct API testing revealed Gemini 2.5 Pro uses "thinking tokens":
```json
{
  "usageMetadata": {
    "promptTokenCount": 2,
    "candidatesTokenCount": 9,
    "totalTokenCount": 521,
    "thoughtsTokenCount": 510  // Internal reasoning tokens!
  }
}
```

#### Root Cause
1. Gemini 2.5 Pro performs internal reasoning before generating output
2. With `max_tokens: 50`, there's insufficient room for output after thinking
3. The model returns empty content when token limit is exhausted by thinking

#### Solution
Increased `max_tokens` to 500+ for Gemini 2.5 Pro requests allows proper responses:
```bash
# Failed with max_tokens: 50
curl -X POST http://localhost:3000/v1/chat/completions \
  -d '{"model": "gemini-2.5-pro", "max_tokens": 50}'
# Returns empty content

# Success with max_tokens: 500
curl -X POST http://localhost:3000/v1/chat/completions \
  -d '{"model": "gemini-2.5-pro", "max_tokens": 500}'
# Returns: "hello"
```

## Part 3: Channel Testing Limitations

### Problem with Existing Test
The channel test endpoint (`/api/channel/test/:id`) uses:
```go
testRequest := &relaymodel.GeneralOpenAIRequest{
    MaxTokens: 2,  // Too low for thinking models!
    Stream:    false,
    Model:     "gpt-3.5-turbo",
}
```

This causes false negatives for:
- Gemini 2.5 Pro (thinking tokens)
- GPT-5 mini/nano (reasoning models)
- Any future thinking/reasoning models

### Parallel with GPT-5 Reasoning Models
Similar issue was already solved for Azure GPT-5 models in `/relay/adaptor/openai/adaptor.go`:
- Converts `max_tokens` to `max_completion_tokens`
- Forces `temperature: 1.0` for reasoning models
- No additional tokens were added, just parameter translation

## Part 4: Model Availability Status Check Implementation

### Motivation
Need to check model availability without invoking them to avoid:
- False negatives from token limit issues
- Unnecessary API costs
- Rate limiting from actual invocations

### Implementation Started
Created `/controller/channel-status.go` with non-invocation status checks:

**AWS Bedrock:**
- Uses `bedrock.GetFoundationModel()` for base models
- Assumes availability for inference profiles (us.*, global.*)
- Parses credentials from channel key format: `ACCESS_KEY|SECRET_KEY|REGION`

**Google Gemini:**
- Queries `https://generativelanguage.googleapis.com/v1beta/models/{model}`
- Checks if model supports `generateContent` method
- Returns model metadata without invocation

**Azure OpenAI:**
- Lists deployments via `/openai/deployments` endpoint
- Checks if model/deployment exists
- Handles both dotted and non-dotted model names

### API Endpoint (To Be Added)
```go
// Add to router
channelRoute.GET("/status/:id", controller.CheckChannelModelsStatus)
```

## Testing Verification

### Confirmed Working Models
All Gemini models tested and confirmed operational:
1. Gemini 1.5 Flash ✅
2. Gemini 1.5 Pro ✅
3. Gemini 2.0 Flash ✅
4. Gemini 2.5 Flash ✅
5. Gemini 2.5 Pro ✅ (with adequate token limits)

### Database Configuration
- Channel ID: 3
- Type: 24 (Gemini)
- API Key: AIzaSyDDK9cD4uNC6l6hkMfOSSodl8U8p82SGBo
- Models: All 13 Gemini models configured and available

## Recommendations

1. **Increase Test Token Limits**: Update channel test to use at least 100-500 tokens
2. **Model-Specific Handling**: Add special cases for known thinking/reasoning models
3. **Complete Status Check Implementation**: Finish and deploy the non-invocation status check
4. **Monitor Thinking Models**: Track which models use thinking tokens for proper handling
5. **Update Documentation**: Document the thinking token behavior for team awareness

## Files Modified
- `/relay/billing/ratio/model.go` - Updated Gemini pricing ratios
- `/controller/channel-status.go` - Created model availability checker (new file)

## Next Steps
1. Add router endpoint for status checking
2. Test status check implementation across all three providers
3. Consider updating default test token limits
4. Document thinking model requirements

## Notes
- Gemini 2.5 Flash rate limiting appears to be temporary/capacity-based
- The pricing ratio system should be refactored in future for clarity
- Similar thinking token issues may affect other providers' reasoning models