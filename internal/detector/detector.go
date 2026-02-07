package detector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Framework represents a detected frontend framework
type Framework string

const (
	FrameworkVite     Framework = "vite"
	FrameworkReactCRA Framework = "react-cra"
	FrameworkVue      Framework = "vue"
	FrameworkStatic   Framework = "static"
	FrameworkUnknown  Framework = "unknown"
)

// DetectionResult contains the detected framework info
type DetectionResult struct {
	Framework Framework
	BuildCmd  string
	OutputDir string
}

// packageJSON represents relevant fields from package.json
type packageJSON struct {
	Scripts      map[string]string `json:"scripts"`
	Dependencies map[string]string `json:"dependencies"`
	DevDeps      map[string]string `json:"devDependencies"`
}

// Detect analyzes the project directory to detect the framework
func Detect(projectDir string) (*DetectionResult, error) {
	// Check for Vite config files
	if fileExists(projectDir, "vite.config.js") || fileExists(projectDir, "vite.config.ts") {
		return &DetectionResult{
			Framework: FrameworkVite,
			BuildCmd:  "build",
			OutputDir: "dist",
		}, nil
	}

	// Check for package.json
	pkgPath := filepath.Join(projectDir, "package.json")
	if fileExists(projectDir, "package.json") {
		pkg, err := readPackageJSON(pkgPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read package.json: %w", err)
		}

		// Check for React CRA (react-scripts)
		if hasPackage(pkg, "react-scripts") {
			return &DetectionResult{
				Framework: FrameworkReactCRA,
				BuildCmd:  "build",
				OutputDir: "build",
			}, nil
		}

		// Check for Vue
		if hasPackage(pkg, "vue") {
			return &DetectionResult{
				Framework: FrameworkVue,
				BuildCmd:  "build",
				OutputDir: "dist",
			}, nil
		}

		// Check for Vite in dependencies (fallback)
		if hasPackage(pkg, "vite") {
			return &DetectionResult{
				Framework: FrameworkVite,
				BuildCmd:  "build",
				OutputDir: "dist",
			}, nil
		}
	}

	// Check for static HTML
	if fileExists(projectDir, "index.html") {
		return &DetectionResult{
			Framework: FrameworkStatic,
			BuildCmd:  "",
			OutputDir: ".",
		}, nil
	}

	return nil, fmt.Errorf("unable to detect framework - no vite.config.js, package.json with known framework, or index.html found")
}

// fileExists checks if a file exists in the project directory
func fileExists(projectDir, filename string) bool {
	path := filepath.Join(projectDir, filename)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// readPackageJSON reads and parses package.json
func readPackageJSON(path string) (*packageJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	return &pkg, nil
}

// hasPackage checks if a package exists in dependencies or devDependencies
func hasPackage(pkg *packageJSON, name string) bool {
	if _, ok := pkg.Dependencies[name]; ok {
		return true
	}
	if _, ok := pkg.DevDeps[name]; ok {
		return true
	}
	return false
}
