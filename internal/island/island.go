// Package island implements the software dynamic island and its local push API.
package island

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ntars/Desop/internal/geometry"
)

// EventKind is one of the supported island event sources.
type EventKind string

const (
	EventMedia    EventKind = "media"
	EventTimer    EventKind = "timer"
	EventDownload EventKind = "download"
	EventCustom   EventKind = "custom"
)

// Event is a transient status pushed into the island. It is never persisted.
type Event struct {
	Kind      EventKind `json:"kind"`
	Title     string    `json:"title"`
	Detail    string    `json:"detail"`
	Progress  int       `json:"progress"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Validate checks event bounds. Progress is 0-100; empty titles are rejected.
func (e Event) Validate() error {
	switch e.Kind {
	case EventMedia, EventTimer, EventDownload, EventCustom:
	default:
		return fmt.Errorf("unsupported event kind %q", e.Kind)
	}
	if strings.TrimSpace(e.Title) == "" {
		return errors.New("event title must not be empty")
	}
	if e.Progress < 0 || e.Progress > 100 {
		return fmt.Errorf("progress %d out of range [0,100]", e.Progress)
	}
	return nil
}

// Snapshot is the current visible island state.
type Snapshot struct {
	Active   bool
	Expanded bool
	Kind     EventKind
	Title    string
	Detail   string
	Progress int
	Bounds   geometry.Rect
}

// Controller holds transient events and computes island geometry.
type Controller struct {
	mu       sync.Mutex
	expanded bool
	event    *Event
	now      func() time.Time
}

// New creates a controller. If now is nil, time.Now is used.
func New(expanded bool, now func() time.Time) *Controller {
	if now == nil {
		now = time.Now
	}
	return &Controller{expanded: expanded, now: now}
}

// Push replaces the active event after validation.
func (c *Controller) Push(e Event) error {
	if err := e.Validate(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	event := e
	c.event = &event
	return nil
}

// Dismiss clears the active event.
func (c *Controller) Dismiss() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.event = nil
}

// SetExpanded controls the manual expanded state.
func (c *Controller) SetExpanded(v bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.expanded = v
}

// Snapshot returns a copy of the visible state. Expired events are cleared.
func (c *Controller) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.event != nil && !c.event.ExpiresAt.IsZero() && !c.now().Before(c.event.ExpiresAt) {
		c.event = nil
	}
	snap := Snapshot{Expanded: c.expanded}
	if c.event != nil {
		snap.Active = true
		snap.Kind = c.event.Kind
		snap.Title = c.event.Title
		snap.Detail = c.event.Detail
		snap.Progress = c.event.Progress
	}
	return snap
}

// Layout computes the island capsule. When an event is active or the island
// is expanded, the capsule widens; otherwise it stays small at the top centre.
func Layout(s Snapshot, screen geometry.Rect) geometry.Rect {
	if screen.Width <= 0 || screen.Height <= 0 {
		screen = geometry.NewRect(0, 0, 1920, 1080)
	}
	width, height := 72, 28
	if s.Active || s.Expanded {
		width, height = 420, 96
	}
	x := screen.X + (screen.Width-width)/2
	return geometry.NewRect(x, screen.Y+8, x+width, screen.Y+8+height)
}

// Server exposes a loopback-only HTTP API for third-party event pushes.
type Server struct {
	controller *Controller
	server     *http.Server
	listener   net.Listener
}

// NewServer binds to 127.0.0.1 only; external interfaces are never exposed.
// Port 0 is accepted for tests and lets the OS choose a free loopback port.
func NewServer(controller *Controller, port int) (*Server, error) {
	if port < 0 || port > 65535 {
		return nil, fmt.Errorf("invalid local API port %d", port)
	}
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil && port > 0 && port < 65535 {
		// A busy port must not prevent the desktop shell from starting.
		for candidate := port + 1; candidate <= port+10 && candidate <= 65535; candidate++ {
			listener, err = net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(candidate)))
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("listen on loopback: %w", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ok":true}`)
	})
	mux.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var e Event
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
		if err := dec.Decode(&e); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if err := controller.Push(e); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/dismiss", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		controller.Dismiss()
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/event/collapse", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		controller.SetExpanded(false)
		w.WriteHeader(http.StatusNoContent)
	})
	return &Server{
		controller: controller,
		listener:   listener,
		server:     &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second},
	}, nil
}

// Addr returns the bound loopback address.
func (s *Server) Addr() string { return s.listener.Addr().String() }

// Start serves requests until Close is called.
func (s *Server) Start() error {
	err := s.server.Serve(s.listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Close shuts down the local API.
func (s *Server) Close() error { return s.server.Close() }

// APIExample returns copy-pasteable local API examples for the CLI.
func APIExample() string {
	return strings.Join([]string{
		"The island API listens on 127.0.0.1 only (default port 17890).",
		"",
		"Push an event (Windows PowerShell):",
		`  Invoke-RestMethod -Method Post -Uri http://127.0.0.1:17890/event -ContentType application/json -Body '{"kind":"download","title":"Update","detail":"Downloading","progress":42}'`,
		"",
		"Dismiss the active event:",
		"  Invoke-RestMethod -Method Post -Uri http://127.0.0.1:17890/dismiss",
		"",
		"Collapse an expanded island without clearing the event:",
		"  Invoke-RestMethod -Method Post -Uri http://127.0.0.1:17890/event/collapse",
		"",
		"Supported kinds: media, timer, download, custom. Progress: 0-100.",
	}, "\n")
}
