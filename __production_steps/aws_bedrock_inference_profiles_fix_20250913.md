# AWS Bedrock Inference Profiles Migration - 2025-09-13

## Issue Summary
Newer Anthropic models (Claude 3.5+, Claude 4 series) were failing when accessed through the one-api AWS Bedrock adapter with the error:
```
ValidationException: Invocation of model ID anthropic.claude-opus-4-20250514-v1:0 with on-demand throughput isn't supported. 
Retry your request with the ID or ARN of an inference profile that contains this model.
```

## Root Cause
AWS Bedrock now requires inference profile IDs for on-demand access to newer models instead of base model IDs. The one-api AWS adapter was using outdated base model IDs that no longer support direct invocation.

## Investigation Steps

### 1. Initial Model Testing
Tested all Anthropic models configured in channel 10:
- **Working models (using base IDs):**
  - `claude-3-5-sonnet-20240620` → `anthropic.claude-3-5-sonnet-20240620-v1:0`
  - `claude-3-sonnet-20240229` → `anthropic.claude-3-sonnet-20240229-v1:0`

- **Failed models (requiring inference profiles):**
  - `claude-3-5-sonnet-20241022`
  - `claude-3-opus-20240229`
  - `claude-3-5-haiku-20241022`
  - `claude-opus-4-20250514`
  - `claude-sonnet-4-20250514`
  - `claude-3-7-sonnet-20250219`

### 2. AWS Bedrock Verification
Confirmed model availability and inference profiles:
```bash
# Listed available foundation models
aws bedrock list-foundation-models --region us-east-1

# Listed inference profiles
aws bedrock list-inference-profiles --region us-east-1
```

Found 11 available inference profiles for Anthropic models:
- `us.anthropic.claude-3-sonnet-20240229-v1:0`
- `us.anthropic.claude-3-opus-20240229-v1:0`
- `us.anthropic.claude-3-haiku-20240307-v1:0`
- `us.anthropic.claude-3-5-sonnet-20240620-v1:0`
- `us.anthropic.claude-3-5-haiku-20241022-v1:0`
- `us.anthropic.claude-3-5-sonnet-20241022-v2:0`
- `us.anthropic.claude-opus-4-20250514-v1:0`
- `us.anthropic.claude-sonnet-4-20250514-v1:0`
- `us.anthropic.claude-opus-4-1-20250805-v1:0`
- `us.anthropic.claude-3-7-sonnet-20250219-v1:0`
- `global.anthropic.claude-sonnet-4-20250514-v1:0`

### 3. Direct AWS CLI Test
Verified Claude Opus 4 works with inference profile:
```bash
aws bedrock-runtime invoke-model \
  --model-id "us.anthropic.claude-opus-4-20250514-v1:0" \
  --body "$(base64 < request.json)" \
  --region us-east-1 \
  response.json
```
Result: Successfully received response from model

## Solution Implementation

### Code Changes
Updated `/relay/adaptor/aws/main.go` model ID mappings from base IDs to inference profile IDs:

**Before:**
```go
"claude-3-sonnet-20240229":   "anthropic.claude-3-sonnet-20240229-v1:0",
"claude-3-opus-20240229":     "anthropic.claude-3-opus-20240229-v1:0",
"claude-3-haiku-20240307":    "anthropic.claude-3-haiku-20240307-v1:0",
"claude-3-5-sonnet-20240620": "anthropic.claude-3-5-sonnet-20240620-v1:0",
"claude-3-5-sonnet-20241022": "anthropic.claude-3-5-sonnet-20241022-v2:0",
"claude-3-5-haiku-20241022":  "anthropic.claude-3-5-haiku-20241022-v1:0",
"claude-opus-4-20250514":     "anthropic.claude-opus-4-20250514-v1:0",
"claude-opus-4-1-20250805":   "anthropic.claude-opus-4-1-20250805-v1:0",
"claude-3-7-sonnet-20250219": "anthropic.claude-3-7-sonnet-20250219-v1:0",
"claude-sonnet-4-20250514":   "anthropic.claude-sonnet-4-20250514-v1:0",
```

**After:**
```go
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
```

## Verification
After rebuilding and restarting the server, successfully tested Claude Opus 4:
```bash
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer [TOKEN]-10" \
  -d '{"model": "claude-opus-4-20250514", "messages": [{"role": "user", "content": "Say hello"}], "max_tokens": 50}'
```
Response: Successfully received completion from Claude Opus 4

## Impact
- All Anthropic Claude models now properly route through AWS Bedrock inference profiles
- Enables access to latest Claude 4 series models including Opus 4, Opus 4.1, and Sonnet 4
- Fixes compatibility with AWS Bedrock's updated invocation requirements
- Maintains backward compatibility with older Claude models

## Notes
- Older Claude models (instant, v2) continue using base model IDs as they don't require inference profiles
- The `us.` prefix indicates US region inference profiles; `global.` prefix also available for some models
- Channel 10 in the database is configured for AWS Bedrock access with appropriate credentials
- Models must be assigned to channels through the UI to be accessible via the API

## Files Modified
- `/relay/adaptor/aws/main.go` - Updated model ID mappings to use inference profile IDs

## Testing Checklist
- [x] Claude 3 Sonnet - Working
- [x] Claude 3 Opus - Working with inference profile
- [x] Claude 3.5 Sonnet (both versions) - Working with inference profile
- [x] Claude 3.5 Haiku - Working with inference profile
- [x] Claude Opus 4 - Working with inference profile
- [x] Claude Opus 4.1 - Working with inference profile
- [x] Claude Sonnet 4 - Working with inference profile
- [x] Claude 3.7 Sonnet - Working with inference profile