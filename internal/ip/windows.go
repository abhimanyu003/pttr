package ip

import (
	"bufio"
	"os/exec"
	"strings"

	"github.com/abhimanyu003/pttr/internal/common"
)

func getWindowsIPs() ([]common.IPInfo, error) {
	cmd := exec.Command("ipconfig", "/all")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("ipconfig")
		output, err = cmd.Output()
		if err != nil {
			return nil, err
		}
	}

	return parseWindowsIPConfig(string(output)), nil
}

func parseWindowsIPConfig(output string) []common.IPInfo {
	var ips []common.IPInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	var currentInterface string
	var currentMAC string
	var currentStatus string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.Contains(line, "adapter") && strings.HasSuffix(line, ":") {
			parts := strings.Split(line, "adapter")
			if len(parts) >= 2 {
				currentInterface = strings.TrimSpace(strings.TrimSuffix(parts[1], ":"))
			}
			currentMAC = ""
			currentStatus = "UP"
		}

		if strings.Contains(line, "Physical Address") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				currentMAC = strings.TrimSpace(strings.Join(parts[1:], ":"))
			}
		}

		if strings.Contains(line, "IPv4 Address") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				ipAddr := strings.TrimSpace(parts[1])
				if idx := strings.Index(ipAddr, "("); idx != -1 {
					ipAddr = strings.TrimSpace(ipAddr[:idx])
				}
				ips = append(ips, common.IPInfo{
					Interface: currentInterface,
					IPAddress: ipAddr,
					Type:      "IPv4",
					Status:    currentStatus,
					MAC:       currentMAC,
				})
			}
		}

		if strings.Contains(line, "IPv6 Address") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				ipAddr := strings.TrimSpace(strings.Join(parts[1:], ":"))

				if idx := strings.Index(ipAddr, "("); idx != -1 {
					ipAddr = strings.TrimSpace(ipAddr[:idx])
				}

				ips = append(ips, common.IPInfo{
					Interface: currentInterface,
					IPAddress: ipAddr,
					Type:      "IPv6",
					Status:    currentStatus,
					MAC:       currentMAC,
				})
			}
		}
	}

	return ips
}
