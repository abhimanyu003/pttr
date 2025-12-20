package common

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func KillProcessWithName(pid, processName string) tea.Cmd {
	return func() tea.Msg {
		if pid == "" || pid == "-" || pid == "unknown" {
			return KillResultMsg{Success: false, Message: "No PID to kill"}
		}
		pidInt, err := strconv.Atoi(pid)
		if err != nil {
			return KillResultMsg{Success: false, Message: fmt.Sprintf("Invalid PID: %s", pid)}
		}
		if !canKillProcess(pid) {
			return KillResultMsg{Success: false, Message: fmt.Sprintf("Permission denied for PID %d. Try running with sudo.", pidInt)}
		}
		var cmd *exec.Cmd
		var fallbackCmd *exec.Cmd

		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("taskkill", "/F", "/PID", pid)
			fallbackCmd = exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%s", pid), "delete")
		case "darwin", "linux", "freebsd", "openbsd", "netbsd", "dragonfly", "solaris":
			cmd = exec.Command("kill", "-TERM", pid)
			fallbackCmd = exec.Command("kill", "-9", pid)
		default:
			cmd = exec.Command("kill", "-9", pid)
		}

		err = cmd.Run()
		if err != nil && fallbackCmd != nil {

			err = fallbackCmd.Run()
		}

		if err != nil {
			if IsPermissionError(err) {
				if runtime.GOOS == "windows" {
					return KillResultMsg{Success: false, Message: fmt.Sprintf("❌ Access denied for PID %d. Try running as Administrator.", pidInt)}
				} else {
					if IsCommandAvailable("sudo") {
						return KillResultMsg{Success: false, Message: fmt.Sprintf("❌ Permission denied for PID %d. Run 'sudo portmanager' or 'sudo kill -9 %s'", pidInt, pid)}
					} else {
						return KillResultMsg{Success: false, Message: fmt.Sprintf("❌ Permission denied for PID %d. Need root privileges.", pidInt)}
					}
				}
			} else {
				return KillResultMsg{Success: false, Message: fmt.Sprintf("❌ Failed to kill PID %d: %v", pidInt, err)}
			}
		}

		var successMsg string
		if processName != "" && processName != pid {
			successMsg = fmt.Sprintf("✅ Successfully killed '%s' (PID %d)", processName, pidInt)
		} else {
			successMsg = fmt.Sprintf("✅ Successfully killed process (PID %d)", pidInt)
		}

		return KillResultMsg{Success: true, Message: successMsg}
	}
}

func canKillProcess(pid string) bool {
	if pid == "" || pid == "-" {
		return false
	}
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %s", pid), "/FO", "CSV", "/NH")
		output, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.TrimSpace(string(output)) != ""
	case "darwin", "linux", "freebsd", "openbsd", "netbsd", "dragonfly", "solaris":
		cmd := exec.Command("kill", "-0", pid)
		err := cmd.Run()
		return err == nil
	default:
		return true
	}
}
