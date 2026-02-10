# Changelog

All notable changes to the CLIProxyAPIPlus project (gf branch) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

#### Enhanced Logging
- Request ID tracking for AI API requests (v1/chat/completions, v1/completions, v1/messages, v1/responses)
- Display provider, model, and account information in log output
- Structured log fields for easy log analysis and filtering
- Log format: `[timestamp] [level] | request_id | status | latency | client_ip | method | path | provider=X | model=Y | account=Z`

#### Embedded Management UI
- Management UI embedded directly into the binary as embedded/management.html
- Offline deployment support - no network download required
- Faster startup with instant UI availability
- Version consistency between UI and backend

#### iFlow API Support
- Session ID generation for iFlow API requests
- HMAC-SHA256 signature generation for request authentication

#### Automated Build Script
- Automated build script (build.sh) with the following features:
  - Check git updates, automatically exit if no new commits
  - Automatically stop/start supervisor service
  - Force build mode with `-f` parameter
  - Automatically build and embed React frontend
  - Auto-install npm dependencies on first build
  - Inject version information into binary
  - Proxychains support for pulling code (removed in later version)

#### Web Submodule
- Web frontend integrated as git submodule
- Separate repository for web development
- Automatic submodule updates

### Changed

#### Build Process
- Simplified code pulling logic
- Optimized submodule update process
- Protected untracked configuration files during updates
- Removed proxychains dependency from build.sh
- Fixed pull_code function to use correct remote branch

#### Logging
- Fixed channel display issue when multiple channels have the same model
- Provider and model information only shown when model is available
- Reads actual provider from gin.Context instead of inferring from model name
- Reads actual model from gin.Context (set by auth manager) as primary source, fallback to request body extraction

#### Auth Persistence
- Fixed disabled state persistence across restarts
- Auth metadata properly merged into auth files
- Ensures disabled auths remain disabled after service restart

#### Management UI
- Improved asset sync handling
- Synchronous availability of management.html
- Better error handling during asset updates

#### SDK Access
- Simplified provider lifecycle and registration logic
- Updated error handling and types
- Improved registry management

### Fixed

- Fixed git log command entering interactive mode in build.sh
- Fixed `-f` parameter not being passed to main function in build.sh
- Fixed leftover merge conflict marker
- Fixed pull_code function issues
- Fixed SSE model name rewriting for Responses API
- Fixed management.html availability issues
- Fixed logging where provider/model/account information was not displayed
- Resolved issue where request body being read by other middlewares prevented model extraction

### Technical Details

#### Logging Implementation
**Files Modified:**
- `internal/logging/gin_logger.go`: Enhanced to extract and log provider, model, and account information
- `sdk/cliproxy/auth/conductor.go`: Stores routeModel, provider, and account info into gin.Context

**Key Changes:**
- In `executeMixedOnce`, `executeCountMixedOnce`, and `executeStreamMixedOnce`:
  ```go
  if ginCtx := ctx.Value("gin"); ginCtx != nil {
      if c, ok := ginCtx.(*gin.Context); ok {
          c.Set("cliproxy.provider", provider)
          c.Set("cliproxy.model", routeModel)
      }
  }
  ```
- In `GinLogrusLogger`:
  ```go
  // First try to get model from gin.Context (set by auth manager)
  model := ""
  if modelVal, exists := c.Get("cliproxy.model"); exists {
      if modelStr, ok := modelVal.(string); ok {
          model = modelStr
      }
  }
  // Fallback to extraction from request body
  if model == "" {
      model = extractModelFromRequest(c)
  }
  ```

#### Auth Persistence Implementation
**Files Modified:**
- `sdk/auth/filestore.go`: Added `mergeMetadataIntoFile()` function
- `internal/watcher/synthesizer/file.go`: Properly handle disabled state

**Key Changes:**
```go
// After saving storage, merge metadata (like disabled state) into the file
if auth.Metadata != nil && len(auth.Metadata) > 0 {
    if err = s.mergeMetadataIntoFile(path, auth); err != nil {
        return "", fmt.Errorf("auth filestore: merge metadata failed: %w", err)
    }
}
```

#### Build Script Improvements
**Files Modified:**
- `build.sh`: Complete rewrite with automation features

**Key Features:**
- Git update checking before build
- Automatic service management
- Web frontend building and embedding
- Version injection
- Submodule management

### Breaking Changes

None

### Migration Notes

#### From Main Branch to GF Branch

1. **Update README**: The gf branch includes enhanced features not present in main
2. **Build Process**: Use `./build.sh` instead of manual `go build`
3. **Web Deployment**: Web is now a git submodule, update it with:
   ```bash
   git submodule update --remote web
   ```
4. **Logging Format**: Logs now include additional fields (provider, model, account)
5. **Management UI**: Access via embedded `/management.html` instead of downloading

### Contributors

- Mainline: router-for-me/CLIProxyAPI
- Plus features: Community contributors
- iFlow integration: router-for-me
- GitHub Copilot: em4go
- Kiro integration: fuko2935, Ravens2121

### Statistics

**Changes from main branch:**
- 32 files changed
- 1,477 insertions(+)
- 650 deletions(-)
- Net: +827 lines

**Key Categories:**
- Logging: 129 insertions (gin_logger.go)
- Build Script: 236 insertions (build.sh)
- Management UI: 193 insertions (internal/managementasset/)
- Auth/SDK: 288 insertions across multiple files
- Documentation: Updated README files