//go:build windows

// Package system reads coarse, read-only Windows status values.
package win32

import (
	"os"
	"time"

	"github.com/ntars/Desop/internal/sysmetrics"
	"github.com/ntars/Desop/internal/topbar"
	"github.com/ntars/Desop/internal/widgets"
)

// StatusReader samples all values needed by the top bar and widgets.
type StatusReader struct {
	lastIdle   uint64
	lastKernel uint64
	lastUser   uint64
	lastNetRX  uint64
	lastNetTX  uint64
	lastNetAt  time.Time
	rxPerSec   uint64
	txPerSec   uint64
}

// NewStatusReader creates a reader with no previous sample.
func NewStatusReader() *StatusReader { return &StatusReader{} }

// TopBarStatus samples battery and network connectivity. Volume and global
// menu are intentionally omitted in MVP because no stable, public,
// dependency-free Win32 API exposes them to a plain user-mode window.
func (r *StatusReader) TopBarStatus() topbar.Status {
	now := time.Now()
	level, state := Battery()
	_, _, ok := NetworkCounters()
	return topbar.Status{
		Now:          now,
		BatteryLevel: level,
		BatteryState: state,
		Online:       ok,
	}
}

// WidgetStats samples CPU, memory, disk and network rates from public counters.
func (r *StatusReader) WidgetStats() widgets.SystemStats {
	idle, kernel, user, ok := SystemTimes()
	cpu := 0.0
	if ok {
		cpu = sysmetrics.CPUPercent(
			sysmetrics.CPUTimes{Idle: r.lastIdle, Kernel: r.lastKernel, User: r.lastUser},
			sysmetrics.CPUTimes{Idle: idle, Kernel: kernel, User: user},
		)
	}
	r.lastIdle, r.lastKernel, r.lastUser = idle, kernel, user

	mem := MemoryStatus()
	memoryPercent := 0.0
	if mem.DwLength > 0 {
		memoryPercent = float64(mem.DwMemoryLoad)
	}
	root := os.Getenv("SystemRoot")
	if root == "" {
		root = "C:\\"
	}
	disk := DiskUsage(root)

	rx, tx, netOK := NetworkCounters()
	now := time.Now()
	if netOK && !r.lastNetAt.IsZero() {
		seconds := now.Sub(r.lastNetAt).Seconds()
		if seconds > 0.2 {
			r.rxPerSec = sysmetrics.RatePerSecond(r.lastNetRX, rx, seconds)
			r.txPerSec = sysmetrics.RatePerSecond(r.lastNetTX, tx, seconds)
		}
	}
	if netOK {
		r.lastNetRX, r.lastNetTX, r.lastNetAt = rx, tx, now
	}

	return widgets.SystemStats{
		CPUPercent:    cpu,
		MemoryPercent: memoryPercent,
		DiskPercent:   disk,
		RxBytesPerSec: r.rxPerSec,
		TxBytesPerSec: r.txPerSec,
	}
}

// Battery reports unknown in MVP. A truthful value requires COM setup that
// would add a third-party dependency; Desop does not invent battery data.
func Battery() (int, topbar.BatteryState) {
	return 0, topbar.BatteryUnknown
}
