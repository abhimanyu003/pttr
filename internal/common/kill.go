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
		success, message := killProcessEscalating(pid, pidInt, processName)
		return KillResultMsg{Success: success, Message: message}
	}
}

func killProcessEscalating(pid string, pidInt int, processName string) (bool, string) {
	var methods []func(string, int) error
	var methodNames []string

	switch runtime.GOOS {
	case "windows":
		methods = []func(string, int) error{
			killWindowsTaskkill,
			killWindowsWmic,
		}
		methodNames = []string{"taskkill", "wmic"}
	case "darwin", "linux", "freebsd", "openbsd", "netbsd", "dragonfly", "solaris":
		methods = []func(string, int) error{
			killUnixTerm,
			killUnixKill9,
			killUnixSudo,
		}
		methodNames = []string{"SIGTERM", "SIGKILL", "sudo kill"}
	default:
		methods = []func(string, int) error{killUnixKill9}
		methodNames = []string{"kill -9"}
	}

	for i, method := range methods {
		err := method(pid, pidInt)
		if err == nil {
			var successMsg string
			if processName != "" && processName != pid {
				successMsg = fmt.Sprintf("✅ Successfully killed '%s' (PID %d) using %s", processName, pidInt, methodNames[i])
			} else {
				successMsg = fmt.Sprintf("✅ Successfully killed process (PID %d) using %s", pidInt, methodNames[i])
			}
			return true, successMsg
		}
	}

	if runtime.GOOS == "windows" {
		return false, fmt.Sprintf("❌ Access denied for PID %d. Try running as Administrator.", pidInt)
	} else {
		return false, fmt.Sprintf("❌ Failed to kill PID %d. Try 'sudo kill -9 %s' manually.", pidInt, pid)
	}
}

func killWindowsTaskkill(pid string, pidInt int) error {
	cmd := exec.Command("taskkill", "/F", "/PID", pid)
	return cmd.Run()
}

func killWindowsWmic(pid string, pidInt int) error {
	cmd := exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%s", pid), "delete")
	return cmd.Run()
}

func killUnixTerm(pid string, pidInt int) error {
	cmd := exec.Command("kill", "-TERM", pid)
	return cmd.Run()
}

func killUnixKill9(pid string, pidInt int) error {
	cmd := exec.Command("kill", "-9", pid)
	return cmd.Run()
}

func killUnixSudo(pid string, pidInt int) error {
	if !IsCommandAvailable("sudo") {
		return fmt.Errorf("sudo not available")
	}
	cmd := exec.Command("sudo", "kill", "-9", pid)
	return cmd.Run()
}

func canKillProcess(pid string) bool {
	if pid == "" || pid == "-" {
		return false
	}
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
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
		if cmd.Run() == nil {
			return true
		}
		if runtime.GOOS == "linux" {
			cmd := exec.Command("test", "-d", fmt.Sprintf("/proc/%d", pidInt))
			if cmd.Run() == nil {
				return true
			}
		}
		cmd = exec.Command("ps", "-p", pid)
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			lines := strings.Split(string(output), "\n")
			return len(lines) >= 2
		}
		return false
	default:
		return true
	}
}
