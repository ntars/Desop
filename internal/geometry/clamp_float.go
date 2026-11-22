package geometry

// ClampIntFloat clamps a float percentage into [0, 100].
func ClampIntFloat(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
