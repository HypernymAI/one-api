# Llama 4 Inference Profiles Implementation - Production Log
## Date: 2025-09-06

### Summary
Successfully implemented Llama 4 models (Scout and Maverick) using AWS Bedrock inference profiles. These models require special handling different from standard on-demand models.

### Key Discovery: Inference Profiles Required

AWS Bedrock has introduced a new access pattern for newer models:
- **On-Demand**: Traditional model access (Llama 3.x models)
- **Inference Profiles**: Required for Llama 4, Claude 4, and other new models

### Implementation Changes

#### Model ID Mapping Update
```go
// Before (didn't work)
"llama-4-scout-17b":    "meta.llama4-scout-17b-instruct-v1:0",
"llama-4-maverick-17b": "meta.llama4-maverick-17b-instruct-v1:0",

// After (WORKING)
"llama-4-scout-17b":    "us.meta.llama4-scout-17b-instruct-v1:0",
"llama-4-maverick-17b": "us.meta.llama4-maverick-17b-instruct-v1:0",
```

The critical change: prefix with `us.` to use the inference profile instead of direct model access.

### Available Inference Profiles

```bash
# List inference profiles
aws bedrock list-inference-profiles --region us-east-1 | jq -r '.inferenceProfileSummaries[] | .inferenceProfileName'

# Llama 4 profiles:
US Llama 4 Scout 17B Instruct: us.meta.llama4-scout-17b-instruct-v1:0
US Llama 4 Maverick 17B Instruct: us.meta.llama4-maverick-17b-instruct-v1:0
```

### Working Models Status

#### AWS Bedrock Channel (ID: 10)

**Working with On-Demand:**
- ✅ llama-3-8b
- ✅ llama-3-70b
- ✅ llama-3.1-8b
- ✅ llama-3.1-70b
- ✅ llama-3.2-1b
- ✅ llama-3.2-3b
- ✅ llama-3.2-11b
- ✅ llama-3.2-90b
- ✅ llama-3.3-70b

**Working with Inference Profiles (us. prefix):**
- ✅ llama-4-scout-17b (17B params, 16 experts)
- ✅ llama-4-maverick-17b (17B active, 400B total, 128 experts)

**Require Inference Profiles (not accessible without setup):**
- ❌ claude-opus-4-20250514
- ❌ claude-opus-4-1-20250805
- ❌ claude-3-7-sonnet-20250219
- ❌ claude-sonnet-4-20250514

### Testing Commands

```bash
# Test Llama 4 Scout
curl -s -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer sk-YOUR-KEY-10" \
  -H "Content-Type: application/json" \
  -d '{"model": "llama-4-scout-17b", "messages": [{"role": "user", "content": "Hi"}], "max_tokens": 10}'

# Test Llama 4 Maverick
curl -s -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer sk-YOUR-KEY-10" \
  -H "Content-Type: application/json" \
  -d '{"model": "llama-4-maverick-17b", "messages": [{"role": "user", "content": "Hi"}], "max_tokens": 10}'
```

### Critical Implementation Notes

1. **Inference Profile Format**: The Go SDK handles inference profiles transparently, but AWS CLI requires special formatting that remains undocumented
2. **No Streaming**: Llama 4 models don't support streaming (InvokeModelWithResponseStream)
3. **Request Format**: Uses same Llama prompt template as Llama 3
4. **Region Specific**: Inference profiles are region-specific (us-east-1)

### Model-Specific API Key Filtering

Implemented channel-specific model filtering for `/v1/models` endpoint:
- Admin keys with suffix (e.g., `-10`) return only that channel's models
- Without suffix, returns all available models based on user permissions

```go
// Added to controller/model.go
if specificChannelId != "" {
    channel, err := model.GetChannelById(channelIdInt, false)
    if err == nil && channel.Models != "" {
        availableModels = strings.Split(channel.Models, ",")
    }
}
```

### Lessons Learned

1. **Inference Profiles vs On-Demand**: Critical distinction for new models
2. **Model ID Prefixes**: `us.` prefix indicates inference profile usage
3. **AWS CLI Limitations**: CLI behavior differs from SDK for inference profiles
4. **Testing Importance**: Always test through actual API, not just AWS CLI

### Files Modified

- `/relay/adaptor/aws/main.go` - Added Llama 4 and Claude 4 model mappings
- `/controller/model.go` - Added channel-specific model filtering

### Next Actions

1. Monitor AWS for additional inference profile regions
2. Document inference profile requirements for users
3. Consider adding inference profile auto-detection

---
Generated: 2025-09-06 05:30 UTC