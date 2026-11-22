package widgets

import (
	"strings"
	"testing"
	"time"
)

func TestRenderMonth(t *testing.T) {
	lines := RenderMonth(time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC))
	if len(lines) < 4 {
		t.Fatalf("month grid too short: %v", lines)
	}
	if lines[0] != "2026-11" || !strings.Contains(lines[1], "Mo Tu") {
		t.Fatalf("unexpected month header: %v", lines[:2])
	}
	joined := strings.Join(lines, " ")
	if !strings.Contains(joined, "22") || !strings.Contains(joined, "30") {
		t.Fatalf("month grid missing dates: %v", lines)
	}
}
