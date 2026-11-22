// Package sysmetrics contains OS-independent parsing and rate calculations
// for read-only Windows performance counters.
package sysmetrics

import "encoding/binary"

// MIB_IFROW layout constants from the Windows SDK. The table header is a
// DWORD row count followed by MIB_IFROW[]. Offsets are byte offsets.
const (
	MIBIfRowSize       = 860
	ifTypeOffset       = 516
	ifOperStatusOffset = 552
	ifInOctetsOffset   = 560
	ifOutOctetsOffset  = 584
	ifTypeLoopback     = 24
	ifOperStatusUp     = 1
)

// InterfaceCounters sums octets from active, non-loopback MIB_IFROW entries.
func InterfaceCounters(table []byte) (rx, tx uint64, ok bool) {
	if len(table) < 4 {
		return 0, 0, false
	}
	entries := int(binary.LittleEndian.Uint32(table[:4]))
	if entries < 0 || entries > (len(table)-4)/MIBIfRowSize {
		return 0, 0, false
	}
	found := false
	for i := 0; i < entries; i++ {
		base := 4 + i*MIBIfRowSize
		row := table[base : base+MIBIfRowSize]
		if binary.LittleEndian.Uint32(row[ifTypeOffset:]) == ifTypeLoopback {
			continue
		}
		if binary.LittleEndian.Uint32(row[ifOperStatusOffset:]) != ifOperStatusUp {
			continue
		}
		rx += uint64(binary.LittleEndian.Uint32(row[ifInOctetsOffset:]))
		tx += uint64(binary.LittleEndian.Uint32(row[ifOutOctetsOffset:]))
		found = true
	}
	return rx, tx, found
}

// CPUTimes are raw GetSystemTimes values in 100ns units.
type CPUTimes struct {
	Idle   uint64
	Kernel uint64
	User   uint64
}

// CPUPercent computes utilisation between two samples. It returns 0 when the
// previous sample is missing or counters wrap/reset.
func CPUPercent(previous, current CPUTimes) float64 {
	if current.Kernel < previous.Kernel || current.User < previous.User || current.Idle < previous.Idle {
		return 0
	}
	// GetSystemTimes' kernel time already includes idle time.
	kernelDelta := current.Kernel - previous.Kernel
	userDelta := current.User - previous.User
	idleDelta := current.Idle - previous.Idle
	busyDelta := kernelDelta + userDelta - idleDelta
	if busyDelta < 0 {
		return 0
	}
	totalDelta := kernelDelta + userDelta
	if totalDelta == 0 {
		return 0
	}
	percent := float64(busyDelta) / float64(totalDelta) * 100
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

// RatePerSecond computes a non-negative rate between two byte counters.
func RatePerSecond(previous, current uint64, seconds float64) uint64 {
	if seconds <= 0 || current < previous {
		return 0
	}
	return uint64(float64(current-previous) / seconds)
}
