# One-API Deployment and Replication Guide

## Overview
This guide explains how to replicate a fully configured one-api instance (including Google OpenAI/Vertex AI channels) to a new server or environment.

## What Gets Transferred via Git vs Manual Copy

### Via Git Pull (Automatic)
When you pull the repository and branch, you get:
- **All source code changes**
  - OAuth implementation for Google Cloud service accounts
  - UI instructions for channel setup
  - Model name handling (e.g., period removal for Azure)
- **Dependencies** (in `go.mod` and `package.json`)
- **Frontend assets** (needs rebuild)
- **Configuration files**

### Manual Copy Required
- **`one-api.db` (SQLite database)** - CRITICAL
  - Contains all channel configurations
  - Service account JSON credentials
  - API tokens
  - User accounts
  - System settings
  - Model mappings

## Complete Deployment Steps

### 1. Clone/Pull Repository
```bash
git clone [repository-url]
cd one-api
git checkout [branch-name]  # e.g., chris/google_openai_vertex_api_as_a_service
```

### 2. Copy Database
```bash
# From source server
scp one-api.db user@new-server:/path/to/one-api/

# Or if local
cp /path/from/source/one-api.db /path/to/new/one-api/
```

**CRITICAL**: The database contains:
- All channel configurations with authentication
- Service account JSON for Google OpenAI channels
- API tokens (e.g., `sk-BoroIrY7uCruh3IR48Cb5b151f8943E88bF86e3d91Ee2bB2`)
- System configuration

### 3. Install Dependencies

#### Go Dependencies
```bash
go get
# or explicitly
go get golang.org/x/oauth2 golang.org/x/oauth2/google
```

#### Frontend Dependencies
```bash
cd web/default
npm install
```

### 4. Build Everything

#### Build Frontend
```bash
cd web/default
npm run build
cd ../..
```

#### Build Go Binary
```bash
go build -ldflags "-s -w" -o one-api-en
```

### 5. Run the Service

#### Direct Run
```bash
./one-api-en --port 3000 --log-dir ./logs
```

#### Run in Screen (Recommended)
```bash
screen -dmS one-api-server bash -c './one-api-en --port 3000 --log-dir ./logs 2>&1 | tee oneapi-debug.log'
```

## Database Structure Reference

### Critical Tables

#### `channels` Table
Contains all provider configurations including:
```sql
id | name | type | key | base_url | models | config | ...
```

For Google OpenAI channels:
- `type`: 40
- `key`: Contains full service account JSON
- `base_url`: e.g., `https://us-central1-aiplatform.googleapis.com`
- `config`: Contains project_id and region
- `models`: Comma-separated list of available models

#### `tokens` Table
API access tokens for one-api:
```sql
id | user_id | key | name | ...
```

#### `options` Table
System-wide settings

## Google Cloud Service Account Details

### Service Account Persistence
The service account JSON stored in the database works from any server because:
- It's tied to the Google Cloud project, not IP addresses
- Authentication is handled via OAuth tokens generated from the JSON
- No additional firewall or access rules needed

### Service Account Structure (Redacted)
```json
{
  "type": "service_account",
  "project_id": "your-project-id",
  "private_key_id": "[REDACTED]",
  "private_key": "-----BEGIN PRIVATE KEY-----\n[REDACTED]\n-----END PRIVATE KEY-----\n",
  "client_email": "service-account@project.iam.gserviceaccount.com",
  "client_id": "[REDACTED]",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/[SERVICE_ACCOUNT_EMAIL]",
  "universe_domain": "googleapis.com"
}
```

## Verification Steps

### 1. Check Service Status
```bash
curl http://localhost:3000
# Should return HTML page
```

### 2. Test API Access
```bash
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer [YOUR_API_TOKEN]" \
  -H "Content-Type: application/json" \
  -d '{"model":"meta/llama-3.1-8b-instruct-maas","messages":[{"role":"user","content":"hi"}]}'
```

### 3. Check Logs
```bash
tail -f oneapi-debug.log
```

## Minimal Replication Checklist

- [ ] Pull correct git branch
- [ ] Copy `one-api.db` from source
- [ ] Run `go get` for dependencies
- [ ] Run `npm install` in web/default
- [ ] Run `npm run build` in web/default
- [ ] Build binary: `go build -ldflags "-s -w" -o one-api-en`
- [ ] Start service in screen session
- [ ] Verify with test API call

## Important Notes

1. **Database is Key**: The `one-api.db` file contains ALL configuration. Without it, you'll need to manually recreate all channels, tokens, and settings.

2. **Service Account Portability**: Google Cloud service accounts work from any location. The JSON credential is self-contained.

3. **Port Configuration**: Default is 3000. Change with `--port` flag if needed.

4. **Log Directory**: Ensure `./logs` directory exists or specify different path with `--log-dir`.

5. **Screen Session**: Use screen or similar to keep service running after disconnect.

## Troubleshooting

### Service won't start
- Check if port 3000 is already in use
- Ensure database file has correct permissions
- Verify all dependencies installed

### Authentication failures
- Verify service account JSON is complete in database
- Check Google Cloud project has Vertex AI API enabled
- Ensure service account has "Vertex AI User" role

### Model not found errors
- Verify model names match exactly (including periods)
- Check region-specific model availability
- Ensure channel is configured for correct region

## Summary

Replication is straightforward:
1. **Code**: Git pull gets all code changes
2. **Config**: Copy `one-api.db` for all settings
3. **Build**: Standard Go and npm build process
4. **Run**: Start with same flags as source

The database file is the critical piece - it contains all the "state" of your one-api instance including the Google Cloud service account credentials needed for Vertex AI access.