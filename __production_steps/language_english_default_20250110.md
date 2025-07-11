# Language Update - Default to English - 2025-01-10

## Purpose
Change one-api to default to English instead of Chinese. Currently the interface shows Chinese by default, requiring manual translation. The upstream repository has already implemented proper internationalization (i18n) that defaults to English.

## Problem
- Current version defaults to Chinese UI
- No language detection from browser
- Manual translation required every time
- Poor user experience for non-Chinese speakers

## Solution
Copy the i18n implementation from upstream which:
- Defaults to English when no language preference is set
- Respects browser's Accept-Language header
- Only shows Chinese to users who specifically request it (zh-* locales)
- Provides proper translation files for both languages

## Files to Copy from Upstream

### 1. Language Middleware
**Source:** `upstream-check/middleware/language.go`
**Destination:** `middleware/language.go`

This file:
- Reads browser's Accept-Language header
- Defaults to "en" if no header present (line 15)
- Only switches to "zh-CN" if browser requests Chinese (line 17-18)
- Sets language in context for other components to use

### 2. I18n Module (entire directory)
**Source:** `upstream-check/common/i18n/`
**Destination:** `common/i18n/`

Contains:
- `i18n.go` - Core translation logic
- `locales/en.json` - English translations
- `locales/zh-CN.json` - Chinese translations

### 3. Update Router
**File:** `router/main.go`
**Add:** `router.Use(middleware.Language())` after other middleware

This activates the language detection for all routes.

## Implementation Steps

1. Copy language middleware:
```bash
cp upstream-check/middleware/language.go middleware/language.go
```

2. Copy i18n module:
```bash
cp -r upstream-check/common/i18n common/
```

3. Edit `router/main.go` to add language middleware:
```go
// Add after other middleware like CORS, Logger, etc.
router.Use(middleware.Language())
```

4. Rebuild and test:
```bash
./build.sh
./one-api
```

## Expected Behavior After Update
- English UI by default for all users
- Chinese UI only for browsers with Chinese language preference
- Automatic language detection based on browser settings
- No more manual translation needed

## Technical Details
The implementation uses Gin middleware to intercept requests and set language context based on:
1. Accept-Language HTTP header from browser
2. Default fallback to English
3. Context key "lang" used throughout the application

## Dependencies
- No new dependencies required
- Uses existing Gin framework
- Translation files are embedded in binary

## Testing
1. Access with English browser - should see English
2. Access with Chinese browser - should see Chinese  
3. Access with no language header - should see English (default)

## Rollback
If issues occur, simply:
1. Delete `middleware/language.go`
2. Delete `common/i18n/` directory
3. Remove middleware line from router
4. Rebuild