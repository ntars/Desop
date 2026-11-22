// Package topbar contains OS-independent status aggregation and layout.
package topbar

import (
	"fmt"
	"strings"
	"time"

	"github.com/ntars/Desop/internal/config"
	"github.com/ntars/Desop/internal/geometry"
)

// BatteryState is the coarse charging state shown in the status area.
type BatteryState string

const (
	BatteryUnknown     BatteryState = "unknown"
	BatteryCharging    BatteryState = "charging"
	BatteryDischarging BatteryState = "discharging"
	BatteryFull        BatteryState = "full"
)

// Status contains read-only system values shown by the top bar.
type Status struct {
	Now          time.Time
	BatteryLevel int
	BatteryState BatteryState
	Online       bool
	Volume       int
	ActiveApp    string
}

// LayoutOptions controls top bar geometry.
type LayoutOptions struct {
	Screen    geometry.Rect
	WorkArea  geometry.Rect
	Height    int
	Padding   int
	Gap       int
	FontWidth int
}

// LayoutResult describes the bar and its three regions.
type LayoutResult struct {
	Bar    geometry.Rect
	Menu   geometry.Rect
	Status geometry.Rect
	Tray   geometry.Rect
}

// Layout places the bar above the work area and splits it into menu, status
// and tray regions. The centre region is clamped so it never overlaps.
func Layout(cfg config.TopBarConfig, opts LayoutOptions) LayoutResult {
	screen := opts.Screen
	if screen.Width <= 0 || screen.Height <= 0 {
		screen = geometry.NewRect(0, 0, 1920, 1080)
	}
	work := opts.WorkArea
	if work.Width <= 0 || work.Height <= 0 {
		work = screen
	}
	height := opts.Height
	if height <= 0 {
		height = 32
	}
	padding := opts.Padding
	if padding < 0 {
		padding = 8
	}
	gap := opts.Gap
	if gap < 0 {
		gap = 12
	}
	// Place the bar at the top of the monitor, overlapping the work area when
	// Windows has already reserved space for a taskbar there.
	barY := screen.Y
	bar := geometry.NewRect(screen.X, barY, screen.Right(), barY+height)
	trayWidth := 140
	if !cfg.TrayConsolidation {
		trayWidth = padding
	}
	statusWidth := screen.Width / 4
	if statusWidth < 120 {
		statusWidth = 120
	}
	menuRight := bar.X + bar.Width - trayWidth - statusWidth - gap
	if menuRight < bar.X+padding {
		menuRight = bar.X + padding
	}
	return LayoutResult{
		Bar:    bar,
		Menu:   geometry.NewRect(bar.X+padding, bar.Y+4, menuRight, bar.Bottom()-4),
		Status: geometry.NewRect(menuRight+gap, bar.Y+4, bar.Right()-trayWidth-gap/2, bar.Bottom()-4),
		Tray:   geometry.NewRect(bar.Right()-trayWidth, bar.Y+4, bar.Right()-padding, bar.Bottom()-4),
	}
}

// RenderStatus builds the status text in a stable left-to-right order.
// Without a battery, no battery segment is invented.
func RenderStatus(cfg config.TopBarConfig, s Status) string {
	var parts []string
	if cfg.ShowClock {
		parts = append(parts, s.Now.Format("15:04"))
		parts = append(parts, s.Now.Format("2006-01-02"))
	}
	if cfg.ShowBattery && s.BatteryState != BatteryUnknown && s.BatteryState != "" {
		level := geometry.Clamp(s.BatteryLevel, 0, 100)
		parts = append(parts, fmt.Sprintf("%s %d%%", s.BatteryState, level))
	}
	if cfg.ShowNetwork {
		if s.Online {
			parts = append(parts, "net:on")
		} else {
			parts = append(parts, "net:off")
		}
	}
	if cfg.ShowVolume {
		parts = append(parts, fmt.Sprintf("vol:%d", geometry.Clamp(s.Volume, 0, 100)))
	}
	return strings.Join(parts, "  ")
}

// RenderMenu returns the active application name when global menu is enabled.
// Windows applications are not required to expose menus; an empty app name
// intentionally yields no menu rather than a fabricated one.
func RenderMenu(cfg config.TopBarConfig, activeApp string) string {
	if !cfg.GlobalMenu || strings.TrimSpace(activeApp) == "" {
		return ""
	}
	return activeApp
}

// ShouldAutoHide implements the documented conservative policy: hide only for
// a confirmed fullscreen window, never because a window merely looks large.
func ShouldAutoHide(cfg config.TopBarConfig, confirmedFullscreen bool) bool {
	return cfg.AutoHideFullscreen && confirmedFullscreen
}
