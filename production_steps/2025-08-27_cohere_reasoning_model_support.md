# Adding Cohere Command A Reasoning Model Support

## Date: 2025-08-27

### Summary
Added support for Cohere's command-a-reasoning-08-2025 model by implementing v2 API compatibility in the Cohere adaptor.

### Changes Made

1. **Added v2 API detection** in `/relay/adaptor/cohere/adaptor.go`:
   - Checks if model name contains "reasoning"
   - Routes to v2 endpoint (`/v2/chat`) for reasoning models
   - Maintains v1 endpoint for other models

2. **Created v2 request/response structures** in `/relay/adaptor/cohere/model.go`:
   - `RequestV2`: Uses `messages` array format
   - `ResponseV2`: Handles content array with thinking/text types
   - `MessageV2Request`: Simple role/content structure for requests

3. **Added v2 conversion functions** in `/relay/adaptor/cohere/main.go`:
   - `ConvertRequestV2`: Converts OpenAI format to Cohere v2
   - `HandlerV2`: Processes v2 responses, extracts text content

4. **Added models to constants**:
   - command-a-03-2025
   - command-a-reasoning-08-2025

### Technical Details

The v2 API differs from v1:
- **Request**: Uses `messages` array instead of single `message` field
- **Response**: Contains `content` array with `type` field ("thinking" or "text")
- **Reasoning**: Shows model's thought process in "thinking" content (currently ignored)

### Testing
```bash
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "command-a-reasoning-08-2025",
    "messages": [{"role": "user", "content": "What is 2+2?"}],
    "max_tokens": 500
  }'
```

### Future Enhancements
- Add support for `thinking` parameter to control reasoning behavior
- Stream reasoning tokens separately
- Support token budgets for reasoning