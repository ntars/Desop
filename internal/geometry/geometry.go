// Package geometry contains small, OS-independent rectangle helpers.
package geometry

import "math"

// Point is an integer screen coordinate.
type Point struct {
	X int
	Y int
}

// Rect is an integer screen rectangle.
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// NewRect builds a rect from left, top, right and bottom coordinates.
func NewRect(left, top, right, bottom int) Rect {
	return Rect{X: left, Y: top, Width: right - left, Height: bottom - top}
}

// Right returns the right edge coordinate.
func (r Rect) Right() int { return r.X + r.Width }

// Bottom returns the bottom edge coordinate.
func (r Rect) Bottom() int { return r.Y + r.Height }

// Center returns the rectangle centre point.
func (r Rect) Center() Point {
	return Point{X: r.X + r.Width/2, Y: r.Y + r.Height/2}
}

// Contains reports whether p lies inside r.
func (r Rect) Contains(p Point) bool {
	return p.X >= r.X && p.X < r.Right() && p.Y >= r.Y && p.Y < r.Bottom()
}

// Intersects reports whether two rectangles overlap.
func (a Rect) Intersects(b Rect) bool {
	return a.X < b.Right() && a.Right() > b.X && a.Y < b.Bottom() && a.Bottom() > b.Y
}

// Scale multiplies every coordinate and dimension by factor.
func (r Rect) Scale(factor float64) Rect {
	return Rect{
		X:      int(math.Round(float64(r.X) * factor)),
		Y:      int(math.Round(float64(r.Y) * factor)),
		Width:  int(math.Round(float64(r.Width) * factor)),
		Height: int(math.Round(float64(r.Height) * factor)),
	}
}

// Clamp keeps value within [low, high].
func Clamp(value, low, high int) int {
	if high < low {
		return low
	}
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

// FitInside moves and shrinks r until it fits inside bounds. Oversized rects
// are aligned to the top-left of bounds.
func FitInside(r, bounds Rect) Rect {
	if r.Width > bounds.Width {
		r.Width = bounds.Width
	}
	if r.Height > bounds.Height {
		r.Height = bounds.Height
	}
	if r.X < bounds.X {
		r.X = bounds.X
	}
	if r.Y < bounds.Y {
		r.Y = bounds.Y
	}
	if r.Right() > bounds.Right() {
		r.X = bounds.Right() - r.Width
	}
	if r.Bottom() > bounds.Bottom() {
		r.Y = bounds.Bottom() - r.Height
	}
	return r
}
