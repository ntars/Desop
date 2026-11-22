package sysmetrics

import (
	"encoding/binary"
	"testing"
)

func makeRow(ifType, operStatus, inOctets, outOctets uint32) []byte {
	row := make([]byte, MIBIfRowSize)
	binary.LittleEndian.PutUint32(row[ifTypeOffset:], ifType)
	binary.LittleEndian.PutUint32(row[ifOperStatusOffset:], operStatus)
	binary.LittleEndian.PutUint32(row[ifInOctetsOffset:], inOctets)
	binary.LittleEndian.PutUint32(row[ifOutOctetsOffset:], outOctets)
	return row
}

func makeTable(rows ...[]byte) []byte {
	table := make([]byte, 4)
	binary.LittleEndian.PutUint32(table, uint32(len(rows)))
	for _, row := range rows {
		table = append(table, row...)
	}
	return table
}

func TestInterfaceCounters(t *testing.T) {
	table := makeTable(
		makeRow(6, ifOperStatusUp, 100, 200),
		makeRow(ifTypeLoopback, ifOperStatusUp, 999, 999),
		makeRow(6, 2, 999, 999),
		makeRow(6, ifOperStatusUp, 50, 70),
	)
	rx, tx, ok := InterfaceCounters(table)
	if !ok || rx != 150 || tx != 270 {
		t.Fatalf("counters = %d, %d, %t; want 150, 270, true", rx, tx, ok)
	}
}

func TestInterfaceCountersRejectsMalformedTable(t *testing.T) {
	for _, data := range [][]byte{
		nil,
		make([]byte, 3),
		append(make([]byte, 4), make([]byte, MIBIfRowSize-1)...),
	} {
		if _, _, ok := InterfaceCounters(data); ok {
			t.Fatalf("malformed table accepted: %d bytes", len(data))
		}
	}
}

func TestCPUPercent(t *testing.T) {
	// GetSystemTimes kernel time includes idle time. Kernel grew by 850,
	// user by 50, idle by 800: busy = 850 + 50 - 800 = 100 of 900 total.
	prev := CPUTimes{Idle: 1000, Kernel: 1200, User: 100}
	cur := CPUTimes{Idle: 1800, Kernel: 2050, User: 150}
	if got := CPUPercent(prev, cur); got < 10.9 || got > 11.2 {
		t.Fatalf("CPU percent = %.2f, want about 11.11", got)
	}
	if got := CPUPercent(cur, prev); got != 0 {
		t.Fatalf("counter reset should yield 0, got %.2f", got)
	}
	if got := CPUPercent(prev, prev); got != 0 {
		t.Fatalf("no delta should yield 0, got %.2f", got)
	}
}

func TestRatePerSecond(t *testing.T) {
	if got := RatePerSecond(100, 300, 2); got != 100 {
		t.Fatalf("rate = %d, want 100", got)
	}
	if got := RatePerSecond(300, 100, 2); got != 0 {
		t.Fatalf("counter reset should yield 0, got %d", got)
	}
	if got := RatePerSecond(100, 200, 0); got != 0 {
		t.Fatalf("zero interval should yield 0, got %d", got)
	}
}
