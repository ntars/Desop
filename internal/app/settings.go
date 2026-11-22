package app

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
)

// Command is a parsed command-line request.
type Command struct {
	Action     string
	Components []string
	ConfigPath string
	Port       int
}

// ParseCommand accepts a small, stable CLI surface:
//
//	desop.exe [start|stop|status|enable <names>|disable <names>|settings|api-event]
//
// Component names are comma-separated: dock,topbar,widgets,island.
func ParseCommand(args []string) (Command, error) {
	cmd := Command{Action: "start"}
	if len(args) == 0 {
		return cmd, nil
	}
	cmd.Action = strings.ToLower(strings.TrimSpace(args[0]))
	switch cmd.Action {
	case "start", "stop", "status", "settings", "api-event", "selftest":
		if len(args) > 1 {
			if cmd.Action == "selftest" && len(args) == 2 {
				seconds, err := strconv.Atoi(strings.TrimSpace(args[1]))
				if err != nil || seconds < 1 || seconds > 3600 {
					return Command{}, fmt.Errorf("selftest duration must be 1-3600 seconds")
				}
				cmd.Port = seconds
				return cmd, nil
			}
			return Command{}, fmt.Errorf("%s does not accept arguments", cmd.Action)
		}
	case "enable", "disable":
		if len(args) != 2 || strings.TrimSpace(args[1]) == "" {
			return Command{}, fmt.Errorf("%s requires component names", cmd.Action)
		}
		for _, part := range strings.Split(args[1], ",") {
			name := strings.ToLower(strings.TrimSpace(part))
			if name == "" {
				continue
			}
			for _, existing := range cmd.Components {
				if existing == name {
					return Command{}, fmt.Errorf("duplicate component %q", name)
				}
			}
			cmd.Components = append(cmd.Components, name)
		}
		if len(cmd.Components) == 0 {
			return Command{}, fmt.Errorf("%s requires component names", cmd.Action)
		}
	default:
		return Command{}, fmt.Errorf("unknown action %q (valid: start, stop, status, enable, disable, settings, api-event, selftest)", args[0])
	}
	return cmd, nil
}

// Usage returns the one-screen CLI reference.
func Usage() string {
	return strings.Join([]string{
		"Desop - macOS-style desktop layer for Windows",
		"",
		"Usage:",
		"  desop.exe                       Start all enabled components",
		"  desop.exe status                Show component status and config path",
		"  desop.exe enable dock,widgets   Enable named components and persist",
		"  desop.exe disable island        Disable named components and persist",
		"  desop.exe settings              Open the graphical settings window",
		"  desop.exe stop                  Stop a running instance",
		"  desop.exe api-event             Show local island API examples",
		"  desop.exe selftest              Run a 10-second idle performance self-test",
		"  desop.exe selftest <seconds>    Run an explicit-duration self-test",
	}, "\n")
}

// StopEventName returns a deterministic, per-user session event name. The
// event exists only in the current Windows session and needs no admin rights.
func StopEventName(configPath string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(strings.ToLower(configPath)))
	return fmt.Sprintf("Local\\Desop_Stop_%016x", h.Sum64())
}
