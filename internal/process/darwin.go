package process

import (
	"bufio"
	"os/exec"
	"strings"

	"github.com/abhimanyu003/pttr/internal/common"
)

func getDarwinProcesses() ([]common.ProcessInfo, error) {
	cmd := exec.Command("ps", "-eo", "pid,pcpu,pmem,comm,command")
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

func parseUnixPs(output string) []common.ProcessInfo {
	var processes []common.ProcessInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	if scanner.Scan() {

	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		pid := fields[0]
		cpu := fields[1]
		memory := fields[2]
		process := fields[3]

		command := ""
		if len(fields) > 4 {
			command = strings.Join(fields[4:], " ")
		} else {
			command = process
		}

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
