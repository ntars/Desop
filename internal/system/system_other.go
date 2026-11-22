//go:build !windows

// Package system provides non-Windows stubs so logic packages remain testable.
package system

import (
	"github.com/ntars/Desop/internal/topbar"
	"github.com/ntars/Desop/internal/widgets"
)

// StatusReader is a no-op stub off Windows.
type StatusReader struct{}

// NewStatusReader returns a stub reader.
func NewStatusReader() *StatusReader { return &StatusReader{} }

// TopBarStatus returns a zero status on non-Windows hosts.
func (r *StatusReader) TopBarStatus() topbar.Status { return topbar.Status{} }

// WidgetStats returns zero stats on non-Windows hosts.
func (r *StatusReader) WidgetStats() widgets.SystemStats { return widgets.SystemStats{} }
