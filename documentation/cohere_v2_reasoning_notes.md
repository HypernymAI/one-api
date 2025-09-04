# Cohere V2 API and Reasoning Models

## Key Differences from V1

### 1. Endpoint Changes
- **V1**: `https://api.cohere.ai/v1/chat`
- **V2**: `https://api.cohere.ai/v2/chat`

### 2. Response Structure
V2 responses include content array with types:
```json
{
  "message": {
    "role": "assistant",
    "content": [
      {
        "type": "thinking",
        "thinking": "Step by step reasoning..."
      },
      {
        "type": "text", 
        "text": "Final answer"
      }
    ]
  }
}
```

### 3. Reasoning Control
```python
# Enable/disable reasoning (enabled by default)
thinking = {
    "type": "disabled"  # or "enabled"
}

# Set token budget
thinking = {
    "token_budget": 500  # limits thinking to 500 tokens
}
```

### 4. Models Requiring V2
- `command-a-reasoning-08-2025`
- Future reasoning models

### 5. Implementation Notes for one-api

To support reasoning models, the Cohere adaptor has:

1. **Dynamic endpoint selection**: Checks model name and uses v2 for models containing "reasoning"
2. **Request format conversion**: Converts OpenAI format to v2 `messages` array
3. **Response parsing**: Extracts text content from content array (thinking currently ignored)
4. **Token management**: Be aware that reasoning models consume many tokens internally

Example usage through one-api:
```json
{
    "model": "command-a-reasoning-08-2025",
    "messages": [
        {"role": "system", "content": "You are an expert..."},
        {"role": "user", "content": "Your question here"}
    ],
    "max_tokens": 650  // Minimum recommended for simple tasks
}
```

**Token Budget Recommendations:**
- Simple questions (like hypernyms): 650+ tokens
- Medium complexity: 1000-2000 tokens  
- Complex reasoning: 3000+ tokens
- Maximum performance: 31000 tokens (allows extensive thinking)

### 6. Token Usage
- V2 returns both `billed_units` and `tokens`
- `billed_units`: What you're charged for (actual output tokens)
- `tokens`: Total tokens including thinking
- **Important**: Reasoning models use significant tokens for internal thinking before generating output
  - With low token limits (e.g., 100-350), all tokens may be consumed by thinking with empty output
  - Recommend minimum 650+ tokens for simple tasks
  - Complex reasoning may require thousands of tokens

### 7. Backward Compatibility
- V1 models continue to use v1 endpoint
- Only reasoning models require v2