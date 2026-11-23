//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ntars/Desop/internal/app"
	"github.com/ntars/Desop/internal/config"
	"github.com/ntars/Desop/internal/island"
	"github.com/ntars/Desop/internal/win32"
)

func run(args []string) error {
	cmd, err := app.ParseCommand(args)
	if err != nil {
		return err
	}
	base := os.Getenv("APPDATA")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("locate user directory: %w", err)
		}
		base = filepath.Join(home, "AppData", "Roaming")
	}
	path := config.ConfigPath(base)

	switch cmd.Action {
	case "api-event":
		fmt.Print(island.APIExample())
		return nil
	case "selftest":
		cfg, err := config.LoadOrCreate(path)
		if err != nil {
			return err
		}
		duration := 10 * time.Second
		if cmd.Port > 0 {
			duration = time.Duration(cmd.Port) * time.Second
		}
		return win32.RunSelfTest(cfg, path, duration)
	case "settings":
		return win32.RunSettings(path)
	case "enable", "disable":
		cfg, err := config.Load(path)
		if errors.Is(err, os.ErrNotExist) {
			cfg, err = config.LoadOrCreate(path)
		}
		if err != nil {
			return err
		}
		enabled := cmd.Action == "enable"
		for _, name := range cmd.Components {
			if err := cfg.SetComponent(name, enabled); err != nil {
				return err
			}
		}
		if err := config.Save(path, cfg); err != nil {
			return err
		}
		fmt.Printf("%s: %v\n", cmd.Action, cfg.EnabledComponents())
		return nil
	case "status":
		cfg, err := config.Load(path)
		if errors.Is(err, os.ErrNotExist) {
			cfg, err = config.LoadOrCreate(path)
		}
		if err != nil {
			return err
		}
		fmt.Printf("config: %s\n", path)
		fmt.Printf("enabled: %v\n", cfg.EnabledComponents())
		fmt.Printf("dock: %s icon=%d magnification=%t\n", cfg.Dock.Position, cfg.Dock.IconSize, cfg.Dock.Magnification)
		fmt.Printf("island local API: %t port=%d\n", cfg.Island.LocalAPI, cfg.Island.LocalAPIPort)
		return nil
	case "stop":
		h, err := win32.OpenStopEvent(app.StopEventName(path))
		if err != nil {
			return fmt.Errorf("no running Desop instance found: %w", err)
		}
		defer win32.CloseHandle(h)
		if err := win32.SetStopEvent(h); err != nil {
			return fmt.Errorf("signal stop: %w", err)
		}
		fmt.Println("stop signal sent")
		return nil
	case "start":
		cfg, err := config.LoadOrCreate(path)
		if err != nil {
			return err
		}
		return startDesktop(cfg, path)
	default:
		return fmt.Errorf("unsupported action %q", cmd.Action)
	}
}

func startDesktop(cfg config.Config, path string) error {
	// Win32 windows and their message queues belong to the creating thread.
	// Keep the desktop owner on one OS thread for its entire lifetime.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	w, err := win32.NewDesktopWindow(cfg, path)
	if err != nil {
		return err
	}
	defer w.Close()
	if err := w.Start(); err != nil {
		return err
	}
	fmt.Println("Desop running. Tray icon: right-click for settings or exit.")
	// The tray icon remains available until the user exits through its menu.
	// A short sleep keeps the polling loop from consuming a full CPU core.
	for {
		if err := w.RefreshDue(); err != nil {
			return err
		}
		if !w.PumpMessages() {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
}
