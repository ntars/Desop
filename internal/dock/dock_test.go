package dock

import (
	"reflect"
	"testing"

	"github.com/ntars/Desop/internal/geometry"
)

func testItems() []Item {
	return []Item{
		{ID: "a", Label: "A", Running: true},
		{ID: "b", Label: "B"},
		{ID: "c", Label: "C", Running: true},
		{ID: "d", Label: "D"},
		{ID: "e", Label: "E"},
	}
}

func TestLayoutCentresOnWorkArea(t *testing.T) {
	opts := LayoutOptions{
		Position: PositionBottom,
		Screen:   geometry.NewRect(0, 0, 1920, 1080),
		WorkArea: geometry.NewRect(0, 0, 1920, 1040),
		Window:   geometry.NewRect(0, 1040-72, 1920, 1040),
		IconSize: 48,
		Gap:      8,
		Radius:   72,
	}
	got := Layout(testItems(), geometry.Point{X: -1000, Y: -1000}, opts)
	if len(got.Icons) != 5 {
		t.Fatalf("got %d icons; want 5", len(got.Icons))
	}
	for _, icon := range got.Icons {
		if icon.Scale != 1 {
			t.Fatalf("distant pointer should not magnify: %+v", icon)
		}
	}
	left := got.Icons[0].Rect.X
	right := got.Icons[len(got.Icons)-1].Rect.Right()
	if centre := (left + right) / 2; centre < 950 || centre > 970 {
		t.Fatalf("dock centre = %d, want ~960", centre)
	}
	for _, icon := range got.Icons {
		if icon.Rect.Y < 900 || icon.Rect.Bottom() > got.Bounds.Bottom()+1 {
			t.Fatalf("icon outside dock: %+v", icon.Rect)
		}
	}
}

func TestMagnificationIsLocal(t *testing.T) {
	opts := LayoutOptions{
		Position:      PositionBottom,
		Screen:        geometry.NewRect(0, 0, 1920, 1080),
		WorkArea:      geometry.NewRect(0, 0, 1920, 1040),
		Window:        geometry.NewRect(0, 1040-72, 1920, 1040),
		IconSize:      48,
		Gap:           8,
		MaxScale:      1.7,
		Spread:        1.2,
		Magnification: true,
		Radius:        72,
	}
	items := testItems()
	centreIcon := items[2]
	centreRect := Layout(items, geometry.Point{X: -1000, Y: 0}, opts).Icons[2].Rect
	pointer := geometry.Point{X: centreRect.Center().X, Y: 1000}
	got := Layout(items, pointer, opts)
	if got.Icons[2].Scale <= 1.3 || got.Icons[2].Scale > opts.MaxScale {
		t.Fatalf("centre scale = %.2f, want (1.3, 1.7]", got.Icons[2].Scale)
	}
	for i, icon := range got.Icons {
		if i > 0 && i < len(got.Icons)-1 && icon.Scale <= 1 {
			t.Fatalf("neighbour %d should magnify slightly: %.2f", i, icon.Scale)
		}
	}
	for i, icon := range got.Icons {
		if i == 0 || i == len(got.Icons)-1 {
			continue
		}
		if icon.Rect.Intersects(got.Icons[i-1].Rect) || icon.Rect.Intersects(got.Icons[i+1].Rect) {
			t.Fatalf("icon %d overlaps a neighbour", i)
		}
	}
	_ = centreIcon
}

func TestVerticalPositions(t *testing.T) {
	screen := geometry.NewRect(0, 0, 1920, 1080)
	for _, position := range []Position{PositionLeft, PositionRight} {
		var window geometry.Rect
		if position == PositionLeft {
			window = geometry.NewRect(0, 0, 52, 1080)
		} else {
			window = geometry.NewRect(1868, 0, 1920, 1080)
		}
		opts := LayoutOptions{Position: position, Screen: screen, WorkArea: screen, Window: window, IconSize: 48, Gap: 8, Radius: 72}
		got := Layout(testItems(), geometry.Point{X: -100, Y: -100}, opts)
		for _, icon := range got.Icons {
			if icon.Rect.Width <= 0 || icon.Rect.Height <= 0 {
				t.Fatalf("%s: invalid icon rect %+v", position, icon.Rect)
			}
			if icon.Rect.X < got.Bounds.X || icon.Rect.Y < got.Bounds.Y || icon.Rect.Right() > got.Bounds.Right() || icon.Rect.Bottom() > got.Bounds.Bottom() {
				t.Fatalf("%s: icon outside dock bounds: icon=%+v bounds=%+v", position, icon.Rect, got.Bounds)
			}
		}
	}
}

func TestRunningIDsSorted(t *testing.T) {
	got := RunningIDs(testItems())
	want := []string{"a", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("running ids = %v, want %v", got, want)
	}
}
