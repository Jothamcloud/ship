package builder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPackageManagerNPM(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shipfe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package-lock.json
	if err := os.WriteFile(filepath.Join(tmpDir, "package-lock.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	pm := detectPackageManager(tmpDir)
	if pm != PackageManagerNPM {
		t.Errorf("expected npm, got %s", pm)
	}
}

func TestDetectPackageManagerYarn(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shipfe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "yarn.lock"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	pm := detectPackageManager(tmpDir)
	if pm != PackageManagerYarn {
		t.Errorf("expected yarn, got %s", pm)
	}
}

func TestDetectPackageManagerPnpm(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shipfe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "pnpm-lock.yaml"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	pm := detectPackageManager(tmpDir)
	if pm != PackageManagerPnpm {
		t.Errorf("expected pnpm, got %s", pm)
	}
}

func TestDetectPackageManagerBun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shipfe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "bun.lockb"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	pm := detectPackageManager(tmpDir)
	if pm != PackageManagerBun {
		t.Errorf("expected bun, got %s", pm)
	}
}

func TestDetectPackageManagerDefault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shipfe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// No lock file
	pm := detectPackageManager(tmpDir)
	if pm != PackageManagerNPM {
		t.Errorf("expected npm as default, got %s", pm)
	}
}

func TestBuildCommand(t *testing.T) {
	tests := []struct {
		pm       PackageManager
		script   string
		expected []string
	}{
		{PackageManagerNPM, "build", []string{"npm", "run", "build"}},
		{PackageManagerYarn, "build", []string{"yarn", "build"}},
		{PackageManagerPnpm, "build", []string{"pnpm", "run", "build"}},
		{PackageManagerBun, "build", []string{"bun", "run", "build"}},
	}

	for _, tc := range tests {
		b := &Builder{packageManager: tc.pm}
		cmd := b.buildCommand(tc.script)

		if len(cmd) != len(tc.expected) {
			t.Errorf("for %s: expected %v, got %v", tc.pm, tc.expected, cmd)
			continue
		}

		for i := range cmd {
			if cmd[i] != tc.expected[i] {
				t.Errorf("for %s: expected %v, got %v", tc.pm, tc.expected, cmd)
				break
			}
		}
	}
}
