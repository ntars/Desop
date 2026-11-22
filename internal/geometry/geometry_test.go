package geometry

import "testing"

func TestContainsAndIntersects(t *testing.T) {
	r := NewRect(10, 20, 30, 40)
	if !r.Contains(Point{X: 10, Y: 20}) || r.Contains(Point{X: 30, Y: 20}) {
		t.Fatal("contains boundary is wrong")
	}
	if !r.Intersects(NewRect(25, 25, 50, 50)) || r.Intersects(NewRect(30, 20, 40, 30)) {
		t.Fatal("intersects boundary is wrong")
	}
}

func TestFitInside(t *testing.T) {
	bounds := NewRect(100, 100, 500, 400)
	got := FitInside(Rect{X: 480, Y: 380, Width: 100, Height: 100}, bounds)
	if got.X != 400 || got.Y != 300 || got.Right() > bounds.Right() || got.Bottom() > bounds.Bottom() {
		t.Fatalf("fit result outside bounds: %+v", got)
	}
	oversized := FitInside(Rect{X: 0, Y: 0, Width: 1000, Height: 1000}, bounds)
	if oversized.Width != bounds.Width || oversized.Height != bounds.Height {
		t.Fatalf("oversized rect not clamped: %+v", oversized)
	}
}
