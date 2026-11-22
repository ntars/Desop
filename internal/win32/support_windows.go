//go:build windows

package win32

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/ntars/Desop/internal/dock"
	"github.com/ntars/Desop/internal/geometry"
)

const (
	DT_RIGHT_ALIGN = 0x00000002
)

func rgb(r, g, b uint8) uint32 {
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16
}

func moveWindow(h HWND, r geometry.Rect) {
	_, _, _ = user32.NewProc("MoveWindow").Call(uintptr(h), uintptr(r.X), uintptr(r.Y), uintptr(r.Width), uintptr(r.Height), 1)
}

func windowText(h HWND) string {
	n, _, _ := procGetWindowTextLengthW.Call(uintptr(h))
	buf := make([]uint16, int(n)+2)
	procGetWindowTextW.Call(uintptr(h), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

func dockItemsFromConfig(items []string) []dock.Item {
	out := make([]dock.Item, 0, len(items))
	for _, item := range items {
		name := filepath.Base(item)
		if ext := filepath.Ext(name); ext != "" {
			name = strings.TrimSuffix(name, ext)
		}
		out = append(out, dock.Item{ID: strings.ToLower(name), Label: name})
	}
	return out
}

func win32RunningSet() map[string]bool {
	raw := RunningExecutables()
	out := make(map[string]bool, len(raw))
	for name := range raw {
		base := strings.ToLower(filepath.Base(name))
		out[base] = true
		trimmed := strings.TrimSuffix(base, filepath.Ext(base))
		if len(trimmed) > 0 && trimmed != base {
			out[trimmed] = true
		}
	}
	return out
}

func componentClass(index int) string {
	return fmt.Sprintf("DesopComponent%d", index)
}

func dockClass() string { return "DesopDock" }

func widgetClass(index int) string { return fmt.Sprintf("DesopWidget%d", index) }

func islandClass() string { return "DesopIsland" }

// WriteAutostart writes the HKCU Run value through reg.exe. It never touches
// machine-wide registry keys and can be removed by Desop's uninstall step.
func writeAutostartValue(command string) error {
	return runReg("add", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", "Desop", "/t", "REG_SZ", "/d", command, "/f")
}

func removeAutostartValue() error {
	err := runReg("delete", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", "Desop", "/f")
	if err != nil && strings.Contains(err.Error(), "ERROR_SUCCESS") {
		return nil
	}
	return err
}

func runReg(args ...string) error {
	cmd := exec.Command("reg.exe", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("reg %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}
