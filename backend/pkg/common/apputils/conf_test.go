package apputils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigReadsYAMLIntoTargetStruct(t *testing.T) {
	type nested struct {
		Enabled bool `yaml:"enabled"`
		Count   int  `yaml:"count"`
	}
	type config struct {
		Name   string `yaml:"name"`
		Nested nested `yaml:"nested"`
	}

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("name: smart-note\nnested:\n  enabled: true\n  count: 3\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var got config
	LoadConfig(&got, path)

	if got.Name != "smart-note" || !got.Nested.Enabled || got.Nested.Count != 3 {
		t.Fatalf("LoadConfig() = %+v, want populated config", got)
	}
}
