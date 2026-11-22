package topbar

import (
	"testing"
	"time"

	"github.com/ntars/Desop/internal/config"
	"github.com/ntars/Desop/internal/geometry"
)

func TestLayoutRegionsDoNotOverlap(t *testing.T) {
	cfg := config.Default().TopBar
	cfg.TrayConsolidation = true
	got := Layout(cfg, LayoutOptions{
		Screen:   geometry.NewRect(0, 0, 1920, 1080),
		WorkArea: geometry.NewRect(0, 32, 1920, 1040),
		Height:   32,
	})
	if got.Bar.Y != 0 || got.Bar.Bottom() != 32 {
		t.Fatalf("bar rect = %+v", got.Bar)
	}
	if got.Menu.Intersects(got.Status) || got.Status.Intersects(got.Tray) {
		t.Fatalf("regions overlap: %+v", got)
	}
	if !got.Status.Contains(geometry.Point{X: got.Status.X + 1, Y: got.Status.Y + 1}) {
		t.Fatalf("invalid status rect: %+v", got.Status)
	}
}

func TestLayoutWithTopTaskbarStaysOnScreen(t *testing.T) {
	cfg := config.Default().TopBar
	got := Layout(cfg, LayoutOptions{
		Screen:   geometry.NewRect(0, 0, 1920, 1080),
		WorkArea: geometry.NewRect(0, 48, 1920, 1080),
		Height:   32,
	})
	if got.Bar.Y < 0 || got.Bar.Bottom() > 1080 {
		t.Fatalf("bar rect off screen: %+v", got.Bar)
	}
}

func TestRenderStatus(t *testing.T) {
	cfg := config.Default().TopBar
	s := Status{
		Now:          time.Date(2026, 11, 22, 21, 7, 0, 0, time.UTC),
		BatteryLevel: 72,
		BatteryState: BatteryDischarging,
		Online:       true,
		Volume:       40,
	}
	got := RenderStatus(cfg, s)
	if got != "21:07  2026-11-22  discharging 72%  net:on  vol:40" {
		t.Fatalf("status text = %q", got)
	}
	cfg.ShowBattery = false
	if RenderStatus(cfg, s) == got {
		t.Fatal("disabling battery should change status text")
	}
	s.BatteryState = BatteryUnknown
	cfg.ShowBattery = true
	if RenderStatus(cfg, s) == got {
		t.Fatal("unknown battery must not be shown")
	}
}

func TestMenuAndAutoHide(t *testing.T) {
	cfg := config.Default().TopBar
	if RenderMenu(cfg, "Notepad") != "" {
		t.Fatal("global menu disabled by default; no menu text expected")
	}
	cfg.GlobalMenu = true
	if RenderMenu(cfg, "") != "" {
		t.Fatal("no fabricated menu for unknown app")
	}
	if RenderMenu(cfg, "Notepad") != "Notepad" {
		t.Fatal("expected active app name")
	}
	if !ShouldAutoHide(cfg, true) || ShouldAutoHide(cfg, false) {
		t.Fatal("fullscreen auto-hide policy is wrong")
	}
}
