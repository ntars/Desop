package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadOrCreateWritesDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	got, err := LoadOrCreate(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, Default()) {
		t.Fatal("first load should return defaults")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("default config file was not created: %v", err)
	}
	reloaded, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(reloaded, Default()) {
		t.Fatal("written default config differs from in-memory defaults")
	}
}

func TestLoadOrCreatePreservesExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	c := Default()
	c.Dock.Position = "right"
	if err := Save(path, c); err != nil {
		t.Fatal(err)
	}
	got, err := LoadOrCreate(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Dock.Position != "right" {
		t.Fatalf("existing config was overwritten: %+v", got.Dock)
	}
}

func TestLoadRecoversFromCorruptConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != Default().Version {
		t.Fatalf("corrupt config did not fall back to defaults: %+v", got)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("corrupt config should be removed for rewrite: %v", err)
	}
}
