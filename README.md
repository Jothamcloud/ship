# Shipfe

Deploy frontend apps to AWS with a single command.

```bash
shipfe deploy
# → https://your-app.cdn.amazonaws.com
```

No infrastructure setup. No CI/CD. Just your code on a global CDN.

## Installation

### macOS

```bash
# Apple Silicon (M1/M2/M3)
curl -L https://github.com/jothamarinze/shipfe/releases/latest/download/shipfe-darwin-arm64 -o shipfe
chmod +x shipfe && sudo mv shipfe /usr/local/bin/

# Intel
curl -L https://github.com/jothamarinze/shipfe/releases/latest/download/shipfe-darwin-amd64 -o shipfe
chmod +x shipfe && sudo mv shipfe /usr/local/bin/
```

### Linux

```bash
# x86_64 / AMD64
curl -L https://github.com/jothamarinze/shipfe/releases/latest/download/shipfe-linux-amd64 -o shipfe
chmod +x shipfe && sudo mv shipfe /usr/local/bin/

# ARM64
curl -L https://github.com/jothamarinze/shipfe/releases/latest/download/shipfe-linux-arm64 -o shipfe
chmod +x shipfe && sudo mv shipfe /usr/local/bin/
```

### Windows

```powershell
# Download from GitHub releases
Invoke-WebRequest -Uri "https://github.com/jothamarinze/shipfe/releases/latest/download/shipfe-windows-amd64.exe" -OutFile "shipfe.exe"

# Add to PATH or move to a directory in your PATH
Move-Item shipfe.exe C:\Windows\System32\
```

### Go Install

```bash
go install github.com/jothamarinze/shipfe@latest
```

### Build from Source

```bash
git clone https://github.com/jothamarinze/shipfe.git
cd shipfe
make install
```

## Prerequisites

- AWS account with credentials configured (`aws configure`)
- Node.js (for projects that need building)

## Quick Start

```bash
cd my-react-app
shipfe deploy
```

That's it. You get back a URL.

## Commands

| Command | Description |
|---------|-------------|
| `shipfe init` | Set up infrastructure (optional, makes first deploy faster) |
| `shipfe deploy` | Build and deploy your app |
| `shipfe status` | Show deployment URL and info |
| `shipfe destroy` | Tear down everything |

### Flags

```bash
shipfe deploy --dry-run      # Preview what will happen
shipfe deploy --skip-build   # Skip the build step
shipfe destroy -y            # Skip confirmation
```

## Supported Frameworks

Auto-detected from your project:

| Framework | Build Command | Output |
|-----------|--------------|--------|
| Vite | `npm run build` | `dist/` |
| Create React App | `npm run build` | `build/` |
| Vue CLI | `npm run build` | `dist/` |
| Static HTML | — | `.` |

Works with npm, yarn, pnpm, and bun.

## How It Works

1. Detects your framework and builds the project
2. Provisions AWS infrastructure (first deploy only)
3. Uploads build files to a global CDN
4. Returns a public HTTPS URL

Subsequent deploys just sync changed files — fast.

## Configuration

Deployment state is stored in `.shipfe/config.json`. Add to `.gitignore`:

```bash
echo ".shipfe/" >> .gitignore
```

## License

MIT
