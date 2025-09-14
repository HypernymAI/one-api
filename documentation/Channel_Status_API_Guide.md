# Channel Status API Guide

## Overview

The Channel Status API endpoint (`/api/channel/status/:id`) allows administrators to check the availability of models configured in a channel without actually invoking them. This is particularly useful for:

- Verifying model availability across cloud providers
- Avoiding unnecessary API costs from test invocations
- Checking models that require special handling (thinking/reasoning models)
- Monitoring channel health without triggering rate limits

## Endpoint Details

### URL
```
GET /api/channel/status/:id
```

### Authentication
- **Required**: Admin access token
- **Header**: `Authorization: Bearer {ADMIN_ACCESS_TOKEN}`
- **Note**: Only users with admin role (role = 100) can access this endpoint

### Parameters
- `:id` - The channel ID to check (from the database)

## How It Works

The endpoint performs non-invocation checks to verify model availability:

1. **Fetches channel configuration** from the database (including stored API keys)
2. **Queries provider APIs** for model metadata (not actual invocations)
3. **Returns availability status** for each model configured in the channel

### Provider-Specific Implementations

#### AWS Bedrock
- Uses AWS SDK to check model status via `GetFoundationModel` API
- For inference profiles (us.*, global.*), assumes availability
- Parses credentials from format: `ACCESS_KEY|SECRET_KEY|REGION`

#### Google Gemini
- Queries each model's metadata endpoint: `/v1beta/models/{model}`
- Verifies the model supports `generateContent` method
- Returns clear status for each model

#### Azure OpenAI
- Lists all available models via `/openai/models` endpoint
- Handles various naming conventions:
  - Date suffixes (e.g., gpt-5-mini-2025-08-07)
  - Dots vs no dots (gpt-4.1 vs gpt-41)
  - Azure-specific naming (gpt-35 vs gpt-3.5)

## Response Format

### Success Response
```json
{
  "success": true,
  "channel": "channel_name",
  "type": "provider_type",
  "models": [
    {
      "model": "model_name",
      "available": true
    },
    {
      "model": "model_name_2",
      "available": false,
      "error": "Model not found"
    }
  ],
  "time": 0.548
}
```

### Error Response
```json
{
  "success": false,
  "message": "Error description"
}
```

## Usage Examples

### 1. Check AWS Bedrock Channel
```bash
# Check channel 10 (AWS Bedrock with Claude and Llama models)
curl -X GET "http://localhost:3000/api/channel/status/10" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" | jq
```

**Example Response:**
```json
{
  "success": true,
  "channel": "aws_claude_llama",
  "type": "AWS Bedrock",
  "models": [
    {
      "model": "claude-3-5-sonnet-20240620",
      "available": true
    },
    {
      "model": "claude-opus-4-20250514",
      "available": true
    },
    {
      "model": "llama-3.3-70b",
      "available": true
    }
  ],
  "time": 1.508
}
```

### 2. Check Google Gemini Channel
```bash
# Check channel 3 (Google Gemini)
curl -X GET "http://localhost:3000/api/channel/status/3" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" | jq
```

**Example Response:**
```json
{
  "success": true,
  "channel": "gemini",
  "type": "Google Gemini",
  "models": [
    {
      "model": "gemini-1.5-flash",
      "available": true
    },
    {
      "model": "gemini-2.5-pro",
      "available": true
    },
    {
      "model": "text-embedding-004",
      "available": false,
      "error": "Model doesn't support generateContent"
    }
  ],
  "time": 1.358
}
```

### 3. Check Azure OpenAI Channel
```bash
# Check channel 8 (Azure OpenAI)
curl -X GET "http://localhost:3000/api/channel/status/8" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" | jq
```

**Example Response:**
```json
{
  "success": true,
  "channel": "azure_openai_gpt4_gpt5_chat",
  "type": "Azure OpenAI",
  "models": [
    {
      "model": "gpt-4.1",
      "available": true
    },
    {
      "model": "gpt-5-mini",
      "available": true
    },
    {
      "model": "gpt-5-nano",
      "available": true
    }
  ],
  "time": 0.548
}
```

## Getting Your Admin Token

### Option 1: From Database
```sql
-- Find admin users and their access tokens
SELECT u.username, u.access_token 
FROM users u 
WHERE u.role = 100;
```

### Option 2: From Tokens Table
```sql
-- Find tokens belonging to admin users
SELECT t.key, t.name, u.username 
FROM tokens t 
JOIN users u ON t.user_id = u.id 
WHERE u.role = 100 
AND t.status = 1;
```

## Common Error Messages

### Authentication Errors
```json
{
  "success": false,
  "message": "无权进行此操作，access token 无效"
}
```
**Solution**: Ensure you're using a valid admin token

### Channel Not Found
```json
{
  "success": false,
  "message": "channel not found"
}
```
**Solution**: Verify the channel ID exists in the database

### Provider-Specific Errors

#### AWS Bedrock
- `"Invalid AWS credentials format"` - Channel key not in `ACCESS_KEY|SECRET_KEY|REGION` format
- `"Model not accessible"` - Model exists but account lacks access
- `"Model mapping not found"` - Model name not recognized

#### Google Gemini
- `"Model not found"` - Model doesn't exist or is deprecated
- `"Model doesn't support generateContent"` - Model is not for text generation (e.g., embedding models)

#### Azure OpenAI
- `"Failed to load config"` - Missing API version in channel configuration
- `"Model not found in available models list"` - Model/deployment not available

## Comparison with Test Endpoint

| Feature | `/api/channel/test/:id` | `/api/channel/status/:id` |
|---------|-------------------------|---------------------------|
| **Purpose** | Test actual model invocation | Check model availability |
| **API Calls** | Sends test message | Queries metadata only |
| **Cost** | Uses tokens (costs money) | No token usage (free) |
| **Speed** | Slower (generation time) | Faster (metadata only) |
| **Rate Limits** | Can trigger limits | Unlikely to trigger |
| **Thinking Models** | May fail with low tokens | Works correctly |
| **Use Case** | Verify model works | Verify model exists |

## Best Practices

### 1. Regular Monitoring
Set up scheduled checks to monitor channel availability:
```bash
# Cron job example - check every hour
0 * * * * curl -X GET "http://localhost:3000/api/channel/status/10" \
  -H "Authorization: Bearer TOKEN" >> /var/log/channel_status.log
```

### 2. Batch Checking
Check multiple channels in sequence:
```bash
#!/bin/bash
CHANNELS="3 8 10"
TOKEN="YOUR_ADMIN_TOKEN"

for id in $CHANNELS; do
  echo "Checking channel $id:"
  curl -s -X GET "http://localhost:3000/api/channel/status/$id" \
    -H "Authorization: Bearer $TOKEN" | jq '.models[] | select(.available==false)'
done
```

### 3. Error Detection
Parse responses to detect unavailable models:
```javascript
// JavaScript example
const response = await fetch(`/api/channel/status/${channelId}`, {
  headers: { 'Authorization': `Bearer ${token}` }
});
const data = await response.json();

const unavailable = data.models.filter(m => !m.available);
if (unavailable.length > 0) {
  console.warn('Unavailable models:', unavailable);
  // Send alert or take action
}
```

## Troubleshooting

### Issue: All models show as unavailable
**Possible Causes:**
- Invalid API credentials in channel configuration
- Network connectivity issues
- Provider API changes

**Solution:**
1. Verify channel credentials are correct
2. Test provider API directly
3. Check provider status page

### Issue: Specific models not found
**Possible Causes:**
- Model deprecated by provider
- Model name mismatch
- Region-specific availability (AWS)

**Solution:**
1. Check provider documentation for current model names
2. Verify model is available in your region
3. Update model configuration in database

### Issue: Slow response times
**Possible Causes:**
- Provider API latency
- Multiple model checks in sequence
- Network issues

**Solution:**
1. Implement caching for status results
2. Check provider API performance
3. Consider parallel checking implementation

## Security Considerations

1. **API Keys**: Channel API keys remain in the database, never exposed to users
2. **Admin Only**: Endpoint restricted to admin users only
3. **Read-Only**: Only checks status, doesn't modify anything
4. **Rate Limiting**: Consider implementing rate limits to prevent abuse

## Future Enhancements

Planned improvements for this endpoint:

1. **Caching**: Cache results for X minutes to reduce API calls
2. **Bulk Check**: Check all channels in single request
3. **Webhooks**: Alert when models become unavailable
4. **Metrics**: Track availability over time
5. **Auto-Update**: Automatically update channel model lists based on availability

## Related Documentation

- [AWS Bedrock Models Guide](./AWS_Bedrock_Models_Guide.md)
- [Azure Model Addition](./Azure_Model_Addition.md)
- [Model Addition Architecture](./model_addition_architecture.md)

## Support

For issues or questions:
1. Check channel configuration in database
2. Verify provider API is accessible
3. Review logs in `oneapi.log`
4. Contact system administrator

---
*Last updated: 2025-09-13*