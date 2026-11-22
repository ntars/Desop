package dock

import (
	"testing"
	"time"

	"github.com/ntars/Desop/internal/geometry"
)

func TestBounceDecaysToRest(t *testing.T) {
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	opts := LayoutOptions{
		Position:    PositionBottom,
		Screen:      geometry.NewRect(0, 0, 1920, 1080),
		WorkArea:    geometry.NewRect(0, 0, 1920, 1040),
		Window:      geometry.NewRect(0, 968, 1920, 1040),
		IconSize:    48,
		Gap:         8,
		Bounce:      true,
		LaunchTimes: map[string]time.Time{"a": now},
		Now:         now,
	}
	items := []Item{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}}
	start := Layout(items, geometry.Point{X: -1000}, opts).Icons[0].Scale
	if start <= 1 {
		t.Fatalf("bounce should start above 1, got %.3f", start)
	}
	opts.Now = now.Add(time.Second)
	rest := Layout(items, geometry.Point{X: -1000}, opts).Icons[0].Scale
	if rest != 1 {
		t.Fatalf("bounce should settle to 1, got %.3f", rest)
	}
}
