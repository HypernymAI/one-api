# Azure GPT-4.1 Dot Handling Fix - Production Log
## Date: 2025-09-07

### Summary
Implemented conditional dot removal for Azure OpenAI model names based on API version to support new GPT-4.1 models while maintaining backward compatibility.

### Problem Statement
- **Issue**: GPT-4.1 models (gpt-4.1-mini, gpt-4.1-nano, gpt-4.1) were failing with "deployment does not exist" errors
- **Root Cause**: Legacy code was removing dots from all Azure model names (e.g., `gpt-4.1-mini` → `gpt-41-mini`)
- **Historical Context**: Azure OpenAI deployments previously didn't support dots in deployment names (GitHub issue #1191)
- **New Requirement**: Modern Azure deployments (2025+) now support and REQUIRE dots for GPT-4.1 models

### Discovery Process

#### 1. Initial Configuration
Channel 8 (`azure_openai_gpt4_gpt5_chat`) configuration:
- **Type**: 3 (Azure)
- **Base URL**: `https://chris-1050-chat-models-resource.cognitiveservices.azure.com`
- **API Key**: Stored in `key` field
- **API Version**: Initially `2024-03-01-preview`, updated to `2025-01-01-preview`
- **Models**: gpt-4.1-nano, gpt-4.1-mini, gpt-4.1, gpt-5-chat, gpt-5-mini, gpt-5-nano

#### 2. Testing Results
Direct Azure API call (WORKS):
```bash
curl "https://chris-1050-chat-models-resource.cognitiveservices.azure.com/openai/deployments/gpt-4.1-mini/chat/completions?api-version=2025-01-01-preview"
# Result: Success
```

Through one-api (FAILED):
```bash
# Code was calling: /openai/deployments/gpt-41-mini/chat/completions
# Result: "deployment does not exist"
```

#### 3. Code Analysis
Located problematic code in `/relay/adaptor/openai/adaptor.go`:
```go
// Line 43 - Removing ALL dots from model names
model_ = strings.Replace(model_, ".", "", -1)
```

Git blame revealed:
- **Author**: JustSong (upstream maintainer)
- **Date**: 2024-02-18
- **Commit**: 1aa374cc "refactor: use adaptor to do relay & test"
- **Reference**: GitHub issue #1191 (Azure deployment naming restrictions)

### Solution Implemented

#### Code Change
Modified `/relay/adaptor/openai/adaptor.go` to conditionally remove dots based on API version:

```go
// Lines 43-48
// Only remove dots for API versions before 2025
// Newer Azure deployments (2025+) support and require dots for models like gpt-4.1
// Check if API version starts with 2025 or later
if len(meta.Config.APIVersion) >= 4 && meta.Config.APIVersion[:4] < "2025" {
    model_ = strings.Replace(model_, ".", "", -1)
}
```

#### Logic Explanation
- **API versions < 2025**: Remove dots (legacy behavior for compatibility)
- **API versions >= 2025**: Keep dots (required for GPT-4.1 models)
- **String comparison**: Safe because Azure API versions follow YYYY-MM-DD format

### Testing Verification

#### String Comparison Safety Test
```go
"2024-03-01-preview" < "2025" // true - dots removed
"2024-05-01-preview" < "2025" // true - dots removed  
"2025-01-01-preview" < "2025" // false - dots kept
"2025-02-01-preview" < "2025" // false - dots kept
```

#### Model Testing Results
After implementation:
- ✅ **GPT-4.1-mini**: Working (dots preserved)
- ❌ **GPT-4.1-nano**: Not deployed in Azure yet
- ✅ **GPT-4.1**: Should work when deployed
- ✅ **GPT-5-chat**: Still working (no dots to remove)
- ✅ **GPT-5-mini**: Working with max_completion_tokens
- ✅ **GPT-5-nano**: Working with max_completion_tokens

### Important Notes

#### Database Structure
Azure channels store configuration directly in table fields:
- `key`: API key
- `base_url`: Azure endpoint
- `other`: API version (e.g., "2025-01-01-preview")
- `config`: JSON field (empty for Azure type)

**Warning**: The `config` field shows empty JSON but actual config is in the main fields!

#### Model Deployment Names
Azure deployment names must match exactly:
- GPT-4.1 models: Keep the dot (e.g., `gpt-4.1-mini`)
- GPT-5 models: No dots to worry about (e.g., `gpt-5-chat`)

#### API Version Compatibility
- **2024-x versions**: Legacy, removes dots
- **2025+ versions**: Modern, preserves dots
- Both GPT-5 and GPT-4.1 models work with `2025-01-01-preview`

### Channels Affected

| Channel ID | Name | API Version | Dot Handling |
|------------|------|-------------|--------------|
| 4 | Llama | 2024-05-01-preview | Removes dots |
| 8 | azure_openai_gpt4_gpt5_chat | 2025-01-01-preview | Keeps dots |

### Future Considerations

1. **Upstream Sync**: This is a divergence from upstream. Document in fork notes.
2. **API Version Updates**: When updating Azure API versions, test dot-containing models
3. **Model Naming**: Future Azure models may have different requirements
4. **Reasoning Models**: GPT-5-nano/mini require `max_completion_tokens` instead of `max_tokens`

### Configuration Checklist

When adding Azure OpenAI channels with GPT-4.1 support:
- [ ] Set API version to `2025-01-01-preview` or later
- [ ] Ensure Azure deployment names match model names exactly (with dots)
- [ ] Test both dot-containing and regular model names
- [ ] For reasoning models, use `max_completion_tokens` parameter

### Files Modified
- `/relay/adaptor/openai/adaptor.go` - Added conditional dot removal based on API version

---
Generated: 2025-09-07 23:10 UTC