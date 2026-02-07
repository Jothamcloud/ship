package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectVite(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "shipfe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create vite.config.js
	if err := os.WriteFile(filepath.Join(tmpDir, "vite.config.js"), []byte("export default {}"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Framework != FrameworkVite {
		t.Errorf("expected framework %s, got %s", FrameworkVite, result.Framework)
	}
	if result.OutputDir != "dist" {
		t.Errorf("expected output dir 'dist', got '%s'", result.OutputDir)
	}
}

func TestDetectViteTS(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shipfe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "vite.config.ts"), []byte("export default {}"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Framework != FrameworkVite {
		t.Errorf("expected framework %s, got %s", FrameworkVite, result.Framework)
	}
}

func TestDetectReactCRA(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shipfe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	pkgJSON := `{
		"dependencies": {
			"react": "^18.0.0",
			"react-scripts": "5.0.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Framework != FrameworkReactCRA {
		t.Errorf("expected framework %s, got %s", FrameworkReactCRA, result.Framework)
	}
	if result.OutputDir != "build" {
		t.Errorf("expected output dir 'build', got '%s'", result.OutputDir)
	}
}

func TestDetectVue(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shipfe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	pkgJSON := `{
		"dependencies": {
			"vue": "^3.0.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Framework != FrameworkVue {
		t.Errorf("expected framework %s, got %s", FrameworkVue, result.Framework)
	}
}

func TestDetectStatic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shipfe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte("<html></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Framework != FrameworkStatic {
		t.Errorf("expected framework %s, got %s", FrameworkStatic, result.Framework)
	}
	if result.OutputDir != "." {
		t.Errorf("expected output dir '.', got '%s'", result.OutputDir)
	}
	if result.BuildCmd != "" {
		t.Errorf("expected empty build cmd, got '%s'", result.BuildCmd)
	}
}

func TestDetectUnknown(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shipfe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Empty directory
	_, err = Detect(tmpDir)
	if err == nil {
		t.Error("expected error for empty directory")
	}
}
