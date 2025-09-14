# Model Token Limits Guide

## Overview
This document provides comprehensive token limits for all models across Google Gemini, AWS Bedrock, and Azure OpenAI as of September 2025.

---

## Google Gemini Models

All limits confirmed via API: `https://generativelanguage.googleapis.com/v1beta/models/{model}`

| Model | Input Token Limit | Output Token Limit | Notes |
|-------|------------------|-------------------|-------|
| **gemini-1.5-flash** | 1,000,000 | 8,192 | Fast, efficient model |
| **gemini-1.5-flash-8b** | 1,000,000 | 8,192 | 8B parameter variant |
| **gemini-1.5-pro** | 2,000,000 | 8,192 | Largest context window |
| **gemini-2.0-flash** | 1,048,576 | 8,192 | Updated architecture |
| **gemini-2.5-flash** | 1,048,576 | 65,536 | 8x larger output than 2.0 |
| **gemini-2.5-pro** | 1,048,576 | 65,536 | Thinking model with large output |

### Key Features
- **Gemini 1.5 Pro** has the largest input context at 2M tokens
- **Gemini 2.5 models** support 65K output tokens (8x more than earlier versions)
- **Thinking tokens**: Gemini 2.5 Pro uses internal thinking tokens that count toward limits

---

## AWS Bedrock Claude Models

### Claude 3 Series
| Model | Context Window | Max Output | Notes |
|-------|---------------|------------|-------|
| **Claude 3 Haiku** | 200,000 | 4,096 | Fast, lightweight |
| **Claude 3 Sonnet** | 200,000 | 4,096 | Balanced performance |
| **Claude 3 Opus** | 200,000 | 4,096 | Most capable Claude 3 |
| **Claude 3.5 Sonnet** | 200,000 | 8,192 | Enhanced capabilities |
| **Claude 3.5 Haiku** | 200,000 | 8,192 | Improved speed |
| **Claude 3.7 Sonnet** | 200,000 | 131,072 | Largest output in Claude 3 series |

### Claude 4 Series
| Model | Context Window | Max Output | Notes |
|-------|---------------|------------|-------|
| **Claude Opus 4** | 200,000 | 65,536 | Advanced reasoning |
| **Claude Opus 4.1** | 200,000 | 65,536 | Latest Opus model |
| **Claude Sonnet 4** | 200,000 (1M in preview) | 65,536 | 1M context in public beta |

### Important AWS Bedrock Considerations
- **On-demand vs Provisioned**: Token limits may vary based on throughput type
- **Inference Profiles**: Newer models require inference profile IDs (us.*, global.*)
- **Extended Thinking**: When using thinking features, thinking budget must be < max_tokens
- **Validation**: Claude 3.7+ returns errors if prompt + max_tokens > context window

---

## Azure OpenAI Models

### GPT-4 Series
| Model | Context Window | Max Output | Notes |
|-------|---------------|------------|-------|
| **GPT-4** | 8,192 / 32,768 | 4,096 | Original GPT-4 |
| **GPT-4 Turbo** | 128,000 | 4,096 | Larger context |
| **GPT-4o** | 128,000 | 16,384 | Optimized version |
| **GPT-4o mini** | 128,000 | 16,384 | Smaller, faster |

### GPT-4.1 Series (2025)
| Model | Context Window | Max Output | Notes |
|-------|---------------|------------|-------|
| **GPT-4.1** | 1,000,000 | 16,384 | 1M token context |
| **GPT-4.1 mini** | 1,000,000 | 16,384 | Faster variant |
| **GPT-4.1 nano** | 1,000,000 | 16,384 | 80.1% MMLU score |

### GPT-5 Series (2025)
| Model | Context Window | Max Output | Notes |
|-------|---------------|------------|-------|
| **GPT-5** | 272,000 | Variable | Deep reasoning model |
| **GPT-5 mini** | 272,000 | Variable | Real-time + reasoning |
| **GPT-5 nano** | 272,000 | Variable | Ultra-low latency |
| **GPT-5 chat** | 272,000 | Variable | Conversational variant |

### O-Series Reasoning Models
| Model | Context Window | Max Output | Notes |
|-------|---------------|------------|-------|
| **o1** | 200,000 | 100,000 | Original reasoning model |
| **o1-mini** | 128,000 | 65,536 | Smaller, faster |
| **o3-mini** | 200,000 | 100,000 | Latest mini (2025-01-31) |
| **o4-mini** | 200,000 | 100,000 | Enhanced reasoning |

### Azure-Specific Considerations
- **Regional Availability**: 1M context for GPT-4.1 not available in all regions
- **API Version**: Newer features require 2025+ API versions
- **Deployment Names**: May differ from model names
- **Registration**: GPT-5 requires registration, GPT-5-mini/nano don't

---

## Token Calculations & Best Practices

### Token Equivalencies
- **1 token** ≈ 4 characters ≈ 0.75 words
- **1,000 tokens** ≈ 750 words ≈ 1.5 pages
- **1M tokens** ≈ 750,000 words ≈ 1,500 pages ≈ 30,000 lines of code

### Choosing the Right Model

#### For Large Context Needs (>200K tokens)
1. **Gemini 1.5 Pro** - 2M input tokens
2. **GPT-4.1 series** - 1M tokens (Azure)
3. **Gemini models** - 1M+ tokens

#### For Large Output Needs (>16K tokens)
1. **Claude 3.7 Sonnet** - 131K output
2. **O-series models** - 100K output
3. **Gemini 2.5 models** - 65K output
4. **Claude 4 series** - 65K output

#### For Fast, Efficient Processing
1. **Gemini 1.5 Flash** - 1M context, fast
2. **GPT-4.1 nano** - 1M context, ultra-fast
3. **Claude 3.5 Haiku** - 200K context, lightweight

### Handling Thinking/Reasoning Models

Models with internal reasoning (Gemini 2.5 Pro, GPT-5 series, O-series) require special consideration:

1. **Reserve tokens for thinking**: These models use tokens internally before generating output
2. **Increase max_tokens**: Set higher than needed output to account for thinking
3. **Monitor usage**: Thinking tokens count toward your limits and costs

### Common Pitfalls to Avoid

1. **Assuming max_tokens = output length**: It's the maximum, not guaranteed
2. **Ignoring thinking tokens**: Can cause empty responses if limit too low
3. **Context overflow**: Prompt + max_tokens can't exceed context window
4. **Regional differences**: Same model may have different limits by region/deployment

---

## Rate Limits & Quotas

### Google Gemini
- Free tier: 1,500 requests/day for Flash models
- Paid: Based on pricing tier

### AWS Bedrock
- On-demand: Lower token limits than provisioned
- Provisioned: Full model capabilities
- Varies by region and model

### Azure OpenAI
- GPT-5: 20K TPM, 200 RPM (most models)
- GPT-5-chat: 50K TPM, 50 RPM
- GPT-4.1: Varies by deployment

---

## Quick Reference Table

| Provider | Largest Input Context | Largest Output | Best Value |
|----------|---------------------|----------------|------------|
| **Gemini** | 2M (1.5 Pro) | 65K (2.5 models) | 1.5 Flash |
| **AWS Bedrock** | 1M (Sonnet 4 preview) | 131K (Claude 3.7) | Claude 3.5 Haiku |
| **Azure** | 1M (GPT-4.1) | 131K (Claude via Azure) | GPT-4o mini |

---

## API Examples

### Checking Token Limits Programmatically

#### Google Gemini
```bash
curl "https://generativelanguage.googleapis.com/v1beta/models/{MODEL}?key={KEY}" \
  | jq '{inputTokenLimit, outputTokenLimit}'
```

#### AWS Bedrock
```bash
aws bedrock get-foundation-model \
  --model-identifier {MODEL_ID} \
  --region us-east-1
```

#### Azure OpenAI
```bash
curl "https://{RESOURCE}.openai.azure.com/openai/models?api-version=2024-02-15-preview" \
  -H "api-key: {KEY}"
```

---

## Updates & Changes

### Recent Updates (September 2025)
- Gemini 2.5 models released with 65K output tokens
- GPT-4.1 family supports 1M context
- GPT-5 series now available with reasoning capabilities
- Claude Sonnet 4 1M context in public beta

### Expected Changes
- Gemini planning 2M context for more models
- Claude 4 series expanding availability
- Azure regional expansion for 1M context models

---

*Last updated: September 13, 2025*
*Data sources: Direct API queries, official documentation, cloud provider announcements*