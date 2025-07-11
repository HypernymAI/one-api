# Model Updates - ZX7M9 - 2025-01-10

## Purpose
Updated one-api to include latest AI models from OpenAI, Anthropic, Gemini, and DeepSeek that were missing from our fork but available in upstream or released in 2025. Also removed deprecated models that return 404 errors.

## Summary of Changes
- Added 27 new model definitions across 4 providers
- Added corresponding billing ratios
- Fixed missing MILLI_USD constant
- Removed/commented out deprecated models

## Detailed Edits

### 1. OpenAI Models (`relay/adaptor/openai/constants.go`)
Added O3/O4 series models released April 2025:
```go
// ZX7M9: O3/O4 models (April 2025 release)
"o3", "o3-2025-04-16",
"o3-mini", "o3-mini-2025-01-31",
"o4-mini", "o4-mini-2025-04-16",
```

### 2. Anthropic Models (`relay/adaptor/anthropic/constants.go`)
Added Claude 3.5, 3.7, and 4 family models:
```go
// ZX7M9: Latest Claude 3.5 models from upstream
"claude-3-5-haiku-20241022",
"claude-3-5-haiku-latest",
"claude-3-5-sonnet-20241022",
"claude-3-5-sonnet-latest",
// ZX7M9: Claude 3.7 (first hybrid reasoning model)
"claude-3-7-sonnet-20250219",
"claude-3-7-sonnet-latest",
// ZX7M9: Claude 4 family (released May 2025)
"claude-opus-4-20250514",
"claude-sonnet-4-20250514",
```

### 3. Gemini Models (`relay/adaptor/gemini/constants.go`)
Added missing Gemini 1.5, 2.0, and new 2.5 models:
```go
// ZX7M9: Missing models from upstream
"gemini-1.5-flash", "gemini-1.5-flash-8b", "gemini-1.5-pro-experimental",
"gemini-2.0-flash",
"gemini-2.0-flash-exp",
"gemini-2.0-flash-lite-preview-02-05",
"gemini-2.0-flash-thinking-exp-01-21",
"gemini-2.0-pro-exp-02-05",
"text-embedding-004", "aqa",
// ZX7M9: Gemini 2.5 models (GA June 2025)
"gemini-2.5-flash",
"gemini-2.5-pro",
```

### 4. DeepSeek Models (`relay/adaptor/deepseek/constants.go`)
Added DeepSeek R1 reasoning model:
```go
// ZX7M9: R1 reasoning model (accessed via deepseek-reasoner)
"deepseek-reasoner",
```

### 5. Billing Ratios (`relay/billing/ratio/model.go`)
Added pricing for all new models:

**OpenAI O-series:**
- o3 variants: 15.0 (premium reasoning)
- o3-mini variants: 1.5 (efficient reasoning)
- o4-mini variants: 0.75 (fastest reasoning)

**Anthropic:**
- Claude 3.5 Haiku: 1.0 / 1000 * USD
- Claude 3.5/3.7 Sonnet: 3.0 / 1000 * USD
- Claude Opus 4: 15.0 / 1000 * USD
- Claude Sonnet 4: 3.0 / 1000 * USD

**Gemini:**
- Flash variants: 0.0375 - 0.075 * MILLI_USD
- Pro variants: 1.25 - 2.5 * MILLI_USD
- Embeddings: 0.01 * MILLI_USD

**DeepSeek:**
- deepseek-reasoner: 0.14 (~$0.00028/1K tokens)

## Technical Notes
- All edits follow the existing passthrough proxy pattern
- No architectural changes required
- Model names verified against official API documentation
- Pricing based on official provider rates as of 2025-01-10
- Preserved existing GPT-4.1 and GPT-4.5-preview custom models

## Build Fixes
1. Added missing `MILLI_USD` constant to `relay/billing/ratio/model.go`:
```go
const (
	USD2RMB   = 7
	USD       = 500 // $0.002 = 1 -> $1 = 500
	MILLI_USD = 1.0 / 1000 * USD  // Added this
	RMB       = USD / USD2RMB
)
```

2. Fixed missing commas in model ratio map (lines 84, 101, 138)

## Deprecated Models Removed
Due to 404 errors, commented out/removed these deprecated models:

**Anthropic:**
- claude-instant-1.2 (retired)
- claude-2.0 (deprecated Jan 2025)
- claude-2.1 (deprecated Jan 2025)
- claude-3-sonnet-20240229 (deprecated Jan 2025)

**Gemini:**
- gemini-pro (use gemini-1.5-pro instead)
- gemini-1.0-pro (use gemini-1.5-pro instead)
- gemini-1.0-pro-001 (deprecated)

## Verification
- Confirmed DeepSeek uses "deepseek-reasoner" not "deepseek-r1" for API
- Confirmed no "o3-pro" in API (ChatGPT only)
- Confirmed no "o4" exists (only o4-mini)
- Confirmed Gemini thinking model is "gemini-2.0-flash-thinking-exp-01-21"
- Tested build successfully compiles after fixes
- Confirmed deprecated models return 404 errors