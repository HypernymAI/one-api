# One-API Model Addition Architecture - Complete Guide

## Overview
This document comprehensively explains how models are added to one-api, why simply adding them to the code isn't sufficient, and the complete data flow from backend constants to UI availability.

## The Model System Architecture

### 1. Model Definition Layer (Backend Code)
Models are defined in channel-specific constant files:
- **Location**: `/relay/adaptor/[provider]/constants.go`
- **Example**: `/relay/adaptor/openai/constants.go`

```go
var ModelList = []string{
    "gpt-3.5-turbo",
    "gpt-4", 
    "gpt-4.1-nano", // Custom models
    "gpt-5-chat",   // Newly added
    // ... more models
}
```

### 2. Model Registration Layer
During initialization, the system aggregates all models from all providers:
- **Location**: `/controller/model.go` (init function)
- **Process**: Iterates through all adaptors, collects their model lists
- **Result**: Creates a master list of all available models in the system

### 3. Channel Configuration Layer (Database)
Channels store which models they support:
- **Table**: `channels`
- **Key Field**: `models` (comma-separated string)
- **Example**: `"gpt-4,gpt-4-turbo,gpt-3.5-turbo,o3,o3-mini"`

**CRITICAL**: A model MUST be in a channel's `models` field to be available!

### 4. Abilities Mapping Layer (Database)
The `abilities` table maps groups → models → channels:
- **Table**: `abilities`
- **Structure**:
  ```sql
  group     | model      | channel_id | enabled | priority
  ----------|------------|------------|---------|----------
  default   | gpt-4      | 1          | 1       | 0
  default   | gpt-4.1    | 1          | 1       | 0
  ```

**This table is automatically populated when channels are created/updated!**

### 5. Model Availability API Layer
When the UI requests available models:
- **Endpoint**: `/api/user/available_models`
- **Process**:
  1. Get user's group (e.g., "default")
  2. Query abilities table: `SELECT DISTINCT model WHERE group = ? AND enabled = 1`
  3. Return model list

### 6. UI Token Edit Layer
The token edit page:
- **Location**: `/web/default/src/pages/Token/EditToken.js`
- **Process**:
  1. Calls `/api/user/available_models`
  2. Populates dropdown with returned models
  3. User selects which models the token can access

## The Model Addition Flow

### What Happens When You Add a Model to Code

1. **Add to Constants** ✓
   ```go
   // In /relay/adaptor/openai/constants.go
   "gpt-5-chat", "gpt-5-mini", "gpt-5-nano",
   ```

2. **Add Pricing Ratios** ✓
   ```go
   // In /relay/billing/ratio/model.go
   "gpt-5-chat": 2.8125,
   "gpt-5-mini": 0.5625,
   "gpt-5-nano": 0.1125,
   ```

3. **Rebuild Backend** ✓
   ```bash
   ./build.sh
   ```

4. **Models Available in System** ✓
   - The models ARE now in the compiled binary
   - They show up in `/api/models` under their channel type

5. **BUT: Models NOT Available to Tokens** ❌
   - They're NOT in any channel's `models` field
   - Therefore NOT in the `abilities` table
   - Therefore NOT returned by `/api/user/available_models`
   - Therefore NOT shown in token edit UI

## Why Models Don't Appear in UI

### The Missing Link: Channel Models List

Even though models are in the code, they must be:
1. Added to a channel's `models` field in the database
2. Which triggers `UpdateAbilities()` to populate the abilities table
3. Which makes them available to `/api/user/available_models`
4. Which makes them appear in the token edit UI

### The Proper Flow

```
Backend Constants → Channel Models List → Abilities Table → Available Models API → UI Dropdown
      (Code)            (Database)          (Database)         (API Endpoint)      (Frontend)
```

## How to Properly Add Models

### Method 1: Through the UI (Recommended)
1. Navigate to Channels page
2. Edit the appropriate channel (e.g., OpenAI channel)
3. In the Models section, add the new models
4. Save the channel
5. This automatically updates the abilities table
6. Models now appear in token edit UI

### Method 2: Understanding the System (Advanced)
1. Models must be added to the channel's models list
2. When a channel is saved, it calls `UpdateAbilities()`
3. This populates/updates the abilities table
4. The abilities table is the source of truth for available models

## Common Misconceptions

### "I added models to the code, why don't they show up?"
- Models in code are just definitions
- They must be assigned to channels
- Channels must update their abilities
- Only then are they available to users

### "The models show in /api/models but not in token edit"
- `/api/models` shows ALL models in the system (from code)
- `/api/user/available_models` shows models from abilities table
- Token edit uses the latter, not the former

### "I rebuilt everything but it still doesn't work"
- Rebuilding compiles the new model definitions
- But doesn't update the database channel configurations
- The channel must be updated to include the new models

## The Abilities System Explained

### Purpose
The abilities table serves as a denormalized cache that:
1. Maps which models are available to which user groups
2. Tracks which channels provide which models
3. Enables fast lookups for model availability
4. Supports priority-based channel selection

### Population
The abilities table is populated/updated when:
1. A new channel is created
2. An existing channel is updated
3. A channel's status is changed

### Structure
```sql
-- For each model in a channel's models list
-- For each group the channel serves
-- Create an ability record
INSERT INTO abilities (group, model, channel_id, enabled, priority)
VALUES ('default', 'gpt-5-chat', 1, 1, 0);
```

## Troubleshooting Model Addition

### Symptom: Models in code but not in UI
**Check**:
1. Is the model in the channel's models list?
   ```sql
   SELECT models FROM channels WHERE type = 1; -- OpenAI
   ```
2. Is the model in the abilities table?
   ```sql
   SELECT * FROM abilities WHERE model LIKE '%gpt-5%';
   ```

### Symptom: Updated channel but models still don't appear
**Check**:
1. Did the channel update trigger abilities update?
2. Is the user's group correct?
3. Are there any errors in the logs?

### Symptom: Models appear for some users but not others
**Check**:
1. User groups may differ
2. Check abilities for specific group:
   ```sql
   SELECT DISTINCT model FROM abilities WHERE group = 'default' AND enabled = 1;
   ```

## The Complete Data Flow

```mermaid
graph TD
    A[Model Constants in Code] -->|Compiled into| B[Binary Executable]
    B -->|Read by| C[Model Controller Init]
    C -->|Available to| D[Channel Configuration]
    D -->|User Updates Channel| E[Channel Models List]
    E -->|Triggers UpdateAbilities| F[Abilities Table]
    F -->|Queried by| G[/api/user/available_models]
    G -->|Fetched by| H[Token Edit UI]
    H -->|Displays to| I[User Dropdown]
```

## Why Manual Database Edits Don't Work Properly

When you manually edit the database:
1. You might update the channel's models list
2. But you don't trigger `UpdateAbilities()`
3. So the abilities table remains out of sync
4. And the models don't appear in the UI

The UI's "Fill" button and save process:
1. Updates the channel through proper API calls
2. Which triggers all necessary updates
3. Including abilities table synchronization
4. Ensuring consistency across the system

## Summary

To add new models to one-api:

1. **Add to code** (constants.go, pricing ratios)
2. **Rebuild the application**
3. **Update channels through the UI** to include new models
4. **Channel update triggers abilities sync**
5. **Models now appear in token edit UI**

The key insight: Models in code are just potential models. They must be assigned to channels, which creates abilities, which makes them available to users. This multi-layer system provides flexibility and access control but requires understanding the complete flow.

---
*Last updated: August 26, 2025*
*Critical for understanding model addition issues*