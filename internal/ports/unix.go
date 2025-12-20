package ports

import (
	"bufio"
	"os/exec"
	"strings"

	"github.com/abhimanyu003/pttr/internal/common"
)

func getBSDPorts() ([]common.PortInfo, error) {
	return getDarwinPorts()
}

func getSolarisPorts() ([]common.PortInfo, error) {
	return getDarwinPorts()
}

func getGenericUnixPorts() ([]common.PortInfo, error) {
	cmd := exec.Command("netstat", "-tuln")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	ports, err := parseNetstatOutput(string(output))
	return ports, err
}

func parseNetstatOutput(output string) ([]common.PortInfo, error) {
	ports := parseGenericNetstat(output)
	return ports, nil
}

func parseGenericNetstat(output string) []common.PortInfo {
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
		var localAddr string
		var state = "LISTEN"
		if len(fields) >= 4 {
			localAddr = fields[3]
		}
		if len(fields) >= 6 && (proto == "TCP" || proto == "TCP6") {
			state = fields[5]
		} else if proto == "UDP" || proto == "UDP6" {
			state = "LISTEN"
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
		if state == "LISTEN" || state == "ESTABLISHED" {
			ports = append(ports, common.PortInfo{
				Port:    port,
				PID:     "-",
				Process: "unknown",
				Proto:   proto,
				State:   state,
			})
		}
	}

	return ports
}
