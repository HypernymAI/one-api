# GPT-5 Models Addition - August 26, 2025

## Summary
Successfully added GPT-5 series models (gpt-5-chat, gpt-5-mini, gpt-5-nano) to one-api. Discovered critical architectural understanding about how models propagate from code to UI availability.

## Models Added

### GPT-5 Series
- **gpt-5-chat**: Standard chat model ($1.25/1M input, $10/1M output)
- **gpt-5-mini**: Reasoning model ($0.25/1M input, $2/1M output)  
- **gpt-5-nano**: Reasoning model ($0.05/1M input, $0.40/1M output)

## Implementation Steps

### 1. Added Models to OpenAI Constants
**File**: `/relay/adaptor/openai/constants.go`
```go
// GPT-5 series models
"gpt-5-chat", "gpt-5-mini", "gpt-5-nano",
```

### 2. Added Pricing Ratios
**File**: `/relay/billing/ratio/model.go`
```go
"gpt-5-chat": 2.8125,  // ~$5.625/1M tokens averaged
"gpt-5-mini": 0.5625,  // ~$1.125/1M tokens averaged  
"gpt-5-nano": 0.1125,  // ~$0.225/1M tokens averaged
```
Note: Ratio calculation is `price_per_1K_tokens / 0.002`

### 3. Added UI Color Mappings
**File**: `/web/air/src/helpers/render.js`
```javascript
'gpt-5-nano': 'rgb(255,20,147)', // Deep pink
'gpt-5-mini': 'rgb(255,69,0)',   // Red-orange  
'gpt-5-chat': 'rgb(220,20,60)',  // Crimson
```

### 4. Rebuilt Application
```bash
# Rebuild web UI
cd web && ./build.sh

# Rebuild Go binary (creates one-api-en, NOT one-api)
./build.sh
```

## Critical Discovery: Model Availability Architecture

### The Problem
After adding models to code and rebuilding, GPT-5 models were NOT appearing in the token edit UI dropdown, despite being in the system.

### Root Cause Analysis
Discovered the multi-layer architecture for model availability:

1. **Backend Constants** → Models defined in code
2. **Channel Models List** → Each channel must include models in its configuration
3. **Abilities Table** → Maps groups→models→channels (populated when channels are saved)
4. **Available Models API** → Returns models from abilities table
5. **Token Edit UI** → Shows models from available models API

### The Missing Step
Models were added to code but NOT to any channel's models list. Without being in a channel, they never populate the abilities table, so they never appear in the UI.

### Solution
Models must be added to channels through the UI:
1. Edit the channel (e.g., OpenAI or Azure OpenAI)
2. Add new models to the channel's models list
3. Save the channel (this triggers UpdateAbilities())
4. Models now appear in token edit UI

### Documentation Created
Created comprehensive guide: `/documentation/model_addition_architecture.md`

## Testing Results

### Azure OpenAI Channel Configuration
**Channel #8**: azure_openai_gpt4_gpt5_chat
- Added GPT-5 models to channel configuration
- Models: gpt-5-chat, gpt-5-mini, gpt-5-nano

### API Parameter Discovery
**Important**: Reasoning models require different API parameters:

**Standard models** (use `max_tokens`):
- gpt-5-chat ✓
- gpt-3.5-turbo ✓
- gpt-4 series ✓

**Reasoning models** (require `max_completion_tokens`):
- gpt-5-mini ✓
- gpt-5-nano ✓
- o3 ✓
- o3-mini ✓
- o4-mini ✓

### Token Allocation for Reasoning Models
Reasoning models use tokens for internal reasoning before generating output:
- Low token limits (e.g., 10) result in empty responses
- Need sufficient tokens (e.g., 1000) for both reasoning and output
- Example: gpt-5-nano used 64 reasoning tokens before generating response

### Test Results
```bash
# gpt-5-chat (standard model)
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer sk-BoroIrY7uCruh3IR48Cb5b151f8943E88bF86e3d91Ee2bB2" \
  -d '{"model": "gpt-5-chat", "messages": [{"role": "user", "content": "Say hello"}], "max_tokens": 10}'
# Result: "Hello! 👋 How are you doing today?"

# gpt-5-mini (reasoning model)  
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer sk-BoroIrY7uCruh3IR48Cb5b151f8943E88bF86e3d91Ee2bB2" \
  -d '{"model": "gpt-5-mini", "messages": [{"role": "user", "content": "Say hello"}], "max_completion_tokens": 1000}'
# Result: "Hello! How can I help you today?"
```

## Issues Encountered

### 1. Wrong Binary Name
- Build script creates `one-api-en`, not `one-api`
- Old `one-api` binary from August was being used
- Solution: Delete old binary, use correct name

### 2. Models Not in UI
- Models added to code but not appearing in token edit dropdown
- Root cause: Models must be in a channel's models list to populate abilities table
- Solution: Add models to channel through UI (or understand the architecture)

### 3. Channel Test Button
- Test button hardcodes "gpt-3.5-turbo" as default model
- Fails for channels without this model
- Code attempts to be smart but falls back to hardcoded default

### 4. API Parameter Differences
- Reasoning models reject `max_tokens` parameter
- Require `max_completion_tokens` instead
- Handled by the API providers, not one-api code

## Lessons Learned

1. **Read ALL documentation** before making changes
2. **Trace existing patterns** (e.g., how gpt-4.1 was added) completely
3. **Understand the full data flow** from code to UI
4. **The abilities table is the source of truth** for model availability
5. **Direct database edits bypass critical update logic**
6. **Build outputs may have unexpected names** (one-api-en vs one-api)

## Status
✅ GPT-5 models successfully added and working
✅ All three models tested and functional
✅ Pricing ratios configured correctly
✅ Comprehensive architecture documentation created
⚠️ Models must still be added to channels via UI for full availability

---
*Implementation completed: August 26, 2025*
*Time spent: ~3 hours (mostly debugging why models weren't appearing)*
*Key insight: Understanding the abilities table architecture is critical*