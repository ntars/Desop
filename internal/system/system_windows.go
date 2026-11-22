//go:build windows

// Package system delegates to the Win32 sampling code. The implementation
// lives in internal/win32 to avoid an import cycle with the desktop shell.
package system

import (
	"github.com/ntars/Desop/internal/topbar"
	"github.com/ntars/Desop/internal/widgets"
	"github.com/ntars/Desop/internal/win32"
)

// StatusReader samples read-only system counters.
type StatusReader struct{ inner *win32.StatusReader }

// NewStatusReader creates a reader with no previous sample.
func NewStatusReader() *StatusReader { return &StatusReader{inner: win32.NewStatusReader()} }

// TopBarStatus returns a conservative status snapshot.
func (r *StatusReader) TopBarStatus() topbar.Status { return r.inner.TopBarStatus() }

// WidgetStats returns CPU, memory, disk and network rates.
func (r *StatusReader) WidgetStats() widgets.SystemStats { return r.inner.WidgetStats() }
