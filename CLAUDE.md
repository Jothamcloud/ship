# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build           # Build for current platform → dist/shipfe
make build-all       # Build for linux, darwin, windows (amd64 + arm64)
make test            # Run all tests
make test-coverage   # Run tests with coverage report
make lint            # Run golangci-lint
make install         # Install locally via go install
go test ./internal/detector/...  # Run tests for a single package
```

## Architecture

Shipfe is a CLI tool that deploys frontend apps to AWS using the user's own AWS credentials. It abstracts away AWS implementation details (S3, CloudFront) from users - terminal output uses generic terms like "storage", "CDN", and "deployment".

### Command Flow

The `cmd/deploy.go` orchestrates the main workflow:
1. **Detection** (`internal/detector`) - Identifies framework (Vite, React CRA, Vue, static HTML) from project files
2. **Build** (`internal/builder`) - Runs build command via detected package manager (npm/yarn/pnpm/bun)
3. **Provision/Redeploy** (`internal/aws/provisioner.go`) - Creates or updates AWS resources

First deploy: creates storage → sets up CDN → uploads files → waits for deployment to be ready.
Redeploy (when `.shipfe/config.json` exists): syncs files → invalidates cache.

### Key Packages

- `internal/aws/client.go` - Initializes AWS SDK clients
- `internal/aws/s3.go` - Storage operations, file sync with content-type detection
- `internal/aws/cloudfront.go` - CDN lifecycle, cache invalidation
- `internal/aws/provisioner.go` - Orchestrates resource creation/destruction (destroy runs silently)
- `internal/config/config.go` - Manages `.shipfe/config.json` state file
- `internal/ui/printer.go` - Colored terminal output helpers

### State Management

Deployment state is stored in `.shipfe/config.json` in the project directory to enable stable URLs across redeploys.

### UX Principles

- User-facing messages should NOT mention S3, CloudFront, OAC, or other AWS service names
- Deploy waits for CDN to be fully ready before returning the URL
- Destroy runs silently (no step-by-step output) - just "Destroying..." → "Destroyed"
- Use `--dry-run` flag to simulate deployment without creating resources
