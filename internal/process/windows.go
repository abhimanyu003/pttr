package process

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/abhimanyu003/pttr/internal/common"
)

func getWindowsProcesses() ([]common.ProcessInfo, error) {
	processes, err := getWindowsProcessesWithWMIC()
	if err == nil && len(processes) > 0 {
		return processes, nil
	}

	cmd := exec.Command("tasklist", "/FO", "CSV")
	output, err := cmd.Output()
	if err != nil {

		cmd = exec.Command("wmic", "process", "get", "ProcessId,Name,CommandLine", "/format:csv")
		output, err = cmd.Output()
		if err != nil {
			return nil, err
		}
		return parseSimpleProcessOutput(string(output)), nil
	}

	return parseWindowsTasklist(string(output)), nil
}

func parseWindowsTasklist(output string) []common.ProcessInfo {
	var processes []common.ProcessInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	if scanner.Scan() {

	}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := parseCSVLine(line)
		if len(fields) < 5 {
			continue
		}
		process := strings.Trim(fields[0], "\"")
		pid := strings.Trim(fields[1], "\"")
		memory := strings.Trim(fields[4], "\"")
		var memoryFloat float64
		if memory != "" {
			memStr := strings.Replace(memory, ",", "", -1)
			memStr = strings.Replace(memStr, " K", "", -1)
			memStr = strings.Replace(memStr, "K", "", -1)
			if memKB, err := strconv.ParseFloat(memStr, 64); err == nil {
				memoryFloat = memKB / 1024
			}
		}
		processes = append(processes, common.ProcessInfo{
			PID:         pid,
			Process:     process,
			CPU:         "",
			Memory:      memory,
			Command:     process,
			CPUFloat:    0.0,
			MemoryFloat: memoryFloat,
			DisplayName: process,
		})
	}

	return processes
}

func parseCSVLine(line string) []string {
	if line == "" {
		return []string{}
	}
	var fields []string
	var current strings.Builder
	inQuotes := false
	runes := []rune(line)
	for i, char := range runes {
		switch char {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(char)
		case ',':
			if inQuotes {
				current.WriteRune(char)
			} else {
				fields = append(fields, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(char)
		}
		if i == len(runes)-1 {
			fields = append(fields, current.String())
		}
	}

	return fields
}

func getWindowsProcessesWithWMIC() ([]common.ProcessInfo, error) {
	cmd := exec.Command("wmic", "process", "get",
		"ProcessId,Name,CommandLine,WorkingSetSize,PageFileUsage,ParentProcessId",
		"/format:csv")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	processes := parseWMICProcessOutput(string(output))
	cpuUsage, err := getWindowsCPUUsage()
	if err == nil {
		for i := range processes {
			if cpu, exists := cpuUsage[processes[i].PID]; exists {
				processes[i].CPU = fmt.Sprintf("%.1f%%", cpu)
				processes[i].CPUFloat = cpu
			}
		}
	}

	return processes, nil
}

func parseWMICProcessOutput(output string) []common.ProcessInfo {
	var processes []common.ProcessInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	headerSkipped := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !headerSkipped && strings.Contains(line, "CommandLine") {
			headerSkipped = true
			continue
		}
		if !headerSkipped {
			continue
		}
		fields := parseCSVLine(line)
		if len(fields) < 6 {
			continue
		}
		commandLine := strings.Trim(fields[1], "\"")
		name := strings.Trim(fields[2], "\"")
		pageFileUsage := strings.Trim(fields[3], "\"")
		ppid := strings.Trim(fields[4], "\"")
		pid := strings.Trim(fields[5], "\"")
		workingSetSize := strings.Trim(fields[6], "\"")

		if pid == "" || name == "" {
			continue
		}
		var memoryDisplay string
		var memoryFloat float64
		if workingSetSize != "" && workingSetSize != "0" {
			if memBytes, err := strconv.ParseFloat(workingSetSize, 64); err == nil {
				memoryMB := memBytes / (1024 * 1024)
				memoryDisplay = fmt.Sprintf("%.1f MB", memoryMB)
				memoryFloat = memoryMB
			}
		}
		if memoryDisplay == "" && pageFileUsage != "" && pageFileUsage != "0" {
			if memBytes, err := strconv.ParseFloat(pageFileUsage, 64); err == nil {
				memoryMB := memBytes / (1024 * 1024)
				memoryDisplay = fmt.Sprintf("%.1f MB", memoryMB)
				memoryFloat = memoryMB
			}
		}
		process := common.ProcessInfo{
			PID:         pid,
			Process:     name,
			CPU:         "",
			Memory:      memoryDisplay,
			Command:     commandLine,
			PPID:        ppid,
			CPUFloat:    0.0,
			MemoryFloat: memoryFloat,
			DisplayName: name,
		}

		processes = append(processes, process)
	}

	return processes
}

func getWindowsCPUUsage() (map[string]float64, error) {
	psScript := `Get-Process | Select-Object Id, CPU | ForEach-Object { "$($_.Id),$($_.CPU)" }`
	cmd := exec.Command("powershell", "-Command", psScript)
	output, err := cmd.Output()
	if err != nil {
		return getWindowsCPUUsageWMIC()
	}
	return parsePowerShellCPUOutput(string(output)), nil
}

func getWindowsCPUUsageWMIC() (map[string]float64, error) {
	cpuUsage := make(map[string]float64)
	cmd := exec.Command("wmic", "process", "get", "ProcessId,Name", "/format:csv")
	_, err := cmd.Output()
	if err != nil {
		return cpuUsage, err
	}

	return cpuUsage, nil
}

func parsePowerShellCPUOutput(output string) map[string]float64 {
	cpuUsage := make(map[string]float64)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			continue
		}
		pid := strings.TrimSpace(parts[0])
		cpuStr := strings.TrimSpace(parts[1])
		if pid == "" || cpuStr == "" || cpuStr == "0" {
			continue
		}
		if cpuTime, err := strconv.ParseFloat(cpuStr, 64); err == nil && cpuTime > 0 {
			cpuPercent := cpuTime / 100.0
			if cpuPercent > 100 {
				cpuPercent = 100
			}
			cpuUsage[pid] = cpuPercent
		}
	}

	return cpuUsage
}
