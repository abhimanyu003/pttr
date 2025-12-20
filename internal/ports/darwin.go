package ports

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/abhimanyu003/pttr/internal/common"
)

func getDarwinPorts() ([]common.PortInfo, error) {
	cmd := exec.Command("lsof", "-i", "-P", "-n")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("netstat", "-tuln")
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("both lsof and netstat failed: %v", err)
		}
		ports, err := parseNetstatOutput(string(output))
		return ports, err
	}
	return parseUnixLsof(string(output)), nil
}

func parseUnixLsof(output string) []common.PortInfo {
	var ports []common.PortInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "COMMAND") {
			continue
		}
		if !strings.Contains(line, "LISTEN") && !strings.Contains(line, "ESTABLISHED") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}
		process := fields[0]
		pid := fields[1]
		proto := ""
		port := ""
		state := ""
		for _, field := range fields {
			if strings.Contains(field, "TCP") {
				proto = "TCP"
			} else if strings.Contains(field, "UDP") {
				proto = "UDP"
			}
			if strings.Contains(field, ":") {
				if strings.Contains(field, "*:") ||
					strings.Contains(field, "127.0.0.1:") ||
					strings.Contains(field, "0.0.0.0:") ||
					strings.Contains(field, "::1:") ||
					strings.Contains(field, "[::]:") ||
					strings.Contains(field, "localhost:") {
					parts := strings.Split(field, ":")
					if len(parts) >= 2 {
						portCandidate := parts[len(parts)-1]

						if len(portCandidate) > 0 && portCandidate != "*" {

							port = portCandidate
						}
					}
				}
			}

			if strings.Contains(field, "(LISTEN)") {
				state = "LISTEN"
			} else if strings.Contains(field, "(ESTABLISHED)") {
				state = "ESTABLISHED"
			}
		}

		if port != "" && proto != "" {
			ports = append(ports, common.PortInfo{
				Port:    port,
				PID:     pid,
				Process: process,
				Proto:   proto,
				State:   state,
			})
		}
	}

	return ports
}
