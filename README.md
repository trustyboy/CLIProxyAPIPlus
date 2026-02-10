# CLIProxyAPI Plus

English | [Chinese](README_CN.md)

This is the Plus version of [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI), adding support for third-party providers on top of the mainline project.

All third-party provider support is maintained by community contributors; CLIProxyAPI does not provide technical support. Please contact the corresponding community maintainer if you need assistance.

The Plus release stays in lockstep with the mainline features.

## Differences from the Mainline

- Added GitHub Copilot support (OAuth login), provided by [em4go](https://github.com/em4go/CLIProxyAPI/tree/feature/github-copilot-auth)
- Added Kiro (AWS CodeWhisperer) support (OAuth login), provided by [fuko2935](https://github.com/fuko2935/CLIProxyAPI/tree/feature/kiro-integration), [Ravens2121](https://github.com/Ravens2121/CLIProxyAPIPlus/)

## New Features (Plus Enhanced)

### Enhanced Logging

**Detailed Request Logging**
- Display request ID, model name, and provider information in logs
- Record actual channel account information used
- Support structured log output for easy log analysis

**Log Format Example**
```
[2025-01-28 04:00:00] [info ] | a1b2c3d4 | 200 |       23.559s | 192.168.1.100 | POST | /v1/chat/completions | provider=gemini | model=gemini-pro | account=oauth:user@example.com
```

**Key Improvements**
- Fixed channel display issue when multiple channels have the same model
- Provider and model information are only shown when model is available
- Reads actual provider from gin.Context instead of inferring from model name

### Embedded Management UI

**Self-contained Dashboard**
- Management UI is now embedded directly into the binary
- No network download required for the dashboard
- Offline deployment support
- Faster startup with instant UI availability
- Version consistency between UI and backend

**Access the Dashboard**
```
http://your-server:8317/management.html
```

**Key Benefits**
- Zero network dependency - works offline
- Faster startup - no download wait time
- Enhanced security - no MITM attack risk during download
- Version consistency - UI and backend are always in sync
- Simplified deployment - single binary contains everything

### Automated Build Script

**build.sh Features**
- Check git updates, automatically exit if no new commits
- Automatically stop/start supervisor service
- Support proxychains for pulling code
- Inject version information into binary
- Support `-f` parameter for forced builds
- **NEW**: Automatically build and embed React frontend
- **NEW**: Auto-install npm dependencies on first build
- **NEW**: Validate embedded HTML file

**Usage**
```bash
# Normal mode (check for updates)
./build.sh

# Force build mode (skip update check)
./build.sh -f
```

**Build Process**
```
1. Check git updates
   ↓
2. Stop service
   ↓
3. Pull code
   ↓
4. Build Web frontend (npm run build)
   ↓
5. Copy to embed directory
   ↓
6. Compile Go binary
   ↓
7. Start service
```

**Configuration Options**
- `PROXY_CHAINS_CMD`: proxychains command (default: proxychains)
- `SERVICE_NAME`: supervisor service name
- `OUTPUT_NAME`: output file name
- `OUTPUT_DIR`: output directory

### Other Enhanced Features

- **Embedded Management UI**: Self-contained dashboard embedded in binary (no download required)
- **OAuth Web Authentication**: Browser-based OAuth login for Kiro with beautiful web UI
- **Rate Limiter**: Built-in request rate limiting to prevent API abuse
- **Background Token Refresh**: Automatic token refresh 10 minutes before expiration
- **Metrics & Monitoring**: Request metrics collection for monitoring and debugging
- **Device Fingerprint**: Device fingerprint generation for enhanced security
- **Cooldown Management**: Smart cooldown mechanism for API rate limits
- **Usage Checker**: Real-time usage monitoring and quota management
- **Model Converter**: Unified model name conversion across providers
- **UTF-8 Stream Processing**: Improved streaming response handling
- **Auth File Persistence**: Fixed disabled state persistence across restarts

## Kiro Authentication

### Web-based OAuth Login

Access the Kiro OAuth web interface at:

```
http://your-server:8080/v0/oauth/kiro
```

This provides a browser-based OAuth flow for Kiro (AWS CodeWhisperer) authentication with:
- AWS Builder ID login
- AWS Identity Center (IDC) login
- Token import from Kiro IDE

## Quick Start

### Docker Deployment

```bash
# Create deployment directory
mkdir -p ~/cli-proxy && cd ~/cli-proxy

# Create docker-compose.yml
cat > docker-compose.yml << 'EOF'
services:
  cli-proxy-api:
    image: eceasy/cli-proxy-api-plus:latest
    container_name: cli-proxy-api-plus
    ports:
      - "8317:8317"
    volumes:
      - ./config.yaml:/CLIProxyAPI/config.yaml
      - ./auths:/root/.cli-proxy-api
      - ./logs:/CLIProxyAPI/logs
    restart: unless-stopped
EOF

# Download example config
curl -o config.yaml https://raw.githubusercontent.com/router-for-me/CLIProxyAPIPlus/main/config.example.yaml

# Start service
docker compose up -d
```

### Source Code Deployment

```bash
# Clone repository
git clone https://github.com/trustyboy/CLIProxyAPIPlus.git
cd CLIProxyAPIPlus

# Switch to gf branch (enhanced features branch)
git checkout gf

# Use automated build script
./build.sh

# Or build manually
go build -o cli-proxy-api ./cmd/server

# Configure config.yaml
cp config.example.yaml config.yaml
vim config.yaml

# Start service
./cli-proxy-api
```

### Configuration

Edit `config.yaml` before starting:

```yaml
# Basic configuration example
server:
  port: 8317

# Add your provider configurations here
```

### Update to Latest Version

```bash
cd ~/cli-proxy
docker compose pull && docker compose up -d
```

## API Endpoints

### OpenAI Compatible APIs

- `POST /v1/chat/completions` - Chat completions
- `POST /v1/completions` - Text completions
- `GET /v1/models` - Model list

### Gemini Compatible APIs

- `POST /v1beta/models/{model}:generateContent` - Content generation
- `GET /v1beta/models` - Model list

### Claude Compatible APIs

- `POST /v1/messages` - Messages API
- `GET /v1/models` - Model list

## Logging

### Log Levels

- `INFO` - Normal requests
- `WARN` - 4xx errors
- `ERROR` - 5xx errors

### Log Fields

- `request_id` - Request ID (for AI API requests)
- `status` - HTTP status code
- `latency` - Request duration
- `client_ip` - Client IP address
- `method` - HTTP method
- `path` - Request path
- `provider` - Provider (shown when model is available)
- `model` - Model name (shown when model is available)
- `account` - Channel account (shown when account info is available)
- `error` - Error message (shown when error occurs)

## Version Information

Check version:
```bash
./cli-proxy-api -version
```

## Contributing

This project only accepts pull requests that relate to third-party provider support. Any pull requests unrelated to third-party provider support will be rejected.

If you need to submit any non-third-party provider changes, please open them against the [mainline](https://github.com/router-for-me/CLIProxyAPI) repository.

## Changelog

For detailed change history and technical implementation details, see [CHANGELOG.md](CHANGELOG.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
