# Google OpenAI Implementation - 2025-08-22

## Overview
Successfully implemented Google OpenAI (Vertex AI) support for the one-api service, including proper model name handling and multi-region support.

## Key Findings

### Model Name Handling
- **Azure OpenAI**: Removes periods from model names (`meta/llama-3.1-8b` → `meta/llama-31-8b`)
- **Google OpenAI**: Keeps periods in model names (`meta/llama-3.1-8b` remains `meta/llama-3.1-8b`)

### Regional Model Availability
Google Cloud has different Llama model availability by region:

**us-central1 region:**
- `meta/llama-3.1-8b-instruct-maas` ✅
- `meta/llama-3.1-405b-instruct-maas` ✅  
- `meta/llama-3.3-70b-instruct-maas` ✅
- `meta/llama-4-maverick-17b-128e-instruct-maas` ❌
- `meta/llama-4-scout-17b-16e-instruct-maas` ❌

**us-east5 region:**
- `meta/llama-3.1-8b-instruct-maas` ❌
- `meta/llama-3.1-405b-instruct-maas` ❌
- `meta/llama-3.3-70b-instruct-maas` ❌
- `meta/llama-4-maverick-17b-128e-instruct-maas` ✅
- `meta/llama-4-scout-17b-16e-instruct-maas` ✅

## Implementation Changes

### 1. Channel Type Definition
Already existed in `relay/channeltype/define.go`:
```go
GoogleOpenAI = 44
```

### 2. Request Processing Fix
Modified `relay/controller/text.go` to ensure GoogleOpenAI requests go through `ConvertRequest`:

```go
// OLD: Only called ConvertRequest for non-OpenAI types
if meta.APIType == apitype.OpenAI {
    // Skip ConvertRequest
}

// NEW: Exclude GoogleOpenAI from the skip
if meta.APIType == apitype.OpenAI && meta.ChannelType != channeltype.GoogleOpenAI {
    // Skip ConvertRequest for regular OpenAI, but not GoogleOpenAI
}
```

### 3. URL Construction
Enhanced `relay/adaptor/openai/adaptor.go` with GoogleOpenAI URL handling:

```go
case channeltype.GoogleOpenAI:
    // https://${ENDPOINT}/v1/projects/${PROJECT_ID}/locations/${REGION}/endpoints/openapi/chat/completions
    task := strings.TrimPrefix(meta.RequestURLPath, "/v1/")
    requestURL := fmt.Sprintf("/v1/projects/%s/locations/%s/endpoints/openapi/%s", meta.Config.ProjectID, meta.Config.Region, task)
    return GetFullRequestURL(meta.BaseURL, requestURL, meta.ChannelType), nil
```

### 4. Model Compatibility
Added GoogleOpenAI to compatible channels in `relay/adaptor/openai/compatible.go`:

```go
var CompatibleChannels = []int{
    channeltype.Azure,
    channeltype.GoogleOpenAI,  // Already existed
    // ... other channels
}

func GetCompatibleChannelMeta(channelType int) (string, []string) {
    switch channelType {
    case channeltype.GoogleOpenAI:
        return "google", googleopenai.ModelList
    // ... other cases
    }
}
```

## Testing Results

### Direct Google Cloud API Tests
```bash
# us-central1 (working)
curl -X POST https://us-central1-aiplatform.googleapis.com/v1/projects/hypernym-api/locations/us-central1/endpoints/openapi/chat/completions \
  -H "Authorization: Bearer $(gcloud auth print-access-token)" \
  -H "Content-Type: application/json" \
  -d '{"model":"meta/llama-3.1-8b-instruct-maas","messages":[{"role":"user","content":"hi"}]}'
# Result: Success ✅

# us-east5 (not working for 3.1 models)  
curl -X POST https://us-east5-aiplatform.googleapis.com/v1/projects/hypernym-api/locations/us-east5/endpoints/openapi/chat/completions \
  -H "Authorization: Bearer $(gcloud auth print-access-token)" \
  -H "Content-Type: application/json" \
  -d '{"model":"meta/llama-3.1-8b-instruct-maas","messages":[{"role":"user","content":"hi"}]}'
# Result: 404 Not Found ❌
```

### One-API Service Tests
```bash
# After fixes - both regions working through separate channels
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer sk-BoroIrY7uCruh3IR48Cb5b151f8943E88bF86e3d91Ee2bB2" \
  -H "Content-Type: application/json" \
  -d '{"model":"meta/llama-3.1-8b-instruct-maas","messages":[{"role":"user","content":"hi"}]}'
# Result: Success ✅ (routes to Channel #6 - us-central1)

curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer sk-BoroIrY7uCruh3IR48Cb5b151f8943E88bF86e3d91Ee2bB2" \
  -H "Content-Type: application/json" \
  -d '{"model":"meta/llama-4-maverick-17b-128e-instruct-maas","messages":[{"role":"user","content":"hi"}]}'
# Result: Success ✅ (routes to Channel #7 - us-east5)
```

## Channel Configuration

### Channel #6 (us-central1)
- **Type**: GoogleOpenAI (40)
- **Base URL**: `https://us-central1-aiplatform.googleapis.com`
- **Project ID**: `hypernym-api`
- **Region**: `us-central1`
- **Models**: `meta/llama-3.1-8b-instruct-maas,meta/llama-3.1-405b-instruct-maas,meta/llama-3.3-70b-instruct-maas`

### Channel #7 (us-east5)  
- **Type**: GoogleOpenAI (40)
- **Base URL**: `https://us-east5-aiplatform.googleapis.com`
- **Project ID**: `hypernym-api` 
- **Region**: `us-east5`
- **Models**: `meta/llama-4-maverick-17b-128e-instruct-maas,meta/llama-4-scout-17b-16e-instruct-maas`

## Issues Resolved

### 1. ConvertRequest Not Being Called
**Problem**: GoogleOpenAI requests bypassed `ConvertRequest` function because they used `apitype.OpenAI`.
**Solution**: Modified condition in `relay/controller/text.go` to specifically include GoogleOpenAI in the `ConvertRequest` path.

### 2. Period Removal Confusion  
**Problem**: Initially tried to remove periods from model names like Azure OpenAI does.
**Solution**: Discovered Google OpenAI requires periods to remain in model names. Removed period transformation.

### 3. Wrong Region Configuration
**Problem**: Channel initially configured for `us-east5` but most Llama models are in `us-central1`.  
**Solution**: Updated channel to `us-central1` and created separate channel for `us-east5` models.

### 4. Test Button 404 Errors
**Problem**: Test button failed with 404 errors due to non-existent models.
**Solution**: Fixed region configuration and ensured proper model availability per region.

## Final Status
- ✅ Google OpenAI integration working
- ✅ Test buttons working for both channels  
- ✅ API calls working through one-api service
- ✅ Proper model name handling (periods preserved)
- ✅ Multi-region support implemented
- ✅ All Llama models accessible through appropriate channels

## Code Changes Summary
1. **relay/controller/text.go**: Fixed ConvertRequest routing for GoogleOpenAI
2. **relay/adaptor/openai/adaptor.go**: Added GoogleOpenAI URL construction, removed debug logging
3. **Database**: Created two GoogleOpenAI channels for different regions
4. **Configuration**: Set correct base URLs, regions, and model lists per channel

## Notes
- Google Cloud's regional model distribution is inconsistent but manageable with multiple channels
- Period handling differs between Azure OpenAI (remove periods) and Google OpenAI (keep periods)
- Both test functionality and API access work correctly after implementation