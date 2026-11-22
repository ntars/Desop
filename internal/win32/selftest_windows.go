//go:build windows

package win32

import (
	"fmt"
	"time"

	"github.com/ntars/Desop/internal/config"
)

// RunSelfTest starts the full desktop for a short idle window and measures
// process CPU and working set. It writes a PASS/FAIL report to stdout so the
// Windows acceptance helper can capture objective performance evidence.
func RunSelfTest(cfg config.Config, cfgPath string, duration time.Duration) error {
	if duration <= 0 {
		duration = 10 * time.Second
	}
	w, err := NewDesktopWindow(cfg, cfgPath)
	if err != nil {
		return err
	}
	defer w.Close()
	if err := w.Start(); err != nil {
		return err
	}

	kernelStart, userStart, ok := ProcessTimes()
	if !ok {
		return fmt.Errorf("selftest: GetProcessTimes failed")
	}
	started := time.Now()
	deadline := started.Add(duration)
	for time.Now().Before(deadline) {
		if !w.PumpMessages() {
			return fmt.Errorf("selftest: desktop loop exited early")
		}
		time.Sleep(10 * time.Millisecond)
	}
	kernelEnd, userEnd, ok := ProcessTimes()
	if !ok {
		return fmt.Errorf("selftest: GetProcessTimes failed")
	}
	elapsed := time.Since(started)
	busy := filetimeToUint(kernelEnd) - filetimeToUint(kernelStart) +
		filetimeToUint(userEnd) - filetimeToUint(userStart)
	cpuPercent := float64(busy) / 10_000_000 / elapsed.Seconds() * 100
	workingSetMB := float64(WorkingSetBytes()) / (1024 * 1024)
	cpuOK := cpuPercent < 1
	memOK := workingSetMB < 150

	fmt.Printf("selftest duration: %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("idle CPU: %.2f%% (target < 1%%)\n", cpuPercent)
	fmt.Printf("working set: %.1f MB (target < 150 MB)\n", workingSetMB)
	fmt.Printf("cpu: %s\n", passFail(cpuOK))
	fmt.Printf("memory: %s\n", passFail(memOK))
	if !cpuOK || !memOK {
		return fmt.Errorf("selftest failed")
	}
	return nil
}

func passFail(ok bool) string {
	if ok {
		return "PASS"
	}
	return "FAIL"
}
