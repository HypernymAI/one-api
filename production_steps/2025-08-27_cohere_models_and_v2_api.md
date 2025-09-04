# Cohere Model Updates and V2 API Implementation

## Date: 2025-08-27

### Overview
Successfully added Cohere's latest models including command-a-reasoning-08-2025, which required implementing v2 API support. The implementation maintains backward compatibility with existing Cohere models while enabling new reasoning capabilities.

### Models Added
1. **command-a-03-2025** - Latest command model (v1 API)
2. **command-a-reasoning-08-2025** - Reasoning model with thought process output (v2 API)

### Implementation Details

#### 1. V2 API Detection
The adaptor now automatically routes requests based on model name:
- Models containing "reasoning" → v2 endpoint (`/v2/chat`)
- All other models → v1 endpoint (`/v1/chat`)

#### 2. Request Format Differences
**V1 Format:**
```json
{
  "message": "user message",
  "chat_history": [{"role": "CHATBOT", "message": "..."}],
  "model": "command-r"
}
```

**V2 Format:**
```json
{
  "messages": [
    {"role": "user", "content": "user message"},
    {"role": "assistant", "content": "assistant response"}
  ],
  "model": "command-a-reasoning-08-2025"
}
```

#### 3. Response Format Differences
**V1 Response:**
```json
{
  "response_id": "...",
  "text": "The answer is...",
  "meta": {
    "tokens": {
      "input_tokens": 10,
      "output_tokens": 20
    }
  }
}
```

**V2 Response:**
```json
{
  "id": "...",
  "message": {
    "role": "assistant",
    "content": [
      {
        "type": "thinking",
        "thinking": "Let me think about this..."
      },
      {
        "type": "text",
        "text": "The answer is..."
      }
    ]
  },
  "usage": {
    "billed_units": {
      "input_tokens": 10,
      "output_tokens": 20
    }
  }
}
```

### Files Modified
1. `/relay/adaptor/cohere/adaptor.go`
   - Added v2 endpoint routing logic
   - Model name detection for API version selection

2. `/relay/adaptor/cohere/model.go`
   - Added RequestV2, ResponseV2, MessageV2Request structures
   - Added ContentPart for thinking/text distinction

3. `/relay/adaptor/cohere/main.go`
   - Added ConvertRequestV2 function
   - Added HandlerV2 for v2 response processing
   - Updated Handler to detect and route v2 responses

4. `/relay/adaptor/cohere/constant.go`
   - Added command-a-03-2025
   - Added command-a-reasoning-08-2025

5. `/relay/billing/ratio/model.go`
   - Added pricing for command-a-03-2025 (4.0/1000*USD)
   - Added pricing for command-a-reasoning-08-2025 (5.0/1000*USD)

### Testing Results

#### V1 Models (Still Working)
```bash
# command-r
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer TOKEN" \
  -d '{"model": "command-r", "messages": [{"role": "user", "content": "What is 2+2?"}]}'
# Response: "The sum of 2+2 = 4."

# command-r-plus  
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer TOKEN" \
  -d '{"model": "command-r-plus", "messages": [{"role": "user", "content": "What is 2+2?"}]}'
# Response: "The answer is 4."

# command-a-03-2025
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer TOKEN" \
  -d '{"model": "command-a-03-2025", "messages": [{"role": "user", "content": "What is 2+2?"}]}'
# Response: "The answer to 2+2 is 4."
```

#### V2 Models (New)
```bash
# command-a-reasoning-08-2025
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer TOKEN" \
  -d '{"model": "command-a-reasoning-08-2025", "messages": [{"role": "user", "content": "What is 2+2?"}]}'
# Response: "The answer to 2 + 2 is 4."
```

### Current Limitations
1. **Thinking content ignored**: The v2 handler currently only extracts the "text" type content, ignoring the "thinking" process
2. **No thinking parameter support**: Cannot control reasoning behavior or set token budgets
3. **Streaming not tested**: V2 streaming response handling may need updates

### Future Improvements
1. Add thinking parameter support for controlling reasoning behavior
2. Implement separate handling for thinking tokens in streaming
3. Add option to include thinking content in responses
4. Support reasoning token budget limits
5. Add proper v2 streaming support

### Backward Compatibility
✅ All existing Cohere models continue to work with v1 API
✅ No changes required for existing integrations
✅ Automatic routing based on model name

### Notes
- Internet variants (e.g., command-r-internet) are automatically created by Cohere's init function
- The reasoning model shows significant thinking process before answering
- Billing is based on billed_units, not total tokens (which includes thinking)