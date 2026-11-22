// Package config defines the persisted Desop settings and their defaults.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Component names used in configuration and command-line flags.
const (
	ComponentDock    = "dock"
	ComponentTopBar  = "topbar"
	ComponentWidgets = "widgets"
	ComponentIsland  = "island"
)

var allComponents = []string{ComponentDock, ComponentTopBar, ComponentWidgets, ComponentIsland}

// Components returns the canonical list of configurable component names.
func Components() []string {
	out := make([]string, len(allComponents))
	copy(out, allComponents)
	return out
}

// DockConfig controls dock placement and behaviour.
type DockConfig struct {
	Enabled       bool     `json:"enabled"`
	Position      string   `json:"position"`
	IconSize      int      `json:"iconSize"`
	Magnification bool     `json:"magnification"`
	RunningDots   bool     `json:"runningDots"`
	Bounce        bool     `json:"bounce"`
	AutoHide      bool     `json:"autoHide"`
	TriggerWidth  int      `json:"triggerWidth"`
	PrimaryOnly   bool     `json:"primaryOnly"`
	Items         []string `json:"items"`
}

// TopBarConfig controls the top status bar.
type TopBarConfig struct {
	Enabled            bool `json:"enabled"`
	ShowClock          bool `json:"showClock"`
	ShowBattery        bool `json:"showBattery"`
	ShowNetwork        bool `json:"showNetwork"`
	ShowVolume         bool `json:"showVolume"`
	GlobalMenu         bool `json:"globalMenu"`
	TrayConsolidation  bool `json:"trayConsolidation"`
	AutoHideFullscreen bool `json:"autoHideFullscreen"`
}

// WidgetsConfig controls desktop widgets.
type WidgetsConfig struct {
	Enabled    bool           `json:"enabled"`
	GridSnap   bool           `json:"gridSnap"`
	GridSize   int            `json:"gridSize"`
	Note       string         `json:"note"`
	WeatherURL string         `json:"weatherURL"`
	Widgets    []WidgetConfig `json:"widgets"`
}

// WidgetConfig stores one placed widget.
type WidgetConfig struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Enabled bool   `json:"enabled"`
}

// IslandConfig controls the software dynamic island and its local API.
type IslandConfig struct {
	Enabled      bool `json:"enabled"`
	Expanded     bool `json:"expanded"`
	LocalAPI     bool `json:"localAPI"`
	LocalAPIPort int  `json:"localAPIPort"`
	MediaEvents  bool `json:"mediaEvents"`
	TimerEvents  bool `json:"timerEvents"`
}

// GeneralConfig controls application-wide behaviour.
type GeneralConfig struct {
	Autostart bool   `json:"autostart"`
	Language  string `json:"language"`
	LogLevel  string `json:"logLevel"`
}

// Config is the full settings document. Version allows future migrations.
type Config struct {
	Version int           `json:"version"`
	General GeneralConfig `json:"general"`
	Dock    DockConfig    `json:"dock"`
	TopBar  TopBarConfig  `json:"topBar"`
	Widgets WidgetsConfig `json:"widgets"`
	Island  IslandConfig  `json:"island"`
}

// Default returns a conservative offline configuration. Only the dock and
// top bar are on so a first launch is immediately useful without extra setup.
func Default() Config {
	return Config{
		Version: 1,
		General: GeneralConfig{Language: "zh-CN", LogLevel: "info"},
		Dock: DockConfig{
			Enabled:       true,
			Position:      "bottom",
			IconSize:      48,
			Magnification: true,
			RunningDots:   true,
			Bounce:        true,
			AutoHide:      false,
			TriggerWidth:  8,
			PrimaryOnly:   true,
			Items:         []string{"explorer.exe", "notepad.exe", "calc.exe"},
		},
		TopBar: TopBarConfig{
			Enabled:            true,
			ShowClock:          true,
			ShowBattery:        true,
			ShowNetwork:        true,
			ShowVolume:         true,
			GlobalMenu:         false,
			TrayConsolidation:  false,
			AutoHideFullscreen: true,
		},
		Widgets: WidgetsConfig{
			Enabled:  true,
			GridSnap: true,
			GridSize: 8,
			Widgets: []WidgetConfig{
				{ID: "clock", Kind: "clock", X: 40, Y: 80, Enabled: true},
				{ID: "status", Kind: "status", X: 300, Y: 80, Enabled: true},
				{ID: "notes", Kind: "notes", X: 40, Y: 360, Enabled: true},
			},
		},
		Island: IslandConfig{
			Enabled:      true,
			LocalAPI:     true,
			LocalAPIPort: 17890,
			MediaEvents:  true,
			TimerEvents:  true,
		},
	}
}

// Validate checks cross-field constraints without mutating c.
func (c Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported config version %d", c.Version)
	}
	switch c.Dock.Position {
	case "bottom", "left", "right":
	default:
		return fmt.Errorf("unsupported dock position %q", c.Dock.Position)
	}
	if c.Dock.IconSize < 24 || c.Dock.IconSize > 128 {
		return fmt.Errorf("dock iconSize %d out of range [24,128]", c.Dock.IconSize)
	}
	if c.Dock.TriggerWidth < 2 || c.Dock.TriggerWidth > 64 {
		return fmt.Errorf("dock triggerWidth %d out of range [2,64]", c.Dock.TriggerWidth)
	}
	if c.Widgets.GridSize < 4 || c.Widgets.GridSize > 64 {
		return fmt.Errorf("widget gridSize %d out of range [4,64]", c.Widgets.GridSize)
	}
	if c.Island.LocalAPIPort < 1024 || c.Island.LocalAPIPort > 65535 {
		return fmt.Errorf("island localAPIPort %d out of range [1024,65535]", c.Island.LocalAPIPort)
	}
	seen := make(map[string]bool, len(c.Widgets.Widgets))
	for _, w := range c.Widgets.Widgets {
		if w.ID == "" {
			return errors.New("widget id must not be empty")
		}
		if seen[w.ID] {
			return fmt.Errorf("duplicate widget id %q", w.ID)
		}
		seen[w.ID] = true
		switch w.Kind {
		case "clock", "status", "notes", "weather":
		case "calendar":
		default:
			return fmt.Errorf("unsupported widget kind %q", w.Kind)
		}
	}
	return nil
}

// EnabledComponents returns the enabled component names in canonical order.
func (c Config) EnabledComponents() []string {
	on := map[string]bool{
		ComponentDock:    c.Dock.Enabled,
		ComponentTopBar:  c.TopBar.Enabled,
		ComponentWidgets: c.Widgets.Enabled,
		ComponentIsland:  c.Island.Enabled,
	}
	var out []string
	for _, name := range allComponents {
		if on[name] {
			out = append(out, name)
		}
	}
	return out
}

// SetComponent enables or disables one component.
func (c *Config) SetComponent(name string, enabled bool) error {
	switch name {
	case ComponentDock:
		c.Dock.Enabled = enabled
	case ComponentTopBar:
		c.TopBar.Enabled = enabled
	case ComponentWidgets:
		c.Widgets.Enabled = enabled
	case ComponentIsland:
		c.Island.Enabled = enabled
	default:
		return fmt.Errorf("unknown component %q (valid: %s)", name, strings.Join(allComponents, ", "))
	}
	return nil
}

// MarshalPretty serializes configuration deterministically.
func MarshalPretty(c Config) ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// Parse validates a JSON document and returns the decoded configuration.
func Parse(data []byte) (Config, error) {
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if err := c.Validate(); err != nil {
		return Config{}, err
	}
	return c, nil
}

// Load reads config from path. A missing file returns defaults without
// creating it, so read-only diagnostics never mutate user data.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	return Parse(data)
}

// LoadOrCreate loads the config and writes defaults when the file is absent.
// This is used by the desktop host so a first run leaves a real, editable file.
func LoadOrCreate(path string) (Config, error) {
	c, err := Load(path)
	if err != nil {
		return Config{}, err
	}
	if _, statErr := os.Stat(path); statErr == nil {
		return c, nil
	}
	if err := Save(path, c); err != nil {
		return Config{}, fmt.Errorf("create default config: %w", err)
	}
	return c, nil
}

// Save atomically writes config to path, creating its parent directory.
func Save(path string, c Config) error {
	if err := c.Validate(); err != nil {
		return err
	}
	data, err := MarshalPretty(c)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".config-*.json")
	if err != nil {
		return fmt.Errorf("create temporary config: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return fmt.Errorf("write temporary config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary config: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}
	return nil
}

// ConfigDir returns the per-user configuration directory. callers pass the
// OS-specific base directory (on Windows: os.Getenv("APPDATA")).
func ConfigDir(base string) string {
	return filepath.Join(base, "Desop")
}

// ConfigPath returns the full settings file path for an OS-specific base dir.
func ConfigPath(base string) string {
	return filepath.Join(ConfigDir(base), "config.json")
}

// SortNames returns component names sorted lexicographically.
func SortNames(names []string) []string {
	out := append([]string(nil), names...)
	sort.Strings(out)
	return out
}
