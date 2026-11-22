package dock

import "testing"

func TestScaleIconSize(t *testing.T) {
	tests := []struct {
		base int
		dpi  uint32
		want int
	}{
		{48, 96, 48},
		{48, 144, 72},
		{48, 192, 96},
		{48, 0, 48},
		{1, 96, 24},
		{1000, 96, 192},
	}
	for _, tt := range tests {
		if got := ScaleIconSize(tt.base, tt.dpi); got != tt.want {
			t.Errorf("ScaleIconSize(%d, %d) = %d, want %d", tt.base, tt.dpi, got, tt.want)
		}
	}
}
