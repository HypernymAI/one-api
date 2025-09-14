# AWS Bedrock Models Complete Guide

## Overview
This document provides a comprehensive guide for AWS Bedrock models available through one-api, including exact model names, configuration requirements, and operational status as of September 2025.

## ⚠️ CRITICAL: Understanding Inference Profiles vs On-Demand Access

AWS Bedrock now has two distinct access patterns:

1. **On-Demand Throughput**: Traditional direct model access
   - Available for: Llama 3.x base models
   - Direct model IDs: `meta.llama3-8b-instruct-v1:0`

2. **Inference Profiles**: Required for newer models
   - Required for: Llama 4, Claude 4, Claude 3.7, and some Claude 3.5 models
   - Use profile IDs: `us.meta.llama4-scout-17b-instruct-v1:0` (note the `us.` prefix)
   - Error if not used: "Invocation with on-demand throughput isn't supported"

## Prerequisites
- AWS Account with Bedrock access enabled
- IAM user with Bedrock permissions
- AWS Access Key and Secret Key
- Supported AWS Region (us-east-1 recommended)

## Channel Configuration

### Setting Up AWS Bedrock Channel in one-api

1. **Channel Type**: AWS Claude (Type 37 internally)
2. **Required Configuration**:
   ```json
   {
     "region": "us-east-1",
     "ak": "YOUR_AWS_ACCESS_KEY_ID",
     "sk": "YOUR_AWS_SECRET_ACCESS_KEY"
   }
   ```

### Important Configuration Notes
- **DO NOT** include field labels in credentials (e.g., don't append "Secret" or "Region" to values)
- The channel type must be set to "AWS Claude" in the UI
- Models must be added to the channel using the UI's "Fill" button

## Working Models (Complete List)

### Claude 4 Series
| Model ID | Display Name | Status | Notes |
|----------|--------------|--------|-------|
| `claude-opus-4-20250514` | Claude Opus 4 | ✅ Working | Latest Opus model, excellent performance |
| `claude-sonnet-4-20250514` | Claude Sonnet 4 | ✅ Working | Balanced performance and cost |
| `claude-opus-4-1-20250805` | Claude Opus 4.1 | ❌ No Permission | Requires special access |

### Claude 3.7 Series
| Model ID | Display Name | Status | Notes |
|----------|--------------|--------|-------|
| `claude-3-7-sonnet-20250219` | Claude 3.7 Sonnet | ✅ Working | February 2025 release |
| `claude-3-7-sonnet-latest` | Claude 3.7 Sonnet Latest | ✅ Working | Always points to latest 3.7 |

### Claude 3.5 Series
| Model ID | Display Name | Status | Notes |
|----------|--------------|--------|-------|
| `claude-3-5-sonnet-20240620` | Claude 3.5 Sonnet (June) | ✅ Working | Stable June 2024 version |
| `claude-3-5-sonnet-20241022` | Claude 3.5 Sonnet (Oct) | ❌ Invocation Error | Region-specific issue |
| `claude-3-5-sonnet-latest` | Claude 3.5 Sonnet Latest | ✅ Working | Points to latest stable |
| `claude-3-5-haiku-20241022` | Claude 3.5 Haiku | ❌ Invocation Error | Region-specific issue |
| `claude-3-5-haiku-latest` | Claude 3.5 Haiku Latest | ✅ Working | Fast, economical option |

### Claude 3 Series
| Model ID | Display Name | Status | Notes |
|----------|--------------|--------|-------|
| `claude-3-opus-20240229` | Claude 3 Opus | ✅ Working | Most capable Claude 3 |
| `claude-3-sonnet-20240229` | Claude 3 Sonnet | ❌ Invocation Error | Use 3.5 or 3.7 instead |
| `claude-3-haiku-20240307` | Claude 3 Haiku | ✅ Working | Fast, lightweight |

### Claude Legacy Models
| Model ID | Display Name | Status | Notes |
|----------|--------------|--------|-------|
| `claude-2.0` | Claude 2.0 | ❌ No Permission | Deprecated |
| `claude-2.1` | Claude 2.1 | ❌ No Permission | Deprecated |
| `claude-instant-1.2` | Claude Instant 1.2 | ❌ No Permission | Deprecated |

### Llama Models
| Model ID | Display Name | Status | Context Window | Notes |
|----------|--------------|--------|---------------|-------|
| `llama-3-8b` | Llama 3 8B | ✅ Working | 8K | On-demand available |
| `llama-3-70b` | Llama 3 70B | ✅ Working | 8K | On-demand available |
| `llama-3.1-8b` | Llama 3.1 8B | ✅ Working | 128K | On-demand available |
| `llama-3.1-70b` | Llama 3.1 70B | ✅ Working | 128K | On-demand available |
| `llama-3.2-1b` | Llama 3.2 1B | ✅ Working | 128K | On-demand available |
| `llama-3.2-3b` | Llama 3.2 3B | ✅ Working | 128K | On-demand available |
| `llama-3.2-11b` | Llama 3.2 11B | ✅ Working | 128K | Multimodal capable |
| `llama-3.2-90b` | Llama 3.2 90B | ✅ Working | 128K | Multimodal capable |
| `llama-3.3-70b` | Llama 3.3 70B | ✅ Working | 128K | Latest 70B variant |
| `llama-4-scout-17b` | Llama 4 Scout 17B | ✅ Working | 3.5M | **Uses inference profile (us. prefix)** |
| `llama-4-maverick-17b` | Llama 4 Maverick 17B | ✅ Working | 1M | **Uses inference profile (us. prefix)** |

## API Usage Examples

### Basic Request
```bash
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ONE_API_KEY" \
  -d '{
    "model": "claude-opus-4-20250514",
    "messages": [{"role": "user", "content": "Hello!"}],
    "max_tokens": 100
  }'
```

### Llama Model Request
```bash
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ONE_API_KEY" \
  -d '{
    "model": "llama-3-70b",
    "messages": [{"role": "user", "content": "Explain quantum computing"}],
    "max_tokens": 500
  }'
```

## Model-Specific Details

### Claude Models
- **Request Format**: Uses Anthropic's message format
- **Max Tokens**: Varies by model, typically 4096-8192
- **Streaming**: Fully supported
- **System Messages**: Supported

### Llama Models
- **Request Format**: Converted to prompt-based format internally
- **Max Output Tokens**: 8192 (max_gen_len)
- **Prompt Template**: Automatically applied with special tokens
- **Streaming**: Fully supported

## Troubleshooting

### Common Issues

1. **"The security token included in the request is invalid"**
   - Check AWS credentials are correct
   - Ensure no extra characters appended to credentials
   - Verify IAM user has Bedrock permissions

2. **"Invocation of model with on-demand throughput isn't supported"**
   - Model requires inference profile setup
   - Use Llama 3 models instead of 3.1+ for on-demand access

3. **"Model not found"**
   - Ensure model is added to channel configuration
   - Use exact model ID as listed above

4. **Region-specific errors**
   - Some models may not be available in all regions
   - Try us-east-1 for maximum availability

## Billing Considerations

### Pricing Tiers (Relative)
- **Most Expensive**: Claude Opus models
- **Balanced**: Claude Sonnet models, Llama 70B
- **Economical**: Claude Haiku models, Llama 8B
- **Budget**: Llama 3 8B

### Token Ratios in one-api
All models have been configured with appropriate billing ratios based on AWS Bedrock pricing.

## Best Practices

1. **Model Selection**:
   - Use Haiku models for simple, high-volume tasks
   - Use Sonnet models for balanced performance
   - Use Opus models for complex reasoning tasks
   - Use Llama models for open-source requirements

2. **Performance Optimization**:
   - Enable streaming for better user experience
   - Set appropriate max_tokens to control costs
   - Use newer model versions for better performance

3. **Error Handling**:
   - Implement retry logic for transient errors
   - Have fallback models configured
   - Monitor model availability by region

## Future Considerations

### Upcoming Support
- Inference profiles for Llama 3.1+ models
- Additional AWS regions
- Cross-region inference profiles

### Model Deprecation Timeline
- Claude 2.x models are already deprecated
- Claude 3 base models may be deprecated in favor of 3.5+
- Always use "latest" variants for long-term stability

## Appendix: AWS Bedrock Model IDs

### Complete Mapping Table
| one-api Model Name | AWS Bedrock Model ID | Provider | Type |
|--------------------|---------------------|----------|------|
| claude-opus-4-20250514 | anthropic.claude-opus-4-20250514-v1:0 | Anthropic | Inference Profile |
| claude-opus-4-1-20250805 | anthropic.claude-opus-4-1-20250805-v1:0 | Anthropic | Inference Profile |
| claude-sonnet-4-20250514 | anthropic.claude-sonnet-4-20250514-v1:0 | Anthropic | Inference Profile |
| claude-3-7-sonnet-20250219 | anthropic.claude-3-7-sonnet-20250219-v1:0 | Anthropic | Inference Profile |
| claude-3-5-sonnet-20240620 | anthropic.claude-3-5-sonnet-20240620-v1:0 | Anthropic | On-Demand |
| claude-3-5-sonnet-20241022 | anthropic.claude-3-5-sonnet-20241022-v2:0 | Anthropic | Inference Profile |
| claude-3-5-haiku-20241022 | anthropic.claude-3-5-haiku-20241022-v1:0 | Anthropic | Inference Profile |
| llama-3-8b | meta.llama3-8b-instruct-v1:0 | Meta | On-Demand |
| llama-3-70b | meta.llama3-70b-instruct-v1:0 | Meta | On-Demand |
| llama-3.1-8b | meta.llama3-1-8b-instruct-v1:0 | Meta | On-Demand |
| llama-3.1-70b | meta.llama3-1-70b-instruct-v1:0 | Meta | On-Demand |
| llama-3.2-1b | meta.llama3-2-1b-instruct-v1:0 | Meta | On-Demand |
| llama-3.2-3b | meta.llama3-2-3b-instruct-v1:0 | Meta | On-Demand |
| llama-3.2-11b | meta.llama3-2-11b-instruct-v1:0 | Meta | On-Demand |
| llama-3.2-90b | meta.llama3-2-90b-instruct-v1:0 | Meta | On-Demand |
| llama-3.3-70b | meta.llama3-3-70b-instruct-v1:0 | Meta | On-Demand |
| **llama-4-scout-17b** | **us.meta.llama4-scout-17b-instruct-v1:0** | Meta | **Inference Profile** |
| **llama-4-maverick-17b** | **us.meta.llama4-maverick-17b-instruct-v1:0** | Meta | **Inference Profile** |

## Important Notes on Llama 4 Models

### Llama 4 Scout (17B)
- **Parameters**: 17B active, 109B total with 16 experts
- **Context Window**: 3.5 million tokens
- **Inference Profile Required**: Must use `us.` prefix in model ID
- **Streaming**: Not supported

### Llama 4 Maverick (17B)  
- **Parameters**: 17B active, 400B total with 128 experts
- **Context Window**: 1 million tokens
- **Inference Profile Required**: Must use `us.` prefix in model ID
- **Streaming**: Not supported

---
*Last Updated: September 6, 2025*
*Tested with one-api version: Latest*