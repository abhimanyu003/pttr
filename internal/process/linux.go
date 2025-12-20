package process

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/abhimanyu003/pttr/internal/common"
)

func getLinuxProcesses() ([]common.ProcessInfo, error) {
	cmd := exec.Command("ps", "-eo", "pid,pcpu,pmem,comm,command")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		processes := parseUnixPs(string(output))
		if len(processes) > 0 {
			return processes, nil
		}
	}

	cmd = exec.Command("ps", "-aux")
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		processes := parseLinuxPsAux(string(output))
		if len(processes) > 0 {
			return processes, nil
		}
	}

	cmd = exec.Command("ps", "-ax")
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("all process detection methods failed (ps -eo, ps -aux, ps -ax): %v", err)
	}

	return parseSimpleProcessOutput(string(output)), nil
}

// parseLinuxPsAux parses 'ps -aux' output format
func parseLinuxPsAux(output string) []common.ProcessInfo {
	var processes []common.ProcessInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	// Skip header line
	if scanner.Scan() {
		// Skip header
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		// ps -aux format: USER PID %CPU %MEM VSZ RSS TTY STAT START TIME COMMAND
		pid := fields[1]
		cpu := fields[2]
		memory := fields[3]
		command := strings.Join(fields[10:], " ")

		// Extract process name from command
		process := fields[10]
		if strings.Contains(process, "/") {
			parts := strings.Split(process, "/")
			process = parts[len(parts)-1]
		}

		// Ensure percentages have % suffix
		if !strings.HasSuffix(cpu, "%") {
			cpu = cpu + "%"
		}
		if !strings.HasSuffix(memory, "%") {
			memory = memory + "%"
		}

		processes = append(processes, common.ProcessInfo{
			PID:     pid,
			Process: process,
			CPU:     cpu,
			Memory:  memory,
			Command: command,
		})
	}

	return processes
}
