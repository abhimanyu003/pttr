package ports

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/abhimanyu003/pttr/internal/common"
)

// getWindowsPorts retrieves port information on Windows systems
func getWindowsPorts() ([]common.PortInfo, error) {
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseWindowsNetstat(string(output)), nil
}

func parseWindowsNetstat(output string) []common.PortInfo {
	var ports []common.PortInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.Contains(line, "LISTENING") && !strings.Contains(line, "ESTABLISHED") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		proto := fields[0]
		localAddr := fields[1]
		state := ""
		pid := ""

		if len(fields) >= 4 {
			if strings.Contains(line, "LISTENING") || strings.Contains(line, "ESTABLISHED") {
				state = fields[3]
				if len(fields) >= 5 {
					pid = fields[4]
				}
			}
		}

		// Extract port from address
		parts := strings.Split(localAddr, ":")
		if len(parts) < 2 {
			continue
		}
		port := parts[len(parts)-1]

		// Get process name
		process := getWindowsProcessName(pid)

		ports = append(ports, common.PortInfo{
			Port:    port,
			PID:     pid,
			Process: process,
			Proto:   proto,
			State:   state,
		})
	}

	return ports
}

func getWindowsProcessName(pid string) string {
	if pid == "" || pid == "-" {
		return "unknown"
	}
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %s", pid), "/FO", "CSV", "/NH")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%s", pid), "get", "Name", "/format:csv")
		output, err = cmd.Output()
		if err != nil {
			return "unknown"
		}
	}
	result := strings.TrimSpace(string(output))
	if result == "" {
		return "unknown"
	}
	lines := strings.SplitSeq(result, "\n")
	for line := range lines {
		if strings.Contains(line, ",") {
			parts := strings.Split(line, ",")
			if len(parts) > 0 {
				name := strings.Trim(parts[0], "\"")
				if name != "" && name != "Name" {
					return name
				}
			}
		}
	}

	return "unknown"
}
