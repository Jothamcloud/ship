package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// PackageManager represents a Node.js package manager
type PackageManager string

const (
	PackageManagerNPM   PackageManager = "npm"
	PackageManagerYarn  PackageManager = "yarn"
	PackageManagerPnpm  PackageManager = "pnpm"
	PackageManagerBun   PackageManager = "bun"
)

// Builder handles project build execution
type Builder struct {
	projectDir     string
	packageManager PackageManager
}

// New creates a new Builder for the given project directory
func New(projectDir string) *Builder {
	return &Builder{
		projectDir:     projectDir,
		packageManager: detectPackageManager(projectDir),
	}
}

// Build runs the build command for the project
func (b *Builder) Build(scriptName string) error {
	if scriptName == "" {
		return nil // No build needed (e.g., static site)
	}

	cmd := b.buildCommand(scriptName)
	return b.run(cmd[0], cmd[1:]...)
}

// buildCommand returns the full command to run for the given npm script
func (b *Builder) buildCommand(scriptName string) []string {
	switch b.packageManager {
	case PackageManagerYarn:
		return []string{"yarn", scriptName}
	case PackageManagerPnpm:
		return []string{"pnpm", "run", scriptName}
	case PackageManagerBun:
		return []string{"bun", "run", scriptName}
	default:
		return []string{"npm", "run", scriptName}
	}
}

// run executes a command with streaming output
func (b *Builder) run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = b.projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s %v: %w", name, args, err)
	}
	return nil
}

// detectPackageManager determines which package manager to use based on lock files
func detectPackageManager(projectDir string) PackageManager {
	lockFiles := map[string]PackageManager{
		"bun.lockb":         PackageManagerBun,
		"pnpm-lock.yaml":    PackageManagerPnpm,
		"yarn.lock":         PackageManagerYarn,
		"package-lock.json": PackageManagerNPM,
	}

	// Check in priority order
	for lockFile, pm := range lockFiles {
		path := filepath.Join(projectDir, lockFile)
		if _, err := os.Stat(path); err == nil {
			return pm
		}
	}

	// Default to npm
	return PackageManagerNPM
}

// GetPackageManager returns the detected package manager
func (b *Builder) GetPackageManager() PackageManager {
	return b.packageManager
}
