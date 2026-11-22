// Package widgets contains OS-independent widget placement and rendering text.
package widgets

import (
	"fmt"
	"strings"
	"time"

	"github.com/ntars/Desop/internal/config"
	"github.com/ntars/Desop/internal/geometry"
)

// SystemStats are read-only values exposed by the OS layer.
type SystemStats struct {
	CPUPercent    float64
	MemoryPercent float64
	DiskPercent   float64
	RxBytesPerSec uint64
	TxBytesPerSec uint64
}

// PlacedWidget is a widget rect after snapping and screen clamping.
type PlacedWidget struct {
	ID    string
	Kind  string
	Rect  geometry.Rect
	Title string
	Lines []string
}

// PlaceAll computes widget rectangles from persisted positions.
func PlaceAll(cfg config.WidgetsConfig, screen geometry.Rect, sizes map[string]geometry.Rect) []PlacedWidget {
	if cfg.GridSnap {
		grid := cfg.GridSize
		if grid <= 0 {
			grid = 8
		}
		for i := range cfg.Widgets {
			size := sizes[cfg.Widgets[i].Kind]
			if size.Width <= 0 || size.Height <= 0 {
				size = geometry.Rect{Width: 240, Height: 180}
			}
			maxX := screen.Right() - size.Width
			maxY := screen.Bottom() - size.Height
			if maxX < screen.X {
				maxX = screen.X
			}
			if maxY < screen.Y {
				maxY = screen.Y
			}
			cfg.Widgets[i].X = geometry.Clamp(cfg.Widgets[i].X/grid*grid, screen.X, maxX)
			cfg.Widgets[i].Y = geometry.Clamp(cfg.Widgets[i].Y/grid*grid, screen.Y, maxY)
		}
	}
	out := make([]PlacedWidget, 0, len(cfg.Widgets))
	for _, w := range cfg.Widgets {
		size := sizes[w.Kind]
		if size.Width <= 0 || size.Height <= 0 {
			size = geometry.Rect{Width: 240, Height: 180}
		}
		rect := geometry.FitInside(geometry.Rect{X: w.X, Y: w.Y, Width: size.Width, Height: size.Height}, screen)
		out = append(out, PlacedWidget{ID: w.ID, Kind: w.Kind, Rect: rect})
	}
	return out
}

// HitTest returns the topmost widget at p (the last placed widget wins).
func HitTest(placed []PlacedWidget, p geometry.Point) *PlacedWidget {
	for i := len(placed) - 1; i >= 0; i-- {
		if placed[i].Rect.Contains(p) {
			return &placed[i]
		}
	}
	return nil
}

// MoveBy applies a drag delta and clamps the widget inside the screen.
func MoveBy(w config.WidgetConfig, dx, dy int, screen geometry.Rect, size geometry.Rect, snap bool, grid int) config.WidgetConfig {
	w.X += dx
	w.Y += dy
	if snap && grid > 0 {
		w.X = (w.X + grid/2) / grid * grid
		w.Y = (w.Y + grid/2) / grid * grid
	}
	w.X = geometry.Clamp(w.X, screen.X, screen.Right()-size.Width)
	w.Y = geometry.Clamp(w.Y, screen.Y, screen.Bottom()-size.Height)
	return w
}

// Render builds the visible lines for one widget. Weather stays offline
// unless the user configured a data source.
func Render(kind string, note string, now time.Time, stats SystemStats, weatherURL string) (string, []string) {
	switch kind {
	case "clock":
		return "Clock", []string{now.Format("15:04:05"), now.Format("2006-01-02 Monday")}
	case "calendar":
		return "Calendar", RenderMonth(now)
	case "status":
		return "System", []string{
			fmt.Sprintf("CPU   %.0f%%", geometry.ClampIntFloat(stats.CPUPercent)),
			fmt.Sprintf("MEM   %.0f%%", geometry.ClampIntFloat(stats.MemoryPercent)),
			fmt.Sprintf("DISK  %.0f%%", geometry.ClampIntFloat(stats.DiskPercent)),
			fmt.Sprintf("NET   ↓ %s/s  ↑ %s/s", humanBytes(stats.RxBytesPerSec), humanBytes(stats.TxBytesPerSec)),
		}
	case "notes":
		lines := strings.Split(strings.ReplaceAll(note, "\r\n", "\n"), "\n")
		if len(lines) == 0 {
			lines = []string{""}
		}
		return "Notes", lines
	case "weather":
		if strings.TrimSpace(weatherURL) == "" {
			return "Weather", []string{"No data source configured", "Offline by default"}
		}
		return "Weather", []string{"Data source configured", "Fetching is disabled in MVP"}
	default:
		return "", nil
	}
}

// RenderMonth returns a compact month grid using Monday as the first day.
func RenderMonth(now time.Time) []string {
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	offset := int(first.Weekday())
	if offset == 0 {
		offset = 7
	}
	offset-- // Monday first
	days := first.AddDate(0, 1, -1).Day()
	lines := []string{now.Format("2006-01"), "Mo Tu We Th Fr Sa Su"}
	row := make([]string, 7)
	for i := range row {
		row[i] = "  "
	}
	for i := 0; i < offset; i++ {
		row[i] = "  "
	}
	for day := 1; day <= days; day++ {
		idx := (offset + day - 1) % 7
		row[idx] = fmt.Sprintf("%2d", day)
		if idx == 6 {
			lines = append(lines, strings.Join(row, " "))
			for i := range row {
				row[i] = "  "
			}
		}
	}
	if strings.TrimSpace(strings.Join(row, " ")) != "" {
		lines = append(lines, strings.Join(row, " "))
	}
	return lines
}

func humanBytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	value := float64(n)
	units := []string{"KB", "MB", "GB", "TB"}
	idx := -1
	for value >= unit && idx < len(units)-1 {
		value /= unit
		idx++
	}
	return fmt.Sprintf("%.1f %s", value, units[idx])
}
