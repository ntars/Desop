package widgets

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ntars/Desop/internal/config"
	"github.com/ntars/Desop/internal/geometry"
)

func TestPlaceAllSnapsAndClamps(t *testing.T) {
	cfg := config.Default().Widgets
	cfg.Widgets = []config.WidgetConfig{
		{ID: "a", Kind: "clock", X: 33, Y: 47},
		{ID: "b", Kind: "notes", X: 1900, Y: 1000},
	}
	screen := geometry.NewRect(0, 0, 1920, 1080)
	sizes := map[string]geometry.Rect{
		"clock": {Width: 240, Height: 180},
		"notes": {Width: 240, Height: 180},
	}
	got := PlaceAll(cfg, screen, sizes)
	if len(got) != 2 {
		t.Fatalf("got %d widgets", len(got))
	}
	if got[0].Rect.X != 32 || got[0].Rect.Y != 40 {
		t.Fatalf("grid snap failed: %+v", got[0].Rect)
	}
	if got[1].Rect.Right() > screen.Right() || got[1].Rect.Bottom() > screen.Bottom() {
		t.Fatalf("widget outside screen: %+v", got[1].Rect)
	}
}

func TestPlaceAllPreservesIDsWithGaps(t *testing.T) {
	cfg := config.Default().Widgets
	cfg.GridSnap = false
	cfg.Widgets = []config.WidgetConfig{
		{ID: "a", Kind: "clock", X: 10, Y: 10, Enabled: true},
		{ID: "b", Kind: "notes", X: 100, Y: 100, Enabled: false},
		{ID: "c", Kind: "calendar", X: 200, Y: 200, Enabled: true},
	}
	sizes := map[string]geometry.Rect{
		"clock":    {Width: 240, Height: 180},
		"notes":    {Width: 240, Height: 180},
		"calendar": {Width: 240, Height: 180},
	}
	got := PlaceAll(cfg, geometry.NewRect(0, 0, 1920, 1080), sizes)
	if len(got) != 3 || got[1].ID != "b" || got[2].ID != "c" {
		t.Fatalf("placed IDs changed: %+v", got)
	}
}

func TestHitTestPrefersLastWidget(t *testing.T) {
	a := PlacedWidget{ID: "a", Rect: geometry.NewRect(0, 0, 100, 100)}
	b := PlacedWidget{ID: "b", Rect: geometry.NewRect(50, 50, 150, 150)}
	got := HitTest([]PlacedWidget{a, b}, geometry.Point{X: 60, Y: 60})
	if got == nil || got.ID != "b" {
		t.Fatalf("hit test = %+v", got)
	}
	if HitTest([]PlacedWidget{a, b}, geometry.Point{X: 200, Y: 200}) != nil {
		t.Fatal("unexpected hit")
	}
}

func TestMoveBy(t *testing.T) {
	screen := geometry.NewRect(0, 0, 1920, 1080)
	size := geometry.Rect{Width: 200, Height: 100}
	w := config.WidgetConfig{ID: "a", Kind: "clock", X: 100, Y: 100}
	got := MoveBy(w, 13, -7, screen, size, true, 8)
	if got.X != 112 || got.Y != 96 {
		t.Fatalf("move result = %+v", got)
	}
	got = MoveBy(w, -1000, 1000, screen, size, false, 8)
	if got.X != 0 || got.Y != 980 {
		t.Fatalf("clamped move = %+v", got)
	}
}

func TestRender(t *testing.T) {
	now := time.Date(2026, 11, 22, 21, 7, 3, 0, time.UTC)
	stats := SystemStats{CPUPercent: 12.5, MemoryPercent: 60, DiskPercent: 42, RxBytesPerSec: 1536, TxBytesPerSec: 512}
	title, lines := Render("status", "", now, stats, "")
	if title != "System" || len(lines) != 4 || !strings.Contains(lines[3], "1.5 KB/s") {
		t.Fatalf("status render = %q %v", title, lines)
	}
	_, lines = Render("notes", "line1\r\nline2", now, stats, "")
	if !reflect.DeepEqual(lines, []string{"line1", "line2"}) {
		t.Fatalf("notes lines = %v", lines)
	}
	_, lines = Render("weather", "", now, stats, "")
	if len(lines) != 2 || !strings.Contains(lines[0], "No data source") {
		t.Fatalf("weather should stay offline: %v", lines)
	}
}
