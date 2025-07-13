# Complete Llama Models Implementation Guide - July 12, 2025

## Executive Summary
This document provides an exhaustive guide for implementing Llama models in one-api using either Google Cloud Platform (GCP) or Microsoft Azure, leveraging existing startup credits.

## Table of Contents
1. [Current Situation](#current-situation)
2. [Why GCP or Azure](#why-gcp-or-azure)
3. [Available Llama Models](#available-llama-models)
4. [Implementation Options](#implementation-options)
5. [Testing & Verification](#testing-verification)
6. [Timeline & Dependencies](#timeline-dependencies)
7. [Related Documentation](#related-documentation)

---

## Current Situation

### What We Have
1. **one-api fork** at `/Users/fieldempress/Desktop/source/plurigrid/one-api`
2. **Startup Credits**:
   - Microsoft Startup Program credits (amount unknown)
   - GCP credits ($300 new customer credits confirmed)
3. **Working Models**:
   - OpenAI models (GPT-4, o3, etc.)
   - Google Gemini models (2.5-pro, 2.5-flash)
   - Various Chinese models (Baidu, Alibaba, etc.)

### What We Need
Access to Meta's Llama models (Llama 2, 3, 3.1, 4) through cloud providers using existing credits.

### Current Llama Support in one-api
```go
// From various provider constants:
// Ollama (local): llama2:7b, llama2:latest, llama3:latest
// Groq (cloud): llama2-7b-2048, llama2-70b-4096, llama3-8b-8192, llama3-70b-8192
// Together AI: meta-llama/Llama-3-70b-chat-hf
// Cloudflare: @cf/meta/llama-3.1-8b-instruct, etc.
```

---

## Why GCP or Azure

### Microsoft Azure Advantages ✅
1. **You have startup credits** through Microsoft Startup Program
2. **Confirmed Llama support** in Azure AI Foundry
3. **Models available**:
   - Llama 2: 7B, 13B, 70B
   - Llama 3.1: 8B, 70B, 405B
   - Llama 4: Scout and Maverick (multimodal)
4. **Integration path**: Azure AI Model Catalog with OpenAI-compatible API
5. **Startup program benefits**:
   - Free access to models
   - 1:1 consulting support
   - Early access to new models

### Google Cloud Platform (GCP) Advantages ✅
1. **$300 free credits** for new customers
2. **Vertex AI Model Garden** hosts Llama models:
   - Llama 3.1: 8B, 70B, 405B (GA)
   - Llama 3.2: Multimodal models
   - Llama 4: Scout and Maverick
3. **Pricing**: Preview models are FREE
4. **API**: Serverless deployment with pay-as-you-go

### Comparison
| Feature | Azure | GCP |
|---------|-------|-----|
| Credits | Startup Program ✅ | $300 new customer |
| Llama 2 | ✅ Yes | ❌ No |
| Llama 3.1 | ✅ Yes | ✅ Yes |
| Llama 4 | ✅ Yes | ✅ Yes |
| Free Preview | ❌ No | ✅ Yes |
| one-api support | ❓ Custom channel | ❓ Custom channel |

**Recommendation**: Start with Azure due to startup credits and broader model selection.

---

## Available Llama Models

### Azure AI Model Catalog
```
# Llama 2 Family
- Llama-2-7b
- Llama-2-7b-chat
- Llama-2-13b
- Llama-2-13b-chat
- Llama-2-70b
- Llama-2-70b-chat

# Llama 3/3.1 Family
- Llama-3-8B-Instruct
- Llama-3-70B-Instruct
- Llama-3.1-8B-Instruct
- Llama-3.1-70B-Instruct
- Llama-3.1-405B-Instruct

# Llama 4 (Multimodal)
- meta-llama/llama-4-scout-17b-16e-instruct
- meta-llama/llama-4-maverick-17b-128e-instruct
```

### GCP Vertex AI
```
# Llama 3.1 Family
- llama3.1-8b
- llama3.1-70b
- llama3.1-405b

# Llama 3.2 (Multimodal)
- llama3.2-1b
- llama3.2-3b
- llama3.2-11b-vision
- llama3.2-90b-vision

# Llama 4
- llama-4-scout
- llama-4-maverick
```

---

## Implementation Options

### Option 1: Azure AI Model Catalog (RECOMMENDED)

#### Step 1: Deploy Model in Azure
```bash
# Prerequisites
az login
az account set --subscription "YOUR_SUBSCRIPTION_ID"

# Create resource group
az group create --name llama-models-rg --location eastus2

# Create AI workspace
az ml workspace create --name llama-workspace --resource-group llama-models-rg

# Deploy Llama model (example with Llama 3.1)
# This would be done through Azure Portal UI typically
```

#### Step 2: Get Endpoint Details
After deployment, you'll receive:
- **Endpoint URL**: `https://your-endpoint.eastus2.inference.ml.azure.com`
- **API Key**: `Bearer YOUR_AZURE_AI_KEY`

#### Step 3: Configure one-api Custom Channel
```json
{
  "name": "Azure Llama",
  "type": 8,  // Custom channel
  "key": "YOUR_AZURE_AI_KEY",
  "base_url": "https://your-endpoint.eastus2.inference.ml.azure.com",
  "models": "Llama-3.1-70B-Instruct",
  "other": {
    "api_version": "2024-05-01-preview"
  }
}
```

#### Step 4: Test via CLI
```bash
# Test Azure endpoint directly
curl -X POST "https://your-endpoint.eastus2.inference.ml.azure.com/chat/completions?api-version=2024-05-01-preview" \
  -H "Authorization: Bearer YOUR_AZURE_AI_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [{"role": "user", "content": "Hello"}],
    "model": "Llama-3.1-70B-Instruct"
  }'

# Test through one-api
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer sk-BoroIrY7uCruh3IR48Cb5b151f8943E88bF86e3d91Ee2bB2" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Llama-3.1-70B-Instruct",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

### Option 2: GCP Vertex AI

#### Step 1: Enable APIs
```bash
gcloud auth login
gcloud config set project YOUR_PROJECT_ID

# Enable required APIs
gcloud services enable aiplatform.googleapis.com
gcloud services enable compute.googleapis.com
```

#### Step 2: Deploy Model
```python
from google.cloud import aiplatform

aiplatform.init(project='YOUR_PROJECT_ID', location='us-central1')

# Deploy Llama model
endpoint = aiplatform.Model.deploy(
    model_name="llama3.1-70b",
    machine_type="n1-standard-8",
    accelerator_type="NVIDIA_TESLA_V100",
    accelerator_count=1
)

print(f"Endpoint: {endpoint.resource_name}")
```

#### Step 3: Configure one-api
Similar to Azure, use Custom Channel with Vertex AI endpoint.

### Option 3: Quick Test with Existing Providers

#### Groq (if you get credits)
```bash
# Already in one-api, just need API key
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer YOUR_ONE_API_KEY" \
  -d '{
    "model": "llama3-70b-8192",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

#### Ollama (local testing)
```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull Llama model
ollama pull llama3

# Run Ollama server
ollama serve

# Configure one-api channel with base_url: http://localhost:11434
```

---

## Testing & Verification

### Test Script
```bash
#!/bin/bash
# test_llama_models.sh

MODELS=(
  "llama2-7b"
  "llama3-8b"
  "llama3.1-70b"
  "llama-4-scout"
)

for model in "${MODELS[@]}"; do
  echo "Testing $model..."
  curl -s -X POST http://localhost:3000/v1/chat/completions \
    -H "Authorization: Bearer sk-BoroIrY7uCruh3IR48Cb5b151f8943E88bF86e3d91Ee2bB2" \
    -H "Content-Type: application/json" \
    -d "{
      \"model\": \"$model\",
      \"messages\": [{\"role\": \"user\", \"content\": \"Say hi\"}],
      \"max_tokens\": 10
    }" | jq -r '.choices[0].message.content // .error.message'
  echo "---"
done
```

### Verification Checklist
- [ ] Endpoint responds to health check
- [ ] Authentication works (Bearer token)
- [ ] Model list includes Llama models
- [ ] Chat completions endpoint works
- [ ] Token counting is accurate
- [ ] Streaming works (if supported)
- [ ] Error handling for rate limits

---

## Timeline & Dependencies

### Phase 1: Research & Setup (Day 1)
1. **Hour 1-2**: Verify Azure/GCP credits
2. **Hour 3-4**: Create cloud resources
3. **Hour 5-6**: Deploy first Llama model

### Phase 2: Integration (Day 2)
1. **Hour 1-2**: Configure one-api custom channel
2. **Hour 3-4**: Test basic functionality
3. **Hour 5-6**: Debug and optimize

### Phase 3: Production (Day 3)
1. **Hour 1-2**: Add all Llama models
2. **Hour 3-4**: Update documentation
3. **Hour 5-6**: Performance testing

### Dependencies
1. **Cloud Credits**: Must be active
2. **API Access**: Proper tier/permissions
3. **one-api**: Running and accessible
4. **Network**: No firewall blocking

---

## Troubleshooting

### Common Issues

#### Issue 1: "Model not found"
```bash
# Check available models
curl https://your-endpoint/models \
  -H "Authorization: Bearer YOUR_KEY"
```

#### Issue 2: "Authentication failed"
```bash
# Verify API key format
# Azure uses: "Bearer KEY"
# Some use: "apikey KEY"
```

#### Issue 3: "Endpoint not compatible"
- Azure AI uses OpenAI-compatible format ✅
- GCP Vertex AI might need adaptation ⚠️
- Check request/response format matches

### Debug Commands
```bash
# Test raw endpoint
curl -v https://endpoint/chat/completions ...

# Check one-api logs
tail -f logs/oneapi-20250712.log | grep -i llama

# Database check
sqlite3 one-api.db "SELECT * FROM channels WHERE name LIKE '%llama%';"
```

---

## Cost Considerations

### Azure Costs (with startup credits)
- **Covered by credits**: Model deployment, inference
- **Not covered**: Storage, networking beyond limits
- **Optimization**: Use smaller models for testing

### GCP Costs
- **Free tier**: $300 credits
- **Preview models**: FREE during preview
- **GA models**: Pay per token

### Cost Optimization
1. Start with smallest model (7B/8B)
2. Use preview/free tier when available
3. Implement caching in one-api
4. Monitor usage closely

---

## Security Considerations

1. **API Keys**: Store securely, rotate regularly
2. **Endpoints**: Use HTTPS only
3. **Access Control**: Limit by IP if possible
4. **Monitoring**: Track usage for anomalies

---

## Next Steps

### Immediate Actions
1. Check Azure subscription status
2. Verify startup program benefits
3. Create test resource group

### Short Term (This Week)
1. Deploy Llama 3.1-70B on Azure
2. Configure one-api custom channel
3. Run comprehensive tests

### Long Term
1. Add all Llama models
2. Optimize for performance
3. Create automated deployment scripts
4. Consider multi-region deployment

---

## References

### Our Documentation
- `/one-api/__production_steps/language_english_default_20250110.md`
- `/one-api/__production_steps/o3_pro_responses_api_20250712.md`
- `/one-api/build_instructions.md`

### External Resources
- [Azure AI Model Catalog](https://ai.azure.com/explore/models)
- [GCP Vertex AI Models](https://cloud.google.com/vertex-ai/docs/generative-ai/models)
- [Meta Llama Official](https://llama.meta.com/)
- [one-api Custom Channels](https://github.com/songquanpeng/one-api/wiki)

### API Documentation
- [Azure AI Inference API](https://learn.microsoft.com/en-us/azure/ai-studio/reference/)
- [Vertex AI API](https://cloud.google.com/vertex-ai/docs/reference)

---

## Appendix: Model Comparison

| Model | Parameters | Context | Use Case | Azure | GCP |
|-------|------------|---------|----------|-------|-----|
| Llama 2-7B | 7B | 4K | General chat | ✅ | ❌ |
| Llama 3-8B | 8B | 8K | Better reasoning | ✅ | ✅ |
| Llama 3.1-70B | 70B | 128K | Complex tasks | ✅ | ✅ |
| Llama 3.1-405B | 405B | 128K | Research | ✅ | ✅ |
| Llama 4 Scout | 17B | Multi-modal | Vision+Text | ✅ | ✅ |

---

## Final Recommendations

1. **Start with Azure** - You have credits
2. **Deploy Llama 3.1-70B** - Best balance
3. **Use Custom Channel** - Most flexible
4. **Test thoroughly** - Before production
5. **Monitor costs** - Even with credits

## Success Criteria
- [ ] At least one Llama model accessible via one-api
- [ ] Cost within credit limits
- [ ] Performance acceptable (<2s latency)
- [ ] All tests passing
- [ ] Documentation updated

---
*Document created: July 12, 2025*
*Author: ZX7M9*
*Status: Implementation Guide*