package process

import (
	"bufio"
	"os/exec"
	"runtime"
	"strings"

	"github.com/abhimanyu003/pttr/internal/common"
)

// getBSDProcesses retrieves process information on BSD systems
func getBSDProcesses() ([]common.ProcessInfo, error) {
	cmd := exec.Command("ps", "-axo", "pid,pcpu,pmem,comm,command")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("ps", "-ax")
		output, err = cmd.Output()
		if err != nil {
			return nil, err
		}
		return parseSimpleProcessOutput(string(output)), nil
	}
	return parseUnixPs(string(output)), nil
}

// getSolarisProcesses retrieves process information on Solaris systems
func getSolarisProcesses() ([]common.ProcessInfo, error) {
	cmd := exec.Command("ps", "-eo", "pid,pcpu,pmem,comm,args")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("ps", "-ef")
		output, err = cmd.Output()
		if err != nil {
			return nil, err
		}
		return parseSimpleProcessOutput(string(output)), nil
	}
	return parseUnixPs(string(output)), nil
}

// getGenericUnixProcesses retrieves process information on generic Unix systems
func getGenericUnixProcesses() ([]common.ProcessInfo, error) {
	cmd := exec.Command("ps", "-ax")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseSimpleProcessOutput(string(output)), nil
}

// parseSimpleProcessOutput handles basic ps output format
func parseSimpleProcessOutput(output string) []common.ProcessInfo {
	var processes []common.ProcessInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	// Skip header line if present
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "PID") || strings.Contains(line, "COMMAND") {
			// This looks like a header, skip it
		} else {
			// This is actual data, process it
			processes = append(processes, parseSimpleProcessLine(line)...)
		}
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		processes = append(processes, parseSimpleProcessLine(line)...)
	}

	return processes
}

func parseSimpleProcessLine(line string) []common.ProcessInfo {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return nil
	}

	pid := fields[0]
	command := ""
	process := ""

	if len(fields) >= 2 {
		// Try to extract process name and command
		if len(fields) >= 4 {
			process = fields[len(fields)-1] // Last field is usually the command
			command = strings.Join(fields[1:], " ")
		} else {
			process = fields[1]
			command = strings.Join(fields[1:], " ")
		}
	}

	// Extract just the executable name from full path
	if strings.Contains(process, "/") {
		parts := strings.Split(process, "/")
		process = parts[len(parts)-1]
	}

	return []common.ProcessInfo{{
		PID:     pid,
		Process: process,
		CPU:     "",
		Memory:  "",
		Command: command,
	}}
}

// addParentPIDs adds parent PID information to processes (batch operation for performance)
func addParentPIDs(processes []common.ProcessInfo) []common.ProcessInfo {
	// Only get parent PIDs when needed for tree view
	enhanced := make([]common.ProcessInfo, len(processes))
	copy(enhanced, processes)

	// Batch get parent PIDs for better performance
	switch runtime.GOOS {
	case "darwin", "linux", "freebsd", "openbsd", "netbsd", "dragonfly", "solaris":
		// Use a single ps command to get all parent PIDs
		cmd := exec.Command("ps", "-eo", "pid,ppid")
		output, err := cmd.Output()
		if err != nil {
			return enhanced
		}

		// Parse the output into a map
		ppidMap := make(map[string]string)
		scanner := bufio.NewScanner(strings.NewReader(string(output)))

		// Skip header
		if scanner.Scan() {
			// Skip header line
		}

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				pid := fields[0]
				ppid := fields[1]
				ppidMap[pid] = ppid
			}
		}

		// Apply parent PIDs to processes
		for i := range enhanced {
			if ppid, exists := ppidMap[enhanced[i].PID]; exists {
				enhanced[i].PPID = ppid
			}
		}

	case "windows":
		// For Windows, getting all parent PIDs in one call is complex
		// Skip parent PID detection for performance
		break
	}

	return enhanced
}
