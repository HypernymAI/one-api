# Azure Model Addition

## Purpose
This document provides a comprehensive guide for adding AI models from Azure AI Model Catalog to one-api using the Azure OpenAI channel type.

## Prerequisites
- Active Azure subscription with AI services access
- Azure AI Studio access (https://ai.azure.com)
- Running one-api instance
- Admin access to one-api

## Core Concepts

### Azure Deployment Names
Azure AI uses "deployment names" to identify model instances. These names:
- Must NOT contain periods (`.`)
- Are used in API calls to specify which model to use
- Can differ from the actual model name

**Example**: The model `Llama-3.3-70B-Instruct` must be deployed as `Llama-33-70B-Instruct`

### Channel Types in One-API
- **Azure OpenAI**: For models hosted on Azure AI (including non-OpenAI models)
- **Custom**: For completely custom endpoints
- Azure OpenAI type provides better integration for Azure-hosted models

## Step-by-Step Process

### 1. Deploy Model in Azure

1. Navigate to Azure AI Studio: https://ai.azure.com
2. Go to **Model Catalog**
3. Search for your desired model
4. Click on the model card
5. Click **Deploy**
6. Configure deployment:
   - **Deployment name**: Remove ALL periods from model name
     - `Llama-3.3-70B` → `Llama-33-70B`
     - `Meta-Llama-3.1-405B` → `Meta-Llama-31-405B`
   - **Compute type**: Select based on model size
   - **Region**: Choose closest to your users
7. Wait for deployment (can take 5-15 minutes)
8. Once deployed, note down:
   - Endpoint URL (e.g., `https://myresource.services.ai.azure.com`)
   - API Key
   - API Version (e.g., `2024-05-01-preview`)

### 2. Configure in One-API

1. Go to one-api admin panel
2. Navigate to **Channels**
3. Click **Create New Channel**
4. Configure as follows:

#### Basic Settings
- **Type**: Azure OpenAI
- **Name**: Descriptive name (e.g., "Azure Llama Models")
- **Group**: Select appropriate group(s)

#### Azure Configuration
- **AZURE_OPENAI_ENDPOINT**: 
  - Enter ONLY the base URL
  - ✅ `https://myresource.services.ai.azure.com`
  - ❌ `https://myresource.services.ai.azure.com/models/chat/completions?api-version=2024-05-01-preview`
- **Default API Version**: `2024-05-01-preview` (or your version)
- **Key**: Your Azure API key

#### Adding Models
For models not in the dropdown list:
1. Scroll to the **Models** section
2. At the bottom, find **"Enter custom model name"** field
3. Type the deployment name EXACTLY as created in Azure
4. Click **Fill** button
5. The model will be added to the selected models list
6. Repeat for each model

**Critical**: Use the Fill button, don't manually edit the database!

### 3. Save and Test

1. Click **Submit** to save the channel
2. Test with curl:

```bash
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer [YOUR_ONE_API_KEY]" \
  -d '{
    "model": "[DEPLOYMENT_NAME]",
    "messages": [{"role": "user", "content": "Hello"}],
    "max_tokens": 100
  }'
```

## Common Issues and Solutions

### "Resource not found" Error
**Cause**: URL formatting issue
**Solution**: 
- Remove `/models/chat/completions` from endpoint URL
- Ensure API version is in separate field
- Check deployment name matches exactly

### Model Not Available in Token Permissions
**Cause**: Model not properly registered
**Solution**: 
- Always use the UI "Fill" button to add models
- Don't edit database directly
- Re-add using the UI if needed

### "Invalid model" Error
**Cause**: Model name mismatch
**Solution**:
- Verify deployment name in Azure
- Ensure no periods in deployment name
- Check for exact match (case-sensitive)

## Best Practices

1. **Naming Convention**: Create a consistent pattern for deployment names
   - Remove periods: `3.1` → `31`
   - Keep other formatting: `Meta-Llama-31-405B-Instruct`

2. **Documentation**: Keep a mapping table:
   ```
   | Original Model | Deployment Name |
   |----------------|-----------------|
   | Llama-3.3-70B  | Llama-33-70B    |
   ```

3. **Testing**: Always test each model after adding

4. **Bulk Addition**: When adding multiple models:
   - Deploy all in Azure first
   - Add all to one-api in one session
   - Test all before announcing availability

## Example: Adding Llama Models

Here's a complete example of adding Llama models:

1. **Azure Deployments Created**:
   - `Llama-33-70B-Instruct-2` (from Llama-3.3-70B-Instruct)
   - `Meta-Llama-31-405B-Instruct` (from Meta-Llama-3.1-405B-Instruct)
   - `Meta-Llama-31-8B-Instruct-2` (from Meta-Llama-3.1-8B-Instruct)

2. **One-API Configuration**:
   - Type: Azure OpenAI
   - Endpoint: `https://hypernymazurellama4-resource.services.ai.azure.com`
   - API Version: `2024-05-01-preview`
   - Models: Added via Fill button

3. **Testing**: All models confirmed working

## Technical Notes

### Why Fill Button is Required
The Fill button in the UI:
- Adds models to React component state
- Triggers proper validation
- Updates internal caches
- Ensures models are available in all dropdowns

Direct database edits bypass these steps and cause issues.

### API Format
Azure OpenAI API expects:
- Base URL + `/deployments/{deployment-name}/chat/completions?api-version={version}`
- One-api handles this URL construction automatically
- Just provide base URL and deployment name

## Related Documentation
- Azure AI Documentation: https://learn.microsoft.com/en-us/azure/ai-studio/
- One-API Channel Types: See main documentation
- Model-specific guides: Check `__production_steps/` directory

---
*Last updated: July 13, 2025*
*Applies to: one-api with Azure OpenAI channel type*