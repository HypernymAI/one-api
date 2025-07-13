# Azure Llama Models Integration - July 13, 2025

## Summary
Successfully integrated Meta's Llama model family (Llama 3.1, 3.3, and 4) into one-api using Azure AI Model Catalog and Azure OpenAI channel type.

## Key Discoveries

### 1. Azure Deployment Name Requirements
- **Critical**: Model deployment names CANNOT contain periods/dots
- Azure's API expects deployment names to match exactly what you send in requests
- Example: `Llama-3.3-70B` must be deployed as `Llama-33-70B` (no dots)

### 2. UI vs Database Model Addition
- **Always use the UI** to add models to channels
- Direct SQLite updates don't properly register models in the system
- The "Fill" button in the custom model field properly registers models

## Step-by-Step Implementation

### Phase 1: Azure Portal Setup

1. **Access Azure AI Studio**
   - Go to https://ai.azure.com
   - Navigate to Model Catalog
   - Search for "Llama" models

2. **Deploy Each Model**
   - Click on desired model (e.g., `Llama-3.3-70B-Instruct`)
   - Click "Deploy"
   - **CRITICAL**: In "Deployment name" field, remove ALL periods
     - Model shows as: `Llama-3.3-70B-Instruct`
     - Deploy it as: `Llama-33-70B-Instruct`
   - Select compute resources
   - Wait for deployment to complete

3. **Get Endpoint Details**
   After deployment, you'll see:
   - Endpoint: `https://[your-resource].services.ai.azure.com`
   - API Key: Long string starting with alphanumeric characters
   - API Version: `2024-05-01-preview` (or similar)

### Phase 2: One-API Configuration

1. **Create Channel in One-API**
   - Go to Channels page
   - Click "Create New Channel"
   - Select Type: **Azure OpenAI** (not Custom!)
   
2. **Configure Azure Details**
   - **AZURE_OPENAI_ENDPOINT**: Enter base URL WITHOUT `/models/chat/completions`
     - ✅ Correct: `https://hypernymazurellama4-resource.services.ai.azure.com`
     - ❌ Wrong: `https://hypernymazurellama4-resource.services.ai.azure.com/models/chat/completions?api-version=2024-05-01-preview`
   - **Default API Version**: `2024-05-01-preview`
   - **Name**: `Azure Llama Models` (or your preference)
   - **Group**: Select appropriate group

3. **Add Models Using Custom Model Field**
   Since Llama models aren't in Azure OpenAI's dropdown:
   - Scroll to bottom of Models section
   - Find "Enter custom model name" field
   - Type deployment name EXACTLY: `Llama-33-70B-Instruct-2`
   - Click **Fill** button
   - Repeat for each model

4. **Key**: Enter your Azure API key

5. **Save**: Click Submit

### Phase 3: Testing

Test each model with curl:
```bash
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ONE_API_KEY" \
  -d '{
    "model": "Llama-33-70B-Instruct-2",
    "messages": [{"role": "user", "content": "Hello"}],
    "max_tokens": 50
  }'
```

## Models Successfully Deployed

| Azure Deployment Name | Original Model Name | Size | Notes |
|----------------------|---------------------|------|-------|
| Llama-4-Scout-17B-16E-Instruct | Llama-4-Scout-17B-16E-Instruct | 17B | Multimodal |
| Llama-4-Maverick-17B-128E-Instruct-FP8 | Llama-4-Maverick-17B-128E-Instruct-FP8 | 17B | Multimodal FP8 |
| Llama-33-70B-Instruct-2 | Llama-3.3-70B-Instruct | 70B | Latest 3.3 |
| Meta-Llama-31-405B-Instruct | Meta-Llama-3.1-405B-Instruct | 405B | Largest |
| Meta-Llama-31-8B-Instruct-2 | Meta-Llama-3.1-8B-Instruct | 8B | Smallest |

## Troubleshooting

### Issue: "Resource not found" error
- Check endpoint URL doesn't include `/models/chat/completions`
- Verify API version is in the separate field
- Ensure deployment name matches exactly (no dots!)

### Issue: Model not appearing in token permissions
- Don't use SQLite to add models directly
- Always use the UI "Fill" button
- This properly registers models in the system

### Issue: "Invalid model" error
- Deployment name must match what you type in one-api
- Remove ALL periods from deployment names
- Check for typos in model names

## Technical Details

### Why the Fill Button Works
When tracing the code, we found:
1. The Fill button (`addCustomModel` function) properly adds models to React state
2. On save, models are joined with commas: `models.join(',')`
3. This triggers proper API validation and cache updates
4. Direct SQLite updates bypass this flow and don't register properly

### Azure API Expectations
- Expects deployment names in requests
- Automatically maps deployment names to actual models
- Requires exact string matching (case-sensitive)

## Future Improvements
1. Add Llama models to default Azure OpenAI model list
2. Automate deployment via Azure CLI/ARM templates
3. Create batch model addition functionality

## References
- Azure AI Model Catalog: https://ai.azure.com/explore/models
- Original Llama implementation guide: `__production_steps/llama_models_complete_guide_20250712.md`

---
*Implementation completed: July 13, 2025*
*Tested all models successfully*
*Author: Assistant*