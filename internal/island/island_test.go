package island

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ntars/Desop/internal/geometry"
)

func TestEventValidation(t *testing.T) {
	e := Event{Kind: EventMedia, Title: "Song", Progress: 50}
	if err := e.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []Event{
		{Kind: "unknown", Title: "x"},
		{Kind: EventMedia, Title: ""},
		{Kind: EventMedia, Title: "x", Progress: 101},
	} {
		if err := bad.Validate(); err == nil {
			t.Fatalf("expected invalid event: %+v", bad)
		}
	}
}

func TestControllerExpiryAndLayout(t *testing.T) {
	now := time.Now()
	c := New(false, func() time.Time { return now })
	screen := geometry.NewRect(0, 0, 1920, 1080)
	collapsed := Layout(c.Snapshot(), screen)
	if collapsed.Width > 100 || collapsed.X < 900 || collapsed.X > 1000 {
		t.Fatalf("collapsed island = %+v", collapsed)
	}
	event := Event{Kind: EventTimer, Title: "Timer", ExpiresAt: now.Add(time.Second)}
	if err := c.Push(event); err != nil {
		t.Fatal(err)
	}
	expanded := Layout(c.Snapshot(), screen)
	if expanded.Width <= collapsed.Width || expanded.Height <= collapsed.Height {
		t.Fatalf("event should expand island: %+v", expanded)
	}
	now = now.Add(2 * time.Second)
	if c.Snapshot().Active {
		t.Fatal("expired event should be cleared")
	}
	c.Push(Event{Kind: EventCustom, Title: "Persistent"})
	c.Dismiss()
	if c.Snapshot().Active {
		t.Fatal("dismiss should clear event")
	}
}

func TestLocalAPIPortValidation(t *testing.T) {
	if _, err := NewServer(New(false, nil), 80); err == nil {
		t.Fatal("privileged port must be rejected")
	}
}

func TestLocalAPI(t *testing.T) {
	if os.Getenv("DESOP_NETWORK_TESTS") == "" {
		t.Skip("set DESOP_NETWORK_TESTS=1 to run loopback HTTP tests")
	}
	c := New(false, nil)
	srv, err := NewServer(c, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(srv.Addr(), "127.0.0.1:") {
		t.Fatalf("API must be loopback only: %s", srv.Addr())
	}
	go func() { _ = srv.Start() }()
	defer srv.Close()
	base := "http://" + srv.Addr()

	resp, err := http.Get(base + "/health")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(string(body), "ok") {
		t.Fatalf("health = %d %s", resp.StatusCode, body)
	}

	resp, err = http.Post(base+"/event", "application/json", bytes.NewBufferString(`{"kind":"download","title":"update","progress":42}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("push status = %d", resp.StatusCode)
	}
	snap := c.Snapshot()
	if !snap.Active || snap.Title != "update" || snap.Progress != 42 {
		t.Fatalf("snapshot after push = %+v", snap)
	}

	resp, err = http.Post(base+"/dismiss", "application/json", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if c.Snapshot().Active {
		t.Fatal("dismiss endpoint failed")
	}

	resp, err = http.Post(base+"/event", "application/json", strings.NewReader(`{"kind":"media","title":"","progress":0}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid event status = %d", resp.StatusCode)
	}
}
