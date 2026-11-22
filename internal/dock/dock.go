// Package dock contains OS-independent dock layout and magnifier logic.
package dock

import (
	"math"
	"sort"
	"time"

	"github.com/ntars/Desop/internal/geometry"
)

// Position is one of the supported dock edges.
type Position string

const (
	PositionBottom Position = "bottom"
	PositionLeft   Position = "left"
	PositionRight  Position = "right"
)

// Item is a dock entry.
type Item struct {
	ID      string
	Label   string
	Running bool
}

// LayoutOptions controls icon placement and magnification.
type LayoutOptions struct {
	Position      Position
	Screen        geometry.Rect
	WorkArea      geometry.Rect
	Window        geometry.Rect
	IconSize      int
	Gap           int
	MaxScale      float64
	Spread        float64
	Magnification bool
	Radius        int
	Bounce        bool
	LaunchTimes   map[string]time.Time
	Now           time.Time
}

// PlacedIcon is one icon after layout and magnification have been applied.
type PlacedIcon struct {
	ID      string
	Label   string
	Rect    geometry.Rect
	Scale   float64
	Offset  int
	Running bool
}

// LayoutResult describes a complete dock frame.
type LayoutResult struct {
	Icons       []PlacedIcon
	Bounds      geometry.Rect
	TriggerZone geometry.Rect
}

// Layout places icons along the chosen edge. Magnification is limited to a
// small neighbourhood and uses smoothstep falloff so no layout state is kept.
func Layout(items []Item, pointer geometry.Point, opts LayoutOptions) LayoutResult {
	size := opts.IconSize
	if size <= 0 {
		size = 48
	}
	gap := opts.Gap
	if gap < 0 {
		gap = 0
	}
	maxScale := opts.MaxScale
	if maxScale <= 1 {
		maxScale = 1.7
	}
	spread := opts.Spread
	if spread <= 0 {
		spread = 1.2
	}
	radius := opts.Radius
	if radius < 0 {
		radius = 0
	}
	screen := opts.Screen
	if screen.Width <= 0 || screen.Height <= 0 {
		screen = geometry.NewRect(0, 0, 1920, 1080)
	}
	work := opts.WorkArea
	if work.Width <= 0 || work.Height <= 0 {
		work = screen
	}
	window := opts.Window
	if window.Width <= 0 || window.Height <= 0 {
		window = geometry.Rect{X: work.X, Y: work.Bottom() - size - radius, Width: work.Width, Height: size + radius}
	}

	total := len(items)*size + max(len(items)-1, 0)*gap
	centre := work.X + work.Width/2
	start := centre - total/2
	var icons []PlacedIcon
	var offsets []float64

	for i, item := range items {
		scale := 1.0
		if opts.Magnification {
			// Magnification is measured along the dock's layout axis.
			layoutOffset := i*(size+gap) + size/2
			axis := float64(start + layoutOffset)
			pointerAxis := float64(pointer.X)
			if opts.Position == PositionLeft || opts.Position == PositionRight {
				// Cross-axis: icons are stacked vertically, so centre them
				// vertically before comparing pointer distance.
				axis = float64(work.Y + work.Height/2 - total/2 + layoutOffset)
				pointerAxis = float64(pointer.Y)
			}
			distance := math.Abs(pointerAxis-axis) / (float64(size+gap) * spread)
			if distance < 1 {
				// smoothstep: 3t^2 - 2t^3. The centre icon gets maxScale.
				scale += (maxScale - 1) * (1 - (3*distance*distance - 2*distance*distance*distance))
			}
		}
		if opts.Bounce && opts.LaunchTimes != nil {
			if launched, ok := opts.LaunchTimes[item.ID]; ok {
				elapsed := opts.Now.Sub(launched)
				if elapsed >= 0 && elapsed < 600*time.Millisecond {
					phase := float64(elapsed) / float64(600*time.Millisecond)
					// Damped sine: starts fast, settles within 600ms.
					scale += 0.35 * math.Exp(-4*phase) * math.Abs(math.Cos(phase*math.Pi*3))
				}
			}
		}
		extra := int(math.Round(float64(size) * (scale - 1)))
		icons = append(icons, PlacedIcon{ID: item.ID, Label: item.Label, Scale: scale, Offset: extra, Running: item.Running})
		offsets = append(offsets, float64(extra))
	}

	// Shift icons adjacent to the magnified one so the dock grows in place
	// rather than overlapping its neighbours.
	cursor := start
	if opts.Position == PositionLeft || opts.Position == PositionRight {
		cursor = work.Y + work.Height/2 - total/2
	}
	for i := range icons {
		scaledSize := int(math.Round(float64(size) * icons[i].Scale))
		centreOfIcon := cursor + scaledSize/2
		var r geometry.Rect
		switch opts.Position {
		case PositionLeft:
			r = geometry.Rect{X: window.X, Y: centreOfIcon - scaledSize/2, Width: scaledSize, Height: scaledSize}
		case PositionRight:
			r = geometry.Rect{X: window.Right() - scaledSize, Y: centreOfIcon - scaledSize/2, Width: scaledSize, Height: scaledSize}
		default:
			r = geometry.Rect{X: centreOfIcon - scaledSize/2, Y: window.Bottom() - scaledSize, Width: scaledSize, Height: scaledSize}
		}
		icons[i].Rect = r
		cursor += scaledSize + gap + int(offsets[i]/2)
	}

	trigger := geometry.Rect{X: work.X, Y: screen.Bottom() - 8, Width: work.Width, Height: 8}
	switch opts.Position {
	case PositionLeft:
		trigger = geometry.Rect{X: screen.X, Y: work.Y, Width: 8, Height: work.Height}
	case PositionRight:
		trigger = geometry.Rect{X: screen.Right() - 8, Y: work.Y, Width: 8, Height: work.Height}
	}

	for i := range icons {
		icons[i].Rect = geometry.FitInside(icons[i].Rect, window)
	}
	return LayoutResult{Icons: icons, Bounds: window, TriggerZone: geometry.FitInside(trigger, screen)}
}

// RunningIDs returns the IDs of running items sorted alphabetically.
func RunningIDs(items []Item) []string {
	var out []string
	for _, item := range items {
		if item.Running {
			out = append(out, item.ID)
		}
	}
	sort.Strings(out)
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ScaleIconSize converts the configured 96-DPI icon size to the effective DPI
// of the monitor hosting a dock. Values are clamped to keep extreme system
// settings from producing unusable docks.
func ScaleIconSize(base int, dpi uint32) int {
	if base <= 0 {
		base = 48
	}
	if dpi < 96 {
		dpi = 96
	}
	scaled := int(math.Round(float64(base) * float64(dpi) / 96))
	return max(24, min(scaled, 192))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
