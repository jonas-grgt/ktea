package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveDebugLogPath(t *testing.T) {
	dir := t.TempDir()
	existingFile := filepath.Join(dir, "existing.log")
	assert.NoError(t, os.WriteFile(existingFile, nil, 0o600))
	missingDir := filepath.Join(dir, "missing") + string(os.PathSeparator)

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty path uses default file", "", "debug.log"},
		{"existing directory", dir, filepath.Join(dir, "debug.log")},
		{"path with trailing separator", missingDir, filepath.Join(dir, "missing", "debug.log")},
		{"existing file", existingFile, existingFile},
		{"new file in existing directory", filepath.Join(dir, "ktea.log"), filepath.Join(dir, "ktea.log")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, resolveDebugLogPath(tt.in))
		})
	}
}

func TestDebugFlag(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantEnabled bool
		wantPath    string
	}{
		{"not set", nil, false, ""},
		{"bare flag", []string{"--debug"}, true, ""},
		{"explicit true", []string{"--debug=true"}, true, ""},
		{"explicit false", []string{"--debug=false"}, false, ""},
		{"file path", []string{"--debug=/x/y.log"}, true, "/x/y.log"},
		{"directory path", []string{"-debug=logs/"}, true, "logs/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d debugFlag
			fs := flag.NewFlagSet("t", flag.ContinueOnError)
			fs.Var(&d, "debug", "")
			assert.NoError(t, fs.Parse(tt.args))
			assert.Equal(t, tt.wantEnabled, d.enabled)
			assert.Equal(t, tt.wantPath, d.path)
		})
	}

	t.Run("empty value is rejected", func(t *testing.T) {
		var d debugFlag
		fs := flag.NewFlagSet("t", flag.ContinueOnError)
		fs.SetOutput(new(nopWriter))
		fs.Var(&d, "debug", "")
		assert.Error(t, fs.Parse([]string{"--debug="}))
	})
}

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }
