package ip

import (
	"bufio"
	"os/exec"
	"strings"

	"github.com/abhimanyu003/pttr/internal/common"
)

func getBSDIPs() ([]common.IPInfo, error) {
	cmd := exec.Command("ifconfig", "-a")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseUnixIfconfig(string(output)), nil
}

func getSolarisIPs() ([]common.IPInfo, error) {
	cmd := exec.Command("ifconfig", "-a")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseUnixIfconfig(string(output)), nil
}

func getGenericUnixIPs() ([]common.IPInfo, error) {
	cmd := exec.Command("ifconfig", "-a")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseUnixIfconfig(string(output)), nil
}

func parseUnixIfconfig(output string) []common.IPInfo {
	var ips []common.IPInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	var currentInterface string
	var currentMAC string
	var currentStatus string

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' && strings.Contains(line, ": flags=") {
			flagsIndex := strings.Index(line, ": flags=")
			if flagsIndex > 0 {
				currentInterface = strings.TrimSpace(line[:flagsIndex])
				if strings.Contains(line, "UP") {
					currentStatus = "UP"
				} else {
					currentStatus = "DOWN"
				}
				currentMAC = ""
			}
		}

		if strings.Contains(line, "ether ") || strings.Contains(line, "lladdr ") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if (part == "ether" || part == "lladdr") && i+1 < len(parts) {
					currentMAC = parts[i+1]
					break
				}
			}
		}

		if strings.Contains(line, "inet ") && !strings.Contains(line, "inet6") {
			parts := strings.Fields(line)
			var ipAddr, netmask, broadcast string

			for i, part := range parts {
				if part == "inet" && i+1 < len(parts) {
					ipAddr = parts[i+1]
				} else if part == "netmask" && i+1 < len(parts) {
					netmask = parts[i+1]
				} else if part == "broadcast" && i+1 < len(parts) {
					broadcast = parts[i+1]
				}
			}

			if ipAddr != "" {
				ips = append(ips, common.IPInfo{
					Interface: currentInterface,
					IPAddress: ipAddr,
					Type:      "IPv4",
					Status:    currentStatus,
					MAC:       currentMAC,
					Netmask:   netmask,
					Broadcast: broadcast,
				})
			}
		}

		if strings.Contains(line, "inet6 ") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "inet6" && i+1 < len(parts) {
					ipAddr := parts[i+1]
					if strings.Contains(ipAddr, "/") {
						ipAddr = strings.Split(ipAddr, "/")[0]
					}
					ips = append(ips, common.IPInfo{
						Interface: currentInterface,
						IPAddress: ipAddr,
						Type:      "IPv6",
						Status:    currentStatus,
						MAC:       currentMAC,
					})
					break
				}
			}
		}
	}

	return ips
}
