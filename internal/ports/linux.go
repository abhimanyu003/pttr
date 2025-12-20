package ports

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/abhimanyu003/pttr/internal/common"
)

func getLinuxPorts() ([]common.PortInfo, error) {
	cmd := exec.Command("lsof", "-i", "-P", "-n")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		ports := parseUnixLsof(string(output))
		if len(ports) > 0 {
			return ports, nil
		}
	}

	cmd = exec.Command("ss", "-tulpn")
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		ports := parseLinuxSS(string(output))
		if len(ports) > 0 {
			return ports, nil
		}
	}

	cmd = exec.Command("netstat", "-tulpn")
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		ports := parseLinuxNetstat(string(output))
		if len(ports) > 0 {
			return ports, nil
		}
	}

	cmd = exec.Command("netstat", "-tuln")
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("all port detection methods failed (lsof, ss, netstat): %v", err)
	}

	ports, err := parseNetstatOutput(string(output))
	return ports, err
}
func parseLinuxSS(output string) []common.PortInfo {
	var ports []common.PortInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Netid") || strings.HasPrefix(line, "State") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		proto := strings.ToUpper(fields[0])
		state := fields[1]
		localAddr := fields[4]

		if state != "LISTEN" && state != "ESTAB" {
			continue
		}

		var port string
		if strings.Contains(localAddr, ":") {
			parts := strings.Split(localAddr, ":")
			port = parts[len(parts)-1]
		} else {
			continue
		}

		if port == "" || port == "*" {
			continue
		}

		var pid, process string = "-", "unknown"
		if len(fields) >= 6 {
			processInfo := fields[len(fields)-1]
			if strings.Contains(processInfo, "pid=") {

				if strings.Contains(processInfo, "((") && strings.Contains(processInfo, "))") {
					start := strings.Index(processInfo, "((\"") + 3
					end := strings.Index(processInfo[start:], "\"")
					if end > 0 {
						process = processInfo[start : start+end]
					}

					pidStart := strings.Index(processInfo, "pid=") + 4
					pidEnd := strings.Index(processInfo[pidStart:], ",")
					if pidEnd > 0 {
						pid = processInfo[pidStart : pidStart+pidEnd]
					} else {
						pidEnd = strings.Index(processInfo[pidStart:], ")")
						if pidEnd > 0 {
							pid = processInfo[pidStart : pidStart+pidEnd]
						}
					}
				}
			}
		}

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
func parseLinuxNetstat(output string) []common.PortInfo {
	var ports []common.PortInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Active") || strings.HasPrefix(line, "Proto") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		proto := strings.ToUpper(fields[0])
		localAddr := fields[3]
		var state = "LISTEN"
		var pidProcess = "-/unknown"

		if proto == "TCP" || proto == "TCP6" {
			if len(fields) >= 6 {
				state = fields[5]
			}
			if len(fields) >= 7 {
				pidProcess = fields[6]
			}
		} else if proto == "UDP" || proto == "UDP6" {
			state = "LISTEN"
			if len(fields) >= 6 {
				pidProcess = fields[5]
			}
		}

		if state != "LISTEN" && state != "ESTABLISHED" {
			continue
		}

		var port string
		if strings.Contains(localAddr, ":") {
			parts := strings.Split(localAddr, ":")
			port = parts[len(parts)-1]
		} else {
			continue
		}

		if port == "" || port == "*" {
			continue
		}

		var pid, process string = "-", "unknown"
		if pidProcess != "-" && pidProcess != "-/unknown" && strings.Contains(pidProcess, "/") {
			parts := strings.Split(pidProcess, "/")
			if len(parts) >= 2 {
				pid = parts[0]
				process = parts[1]
			}
		}

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
