package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDefaultValid(t *testing.T) {
	c := Default()
	if err := c.Validate(); err != nil {
		t.Fatalf("default config is invalid: %v", err)
	}
	got := c.EnabledComponents()
	want := []string{ComponentDock, ComponentTopBar, ComponentWidgets, ComponentIsland}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("enabled components = %v, want %v", got, want)
	}
}

func TestSetComponent(t *testing.T) {
	c := Default()
	if err := c.SetComponent(ComponentDock, false); err != nil {
		t.Fatal(err)
	}
	if err := c.SetComponent("bogus", true); err == nil {
		t.Fatal("expected error for unknown component")
	}
	got := c.EnabledComponents()
	want := []string{ComponentTopBar, ComponentWidgets, ComponentIsland}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("enabled components = %v, want %v", got, want)
	}
}

func TestValidateRejectsBadValues(t *testing.T) {
	cases := []func(*Config){
		func(c *Config) { c.Version = 2 },
		func(c *Config) { c.Dock.Position = "top" },
		func(c *Config) { c.Dock.IconSize = 4 },
		func(c *Config) { c.Dock.TriggerWidth = 100 },
		func(c *Config) { c.Widgets.GridSize = 1 },
		func(c *Config) { c.Island.LocalAPIPort = 80 },
		func(c *Config) {
			c.Widgets.Widgets = append(c.Widgets.Widgets, WidgetConfig{ID: "clock", Kind: "notes"})
		},
		func(c *Config) { c.Widgets.Widgets[0].Kind = "unknown" },
	}
	for i, mutate := range cases {
		c := Default()
		mutate(&c)
		if err := c.Validate(); err == nil {
			t.Errorf("case %d: expected validation failure", i)
		}
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := ConfigPath(dir)
	c := Default()
	c.Dock.Position = "left"
	c.Widgets.Note = "hello\nworld"
	if err := Save(path, c); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Dock.Position != "left" || got.Widgets.Note != "hello\nworld" {
		t.Fatalf("round trip mismatch: %+v", got)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "\"version\": 1") {
		t.Fatalf("saved config is not pretty JSON: %s", data)
	}
}

func TestLoadMissingReturnsDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, Default()) {
		t.Fatal("missing config should return defaults")
	}
}
