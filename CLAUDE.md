# Claude Development Guidelines for one-api

## Important Commands and Processes

### Starting the one-api Server

**ALWAYS use screen with logging to avoid blocking processes:**

```bash
# Kill any existing processes first
pkill -f one-api

# Start in a screen session with logging
screen -dmS oneapi bash -c './one-api-en 2>&1 | tee -a oneapi.log'

# To view the logs
tail -f oneapi.log

# To attach to the screen session (if needed)
screen -r oneapi
```

**NEVER run blocking processes directly like:**
- `./one-api-en` (blocks the terminal)
- `./one-api-en &` (background without proper logging)
- `nohup ./one-api-en` (harder to manage)

### Building one-api

```bash
# The build script creates one-api-en, NOT one-api
./build.sh
```

### Testing and Linting

When completing tasks, ALWAYS run:
```bash
# Add specific lint/typecheck commands here when known
# npm run lint
# npm run typecheck
```

## Model Addition Process

When adding new models:
1. Add to channel constants (e.g., `/relay/adaptor/[provider]/constants.go`)
2. Add pricing ratios to `/relay/billing/ratio/model.go`
3. Rebuild with `./build.sh`
4. Update channel models list through UI (models must be assigned to channels to appear)

## Key Architecture Notes

- Models in code are just definitions - they must be assigned to channels
- The abilities table controls what models appear in the UI
- Never manually edit the SQLite database
- Use the UI's "Fill" button to properly update model lists

## File Locations

- Binary location: `./one-api-en` (NOT `./one-api`)
- Log files: `oneapi.log`, `oneapi-debug.log`
- Database: SQLite (do not edit manually)

## Common Issues and Solutions

1. **Models not appearing in UI**: Need to update channel models list through UI
2. **Wrong binary**: Use `one-api-en`, not `one-api`
3. **Blocking processes**: Always use screen with logging
4. **Reasoning models**: May require `max_completion_tokens` instead of `max_tokens`
5. **Cohere reasoning models**: Automatically use v2 API (models with "reasoning" in name)

---
Last updated: 2025-08-27